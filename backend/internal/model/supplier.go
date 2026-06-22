package model

import (
	"time"

	"github.com/shopspring/decimal"
)

type Supplier struct {
	BaseModel
	StoreID          uint            `gorm:"not null;index" json:"store_id"`
	SupplierNo       string          `gorm:"size:50;not null;uniqueIndex" json:"supplier_no"`
	Name             string          `gorm:"size:100;not null" json:"name"`
	ShortName        string          `gorm:"size:50" json:"short_name"`
	Category         string          `gorm:"size:50;index" json:"category"`
	ContactPerson    string          `gorm:"size:50" json:"contact_person"`
	Phone            string          `gorm:"size:20" json:"phone"`
	Mobile           string          `gorm:"size:20" json:"mobile"`
	Email            string          `gorm:"size:100" json:"email"`
	Fax              string          `gorm:"size:20" json:"fax"`
	Address          string          `gorm:"size:255" json:"address"`
	Province         string          `gorm:"size:50" json:"province"`
	City             string          `gorm:"size:50" json:"city"`
	District         string          `gorm:"size:50" json:"district"`
	BankName         string          `gorm:"size:100" json:"bank_name"`
	BankAccount      string          `gorm:"size:50" json:"bank_account"`
	BankAccountName  string          `gorm:"size:100" json:"bank_account_name"`
	TaxNo            string          `gorm:"size:50" json:"tax_no"`
	PaymentTerm      int             `gorm:"default:0" json:"payment_term"`
	PaymentTermDesc  string          `gorm:"size:50" json:"payment_term_desc"`
	SettlementMethod string          `gorm:"size:30;default:bank_transfer" json:"settlement_method"`
	CreditLimit      decimal.Decimal `gorm:"type:decimal(12,2);default:0" json:"credit_limit"`
	CurrentPayable   decimal.Decimal `gorm:"type:decimal(12,2);default:0" json:"current_payable"`
	TotalPurchase    decimal.Decimal `gorm:"type:decimal(12,2);default:0" json:"total_purchase"`
	TotalPaid        decimal.Decimal `gorm:"type:decimal(12,2);default:0" json:"total_paid"`
	Status           int             `gorm:"default:1;index" json:"status"`
	Remark           string          `gorm:"size:255" json:"remark"`
	Store            Store           `gorm:"foreignKey:StoreID" json:"store,omitempty"`
}

type PurchaseReceive struct {
	BaseModel
	StoreID        uint            `gorm:"not null;index" json:"store_id"`
	PurchaseID     uint            `gorm:"not null;index" json:"purchase_id"`
	PurchaseNo     string          `gorm:"size:50;index" json:"purchase_no"`
	SupplierID     uint            `gorm:"index" json:"supplier_id"`
	SupplierName   string          `gorm:"size:100" json:"supplier_name"`
	ReceiveNo      string          `gorm:"size:50;not null;uniqueIndex" json:"receive_no"`
	ReceiveType    string          `gorm:"size:20;default:full" json:"receive_type"`
	TotalQty       decimal.Decimal `gorm:"type:decimal(12,2);default:0" json:"total_qty"`
	TotalAmount    decimal.Decimal `gorm:"type:decimal(12,2);default:0" json:"total_amount"`
	Remark         string          `gorm:"size:255" json:"remark"`
	OperatorID     uint            `json:"operator_id"`
	OperatorName   string          `gorm:"size:50" json:"operator_name"`
	ReceivedAt     *time.Time      `json:"received_at"`
	Items          []PurchaseReceiveItem `gorm:"foreignKey:ReceiveID" json:"items,omitempty"`
	Purchase       PurchaseOrder   `gorm:"foreignKey:PurchaseID" json:"purchase,omitempty"`
	Store          Store           `gorm:"foreignKey:StoreID" json:"store,omitempty"`
}

type PurchaseReceiveItem struct {
	BaseModel
	ReceiveID       uint            `gorm:"not null;index" json:"receive_id"`
	PurchaseItemID  uint            `gorm:"index" json:"purchase_item_id"`
	IngredientID    uint            `gorm:"not null;index" json:"ingredient_id"`
	IngredientName  string          `gorm:"size:100;not null" json:"ingredient_name"`
	Category        string          `gorm:"size:50" json:"category"`
	Unit            string          `gorm:"size:20" json:"unit"`
	PurchaseQty     decimal.Decimal `gorm:"type:decimal(10,2)" json:"purchase_qty"`
	ReceivedQty     decimal.Decimal `gorm:"type:decimal(10,2)" json:"received_qty"`
	QualifiedQty    decimal.Decimal `gorm:"type:decimal(10,2)" json:"qualified_qty"`
	RejectedQty     decimal.Decimal `gorm:"type:decimal(10,2)" json:"rejected_qty"`
	UnitPrice       decimal.Decimal `gorm:"type:decimal(10,2)" json:"unit_price"`
	Subtotal        decimal.Decimal `gorm:"type:decimal(12,2)" json:"subtotal"`
	BatchNo         string          `gorm:"size:50;index" json:"batch_no"`
	ExpiryDate      string          `gorm:"size:10" json:"expiry_date"`
	RejectReason    string          `gorm:"size:255" json:"reject_reason"`
	SortOrder       int             `gorm:"default:0" json:"sort_order"`
}

type AccountsPayable struct {
	BaseModel
	StoreID        uint            `gorm:"not null;index" json:"store_id"`
	SupplierID     uint            `gorm:"not null;index" json:"supplier_id"`
	SupplierName   string          `gorm:"size:100" json:"supplier_name"`
	PayableNo      string          `gorm:"size:50;not null;uniqueIndex" json:"payable_no"`
	BusinessType   string          `gorm:"size:30;default:purchase" json:"business_type"`
	BusinessID     uint            `gorm:"index" json:"business_id"`
	BusinessNo     string          `gorm:"size:50;index" json:"business_no"`
	Amount         decimal.Decimal `gorm:"type:decimal(12,2);not null" json:"amount"`
	PaidAmount     decimal.Decimal `gorm:"type:decimal(12,2);default:0" json:"paid_amount"`
	Balance        decimal.Decimal `gorm:"type:decimal(12,2);default:0" json:"balance"`
	DueDate        string          `gorm:"size:10;index" json:"due_date"`
	Status         string          `gorm:"size:20;default:unpaid;index" json:"status"`
	IsOverdue      int             `gorm:"default:0;index" json:"is_overdue"`
	Remark         string          `gorm:"size:255" json:"remark"`
	PaidAt         *time.Time      `json:"paid_at"`
	Supplier       Supplier        `gorm:"foreignKey:SupplierID" json:"supplier,omitempty"`
	Store          Store           `gorm:"foreignKey:StoreID" json:"store,omitempty"`
	Payments       []PayablePayment `gorm:"foreignKey:PayableID" json:"payments,omitempty"`
}

type PayablePayment struct {
	BaseModel
	StoreID        uint            `gorm:"not null;index" json:"store_id"`
	SupplierID     uint            `gorm:"not null;index" json:"supplier_id"`
	SupplierName   string          `gorm:"size:100" json:"supplier_name"`
	PayableID      uint            `gorm:"index" json:"payable_id"`
	PaymentNo      string          `gorm:"size:50;not null;uniqueIndex" json:"payment_no"`
	Amount         decimal.Decimal `gorm:"type:decimal(12,2);not null" json:"amount"`
	PaymentMethod  string          `gorm:"size:30;default:bank_transfer" json:"payment_method"`
	PaymentDate    string          `gorm:"size:10;index" json:"payment_date"`
	OperatorID     uint            `json:"operator_id"`
	OperatorName   string          `gorm:"size:50" json:"operator_name"`
	VoucherNo      string          `gorm:"size:50" json:"voucher_no"`
	VoucherURL     string          `gorm:"size:255" json:"voucher_url"`
	Remark         string          `gorm:"size:255" json:"remark"`
	Payable        AccountsPayable `gorm:"foreignKey:PayableID" json:"payable,omitempty"`
}

type Reconciliation struct {
	BaseModel
	StoreID          uint            `gorm:"not null;index" json:"store_id"`
	SupplierID       uint            `gorm:"not null;index" json:"supplier_id"`
	SupplierName     string          `gorm:"size:100" json:"supplier_name"`
	ReconcileNo      string          `gorm:"size:50;not null;uniqueIndex" json:"reconcile_no"`
	PeriodStart      string          `gorm:"size:10;index" json:"period_start"`
	PeriodEnd        string          `gorm:"size:10;index" json:"period_end"`
	SystemAmount     decimal.Decimal `gorm:"type:decimal(12,2);default:0" json:"system_amount"`
	SupplierAmount   decimal.Decimal `gorm:"type:decimal(12,2);default:0" json:"supplier_amount"`
	DiffAmount       decimal.Decimal `gorm:"type:decimal(12,2);default:0" json:"diff_amount"`
	Status           string          `gorm:"size:20;default:draft;index" json:"status"`
	ConfirmedAt      *time.Time      `json:"confirmed_at"`
	ConfirmedBy      string          `gorm:"size:50" json:"confirmed_by"`
	Remark           string          `gorm:"size:255" json:"remark"`
	DifferenceRemark string          `gorm:"size:500" json:"difference_remark"`
	Items            []ReconciliationItem `gorm:"foreignKey:ReconcileID" json:"items,omitempty"`
	Supplier         Supplier        `gorm:"foreignKey:SupplierID" json:"supplier,omitempty"`
}

type ReconciliationItem struct {
	BaseModel
	ReconcileID   uint            `gorm:"not null;index" json:"reconcile_id"`
	BusinessType  string          `gorm:"size:30" json:"business_type"`
	BusinessID    uint            `json:"business_id"`
	BusinessNo    string          `gorm:"size:50" json:"business_no"`
	BusinessDate  string          `gorm:"size:10" json:"business_date"`
	SystemAmount  decimal.Decimal `gorm:"type:decimal(12,2)" json:"system_amount"`
	SupplierAmount decimal.Decimal `gorm:"type:decimal(12,2)" json:"supplier_amount"`
	DiffAmount    decimal.Decimal `gorm:"type:decimal(12,2)" json:"diff_amount"`
	Remark        string          `gorm:"size:255" json:"remark"`
}
