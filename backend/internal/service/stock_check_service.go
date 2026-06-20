package service

import (
	"errors"
	"fmt"
	"math"
	"stalll-hub-pos/backend/internal/model"
	"stalll-hub-pos/backend/pkg/database"
	"time"

	"github.com/shopspring/decimal"
)

type StockCheckService struct {
	dingTalk *DingTalkService
}

func NewStockCheckService() *StockCheckService {
	return &StockCheckService{
		dingTalk: NewDingTalkService(),
	}
}

type StockCheckItemDTO struct {
	ProductID   uint    `json:"product_id"`
	SKUID       uint    `json:"sku_id"`
	SKUCode     string  `json:"sku_code"`
	ProductName string  `json:"product_name"`
	SpecName    string  `json:"spec_name"`
	ActualStock int     `json:"actual_stock"`
	SystemStock int     `json:"system_stock"`
	CostPrice   float64 `json:"cost_price"`
	Remark      string  `json:"remark"`
}

type CreateStockCheckReq struct {
	StoreID     uint                `json:"store_id"`
	Title       string              `json:"title"`
	CheckType   string              `json:"check_type"`
	CategoryIDs []uint              `json:"category_ids"`
	OperatorID  uint                `json:"operator_id"`
	OperatorName string             `json:"operator_name"`
	Remark      string              `json:"remark"`
}

func (s *StockCheckService) Create(req *CreateStockCheckReq) (*model.StockCheck, error) {
	if req.StoreID == 0 {
		return nil, errors.New("store_id required")
	}
	if req.Title == "" {
		return nil, errors.New("title required")
	}

	checkNo := generateCheckNo()
	now := time.Now()

	check := &model.StockCheck{
		StoreID:      req.StoreID,
		CheckNo:      checkNo,
		Title:        req.Title,
		CheckType:    req.CheckType,
		Status:       0,
		TotalSKU:     0,
		CheckedSKU:   0,
		OperatorID:   req.OperatorID,
		OperatorName: req.OperatorName,
		Remark:       req.Remark,
		StartTime:    &now,
	}

	tx := database.DB.Begin()
	if err := tx.Create(check).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	var skus []model.ProductSKU
	query := tx.Model(&model.ProductSKU{}).
		Where("store_id = ? AND status = 1", req.StoreID).
		Preload("Product").Preload("Product.Category")

	if len(req.CategoryIDs) > 0 {
		query = query.Joins("LEFT JOIN products ON products.id = product_skus.product_id").
			Where("products.category_id IN ?", req.CategoryIDs)
	}

	if err := query.Find(&skus).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	check.TotalSKU = len(skus)
	items := make([]model.StockCheckItem, 0, len(skus))

	for _, sku := range skus {
		costPrice, _ := sku.Price.Float64()
		categoryName := ""
		if sku.Product.Category.Name != "" {
			categoryName = sku.Product.Category.Name
		}

		items = append(items, model.StockCheckItem{
			CheckID:      check.ID,
			ProductID:    sku.ProductID,
			SKUID:        sku.ID,
			SKUCode:      sku.SKUCode,
			ProductName:  sku.Product.Name,
			SpecName:     sku.SpecName,
			CategoryID:   sku.Product.CategoryID,
			CategoryName: categoryName,
			SystemStock:  sku.Stock,
			ActualStock:  0,
			DiffQty:      -sku.Stock,
			CostPrice:    costPrice,
			DiffAmount:   -costPrice * float64(sku.Stock),
			Status:       0,
		})
	}

	if len(items) > 0 {
		if err := tx.Create(&items).Error; err != nil {
			tx.Rollback()
			return nil, err
		}
	}

	if err := tx.Save(check).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	tx.Commit()
	check.Items = items

	return check, nil
}

func (s *StockCheckService) GetByID(id uint) (*model.StockCheck, error) {
	var check model.StockCheck
	err := database.DB.Preload("Items").First(&check, id).Error
	return &check, err
}

func (s *StockCheckService) GetByCheckNo(checkNo string) (*model.StockCheck, error) {
	var check model.StockCheck
	err := database.DB.Where("check_no = ?", checkNo).Preload("Items").First(&check).Error
	return &check, err
}

func (s *StockCheckService) List(storeID uint, status int, page, pageSize int) ([]model.StockCheck, int64, error) {
	var list []model.StockCheck
	var total int64

	query := database.DB.Model(&model.StockCheck{})
	if storeID > 0 {
		query = query.Where("store_id = ?", storeID)
	}
	if status >= 0 {
		query = query.Where("status = ?", status)
	}

	query.Count(&total)

	err := query.Order("id DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&list).Error

	return list, total, err
}

func (s *StockCheckService) GetItems(checkID uint, status int, keyword string, page, pageSize int) ([]model.StockCheckItem, int64, error) {
	var items []model.StockCheckItem
	var total int64

	query := database.DB.Model(&model.StockCheckItem{}).Where("check_id = ?", checkID)
	if status >= 0 {
		query = query.Where("status = ?", status)
	}
	if keyword != "" {
		query = query.Where("sku_code LIKE ? OR product_name LIKE ? OR spec_name LIKE ?",
			"%"+keyword+"%", "%"+keyword+"%", "%"+keyword+"%")
	}

	query.Count(&total)

	err := query.Order("id ASC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&items).Error

	return items, total, err
}

func (s *StockCheckService) UploadItems(checkID uint, items []StockCheckItemDTO) error {
	var check model.StockCheck
	if err := database.DB.First(&check, checkID).Error; err != nil {
		return errors.New("check not found")
	}
	if check.Status == 2 {
		return errors.New("check already completed")
	}

	tx := database.DB.Begin()

	checkedCount := 0
	totalDiffQty := 0
	totalDiffAmount := 0.0

	for _, item := range items {
		if item.SKUID == 0 && item.SKUCode == "" {
			continue
		}

		var dbItem model.StockCheckItem
		q := tx.Where("check_id = ?", checkID)
		if item.SKUID > 0 {
			q = q.Where("sku_id = ?", item.SKUID)
		} else if item.SKUCode != "" {
			q = q.Where("sku_code = ?", item.SKUCode)
		}

		if err := q.First(&dbItem).Error; err != nil {
			continue
		}

		diffQty := item.ActualStock - dbItem.SystemStock
		diffAmount := float64(diffQty) * dbItem.CostPrice

		dbItem.ActualStock = item.ActualStock
		dbItem.DiffQty = diffQty
		dbItem.DiffAmount = diffAmount
		dbItem.Status = 1
		if item.Remark != "" {
			dbItem.Remark = item.Remark
		}

		if err := tx.Save(&dbItem).Error; err != nil {
			tx.Rollback()
			return err
		}

		checkedCount++
		totalDiffQty += diffQty
		totalDiffAmount += diffAmount
	}

	check.CheckedSKU = checkedCount
	check.TotalDiffQty = totalDiffQty
	check.TotalDiffAmount = math.Round(totalDiffAmount*100) / 100

	if err := tx.Save(&check).Error; err != nil {
		tx.Rollback()
		return err
	}

	tx.Commit()
	return nil
}

func (s *StockCheckService) UpdateItem(itemID uint, actualStock int, remark string) error {
	var item model.StockCheckItem
	if err := database.DB.First(&item, itemID).Error; err != nil {
		return err
	}

	diffQty := actualStock - item.SystemStock
	diffAmount := float64(diffQty) * item.CostPrice

	item.ActualStock = actualStock
	item.DiffQty = diffQty
	item.DiffAmount = diffAmount
	item.Status = 1
	if remark != "" {
		item.Remark = remark
	}

	tx := database.DB.Begin()

	if err := tx.Save(&item).Error; err != nil {
		tx.Rollback()
		return err
	}

	var checkedCount int64
	tx.Model(&model.StockCheckItem{}).
		Where("check_id = ? AND status = 1", item.CheckID).
		Count(&checkedCount)

	var totalDiffQty int
	var totalDiffAmount float64
	var allItems []model.StockCheckItem
	tx.Where("check_id = ?", item.CheckID).Find(&allItems)
	for _, it := range allItems {
		if it.Status == 1 {
			totalDiffQty += it.DiffQty
			totalDiffAmount += it.DiffAmount
		}
	}

	tx.Model(&model.StockCheck{}).
		Where("id = ?", item.CheckID).
		Updates(map[string]interface{}{
			"checked_sku":       checkedCount,
			"total_diff_qty":    totalDiffQty,
			"total_diff_amount": totalDiffAmount,
		})

	tx.Commit()
	return nil
}

func (s *StockCheckService) Complete(checkID uint) (*model.StockCheck, error) {
	var check model.StockCheck
	if err := database.DB.Preload("Items").First(&check, checkID).Error; err != nil {
		return nil, errors.New("check not found")
	}
	if check.Status == 2 {
		return &check, nil
	}

	tx := database.DB.Begin()

	now := time.Now()
	check.Status = 2
	check.EndTime = &now

	diffCount := 0
	for _, item := range check.Items {
		if item.Status == 1 && item.DiffQty != 0 {
			diffCount++
		}
	}

	var store model.Store
	tx.First(&store, check.StoreID)
	storeName := store.Name
	if storeName == "" {
		storeName = fmt.Sprintf("门店%d", check.StoreID)
	}

	if err := tx.Save(&check).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	for _, item := range check.Items {
		if item.Status != 1 || item.DiffQty == 0 {
			continue
		}
		tx.Model(&model.ProductSKU{}).
			Where("id = ?", item.SKUID).
			Update("stock", item.ActualStock)
	}

	tx.Commit()

	go s.dingTalk.SendStockCheckComplete(
		check.CheckNo,
		check.Title,
		check.TotalSKU,
		diffCount,
		check.TotalDiffAmount,
	)

	return &check, nil
}

func (s *StockCheckService) GenerateDiffReport(checkID uint) (map[string]interface{}, error) {
	var check model.StockCheck
	if err := database.DB.Preload("Items").First(&check, checkID).Error; err != nil {
		return nil, errors.New("check not found")
	}

	diffItems := make([]model.StockCheckItem, 0)
	profitItems := make([]model.StockCheckItem, 0)
	lossItems := make([]model.StockCheckItem, 0)
	normalItems := make([]model.StockCheckItem, 0)

	for _, item := range check.Items {
		if item.Status != 1 {
			continue
		}
		if item.DiffQty == 0 {
			normalItems = append(normalItems, item)
		} else {
			diffItems = append(diffItems, item)
			if item.DiffQty > 0 {
				profitItems = append(profitItems, item)
			} else {
				lossItems = append(lossItems, item)
			}
		}
	}

	result := map[string]interface{}{
		"check":          check,
		"diff_items":     diffItems,
		"profit_items":   profitItems,
		"loss_items":     lossItems,
		"normal_items":   normalItems,
		"diff_count":     len(diffItems),
		"profit_count":   len(profitItems),
		"loss_count":     len(lossItems),
		"normal_count":   len(normalItems),
		"total_diff_qty": check.TotalDiffQty,
		"total_diff_amount": check.TotalDiffAmount,
	}

	return result, nil
}

func generateCheckNo() string {
	now := time.Now()
	return fmt.Sprintf("PD%s%06d", now.Format("20060102150405"), time.Now().UnixNano()%1000000)
}
