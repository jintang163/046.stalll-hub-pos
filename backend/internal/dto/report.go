package dto

import (
	"time"

	"github.com/shopspring/decimal"
)

type DailyReportQueryDTO struct {
	StoreID   uint   `form:"store_id"`
	StartDate string `form:"start_date" binding:"required"`
	EndDate   string `form:"end_date" binding:"required"`
}

type DailyReportResponse struct {
	ID                 uint            `json:"id"`
	StoreID            uint            `json:"store_id"`
	StoreName          string          `json:"store_name"`
	ReportDate         string          `json:"report_date"`
	TotalOrders        int             `json:"total_orders"`
	TotalAmount        decimal.Decimal `json:"total_amount"`
	PayAmount          decimal.Decimal `json:"pay_amount"`
	RefundAmount       decimal.Decimal `json:"refund_amount"`
	NetAmount          decimal.Decimal `json:"net_amount"`
	DiscountAmount     decimal.Decimal `json:"discount_amount"`
	CouponAmount       decimal.Decimal `json:"coupon_amount"`
	PointsUsed         int             `json:"points_used"`
	PointsEarned       int             `json:"points_earned"`
	NewMembers         int             `json:"new_members"`
	ActiveMembers      int             `json:"active_members"`
	AverageOrderAmount decimal.Decimal `json:"average_order_amount"`
	PeakHourOrders     int             `json:"peak_hour_orders"`
	PeakHour           string          `json:"peak_hour"`
	CanceledOrders     int             `json:"canceled_orders"`
	RefundedOrders     int             `json:"refunded_orders"`
	CreatedAt          time.Time       `json:"created_at"`
}

type ProductSalesReportDTO struct {
	StoreID    uint   `form:"store_id"`
	StartDate  string `form:"start_date" binding:"required"`
	EndDate    string `form:"end_date" binding:"required"`
	CategoryID uint   `form:"category_id"`
	TopN       int    `form:"top_n"`
}

type ProductSalesResponse struct {
	ProductID      uint            `json:"product_id"`
	ProductName    string          `json:"product_name"`
	ProductImage   string          `json:"product_image"`
	CategoryID     uint            `json:"category_id"`
	CategoryName   string          `json:"category_name"`
	TotalQuantity  int             `json:"total_quantity"`
	TotalAmount    decimal.Decimal `json:"total_amount"`
	RefundQuantity int             `json:"refund_quantity"`
	RefundAmount   decimal.Decimal `json:"refund_amount"`
	NetAmount      decimal.Decimal `json:"net_amount"`
	Rank           int             `json:"rank"`
}

type CategorySalesResponse struct {
	CategoryID    uint            `json:"category_id"`
	CategoryName  string          `json:"category_name"`
	ProductCount  int             `json:"product_count"`
	TotalQuantity int             `json:"total_quantity"`
	TotalAmount   decimal.Decimal `json:"total_amount"`
	Percentage    decimal.Decimal `json:"percentage"`
}

type HourlySalesResponse struct {
	Hour          int             `json:"hour"`
	TotalOrders   int             `json:"total_orders"`
	TotalAmount   decimal.Decimal `json:"total_amount"`
	Percentage    decimal.Decimal `json:"percentage"`
}

type PaymentReportDTO struct {
	StoreID   uint   `form:"store_id"`
	StartDate string `form:"start_date" binding:"required"`
	EndDate   string `form:"end_date" binding:"required"`
}

type PaymentSalesResponse struct {
	PayMethod     string          `json:"pay_method"`
	PayMethodName string          `json:"pay_method_name"`
	TotalOrders   int             `json:"total_orders"`
	TotalAmount   decimal.Decimal `json:"total_amount"`
	Percentage    decimal.Decimal `json:"percentage"`
	AverageAmount decimal.Decimal `json:"average_amount"`
}

type OverviewResponse struct {
	StoreID              uint            `json:"store_id"`
	TodayOrders          int             `json:"today_orders"`
	TodayAmount          decimal.Decimal `json:"today_amount"`
	TodayCustomers       int             `json:"today_customers"`
	YesterdayOrders      int             `json:"yesterday_orders"`
	YesterdayAmount      decimal.Decimal `json:"yesterday_amount"`
	YesterdayCustomers   int             `json:"yesterday_customers"`
	MonthOrders          int             `json:"month_orders"`
	MonthAmount          decimal.Decimal `json:"month_amount"`
	MonthCustomers       int             `json:"month_customers"`
	StockWarningCount    int             `json:"stock_warning_count"`
	PendingOrders        int             `json:"pending_orders"`
	ActiveMembers        int             `json:"active_members"`
	AvailableCoupons     int             `json:"available_coupons"`
	TodayHourlySales     []HourlySalesResponse `json:"today_hourly_sales"`
	TopProducts          []ProductSalesResponse `json:"top_products"`
}

type ExportReportDTO struct {
	ReportType string `json:"report_type" binding:"required,oneof=daily product category payment"`
	StoreID    uint   `json:"store_id"`
	StartDate  string `json:"start_date" binding:"required"`
	EndDate    string `json:"end_date" binding:"required"`
	Format     string `json:"format" binding:"required,oneof=excel csv"`
}

type ExportResponse struct {
	DownloadURL string `json:"download_url"`
	FileName    string `json:"file_name"`
}
