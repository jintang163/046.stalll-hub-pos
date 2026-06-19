package dto

type RecommendConfigDTO struct {
	ID                     uint    `json:"id"`
	StoreID                uint    `json:"store_id"`
	CFWeight               float64 `json:"cf_weight"`
	HotWeight              float64 `json:"hot_weight"`
	CategoryDiversityWeight float64 `json:"category_diversity_weight"`
	RecommendCount         int     `json:"recommend_count"`
	MinOrderPairs          int     `json:"min_order_pairs"`
	MinSimilarity          float64 `json:"min_similarity"`
	HotDays                int     `json:"hot_days"`
	CFDays                 int     `json:"cf_days"`
	Enabled                bool    `json:"enabled"`
	AutoRefresh            bool    `json:"auto_refresh"`
	RefreshIntervalHours   int     `json:"refresh_interval_hours"`
}

type UpdateRecommendConfigRequest struct {
	CFWeight               *float64 `json:"cf_weight"`
	HotWeight              *float64 `json:"hot_weight"`
	CategoryDiversityWeight *float64 `json:"category_diversity_weight"`
	RecommendCount         *int     `json:"recommend_count"`
	MinOrderPairs          *int     `json:"min_order_pairs"`
	MinSimilarity          *float64 `json:"min_similarity"`
	HotDays                *int     `json:"hot_days"`
	CFDays                 *int     `json:"cf_days"`
	Enabled                *bool    `json:"enabled"`
	AutoRefresh            *bool    `json:"auto_refresh"`
	RefreshIntervalHours   *int     `json:"refresh_interval_hours"`
}

type RecommendItemDTO struct {
	ProductID   uint   `json:"product_id"`
	ProductName string `json:"product_name"`
	CategoryID  uint   `json:"category_id"`
	MainImage   string `json:"main_image"`
	Price       string `json:"price"`
	Score       float64 `json:"score"`
	Reason      string `json:"reason"`
}

type GetRecommendRequest struct {
	StoreID    uint   `form:"store_id" json:"store_id"`
	ProductIDs []uint `form:"product_ids" json:"product_ids"`
	Count      int    `form:"count" json:"count"`
}

type TriggerRefreshRequest struct {
	StoreID uint `json:"store_id" binding:"required"`
}

type RefreshStatusDTO struct {
	StoreID         uint   `json:"store_id"`
	Enabled         bool   `json:"enabled"`
	LastRefreshedAt string `json:"last_refreshed_at"`
	IsRunning       bool   `json:"is_running"`
	TotalProducts   int    `json:"total_products"`
	TotalPairs      int    `json:"total_pairs"`
}
