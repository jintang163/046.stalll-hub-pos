package repository

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"gorm.io/gorm"
	"stalll-hub-pos/backend/internal/dto"
	"stalll-hub-pos/backend/internal/model"
	"stalll-hub-pos/backend/pkg/database"
)

type PromotionRepository struct {
	db *gorm.DB
}

func NewPromotionRepository(db *gorm.DB) *PromotionRepository {
	return &PromotionRepository{db: db}
}

func (r *PromotionRepository) Create(promotion *model.Promotion, tiers []model.PromotionTier) error {
	return database.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(promotion).Error; err != nil {
			return err
		}
		for i := range tiers {
			tiers[i].PromotionID = promotion.ID
		}
		if len(tiers) > 0 {
			if err := tx.Create(&tiers).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *PromotionRepository) GetByID(id uint) (*model.Promotion, error) {
	var promotion model.Promotion
	err := database.DB.Preload("Store").First(&promotion, id).Error
	if err != nil {
		return nil, err
	}
	return &promotion, nil
}

func (r *PromotionRepository) GetTiers(promotionID uint) ([]model.PromotionTier, error) {
	var tiers []model.PromotionTier
	err := database.DB.Where("promotion_id = ?", promotionID).Order("min_amount ASC").Find(&tiers).Error
	return tiers, err
}

func (r *PromotionRepository) List(name, promotionType string, status, page, pageSize int) ([]model.Promotion, int64, error) {
	var promotions []model.Promotion
	var total int64

	db := database.DB.Model(&model.Promotion{})
	if name != "" {
		db = db.Where("name LIKE ?", "%"+name+"%")
	}
	if promotionType != "" {
		db = db.Where("type = ?", promotionType)
	}
	if status > 0 {
		db = db.Where("status = ?", status)
	}

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	err := db.Order("priority DESC, id DESC").Offset(offset).Limit(pageSize).Find(&promotions).Error
	return promotions, total, err
}

func (r *PromotionRepository) Update(promotion *model.Promotion, tiers []model.PromotionTier) error {
	return database.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Save(promotion).Error; err != nil {
			return err
		}
		if err := tx.Where("promotion_id = ?", promotion.ID).Delete(&model.PromotionTier{}).Error; err != nil {
			return err
		}
		for i := range tiers {
			tiers[i].PromotionID = promotion.ID
		}
		if len(tiers) > 0 {
			if err := tx.Create(&tiers).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *PromotionRepository) Delete(id uint) error {
	return database.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("promotion_id = ?", id).Delete(&model.PromotionTier{}).Error; err != nil {
			return err
		}
		if err := tx.Delete(&model.Promotion{}, id).Error; err != nil {
			return err
		}
		return nil
	})
}

func (r *PromotionRepository) GetActivePromotions(storeID uint) ([]model.Promotion, error) {
	var promotions []model.Promotion
	now := time.Now()
	db := database.DB.Where("status = 1")
	if storeID > 0 {
		db = db.Where("store_id = ?", storeID)
	}
	db = db.Where("(start_time IS NULL OR start_time <= ?) AND (end_time IS NULL OR end_time >= ?)", now, now)
	err := db.Order("priority DESC, id ASC").Find(&promotions).Error
	return promotions, err
}

func (r *PromotionRepository) UpdateStatus(id, status uint) error {
	return database.DB.Model(&model.Promotion{}).Where("id = ?", id).Update("status", status).Error
}

func (r *PromotionRepository) ActivatePendingPromotions() (int64, error) {
	now := time.Now()
	result := database.DB.Model(&model.Promotion{}).
		Where("status = 0 AND start_time IS NOT NULL AND start_time <= ? AND (end_time IS NULL OR end_time > ?)", now, now).
		Update("status", 1)
	return result.RowsAffected, result.Error
}

func (r *PromotionRepository) DeactivateExpiredPromotions() (int64, error) {
	now := time.Now()
	result := database.DB.Model(&model.Promotion{}).
		Where("status = 1 AND end_time IS NOT NULL AND end_time < ?", now).
		Update("status", 2)
	return result.RowsAffected, result.Error
}

func (r *PromotionRepository) ParseIDs(idsStr string) []uint {
	if idsStr == "" {
		return nil
	}
	parts := strings.Split(idsStr, ",")
	var ids []uint
	for _, part := range parts {
		var id uint
		if _, err := fmt.Sscanf(part, "%d", &id); err == nil {
			ids = append(ids, id)
		}
	}
	return ids
}

func (r *PromotionRepository) JoinIDs(ids []uint) string {
	if len(ids) == 0 {
		return ""
	}
	parts := make([]string, len(ids))
	for i, id := range ids {
		parts[i] = strconv.FormatUint(uint64(id), 10)
	}
	return strings.Join(parts, ",")
}

func (r *PromotionRepository) TiersToJSON(tiers []dto.PromotionTierDTO) string {
	if len(tiers) == 0 {
		return ""
	}
	b, _ := json.Marshal(tiers)
	return string(b)
}

func (r *PromotionRepository) JSONToTiers(jsonStr string) []dto.PromotionTierDTO {
	if jsonStr == "" {
		return nil
	}
	var tiers []dto.PromotionTierDTO
	_ = json.Unmarshal([]byte(jsonStr), &tiers)
	return tiers
}
