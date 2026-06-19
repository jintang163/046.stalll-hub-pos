package repository

import (
	"time"

	"stalll-hub-pos/backend/internal/model"
	"stalll-hub-pos/backend/pkg/database"

	"gorm.io/gorm"
)

type RechargeActivityRepository struct {
	db *gorm.DB
}

func NewRechargeActivityRepository(db *gorm.DB) *RechargeActivityRepository {
	return &RechargeActivityRepository{db: db}
}

func (r *RechargeActivityRepository) Create(activity *model.RechargeActivity) error {
	return database.DB.Create(activity).Error
}

func (r *RechargeActivityRepository) GetByID(id uint) (*model.RechargeActivity, error) {
	var activity model.RechargeActivity
	err := database.DB.First(&activity, id).Error
	if err != nil {
		return nil, err
	}
	return &activity, nil
}

func (r *RechargeActivityRepository) List(storeID uint, status int, page, pageSize int) ([]model.RechargeActivity, int64, error) {
	var activities []model.RechargeActivity
	var total int64

	db := database.DB.Model(&model.RechargeActivity{})
	if storeID > 0 {
		db = db.Where("store_id = ?", storeID)
	}
	if status > 0 {
		db = db.Where("status = ?", status)
	}

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	err := db.Order("id DESC").Offset(offset).Limit(pageSize).Find(&activities).Error
	return activities, total, err
}

func (r *RechargeActivityRepository) Update(activity *model.RechargeActivity) error {
	return database.DB.Save(activity).Error
}

func (r *RechargeActivityRepository) Delete(id uint) error {
	return database.DB.Delete(&model.RechargeActivity{}, id).Error
}

func (r *RechargeActivityRepository) GetActiveActivities(storeID uint) ([]model.RechargeActivity, error) {
	var activities []model.RechargeActivity
	now := time.Now()
	err := database.DB.Where("store_id = ? AND status = 1 AND start_time <= ? AND end_time >= ?",
		storeID, now, now).Order("min_amount ASC").Find(&activities).Error
	return activities, err
}

func (r *RechargeActivityRepository) GetPendingAutoActivate() ([]model.RechargeActivity, error) {
	var activities []model.RechargeActivity
	now := time.Now()
	err := database.DB.Where("auto_activate = true AND status = 0 AND start_time <= ?", now).
		Find(&activities).Error
	return activities, err
}

func (r *RechargeActivityRepository) GetExpiredActive() ([]model.RechargeActivity, error) {
	var activities []model.RechargeActivity
	now := time.Now()
	err := database.DB.Where("status = 1 AND end_time < ?", now).
		Find(&activities).Error
	return activities, err
}

type MemberRechargeRepository struct {
	db *gorm.DB
}

func NewMemberRechargeRepository(db *gorm.DB) *MemberRechargeRepository {
	return &MemberRechargeRepository{db: db}
}

func (r *MemberRechargeRepository) Create(recharge *model.MemberRecharge) error {
	return database.DB.Create(recharge).Error
}

func (r *MemberRechargeRepository) GetByID(id uint) (*model.MemberRecharge, error) {
	var recharge model.MemberRecharge
	err := database.DB.Preload("Member").Preload("Activity").First(&recharge, id).Error
	if err != nil {
		return nil, err
	}
	return &recharge, nil
}

func (r *MemberRechargeRepository) List(memberID, storeID uint, page, pageSize int) ([]model.MemberRecharge, int64, error) {
	var recharges []model.MemberRecharge
	var total int64

	db := database.DB.Model(&model.MemberRecharge{})
	if memberID > 0 {
		db = db.Where("member_id = ?", memberID)
	}
	if storeID > 0 {
		db = db.Where("store_id = ?", storeID)
	}

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	err := db.Preload("Activity").Order("id DESC").Offset(offset).Limit(pageSize).Find(&recharges).Error
	return recharges, total, err
}
