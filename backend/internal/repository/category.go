package repository

import (
	"stalll-hub-pos/backend/internal/model"
	"stalll-hub-pos/backend/pkg/database"

	"gorm.io/gorm"
)

type CategoryRepository struct {
	db *gorm.DB
}

func NewCategoryRepository() *CategoryRepository {
	return &CategoryRepository{
		db: database.DB,
	}
}

func (r *CategoryRepository) Create(category *model.Category) error {
	return r.db.Create(category).Error
}

func (r *CategoryRepository) Update(category *model.Category) error {
	return r.db.Save(category).Error
}

func (r *CategoryRepository) Delete(id uint) error {
	return r.db.Delete(&model.Category{}, id).Error
}

func (r *CategoryRepository) GetByID(id uint) (*model.Category, error) {
	var category model.Category
	err := r.db.First(&category, id).Error
	return &category, err
}

func (r *CategoryRepository) List(storeID uint) ([]model.Category, error) {
	var categories []model.Category
	err := r.db.Where("store_id = ?", storeID).
		Preload("Products").
		Order("sort_order ASC, id ASC").
		Find(&categories).Error
	return categories, err
}

func (r *CategoryRepository) ListWithProductCount(storeID uint) ([]model.Category, error) {
	var categories []model.Category
	err := r.db.Where("store_id = ?", storeID).
		Select("categories.*, COUNT(products.id) as product_count").
		Joins("LEFT JOIN products ON products.category_id = categories.id AND products.status = 1").
		Group("categories.id").
		Order("sort_order ASC, id ASC").
		Find(&categories).Error
	return categories, err
}
