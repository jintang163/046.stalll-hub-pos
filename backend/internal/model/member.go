package model

import (
	"time"

	"github.com/shopspring/decimal"
)

type Member struct {
	BaseModel
	StoreID        uint            `gorm:"not null;index" json:"store_id"`
	MemberNo       string          `gorm:"size:32;unique;not null" json:"member_no"`
	Name           string          `gorm:"size:50" json:"name"`
	Phone          string          `gorm:"size:20;uniqueIndex" json:"phone"`
	OpenID         string          `gorm:"size:100" json:"open_id"`
	UnionID        string          `gorm:"size:100" json:"union_id"`
	Avatar         string          `gorm:"size:255" json:"avatar"`
	LevelID        uint            `gorm:"default:1" json:"level_id"`
	Points         int             `gorm:"default:0" json:"points"`
	TotalPoints    int             `gorm:"default:0" json:"total_points"`
	Balance        decimal.Decimal `gorm:"type:decimal(10,2);default:0" json:"balance"`
	TotalConsume   decimal.Decimal `gorm:"type:decimal(10,2);default:0" json:"total_consume"`
	OrderCount     int             `gorm:"default:0" json:"order_count"`
	Status         int             `gorm:"default:1" json:"status"`
	Birthday       *time.Time      `json:"birthday"`
	Gender         int             `gorm:"default:0" json:"gender"`
	Store          Store           `gorm:"foreignKey:StoreID" json:"store,omitempty"`
	Level          MemberLevel     `gorm:"foreignKey:LevelID" json:"level,omitempty"`
}

type MemberLevel struct {
	BaseModel
	StoreID         uint            `gorm:"not null;index" json:"store_id"`
	Name            string          `gorm:"size:50;not null" json:"name"`
	MinPoints       int             `gorm:"default:0" json:"min_points"`
	DiscountRate    decimal.Decimal `gorm:"type:decimal(5,2);default:100" json:"discount_rate"`
	PointsRate      decimal.Decimal `gorm:"type:decimal(5,2);default:1" json:"points_rate"`
	Description     string          `gorm:"size:255" json:"description"`
	Color           string          `gorm:"size:20" json:"color"`
	Store           Store           `gorm:"foreignKey:StoreID" json:"store,omitempty"`
}

type MemberPointsRecord struct {
	BaseModel
	StoreID     uint   `gorm:"not null;index" json:"store_id"`
	MemberID    uint   `gorm:"not null;index" json:"member_id"`
	OrderID     uint   `json:"order_id"`
	Type        string `gorm:"size:20;not null" json:"type"`
	Points      int    `gorm:"not null" json:"points"`
	Balance     int    `gorm:"not null" json:"balance"`
	Remark      string `gorm:"size:255" json:"remark"`
	Store       Store  `gorm:"foreignKey:StoreID" json:"store,omitempty"`
	Member      Member `gorm:"foreignKey:MemberID" json:"member,omitempty"`
	Order       *Order `gorm:"foreignKey:OrderID" json:"order,omitempty"`
}

type Coupon struct {
	BaseModel
	StoreID         uint            `gorm:"not null;index" json:"store_id"`
	Name            string          `gorm:"size:100;not null" json:"name"`
	Type            string          `gorm:"size:20;not null" json:"type"`
	Value           decimal.Decimal `gorm:"type:decimal(10,2);not null" json:"value"`
	MinConsume      decimal.Decimal `gorm:"type:decimal(10,2);default:0" json:"min_consume"`
	DiscountRate    decimal.Decimal `gorm:"type:decimal(5,2)" json:"discount_rate"`
	MaxDiscount     decimal.Decimal `gorm:"type:decimal(10,2)" json:"max_discount"`
	TotalCount      int             `gorm:"default:0" json:"total_count"`
	UsedCount       int             `gorm:"default:0" json:"used_count"`
	PerUserLimit    int             `gorm:"default:1" json:"per_user_limit"`
	ValidType       string          `gorm:"size:20;default:fixed" json:"valid_type"`
	ValidDays       int             `gorm:"default:0" json:"valid_days"`
	StartTime       *time.Time      `json:"start_time"`
	EndTime         *time.Time      `json:"end_time"`
	ApplyScope      string          `gorm:"size:20;default:all" json:"apply_scope"`
	ProductIDs      string          `gorm:"size:1000" json:"product_ids"`
	ExcludeProducts string          `gorm:"size:1000" json:"exclude_products"`
	Status          int             `gorm:"default:1" json:"status"`
	Description     string          `gorm:"size:500" json:"description"`
	Store           Store           `gorm:"foreignKey:StoreID" json:"store,omitempty"`
}

type MemberCoupon struct {
	BaseModel
	StoreID    uint       `gorm:"not null;index" json:"store_id"`
	MemberID   uint       `gorm:"not null;index" json:"member_id"`
	CouponID   uint       `gorm:"not null;index" json:"coupon_id"`
	OrderID    uint       `json:"order_id"`
	Status     int        `gorm:"default:0" json:"status"`
	UsedTime   *time.Time `json:"used_time"`
	ExpireTime *time.Time `json:"expire_time"`
	Store      Store      `gorm:"foreignKey:StoreID" json:"store,omitempty"`
	Member     Member     `gorm:"foreignKey:MemberID" json:"member,omitempty"`
	Coupon     Coupon     `gorm:"foreignKey:CouponID" json:"coupon,omitempty"`
	Order      *Order     `gorm:"foreignKey:OrderID" json:"order,omitempty"`
}
