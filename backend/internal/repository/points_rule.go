package repository

import (
	"stalll-hub-pos/backend/internal/model"
	"stalll-hub-pos/backend/pkg/database"

	"gorm.io/gorm"
)

type PointsRuleRepository struct {
	db *gorm.DB
}

func NewPointsRuleRepository(db *gorm.DB) *PointsRuleRepository {
	return &PointsRuleRepository{db: db}
}

func (r *PointsRuleRepository) Create(rule *model.PointsRule) error {
	return database.DB.Create(rule).Error
}

func (r *PointsRuleRepository) GetByID(id uint) (*model.PointsRule, error) {
	var rule model.PointsRule
	err := database.DB.First(&rule, id).Error
	if err != nil {
		return nil, err
	}
	return &rule, nil
}

func (r *PointsRuleRepository) GetByRuleKey(storeID uint, ruleKey string) (*model.PointsRule, error) {
	var rule model.PointsRule
	err := database.DB.Where("store_id = ? AND rule_key = ?", storeID, ruleKey).First(&rule).Error
	if err != nil {
		return nil, err
	}
	return &rule, nil
}

func (r *PointsRuleRepository) List(storeID uint, ruleType string, status int, page, pageSize int) ([]model.PointsRule, int64, error) {
	var rules []model.PointsRule
	var total int64

	db := database.DB.Model(&model.PointsRule{})
	if storeID > 0 {
		db = db.Where("store_id = ?", storeID)
	}
	if ruleType != "" {
		db = db.Where("rule_type = ?", ruleType)
	}
	if status > 0 {
		db = db.Where("status = ?", status)
	}

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	err := db.Order("priority ASC, id ASC").Offset(offset).Limit(pageSize).Find(&rules).Error
	return rules, total, err
}

func (r *PointsRuleRepository) Update(rule *model.PointsRule) error {
	return database.DB.Save(rule).Error
}

func (r *PointsRuleRepository) Delete(id uint) error {
	return database.DB.Delete(&model.PointsRule{}, id).Error
}

func (r *PointsRuleRepository) GetActiveRulesByType(storeID uint, ruleType string) ([]model.PointsRule, error) {
	var rules []model.PointsRule
	err := database.DB.Where("store_id = ? AND rule_type = ? AND status = 1", storeID, ruleType).
		Order("priority ASC").Find(&rules).Error
	return rules, err
}
