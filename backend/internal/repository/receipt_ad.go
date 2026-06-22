package repository

import (
	"stalll-hub-pos/backend/internal/model"
	"stalll-hub-pos/backend/pkg/database"
	"time"
)

type ReceiptAdRepository struct{}

func NewReceiptAdRepository() *ReceiptAdRepository {
	return &ReceiptAdRepository{}
}

func (r *ReceiptAdRepository) Create(ad *model.ReceiptAd) error {
	return database.DB.Create(ad).Error
}

func (r *ReceiptAdRepository) Update(ad *model.ReceiptAd) error {
	return database.DB.Save(ad).Error
}

func (r *ReceiptAdRepository) Delete(id uint) error {
	return database.DB.Delete(&model.ReceiptAd{}, id).Error
}

func (r *ReceiptAdRepository) GetByID(id uint) (*model.ReceiptAd, error) {
	var ad model.ReceiptAd
	err := database.DB.First(&ad, id).Error
	if err != nil {
		return nil, err
	}
	return &ad, nil
}

func (r *ReceiptAdRepository) List(storeID uint, page, pageSize int, status int, position, adType, keyword string) ([]model.ReceiptAd, int64, error) {
	var list []model.ReceiptAd
	var total int64

	db := database.DB.Model(&model.ReceiptAd{})
	if storeID > 0 {
		db = db.Where("store_id = ?", storeID)
	}
	if status > 0 {
		db = db.Where("status = ?", status)
	}
	if position != "" {
		db = db.Where("position = ?", position)
	}
	if adType != "" {
		db = db.Where("ad_type = ?", adType)
	}
	if keyword != "" {
		db = db.Where("title LIKE ? OR content LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}

	err := db.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	err = db.Order("sort_order ASC, id DESC").Offset(offset).Limit(pageSize).Find(&list).Error
	if err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

func (r *ReceiptAdRepository) GetActiveAds(storeID uint, position string) ([]model.ReceiptAd, error) {
	var list []model.ReceiptAd
	today := time.Now().Format("2006-01-02")
	nowTime := time.Now().Format("15:04")

	db := database.DB.Model(&model.ReceiptAd{}).Where("store_id = ? AND status = 1", storeID)
	if position != "" {
		db = db.Where("position = ?", position)
	}

	db = db.Where(`
		(start_date = '' OR start_date <= ?) AND
		(end_date = '' OR end_date >= ?)
	`, today, today)

	db = db.Where(`
		(start_time = '' OR start_time <= ?) AND
		(end_time = '' OR end_time >= ?)
	`, nowTime, nowTime)

	err := db.Order("sort_order ASC, id DESC").Find(&list).Error
	return list, err
}

func (r *ReceiptAdRepository) IncrementViewCount(id uint) error {
	return database.DB.Model(&model.ReceiptAd{}).Where("id = ?", id).
		UpdateColumn("view_count", database.DB.Raw("view_count + 1")).Error
}

func (r *ReceiptAdRepository) IncrementClickCount(id uint) error {
	return database.DB.Model(&model.ReceiptAd{}).Where("id = ?", id).
		UpdateColumn("click_count", database.DB.Raw("click_count + 1")).Error
}

func (r *ReceiptAdRepository) UpdateStatus(id uint, status int) error {
	return database.DB.Model(&model.ReceiptAd{}).Where("id = ?", id).
		Update("status", status).Error
}

type ReceiptAdClickRepository struct{}

func NewReceiptAdClickRepository() *ReceiptAdClickRepository {
	return &ReceiptAdClickRepository{}
}

func (r *ReceiptAdClickRepository) Create(click *model.ReceiptAdClick) error {
	return database.DB.Create(click).Error
}

func (r *ReceiptAdClickRepository) GetStats(storeID uint, adID uint, startDate, endDate string) ([]model.ReceiptAdStats, error) {
	var stats []model.ReceiptAdStats

	type result struct {
		Date      string
		ViewCount int
		ClickCount int
	}

	var results []result

	db := database.DB.Table("receipt_ad_clicks").
		Select("DATE(created_at) as date, COUNT(*) as click_count, 0 as view_count").
		Where("store_id = ?", storeID)

	if adID > 0 {
		db = db.Where("ad_id = ?", adID)
	}
	if startDate != "" {
		db = db.Where("DATE(created_at) >= ?", startDate)
	}
	if endDate != "" {
		db = db.Where("DATE(created_at) <= ?", endDate)
	}

	err := db.Group("DATE(created_at)").Order("date ASC").Scan(&results).Error
	if err != nil {
		return nil, err
	}

	for _, r := range results {
		stats = append(stats, model.ReceiptAdStats{
			Date:       r.Date,
			ViewCount:  r.ViewCount,
			ClickCount: r.ClickCount,
		})
	}

	return stats, nil
}
