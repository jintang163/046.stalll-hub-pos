package model

import (
	"time"

	"github.com/shopspring/decimal"
)

type DailyReport struct {
	BaseModel
	StoreID          uint            `gorm:"not null;index" json:"store_id"`
	ReportDate       string          `gorm:"size:10;not null;index" json:"report_date"`
	OrderCount       int             `gorm:"default:0" json:"order_count"`
	TotalAmount      decimal.Decimal `gorm:"type:decimal(12,2);default:0" json:"total_amount"`
	DiscountAmount   decimal.Decimal `gorm:"type:decimal(12,2);default:0" json:"discount_amount"`
	CouponAmount     decimal.Decimal `gorm:"type:decimal(12,2);default:0" json:"coupon_amount"`
	RefundAmount     decimal.Decimal `gorm:"type:decimal(12,2);default:0" json:"refund_amount"`
	NetAmount        decimal.Decimal `gorm:"type:decimal(12,2);default:0" json:"net_amount"`
	WechatPayAmount  decimal.Decimal `gorm:"type:decimal(12,2);default:0" json:"wechat_pay_amount"`
	AlipayAmount     decimal.Decimal `gorm:"type:decimal(12,2);default:0" json:"alipay_amount"`
	CashAmount       decimal.Decimal `gorm:"type:decimal(12,2);default:0" json:"cash_amount"`
	MemberCount      int             `gorm:"default:0" json:"member_count"`
	PointsEarned     int             `gorm:"default:0" json:"points_earned"`
	PointsUsed       int             `gorm:"default:0" json:"points_used"`
	CouponUsedCount  int             `gorm:"default:0" json:"coupon_used_count"`
	NewMemberCount   int             `gorm:"default:0" json:"new_member_count"`
	Store            Store           `gorm:"foreignKey:StoreID" json:"store,omitempty"`
}

type ProductSalesReport struct {
	BaseModel
	StoreID        uint            `gorm:"not null;index" json:"store_id"`
	ProductID      uint            `gorm:"not null;index" json:"product_id"`
	SKUID          uint            `gorm:"not null;index" json:"sku_id"`
	ReportDate     string          `gorm:"size:10;not null;index" json:"report_date"`
	Quantity       int             `gorm:"default:0" json:"quantity"`
	Amount         decimal.Decimal `gorm:"type:decimal(12,2);default:0" json:"amount"`
	RefundQuantity int             `gorm:"default:0" json:"refund_quantity"`
	RefundAmount   decimal.Decimal `gorm:"type:decimal(12,2);default:0" json:"refund_amount"`
	Product        Product         `gorm:"foreignKey:ProductID" json:"product,omitempty"`
	SKU            ProductSKU      `gorm:"foreignKey:SKUID" json:"sku,omitempty"`
	Store          Store           `gorm:"foreignKey:StoreID" json:"store,omitempty"`
}

type CategorySalesReport struct {
	BaseModel
	StoreID    uint            `gorm:"not null;index" json:"store_id"`
	CategoryID uint            `gorm:"not null;index" json:"category_id"`
	ReportDate string          `gorm:"size:10;not null;index" json:"report_date"`
	OrderCount int             `gorm:"default:0" json:"order_count"`
	Quantity   int             `gorm:"default:0" json:"quantity"`
	Amount     decimal.Decimal `gorm:"type:decimal(12,2);default:0" json:"amount"`
	Category   Category        `gorm:"foreignKey:CategoryID" json:"category,omitempty"`
	Store      Store           `gorm:"foreignKey:StoreID" json:"store,omitempty"`
}

type HourlyReport struct {
	BaseModel
	StoreID       uint            `gorm:"not null;index" json:"store_id"`
	ReportDate    string          `gorm:"size:10;not null" json:"report_date"`
	Hour          int             `gorm:"not null" json:"hour"`
	OrderCount    int             `gorm:"default:0" json:"order_count"`
	TotalAmount   decimal.Decimal `gorm:"type:decimal(12,2);default:0" json:"total_amount"`
	Store         Store           `gorm:"foreignKey:StoreID" json:"store,omitempty"`
}

type PaymentReport struct {
	BaseModel
	StoreID     uint            `gorm:"not null;index" json:"store_id"`
	ReportDate  string          `gorm:"size:10;not null;index" json:"report_date"`
	PayMethod   string          `gorm:"size:20;not null" json:"pay_method"`
	OrderCount  int             `gorm:"default:0" json:"order_count"`
	Amount      decimal.Decimal `gorm:"type:decimal(12,2);default:0" json:"amount"`
	RefundCount int             `gorm:"default:0" json:"refund_count"`
	RefundAmount decimal.Decimal `gorm:"type:decimal(12,2);default:0" json:"refund_amount"`
	Store       Store           `gorm:"foreignKey:StoreID" json:"store,omitempty"`
}

type ReportTask struct {
	BaseModel
	StoreID      uint       `gorm:"not null;index" json:"store_id"`
	ReportType   string     `gorm:"size:20;not null" json:"report_type"`
	StartDate    string     `gorm:"size:10;not null" json:"start_date"`
	EndDate      string     `gorm:"size:10;not null" json:"end_date"`
	Status       int        `gorm:"default:0" json:"status"`
	FileURL      string     `gorm:"size:255" json:"file_url"`
	GeneratedAt  *time.Time `json:"generated_at"`
	Store        Store      `gorm:"foreignKey:StoreID" json:"store,omitempty"`
}
