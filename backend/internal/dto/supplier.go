package dto

import (
	"time"

	"github.com/shopspring/decimal"
)

type SupplierQueryDTO struct {
	StoreID  uint   `form:"store_id"`
	Category string `form:"category"`
	Keyword  string `form:"keyword"`
	Status   int    `form:"status"`
	Page     int    `form:"page,default=1"`
	PageSize int    `form:"page_size,default=20"`
}

type SupplierCreateDTO struct {
	StoreID          uint            `json:"store_id" binding:"required"`
	Name             string          `json:"name" binding:"required"`
	ShortName        string          `json:"short_name"`
	Category         string          `json:"category"`
	ContactPerson    string          `json:"contact_person"`
	Phone            string          `json:"phone"`
	Mobile           string          `json:"mobile"`
	Email            string          `json:"email"`
	Fax              string          `json:"fax"`
	Address          string          `json:"address"`
	Province         string          `json:"province"`
	City             string          `json:"city"`
	District         string          `json:"district"`
	BankName         string          `json:"bank_name"`
	BankAccount      string          `json:"bank_account"`
	BankAccountName  string          `json:"bank_account_name"`
	TaxNo            string          `json:"tax_no"`
	PaymentTerm      int             `json:"payment_term"`
	PaymentTermDesc  string          `json:"payment_term_desc"`
	SettlementMethod string          `json:"settlement_method"`
	CreditLimit      decimal.Decimal `json:"credit_limit"`
	Status           int             `json:"status"`
	Remark           string          `json:"remark"`
}

type SupplierUpdateDTO struct {
	Name             string          `json:"name"`
	ShortName        string          `json:"short_name"`
	Category         string          `json:"category"`
	ContactPerson    string          `json:"contact_person"`
	Phone            string          `json:"phone"`
	Mobile           string          `json:"mobile"`
	Email            string          `json:"email"`
	Fax              string          `json:"fax"`
	Address          string          `json:"address"`
	Province         string          `json:"province"`
	City             string          `json:"city"`
	District         string          `json:"district"`
	BankName         string          `json:"bank_name"`
	BankAccount      string          `json:"bank_account"`
	BankAccountName  string          `json:"bank_account_name"`
	TaxNo            string          `json:"tax_no"`
	PaymentTerm      int             `json:"payment_term"`
	PaymentTermDesc  string          `json:"payment_term_desc"`
	SettlementMethod string          `json:"settlement_method"`
	CreditLimit      decimal.Decimal `json:"credit_limit"`
	Status           int             `json:"status"`
	Remark           string          `json:"remark"`
}

type SupplierResponse struct {
	ID               uint            `json:"id"`
	StoreID          uint            `json:"store_id"`
	StoreName        string          `json:"store_name"`
	SupplierNo       string          `json:"supplier_no"`
	Name             string          `json:"name"`
	ShortName        string          `json:"short_name"`
	Category         string          `json:"category"`
	ContactPerson    string          `json:"contact_person"`
	Phone            string          `json:"phone"`
	Mobile           string          `json:"mobile"`
	Email            string          `json:"email"`
	Fax              string          `json:"fax"`
	Address          string          `json:"address"`
	Province         string          `json:"province"`
	City             string          `json:"city"`
	District         string          `json:"district"`
	BankName         string          `json:"bank_name"`
	BankAccount      string          `json:"bank_account"`
	BankAccountName  string          `json:"bank_account_name"`
	TaxNo            string          `json:"tax_no"`
	PaymentTerm      int             `json:"payment_term"`
	PaymentTermDesc  string          `json:"payment_term_desc"`
	SettlementMethod string          `json:"settlement_method"`
	CreditLimit      decimal.Decimal `json:"credit_limit"`
	CurrentPayable   decimal.Decimal `json:"current_payable"`
	TotalPurchase    decimal.Decimal `json:"total_purchase"`
	TotalPaid        decimal.Decimal `json:"total_paid"`
	Status           int             `json:"status"`
	StatusText       string          `json:"status_text"`
	Remark           string          `json:"remark"`
	CreatedAt        time.Time       `json:"created_at"`
	UpdatedAt        time.Time       `json:"updated_at"`
}

type SupplierStatsResponse struct {
	TotalSupplier    int             `json:"total_supplier"`
	ActiveSupplier   int             `json:"active_supplier"`
	TotalPayable     decimal.Decimal `json:"total_payable"`
	OverduePayable   decimal.Decimal `json:"overdue_payable"`
	TotalPurchase    decimal.Decimal `json:"total_purchase"`
	TotalPaid        decimal.Decimal `json:"total_paid"`
}

type PurchaseOrderCreateV2DTO struct {
	StoreID       uint                 `json:"store_id" binding:"required"`
	SupplierID    uint                 `json:"supplier_id"`
	ForecastDate  string               `json:"forecast_date"`
	ForecastDays  int                  `json:"forecast_days"`
	ExpectedDate  string               `json:"expected_date"`
	PaymentTerm   int                  `json:"payment_term"`
	SupplierName  string               `json:"supplier_name"`
	SupplierPhone string               `json:"supplier_phone"`
	SupplierEmail string               `json:"supplier_email"`
	Items         []PurchaseItemCreate `json:"items" binding:"required"`
	Remark        string               `json:"remark"`
}

type PurchaseReceiveCreateDTO struct {
	StoreID      uint                       `json:"store_id" binding:"required"`
	PurchaseID   uint                       `json:"purchase_id" binding:"required"`
	ReceiveType  string                     `json:"receive_type"`
	Remark       string                     `json:"remark"`
	OperatorID   uint                       `json:"operator_id"`
	OperatorName string                     `json:"operator_name"`
	Items        []PurchaseReceiveItemDTO   `json:"items" binding:"required"`
}

type PurchaseReceiveItemDTO struct {
	PurchaseItemID uint            `json:"purchase_item_id"`
	IngredientID   uint            `json:"ingredient_id" binding:"required"`
	IngredientName string          `json:"ingredient_name"`
	Category       string          `json:"category"`
	Unit           string          `json:"unit"`
	PurchaseQty    decimal.Decimal `json:"purchase_qty"`
	ReceivedQty    decimal.Decimal `json:"received_qty" binding:"required"`
	QualifiedQty   decimal.Decimal `json:"qualified_qty"`
	RejectedQty    decimal.Decimal `json:"rejected_qty"`
	UnitPrice      decimal.Decimal `json:"unit_price"`
	BatchNo        string          `json:"batch_no"`
	ExpiryDate     string          `json:"expiry_date"`
	RejectReason   string          `json:"reject_reason"`
}

type PurchaseReceiveQueryDTO struct {
	StoreID    uint   `form:"store_id"`
	PurchaseID uint   `form:"purchase_id"`
	SupplierID uint   `form:"supplier_id"`
	Keyword    string `form:"keyword"`
	StartDate  string `form:"start_date"`
	EndDate    string `form:"end_date"`
	Page       int    `form:"page,default=1"`
	PageSize   int    `form:"page_size,default=20"`
}

type PurchaseReceiveResponse struct {
	ID           uint            `json:"id"`
	StoreID      uint            `json:"store_id"`
	StoreName    string          `json:"store_name"`
	PurchaseID   uint            `json:"purchase_id"`
	PurchaseNo   string          `json:"purchase_no"`
	SupplierID   uint            `json:"supplier_id"`
	SupplierName string          `json:"supplier_name"`
	ReceiveNo    string          `json:"receive_no"`
	ReceiveType  string          `json:"receive_type"`
	TotalQty     decimal.Decimal `json:"total_qty"`
	TotalAmount  decimal.Decimal `json:"total_amount"`
	Remark       string          `json:"remark"`
	OperatorID   uint            `json:"operator_id"`
	OperatorName string          `json:"operator_name"`
	ReceivedAt   *time.Time      `json:"received_at"`
	CreatedAt    time.Time       `json:"created_at"`
	Items        []PurchaseReceiveItemResponse `json:"items,omitempty"`
}

type PurchaseReceiveItemResponse struct {
	ID             uint            `json:"id"`
	ReceiveID      uint            `json:"receive_id"`
	PurchaseItemID uint            `json:"purchase_item_id"`
	IngredientID   uint            `json:"ingredient_id"`
	IngredientName string          `json:"ingredient_name"`
	Category       string          `json:"category"`
	Unit           string          `json:"unit"`
	PurchaseQty    decimal.Decimal `json:"purchase_qty"`
	ReceivedQty    decimal.Decimal `json:"received_qty"`
	QualifiedQty   decimal.Decimal `json:"qualified_qty"`
	RejectedQty    decimal.Decimal `json:"rejected_qty"`
	UnitPrice      decimal.Decimal `json:"unit_price"`
	Subtotal       decimal.Decimal `json:"subtotal"`
	BatchNo        string          `json:"batch_no"`
	ExpiryDate     string          `json:"expiry_date"`
	RejectReason   string          `json:"reject_reason"`
}

type PayableQueryDTO struct {
	StoreID    uint   `form:"store_id"`
	SupplierID uint   `form:"supplier_id"`
	Status     string `form:"status"`
	IsOverdue  int    `form:"is_overdue"`
	Keyword    string `form:"keyword"`
	StartDate  string `form:"start_date"`
	EndDate    string `form:"end_date"`
	DueBefore  string `form:"due_before"`
	Page       int    `form:"page,default=1"`
	PageSize   int    `form:"page_size,default=20"`
}

type PayablePaymentCreateDTO struct {
	StoreID       uint            `json:"store_id" binding:"required"`
	SupplierID    uint            `json:"supplier_id" binding:"required"`
	PayableID     uint            `json:"payable_id"`
	Amount        decimal.Decimal `json:"amount" binding:"required"`
	PaymentMethod string          `json:"payment_method"`
	PaymentDate   string          `json:"payment_date"`
	OperatorID    uint            `json:"operator_id"`
	OperatorName  string          `json:"operator_name"`
	VoucherNo     string          `json:"voucher_no"`
	VoucherURL    string          `json:"voucher_url"`
	Remark        string          `json:"remark"`
}

type PayableResponse struct {
	ID           uint            `json:"id"`
	StoreID      uint            `json:"store_id"`
	StoreName    string          `json:"store_name"`
	SupplierID   uint            `json:"supplier_id"`
	SupplierName string          `json:"supplier_name"`
	PayableNo    string          `json:"payable_no"`
	BusinessType string          `json:"business_type"`
	BusinessID   uint            `json:"business_id"`
	BusinessNo   string          `json:"business_no"`
	BillDate     string          `json:"bill_date"`
	Amount       decimal.Decimal `json:"amount"`
	PaidAmount   decimal.Decimal `json:"paid_amount"`
	Balance      decimal.Decimal `json:"balance"`
	DueDate      string          `json:"due_date"`
	Status       string          `json:"status"`
	StatusText   string          `json:"status_text"`
	IsOverdue    int             `json:"is_overdue"`
	Remark       string          `json:"remark"`
	PaidAt       *time.Time      `json:"paid_at"`
	CreatedAt    time.Time       `json:"created_at"`
	Payments     []PayablePaymentResponse `json:"payments,omitempty"`
}

type PayablePaymentResponse struct {
	ID           uint            `json:"id"`
	StoreID      uint            `json:"store_id"`
	SupplierID   uint            `json:"supplier_id"`
	SupplierName string          `json:"supplier_name"`
	PayableID    uint            `json:"payable_id"`
	PaymentNo    string          `json:"payment_no"`
	Amount       decimal.Decimal `json:"amount"`
	PaymentMethod string         `json:"payment_method"`
	PaymentMethodText string     `json:"payment_method_text"`
	PaymentDate  string          `json:"payment_date"`
	OperatorID   uint            `json:"operator_id"`
	OperatorName string          `json:"operator_name"`
	VoucherNo    string          `json:"voucher_no"`
	VoucherURL   string          `json:"voucher_url"`
	Remark       string          `json:"remark"`
	CreatedAt    time.Time       `json:"created_at"`
}

type PayablePaymentQueryDTO struct {
	StoreID    uint   `form:"store_id"`
	SupplierID uint   `form:"supplier_id"`
	PayableID  uint   `form:"payable_id"`
	Keyword    string `form:"keyword"`
	StartDate  string `form:"start_date"`
	EndDate    string `form:"end_date"`
	Page       int    `form:"page,default=1"`
	PageSize   int    `form:"page_size,default=20"`
}

type PayableStatsResponse struct {
	TotalCount     int             `json:"total_count"`
	TotalAmount    decimal.Decimal `json:"total_amount"`
	TotalPaid      decimal.Decimal `json:"total_paid"`
	TotalBalance   decimal.Decimal `json:"total_balance"`
	OverdueCount   int             `json:"overdue_count"`
	OverdueAmount  decimal.Decimal `json:"overdue_amount"`
}

type ReconciliationCreateDTO struct {
	StoreID        uint   `json:"store_id" binding:"required"`
	SupplierID     uint   `json:"supplier_id" binding:"required"`
	PeriodStart    string `json:"period_start" binding:"required"`
	PeriodEnd      string `json:"period_end" binding:"required"`
	Remark         string `json:"remark"`
}

type ReconciliationConfirmDTO struct {
	SupplierAmount decimal.Decimal `json:"supplier_amount" binding:"required"`
	Remark         string          `json:"remark"`
	ConfirmedBy    string          `json:"confirmed_by"`
}

type ReconciliationSupplierAmountDTO struct {
	SupplierAmount     decimal.Decimal `json:"supplier_amount" binding:"required"`
	DifferenceRemark   string          `json:"difference_remark"`
}

type ReconciliationQueryDTO struct {
	StoreID    uint   `form:"store_id"`
	SupplierID uint   `form:"supplier_id"`
	Status     string `form:"status"`
	Keyword    string `form:"keyword"`
	StartDate  string `form:"start_date"`
	EndDate    string `form:"end_date"`
	Page       int    `form:"page,default=1"`
	PageSize   int    `form:"page_size,default=20"`
}

type ReconciliationResponse struct {
	ID             uint            `json:"id"`
	StoreID        uint            `json:"store_id"`
	StoreName      string          `json:"store_name"`
	SupplierID     uint            `json:"supplier_id"`
	SupplierName   string          `json:"supplier_name"`
	ReconcileNo    string          `json:"reconcile_no"`
	PeriodStart    string          `json:"period_start"`
	PeriodEnd      string          `json:"period_end"`
	SystemAmount   decimal.Decimal `json:"system_amount"`
	SupplierAmount decimal.Decimal `json:"supplier_amount"`
	DiffAmount     decimal.Decimal `json:"diff_amount"`
	Status         string          `json:"status"`
	StatusText     string          `json:"status_text"`
	ConfirmedAt    *time.Time      `json:"confirmed_at"`
	ConfirmedBy    string          `json:"confirmed_by"`
	Remark         string          `json:"remark"`
	CreatedAt      time.Time       `json:"created_at"`
	Items          []ReconciliationItemResponse `json:"items,omitempty"`
}

type ReconciliationItemResponse struct {
	ID             uint            `json:"id"`
	ReconcileID    uint            `json:"reconcile_id"`
	BusinessType   string          `json:"business_type"`
	BusinessTypeText string        `json:"business_type_text"`
	BusinessID     uint            `json:"business_id"`
	BusinessNo     string          `json:"business_no"`
	BusinessDate   string          `json:"business_date"`
	SystemAmount   decimal.Decimal `json:"system_amount"`
	SupplierAmount decimal.Decimal `json:"supplier_amount"`
	DiffAmount     decimal.Decimal `json:"diff_amount"`
	Remark         string          `json:"remark"`
}

type NotifySupplierDTO struct {
	NotifyType []string `json:"notify_type"`
	TemplateID uint     `json:"template_id"`
	Content    string   `json:"content"`
}

type PurchaseOrderV2Response struct {
	ID               uint            `json:"id"`
	StoreID          uint            `json:"store_id"`
	StoreName        string          `json:"store_name"`
	SupplierID       uint            `json:"supplier_id"`
	PurchaseNo       string          `json:"purchase_no"`
	SupplierName     string          `json:"supplier_name"`
	SupplierPhone    string          `json:"supplier_phone"`
	SupplierEmail    string          `json:"supplier_email"`
	TotalAmount      decimal.Decimal `json:"total_amount"`
	ReceivedAmount   decimal.Decimal `json:"received_amount"`
	TotalQuantity    int             `json:"total_quantity"`
	ReceivedQuantity int             `json:"received_quantity"`
	ItemCount        int             `json:"item_count"`
	Status           int             `json:"status"`
	StatusText       string          `json:"status_text"`
	ForecastDate     string          `json:"forecast_date"`
	ForecastDays     int             `json:"forecast_days"`
	PaymentTerm      int             `json:"payment_term"`
	PaymentTermText  string          `json:"payment_term_text"`
	ExpectedDate     string          `json:"expected_date"`
	Remark           string          `json:"remark"`
	SentAt           *time.Time      `json:"sent_at"`
	CreatedAt        time.Time       `json:"created_at"`
	Items            []PurchaseItemV2Response `json:"items,omitempty"`
}

type PurchaseItemV2Response struct {
	ID             uint            `json:"id"`
	IngredientID   uint            `json:"ingredient_id"`
	IngredientName string          `json:"ingredient_name"`
	Category       string          `json:"category"`
	Unit           string          `json:"unit"`
	ForecastQty    decimal.Decimal `json:"forecast_qty"`
	SafetyStockQty decimal.Decimal `json:"safety_stock_qty"`
	CurrentStock   decimal.Decimal `json:"current_stock"`
	PurchaseQty    decimal.Decimal `json:"purchase_qty"`
	ReceivedQty    decimal.Decimal `json:"received_qty"`
	UnitPrice      decimal.Decimal `json:"unit_price"`
	Subtotal       decimal.Decimal `json:"subtotal"`
}

type PageResponse struct {
	List  interface{} `json:"list"`
	Total int64       `json:"total"`
	Page  int         `json:"page"`
	Size  int         `json:"size"`
}
