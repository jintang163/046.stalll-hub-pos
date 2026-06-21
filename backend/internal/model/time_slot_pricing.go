package model

import (
	"time"

	"github.com/shopspring/decimal"
)

type TimeSlotPricing struct {
	BaseModel
	StoreID        uint            `gorm:"not null;index" json:"store_id"`
	Name           string          `gorm:"size:50;not null" json:"name"`
	StartTime      string          `gorm:"size:5;not null" json:"start_time"`
	EndTime        string          `gorm:"size:5;not null" json:"end_time"`
	PricingType    string          `gorm:"size:20;not null;default:discount" json:"pricing_type"`
	DiscountRate   decimal.Decimal `gorm:"type:decimal(5,2);default:100" json:"discount_rate"`
	DiscountAmount decimal.Decimal `gorm:"type:decimal(10,2);default:0" json:"discount_amount"`
	MinAmount      decimal.Decimal `gorm:"type:decimal(10,2);default:0" json:"min_amount"`
	ApplicableType string          `gorm:"size:20;default:all" json:"applicable_type"`
	ApplicableIDs  string          `gorm:"size:500" json:"applicable_ids"`
	Weekdays       string          `gorm:"size:20;default:1,2,3,4,5,6,7" json:"weekdays"`
	Priority       int             `gorm:"default:100" json:"priority"`
	Status         int             `gorm:"default:1" json:"status"`
	Description    string          `gorm:"size:255" json:"description"`
	Store          Store           `gorm:"foreignKey:StoreID" json:"store,omitempty"`
}

type StockReservation struct {
	BaseModel
	StoreID     uint      `gorm:"not null;index" json:"store_id"`
	OrderID     uint      `gorm:"not null;index" json:"order_id"`
	SKUID       uint      `gorm:"not null;index" json:"sku_id"`
	ProductID   uint      `gorm:"not null" json:"product_id"`
	Quantity    int       `gorm:"not null" json:"quantity"`
	ReservedAt  time.Time `gorm:"not null" json:"reserved_at"`
	ExpireAt    time.Time `gorm:"not null;index" json:"expire_at"`
	IsReleased  bool      `gorm:"default:false" json:"is_released"`
	ReleasedAt  *time.Time `json:"released_at"`
	Order       Order     `gorm:"foreignKey:OrderID" json:"order,omitempty"`
	SKU         ProductSKU `gorm:"foreignKey:SKUID" json:"sku,omitempty"`
	Product     Product   `gorm:"foreignKey:ProductID" json:"product,omitempty"`
}

type ReservationReminder struct {
	BaseModel
	StoreID     uint      `gorm:"not null;index" json:"store_id"`
	OrderID     uint      `gorm:"not null;uniqueIndex" json:"order_id"`
	RemindAt    time.Time `gorm:"not null;index" json:"remind_at"`
	IsSent      bool      `gorm:"default:false" json:"is_sent"`
	SentAt      *time.Time `json:"sent_at"`
	Order       Order     `gorm:"foreignKey:OrderID" json:"order,omitempty"`
}
