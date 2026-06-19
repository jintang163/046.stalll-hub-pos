package dto

import (
	"time"

	"github.com/shopspring/decimal"
)

type PageQuery struct {
	Page     int `form:"page" json:"page"`
	PageSize int `form:"page_size" json:"page_size"`
}

type PointsRuleCreateDTO struct {
	StoreID          uint            `json:"store_id"`
	RuleKey          string          `json:"rule_key" binding:"required,max=50"`
	RuleName         string          `json:"rule_name" binding:"required,max=100"`
	RuleType         string          `json:"rule_type" binding:"required,oneof=earn redeem register bonus"`
	PointsPerYuan    decimal.Decimal `json:"points_per_yuan" binding:"min=0"`
	RedeemRate       decimal.Decimal `json:"redeem_rate" binding:"min=0"`
	MinRedeemPoints  int             `json:"min_redeem_points" binding:"min=0"`
	BonusPoints      int             `json:"bonus_points" binding:"min=0"`
	MinConsumeAmount decimal.Decimal `json:"min_consume_amount" binding:"min=0"`
	Priority         int             `json:"priority"`
	Status           int             `json:"status" binding:"oneof=0 1"`
}

type PointsRuleUpdateDTO struct {
	RuleName         string          `json:"rule_name" binding:"max=100"`
	PointsPerYuan    decimal.Decimal `json:"points_per_yuan" binding:"min=0"`
	RedeemRate       decimal.Decimal `json:"redeem_rate" binding:"min=0"`
	MinRedeemPoints  int             `json:"min_redeem_points" binding:"min=0"`
	BonusPoints      int             `json:"bonus_points" binding:"min=0"`
	MinConsumeAmount decimal.Decimal `json:"min_consume_amount" binding:"min=0"`
	Priority         int             `json:"priority"`
	Status           int             `json:"status" binding:"oneof=0 1"`
}

type PointsRuleQueryDTO struct {
	PageQuery
	StoreID  uint   `form:"store_id"`
	RuleType string `form:"rule_type"`
	Status   int    `form:"status"`
}

type PointsRuleResponse struct {
	ID               uint            `json:"id"`
	StoreID          uint            `json:"store_id"`
	RuleKey          string          `json:"rule_key"`
	RuleName         string          `json:"rule_name"`
	RuleType         string          `json:"rule_type"`
	PointsPerYuan    decimal.Decimal `json:"points_per_yuan"`
	RedeemRate       decimal.Decimal `json:"redeem_rate"`
	MinRedeemPoints  int             `json:"min_redeem_points"`
	BonusPoints      int             `json:"bonus_points"`
	MinConsumeAmount decimal.Decimal `json:"min_consume_amount"`
	Priority         int             `json:"priority"`
	Status           int             `json:"status"`
	CreatedAt        time.Time       `json:"created_at"`
	UpdatedAt        time.Time       `json:"updated_at"`
}

type PointsEarnRequest struct {
	MemberID     uint            `json:"member_id" binding:"required"`
	OrderID      uint            `json:"order_id"`
	ConsumeAmount decimal.Decimal `json:"consume_amount" binding:"required,min=0.01"`
}

type PointsRedeemRequest struct {
	MemberID uint `json:"member_id" binding:"required"`
	OrderID  uint `json:"order_id"`
	Points   int  `json:"points" binding:"required,min=1"`
}

type PointsCalcResult struct {
	PointsEarned   int             `json:"points_earned"`
	DiscountAmount decimal.Decimal `json:"discount_amount"`
	PointsUsed     int             `json:"points_used"`
}
