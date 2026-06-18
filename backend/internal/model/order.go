package model

import (
	"time"

	"github.com/shopspring/decimal"
)

type Order struct {
	BaseModel
	OrderNo         string          `gorm:"size:32;unique;not null" json:"order_no"`
	StoreID         uint            `gorm:"not null;index" json:"store_id"`
	MemberID        uint            `gorm:"index" json:"member_id"`
	UserID          uint            `json:"user_id"`
	OrderType       string          `gorm:"size:20;default:dine_in" json:"order_type"`
	TableNo         string          `gorm:"size:20" json:"table_no"`
	TotalAmount     decimal.Decimal `gorm:"type:decimal(10,2);not null" json:"total_amount"`
	DiscountAmount  decimal.Decimal `gorm:"type:decimal(10,2);default:0" json:"discount_amount"`
	CouponAmount    decimal.Decimal `gorm:"type:decimal(10,2);default:0" json:"coupon_amount"`
	PayAmount       decimal.Decimal `gorm:"type:decimal(10,2);not null" json:"pay_amount"`
	PayMethod       string          `gorm:"size:20" json:"pay_method"`
	PayStatus       int             `gorm:"default:0" json:"pay_status"`
	PayTime         *time.Time      `json:"pay_time"`
	OrderStatus     int             `gorm:"default:1" json:"order_status"`
	PrintStatus     int             `gorm:"default:0" json:"print_status"`
	PointsEarned    int             `gorm:"default:0" json:"points_earned"`
	PointsUsed      int             `gorm:"default:0" json:"points_used"`
	CouponID        uint            `json:"coupon_id"`
	Remark          string          `gorm:"size:500" json:"remark"`
	Source          string          `gorm:"size:20;default:cashier" json:"source"`
	Store           Store           `gorm:"foreignKey:StoreID" json:"store,omitempty"`
	Member          *Member         `gorm:"foreignKey:MemberID" json:"member,omitempty"`
	Items           []OrderItem     `gorm:"foreignKey:OrderID" json:"items,omitempty"`
	Payments        []OrderPayment  `gorm:"foreignKey:OrderID" json:"payments,omitempty"`
	Refunds         []OrderRefund   `gorm:"foreignKey:OrderID" json:"refunds,omitempty"`
}

type OrderItem struct {
	BaseModel
	OrderID         uint            `gorm:"not null;index" json:"order_id"`
	ProductID       uint            `gorm:"not null" json:"product_id"`
	SKUID           uint            `gorm:"not null" json:"sku_id"`
	ProductName     string          `gorm:"size:100;not null" json:"product_name"`
	SKUName         string          `gorm:"size:100;not null" json:"sku_name"`
	AttributeValues string          `gorm:"size:500" json:"attribute_values"`
	Image           string          `gorm:"size:255" json:"image"`
	Price           decimal.Decimal `gorm:"type:decimal(10,2);not null" json:"price"`
	Quantity        int             `gorm:"not null" json:"quantity"`
	Subtotal        decimal.Decimal `gorm:"type:decimal(10,2);not null" json:"subtotal"`
	Status          int             `gorm:"default:1" json:"status"`
	PrintStatus     int             `gorm:"default:0" json:"print_status"`
	CookStatus      int             `gorm:"default:0" json:"cook_status"`
	Order           Order           `gorm:"foreignKey:OrderID" json:"order,omitempty"`
}

type OrderPayment struct {
	BaseModel
	OrderID       uint            `gorm:"not null;index" json:"order_id"`
	PayMethod     string          `gorm:"size:20;not null" json:"pay_method"`
	Amount        decimal.Decimal `gorm:"type:decimal(10,2);not null" json:"amount"`
	TransactionID string          `gorm:"size:100" json:"transaction_id"`
	PayStatus     int             `gorm:"default:0" json:"pay_status"`
	PayTime       *time.Time      `json:"pay_time"`
	Order         Order           `gorm:"foreignKey:OrderID" json:"order,omitempty"`
}

type OrderRefund struct {
	BaseModel
	OrderID       uint            `gorm:"not null;index" json:"order_id"`
	RefundNo      string          `gorm:"size:32;unique;not null" json:"refund_no"`
	RefundAmount  decimal.Decimal `gorm:"type:decimal(10,2);not null" json:"refund_amount"`
	RefundReason  string          `gorm:"size:500" json:"refund_reason"`
	RefundType    string          `gorm:"size:20;default:partial" json:"refund_type"`
	RefundStatus  int             `gorm:"default:0" json:"refund_status"`
	RefundTime    *time.Time      `json:"refund_time"`
	OperatorID    uint            `json:"operator_id"`
	Remark        string          `gorm:"size:500" json:"remark"`
	Order         Order           `gorm:"foreignKey:OrderID" json:"order,omitempty"`
	Items         []RefundItem    `gorm:"foreignKey:RefundID" json:"items,omitempty"`
}

type RefundItem struct {
	BaseModel
	RefundID   uint `gorm:"not null;index" json:"refund_id"`
	OrderItemID uint `gorm:"not null" json:"order_item_id"`
	Quantity   int  `gorm:"not null" json:"quantity"`
	Refund     OrderRefund `gorm:"foreignKey:RefundID" json:"refund,omitempty"`
	OrderItem  OrderItem   `gorm:"foreignKey:OrderItemID" json:"order_item,omitempty"`
}

type SyncRecord struct {
	BaseModel
	StoreID     uint   `gorm:"not null;index" json:"store_id"`
	SyncType    string `gorm:"size:20;not null" json:"sync_type"`
	LastSyncID  uint   `gorm:"default:0" json:"last_sync_id"`
	SyncStatus  int    `gorm:"default:0" json:"sync_status"`
	TotalCount  int    `gorm:"default:0" json:"total_count"`
	SuccessCount int   `gorm:"default:0" json:"success_count"`
	ErrorMsg    string `gorm:"size:1000" json:"error_msg"`
	Store       Store  `gorm:"foreignKey:StoreID" json:"store,omitempty"`
}

type OrderQueue struct {
	BaseModel
	StoreID     uint   `gorm:"not null;index" json:"store_id"`
	OrderData   string `gorm:"type:text;not null" json:"order_data"`
	Status      int    `gorm:"default:0" json:"status"`
	RetryCount  int    `gorm:"default:0" json:"retry_count"`
	ErrorMsg    string `gorm:"size:1000" json:"error_msg"`
	Store       Store  `gorm:"foreignKey:StoreID" json:"store,omitempty"`
}
