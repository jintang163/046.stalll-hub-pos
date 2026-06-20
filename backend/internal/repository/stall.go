package repository

import (
	"stalll-hub-pos/backend/internal/model"
	"stalll-hub-pos/backend/pkg/database"
	"time"

	"gorm.io/gorm"
)

type StallRepository struct {
	db *gorm.DB
}

func NewStallRepository(db *gorm.DB) *StallRepository {
	return &StallRepository{db: db}
}

func (r *StallRepository) Create(stall *model.Stall) error {
	return database.DB.Create(stall).Error
}

func (r *StallRepository) GetByID(id uint) (*model.Stall, error) {
	var stall model.Stall
	err := database.DB.Preload("Store").First(&stall, id).Error
	if err != nil {
		return nil, err
	}
	return &stall, nil
}

func (r *StallRepository) GetByStallNo(storeID uint, stallNo string) (*model.Stall, error) {
	var stall model.Stall
	err := database.DB.Where("store_id = ? AND stall_no = ?", storeID, stallNo).First(&stall).Error
	if err != nil {
		return nil, err
	}
	return &stall, nil
}

func (r *StallRepository) List(storeID uint, name string, status int, page, pageSize int) ([]model.Stall, int64, error {
	var stalls []model.Stall
	var total int64

	db := database.DB.Model(&model.Stall{})
	if storeID > 0 {
		db = db.Where("store_id = ?", storeID)
	}
	if name != "" {
		db = db.Where("name LIKE ?", "%"+name+"%")
	}
	if status > 0 {
		db = db.Where("status = ?", status)
	}

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	err := db.Preload("Store").Order("sort_order ASC, id DESC").Offset(offset).Limit(pageSize).Find(&stalls).Error
	return stalls, total, err
}

func (r *StallRepository) Update(stall *model.Stall) error {
	return database.DB.Save(stall).Error
}

func (r *StallRepository) Delete(id uint) error {
	return database.DB.Delete(&model.Stall{}, id).Error
}

func (r *StallRepository) GetAll(storeID uint) ([]model.Stall, error) {
	var stalls []model.Stall
	db := database.DB.Where("status = 1")
	if storeID > 0 {
		db = db.Where("store_id = ?", storeID)
	}
	err := db.Order("sort_order ASC, id ASC").Find(&stalls).Error
	return stalls, err
}

type StallDeviceRepository struct {
	db *gorm.DB
}

func NewStallDeviceRepository(db *gorm.DB) *StallDeviceRepository {
	return &StallDeviceRepository{db: db}
}

func (r *StallDeviceRepository) Create(device *model.StallDevice) error {
	return database.DB.Create(device).Error
}

func (r *StallDeviceRepository) GetByID(id uint) (*model.StallDevice, error) {
	var device model.StallDevice
	err := database.DB.Preload("Stall").Preload("Store").First(&device, id).Error
	if err != nil {
		return nil, err
	}
	return &device, nil
}

func (r *StallDeviceRepository) GetByDeviceID(deviceID string) (*model.StallDevice, error) {
	var device model.StallDevice
	err := database.DB.Where("device_id = ?", deviceID).Preload("Stall").Preload("Store").First(&device).Error
	if err != nil {
		return nil, err
	}
	return &device, nil
}

func (r *StallDeviceRepository) List(storeID, stallID uint, page, pageSize int) ([]model.StallDevice, int64, error) {
	var devices []model.StallDevice
	var total int64

	db := database.DB.Model(&model.StallDevice{})
	if storeID > 0 {
		db = db.Where("store_id = ?", storeID)
	}
	if stallID > 0 {
		db = db.Where("stall_id = ?", stallID)
	}

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	err := db.Preload("Stall").Preload("Store").Order("id DESC").Offset(offset).Limit(pageSize).Find(&devices).Error
	return devices, total, err
}

func (r *StallDeviceRepository) Update(device *model.StallDevice) error {
	return database.DB.Save(device).Error
}

func (r *StallDeviceRepository) UpdateHeartbeat(deviceID string) error {
	now := time.Now()
	return database.DB.Model(&model.StallDevice{}).
		Where("device_id = ?", deviceID).
		Updates(map[string]interface{}{
			"last_heartbeat_at": &now,
			"last_online_at":     &now,
		}).Error
}

func (r *StallDeviceRepository) Delete(id uint) error {
	return database.DB.Delete(&model.StallDevice{}, id).Error
}

func (r *StallDeviceRepository) GetByStallID(stallID uint) ([]model.StallDevice, error) {
	var devices []model.StallDevice
	err := database.DB.Where("stall_id = ? AND status = 1", stallID).Find(&devices).Error
	return devices, err
}

type StallUserRepository struct {
	db *gorm.DB
}

func NewStallUserRepository(db *gorm.DB) *StallUserRepository {
	return &StallUserRepository{db: db}
}

func (r *StallUserRepository) Create(user *model.StallUser) error {
	return database.DB.Create(user).Error
}

func (r *StallUserRepository) GetByID(id uint) (*model.StallUser, error) {
	var user model.StallUser
	err := database.DB.Preload("Stall").Preload("Store").First(&user, id).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *StallUserRepository) GetByUsername(username string) (*model.StallUser, error) {
	var user model.StallUser
	err := database.DB.Where("username = ?", username).Preload("Stall").Preload("Store").First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *StallUserRepository) List(storeID, stallID uint, username string, status int, page, pageSize int) ([]model.StallUser, int64, error) {
	var users []model.StallUser
	var total int64

	db := database.DB.Model(&model.StallUser{})
	if storeID > 0 {
		db = db.Where("store_id = ?", storeID)
	}
	if stallID > 0 {
		db = db.Where("stall_id = ?", stallID)
	}
	if username != "" {
		db = db.Where("username LIKE ?", "%"+username+"%")
	}
	if status > 0 {
		db = db.Where("status = ?", status)
	}

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	err := db.Preload("Stall").Preload("Store").Order("id DESC").Offset(offset).Limit(pageSize).Find(&users).Error
	return users, total, err
}

func (r *StallUserRepository) Update(user *model.StallUser) error {
	return database.DB.Save(user).Error
}

func (r *StallUserRepository) Delete(id uint) error {
	return database.DB.Delete(&model.StallUser{}, id).Error
}

type StallSettlementRepository struct {
	db *gorm.DB
}

func NewStallSettlementRepository(db *gorm.DB) *StallSettlementRepository {
	return &StallSettlementRepository{db: db}
}

func (r *StallSettlementRepository) Create(settlement *model.StallSettlement) error {
	return database.DB.Create(settlement).Error
}

func (r *StallSettlementRepository) GetByID(id uint) (*model.StallSettlement, error) {
	var settlement model.StallSettlement
	err := database.DB.Preload("Stall").Preload("Store").First(&settlement, id).Error
	if err != nil {
		return nil, err
	}
	return &settlement, nil
}

func (r *StallSettlementRepository) List(storeID, stallID uint, settlementDate string, settlementStatus int, page, pageSize int) ([]model.StallSettlement, int64, error) {
	var settlements []model.StallSettlement
	var total int64

	db := database.DB.Model(&model.StallSettlement{})
	if storeID > 0 {
		db = db.Where("store_id = ?", storeID)
	}
	if stallID > 0 {
		db = db.Where("stall_id = ?", stallID)
	}
	if settlementDate != "" {
		db = db.Where("settlement_date = ?", settlementDate)
	}
	if settlementStatus >= 0 {
		db = db.Where("settlement_status = ?", settlementStatus)
	}

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	err := db.Preload("Stall").Preload("Store").Order("id DESC").Offset(offset).Limit(pageSize).Find(&settlements).Error
	return settlements, total, err
}

func (r *StallSettlementRepository) Update(settlement *model.StallSettlement) error {
	return database.DB.Save(settlement).Error
}

func (r *StallSettlementRepository) Delete(id uint) error {
	return database.DB.Delete(&model.StallSettlement{}, id).Error
}

type StallDailyReportRepository struct {
	db *gorm.DB
}

func NewStallDailyReportRepository(db *gorm.DB) *StallDailyReportRepository {
	return &StallDailyReportRepository{db: db}
}

func (r *StallDailyReportRepository) Create(report *model.StallDailyReport) error {
	return database.DB.Create(report).Error
}

func (r *StallDailyReportRepository) GetByID(id uint) (*model.StallDailyReport, error) {
	var report model.StallDailyReport
	err := database.DB.Preload("Stall").Preload("Store").First(&report, id).Error
	if err != nil {
		return nil, err
	}
	return &report, nil
}

func (r *StallDailyReportRepository) List(storeID, stallID uint, startDate, endDate string) ([]model.StallDailyReport, error) {
	var reports []model.StallDailyReport
	db := database.DB.Model(&model.StallDailyReport{})
	if storeID > 0 {
		db = db.Where("store_id = ?", storeID)
	}
	if stallID > 0 {
		db = db.Where("stall_id = ?", stallID)
	}
	if startDate != "" {
		db = db.Where("report_date >= ?", startDate)
	}
	if endDate != "" {
		db = db.Where("report_date <= ?", endDate)
	}
	err := db.Preload("Stall").Preload("Store").Order("report_date DESC").Find(&reports).Error
	return reports, err
}

func (r *StallDailyReportRepository) Update(report *model.StallDailyReport) error {
	return database.DB.Save(report).Error
}

func (r *StallDailyReportRepository) GetByDate(storeID, stallID uint, reportDate string) (*model.StallDailyReport, error) {
	var report model.StallDailyReport
	err := database.DB.Where("store_id = ? AND stall_id = ? AND report_date = ?", storeID, stallID, reportDate).
		Preload("Stall").Preload("Store").First(&report).Error
	if err != nil {
		return nil, err
	}
	return &report, nil
}

func (r *StallDailyReportRepository) Upsert(report *model.StallDailyReport) error {
	return database.DB.Save(report).Error
}
