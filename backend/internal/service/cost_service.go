package service

import (
	"fmt"
	"log"
	"time"

	"github.com/shopspring/decimal"
	"stalll-hub-pos/backend/internal/dto"
	"stalll-hub-pos/backend/internal/model"
	"stalll-hub-pos/backend/pkg/database"

	"github.com/xuri/excelize/v2"
)

type CostService struct{}

func NewCostService() *CostService {
	return &CostService{}
}

func (s *CostService) ImportCostExcel(filePath string, effectiveDate string) (*model.CostImportBatch, error) {
	f, err := excelize.OpenFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open excel file: %v", err)
	}
	defer f.Close()

	sheetName := f.GetSheetName(0)
	rows, err := f.GetRows(sheetName)
	if err != nil {
		return nil, fmt.Errorf("failed to read excel rows: %v", err)
	}

	if len(rows) < 2 {
		return nil, fmt.Errorf("excel file is empty or has no data rows")
	}

	batchNo := fmt.Sprintf("COST%s%d", time.Now().Format("20060102150405"), time.Now().UnixNano()%10000)
	batch := &model.CostImportBatch{
		BatchNo:       batchNo,
		FileName:      filePath,
		TotalRows:     len(rows) - 1,
		Status:        0,
		EffectiveDate: effectiveDate,
	}

	if err := database.DB.Create(batch).Error; err != nil {
		return nil, fmt.Errorf("failed to create batch record: %v", err)
	}

	successCount := 0
	failCount := 0

	for i, row := range rows {
		if i == 0 {
			continue
		}

		if len(row) < 3 {
			failCount++
			continue
		}

		productName := row[0]
		unitCostStr := row[1]
		priceStr := ""
		if len(row) >= 3 {
			priceStr = row[2]
		}

		unitCost, err := decimal.NewFromString(unitCostStr)
		if err != nil {
			failCount++
			continue
		}

		var price decimal.Decimal
		if priceStr != "" {
			price, _ = decimal.NewFromString(priceStr)
		}

		var product model.Product
		if err := database.DB.Where("name = ?", productName).First(&product).Error; err != nil {
			failCount++
			continue
		}

		grossProfit := price.Sub(unitCost)
		var grossMargin decimal.Decimal
		if price.GreaterThan(decimal.Zero) {
			grossMargin = grossProfit.Div(price).Mul(decimal.NewFromInt(100))
		}

		cost := &model.ProductCost{
			ProductID:     product.ID,
			ProductName:   productName,
			UnitCost:      unitCost,
			Price:         price,
			GrossProfit:   grossProfit,
			GrossMargin:   grossMargin,
			EffectiveDate: effectiveDate,
			BatchNo:       batchNo,
		}

		if err := database.DB.Create(cost).Error; err != nil {
			failCount++
			continue
		}

		successCount++
	}

	now := time.Now()
	batch.SuccessCount = successCount
	batch.FailCount = failCount
	batch.Status = 1
	batch.CompletedAt = &now
	database.DB.Save(batch)

	log.Printf("[CostImport] Batch %s: total=%d, success=%d, fail=%d", batchNo, batch.TotalRows, successCount, failCount)
	return batch, nil
}

func (s *CostService) GetCostList(query *dto.CostQueryDTO) ([]model.ProductCost, int64, error) {
	var costs []model.ProductCost
	var total int64

	db := database.DB.Model(&model.ProductCost{})
	if query.ProductID > 0 {
		db = db.Where("product_id = ?", query.ProductID)
	}
	if query.EffectiveDate != "" {
		db = db.Where("effective_date = ?", query.EffectiveDate)
	}

	db.Count(&total)

	if err := db.Preload("Product").
		Order("id DESC").
		Offset((query.Page - 1) * query.PageSize).
		Limit(query.PageSize).
		Find(&costs).Error; err != nil {
		return nil, 0, err
	}

	return costs, total, nil
}

func (s *CostService) GetProfitReport(query *dto.ProfitReportQueryDTO) ([]dto.ProfitReportResponse, error) {
	var results []dto.ProfitReportResponse

	startDate := query.StartDate
	endDate := query.EndDate

	type SalesData struct {
		ProductID   uint
		ProductName string
		Quantity    int
		Revenue     decimal.Decimal
	}

	var salesData []SalesData
	db := database.DB.Table("order_items oi").
		Select(`oi.product_id, oi.product_name, SUM(oi.quantity) as quantity, SUM(oi.subtotal) as revenue`).
		Joins("LEFT JOIN orders o ON oi.order_id = o.id").
		Where("o.created_at >= ? AND o.created_at <= ?", startDate+" 00:00:00", endDate+" 23:59:59").
		Where("o.order_status != ?", -1)

	if query.StoreID > 0 {
		db = db.Where("o.store_id = ?", query.StoreID)
	}
	db = db.Group("oi.product_id, oi.product_name").Scan(&salesData)

	var totalRevenue, totalCost decimal.Decimal
	var orderCount int64

	database.DB.Model(&model.Order{}).
		Where("created_at >= ? AND created_at <= ?", startDate+" 00:00:00", endDate+" 23:59:59").
		Where("order_status != ?", -1).
		Count(&orderCount)

	type ProductCostRow struct {
		ProductID uint
		UnitCost  decimal.Decimal
	}

	var costRows []ProductCostRow
	if len(salesData) > 0 {
		productIDs := make([]uint, len(salesData))
		for i, s := range salesData {
			productIDs[i] = s.ProductID
		}

		database.DB.Model(&model.ProductCost{}).
			Where("product_id IN ? AND effective_date <= ?", productIDs, endDate).
			Group("product_id").
			Select("product_id, MAX(unit_cost) as unit_cost").
			Scan(&costRows)
	}

	costMap := make(map[uint]decimal.Decimal)
	for _, c := range costRows {
		costMap[c.ProductID] = c.UnitCost
	}

	for _, s := range salesData {
		unitCost, hasCost := costMap[s.ProductID]
		if !hasCost {
			unitCost = decimal.Zero
		}

		totalProductCost := unitCost.Mul(decimal.NewFromInt(int64(s.Quantity)))
		grossProfit := s.Revenue.Sub(totalProductCost)
		var grossMargin decimal.Decimal
		if s.Revenue.GreaterThan(decimal.Zero) {
			grossMargin = grossProfit.Div(s.Revenue).Mul(decimal.NewFromInt(100))
		}

		totalRevenue = totalRevenue.Add(s.Revenue)
		totalCost = totalCost.Add(totalProductCost)

		results = append(results, dto.ProfitReportResponse{
			ProductID:    s.ProductID,
			ProductName:  s.ProductName,
			Quantity:     s.Quantity,
			Revenue:      s.Revenue,
			UnitCost:     unitCost,
			TotalCost:    totalProductCost,
			GrossProfit:  grossProfit,
			GrossMargin:  grossMargin,
		})
	}

	var overallGrossMargin decimal.Decimal
	overallGrossProfit := totalRevenue.Sub(totalCost)
	if totalRevenue.GreaterThan(decimal.Zero) {
		overallGrossMargin = overallGrossProfit.Div(totalRevenue).Mul(decimal.NewFromInt(100))
	}

	_ = orderCount

	return results, nil
}

func (s *CostService) GetProfitSummary(query *dto.ProfitReportQueryDTO) (*dto.ProfitSummaryResponse, error) {
	report, err := s.GetProfitReport(query)
	if err != nil {
		return nil, err
	}

	var totalRevenue, totalCost decimal.Decimal
	var productCount int

	for _, r := range report {
		totalRevenue = totalRevenue.Add(r.Revenue)
		totalCost = totalCost.Add(r.TotalCost)
		productCount++
	}

	grossProfit := totalRevenue.Sub(totalCost)
	var grossMargin decimal.Decimal
	if totalRevenue.GreaterThan(decimal.Zero) {
		grossMargin = grossProfit.Div(totalRevenue).Mul(decimal.NewFromInt(100))
	}

	return &dto.ProfitSummaryResponse{
		TotalRevenue:  totalRevenue,
		TotalCost:     totalCost,
		GrossProfit:   grossProfit,
		GrossMargin:   grossMargin,
		ProductCount:  productCount,
	}, nil
}
