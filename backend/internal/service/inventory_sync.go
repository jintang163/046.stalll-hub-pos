package service

import (
	"fmt"
	"log"
	"time"

	"github.com/shopspring/decimal"
	"stalll-hub-pos/backend/config"
	"stalll-hub-pos/backend/internal/model"
	"stalll-hub-pos/backend/pkg/database"
)

type InventorySyncService struct {
	client       *InventoryClient
	alertService *CostAlertService
}

func NewInventorySyncService() *InventorySyncService {
	return &InventorySyncService{
		client:       NewInventoryClient(),
		alertService: NewCostAlertService(),
	}
}

func (s *InventorySyncService) StartSyncScheduler() {
	if !config.AppConfig.Inventory.Enabled {
		log.Println("[InventorySync] Inventory sync disabled, skipping scheduler")
		return
	}

	interval := config.AppConfig.Inventory.SyncInterval
	if interval <= 0 {
		interval = 60
	}

	go s.runSyncScheduler(time.Duration(interval) * time.Minute)

	log.Printf("[InventorySync] Scheduler started, interval: %d minutes", interval)

	go func() {
		time.Sleep(10 * time.Second)
		s.SyncAllStores()
	}()
}

func (s *InventorySyncService) runSyncScheduler(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for range ticker.C {
		s.SyncAllStores()
	}
}

func (s *InventorySyncService) SyncAllStores() {
	var stores []model.Store
	if err := database.DB.Where("status = ?", 1).Find(&stores).Error; err != nil {
		log.Printf("[InventorySync] Failed to get stores: %v", err)
		return
	}

	for _, store := range stores {
		if err := s.SyncStoreIngredients(store.ID); err != nil {
			log.Printf("[InventorySync] Store %d sync failed: %v", store.ID, err)
		}
	}
}

func (s *InventorySyncService) SyncStoreIngredients(storeID uint) error {
	log.Printf("[InventorySync] Start syncing ingredients for store %d", storeID)

	ingredients, err := s.client.GetAllIngredients(storeID)
	if err != nil {
		return fmt.Errorf("failed to fetch ingredients: %v", err)
	}

	batchNo := fmt.Sprintf("INV%s%d", time.Now().Format("20060102150405"), time.Now().UnixNano()%10000)
	effectiveDate := time.Now().Format("2006-01-02")

	var successCount, failCount int
	var updatedIngredients []model.Ingredient

	for _, invIng := range ingredients {
		if err := s.syncSingleIngredient(storeID, invIng, batchNo, effectiveDate, &updatedIngredients); err != nil {
			log.Printf("[InventorySync] Failed to sync ingredient %s: %v", invIng.IngredientNo, err)
			failCount++
			continue
		}
		successCount++
	}

	if len(updatedIngredients) > 0 && config.AppConfig.CostAlert.Enabled {
		go s.alertService.DetectAndAlert(storeID, updatedIngredients, batchNo)
	}

	log.Printf("[InventorySync] Store %d sync completed: success=%d, fail=%d, batch=%s",
		storeID, successCount, failCount, batchNo)
	return nil
}

func (s *InventorySyncService) syncSingleIngredient(storeID uint, invIng InventoryIngredient,
	batchNo, effectiveDate string, updatedIngredients *[]model.Ingredient) error {

	var ingredient model.Ingredient
	err := database.DB.Where("store_id = ? AND ingredient_no = ?", storeID, invIng.IngredientNo).First(&ingredient).Error

	if err != nil {
		ingredient = model.Ingredient{
			StoreID:      storeID,
			IngredientNo: invIng.IngredientNo,
			Name:         invIng.Name,
			Category:     invIng.Category,
			Unit:         invIng.Unit,
			CurrentPrice: invIng.Price,
			CurrentStock: invIng.CurrentStock,
			StockUnit:    invIng.StockUnit,
			Supplier:     invIng.Supplier,
			SupplierPhone: invIng.SupplierPhone,
			SupplierEmail: invIng.SupplierEmail,
			Status:       invIng.Status,
		}
		if err := database.DB.Create(&ingredient).Error; err != nil {
			return fmt.Errorf("create ingredient failed: %v", err)
		}
	} else {
		oldPrice := ingredient.CurrentPrice
		ingredient.Name = invIng.Name
		ingredient.Category = invIng.Category
		ingredient.Unit = invIng.Unit
		ingredient.CurrentPrice = invIng.Price
		ingredient.CurrentStock = invIng.CurrentStock
		ingredient.StockUnit = invIng.StockUnit
		ingredient.Supplier = invIng.Supplier
		ingredient.SupplierPhone = invIng.SupplierPhone
		ingredient.SupplierEmail = invIng.SupplierEmail
		ingredient.Status = invIng.Status

		if err := database.DB.Save(&ingredient).Error; err != nil {
			return fmt.Errorf("update ingredient failed: %v", err)
		}

		if !oldPrice.Equal(invIng.Price) {
			*updatedIngredients = append(*updatedIngredients, ingredient)
		}
	}

	previousPrice := decimal.Zero
	var lastPrice model.IngredientPrice
	if err := database.DB.Where("store_id = ? AND ingredient_id = ?", storeID, ingredient.ID).
		Order("id DESC").First(&lastPrice).Error; err == nil {
		previousPrice = lastPrice.Price
	}

	priceChange := decimal.Zero
	if previousPrice.GreaterThan(decimal.Zero) {
		priceChange = invIng.Price.Sub(previousPrice).Div(previousPrice).Mul(decimal.NewFromInt(100))
	}

	priceRecord := model.IngredientPrice{
		StoreID:        storeID,
		IngredientID:   ingredient.ID,
		IngredientNo:   invIng.IngredientNo,
		IngredientName: invIng.Name,
		Price:          invIng.Price,
		PreviousPrice:  previousPrice,
		PriceChange:    priceChange,
		Supplier:       invIng.Supplier,
		EffectiveDate:  effectiveDate,
		Source:         "inventory",
		BatchNo:        batchNo,
	}
	if err := database.DB.Create(&priceRecord).Error; err != nil {
		return fmt.Errorf("create price record failed: %v", err)
	}

	return nil
}
