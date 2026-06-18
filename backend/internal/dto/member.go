package dto

import (
	"time"

	"github.com/shopspring/decimal"
)

type MemberCreateDTO struct {
	StoreID   uint   `json:"store_id" binding:"required"`
	Name      string `json:"name" binding:"required,max=50"`
	Phone     string `json:"phone" binding:"required,max=20"`
	Gender    int    `json:"gender" binding:"oneof=0 1 2"`
	Birthday  string `json:"birthday" binding:"max=20"`
	Address   string `json:"address" binding:"max=255"`
	Avatar    string `json:"avatar" binding:"max=255"`
	LevelID   uint   `json:"level_id"`
	Status    int    `json:"status" binding:"oneof=0 1"`
	Source    string `json:"source" binding:"max=50"`
	Remark    string `json:"remark" binding:"max=500"`
}

type MemberUpdateDTO struct {
	Name     string `json:"name" binding:"max=50"`
	Phone    string `json:"phone" binding:"max=20"`
	Gender   int    `json:"gender" binding:"oneof=0 1 2"`
	Birthday string `json:"birthday" binding:"max=20"`
	Address  string `json:"address" binding:"max=255"`
	Avatar   string `json:"avatar" binding:"max=255"`
	LevelID  uint   `json:"level_id"`
	Status   int    `json:"status" binding:"oneof=0 1"`
	Remark   string `json:"remark" binding:"max=500"`
}

type MemberQueryDTO struct {
	PageQuery
	StoreID uint   `form:"store_id"`
	Name    string `form:"name"`
	Phone   string `form:"phone"`
	LevelID uint   `form:"level_id"`
	Status  int    `form:"status"`
}

type MemberResponse struct {
	ID            uint            `json:"id"`
	StoreID       uint            `json:"store_id"`
	StoreName     string          `json:"store_name"`
	Name          string          `json:"name"`
	Phone         string          `json:"phone"`
	Gender        int             `json:"gender"`
	Birthday      string          `json:"birthday"`
	Address       string          `json:"address"`
	Avatar        string          `json:"avatar"`
	LevelID       uint            `json:"level_id"`
	LevelName     string          `json:"level_name"`
	LevelDiscount decimal.Decimal `json:"level_discount"`
	Points        int             `json:"points"`
	TotalSpent    decimal.Decimal `json:"total_spent"`
	TotalOrders   int             `json:"total_orders"`
	Status        int             `json:"status"`
	Source        string          `json:"source"`
	Remark        string          `json:"remark"`
	CreatedAt     time.Time       `json:"created_at"`
}

type MemberLevelCreateDTO struct {
	Name              string          `json:"name" binding:"required,max=50"`
	PointsRequired    int             `json:"points_required" binding:"min=0"`
	Discount          decimal.Decimal `json:"discount" binding:"required,min=0.1,max=1"`
	Description       string          `json:"description" binding:"max=200"`
	Color             string          `json:"color" binding:"max=20"`
	Status            int             `json:"status" binding:"oneof=0 1"`
}

type MemberLevelUpdateDTO struct {
	Name              string          `json:"name" binding:"max=50"`
	PointsRequired    int             `json:"points_required" binding:"min=0"`
	Discount          decimal.Decimal `json:"discount" binding:"min=0.1,max=1"`
	Description       string          `json:"description" binding:"max=200"`
	Color             string          `json:"color" binding:"max=20"`
	Status            int             `json:"status" binding:"oneof=0 1"`
}

type MemberLevelResponse struct {
	ID              uint            `json:"id"`
	Name            string          `json:"name"`
	PointsRequired  int             `json:"points_required"`
	Discount        decimal.Decimal `json:"discount"`
	Description     string          `json:"description"`
	Color           string          `json:"color"`
	MemberCount     int             `json:"member_count"`
	Status          int             `json:"status"`
	CreatedAt       time.Time       `json:"created_at"`
}

type PointsRecordQueryDTO struct {
	PageQuery
	MemberID uint   `form:"member_id"`
	StoreID  uint   `form:"store_id"`
	Type     string `form:"type"`
	StartDate string `form:"start_date"`
	EndDate   string `form:"end_date"`
}

type PointsRecordResponse struct {
	ID         uint      `json:"id"`
	MemberID   uint      `json:"member_id"`
	MemberName string    `json:"member_name"`
	StoreID    uint      `json:"store_id"`
	StoreName  string    `json:"store_name"`
	Type       string    `json:"type"`
	Points     int       `json:"points"`
	Balance    int       `json:"balance"`
	OrderID    uint      `json:"order_id"`
	OrderNo    string    `json:"order_no"`
	Remark     string    `json:"remark"`
	CreatedAt  time.Time `json:"created_at"`
}

type MemberLoginDTO struct {
	Phone    string `json:"phone" binding:"required"`
	Password string `json:"password"`
	Code     string `json:"code"`
	StoreID  uint   `json:"store_id"`
}

type MemberLoginResponse struct {
	Token  string         `json:"token"`
	Member MemberResponse `json:"member"`
}

type AdjustPointsDTO struct {
	MemberID uint   `json:"member_id" binding:"required"`
	Points   int    `json:"points" binding:"required"`
	Remark   string `json:"remark" binding:"required"`
}
