package service

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/shopspring/decimal"
	"stalll-hub-pos/backend/config"
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

	if len(rows) == 0 {
		return nil, fmt.Errorf("excel file is empty")
	}

	batchNo := fmt.Sprintf("COST%s%d", time.Now().Format("20060102150405"), time.Now().UnixNano()%10000)
	batch := &model.CostImportBatch{
		BatchNo:       batchNo,
		FileName:      filePath,
		TotalRows:     0,
		Status:        0,
		EffectiveDate: effectiveDate,
	}

	if err := database.DB.Create(batch).Error; err != nil {
		return nil, fmt.Errorf("failed to create batch record: %v", err)
	}

	colMapping := detectColumns(rows[0])

	startRow := 0
	if colMapping.hasHeader {
		startRow = 1
	}

	dataRows := rows[startRow:]
	batch.TotalRows = len(dataRows)

	successCount := 0
	failCount := 0

	for i, row := range dataRows {
		rowNum := startRow + i + 1
		if isEmptyRow(row) {
			failCount++
			continue
		}

		productName := getColValue(row, colMapping.nameIdx)
		costStr := getColValue(row, colMapping.costIdx)
		priceStr := getColValue(row, colMapping.priceIdx)

		productName = strings.TrimSpace(productName)
		costStr = strings.TrimSpace(costStr)
		priceStr = strings.TrimSpace(priceStr)

		if productName == "" {
			log.Printf("[CostImport] Row %d: product name is empty, skip", rowNum)
			failCount++
			continue
		}

		if costStr == "" {
			log.Printf("[CostImport] Row %d: cost is empty, skip (product: %s)", rowNum, productName)
			failCount++
			continue
		}

		unitCost, err := parseDecimal(costStr)
		if err != nil || unitCost.LessThan(decimal.Zero) {
			log.Printf("[CostImport] Row %d: invalid cost value '%s' (product: %s)", rowNum, costStr, productName)
			failCount++
			continue
		}

		var price decimal.Decimal
		if priceStr != "" {
			if p, err := parseDecimal(priceStr); err == nil && p.GreaterThanOrEqual(decimal.Zero) {
				price = p
			}
		}

		var product model.Product
		if err := database.DB.Where("name = ?", productName).First(&product).Error; err != nil {
			log.Printf("[CostImport] Row %d: product '%s' not found", rowNum, productName)
			failCount++
			continue
		}

		if price.IsZero() && !product.Price.IsZero() {
			price = product.Price
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
			log.Printf("[CostImport] Row %d: create cost record failed: %v", rowNum, err)
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

type columnMapping struct {
	nameIdx    int
	costIdx    int
	priceIdx   int
	hasHeader  bool
}

func detectColumns(headerRow []string) columnMapping {
	mapping := columnMapping{
		nameIdx:  0,
		costIdx:  1,
		priceIdx: 2,
	}

	if len(headerRow) == 0 {
		return mapping
	}

	nameKeywords := []string{"菜品", "商品", "名称", "product", "name", "菜名"}
	costKeywords := []string{"成本", "进价", "cost", "unit_cost", "单位成本"}
	priceKeywords := []string{"售价", "价格", "price", "销售价", "定价"}

	foundName := -1
	foundCost := -1
	foundPrice := -1

	for i, cell := range headerRow {
		cellLower := strings.ToLower(strings.TrimSpace(cell))
		cellTrim := strings.TrimSpace(cell)

		if foundName == -1 && containsAny(cellLower, cellTrim, nameKeywords) {
			foundName = i
		}
		if foundCost == -1 && containsAny(cellLower, cellTrim, costKeywords) {
			foundCost = i
		}
		if foundPrice == -1 && containsAny(cellLower, cellTrim, priceKeywords) {
			foundPrice = i
		}
	}

	if foundName >= 0 || foundCost >= 0 || foundPrice >= 0 {
		mapping.hasHeader = true
		if foundName >= 0 {
			mapping.nameIdx = foundName
		}
		if foundCost >= 0 {
			mapping.costIdx = foundCost
		}
		if foundPrice >= 0 {
			mapping.priceIdx = foundPrice
		}
	}

	return mapping
}

func containsAny(lower, raw string, keywords []string) bool {
	for _, kw := range keywords {
		if strings.Contains(lower, strings.ToLower(kw)) || strings.Contains(raw, kw) {
			return true
		}
	}
	return false
}

func getColValue(row []string, idx int) string {
	if idx >= 0 && idx < len(row) {
		return row[idx]
	}
	return ""
}

func isEmptyRow(row []string) bool {
	for _, cell := range row {
		if strings.TrimSpace(cell) != "" {
			return false
		}
	}
	return true
}

func parseDecimal(s string) (decimal.Decimal, error) {
	s = strings.TrimSpace(s)
	s = strings.ReplaceAll(s, "¥", "")
	s = strings.ReplaceAll(s, "￥", "")
	s = strings.ReplaceAll(s, ",", "")
	s = strings.ReplaceAll(s, "，", "")
	s = strings.TrimSpace(s)

	if s == "" {
		return decimal.Zero, nil
	}

	if _, err := strconv.ParseFloat(s, 64); err != nil {
		return decimal.Zero, fmt.Errorf("invalid number: %s", s)
	}

	return decimal.NewFromString(s)
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
		Where("o.order_status != ? AND o.pay_status = ? AND oi.status = ?", -1, 1, 1)

	if query.StoreID > 0 {
		db = db.Where("o.store_id = ?", query.StoreID)
	}
	db = db.Group("oi.product_id, oi.product_name").Order("revenue DESC").Scan(&salesData)

	var totalRevenue, totalCost decimal.Decimal
	totalRevenue = decimal.Zero
	totalCost = decimal.Zero
	_ = totalRevenue
	_ = totalCost

	var productIDs []uint
	for _, s := range salesData {
		productIDs = append(productIDs, s.ProductID)
	}

	type ProductCostRow struct {
		ProductID uint
		UnitCost  decimal.Decimal
	}

	var costRows []ProductCostRow
	if len(productIDs) > 0 {
		database.DB.Model(&model.ProductCost{}).
			Where("product_id IN ? AND effective_date <= ?", productIDs, endDate).
			Select("product_id, MAX(unit_cost) as unit_cost").
			Group("product_id").
			Scan(&costRows)
	}

	costMap := make(map[uint]decimal.Decimal)
	for _, c := range costRows {
		costMap[c.ProductID] = c.UnitCost
	}

	var results []dto.ProfitReportResponse
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
		TotalRevenue: totalRevenue,
		TotalCost:    totalCost,
		GrossProfit:  grossProfit,
		GrossMargin:  grossMargin,
		ProductCount: productCount,
	}, nil
}

func (s *CostService) GetProfitReportV2(query *dto.ProfitReportQueryDTO) ([]dto.ProfitReportV2Response, error) {
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
		Where("o.order_status != ? AND o.pay_status = ? AND oi.status = ?", -1, 1, 1)

	if query.StoreID > 0 {
		db = db.Where("o.store_id = ?", query.StoreID)
	}
	db = db.Group("oi.product_id, oi.product_name").Order("revenue DESC").Scan(&salesData)

	var productIDs []uint
	for _, s := range salesData {
		productIDs = append(productIDs, s.ProductID)
	}

	bomService := NewBOMService()
	bomCostMap, _ := bomService.BatchCalculateProductCosts(productIDs)

	staticCostMap := make(map[uint]decimal.Decimal)
	if len(productIDs) > 0 {
		type ProductCostRow struct {
			ProductID uint
			UnitCost  decimal.Decimal
		}
		var costRows []ProductCostRow
		database.DB.Model(&model.ProductCost{}).
			Where("product_id IN ? AND effective_date <= ?", productIDs, endDate).
			Select("product_id, MAX(unit_cost) as unit_cost").
			Group("product_id").
			Scan(&costRows)
		for _, c := range costRows {
			staticCostMap[c.ProductID] = c.UnitCost
		}
	}

	var results []dto.ProfitReportV2Response
	for _, s := range salesData {
		unitCost, hasBOMCost := bomCostMap[s.ProductID]
		if !hasBOMCost || unitCost.IsZero() {
			unitCost = staticCostMap[s.ProductID]
		}

		unitPrice := decimal.Zero
		if s.Quantity > 0 {
			unitPrice = s.Revenue.Div(decimal.NewFromInt(int64(s.Quantity)))
		}

		totalMaterialCost := unitCost.Mul(decimal.NewFromInt(int64(s.Quantity)))
		grossProfit := s.Revenue.Sub(totalMaterialCost)
		var grossMargin decimal.Decimal
		if s.Revenue.GreaterThan(decimal.Zero) {
			grossMargin = grossProfit.Div(s.Revenue).Mul(decimal.NewFromInt(100))
		}

		results = append(results, dto.ProfitReportV2Response{
			ProductID:    s.ProductID,
			ProductName:  s.ProductName,
			Quantity:     s.Quantity,
			Revenue:      s.Revenue,
			MaterialCost: totalMaterialCost,
			GrossProfit:  grossProfit,
			GrossMargin:  grossMargin,
			UnitPrice:    unitPrice,
			UnitCost:     unitCost,
		})
	}

	return results, nil
}

func (s *CostService) GetProfitSummaryV2(query *dto.ProfitReportQueryDTO) (*dto.ProfitSummaryV2Response, error) {
	report, err := s.GetProfitReportV2(query)
	if err != nil {
		return nil, err
	}

	var totalRevenue, totalMaterialCost decimal.Decimal
	var productCount, orderCount int

	for _, r := range report {
		totalRevenue = totalRevenue.Add(r.Revenue)
		totalMaterialCost = totalMaterialCost.Add(r.MaterialCost)
		productCount++
	}

	if query.StoreID > 0 {
		type OrderCountResult struct {
			Count int64
		}
		var result OrderCountResult
		database.DB.Table("orders").
			Where("store_id = ? AND created_at >= ? AND created_at <= ?",
				query.StoreID, query.StartDate+" 00:00:00", query.EndDate+" 23:59:59").
			Where("order_status != ? AND pay_status = ?", -1, 1).
			Select("COUNT(DISTINCT id) as count").
			Scan(&result)
		orderCount = int(result.Count)
	}

	grossProfit := totalRevenue.Sub(totalMaterialCost)
	var grossMargin decimal.Decimal
	if totalRevenue.GreaterThan(decimal.Zero) {
		grossMargin = grossProfit.Div(totalRevenue).Mul(decimal.NewFromInt(100))
	}

	operatingExpenseRate := decimal.NewFromFloat(config.AppConfig.CostAlert.OperatingExpenseRate)
	if operatingExpenseRate.LessThanOrEqual(decimal.Zero) {
		operatingExpenseRate = decimal.NewFromFloat(15.0)
	}
	operatingExpense := totalRevenue.Mul(operatingExpenseRate).Div(decimal.NewFromInt(100))
	netProfit := grossProfit.Sub(operatingExpense)
	var netMargin decimal.Decimal
	if totalRevenue.GreaterThan(decimal.Zero) {
		netMargin = netProfit.Div(totalRevenue).Mul(decimal.NewFromInt(100))
	}

	return &dto.ProfitSummaryV2Response{
		TotalRevenue:        totalRevenue,
		TotalMaterialCost:   totalMaterialCost,
		GrossProfit:         grossProfit,
		GrossMargin:         grossMargin,
		OperatingExpense:    operatingExpense,
		OperatingExpenseRate: operatingExpenseRate,
		NetProfit:           netProfit,
		NetMargin:           netMargin,
		ProductCount:        productCount,
		OrderCount:          orderCount,
	}, nil
}
