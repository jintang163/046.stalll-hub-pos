package dto

type RecommendConfigDTO struct {
	ID                     uint    `json:"id"`
	StoreID                uint    `json:"store_id"`
	CFWeight               float64 `json:"cf_weight"`
	HotWeight              float64 `json:"hot_weight"`
	UserHistoryWeight      float64 `json:"user_history_weight"`
	CategoryDiversityWeight float64 `json:"category_diversity_weight"`
	RecommendCount         int     `json:"recommend_count"`
	MinOrderPairs          int     `json:"min_order_pairs"`
	MinSimilarity          float64 `json:"min_similarity"`
	HotDays                int     `json:"hot_days"`
	CFDays                 int     `json:"cf_days"`
	UserHistoryDays        int     `json:"user_history_days"`
	UserHistoryTopK        int     `json:"user_history_top_k"`
	Enabled                bool    `json:"enabled"`
	AutoRefresh            bool    `json:"auto_refresh"`
	RefreshIntervalHours   int     `json:"refresh_interval_hours"`
}

type UpdateRecommendConfigRequest struct {
	CFWeight               *float64 `json:"cf_weight"`
	HotWeight              *float64 `json:"hot_weight"`
	UserHistoryWeight      *float64 `json:"user_history_weight"`
	CategoryDiversityWeight *float64 `json:"category_diversity_weight"`
	RecommendCount         *int     `json:"recommend_count"`
	MinOrderPairs          *int     `json:"min_order_pairs"`
	MinSimilarity          *float64 `json:"min_similarity"`
	HotDays                *int     `json:"hot_days"`
	CFDays                 *int     `json:"cf_days"`
	UserHistoryDays        *int     `json:"user_history_days"`
	UserHistoryTopK        *int     `json:"user_history_top_k"`
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
	SKUID       uint   `json:"sku_id"`
	Score       float64 `json:"score"`
	Reason      string `json:"reason"`
	ReasonType  string `json:"reason_type"`
}

type GetRecommendRequest struct {
	StoreID    uint   `form:"store_id" json:"store_id"`
	ProductIDs []uint `form:"product_ids" json:"product_ids"`
	MemberID   uint   `form:"member_id" json:"member_id"`
	UserID     uint   `form:"user_id" json:"user_id"`
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

type ConfigMetaItem struct {
	Key         string      `json:"key"`
	Label       string      `json:"label"`
	Description string      `json:"description"`
	Min         float64     `json:"min"`
	Max         float64     `json:"max"`
	Step        float64     `json:"step"`
	Type        string      `json:"type"`
	Unit        string      `json:"unit,omitempty"`
	Default     interface{} `json:"default"`
	Options     []OptionItem `json:"options,omitempty"`
}

type OptionItem struct {
	Label string      `json:"label"`
	Value interface{} `json:"value"`
}

