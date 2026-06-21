package model

import (
	"time"

	"github.com/shopspring/decimal"
)

type DeliveryOrder struct {
	BaseModel
	OrderID         uint            `gorm:"not null;index" json:"order_id"`
	StoreID         uint            `gorm:"not null;index" json:"store_id"`
	DeliveryType    string          `gorm:"size:20;not null;default:self" json:"delivery_type"`
	DeliveryStatus  int             `gorm:"default:0" json:"delivery_status"`
	RiderID         uint            `gorm:"index" json:"rider_id"`
	RiderName       string          `gorm:"size:50" json:"rider_name"`
	RiderPhone      string          `gorm:"size:20" json:"rider_phone"`
	RiderLng        float64         `json:"rider_lng"`
	RiderLat        float64         `json:"rider_lat"`
	DeliveryFee     decimal.Decimal `gorm:"type:decimal(10,2);default:0" json:"delivery_fee"`
	Distance        float64         `gorm:"default:0" json:"distance"`
	Duration        int             `gorm:"default:0" json:"duration"`
	SenderName      string          `gorm:"size:50" json:"sender_name"`
	SenderPhone     string          `gorm:"size:20" json:"sender_phone"`
	SenderAddress   string          `gorm:"size:255" json:"sender_address"`
	SenderLng       float64         `json:"sender_lng"`
	SenderLat       float64         `json:"sender_lat"`
	ReceiverName    string          `gorm:"size:50" json:"receiver_name"`
	ReceiverPhone   string          `gorm:"size:20" json:"receiver_phone"`
	ReceiverAddress string          `gorm:"size:255" json:"receiver_address"`
	ReceiverLng     float64         `json:"receiver_lng"`
	ReceiverLat     float64         `json:"receiver_lat"`
	PlatformOrderID string          `gorm:"size:100" json:"platform_order_id"`
	PlatformType    string          `gorm:"size:20" json:"platform_type"`
	EstimatedTime   *time.Time      `json:"estimated_time"`
	PickedUpAt      *time.Time      `json:"picked_up_at"`
	DeliveredAt     *time.Time      `json:"delivered_at"`
	RouteData       string          `gorm:"type:text" json:"route_data"`
	Order           Order           `gorm:"foreignKey:OrderID" json:"order,omitempty"`
	Rider           *Rider          `gorm:"foreignKey:RiderID" json:"rider,omitempty"`
}

type PickupCode struct {
	BaseModel
	OrderID     uint      `gorm:"not null;index" json:"order_id"`
	StoreID     uint      `gorm:"not null;index" json:"store_id"`
	Code        string    `gorm:"size:8;unique;not null" json:"code"`
	Status      int       `gorm:"default:0" json:"status"`
	ExpiredAt   time.Time `json:"expired_at"`
	UsedAt      *time.Time `json:"used_at"`
	Order       Order     `gorm:"foreignKey:OrderID" json:"order,omitempty"`
}

type Rider struct {
	BaseModel
	StoreID     uint    `gorm:"not null;index" json:"store_id"`
	Name        string  `gorm:"size:50;not null" json:"name"`
	Phone       string  `gorm:"size:20;not null" json:"phone"`
	Status      int     `gorm:"default:1" json:"status"`
	CurrentLng  float64 `json:"current_lng"`
	CurrentLat  float64 `json:"current_lat"`
	OrderCount  int     `gorm:"default:0" json:"order_count"`
	Store       Store   `gorm:"foreignKey:StoreID" json:"store,omitempty"`
}

type DeliveryTracking struct {
	BaseModel
	DeliveryOrderID uint    `gorm:"not null;index" json:"delivery_order_id"`
	RiderID         uint    `gorm:"index" json:"rider_id"`
	Lng             float64 `json:"lng"`
	Lat             float64 `json:"lat"`
	Speed           float64 `json:"speed"`
	Heading         float64 `json:"heading"`
	Timestamp       int64   `json:"timestamp"`
	DeliveryOrder   DeliveryOrder `gorm:"foreignKey:DeliveryOrderID" json:"delivery_order,omitempty"`
}
