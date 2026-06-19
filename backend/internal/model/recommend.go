package model

import "time"

type RecommendConfig struct {
	BaseModel
	StoreID              uint    `gorm:"not null;index" json:"store_id"`
	CFWeight             float64 `gorm:"default:0.5" json:"cf_weight"`
	HotWeight            float64 `gorm:"default:0.25" json:"hot_weight"`
	UserHistoryWeight    float64 `gorm:"default:0.25" json:"user_history_weight"`
	CategoryDiversityWeight float64 `gorm:"default:0.1" json:"category_diversity_weight"`
	RecommendCount       int     `gorm:"default:8" json:"recommend_count"`
	MinOrderPairs        int     `gorm:"default:3" json:"min_order_pairs"`
	MinSimilarity        float64 `gorm:"default:0.05" json:"min_similarity"`
	HotDays              int     `gorm:"default:30" json:"hot_days"`
	CFDays               int     `gorm:"default:90" json:"cf_days"`
	UserHistoryDays      int     `gorm:"default:180" json:"user_history_days"`
	UserHistoryTopK      int     `gorm:"default:20" json:"user_history_top_k"`
	Enabled              bool    `gorm:"default:true" json:"enabled"`
	AutoRefresh          bool    `gorm:"default:true" json:"auto_refresh"`
	RefreshIntervalHours int     `gorm:"default:6" json:"refresh_interval_hours"`
	LastRefreshedAt      *time.Time `json:"last_refreshed_at"`
	Store                Store   `gorm:"foreignKey:StoreID" json:"store,omitempty"`
}

type RecommendResult struct {
	BaseModel
	StoreID          uint   `gorm:"not null;index" json:"store_id"`
	ProductID        uint   `gorm:"not null;index" json:"product_id"`
	RecommendProductID uint `gorm:"not null" json:"recommend_product_id"`
	Score            float64 `gorm:"default:0" json:"score"`
	CFScore          float64 `gorm:"default:0" json:"cf_score"`
	HotScore         float64 `gorm:"default:0" json:"hot_score"`
	UserHistoryScore float64 `gorm:"default:0" json:"user_history_score"`
	Reason           string `gorm:"size:100" json:"reason"`
	Store            Store  `gorm:"foreignKey:StoreID" json:"store,omitempty"`
	Product          Product `gorm:"foreignKey:ProductID" json:"product,omitempty"`
	RecommendProduct Product `gorm:"foreignKey:RecommendProductID" json:"recommend_product,omitempty"`
}

type HotProduct struct {
	StoreID     uint
	ProductID   uint
	ProductName string
	CategoryID  uint
	SoldCount   int
	HotScore    float64
}
