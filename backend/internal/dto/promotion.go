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
	StoreID    uint            `json:"store_id" binding:"required"`
	Amount     decimal.Decimal `json:"amount" binding:"required,min=0"`
	ProductIDs []uint          `json:"product_ids"`
}

type PromotionCalcResult struct {
	PromotionID    uint            `json:"promotion_id"`
	PromotionName  string          `json:"promotion_name"`
	PromotionType  string          `json:"promotion_type"`
	DiscountAmount decimal.Decimal `json:"discount_amount"`
}

type BestPromotionResponse struct {
	Promotions     []PromotionCalcResult `json:"promotions"`
	TotalDiscount  decimal.Decimal       `json:"total_discount"`
	FinalAmount    decimal.Decimal       `json:"final_amount"`
}

type MemberClaimCouponDTO struct {
	CouponID uint `json:"coupon_id" binding:"required"`
}

type AvailableCouponQuery struct {
	StoreID    uint   `form:"store_id"`
	Amount     string `form:"amount"`
	ProductIDs string `form:"product_ids"`
}
