package service

import (
	"fmt"
	"log"
	"time"

	"github.com/shopspring/decimal"
	"github.com/xuri/excelize/v2"
	"stalll-hub-pos/backend/internal/dto"
	"stalll-hub-pos/backend/internal/model"
	"stalll-hub-pos/backend/pkg/database"
)

type PurchaseService struct {
	dingTalk *DingTalkService
}

func NewPurchaseService() *PurchaseService {
	return &PurchaseService{
		dingTalk: NewDingTalkService(),
	}
}

var purchaseStatusMap = map[int]string{
	0: "待发送",
	1: "已发送",
	2: "已确认",
	3: "部分入库",
	4: "已完成",
	5: "已取消",
}

func (s *PurchaseService) GeneratePurchaseOrder(req *dto.PurchaseOrderCreateRequest) (*model.PurchaseOrder, error) {
	purchaseNo := fmt.Sprintf("PO%s%06d", time.Now().Format("20060102"), time.Now().Unix()%1000000)

	totalAmount := decimal.Zero
	totalQty := decimal.Zero
	itemCount := 0

	var items []model.PurchaseOrderItem
	for idx, item := range req.Items {
		subtotal := item.PurchaseQty.Mul(item.UnitPrice)
		totalAmount = totalAmount.Add(subtotal)
		totalQty = totalQty.Add(item.PurchaseQty)
		itemCount++

		items = append(items, model.PurchaseOrderItem{
			IngredientID:   item.IngredientID,
			IngredientName: item.IngredientName,
			Category:       item.Category,
			Unit:           item.Unit,
			ForecastQty:    item.ForecastQty,
			SafetyStockQty: item.SafetyStockQty,
			CurrentStock:   item.CurrentStock,
			PurchaseQty:    item.PurchaseQty,
			UnitPrice:      item.UnitPrice,
			Subtotal:       subtotal,
			SortOrder:      idx,
		})
	}

	purchase := &model.PurchaseOrder{
		StoreID:       req.StoreID,
		PurchaseNo:    purchaseNo,
		SupplierName:  req.SupplierName,
		SupplierPhone: req.SupplierPhone,
		SupplierEmail: req.SupplierEmail,
		TotalAmount:   totalAmount,
		TotalQuantity: int(totalQty.IntPart()),
		ItemCount:     itemCount,
		Status:        0,
		ForecastDate:  req.ForecastDate,
		ForecastDays:  req.ForecastDays,
		Remark:        req.Remark,
		Items:         items,
	}

	if err := database.DB.Create(purchase).Error; err != nil {
		return nil, fmt.Errorf("create purchase order failed: %w", err)
	}

	log.Printf("[PurchaseService] Created purchase order %s, %d items, total: %s",
		purchaseNo, itemCount, totalAmount.String())

	return purchase, nil
}

func (s *PurchaseService) GetPurchaseOrder(id uint) (*model.PurchaseOrder, error) {
	var purchase model.PurchaseOrder
	if err := database.DB.Preload("Items").Preload("Store").First(&purchase, id).Error; err != nil {
		return nil, fmt.Errorf("purchase order not found: %w", err)
	}
	return &purchase, nil
}

func (s *PurchaseService) ListPurchaseOrders(query *dto.PurchaseOrderQuery) ([]model.PurchaseOrder, int64, error) {
	var orders []model.PurchaseOrder
	var total int64

	db := database.DB.Model(&model.PurchaseOrder{})

	if query.StoreID > 0 {
		db = db.Where("store_id = ?", query.StoreID)
	}
	if query.Status > 0 {
		db = db.Where("status = ?", query.Status)
	}
	if query.Date != "" {
		db = db.Where("forecast_date = ?", query.Date)
	}
	if query.Keyword != "" {
		db = db.Where("purchase_no LIKE ? OR supplier_name LIKE ?",
			"%"+query.Keyword+"%", "%"+query.Keyword+"%")
	}

	db.Count(&total)

	offset := (query.Page - 1) * query.Size
	if err := db.Preload("Store").
		Order("created_at DESC").
		Limit(query.Size).
		Offset(offset).
		Find(&orders).Error; err != nil {
		return nil, 0, err
	}

	return orders, total, nil
}

func (s *PurchaseService) UpdateStatus(id uint, status int) error {
	result := database.DB.Model(&model.PurchaseOrder{}).
		Where("id = ?", id).
		Update("status", status)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("purchase order not found")
	}

	if status == 1 {
		now := time.Now()
		database.DB.Model(&model.PurchaseOrder{}).
			Where("id = ?", id).
			Update("sent_at", &now)
	}

	return nil
}

func (s *PurchaseService) GenerateExcel(id uint) (string, error) {
	purchase, err := s.GetPurchaseOrder(id)
	if err != nil {
		return "", err
	}

	f := excelize.NewFile()
	defer f.Close()

	sheetName := "采购单"
	f.SetSheetName("Sheet1", sheetName)

	headerStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true, Size: 12},
		Fill: excelize.Fill{Type: "pattern", Color: []string{"#E0E0E0"}, Pattern: 1},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
		Border: []excelize.Border{
			{Type: "left", Color: "#000000", Style: 1},
			{Type: "top", Color: "#000000", Style: 1},
			{Type: "right", Color: "#000000", Style: 1},
			{Type: "bottom", Color: "#000000", Style: 1},
		},
	})

	titleStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true, Size: 16},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
	})

	infoStyle, _ := f.NewStyle(&excelize.Alignment{
		Horizontal: "left",
		Vertical:   "center",
	})

	f.MergeCell(sheetName, "A1", "H1")
	f.SetCellValue(sheetName, "A1", "采购订单")
	f.SetCellStyle(sheetName, "A1", "A1", titleStyle)
	f.SetRowHeight(sheetName, 1, 30)

	infoRow := 3
	f.SetCellValue(sheetName, fmt.Sprintf("A%d", infoRow), "采购单号:")
	f.SetCellValue(sheetName, fmt.Sprintf("B%d", infoRow), purchase.PurchaseNo)
	f.SetCellValue(sheetName, fmt.Sprintf("D%d", infoRow), "供应商:")
	f.SetCellValue(sheetName, fmt.Sprintf("E%d", infoRow), purchase.SupplierName)
	f.SetCellValue(sheetName, fmt.Sprintf("G%d", infoRow), "日期:")
	f.SetCellValue(sheetName, fmt.Sprintf("H%d", infoRow), purchase.CreatedAt.Format("2006-01-02"))

	f.SetCellValue(sheetName, fmt.Sprintf("A%d", infoRow+1), "联系电话:")
	f.SetCellValue(sheetName, fmt.Sprintf("B%d", infoRow+1), purchase.SupplierPhone)
	f.SetCellValue(sheetName, fmt.Sprintf("D%d", infoRow+1), "邮箱:")
	f.SetCellValue(sheetName, fmt.Sprintf("E%d", infoRow+1), purchase.SupplierEmail)
	f.SetCellValue(sheetName, fmt.Sprintf("G%d", infoRow+1), "状态:")
	f.SetCellValue(sheetName, fmt.Sprintf("H%d", infoRow+1), purchaseStatusMap[purchase.Status])
	f.SetCellStyle(sheetName, fmt.Sprintf("A%d", infoRow), fmt.Sprintf("H%d", infoRow+1), infoStyle)

	headerRow := infoRow + 3
	headers := []string{"序号", "食材名称", "分类", "单位", "预计用量", "安全库存", "当前库存", "采购数量", "单价", "金额", "备注"}
	for i, h := range headers {
		cell := fmt.Sprintf("%c%d", 'A'+i, headerRow)
		f.SetCellValue(sheetName, cell, h)
		f.SetCellStyle(sheetName, cell, cell, headerStyle)
	}
	f.SetRowHeight(sheetName, headerRow, 25)

	for i, item := range purchase.Items {
		row := headerRow + 1 + i
		f.SetCellValue(sheetName, fmt.Sprintf("A%d", row), i+1)
		f.SetCellValue(sheetName, fmt.Sprintf("B%d", row), item.IngredientName)
		f.SetCellValue(sheetName, fmt.Sprintf("C%d", row), item.Category)
		f.SetCellValue(sheetName, fmt.Sprintf("D%d", row), item.Unit)
		f.SetCellValue(sheetName, fmt.Sprintf("E%d", row), item.ForecastQty.String())
		f.SetCellValue(sheetName, fmt.Sprintf("F%d", row), item.SafetyStockQty.String())
		f.SetCellValue(sheetName, fmt.Sprintf("G%d", row), item.CurrentStock.String())
		f.SetCellValue(sheetName, fmt.Sprintf("H%d", row), item.PurchaseQty.String())
		f.SetCellValue(sheetName, fmt.Sprintf("I%d", row), item.UnitPrice.String())
		f.SetCellValue(sheetName, fmt.Sprintf("J%d", row), item.Subtotal.String())
		f.SetCellValue(sheetName, fmt.Sprintf("K%d", row), "")
	}

	totalRow := headerRow + 1 + len(purchase.Items)
	f.MergeCell(sheetName, fmt.Sprintf("A%d", totalRow), fmt.Sprintf("G%d", totalRow))
	f.SetCellValue(sheetName, fmt.Sprintf("A%d", totalRow), "合计")
	f.SetCellStyle(sheetName, fmt.Sprintf("A%d", totalRow), fmt.Sprintf("A%d", totalRow), headerStyle)
	f.SetCellValue(sheetName, fmt.Sprintf("H%d", totalRow), purchase.TotalQuantity)
	f.SetCellValue(sheetName, fmt.Sprintf("J%d", totalRow), purchase.TotalAmount.String())

	remarkRow := totalRow + 2
	f.SetCellValue(sheetName, fmt.Sprintf("A%d", remarkRow), "备注:")
	f.SetCellValue(sheetName, fmt.Sprintf("B%d", remarkRow), purchase.Remark)

	colWidths := map[string]float64{
		"A": 6, "B": 20, "C": 12, "D": 8, "E": 10,
		"F": 10, "G": 10, "H": 10, "I": 10, "J": 12, "K": 15,
	}
	for col, width := range colWidths {
		f.SetColWidth(sheetName, col, col, width)
	}

	fileName := fmt.Sprintf("purchase_%s_%d.xlsx", purchase.PurchaseNo, time.Now().Unix())
	filePath := fmt.Sprintf("./tmp/%s", fileName)

	if err := f.SaveAs(filePath); err != nil {
		return "", fmt.Errorf("save excel failed: %w", err)
	}

	log.Printf("[PurchaseService] Generated Excel for purchase order %s: %s", purchase.PurchaseNo, filePath)

	return filePath, nil
}

func (s *PurchaseService) SendToSupplier(id uint) error {
	purchase, err := s.GetPurchaseOrder(id)
	if err != nil {
		return err
	}

	if purchase.Status != 0 && purchase.Status != 5 {
		return fmt.Errorf("purchase order already sent or completed")
	}

	excelPath, err := s.GenerateExcel(id)
	if err != nil {
		return fmt.Errorf("generate excel failed: %w", err)
	}

	go s.dingTalk.SendPurchaseOrderNotification(purchase, excelPath)

	s.UpdateStatus(id, 1)

	log.Printf("[PurchaseService] Purchase order %s sent to supplier %s", purchase.PurchaseNo, purchase.SupplierName)

	return nil
}

func (s *PurchaseService) ConvertToResponse(purchase *model.PurchaseOrder) dto.PurchaseOrderResponse {
	resp := dto.PurchaseOrderResponse{
		ID:            purchase.ID,
		StoreID:       purchase.StoreID,
		PurchaseNo:    purchase.PurchaseNo,
		SupplierName:  purchase.SupplierName,
		SupplierPhone: purchase.SupplierPhone,
		SupplierEmail: purchase.SupplierEmail,
		TotalAmount:   purchase.TotalAmount,
		TotalQuantity: purchase.TotalQuantity,
		ItemCount:     purchase.ItemCount,
		Status:        purchase.Status,
		StatusText:    purchaseStatusMap[purchase.Status],
		ForecastDate:  purchase.ForecastDate,
		ForecastDays:  purchase.ForecastDays,
		Remark:        purchase.Remark,
		SentAt:        purchase.SentAt,
		CreatedAt:     purchase.CreatedAt,
	}

	if purchase.Store.ID > 0 {
		resp.StoreName = purchase.Store.Name
	}

	for _, item := range purchase.Items {
		resp.Items = append(resp.Items, dto.PurchaseItemResponse{
			ID:             item.ID,
			IngredientID:   item.IngredientID,
			IngredientName: item.IngredientName,
			Category:       item.Category,
			Unit:           item.Unit,
			ForecastQty:    item.ForecastQty,
			SafetyStockQty: item.SafetyStockQty,
			CurrentStock:   item.CurrentStock,
			PurchaseQty:    item.PurchaseQty,
			UnitPrice:      item.UnitPrice,
			Subtotal:       item.Subtotal,
		})
	}

	return resp
}

func (s *PurchaseService) AutoGenerateFromForecast(
	storeID uint,
	forecast *dto.StoreForecastResponse,
	suggestions *dto.StockingSuggestionResponse,
	supplierName string,
) (*model.PurchaseOrder, error) {
	var items []dto.PurchaseItemCreate

	for _, sug := range suggestions.Suggestions {
		if sug.SuggestedQty.LessThanOrEqual(decimal.Zero) {
			continue
		}

		items = append(items, dto.PurchaseItemCreate{
			IngredientID:   sug.IngredientID,
			IngredientName: sug.IngredientName,
			Category:       sug.Category,
			Unit:           sug.Unit,
			ForecastQty:    sug.ForecastUsage,
			SafetyStockQty: sug.SafetyStock,
			CurrentStock:   sug.CurrentStock,
			PurchaseQty:    sug.SuggestedQty,
			UnitPrice:      sug.UnitPrice,
		})
	}

	if len(items) == 0 {
		return nil, fmt.Errorf("no items need to purchase")
	}

	req := &dto.PurchaseOrderCreateRequest{
		StoreID:       storeID,
		ForecastDate:  forecast.ForecastDate,
		ForecastDays:  forecast.ForecastDays,
		SupplierName:  supplierName,
		Items:         items,
		Remark:        fmt.Sprintf("基于 %d 天销量预测自动生成", forecast.ForecastDays),
	}

	return s.GeneratePurchaseOrder(req)
}
