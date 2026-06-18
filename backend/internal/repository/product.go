package repository

import (
	"stalll-hub-pos/backend/internal/model"
	"stalll-hub-pos/backend/pkg/database"

	"gorm.io/gorm"
)

type ProductRepository struct {
	db *gorm.DB
}

func NewProductRepository() *ProductRepository {
	return &ProductRepository{
		db: database.DB,
	}
}

func (r *ProductRepository) Create(product *model.Product) error {
	return r.db.Create(product).Error
}

func (r *ProductRepository) Update(product *model.Product) error {
	return r.db.Save(product).Error
}

func (r *ProductRepository) Delete(id uint) error {
	return r.db.Delete(&model.Product{}, id).Error
}

func (r *ProductRepository) GetByID(id uint) (*model.Product, error) {
	var product model.Product
	err := r.db.Preload("Category").Preload("SKUs").Preload("SKUs.AttributeValues").
		Preload("Attributes").Preload("Attributes.Values").
		First(&product, id).Error
	return &product, err
}

func (r *ProductRepository) List(storeID, categoryID uint, name string, status *int, isHot, isRecommend *bool, offset, limit int) ([]model.Product, int64, error) {
	var products []model.Product
	var total int64

	query := r.db.Model(&model.Product{}).Where("store_id = ?", storeID)

	if categoryID > 0 {
		query = query.Where("category_id = ?", categoryID)
	}
	if name != "" {
		query = query.Where("name LIKE ?", "%"+name+"%")
	}
	if status != nil {
		query = query.Where("status = ?", *status)
	}
	if isHot != nil {
		query = query.Where("is_hot = ?", *isHot)
	}
	if isRecommend != nil {
		query = query.Where("is_recommend = ?", *isRecommend)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := query.Preload("Category").Preload("SKUs").Preload("Attributes").
		Order("sort_order ASC, id DESC").
		Offset(offset).Limit(limit).Find(&products).Error

	return products, total, err
}

func (r *ProductRepository) CreateSKU(sku *model.ProductSKU) error {
	return r.db.Create(sku).Error
}

func (r *ProductRepository) UpdateSKU(sku *model.ProductSKU) error {
	return r.db.Save(sku).Error
}

func (r *ProductRepository) DeleteSKU(id uint) error {
	return r.db.Delete(&model.ProductSKU{}, id).Error
}

func (r *ProductRepository) GetSKUByID(id uint) (*model.ProductSKU, error) {
	var sku model.ProductSKU
	err := r.db.Preload("AttributeValues").First(&sku, id).Error
	return &sku, err
}

func (r *ProductRepository) GetSKUsByProductID(productID uint) ([]model.ProductSKU, error) {
	var skus []model.ProductSKU
	err := r.db.Where("product_id = ?", productID).Preload("AttributeValues").Find(&skus).Error
	return skus, err
}

func (r *ProductRepository) UpdateSKUStock(skuID uint, stock int) (int, error) {
	var sku model.ProductSKU
	if err := r.db.First(&sku, skuID).Error; err != nil {
		return 0, err
	}
	oldStock := sku.Stock
	err := r.db.Model(&sku).Update("stock", stock).Error
	return oldStock, err
}

func (r *ProductRepository) DecreaseStockWithOptimisticLock(skuID uint, quantity int) (int, int, error) {
	var sku model.ProductSKU
	if err := r.db.First(&sku, skuID).Error; err != nil {
		return 0, 0, err
	}

	oldStock := sku.Stock
	if oldStock < quantity {
		return oldStock, oldStock, nil
	}

	newStock := oldStock - quantity
	result := r.db.Model(&model.ProductSKU{}).
		Where("id = ? AND stock = ?", skuID, oldStock).
		Update("stock", newStock)

	if result.Error != nil {
		return oldStock, oldStock, result.Error
	}

	if result.RowsAffected == 0 {
		return oldStock, oldStock, nil
	}

	soldCount := oldStock - newStock
	if soldCount > 0 {
		r.db.Model(&model.ProductSKU{}).
			Where("id = ?", skuID).
			UpdateColumn("sold_count", gorm.Expr("sold_count + ?", soldCount))
	}

	return oldStock, newStock, nil
}

func (r *ProductRepository) CreateAttribute(attr *model.ProductAttribute) error {
	return r.db.Create(attr).Error
}

func (r *ProductRepository) UpdateAttribute(attr *model.ProductAttribute) error {
	return r.db.Save(attr).Error
}

func (r *ProductRepository) DeleteAttribute(id uint) error {
	return r.db.Delete(&model.ProductAttribute{}, id).Error
}

func (r *ProductRepository) CreateAttributeValue(value *model.AttributeValue) error {
	return r.db.Create(value).Error
}

func (r *ProductRepository) UpdateAttributeValue(value *model.AttributeValue) error {
	return r.db.Save(value).Error
}

func (r *ProductRepository) DeleteAttributeValue(id uint) error {
	return r.db.Delete(&model.AttributeValue{}, id).Error
}

func (r *ProductRepository) GetAttributesByProductID(productID uint) ([]model.ProductAttribute, error) {
	var attrs []model.ProductAttribute
	err := r.db.Where("product_id = ?", productID).Preload("Values").Find(&attrs).Error
	return attrs, err
}

func (r *ProductRepository) CreateSKUAttributeValue(sav *model.SKUAttributeValue) error {
	return r.db.Create(sav).Error
}

func (r *ProductRepository) DeleteSKUAttributeValues(skuID uint) error {
	return r.db.Where("sku_id = ?", skuID).Delete(&model.SKUAttributeValue{}).Error
}

func (r *ProductRepository) SyncProducts(storeID uint, lastSyncID uint, limit int) ([]model.Product, uint, int64, error) {
	var products []model.Product
	var total int64

	query := r.db.Model(&model.Product{}).Where("store_id = ? AND id > ?", storeID, lastSyncID)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, 0, err
	}

	err := query.Preload("Category").Preload("SKUs").Preload("SKUs.AttributeValues").
		Preload("Attributes").Preload("Attributes.Values").
		Order("id ASC").
		Limit(limit).Find(&products).Error

	var lastID uint = 0
	if len(products) > 0 {
		lastID = products[len(products)-1].ID
	}

	return products, lastID, total, err
}

func (r *ProductRepository) BatchUpdatePrice(productIDs []uint, priceType string, price float64) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if priceType == "fixed" {
			return tx.Model(&model.ProductSKU{}).
				Where("product_id IN ?", productIDs).
				Update("price", price).Error
		} else {
			return tx.Model(&model.ProductSKU{}).
				Where("product_id IN ?", productIDs).
				Update("price", gorm.Expr("price * ?", price/100)).Error
		}
	})
}

func (r *ProductRepository) GetStockWarnings(storeID uint) ([]model.StockWarning, int64, error) {
	var warnings []model.StockWarning
	var total int64

	query := r.db.Model(&model.StockWarning{}).
		Where("store_id = ? AND status = 0", storeID).
		Preload("Product").Preload("SKU")

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := query.Order("id DESC").Find(&warnings).Error
	return warnings, total, err
}

func (r *ProductRepository) CheckStockWarning(storeID, skuID, productID uint, stock, threshold int) error {
	var count int64
	r.db.Model(&model.StockWarning{}).
		Where("store_id = ? AND sku_id = ? AND status = 0", storeID, skuID).
		Count(&count)

	if stock <= threshold && count == 0 {
		warning := &model.StockWarning{
			StoreID:      storeID,
			SKUID:        skuID,
			ProductID:    productID,
			CurrentStock: stock,
			Threshold:    threshold,
			Status:       0,
		}
		return r.db.Create(warning).Error
	} else if stock > threshold && count > 0 {
		return r.db.Where("store_id = ? AND sku_id = ? AND status = 0", storeID, skuID).
			Delete(&model.StockWarning{}).Error
	}
	return nil
}

func (r *ProductRepository) GetByIDs(ids []uint) ([]model.Product, error) {
	var products []model.Product
	err := r.db.Where("id IN ?", ids).
		Preload("SKUs").Preload("Attributes").Preload("Attributes.Values").
		Find(&products).Error
	return products, err
}
