package dto

import "github.com/shopspring/decimal"

type StallCreateDTO struct {
	StoreID       uint            `json:"store_id" binding:"required"`
	StallNo       string          `json:"stall_no" binding:"required,max=20"`
	Name          string          `json:"name" binding:"required,max=100"`
	Type          string          `json:"type" binding:"max=20"`
	Description   string          `json:"description" binding:"max=500"`
	Logo          string          `json:"logo" binding:"max=255"`
	RevenueRatio  decimal.Decimal `json:"revenue_ratio"`
	PlatformRatio decimal.Decimal `json:"platform_ratio"`
	ContactName   string          `json:"contact_name" binding:"max=50"`
	ContactPhone  string          `json:"contact_phone" binding:"max=20"`
	PrinterName   string          `json:"printer_name" binding:"max=100"`
	SortOrder     int             `json:"sort_order"`
	Status        int             `json:"status" binding:"oneof=0 1"`
}

type StallUpdateDTO struct {
	Name          string          `json:"name" binding:"max=100"`
	Type          string          `json:"type" binding:"max=20"`
	Description   string          `json:"description" binding:"max=500"`
	Logo          string          `json:"logo" binding:"max=255"`
	RevenueRatio  *decimal.Decimal `json:"revenue_ratio"`
	PlatformRatio *decimal.Decimal `json:"platform_ratio"`
	ContactName   string          `json:"contact_name" binding:"max=50"`
	ContactPhone  string          `json:"contact_phone" binding:"max=20"`
	PrinterName   string          `json:"printer_name" binding:"max=100"`
	SortOrder     *int             `json:"sort_order"`
	Status        int              `json:"status" binding:"oneof=0 1"`
}

type StallQueryDTO struct {
	PageQuery
	StoreID uint   `form:"store_id"`
	Name    string `form:"name"`
	Status  int    `form:"status"`
	Type    string `form:"type"`
}

type StallResponse struct {
	ID            uint            `json:"id"`
	StoreID       uint            `json:"store_id"`
	StoreName     string          `json:"store_name"`
	StallNo       string          `json:"stall_no"`
	Name          string          `json:"name"`
	Type          string          `json:"type"`
	Description   string          `json:"description"`
	Logo          string          `json:"logo"`
	RevenueRatio  decimal.Decimal `json:"revenue_ratio"`
	PlatformRatio decimal.Decimal `json:"platform_ratio"`
	ContactName   string          `json:"contact_name"`
	ContactPhone  string          `json:"contact_phone"`
	PrinterName   string          `json:"printer_name"`
	SortOrder     int             `json:"sort_order"`
	Status        int             `json:"status"`
	CreatedAt     string          `json:"created_at"`
}

type StallDeviceRegisterDTO struct {
	StoreID    uint   `json:"store_id" binding:"required"`
	StallID    uint   `json:"stall_id" binding:"required"`
	DeviceID   string `json:"device_id" binding:"required,max=64"`
	DeviceName string `json:"device_name" binding:"max=100"`
	DeviceType string `json:"device_type" binding:"max=20"`
	OSVersion  string `json:"os_version" binding:"max=50"`
	AppVersion string `json:"app_version" binding:"max=50"`
}

type StallDeviceHeartbeatDTO struct {
	DeviceID   string `json:"device_id" binding:"required"`
	AppVersion string `json:"app_version" binding:"max=50"`
}

type StallDeviceResponse struct {
	ID              uint   `json:"id"`
	StoreID         uint   `json:"store_id"`
	StallID         uint   `json:"stall_id"`
	StallName       string `json:"stall_name"`
	DeviceID        string `json:"device_id"`
	DeviceName      string `json:"device_name"`
	DeviceType      string `json:"device_type"`
	OSVersion       string `json:"os_version"`
	AppVersion      string `json:"app_version"`
	IsOnline        bool   `json:"is_online"`
	LastOnlineAt    string `json:"last_online_at"`
	LastHeartbeatAt string `json:"last_heartbeat_at"`
	Status          int    `json:"status"`
}

type StallUserCreateDTO struct {
	StoreID  uint   `json:"store_id" binding:"required"`
	StallID  uint   `json:"stall_id" binding:"required"`
	Username string `json:"username" binding:"required,max=50"`
	Password string `json:"password" binding:"required,max=50"`
	RealName string `json:"real_name" binding:"max=50"`
	Phone    string `json:"phone" binding:"max=20"`
	Role     string `json:"role" binding:"max=20"`
	Status   int    `json:"status" binding:"oneof=0 1"`
}

type StallUserUpdateDTO struct {
	RealName string `json:"real_name" binding:"max=50"`
	Phone    string `json:"phone" binding:"max=20"`
	Role     string `json:"role" binding:"max=20"`
	Status   int    `json:"status" binding:"oneof=0 1"`
	Password string `json:"password" binding:"max=50"`
}

type StallUserQueryDTO struct {
	PageQuery
	StoreID uint   `form:"store_id"`
	StallID uint   `form:"stall_id"`
	Username string `form:"username"`
	Status  int    `form:"status"`
}

type StallUserResponse struct {
	ID        uint   `json:"id"`
	StoreID   uint   `json:"store_id"`
	StallID   uint   `json:"stall_id"`
	StallName string `json:"stall_name"`
	Username  string `json:"username"`
	RealName  string `json:"real_name"`
	Phone     string `json:"phone"`
	Role      string `json:"role"`
	Status    int    `json:"status"`
	CreatedAt string `json:"created_at"`
}

type StallLoginDTO struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type StallLoginResponse struct {
	Token    string          `json:"token"`
	User     StallUserResponse `json:"user"`
	Stall    StallResponse    `json:"stall"`
}

type StallSettlementCreateDTO struct {
	StoreID        uint   `json:"store_id" binding:"required"`
	StallID        uint   `json:"stall_id" binding:"required"`
	SettlementDate string `json:"settlement_date" binding:"required"`
	Remark         string `json:"remark" binding:"max=500"`
}

type StallSettlementQueryDTO struct {
	PageQuery
	StoreID          uint   `form:"store_id"`
	StallID          uint   `form:"stall_id"`
	SettlementDate   string `form:"settlement_date"`
	SettlementStatus int    `form:"settlement_status"`
}

type StallSettlementResponse struct {
	ID               uint            `json:"id"`
	StoreID          uint            `json:"store_id"`
	StallID          uint            `json:"stall_id"`
	StallName        string          `json:"stall_name"`
	SettlementNo     string          `json:"settlement_no"`
	SettlementDate   string          `json:"settlement_date"`
	OrderCount       int             `json:"order_count"`
	TotalAmount      decimal.Decimal `json:"total_amount"`
	RefundAmount     decimal.Decimal `json:"refund_amount"`
	NetAmount        decimal.Decimal `json:"net_amount"`
	StallAmount      decimal.Decimal `json:"stall_amount"`
	PlatformAmount   decimal.Decimal `json:"platform_amount"`
	SettlementStatus int             `json:"settlement_status"`
	SettledAt        string          `json:"settled_at"`
	Remark           string          `json:"remark"`
}

type StallDailyReportQueryDTO struct {
	StoreID    uint   `form:"store_id"`
	StallID    uint   `form:"stall_id"`
	StartDate  string `form:"start_date" binding:"required"`
	EndDate    string `form:"end_date" binding:"required"`
}

type StallDailyReportResponse struct {
	ID             uint            `json:"id"`
	StoreID        uint            `json:"store_id"`
	StallID        uint            `json:"stall_id"`
	StallName      string          `json:"stall_name"`
	ReportDate     string          `json:"report_date"`
	OrderCount     int             `json:"order_count"`
	TotalAmount    decimal.Decimal `json:"total_amount"`
	DiscountAmount decimal.Decimal `json:"discount_amount"`
	CouponAmount   decimal.Decimal `json:"coupon_amount"`
	RefundAmount   decimal.Decimal `json:"refund_amount"`
	NetAmount      decimal.Decimal `json:"net_amount"`
	StallAmount    decimal.Decimal `json:"stall_amount"`
	PlatformAmount decimal.Decimal `json:"platform_amount"`
}

type StallSummaryDTO struct {
	StoreID           uint            `json:"store_id"`
	StallCount        uint            `json:"stall_id"`
	StallName        string          `json:"stall_name"`
	TodayOrders      int             `json:"today_orders"`
	TodayAmount      decimal.Decimal `json:"today_amount"`
	TodayStallAmount decimal.Decimal `json:"today_stall_amount"`
	IsOnline         bool            `json:"is_online"`
	DeviceCount      int             `json:"device_count"`
}
