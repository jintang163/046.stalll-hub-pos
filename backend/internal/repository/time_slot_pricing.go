package repository

import (
	"time"

	"gorm.io/gorm"
	"stalll-hub-pos/backend/internal/model"
	"stalll-hub-pos/backend/pkg/database"
)

type TimeSlotPricingRepository struct {
	db *gorm.DB
}

func NewTimeSlotPricingRepository() *TimeSlotPricingRepository {
	return &TimeSlotPricingRepository{
		db: database.DB,
	}
}

func (r *TimeSlotPricingRepository) Create(pricing *model.TimeSlotPricing) error {
	return r.db.Create(pricing).Error
}

func (r *TimeSlotPricingRepository) Update(pricing *model.TimeSlotPricing) error {
	return r.db.Save(pricing).Error
}

func (r *TimeSlotPricingRepository) Delete(id uint) error {
	return r.db.Delete(&model.TimeSlotPricing{}, id).Error
}

func (r *TimeSlotPricingRepository) GetByID(id uint) (*model.TimeSlotPricing, error) {
	var pricing model.TimeSlotPricing
	err := r.db.Preload("Store").First(&pricing, id).Error
	if err != nil {
		return nil, err
	}
	return &pricing, nil
}

func (r *TimeSlotPricingRepository) List(storeID uint, name string, status *int, offset, limit int) ([]model.TimeSlotPricing, int64, error) {
	var pricings []model.TimeSlotPricing
	var total int64

	query := r.db.Model(&model.TimeSlotPricing{}).Where("store_id = ?", storeID)

	if name != "" {
		query = query.Where("name LIKE ?", "%"+name+"%")
	}
	if status != nil {
		query = query.Where("status = ?", *status)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := query.Preload("Store").
		Order("priority DESC, id DESC").
		Offset(offset).Limit(limit).Find(&pricings).Error

	return pricings, total, err
}

func (r *TimeSlotPricingRepository) GetActiveByStoreAndTime(storeID uint, checkTime time.Time) ([]model.TimeSlotPricing, error) {
	var pricings []model.TimeSlotPricing
	weekday := int(checkTime.Weekday())
	if weekday == 0 {
		weekday = 7
	}
	timeStr := checkTime.Format("15:04")

	err := r.db.Where("store_id = ? AND status = 1", storeID).
		Where("FIND_IN_SET(?, weekdays) > 0", weekday).
		Where("start_time <= ? AND end_time >= ?", timeStr, timeStr).
		Order("priority DESC, id ASC").
		Find(&pricings).Error

	return pricings, err
}

func (r *TimeSlotPricingRepository) GetActivePricings(storeID uint) ([]model.TimeSlotPricing, error) {
	var pricings []model.TimeSlotPricing
	err := r.db.Where("store_id = ? AND status = 1", storeID).
		Order("priority DESC, id ASC").
		Find(&pricings).Error
	return pricings, err
}

type StockReservationRepository struct {
	db *gorm.DB
}

func NewStockReservationRepository() *StockReservationRepository {
	return &StockReservationRepository{
		db: database.DB,
	}
}

func (r *StockReservationRepository) Create(reservation *model.StockReservation) error {
	return r.db.Create(reservation).Error
}

func (r *StockReservationRepository) ReleaseByOrderID(orderID uint) error {
	now := time.Now()
	return r.db.Model(&model.StockReservation{}).
		Where("order_id = ? AND is_released = ?", orderID, false).
		Updates(map[string]interface{}{
			"is_released": true,
			"released_at": now,
		}).Error
}

func (r *StockReservationRepository) GetByOrderID(orderID uint) ([]model.StockReservation, error) {
	var reservations []model.StockReservation
	err := r.db.Where("order_id = ?", orderID).
		Preload("SKU").Preload("Product").
		Find(&reservations).Error
	return reservations, err
}

func (r *StockReservationRepository) CleanExpiredReservations(now time.Time) (int64, error) {
	result := r.db.Model(&model.StockReservation{}).
		Where("expire_at <= ? AND is_released = ?", now, false).
		Updates(map[string]interface{}{
			"is_released": true,
			"released_at": now,
		})
	return result.RowsAffected, result.Error
}

type ReservationReminderRepository struct {
	db *gorm.DB
}

func NewReservationReminderRepository() *ReservationReminderRepository {
	return &ReservationReminderRepository{
		db: database.DB,
	}
}

func (r *ReservationReminderRepository) Create(reminder *model.ReservationReminder) error {
	return r.db.Create(reminder).Error
}

func (r *ReservationReminderRepository) GetPendingReminders(now time.Time) ([]model.ReservationReminder, error) {
	var reminders []model.ReservationReminder
	err := r.db.Where("remind_at <= ? AND is_sent = ?", now, false).
		Preload("Order").
		Order("remind_at ASC").
		Find(&reminders).Error
	return reminders, err
}

func (r *ReservationReminderRepository) MarkAsSent(id uint, sentAt time.Time) error {
	return r.db.Model(&model.ReservationReminder{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"is_sent": true,
			"sent_at": sentAt,
		}).Error
}
