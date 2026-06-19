package dto

import (
	"time"

	"github.com/shopspring/decimal"
)

type CouponCreateDTO struct {
	RuleKey           string          `json:"rule_key" binding:"max=50"`
	Name              string          `json:"name" binding:"required,max=100"`
	Type              string          `json:"type" binding:"required,oneof=fixed percentage"`
	Value             decimal.Decimal `json:"value" binding:"required,min=0.01"`
	MinAmount         decimal.Decimal `json:"min_amount" binding:"min=0"`
	MaxDiscount       decimal.Decimal `json:"max_discount" binding:"min=0"`
	TotalCount        int             `json:"total_count" binding:"min=1"`
	PerUserLimit      int             `json:"per_user_limit" binding:"min=1"`
	ValidityType      string          `json:"validity_type" binding:"required,oneof=fixed relative"`
	ValidityDays      int             `json:"validity_days" binding:"min=1"`
	StartTime         *time.Time      `json:"start_time"`
	EndTime           *time.Time      `json:"end_time"`
	ApplicableType    string          `json:"applicable_type" binding:"required,oneof=all category product"`
	ApplicableIDs     []uint          `json:"applicable_ids"`
	ExcludeProducts   []uint          `json:"exclude_products"`
	Stackable         bool            `json:"stackable"`
	Description       string          `json:"description" binding:"max=500"`
	Status            int             `json:"status" binding:"oneof=0 1"`
}

type CouponUpdateDTO struct {
	RuleKey           string          `json:"rule_key" binding:"max=50"`
	Name              string          `json:"name" binding:"max=100"`
	Value             decimal.Decimal `json:"value" binding:"min=0.01"`
	MinAmount         decimal.Decimal `json:"min_amount" binding:"min=0"`
	MaxDiscount       decimal.Decimal `json:"max_discount" binding:"min=0"`
	TotalCount        int             `json:"total_count" binding:"min=1"`
	PerUserLimit      int             `json:"per_user_limit" binding:"min=1"`
	ValidityType      string          `json:"validity_type" binding:"oneof=fixed relative"`
	ValidityDays      int             `json:"validity_days" binding:"min=1"`
	StartTime         *time.Time      `json:"start_time"`
	EndTime           *time.Time      `json:"end_time"`
	ApplicableType    string          `json:"applicable_type" binding:"oneof=all category product"`
	ApplicableIDs     []uint          `json:"applicable_ids"`
	ExcludeProducts   []uint          `json:"exclude_products"`
	Stackable         *bool           `json:"stackable"`
	Description       string          `json:"description" binding:"max=500"`
	Status            int             `json:"status" binding:"oneof=0 1"`
}

type CouponQueryDTO struct {
	PageQuery
	Name     string `form:"name"`
	Type     string `form:"type"`
	Status   int    `form:"status"`
	StoreID  uint   `form:"store_id"`
}

type CouponResponse struct {
	ID                uint            `json:"id"`
	RuleKey           string          `json:"rule_key"`
	Name              string          `json:"name"`
	Type              string          `json:"type"`
	Value             decimal.Decimal `json:"value"`
	MinAmount         decimal.Decimal `json:"min_amount"`
	MaxDiscount       decimal.Decimal `json:"max_discount"`
	TotalCount        int             `json:"total_count"`
	UsedCount         int             `json:"used_count"`
	PerUserLimit      int             `json:"per_user_limit"`
	ValidityType      string          `json:"validity_type"`
	ValidityDays      int             `json:"validity_days"`
	StartTime         *time.Time      `json:"start_time"`
	EndTime           *time.Time      `json:"end_time"`
	ApplicableType    string          `json:"applicable_type"`
	ApplicableIDs     []uint          `json:"applicable_ids"`
	ExcludeProducts   []uint          `json:"exclude_products"`
	Stackable         bool            `json:"stackable"`
	Description       string          `json:"description"`
	Status            int             `json:"status"`
	CreatedAt         time.Time       `json:"created_at"`
}

type MemberCouponResponse struct {
	ID         uint            `json:"id"`
	MemberID   uint            `json:"member_id"`
	CouponID   uint            `json:"coupon_id"`
	Coupon     CouponResponse  `json:"coupon"`
	Code       string          `json:"code"`
	Status     int             `json:"status"`
	UsedAt     *time.Time      `json:"used_at"`
	ExpireAt   *time.Time      `json:"expire_at"`
	OrderID    uint            `json:"order_id"`
	CreatedAt  time.Time       `json:"created_at"`
}

type IssueCouponDTO struct {
	CouponID    uint   `json:"coupon_id" binding:"required"`
	MemberIDs   []uint `json:"member_ids" binding:"required,min=1"`
	SendMessage bool   `json:"send_message"`
}

type IssueCouponResult struct {
	SuccessCount int      `json:"success_count"`
	FailCount    int      `json:"fail_count"`
	Errors       []string `json:"errors,omitempty"`
}

type VerifyCouponDTO struct {
	CouponID  uint            `json:"coupon_id" binding:"required"`
	MemberID  uint            `json:"member_id" binding:"required"`
	StoreID   uint            `json:"store_id"`
	Amount    decimal.Decimal `json:"amount" binding:"required,min=0"`
	ProductIDs []uint         `json:"product_ids"`
}

type VerifyCouponResponse struct {
	Valid          bool            `json:"valid"`
	DiscountAmount decimal.Decimal `json:"discount_amount"`
	Message        string          `json:"message"`
}

type MemberCouponQueryDTO struct {
	PageQuery
	MemberID uint `form:"member_id"`
	Status   int  `form:"status"`
}
