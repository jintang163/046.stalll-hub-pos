package model

import (
	"time"

	"github.com/shopspring/decimal"
)

type PlatformReviewRating struct {
	BaseModel
	StoreID        uint            `gorm:"not null;index" json:"store_id"`
	Platform       string          `gorm:"size:20;not null;index" json:"platform"`
	OverallRating  decimal.Decimal `gorm:"type:decimal(3,2);not null" json:"overall_rating"`
	TasteRating    decimal.Decimal `gorm:"type:decimal(3,2)" json:"taste_rating"`
	EnvironmentRating decimal.Decimal `gorm:"type:decimal(3,2)" json:"environment_rating"`
	ServiceRating  decimal.Decimal `gorm:"type:decimal(3,2)" json:"service_rating"`
	ReviewCount    int             `gorm:"default:0" json:"review_count"`
	GoodReviewCount int            `gorm:"default:0" json:"good_review_count"`
	MidReviewCount int             `gorm:"default:0" json:"mid_review_count"`
	BadReviewCount int             `gorm:"default:0" json:"bad_review_count"`
	GoodReviewRate decimal.Decimal `gorm:"type:decimal(5,2)" json:"good_review_rate"`
	RatingDate     string          `gorm:"size:10;not null;index" json:"rating_date"`
	SnapshotTime   time.Time       `gorm:"not null;index" json:"snapshot_time"`
	Store          Store           `gorm:"foreignKey:StoreID" json:"store,omitempty"`
}

type PlatformReview struct {
	BaseModel
	StoreID      uint            `gorm:"not null;index" json:"store_id"`
	Platform     string          `gorm:"size:20;not null;index" json:"platform"`
	PlatformID   string          `gorm:"size:100;uniqueIndex" json:"platform_id"`
	UserNickname string          `gorm:"size:100" json:"user_nickname"`
	UserAvatar   string          `gorm:"size:500" json:"user_avatar"`
	UserLevel    string          `gorm:"size:50" json:"user_level"`
	Rating       decimal.Decimal `gorm:"type:decimal(3,2);not null" json:"rating"`
	TasteRating  decimal.Decimal `gorm:"type:decimal(3,2)" json:"taste_rating"`
	EnvironmentRating decimal.Decimal `gorm:"type:decimal(3,2)" json:"environment_rating"`
	ServiceRating decimal.Decimal `gorm:"type:decimal(3,2)" json:"service_rating"`
	Content      string          `gorm:"type:text" json:"content"`
	Images       string          `gorm:"type:text" json:"images"`
	ReviewTime   time.Time       `gorm:"not null;index" json:"review_time"`
	ReplyContent string          `gorm:"type:text" json:"reply_content"`
	ReplyTime    *time.Time      `json:"reply_time"`
	IsBadReview  bool            `gorm:"default:false;index" json:"is_bad_review"`
	IsReplied    bool            `gorm:"default:false;index" json:"is_replied"`
	IsWorkOrderCreated bool     `gorm:"default:false" json:"is_work_order_created"`
	OrderNo      string          `gorm:"size:50" json:"order_no"`
	PerCapita    decimal.Decimal `gorm:"type:decimal(10,2)" json:"per_capita"`
	Store        Store           `gorm:"foreignKey:StoreID" json:"store,omitempty"`
}

type ReviewWorkOrder struct {
	BaseModel
	StoreID       uint      `gorm:"not null;index" json:"store_id"`
	ReviewID      uint      `gorm:"not null;uniqueIndex" json:"review_id"`
	WorkOrderNo   string    `gorm:"size:50;uniqueIndex;not null" json:"work_order_no"`
	Type          string    `gorm:"size:20;default:bad_review" json:"type"`
	Title         string    `gorm:"size:200;not null" json:"title"`
	Description   string    `gorm:"type:text" json:"description"`
	Priority      string    `gorm:"size:20;default:high" json:"priority"`
	Status        string    `gorm:"size:20;default:pending;index" json:"status"`
	AssigneeID    uint      `gorm:"index" json:"assignee_id"`
	AssigneeName  string    `gorm:"size:50" json:"assignee_name"`
	HandlerID     uint      `json:"handler_id"`
	HandlerName   string    `gorm:"size:50" json:"handler_name"`
	HandleTime    *time.Time `json:"handle_time"`
	HandleResult  string    `gorm:"type:text" json:"handle_result"`
	DueTime       *time.Time `json:"due_time"`
	Review        PlatformReview `gorm:"foreignKey:ReviewID" json:"review,omitempty"`
	Store         Store     `gorm:"foreignKey:StoreID" json:"store,omitempty"`
}

type RatingAlert struct {
	BaseModel
	StoreID      uint            `gorm:"not null;index" json:"store_id"`
	Platform     string          `gorm:"size:20;not null" json:"platform"`
	AlertType    string          `gorm:"size:50;not null;index" json:"alert_type"`
	Title        string          `gorm:"size:200;not null" json:"title"`
	Content      string          `gorm:"type:text" json:"content"`
	PrevRating   decimal.Decimal `gorm:"type:decimal(3,2)" json:"prev_rating"`
	CurrRating   decimal.Decimal `gorm:"type:decimal(3,2)" json:"curr_rating"`
	RatingDrop   decimal.Decimal `gorm:"type:decimal(3,2)" json:"rating_drop"`
	Threshold    decimal.Decimal `gorm:"type:decimal(3,2)" json:"threshold"`
	Status       string          `gorm:"size:20;default:unhandled;index" json:"status"`
	AlertTime    time.Time       `gorm:"not null;index" json:"alert_time"`
	HandlerID    uint            `json:"handler_id"`
	HandlerName  string          `gorm:"size:50" json:"handler_name"`
	HandleTime   *time.Time      `json:"handle_time"`
	HandleRemark string          `gorm:"type:text" json:"handle_remark"`
	Store        Store           `gorm:"foreignKey:StoreID" json:"store,omitempty"`
}

type StorePlatformAuth struct {
	BaseModel
	StoreID         uint      `gorm:"not null;uniqueIndex:idx_store_platform" json:"store_id"`
	Platform        string    `gorm:"size:20;not null;uniqueIndex:idx_store_platform" json:"platform"`
	StoreUrl        string    `gorm:"size:500" json:"store_url"`
	AuthToken       string    `gorm:"size:500" json:"auth_token"`
	RefreshToken    string    `gorm:"size:500" json:"refresh_token"`
	ExpireTime      *time.Time `json:"expire_time"`
	ShopID          string    `gorm:"size:100" json:"shop_id"`
	Status          int       `gorm:"default:0;index" json:"status"`
	LastSyncTime    *time.Time `json:"last_sync_time"`
	SyncStatus      string    `gorm:"size:20;default:pending" json:"sync_status"`
	SyncError       string    `gorm:"size:500" json:"sync_error"`
	Store           Store     `gorm:"foreignKey:StoreID" json:"store,omitempty"`
}
