package repository

import (
	"errors"
	"fmt"
	"time"

	"stalll-hub-pos/backend/internal/dto"
	"stalll-hub-pos/backend/internal/model"
	"stalll-hub-pos/backend/pkg/database"

	"gorm.io/gorm"
)

type TableRepository struct{}

func NewTableRepository() *TableRepository {
	return &TableRepository{}
}

func (r *TableRepository) Create(table *model.Table) error {
	return database.DB.Transaction(func(tx *gorm.DB) error {
		var existing int64
		if err := tx.Model(&model.Table{}).Where("store_id = ? AND table_no = ?", table.StoreID, table.TableNo).Count(&existing).Error; err != nil {
			return err
		}
		if existing > 0 {
			return errors.New("桌号已存在")
		}
		if err := tx.Create(table).Error; err != nil {
			return err
		}
		return tx.Model(&model.Store{}).Where("id = ?", table.StoreID).UpdateColumn("tables_count", gorm.Expr("tables_count + 1")).Error
	})
}

func (r *TableRepository) Update(id uint, table *model.Table) error {
	return database.DB.Model(&model.Table{}).Where("id = ?", id).Updates(table).Error
}

func (r *TableRepository) Delete(id uint) error {
	return database.DB.Transaction(func(tx *gorm.DB) error {
		var table model.Table
		if err := tx.First(&table, id).Error; err != nil {
			return err
		}
		if err := tx.Delete(&model.Table{}, id).Error; err != nil {
			return err
		}
		return tx.Model(&model.Store{}).Where("id = ?", table.StoreID).UpdateColumn("tables_count", gorm.Expr("tables_count - 1")).Error
	})
}

func (r *TableRepository) GetByID(id uint) (*model.Table, error) {
	var table model.Table
	err := database.DB.Preload("Store").First(&table, id).Error
	return &table, err
}

func (r *TableRepository) GetByStoreAndNo(storeID uint, tableNo string) (*model.Table, error) {
	var table model.Table
	err := database.DB.Where("store_id = ? AND table_no = ?", storeID, tableNo).First(&table).Error
	return &table, err
}

func (r *TableRepository) List(query *dto.TableQueryDTO) ([]model.Table, int64, error) {
	var tables []model.Table
	var total int64
	db := database.DB.Model(&model.Table{})

	if query.StoreID > 0 {
		db = db.Where("store_id = ?", query.StoreID)
	}
	if query.Status > 0 {
		db = db.Where("status = ?", query.Status)
	}
	if query.Floor > 0 {
		db = db.Where("floor = ?", query.Floor)
	}
	if query.Area != "" {
		db = db.Where("area = ?", query.Area)
	}
	if query.Type != "" {
		db = db.Where("type = ?", query.Type)
	}
	if query.Keyword != "" {
		db = db.Where("table_no LIKE ? OR name LIKE ?", "%"+query.Keyword+"%", "%"+query.Keyword+"%")
	}

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (query.PageNum - 1) * query.PageSize
	err := db.Preload("Store").Order("floor ASC, area ASC, table_no ASC").
		Offset(offset).Limit(query.PageSize).Find(&tables).Error

	return tables, total, err
}

func (r *TableRepository) GetOccupiedTables(storeID uint) ([]model.Table, error) {
	var tables []model.Table
	err := database.DB.Where("store_id = ? AND current_order_id IS NOT NULL AND current_order_id > 0", storeID).
		Order("checkin_time ASC").Find(&tables).Error
	return tables, err
}

func (r *TableRepository) Checkin(tableID uint, peopleCount int, orderID uint) error {
	now := time.Now()
	return database.DB.Model(&model.Table{}).Where("id = ?", tableID).Updates(map[string]interface{}{
		"current_order_id":        orderID,
		"current_customer_count":  peopleCount,
		"checkin_time":            &now,
	}).Error
}

func (r *TableRepository) Checkout(tableID uint, orderAmount float64) error {
	var table model.Table
	if err := database.DB.First(&table, tableID).Error; err != nil {
		return err
	}

	duration := 0
	if table.CheckinTime != nil {
		duration = int(time.Since(*table.CheckinTime).Minutes())
	}

	return database.DB.Model(&model.Table{}).Where("id = ?", tableID).Updates(map[string]interface{}{
		"current_order_id":        nil,
		"current_customer_count":  0,
		"checkin_time":            nil,
		"occupied_duration":       gorm.Expr("occupied_duration + ?", duration),
		"total_orders":            gorm.Expr("total_orders + 1"),
		"total_amount":            gorm.Expr("total_amount + ?", orderAmount),
	}).Error
}

func (r *TableRepository) UpdateQRCode(tableID uint, qrCode, qrCodeUrl string) error {
	return database.DB.Model(&model.Table{}).Where("id = ?", tableID).Updates(map[string]interface{}{
		"qr_code":     qrCode,
		"qr_code_url": qrCodeUrl,
	}).Error
}

func (r *TableRepository) BatchCreate(tables []model.Table) error {
	return database.DB.Create(&tables).Error
}

func (r *TableRepository) GetAvailableTables(storeID uint, peopleCount int) ([]model.Table, error) {
	var tables []model.Table
	err := database.DB.Where("store_id = ? AND status = 1 AND capacity >= ? AND (current_order_id IS NULL OR current_order_id = 0)",
		storeID, peopleCount).Order("capacity ASC, table_no ASC").Find(&tables).Error
	return tables, err
}

func (r *TableRepository) GetStoreMapInfo() ([]model.Store, error) {
	var stores []model.Store
	err := database.DB.Where("latitude IS NOT NULL AND longitude IS NOT NULL AND latitude != '' AND longitude != ''").
		Select("id, name, address, latitude, longitude, status, tables_count, open_time, close_time, phone").
		Find(&stores).Error
	return stores, err
}

type TableAreaRepository struct{}

func NewTableAreaRepository() *TableAreaRepository {
	return &TableAreaRepository{}
}

func (r *TableAreaRepository) Create(area *model.TableArea) error {
	return database.DB.Create(area).Error
}

func (r *TableAreaRepository) Update(id uint, area *model.TableArea) error {
	return database.DB.Model(&model.TableArea{}).Where("id = ?", id).Updates(area).Error
}

func (r *TableAreaRepository) Delete(id uint) error {
	return database.DB.Delete(&model.TableArea{}, id).Error
}

func (r *TableAreaRepository) GetByID(id uint) (*model.TableArea, error) {
	var area model.TableArea
	err := database.DB.First(&area, id).Error
	return &area, err
}

func (r *TableAreaRepository) List(storeID uint) ([]model.TableArea, error) {
	var areas []model.TableArea
	err := database.DB.Where("store_id = ?", storeID).Order("sort_order ASC, id ASC").Find(&areas).Error
	return areas, err
}

type ReservationRepository struct{}

func NewReservationRepository() *ReservationRepository {
	return &ReservationRepository{}
}

func (r *ReservationRepository) Create(reservation *model.TableReservation) error {
	return database.DB.Create(reservation).Error
}

func (r *ReservationRepository) Update(id uint, reservation *model.TableReservation) error {
	return database.DB.Model(&model.TableReservation{}).Where("id = ?", id).Updates(reservation).Error
}

func (r *ReservationRepository) Delete(id uint) error {
	return database.DB.Delete(&model.TableReservation{}, id).Error
}

func (r *ReservationRepository) GetByID(id uint) (*model.TableReservation, error) {
	var reservation model.TableReservation
	err := database.DB.Preload("Table").Preload("Member").First(&reservation, id).Error
	return &reservation, err
}

func (r *ReservationRepository) List(query *dto.ReservationQueryDTO) ([]model.TableReservation, int64, error) {
	var reservations []model.TableReservation
	var total int64
	db := database.DB.Model(&model.TableReservation{})

	if query.StoreID > 0 {
		db = db.Where("store_id = ?", query.StoreID)
	}
	if query.MemberID > 0 {
		db = db.Where("member_id = ?", query.MemberID)
	}
	if query.Status > 0 {
		db = db.Where("status = ?", query.Status)
	}
	if query.ReserveDate != "" {
		db = db.Where("reserve_date = ?", query.ReserveDate)
	}
	if query.CheckinStatus >= 0 {
		db = db.Where("checkin_status = ?", query.CheckinStatus)
	}
	if query.Keyword != "" {
		db = db.Where("member_name LIKE ? OR member_phone LIKE ? OR table_no LIKE ?",
			"%"+query.Keyword+"%", "%"+query.Keyword+"%", "%"+query.Keyword+"%")
	}

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (query.PageNum - 1) * query.PageSize
	err := db.Preload("Table").Preload("Member").Order("reserve_date DESC, reserve_time DESC, id DESC").
		Offset(offset).Limit(query.PageSize).Find(&reservations).Error

	return reservations, total, err
}

func (r *ReservationRepository) Checkin(id uint) error {
	now := time.Now()
	return database.DB.Model(&model.TableReservation{}).Where("id = ?", id).Updates(map[string]interface{}{
		"checkin_status": 1,
		"checkin_time":   &now,
		"status":         2,
	}).Error
}

func (r *ReservationRepository) Cancel(id uint) error {
	now := time.Now()
	return database.DB.Model(&model.TableReservation{}).Where("id = ?", id).Updates(map[string]interface{}{
		"status":      3,
		"cancel_time": &now,
	}).Error
}

func (r *ReservationRepository) GetTimeSlotAvailability(storeID uint, reserveDate string, peopleCount int) (map[string]int, error) {
	type Result struct {
		ReserveTime string
		Count       int
	}
	var results []Result

	err := database.DB.Table("table_reservations").
		Select("reserve_time, COUNT(*) as count").
		Where("store_id = ? AND reserve_date = ? AND status IN (1,2) AND people_count <= ? + 2",
			storeID, reserveDate, peopleCount).
		Group("reserve_time").Scan(&results).Error

	if err != nil {
		return nil, err
	}

	availability := make(map[string]int)
	for _, r := range results {
		availability[r.ReserveTime] = r.Count
	}

	return availability, nil
}

func (r *ReservationRepository) GetCountByTableAndTime(tableID uint, reserveDate, reserveTime string) (int64, error) {
	var count int64
	err := database.DB.Model(&model.TableReservation{}).
		Where("table_id = ? AND reserve_date = ? AND reserve_time = ? AND status IN (1,2)",
			tableID, reserveDate, reserveTime).
		Count(&count).Error
	return count, err
}

type QueueRepository struct{}

func NewQueueRepository() *QueueRepository {
	return &QueueRepository{}
}

func (r *QueueRepository) Create(queue *model.QueueNumber, config *model.QueueConfig) error {
	return database.DB.Transaction(func(tx *gorm.DB) error {
		today := time.Now().Format("2006-01-02")

		var lastSeq int
		err := tx.Model(&model.QueueNumber{}).
			Where("store_id = ? AND queue_type = ? AND DATE(created_at) = ?",
				queue.StoreID, queue.QueueType, today).
			Select("COALESCE(MAX(sequence), 0)").Scan(&lastSeq).Error
		if err != nil {
			return err
		}

		queue.Sequence = lastSeq + 1

		var prefix string
		switch queue.QueueType {
		case "small":
			prefix = config.SmallPrefix
		case "medium":
			prefix = config.MediumPrefix
		case "large":
			prefix = config.LargePrefix
		}
		queue.QueueNumber = fmt.Sprintf("%s%03d", prefix, queue.Sequence)

		return tx.Create(queue).Error
	})
}

func (r *QueueRepository) Call(queueID uint) error {
	now := time.Now()
	return database.DB.Transaction(func(tx *gorm.DB) error {
		var queue model.QueueNumber
		if err := tx.First(&queue, queueID).Error; err != nil {
			return err
		}

		updates := map[string]interface{}{
			"call_count":     gorm.Expr("call_count + 1"),
			"last_call_time": &now,
		}
		if queue.CallTime == nil {
			updates["call_time"] = &now
		}

		return tx.Model(&queue).Updates(updates).Error
	})
}

func (r *QueueRepository) Cancel(queueID uint, reason string) error {
	now := time.Now()
	return database.DB.Model(&model.QueueNumber{}).Where("id = ?", queueID).Updates(map[string]interface{}{
		"status":      3,
		"cancel_time": &now,
		"remark":      reason,
	}).Error
}

func (r *QueueRepository) Arrive(queueID uint, tableID uint, tableNo string) error {
	now := time.Now()
	return database.DB.Model(&model.QueueNumber{}).Where("id = ?", queueID).Updates(map[string]interface{}{
		"status":     2,
		"arrive_time": &now,
		"table_id":   tableID,
		"table_no":   tableNo,
	}).Error
}

func (r *QueueRepository) GetByID(id uint) (*model.QueueNumber, error) {
	var queue model.QueueNumber
	err := database.DB.Preload("Member").First(&queue, id).Error
	return &queue, err
}

func (r *QueueRepository) List(query *dto.QueueQueryDTO) ([]model.QueueNumber, int64, error) {
	var queues []model.QueueNumber
	var total int64
	db := database.DB.Model(&model.QueueNumber{})

	if query.StoreID > 0 {
		db = db.Where("store_id = ?", query.StoreID)
	}
	if query.QueueType != "" {
		db = db.Where("queue_type = ?", query.QueueType)
	}
	if query.Status > 0 {
		db = db.Where("status = ?", query.Status)
	}
	if query.MemberID > 0 {
		db = db.Where("member_id = ?", query.MemberID)
	}
	if query.Keyword != "" {
		db = db.Where("queue_number LIKE ? OR member_name LIKE ? OR member_phone LIKE ?",
			"%"+query.Keyword+"%", "%"+query.Keyword+"%", "%"+query.Keyword+"%")
	}

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (query.PageNum - 1) * query.PageSize
	err := db.Preload("Member").Order("status ASC, queue_type ASC, sequence ASC").
		Offset(offset).Limit(query.PageSize).Find(&queues).Error

	return queues, total, err
}

func (r *QueueRepository) GetWaitingCount(storeID uint, queueType string) (int, error) {
	var count int64
	err := database.DB.Model(&model.QueueNumber{}).
		Where("store_id = ? AND queue_type = ? AND status = 1", storeID, queueType).
		Count(&count).Error
	return int(count), err
}

func (r *QueueRepository) GetAheadCount(storeID uint, queueType string, sequence int) (int, error) {
	var count int64
	err := database.DB.Model(&model.QueueNumber{}).
		Where("store_id = ? AND queue_type = ? AND status = 1 AND sequence < ?", storeID, queueType, sequence).
		Count(&count).Error
	return int(count), err
}

func (r *QueueRepository) GetMyQueue(memberID uint, storeID uint) ([]model.QueueNumber, error) {
	var queues []model.QueueNumber
	err := database.DB.Where("member_id = ? AND store_id = ? AND status IN (1,2)", memberID, storeID).
		Order("created_at DESC").Find(&queues).Error
	return queues, err
}

func (r *QueueRepository) GetNextToCall(storeID uint, queueType string) (*model.QueueNumber, error) {
	var queue model.QueueNumber
	err := database.DB.Where("store_id = ? AND queue_type = ? AND status = 1", storeID, queueType).
		Order("sequence ASC").First(&queue).Error
	if err != nil {
		return nil, err
	}
	return &queue, nil
}

func (r *QueueRepository) GetConfig(storeID uint) (*model.QueueConfig, error) {
	var config model.QueueConfig
	err := database.DB.Where("store_id = ?", storeID).First(&config).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &model.QueueConfig{
				StoreID:        storeID,
				SmallPrefix:    "A",
				SmallCapacity:  4,
				MediumPrefix:   "B",
				MediumCapacity: 6,
				LargePrefix:    "C",
				LargeCapacity:  10,
				AutoCall:       true,
				CallInterval:   300,
				MaxCallCount:   3,
				AutoExpire:     true,
				ExpireMinutes:  15,
				VoiceNotify:    true,
				SMSNotify:      false,
			}, nil
		}
		return nil, err
	}
	return &config, nil
}

func (r *QueueRepository) SaveConfig(config *model.QueueConfig) error {
	var existing model.QueueConfig
	err := database.DB.Where("store_id = ?", config.StoreID).First(&existing).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return database.DB.Create(config).Error
		}
		return err
	}
	return database.DB.Model(&existing).Updates(config).Error
}

func (r *QueueRepository) ExpireOverdue(storeID uint, expireMinutes int) error {
	expireTime := time.Now().Add(-time.Duration(expireMinutes) * time.Minute)
	return database.DB.Model(&model.QueueNumber{}).
		Where("store_id = ? AND status = 1 AND last_call_time IS NOT NULL AND last_call_time < ?",
			storeID, expireTime).
		Updates(map[string]interface{}{
			"status": 4,
		}).Error
}
