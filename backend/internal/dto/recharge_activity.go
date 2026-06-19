package dto

import (
	"time"

	"github.com/shopspring/decimal"
)

type RechargeActivityCreateDTO struct {
	StoreID       uint            `json:"store_id"`
	Name          string          `json:"name" binding:"required,max=100"`
	MinAmount     decimal.Decimal `json:"min_amount" binding:"required,min=0.01"`
	BonusAmount   decimal.Decimal `json:"bonus_amount" binding:"min=0"`
	BonusPoints   int             `json:"bonus_points" binding:"min=0"`
	BonusCouponID uint            `json:"bonus_coupon_id"`
	StartTime     *time.Time      `json:"start_time" binding:"required"`
	EndTime       *time.Time      `json:"end_time" binding:"required"`
	AutoActivate  bool            `json:"auto_activate"`
	Status        int             `json:"status" binding:"oneof=0 1"`
	Description   string          `json:"description" binding:"max=500"`
}

type RechargeActivityUpdateDTO struct {
	Name          string          `json:"name" binding:"max=100"`
	MinAmount     decimal.Decimal `json:"min_amount" binding:"min=0.01"`
	BonusAmount   decimal.Decimal `json:"bonus_amount" binding:"min=0"`
	BonusPoints   int             `json:"bonus_points" binding:"min=0"`
	BonusCouponID uint            `json:"bonus_coupon_id"`
	StartTime     *time.Time      `json:"start_time"`
	EndTime       *time.Time      `json:"end_time"`
	AutoActivate  *bool           `json:"auto_activate"`
	Status        int             `json:"status" binding:"oneof=0 1"`
	Description   string          `json:"description" binding:"max=500"`
}

type RechargeActivityQueryDTO struct {
	PageQuery
	StoreID uint `form:"store_id"`
	Status  int  `form:"status"`
}

type RechargeActivityResponse struct {
	ID            uint            `json:"id"`
	StoreID       uint            `json:"store_id"`
	Name          string          `json:"name"`
	MinAmount     decimal.Decimal `json:"min_amount"`
	BonusAmount   decimal.Decimal `json:"bonus_amount"`
	BonusPoints   int             `json:"bonus_points"`
	BonusCouponID uint            `json:"bonus_coupon_id"`
	StartTime     *time.Time      `json:"start_time"`
	EndTime       *time.Time      `json:"end_time"`
	AutoActivate  bool            `json:"auto_activate"`
	Status        int             `json:"status"`
	Description   string          `json:"description"`
	CreatedAt     time.Time       `json:"created_at"`
	UpdatedAt     time.Time       `json:"updated_at"`
}

type MemberRechargeDTO struct {
	MemberID   uint            `json:"member_id" binding:"required"`
	Amount     decimal.Decimal `json:"amount" binding:"required,min=0.01"`
	ActivityID uint            `json:"activity_id"`
	PayMethod  string          `json:"pay_method" binding:"max=20"`
}

type MemberRechargeResponse struct {
	ID          uint            `json:"id"`
	StoreID     uint            `json:"store_id"`
	MemberID    uint            `json:"member_id"`
	Amount      decimal.Decimal `json:"amount"`
	BonusAmount decimal.Decimal `json:"bonus_amount"`
	BonusPoints int             `json:"bonus_points"`
	ActivityID  uint            `json:"activity_id"`
	PayMethod   string          `json:"pay_method"`
	Status      int             `json:"status"`
	CreatedAt   time.Time       `json:"created_at"`
}

type RechargeActivityMatchResult struct {
	ActivityID  uint            `json:"activity_id"`
	ActivityName string         `json:"activity_name"`
	BonusAmount decimal.Decimal `json:"bonus_amount"`
	BonusPoints int             `json:"bonus_points"`
	BonusCouponID uint          `json:"bonus_coupon_id"`
}
