package model

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/shopspring/decimal"
)

type TableOrderHistory struct {
	BaseModel
	StoreID     uint            `gorm:"not null;index" json:"store_id"`
	TableHash   string          `gorm:"size:64;not null;index" json:"table_hash"`
	VisitCount  int             `gorm:"default:1" json:"visit_count"`
	LastVisit   *time.Time      `json:"last_visit"`
	TotalAmount decimal.Decimal `gorm:"type:decimal(12,2);default:0" json:"total_amount"`
}

type TableOrderItem struct {
	BaseModel
	HistoryID   uint            `gorm:"not null;index" json:"history_id"`
	StoreID     uint            `gorm:"not null;index" json:"store_id"`
	TableHash   string          `gorm:"size:64;not null;index" json:"table_hash"`
	ProductID   uint            `gorm:"not null;index" json:"product_id"`
	ProductName string          `gorm:"size:100;not null" json:"product_name"`
	CategoryID  uint            `gorm:"index" json:"category_id"`
	OrderCount  int             `gorm:"default:1" json:"order_count"`
	TotalQty    int             `gorm:"default:0" json:"total_qty"`
	TotalAmount decimal.Decimal `gorm:"type:decimal(12,2);default:0" json:"total_amount"`
	LastOrderAt *time.Time      `json:"last_order_at"`
	Product     Product         `gorm:"foreignKey:ProductID" json:"product,omitempty"`
}

type ProductRecommendScore struct {
	BaseModel
	StoreID     uint            `gorm:"not null;index" json:"store_id"`
	ProductID   uint            `gorm:"not null;index" json:"product_id"`
	Score       decimal.Decimal `gorm:"type:decimal(10,4);default:0" json:"score"`
	HotScore    decimal.Decimal `gorm:"type:decimal(10,4);default:0" json:"hot_score"`
	TimeScore   decimal.Decimal `gorm:"type:decimal(10,4);default:0" json:"time_score"`
	HistoryScore decimal.Decimal `gorm:"type:decimal(10,4);default:0" json:"history_score"`
	UpdatedAt   *time.Time      `json:"updated_at"`
}

const tableHashSecret = "stalll-pos-table-anonymization-v1"

func GenerateTableHash(storeID uint, tableNo string) string {
	data := fmt.Sprintf("%d:%s", storeID, tableNo)
	mac := hmac.New(sha256.New, []byte(tableHashSecret))
	mac.Write([]byte(data))
	return hex.EncodeToString(mac.Sum(nil))
}

type RecommendConfig struct {
	BaseModel
	StoreID                 uint            `gorm:"not null;uniqueIndex" json:"store_id"`
	CFWeight                float64         `gorm:"type:decimal(5,4);default:0.5" json:"cf_weight"`
	HotWeight               float64         `gorm:"type:decimal(5,4);default:0.25" json:"hot_weight"`
	UserHistoryWeight       float64         `gorm:"type:decimal(5,4);default:0.25" json:"user_history_weight"`
	CategoryDiversityWeight float64         `gorm:"type:decimal(5,4);default:0.1" json:"category_diversity_weight"`
	RecommendCount          int             `gorm:"default:8" json:"recommend_count"`
	MinOrderPairs           int             `gorm:"default:3" json:"min_order_pairs"`
	MinSimilarity           float64         `gorm:"type:decimal(5,4);default:0.05" json:"min_similarity"`
	HotDays                 int             `gorm:"default:30" json:"hot_days"`
	CFDays                  int             `gorm:"default:90" json:"cf_days"`
	UserHistoryDays         int             `gorm:"default:180" json:"user_history_days"`
	UserHistoryTopK         int             `gorm:"default:20" json:"user_history_top_k"`
	Enabled                 bool            `gorm:"default:true" json:"enabled"`
	AutoRefresh             bool            `gorm:"default:true" json:"auto_refresh"`
	RefreshIntervalHours    int             `gorm:"default:6" json:"refresh_interval_hours"`
	LastRefreshedAt         *time.Time      `json:"last_refreshed_at"`
}

type RecommendResult struct {
	BaseModel
	StoreID            uint            `gorm:"not null;index:idx_store_product" json:"store_id"`
	ProductID          uint            `gorm:"not null;index:idx_store_product" json:"product_id"`
	RecommendProductID uint            `gorm:"not null;index" json:"recommend_product_id"`
	CFScore            float64         `gorm:"type:decimal(10,4);default:0" json:"cf_score"`
	HotScore           float64         `gorm:"type:decimal(10,4);default:0" json:"hot_score"`
	Score              float64         `gorm:"type:decimal(10,4);default:0;index" json:"score"`
	Reason             string          `gorm:"size:50" json:"reason"`
	RecommendProduct   Product         `gorm:"foreignKey:RecommendProductID" json:"recommend_product,omitempty"`
}

type HotProduct struct {
	ProductID   uint            `json:"product_id"`
	ProductName string          `json:"product_name"`
	CategoryID  uint            `json:"category_id"`
	TotalQty    int             `json:"total_qty"`
	SoldCount   int             `json:"sold_count"`
	HotScore    float64         `json:"hot_score"`
}
