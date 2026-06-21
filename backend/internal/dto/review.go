package dto

import (
	"time"

	"github.com/shopspring/decimal"
)

type ReviewRatingQueryDTO struct {
	PageQuery
	StoreID   uint   `form:"store_id"`
	Platform  string `form:"platform"`
	StartDate string `form:"start_date"`
	EndDate   string `form:"end_date"`
}

type ReviewRatingResponse struct {
	ID                uint            `json:"id"`
	StoreID           uint            `json:"store_id"`
	StoreName         string          `json:"store_name"`
	Platform          string          `json:"platform"`
	OverallRating     decimal.Decimal `json:"overall_rating"`
	TasteRating       decimal.Decimal `json:"taste_rating"`
	EnvironmentRating decimal.Decimal `json:"environment_rating"`
	ServiceRating     decimal.Decimal `json:"service_rating"`
	ReviewCount       int             `json:"review_count"`
	GoodReviewCount   int             `json:"good_review_count"`
	MidReviewCount    int             `json:"mid_review_count"`
	BadReviewCount    int             `json:"bad_review_count"`
	GoodReviewRate    decimal.Decimal `json:"good_review_rate"`
	RatingDate        string          `json:"rating_date"`
	SnapshotTime      time.Time       `json:"snapshot_time"`
	CreatedAt         time.Time       `json:"created_at"`
	UpdatedAt         time.Time       `json:"updated_at"`
}

type RatingTrendResponse struct {
	Date              string          `json:"date"`
	OverallRating     decimal.Decimal `json:"overall_rating"`
	TasteRating       decimal.Decimal `json:"taste_rating"`
	EnvironmentRating decimal.Decimal `json:"environment_rating"`
	ServiceRating     decimal.Decimal `json:"service_rating"`
	ReviewCount       int             `json:"review_count"`
}

type ReviewQueryDTO struct {
	PageQuery
	StoreID      uint   `form:"store_id"`
	Platform     string `form:"platform"`
	RatingMin    string `form:"rating_min"`
	RatingMax    string `form:"rating_max"`
	IsBadReview  *bool  `form:"is_bad_review"`
	IsReplied    *bool  `form:"is_replied"`
	Keyword      string `form:"keyword"`
	StartDate    string `form:"start_date"`
	EndDate      string `form:"end_date"`
}

type ReviewResponse struct {
	ID                 uint            `json:"id"`
	StoreID            uint            `json:"store_id"`
	StoreName          string          `json:"store_name"`
	Platform           string          `json:"platform"`
	PlatformID         string          `json:"platform_id"`
	UserNickname       string          `json:"user_nickname"`
	UserAvatar         string          `json:"user_avatar"`
	UserLevel          string          `json:"user_level"`
	Rating             decimal.Decimal `json:"rating"`
	TasteRating        decimal.Decimal `json:"taste_rating"`
	EnvironmentRating  decimal.Decimal `json:"environment_rating"`
	ServiceRating      decimal.Decimal `json:"service_rating"`
	Content            string          `json:"content"`
	Images             string          `json:"images"`
	ReviewTime         time.Time       `json:"review_time"`
	ReplyContent       string          `json:"reply_content"`
	ReplyTime          *time.Time      `json:"reply_time"`
	IsBadReview        bool            `json:"is_bad_review"`
	IsReplied          bool            `json:"is_replied"`
	IsWorkOrderCreated bool            `json:"is_work_order_created"`
	OrderNo            string          `json:"order_no"`
	PerCapita          decimal.Decimal `json:"per_capita"`
	CreatedAt          time.Time       `json:"created_at"`
	UpdatedAt          time.Time       `json:"updated_at"`
}

type ReviewReplyRequest struct {
	ReplyContent string `json:"reply_content" binding:"required"`
}

type WorkOrderQueryDTO struct {
	PageQuery
	StoreID    uint   `form:"store_id"`
	Status     string `form:"status"`
	Priority   string `form:"priority"`
	AssigneeID uint   `form:"assignee_id"`
	Keyword    string `form:"keyword"`
}

type WorkOrderResponse struct {
	ID            uint       `json:"id"`
	StoreID       uint       `json:"store_id"`
	StoreName     string     `json:"store_name"`
	ReviewID      uint       `json:"review_id"`
	ReviewContent string     `json:"review_content"`
	WorkOrderNo   string     `json:"work_order_no"`
	Type          string     `json:"type"`
	Title         string     `json:"title"`
	Description   string     `json:"description"`
	Priority      string     `json:"priority"`
	Status        string     `json:"status"`
	AssigneeID    uint       `json:"assignee_id"`
	AssigneeName  string     `json:"assignee_name"`
	HandlerID     uint       `json:"handler_id"`
	HandlerName   string     `json:"handler_name"`
	HandleTime    *time.Time `json:"handle_time"`
	HandleResult  string     `json:"handle_result"`
	DueTime       *time.Time `json:"due_time"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

type WorkOrderCreateRequest struct {
	ReviewID     uint   `json:"review_id" binding:"required"`
	Priority     string `json:"priority" binding:"required,oneof=low medium high urgent"`
	AssigneeID   uint   `json:"assignee_id"`
	AssigneeName string `json:"assignee_name"`
}

type WorkOrderHandleRequest struct {
	Status       string `json:"status" binding:"required,oneof=processing completed cancelled"`
	HandleResult string `json:"handle_result" binding:"required"`
	HandlerName  string `json:"handler_name"`
}

type AlertQueryDTO struct {
	PageQuery
	StoreID   uint   `form:"store_id"`
	Status    string `form:"status"`
	AlertType string `form:"alert_type"`
	StartDate string `form:"start_date"`
	EndDate   string `form:"end_date"`
}

type AlertResponse struct {
	ID           uint            `json:"id"`
	StoreID      uint            `json:"store_id"`
	StoreName    string          `json:"store_name"`
	Platform     string          `json:"platform"`
	AlertType    string          `json:"alert_type"`
	Title        string          `json:"title"`
	Content      string          `json:"content"`
	PrevRating   decimal.Decimal `json:"prev_rating"`
	CurrRating   decimal.Decimal `json:"curr_rating"`
	RatingDrop   decimal.Decimal `json:"rating_drop"`
	Threshold    decimal.Decimal `json:"threshold"`
	Status       string          `json:"status"`
	AlertTime    time.Time       `json:"alert_time"`
	HandlerID    uint            `json:"handler_id"`
	HandlerName  string          `json:"handler_name"`
	HandleTime   *time.Time      `json:"handle_time"`
	HandleRemark string          `json:"handle_remark"`
	CreatedAt    time.Time       `json:"created_at"`
	UpdatedAt    time.Time       `json:"updated_at"`
}

type AlertHandleRequest struct {
	Status       string `json:"status" binding:"required,oneof=processing resolved ignored"`
	HandleRemark string `json:"handle_remark"`
	HandlerName  string `json:"handler_name"`
}

type PlatformAuthDTO struct {
	StoreID      uint   `json:"store_id" binding:"required"`
	Platform     string `json:"platform" binding:"required,oneof=dianping meituan"`
	StoreUrl     string `json:"store_url"`
	ShopID       string `json:"shop_id"`
	AuthToken    string `json:"auth_token"`
	RefreshToken string `json:"refresh_token"`
}

type PlatformAuthResponse struct {
	ID           uint       `json:"id"`
	StoreID      uint       `json:"store_id"`
	Platform     string     `json:"platform"`
	StoreUrl     string     `json:"store_url"`
	AuthToken    string     `json:"auth_token"`
	RefreshToken string     `json:"refresh_token"`
	ExpireTime   *time.Time `json:"expire_time"`
	ShopID       string     `json:"shop_id"`
	Status       int        `json:"status"`
	LastSyncTime *time.Time `json:"last_sync_time"`
	SyncStatus   string     `json:"sync_status"`
	SyncError    string     `json:"sync_error"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

type SyncRequest struct {
	StoreID  uint   `json:"store_id" binding:"required"`
	Platform string `json:"platform" binding:"required,oneof=dianping meituan"`
}

type SyncStatusResponse struct {
	StoreID      uint       `json:"store_id"`
	Platform     string     `json:"platform"`
	SyncStatus   string     `json:"sync_status"`
	SyncError    string     `json:"sync_error"`
	LastSyncTime *time.Time `json:"last_sync_time"`
}
