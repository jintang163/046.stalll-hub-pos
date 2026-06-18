package repository

import (
	"time"

	"stalll-hub-pos/backend/internal/model"
	"stalll-hub-pos/backend/pkg/database"

	"gorm.io/gorm"
)

type MemberRepository struct {
	db *gorm.DB
}

func NewMemberRepository(db *gorm.DB) *MemberRepository {
	return &MemberRepository{db: db}
}

func (r *MemberRepository) Create(member *model.Member) error {
	return database.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(member).Error; err != nil {
			return err
		}
		if member.Points > 0 {
			record := &model.PointsRecord{
				MemberID: member.ID,
				StoreID:  member.StoreID,
				Type:     "register",
				Points:   member.Points,
				Balance:  member.Points,
				Remark:   "注册赠送",
			}
			return tx.Create(record).Error
		}
		return nil
	})
}

func (r *MemberRepository) GetByID(id uint) (*model.Member, error) {
	var member model.Member
	err := database.DB.Preload("Level").Preload("Store").First(&member, id).Error
	if err != nil {
		return nil, err
	}
	return &member, nil
}

func (r *MemberRepository) GetByPhone(storeID uint, phone string) (*model.Member, error) {
	var member model.Member
	err := database.DB.Where("store_id = ? AND phone = ?", storeID, phone).
		Preload("Level").Preload("Store").First(&member).Error
	if err != nil {
		return nil, err
	}
	return &member, nil
}

func (r *MemberRepository) List(storeID uint, name, phone string, levelID, status int, page, pageSize int) ([]model.Member, int64, error) {
	var members []model.Member
	var total int64

	db := database.DB.Model(&model.Member{})
	if storeID > 0 {
		db = db.Where("store_id = ?", storeID)
	}
	if name != "" {
		db = db.Where("name LIKE ?", "%"+name+"%")
	}
	if phone != "" {
		db = db.Where("phone LIKE ?", "%"+phone+"%")
	}
	if levelID > 0 {
		db = db.Where("level_id = ?", levelID)
	}
	if status > 0 {
		db = db.Where("status = ?", status)
	}

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	err := db.Preload("Level").Preload("Store").
		Order("id DESC").Offset(offset).Limit(pageSize).Find(&members).Error
	return members, total, err
}

func (r *MemberRepository) Update(member *model.Member) error {
	return database.DB.Save(member).Error
}

func (r *MemberRepository) AdjustPoints(id uint, points int, remark string, orderID uint) error {
	return database.DB.Transaction(func(tx *gorm.DB) error {
		var member model.Member
		if err := tx.First(&member, id).Error; err != nil {
			return err
		}

		newPoints := member.Points + points
		if newPoints < 0 {
			newPoints = 0
		}

		if err := tx.Model(&member).Update("points", newPoints).Error; err != nil {
			return err
		}

		record := &model.PointsRecord{
			MemberID: id,
			StoreID:  member.StoreID,
			Type:     "adjust",
			Points:   points,
			Balance:  newPoints,
			OrderID:  orderID,
			Remark:   remark,
		}
		if points > 0 {
			record.Type = "earn"
		} else if points < 0 {
			record.Type = "spend"
		}
		return tx.Create(record).Error
	})
}

func (r *MemberRepository) GetPointsRecords(memberID, storeID uint, recordType, startDate, endDate string, page, pageSize int) ([]model.PointsRecord, int64, error) {
	var records []model.PointsRecord
	var total int64

	db := database.DB.Model(&model.PointsRecord{})
	if memberID > 0 {
		db = db.Where("member_id = ?", memberID)
	}
	if storeID > 0 {
		db = db.Where("store_id = ?", storeID)
	}
	if recordType != "" {
		db = db.Where("type = ?", recordType)
	}
	if startDate != "" {
		db = db.Where("created_at >= ?", startDate)
	}
	if endDate != "" {
		db = db.Where("created_at <= ?", endDate+" 23:59:59")
	}

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	err := db.Preload("Member").Preload("Store").
		Order("id DESC").Offset(offset).Limit(pageSize).Find(&records).Error
	return records, total, err
}

type MemberLevelRepository struct {
	db *gorm.DB
}

func NewMemberLevelRepository(db *gorm.DB) *MemberLevelRepository {
	return &MemberLevelRepository{db: db}
}

func (r *MemberLevelRepository) Create(level *model.MemberLevel) error {
	return database.DB.Create(level).Error
}

func (r *MemberLevelRepository) GetByID(id uint) (*model.MemberLevel, error) {
	var level model.MemberLevel
	err := database.DB.First(&level, id).Error
	if err != nil {
		return nil, err
	}
	return &level, nil
}

func (r *MemberLevelRepository) List() ([]model.MemberLevel, error) {
	var levels []model.MemberLevel
	err := database.DB.Order("points_required ASC").Find(&levels).Error
	return levels, err
}

func (r *MemberLevelRepository) Update(level *model.MemberLevel) error {
	return database.DB.Save(level).Error
}

func (r *MemberLevelRepository) Delete(id uint) error {
	return database.DB.Delete(&model.MemberLevel{}, id).Error
}

func (r *MemberLevelRepository) GetByPoints(points int) (*model.MemberLevel, error) {
	var level model.MemberLevel
	err := database.DB.Where("points_required <= ?", points).
		Order("points_required DESC").First(&level).Error
	if err != nil {
		return nil, err
	}
	return &level, nil
}

func (r *MemberLevelRepository) CountMembers(levelID uint) (int64, error) {
	var count int64
	err := database.DB.Model(&model.Member{}).Where("level_id = ?", levelID).Count(&count).Error
	return count, err
}

func (r *MemberRepository) UpdateStats(memberID uint, orderAmount float64) error {
	return database.DB.Model(&model.Member{}).Where("id = ?", memberID).
		Updates(map[string]interface{}{
			"total_spent":  gorm.Expr("total_spent + ?", orderAmount),
			"total_orders": gorm.Expr("total_orders + 1"),
		}).Error
}

func (r *MemberRepository) GetByOpenID(storeID uint, openID string) (*model.Member, error) {
	var member model.Member
	err := database.DB.Where("store_id = ? AND open_id = ?", storeID, openID).
		Preload("Level").Preload("Store").First(&member).Error
	if err != nil {
		return nil, err
	}
	return &member, nil
}

func (r *MemberRepository) UpdateLastActive(memberID uint) error {
	return database.DB.Model(&model.Member{}).Where("id = ?", memberID).
		Update("last_active", time.Now()).Error
}
