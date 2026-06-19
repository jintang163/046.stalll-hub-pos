package repository

import (
	"math"
	"time"

	"gorm.io/gorm"
	"stalll-hub-pos/backend/internal/model"
	"stalll-hub-pos/backend/pkg/database"
)

type RecommendRepository struct {
	db *gorm.DB
}

func NewRecommendRepository() *RecommendRepository {
	return &RecommendRepository{db: database.DB}
}

func (r *RecommendRepository) GetConfigByStoreID(storeID uint) (*model.RecommendConfig, error) {
	var config model.RecommendConfig
	err := r.db.Where("store_id = ?", storeID).First(&config).Error
	if err != nil {
		return nil, err
	}
	return &config, nil
}

func (r *RecommendRepository) CreateConfig(config *model.RecommendConfig) error {
	return r.db.Create(config).Error
}

func (r *RecommendRepository) UpdateConfig(config *model.RecommendConfig) error {
	return r.db.Save(config).Error
}

func (r *RecommendRepository) GetAllConfigs() ([]model.RecommendConfig, error) {
	var configs []model.RecommendConfig
	err := r.db.Find(&configs).Error
	return configs, err
}

func (r *RecommendRepository) GetOrderItemsForCF(storeID uint, days int) ([]model.OrderItem, error) {
	var items []model.OrderItem
	since := time.Now().AddDate(0, 0, -days)
	err := r.db.Table("order_items oi").
		Select("oi.id, oi.order_id, oi.product_id, oi.quantity, oi.created_at").
		Joins("JOIN orders o ON o.id = oi.order_id").
		Where("o.store_id = ? AND o.order_status IN ? AND o.created_at >= ?",
			storeID, []int{2, 3, 4}, since).
		Find(&items).Error
	return items, err
}

func (r *RecommendRepository) GetHotProducts(storeID uint, days int, limit int) ([]model.HotProduct, error) {
	var hots []model.HotProduct
	since := time.Now().AddDate(0, 0, -days)

	err := r.db.Table("order_items oi").
		Select("o.store_id, oi.product_id, p.name as product_name, p.category_id, SUM(oi.quantity) as sold_count").
		Joins("JOIN orders o ON o.id = oi.order_id").
		Joins("JOIN products p ON p.id = oi.product_id").
		Where("o.store_id = ? AND o.order_status IN ? AND o.created_at >= ? AND p.status = 1",
			storeID, []int{2, 3, 4}, since).
		Group("oi.product_id, o.store_id, p.name, p.category_id").
		Order("sold_count DESC").
		Limit(limit).
		Scan(&hots).Error

	if len(hots) > 0 {
		maxSold := 0
		for _, h := range hots {
			if h.SoldCount > maxSold {
				maxSold = h.SoldCount
			}
		}
		if maxSold > 0 {
			for i := range hots {
				hots[i].HotScore = float64(hots[i].SoldCount) / float64(maxSold)
			}
		}
	}

	return hots, err
}

func (r *RecommendRepository) ClearResultsByStore(storeID uint) error {
	return r.db.Where("store_id = ?", storeID).Delete(&model.RecommendResult{}).Error
}

func (r *RecommendRepository) BatchCreateResults(results []model.RecommendResult) error {
	if len(results) == 0 {
		return nil
	}
	return r.db.CreateInBatches(results, 500).Error
}

func (r *RecommendRepository) GetResultsByProducts(storeID uint, productIDs []uint, count int) ([]model.RecommendResult, error) {
	var results []model.RecommendResult
	if len(productIDs) == 0 {
		return results, nil
	}
	err := r.db.Preload("RecommendProduct").
		Preload("RecommendProduct.SKUs").
		Where("store_id = ? AND product_id IN ?", storeID, productIDs).
		Order("score DESC").
		Limit(count * 3).
		Find(&results).Error
	return results, err
}

func (r *RecommendRepository) GetAllStoreIDs() ([]uint, error) {
	var ids []uint
	err := r.db.Model(&model.Store{}).Where("status = 1").Pluck("id", &ids).Error
	return ids, err
}

func (r *RecommendRepository) GetValidProducts(storeID uint) ([]model.Product, error) {
	var products []model.Product
	err := r.db.Where("store_id = ? AND status = 1", storeID).Find(&products).Error
	return products, err
}

func (r *RecommendRepository) GetResultStats(storeID uint) (int, int) {
	var productCount int64
	var pairCount int64
	r.db.Model(&model.RecommendResult{}).Where("store_id = ?", storeID).
		Distinct("product_id").Count(&productCount)
	r.db.Model(&model.RecommendResult{}).Where("store_id = ?", storeID).Count(&pairCount)
	return int(productCount), int(pairCount)
}

type UserHistoryProduct struct {
	ProductID    uint
	ProductName  string
	CategoryID   uint
	BuyCount     int
	LastBuyDays  int
	HistoryScore float64
}

func (r *RecommendRepository) GetMemberHistoryProducts(storeID uint, memberID uint, userID uint, days int, topK int) ([]UserHistoryProduct, error) {
	var list []UserHistoryProduct
	if memberID == 0 && userID == 0 {
		return list, nil
	}
	since := time.Now().AddDate(0, 0, -days)

	query := r.db.Table("order_items oi").
		Select("oi.product_id, p.name as product_name, p.category_id, "+
			"SUM(oi.quantity) as buy_count, "+
			"MIN(DATEDIFF(NOW(), o.created_at)) as last_buy_days").
		Joins("JOIN orders o ON o.id = oi.order_id").
		Joins("JOIN products p ON p.id = oi.product_id").
		Where("o.store_id = ? AND o.order_status IN ? AND o.created_at >= ? AND p.status = 1",
			storeID, []int{2, 3, 4}, since).
		Group("oi.product_id, p.name, p.category_id")

	if memberID > 0 {
		query = query.Where("o.member_id = ?", memberID)
	} else if userID > 0 {
		query = query.Where("o.user_id = ?", userID)
	}

	err := query.Order("buy_count DESC, last_buy_days ASC").
		Limit(topK).
		Scan(&list).Error

	if len(list) > 0 {
		maxCount := 0
		for _, h := range list {
			if h.BuyCount > maxCount {
				maxCount = h.BuyCount
			}
		}
		if maxCount > 0 {
			for i := range list {
				freqPart := float64(list[i].BuyCount) / float64(maxCount)
				recencyPart := 1.0
				if list[i].LastBuyDays > 0 {
					recencyPart = 1.0 / math.Log2(float64(list[i].LastBuyDays)+1.0)
				}
				if recencyPart > 1 {
					recencyPart = 1.0
				}
				list[i].HistoryScore = 0.7*freqPart + 0.3*recencyPart
			}
		}
	}
	return list, err
}
