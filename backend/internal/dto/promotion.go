package dto

import (
	"time"

	"github.com/shopspring/decimal"
)

type PromotionTierDTO struct {
	MinAmount      decimal.Decimal `json:"min_amount" binding:"required,min=0"`
	DiscountAmount decimal.Decimal `json:"discount_amount" binding:"required,min=0.01"`
}

type PromotionCreateDTO struct {
	RuleKey        string             `json:"rule_key" binding:"max=50"`
	Name           string             `json:"name" binding:"required,max=100"`
	Type           string             `json:"type" binding:"required,oneof=full_reduction discount tiered"`
	MinAmount      decimal.Decimal    `json:"min_amount" binding:"min=0"`
	DiscountAmount decimal.Decimal    `json:"discount_amount" binding:"min=0"`
	DiscountRate   decimal.Decimal    `json:"discount_rate" binding:"min=0,max=100"`
	MaxDiscount    decimal.Decimal    `json:"max_discount" binding:"min=0"`
	ApplicableType string             `json:"applicable_type" binding:"required,oneof=all category product"`
	ApplicableIDs  []uint             `json:"applicable_ids"`
	StartTime      *time.Time         `json:"start_time"`
	EndTime        *time.Time         `json:"end_time"`
	Tiers          []PromotionTierDTO `json:"tiers"`
	Priority       int                `json:"priority"`
	Stackable      bool               `json:"stackable"`
	Status         int                `json:"status" binding:"oneof=0 1"`
	Description    string             `json:"description" binding:"max=500"`
}

type PromotionUpdateDTO struct {
	RuleKey        string             `json:"rule_key" binding:"max=50"`
	Name           string             `json:"name" binding:"max=100"`
	MinAmount      decimal.Decimal    `json:"min_amount" binding:"min=0"`
	DiscountAmount decimal.Decimal    `json:"discount_amount" binding:"min=0"`
	DiscountRate   decimal.Decimal    `json:"discount_rate" binding:"min=0,max=100"`
	MaxDiscount    decimal.Decimal    `json:"max_discount" binding:"min=0"`
	ApplicableType string             `json:"applicable_type" binding:"oneof=all category product"`
	ApplicableIDs  []uint             `json:"applicable_ids"`
	StartTime      *time.Time         `json:"start_time"`
	EndTime        *time.Time         `json:"end_time"`
	Tiers          []PromotionTierDTO `json:"tiers"`
	Priority       *int               `json:"priority"`
	Stackable      *bool              `json:"stackable"`
	Status         int                `json:"status" binding:"oneof=0 1"`
	Description    string             `json:"description" binding:"max=500"`
}

type PromotionQueryDTO struct {
	PageQuery
	Name    string `form:"name"`
	Type    string `form:"type"`
	Status  int    `form:"status"`
	StoreID uint   `form:"store_id"`
}

type PromotionResponse struct {
	ID             uint                `json:"id"`
	StoreID        uint                `json:"store_id"`
	RuleKey        string              `json:"rule_key"`
	Name           string              `json:"name"`
	Type           string              `json:"type"`
	MinAmount      decimal.Decimal     `json:"min_amount"`
	DiscountAmount decimal.Decimal     `json:"discount_amount"`
	DiscountRate   decimal.Decimal     `json:"discount_rate"`
	MaxDiscount    decimal.Decimal     `json:"max_discount"`
	ApplicableType string              `json:"applicable_type"`
	ApplicableIDs  []uint              `json:"applicable_ids"`
	StartTime      *time.Time          `json:"start_time"`
	EndTime        *time.Time          `json:"end_time"`
	Tiers          []PromotionTierDTO  `json:"tiers"`
	Priority       int                 `json:"priority"`
	Stackable      bool                `json:"stackable"`
	Status         int                 `json:"status"`
	Description    string              `json:"description"`
	CreatedAt      time.Time           `json:"created_at"`
}

type PromotionCalcRequest struct {
	StoreID         uint            `json:"store_id" binding:"required"`
	Amount          decimal.Decimal `json:"amount" binding:"required,min=0"`
	ProductIDs      []uint          `json:"product_ids"`
	MemberCouponID  uint            `json:"member_coupon_id"`
	MemberID        uint            `json:"member_id"`
}

type PromotionCalcResult struct {
	PromotionID   uint            `json:"promotion_id"`
	CouponID      uint            `json:"coupon_id"`
	Name          string          `json:"name"`
	Type          string          `json:"type"`
	Discount      decimal.Decimal `json:"discount"`
}

type BestPromotionResponse struct {
	Promotions    []PromotionCalcResult `json:"promotions"`
	TotalDiscount decimal.Decimal       `json:"total_discount"`
	FinalAmount   decimal.Decimal       `json:"final_amount"`
}

type MemberClaimCouponDTO struct {
	CouponID uint `json:"coupon_id" binding:"required"`
}

type ClaimableCouponResponse struct {
	ID              uint            `json:"id"`
	StoreID         uint            `json:"store_id"`
	RuleKey         string          `json:"rule_key"`
	Name            string          `json:"name"`
	Type            string          `json:"type"`
	Value           decimal.Decimal `json:"value"`
	MinAmount       decimal.Decimal `json:"min_amount"`
	DiscountRate    decimal.Decimal `json:"discount_rate"`
	MaxDiscount     decimal.Decimal `json:"max_discount"`
	TotalCount      int             `json:"total_count"`
	UsedCount       int             `json:"used_count"`
	PerUserLimit    int             `json:"per_user_limit"`
	ValidityType    string          `json:"validity_type"`
	ValidityDays    int             `json:"validity_days"`
	StartTime       *time.Time      `json:"start_time"`
	EndTime         *time.Time      `json:"end_time"`
	ApplicableType  string          `json:"applicable_type"`
	ApplicableIDs   []uint          `json:"applicable_ids"`
	Stackable       bool            `json:"stackable"`
	Status          int             `json:"status"`
	Description     string          `json:"description"`
	ExchangeProductID uint          `json:"exchange_product_id"`
	RemainingCount  int             `json:"remaining_count"`
	ClaimedCount    int             `json:"claimed_count"`
	CanClaim        bool            `json:"can_claim"`
}

type AvailableCouponQuery struct {
	StoreID    uint   `form:"store_id"`
	Amount     string `form:"amount"`
	ProductIDs string `form:"product_ids"`
}
