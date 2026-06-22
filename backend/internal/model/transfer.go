package model

import (
	"time"

	"github.com/shopspring/decimal"
)

const (
	TransferStatusPendingAccept  = 0
	TransferStatusPendingOut   = 1
	TransferStatusOutConfirmed = 2
	TransferStatusInTransit = 3
	TransferStatusReceived  = 4
	TransferStatusCompleted = 5
	TransferStatusCancelled = 6
)

type TransferOrder struct {
	BaseModel
	TransferNo      string            `gorm:"size:32;not null;uniqueIndex" json:"transfer_no"`
	FromStoreID     uint              `gorm:"not null;index" json:"from_store_id"`
	ToStoreID       uint              `gorm:"not null;index" json:"to_store_id"`
	Status          int               `gorm:"default:0;index" json:"status"`
	TotalQty        decimal.Decimal   `gorm:"type:decimal(12,2);default:0" json:"total_qty"`
	TotalAmount     decimal.Decimal   `gorm:"type:decimal(12,2);default:0" json:"total_amount"`
	TransferType    string            `gorm:"size:20;default:normal" json:"transfer_type"`
	Priority        string            `gorm:"size:20;default:normal" json:"priority"`
	OutOperatorID   uint              `json:"out_operator_id"`
	OutOperatorName string            `gorm:"size:50" json:"out_operator_name"`
	OutConfirmedAt  *time.Time        `json:"out_confirmed_at"`
	AcceptOperatorID   uint            `json:"accept_operator_id"`
	AcceptOperatorName string          `gorm:"size:50" json:"accept_operator_name"`
	AcceptedAt      *time.Time          `json:"accepted_at"`
	InOperatorID    uint              `json:"in_operator_id"`
	InOperatorName  string            `gorm:"size:50" json:"in_operator_name"`
	ReceivedAt      *time.Time        `json:"received_at"`
	CompletedAt     *time.Time        `json:"completed_at"`
	LogisticsCompany string           `gorm:"size:50" json:"logistics_company"`
	TrackingNo      string            `gorm:"size:50;index" json:"tracking_no"`
	SenderName      string            `gorm:"size:50" json:"sender_name"`
	SenderPhone     string            `gorm:"size:20" json:"sender_phone"`
	ReceiverName    string            `gorm:"size:50" json:"receiver_name"`
	ReceiverPhone   string            `gorm:"size:20" json:"receiver_phone"`
	ReceiverAddress string            `gorm:"size:255" json:"receiver_address"`
	Remark          string            `gorm:"size:500" json:"remark"`
	HasDiff         bool              `gorm:"default:false" json:"has_diff"`
	DiffRemark      string            `gorm:"size:500" json:"diff_remark"`
	FromStore       Store             `gorm:"foreignKey:FromStoreID" json:"from_store,omitempty"`
	ToStore         Store             `gorm:"foreignKey:ToStoreID" json:"to_store,omitempty"`
	Items           []TransferOrderItem `gorm:"foreignKey:TransferID" json:"items,omitempty"`
	LogisticsTracks []TransferLogistics `gorm:"foreignKey:TransferID" json:"logistics_tracks,omitempty"`
}

type TransferOrderItem struct {
	BaseModel
	TransferID     uint            `gorm:"not null;index" json:"transfer_id"`
	IngredientID   uint            `gorm:"not null;index" json:"ingredient_id"`
	IngredientNo   string          `gorm:"size:50" json:"ingredient_no"`
	IngredientName string          `gorm:"size:100;not null" json:"ingredient_name"`
	Unit           string          `gorm:"size:20" json:"unit"`
	OutQty         decimal.Decimal `gorm:"type:decimal(12,2);default:0" json:"out_qty"`
	InQty          decimal.Decimal `gorm:"type:decimal(12,2);default:0" json:"in_qty"`
	UnitPrice      decimal.Decimal `gorm:"type:decimal(10,2);default:0" json:"unit_price"`
	Amount         decimal.Decimal `gorm:"type:decimal(12,2);default:0" json:"amount"`
	DiffQty        decimal.Decimal `gorm:"type:decimal(12,2);default:0" json:"diff_qty"`
	DiffAmount     decimal.Decimal `gorm:"type:decimal(12,2);default:0" json:"diff_amount"`
	Remark         string          `gorm:"size:255" json:"remark"`
	Ingredient     Ingredient      `gorm:"foreignKey:IngredientID" json:"ingredient,omitempty"`
}

type TransferLogistics struct {
	BaseModel
	TransferID    uint   `gorm:"not null;index" json:"transfer_id"`
	TrackingNo    string `gorm:"size:50;index" json:"tracking_no"`
	LogisticsCode string `gorm:"size:20" json:"logistics_code"`
	LogisticsName string `gorm:"size:50" json:"logistics_name"`
	Status        string `gorm:"size:20" json:"status"`
	Location      string `gorm:"size:255" json:"location"`
	Description   string `gorm:"size:500" json:"description"`
	Operator      string `gorm:"size:50" json:"operator"`
	OperatorPhone string `gorm:"size:20" json:"operator_phone"`
	TrackTime     *time.Time `json:"track_time"`
}
