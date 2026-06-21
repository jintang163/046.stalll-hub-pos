package service

import (
	"fmt"
	"log"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
	"stalll-hub-pos/backend/internal/dto"
	"stalll-hub-pos/backend/internal/model"
	"stalll-hub-pos/backend/pkg/database"
)

type BOMService struct{}

func NewBOMService() *BOMService {
	return &BOMService{}
}

func (s *BOMService) GetIngredients(query *dto.IngredientQueryDTO) ([]model.Ingredient, int64, error) {
	var ingredients []model.Ingredient
	var total int64

	db := database.DB.Model(&model.Ingredient{})
	if query.StoreID > 0 {
		db = db.Where("store_id = ?", query.StoreID)
	}
	if query.Category != "" {
		db = db.Where("category = ?", query.Category)
	}
	if query.Keyword != "" {
		db = db.Where("name LIKE ? OR ingredient_no LIKE ?", "%"+query.Keyword+"%", "%"+query.Keyword+"%")
	}
	if query.Status >= 0 {
		db = db.Where("status = ?", query.Status)
	}

	db.Count(&total)

	if err := db.Order("id DESC").
		Offset((query.Page - 1) * query.PageSize).
		Limit(query.PageSize).
		Find(&ingredients).Error; err != nil {
		return nil, 0, err
	}

	return ingredients, total, nil
}

func (s *BOMService) GetIngredientByID(id uint) (*model.Ingredient, error) {
	var ingredient model.Ingredient
	if err := database.DB.First(&ingredient, id).Error; err != nil {
		return nil, fmt.Errorf("ingredient not found: %v", err)
	}
	return &ingredient, nil
}

func (s *BOMService) CreateIngredient(ingredient *model.Ingredient) error {
	return database.DB.Create(ingredient).Error
}

func (s *BOMService) UpdateIngredient(ingredient *model.Ingredient) error {
	return database.DB.Save(ingredient).Error
}

func (s *BOMService) DeleteIngredient(id uint) error {
	return database.DB.Delete(&model.Ingredient{}, id).Error
}

func (s *BOMService) GetProductBOM(productID, skuID uint) ([]model.ProductBOM, error) {
	var items []model.ProductBOM
	db := database.DB.Where("product_id = ?", productID)
	if skuID > 0 {
		db = db.Where("sku_id = ?", skuID)
	}
	if err := db.Preload("Ingredient").
		Order("sort_order ASC, id ASC").
		Find(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}

func (s *BOMService) AddBOMItem(item *model.ProductBOM) error {
	return database.DB.Create(item).Error
}

func (s *BOMService) UpdateBOMItem(item *model.ProductBOM) error {
	return database.DB.Save(item).Error
}

func (s *BOMService) DeleteBOMItem(id uint) error {
	return database.DB.Delete(&model.ProductBOM{}, id).Error
}

func (s *BOMService) CalculateProductCost(productID, skuID uint) (decimal.Decimal, error) {
	bomItems, err := s.GetProductBOM(productID, skuID)
	if err != nil {
		return decimal.Zero, err
	}

	if len(bomItems) == 0 {
		return decimal.Zero, nil
	}

	totalCost := decimal.Zero
	for _, item := range bomItems {
		var ingredient model.Ingredient
		if err := database.DB.First(&ingredient, item.IngredientID).Error; err != nil {
			log.Printf("[BOM] Ingredient %d not found, skip", item.IngredientID)
			continue
		}

		actualQty := item.Quantity
		if item.WastageRate.GreaterThan(decimal.Zero) {
			wastage := item.Quantity.Mul(item.WastageRate).Div(decimal.NewFromInt(100))
			actualQty = actualQty.Add(wastage)
		}

		itemCost := ingredient.CurrentPrice.Mul(actualQty)
		totalCost = totalCost.Add(itemCost)
	}

	return totalCost, nil
}

func (s *BOMService) BatchCalculateProductCosts(productIDs []uint) (map[uint]decimal.Decimal, error) {
	result := make(map[uint]decimal.Decimal)

	for _, productID := range productIDs {
		cost, err := s.CalculateProductCost(productID, 0)
		if err != nil {
			return nil, err
		}
		result[productID] = cost
	}

	return result, nil
}

func (s *BOMService) GetProductCostDetail(productID, skuID uint) (*dto.ProductCostDetailDTO, error) {
	bomItems, err := s.GetProductBOM(productID, skuID)
	if err != nil {
		return nil, err
	}

	var ingredientCost = make([]dto.BOMIngredientCostDTO, 0, len(bomItems))
	totalCost := decimal.Zero

	for _, item := range bomItems {
		var ingredient model.Ingredient
		if err := database.DB.First(&ingredient, item.IngredientID).Error; err != nil {
			continue
		}

		actualQty := item.Quantity
		if item.WastageRate.GreaterThan(decimal.Zero) {
			wastage := item.Quantity.Mul(item.WastageRate).Div(decimal.NewFromInt(100))
			actualQty = actualQty.Add(wastage)
		}

		itemCost := ingredient.CurrentPrice.Mul(actualQty)
		totalCost = totalCost.Add(itemCost)

		ingredientCost = append(ingredientCost, dto.BOMIngredientCostDTO{
			IngredientID:   item.IngredientID,
			IngredientName: item.IngredientName,
			Unit:           ingredient.Unit,
			UnitPrice:      ingredient.CurrentPrice,
			Quantity:       item.Quantity,
			WastageRate:    item.WastageRate,
			ActualQty:      actualQty,
			TotalCost:      itemCost,
		})
	}

	return &dto.ProductCostDetailDTO{
		ProductID:       productID,
		SKUID:           skuID,
		TotalCost:       totalCost,
		IngredientCost:  ingredientCost,
		IngredientCount: len(ingredientCost),
	}, nil
}

func (s *BOMService) GetIngredientCategories(storeID uint) ([]string, error) {
	var categories []string
	db := database.DB.Model(&model.Ingredient{})
	if storeID > 0 {
		db = db.Where("store_id = ?", storeID)
	}
	if err := db.Distinct("category").Where("category != ''").Pluck("category", &categories).Error; err != nil {
		return nil, err
	}
	return categories, nil
}

func (s *BOMService) SaveProductBOMList(storeID, productID uint, skuID uint, items []dto.BOMItemSaveDTO) error {
	return database.DB.Transaction(func(tx *gorm.DB) error {
		query := tx.Where("product_id = ? AND store_id = ?", productID, storeID)
		if skuID > 0 {
			query = query.Where("sku_id = ?", skuID)
		} else {
			query = query.Where("sku_id = 0 OR sku_id IS NULL")
		}
		if err := query.Delete(&model.ProductBOM{}).Error; err != nil {
			return err
		}

		for i, item := range items {
			bomItem := model.ProductBOM{
				StoreID:        storeID,
				ProductID:      productID,
				SKUID:          skuID,
				IngredientID:   item.IngredientID,
				IngredientName: item.IngredientName,
				Quantity:       item.Quantity,
				Unit:           item.Unit,
				WastageRate:    item.WastageRate,
				SortOrder:      i + 1,
			}
			if err := tx.Create(&bomItem).Error; err != nil {
				return err
			}
		}

		return nil
	})
}
