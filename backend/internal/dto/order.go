package dto

import (
	"time"

	"github.com/shopspring/decimal"
)

type OrderItemDTO struct {
	ProductID       uint            `json:"product_id" binding:"required"`
	SKUID           uint            `json:"sku_id" binding:"required"`
	ProductName     string          `json:"product_name"`
	SKUName         string          `json:"sku_name"`
	AttributeValues []AttributeValue `json:"attribute_values"`
	Price           decimal.Decimal `json:"price" binding:"required"`
	Quantity        int             `json:"quantity" binding:"required,min=1"`
}

type CreateOrderRequest struct {
	StoreID        uint            `json:"store_id" binding:"required"`
	MemberID       uint            `json:"member_id"`
	TableNo        string          `json:"table_no"`
	OrderType      string          `json:"order_type" binding:"required,oneof=dine_in takeout delivery"`
	Items          []OrderItemDTO  `json:"items" binding:"required,min=1"`
	CouponID       uint            `json:"coupon_id"`
	MemberCouponID uint            `json:"member_coupon_id"`
	PointsUsed     int             `json:"points_used"`
	Remark         string          `json:"remark"`
	Source         string          `json:"source"`
}

type CreateOrderResponse struct {
	OrderID     uint            `json:"order_id"`
	OrderNo     string          `json:"order_no"`
	PayAmount   decimal.Decimal `json:"pay_amount"`
}

type OrderQuery struct {
	PageQuery
	StoreID     uint   `form:"store_id"`
	MemberID    uint   `form:"member_id"`
	OrderStatus int    `form:"order_status"`
	PayStatus   int    `form:"pay_status"`
	OrderType   string `form:"order_type"`
	OrderNo     string `form:"order_no"`
	StartDate   string `form:"start_date"`
	EndDate     string `form:"end_date"`
}

type OrderListResponse struct {
	ID             uint            `json:"id"`
	OrderNo        string          `json:"order_no"`
	StoreID        uint            `json:"store_id"`
	StoreName      string          `json:"store_name"`
	MemberID       uint            `json:"member_id"`
	MemberName     string          `json:"member_name"`
	TableNo        string          `json:"table_no"`
	OrderType      string          `json:"order_type"`
	TotalAmount    decimal.Decimal `json:"total_amount"`
	DiscountAmount decimal.Decimal `json:"discount_amount"`
	CouponAmount   decimal.Decimal `json:"coupon_amount"`
	PayAmount      decimal.Decimal `json:"pay_amount"`
	PayStatus      int             `json:"pay_status"`
	OrderStatus    int             `json:"order_status"`
	ItemCount      int             `json:"item_count"`
	Remark         string          `json:"remark"`
	Source         string          `json:"source"`
	CreatedAt      time.Time       `json:"created_at"`
}

type OrderDetailResponse struct {
	ID              uint                  `json:"id"`
	OrderNo         string                `json:"order_no"`
	StoreID         uint                  `json:"store_id"`
	StoreName       string                `json:"store_name"`
	MemberID        uint                  `json:"member_id"`
	MemberName      string                `json:"member_name"`
	TableNo         string                `json:"table_no"`
	OrderType       string                `json:"order_type"`
	TotalAmount     decimal.Decimal       `json:"total_amount"`
	DiscountAmount  decimal.Decimal       `json:"discount_amount"`
	CouponAmount    decimal.Decimal       `json:"coupon_amount"`
	PayAmount       decimal.Decimal       `json:"pay_amount"`
	PayMethod       string                `json:"pay_method"`
	PayStatus       int                   `json:"pay_status"`
	PayTime         *time.Time            `json:"pay_time"`
	OrderStatus     int                   `json:"order_status"`
	PrintStatus     int                   `json:"print_status"`
	PointsEarned    int                   `json:"points_earned"`
	PointsUsed      int                   `json:"points_used"`
	CouponID        uint                  `json:"coupon_id"`
	MemberCouponID  uint                  `json:"member_coupon_id"`
	Remark          string                `json:"remark"`
	Source          string                `json:"source"`
	Items           []OrderItemDetail     `json:"items"`
	Refunds         []OrderRefundDetail   `json:"refunds,omitempty"`
	CreatedAt       time.Time             `json:"created_at"`
}

type OrderItemDetail struct {
	ID              uint            `json:"id"`
	ProductID       uint            `json:"product_id"`
	SKUID           uint            `json:"sku_id"`
	CategoryID      uint            `json:"category_id"`
	CategoryName    string          `json:"category_name"`
	ProductName     string          `json:"product_name"`
	SKUName         string          `json:"sku_name"`
	AttributeValues string          `json:"attribute_values"`
	Image           string          `json:"image"`
	Price           decimal.Decimal `json:"price"`
	Quantity        int             `json:"quantity"`
	Subtotal        decimal.Decimal `json:"subtotal"`
	Status          int             `json:"status"`
	PrintStatus     int             `json:"print_status"`
	CookStatus      int             `json:"cook_status"`
}

type OrderRefundDetail struct {
	ID            uint            `json:"id"`
	RefundNo      string          `json:"refund_no"`
	RefundAmount  decimal.Decimal `json:"refund_amount"`
	RefundReason  string          `json:"refund_reason"`
	RefundType    string          `json:"refund_type"`
	RefundStatus  int             `json:"refund_status"`
	RefundTime    *time.Time      `json:"refund_time"`
	Remark        string          `json:"remark"`
}

type UpdateOrderStatusRequest struct {
	OrderStatus int `json:"order_status" binding:"required,oneof=1 2 3 4 5 -1"`
}

type CancelOrderRequest struct {
	Reason string `json:"reason" binding:"required"`
}

type RefundOrderRequest struct {
	RefundType   string                  `json:"refund_type" binding:"required,oneof=full partial"`
	RefundAmount decimal.Decimal         `json:"refund_amount" binding:"required"`
	RefundReason string                  `json:"refund_reason" binding:"required"`
	Items        []RefundItemRequest     `json:"items"`
}

type RefundItemRequest struct {
	OrderItemID uint `json:"order_item_id" binding:"required"`
	Quantity    int  `json:"quantity" binding:"required,min=1"`
}

type PaymentParamsRequest struct {
	OrderID uint   `json:"order_id" binding:"required"`
	PayType string `json:"pay_type" binding:"required,oneof=wechat alipay cash"`
}

type PaymentParamsResponse struct {
	PayType   string            `json:"pay_type"`
	OrderID   uint              `json:"order_id"`
	OrderNo   string            `json:"order_no"`
	Amount    decimal.Decimal   `json:"amount"`
	Params    map[string]string `json:"params"`
}

type BatchOrderRequest struct {
	Orders []CreateOrderRequest `json:"orders" binding:"required,min=1"`
}

type BatchOrderResponse struct {
	SuccessCount int      `json:"success_count"`
	FailCount    int      `json:"fail_count"`
	OrderIDs     []uint   `json:"order_ids"`
	Errors       []string `json:"errors,omitempty"`
}
