package model

import (
	"time"

	"github.com/shopspring/decimal"
)

type Stall struct {
	BaseModel
	StoreID         uint            `gorm:"not null;index" json:"store_id"`
	StallNo         string          `gorm:"size:20;not null;index" json:"stall_no"`
	Name            string          `gorm:"size:100;not null" json:"name"`
	Type            string          `gorm:"size:20;default:normal" json:"type"`
	Description     string          `gorm:"size:500" json:"description"`
	Logo            string          `gorm:"size:255" json:"logo"`
	RevenueRatio    decimal.Decimal `gorm:"type:decimal(5,4);not null;default:0.7000" json:"revenue_ratio"`
	PlatformRatio   decimal.Decimal `gorm:"type:decimal(5,4);not null;default:0.3000" json:"platform_ratio"`
	ContactName     string          `gorm:"size:50" json:"contact_name"`
	ContactPhone    string          `gorm:"size:20" json:"contact_phone"`
	PrinterName     string          `gorm:"size:100" json:"printer_name"`
	SortOrder       int             `gorm:"default:0" json:"sort_order"`
	Status          int             `gorm:"default:1" json:"status"`
	Store           Store           `gorm:"foreignKey:StoreID" json:"store,omitempty"`
}

type StallDevice struct {
	BaseModel
	StoreID      uint      `gorm:"not null;index" json:"store_id"`
	StallID      uint      `gorm:"not null;index" json:"stall_id"`
	DeviceID     string    `gorm:"size:64;not null;uniqueIndex" json:"device_id"`
	DeviceName   string    `gorm:"size:100" json:"device_name"`
	DeviceType   string    `gorm:"size:20;default:mobile" json:"device_type"`
	OSVersion    string    `gorm:"size:50" json:"os_version"`
	AppVersion   string    `gorm:"size:50" json:"app_version"`
	LastOnlineAt *time.Time `json:"last_online_at"`
	LastHeartbeatAt *time.Time `json:"last_heartbeat_at"`
	Status       int       `gorm:"default:1" json:"status"`
	Stall        Stall     `gorm:"foreignKey:StallID" json:"stall,omitempty"`
	Store        Store     `gorm:"foreignKey:StoreID" json:"store,omitempty"`
}

type StallUser struct {
	BaseModel
	StoreID    uint   `gorm:"not null;index" json:"store_id"`
	StallID    uint   `gorm:"not null;index" json:"stall_id"`
	Username   string `gorm:"size:50;not null;uniqueIndex" json:"username"`
	Password   string `gorm:"size:255;not null" json:"-"`
	RealName   string `gorm:"size:50" json:"real_name"`
	Phone      string `gorm:"size:20" json:"phone"`
	Role       string `gorm:"size:20;default:stall_staff" json:"role"`
	Status     int    `gorm:"default:1" json:"status"`
	Stall      Stall  `gorm:"foreignKey:StallID" json:"stall,omitempty"`
	Store      Store  `gorm:"foreignKey:StoreID" json:"store,omitempty"`
}

type StallSettlement struct {
	BaseModel
	StoreID         uint            `gorm:"not null;index" json:"store_id"`
	StallID         uint            `gorm:"not null;index" json:"stall_id"`
	SettlementNo    string          `gorm:"size:32;unique;not null" json:"settlement_no"`
	SettlementDate  string          `gorm:"size:10;not null;index" json:"settlement_date"`
	OrderCount      int             `gorm:"default:0" json:"order_count"`
	TotalAmount     decimal.Decimal `gorm:"type:decimal(12,2);default:0" json:"total_amount"`
	RefundAmount    decimal.Decimal `gorm:"type:decimal(12,2);default:0" json:"refund_amount"`
	NetAmount       decimal.Decimal `gorm:"type:decimal(12,2);default:0" json:"net_amount"`
	StallAmount     decimal.Decimal `gorm:"type:decimal(12,2);default:0" json:"stall_amount"`
	PlatformAmount  decimal.Decimal `gorm:"type:decimal(12,2);default:0" json:"platform_amount"`
	SettlementStatus int            `gorm:"default:0" json:"settlement_status"`
	SettledAt       *time.Time      `json:"settled_at"`
	OperatorID      uint            `json:"operator_id"`
	Remark          string          `gorm:"size:500" json:"remark"`
	Stall           Stall           `gorm:"foreignKey:StallID" json:"stall,omitempty"`
	Store           Store           `gorm:"foreignKey:StoreID" json:"store,omitempty"`
}

type StallSettlementItem struct {
	BaseModel
	SettlementID   uint            `gorm:"not null;index" json:"settlement_id"`
	OrderID        uint            `gorm:"not null;index" json:"order_id"`
	OrderNo        string          `gorm:"size:32;not null;index" json:"order_no"`
	OrderAmount    decimal.Decimal `gorm:"type:decimal(12,2);not null" json:"order_amount"`
	StallAmount    decimal.Decimal `gorm:"type:decimal(12,2);not null" json:"stall_amount"`
	PlatformAmount decimal.Decimal `gorm:"type:decimal(12,2);not null" json:"platform_amount"`
	Settlement     StallSettlement `gorm:"foreignKey:SettlementID" json:"settlement,omitempty"`
}

type StallDailyReport struct {
	BaseModel
	StoreID         uint            `gorm:"not null;index" json:"store_id"`
	StallID         uint            `gorm:"not null;index" json:"stall_id"`
	ReportDate      string          `gorm:"size:10;not null;index" json:"report_date"`
	OrderCount      int             `gorm:"default:0" json:"order_count"`
	TotalAmount     decimal.Decimal `gorm:"type:decimal(12,2);default:0" json:"total_amount"`
	DiscountAmount  decimal.Decimal `gorm:"type:decimal(12,2);default:0" json:"discount_amount"`
	CouponAmount    decimal.Decimal `gorm:"type:decimal(12,2);default:0" json:"coupon_amount"`
	RefundAmount    decimal.Decimal `gorm:"type:decimal(12,2);default:0" json:"refund_amount"`
	NetAmount       decimal.Decimal `gorm:"type:decimal(12,2);default:0" json:"net_amount"`
	StallAmount     decimal.Decimal `gorm:"type:decimal(12,2);default:0" json:"stall_amount"`
	PlatformAmount  decimal.Decimal `gorm:"type:decimal(12,2);default:0" json:"platform_amount"`
	Stall           Stall           `gorm:"foreignKey:StallID" json:"stall,omitempty"`
	Store           Store           `gorm:"foreignKey:StoreID" json:"store,omitempty"`
}
