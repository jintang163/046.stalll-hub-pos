package repository

import (
	"stalll-hub-pos/backend/internal/model"
	"stalll-hub-pos/backend/pkg/database"

	"gorm.io/gorm"
)

type StoreRepository struct {
	db *gorm.DB
}

func NewStoreRepository(db *gorm.DB) *StoreRepository {
	return &StoreRepository{db: db}
}

func (r *StoreRepository) Create(store *model.Store) error {
	return database.DB.Create(store).Error
}

func (r *StoreRepository) GetByID(id uint) (*model.Store, error) {
	var store model.Store
	err := database.DB.First(&store, id).Error
	if err != nil {
		return nil, err
	}
	return &store, nil
}

func (r *StoreRepository) List(query string, status int, page, pageSize int) ([]model.Store, int64, error) {
	var stores []model.Store
	var total int64

	db := database.DB.Model(&model.Store{})
	if query != "" {
		db = db.Where("name LIKE ?", "%"+query+"%")
	}
	if status > 0 {
		db = db.Where("status = ?", status)
	}

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	err := db.Order("id DESC").Offset(offset).Limit(pageSize).Find(&stores).Error
	return stores, total, err
}

func (r *StoreRepository) Update(store *model.Store) error {
	return database.DB.Save(store).Error
}

func (r *StoreRepository) Delete(id uint) error {
	return database.DB.Delete(&model.Store{}, id).Error
}

func (r *StoreRepository) GetAll() ([]model.Store, error) {
	var stores []model.Store
	err := database.DB.Where("status = 1").Order("id ASC").Find(&stores).Error
	return stores, err
}

type PrinterRepository struct {
	db *gorm.DB
}

func NewPrinterRepository(db *gorm.DB) *PrinterRepository {
	return &PrinterRepository{db: db}
}

func (r *PrinterRepository) Create(printer *model.Printer) error {
	return database.DB.Create(printer).Error
}

func (r *PrinterRepository) GetByID(id uint) (*model.Printer, error) {
	var printer model.Printer
	err := database.DB.Preload("Store").First(&printer, id).Error
	if err != nil {
		return nil, err
	}
	return &printer, nil
}

func (r *PrinterRepository) List(storeID uint, printerType string, status int, page, pageSize int) ([]model.Printer, int64, error) {
	var printers []model.Printer
	var total int64

	db := database.DB.Model(&model.Printer{})
	if storeID > 0 {
		db = db.Where("store_id = ?", storeID)
	}
	if printerType != "" {
		db = db.Where("type = ?", printerType)
	}
	if status > 0 {
		db = db.Where("status = ?", status)
	}

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	err := db.Preload("Store").Order("id DESC").Offset(offset).Limit(pageSize).Find(&printers).Error
	return printers, total, err
}

func (r *PrinterRepository) Update(printer *model.Printer) error {
	return database.DB.Save(printer).Error
}

func (r *PrinterRepository) Delete(id uint) error {
	return database.DB.Delete(&model.Printer{}, id).Error
}

func (r *PrinterRepository) GetByStoreAndType(storeID uint, printerType string) ([]model.Printer, error) {
	var printers []model.Printer
	err := database.DB.Where("store_id = ? AND type = ? AND status = 1", storeID, printerType).Find(&printers).Error
	return printers, err
}

func (r *PrinterRepository) GetByStore(storeID uint) ([]model.Printer, error) {
	var printers []model.Printer
	err := database.DB.Where("store_id = ? AND status = 1", storeID).Find(&printers).Error
	return printers, err
}
