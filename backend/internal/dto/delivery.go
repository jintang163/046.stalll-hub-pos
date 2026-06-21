package dto

import (
	"time"

	"github.com/shopspring/decimal"
)

type DeliveryOrderCreateRequest struct {
	OrderID         uint   `json:"order_id" binding:"required"`
	DeliveryType    string `json:"delivery_type" binding:"required,oneof=self meituan eleme"`
	SenderName      string `json:"sender_name"`
	SenderPhone     string `json:"sender_phone"`
	SenderAddress   string `json:"sender_address"`
	SenderLng       float64 `json:"sender_lng"`
	SenderLat       float64 `json:"sender_lat"`
	ReceiverName    string `json:"receiver_name" binding:"required"`
	ReceiverPhone   string `json:"receiver_phone" binding:"required"`
	ReceiverAddress string `json:"receiver_address" binding:"required"`
	ReceiverLng     float64 `json:"receiver_lng"`
	ReceiverLat     float64 `json:"receiver_lat"`
}

type DeliveryOrderResponse struct {
	ID              uint            `json:"id"`
	OrderID         uint            `json:"order_id"`
	OrderNo         string          `json:"order_no"`
	StoreID         uint            `json:"store_id"`
	DeliveryType    string          `json:"delivery_type"`
	DeliveryStatus  int             `json:"delivery_status"`
	RiderID         uint            `json:"rider_id"`
	RiderName       string          `json:"rider_name"`
	RiderPhone      string          `json:"rider_phone"`
	RiderLng        float64         `json:"rider_lng"`
	RiderLat        float64         `json:"rider_lat"`
	DeliveryFee     decimal.Decimal `json:"delivery_fee"`
	Distance        float64         `json:"distance"`
	Duration        int             `json:"duration"`
	SenderName      string          `json:"sender_name"`
	SenderPhone     string          `json:"sender_phone"`
	SenderAddress   string          `json:"sender_address"`
	ReceiverName    string          `json:"receiver_name"`
	ReceiverPhone   string          `json:"receiver_phone"`
	ReceiverAddress string          `json:"receiver_address"`
	PlatformOrderID string          `json:"platform_order_id"`
	PlatformType    string          `json:"platform_type"`
	EstimatedTime   *time.Time      `json:"estimated_time"`
	PickedUpAt      *time.Time      `json:"picked_up_at"`
	DeliveredAt     *time.Time      `json:"delivered_at"`
	CreatedAt       time.Time       `json:"created_at"`
}

type PickupCodeResponse struct {
	OrderID   uint   `json:"order_id"`
	Code      string `json:"code"`
	Status    int    `json:"status"`
	ExpiredAt string `json:"expired_at,omitempty"`
}

type RiderLocationUpdate struct {
	RiderID uint    `json:"rider_id" binding:"required"`
	Lng     float64 `json:"lng" binding:"required"`
	Lat     float64 `json:"lat" binding:"required"`
	Speed   float64 `json:"speed"`
	Heading float64 `json:"heading"`
}

type RiderLocationResponse struct {
	RiderID   uint    `json:"rider_id"`
	RiderName string  `json:"rider_name"`
	Lng       float64 `json:"lng"`
	Lat       float64 `json:"lat"`
	Speed     float64 `json:"speed"`
	Heading   float64 `json:"heading"`
	UpdatedAt string  `json:"updated_at"`
}

type RoutePlanRequest struct {
	OriginLng float64 `json:"origin_lng" binding:"required"`
	OriginLat float64 `json:"origin_lat" binding:"required"`
	DestLng   float64 `json:"dest_lng" binding:"required"`
	DestLat   float64 `json:"dest_lat" binding:"required"`
}

type RoutePlanResponse struct {
	Distance float64 `json:"distance"`
	Duration int     `json:"duration"`
	Route    string  `json:"route"`
	Fee      decimal.Decimal `json:"fee"`
}

type GeocodeRequest struct {
	Address string `json:"address" binding:"required"`
	City    string `json:"city"`
}

type GeocodeResponse struct {
	Lng       float64 `json:"lng"`
	Lat       float64 `json:"lat"`
	Formatted string  `json:"formatted"`
}

type CreateRiderRequest struct {
	Name  string `json:"name" binding:"required"`
	Phone string `json:"phone" binding:"required"`
}

type DeliveryOrderQuery struct {
	PageQuery
	StoreID        uint   `form:"store_id"`
	DeliveryType   string `form:"delivery_type"`
	DeliveryStatus int    `form:"delivery_status"`
	OrderID        uint   `form:"order_id"`
}

type DeliveryStatusUpdateRequest struct {
	DeliveryStatus int `json:"delivery_status" binding:"required"`
}

type AssignRiderRequest struct {
	RiderID uint `json:"rider_id" binding:"required"`
}

type VerifyPickupCodeRequest struct {
	Code    string `json:"code" binding:"required"`
	StoreID uint   `json:"store_id" binding:"required"`
}

type DeliveryTrackingResponse struct {
	OrderNo         string               `json:"order_no"`
	DeliveryType    string               `json:"delivery_type"`
	DeliveryStatus  int                  `json:"delivery_status"`
	RiderID         uint                 `json:"rider_id"`
	RiderName       string               `json:"rider_name"`
	RiderPhone      string               `json:"rider_phone"`
	RiderLng        float64              `json:"rider_lng"`
	RiderLat        float64              `json:"rider_lat"`
	Distance        float64              `json:"distance"`
	Duration        int                  `json:"duration"`
	ReceiverAddress string               `json:"receiver_address"`
	SenderAddress   string               `json:"sender_address"`
	EstimatedTime   *time.Time           `json:"estimated_time"`
	Trackings       []TrackingPoint      `json:"trackings"`
}

type TrackingPoint struct {
	Lng       float64 `json:"lng"`
	Lat       float64 `json:"lat"`
	Speed     float64 `json:"speed"`
	Timestamp int64   `json:"timestamp"`
}
