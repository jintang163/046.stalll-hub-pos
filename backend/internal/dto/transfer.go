package dto

import "github.com/shopspring/decimal"

type TransferOrderQueryDTO struct {
	Pagination
	FromStoreID uint   `form:"from_store_id"`
	ToStoreID   uint   `form:"to_store_id"`
	Status      int    `form:"status"`
	Keyword     string `form:"keyword"`
	TransferNo  string `form:"transfer_no"`
}

type CreateTransferOrderDTO struct {
	FromStoreID     uint                     `json:"from_store_id" binding:"required"`
	ToStoreID       uint                     `json:"to_store_id" binding:"required"`
	TransferType    string                   `json:"transfer_type"`
	Priority        string                   `json:"priority"`
	Remark          string                   `json:"remark"`
	SenderName      string                   `json:"sender_name"`
	SenderPhone     string                   `json:"sender_phone"`
	ReceiverName    string                   `json:"receiver_name"`
	ReceiverPhone   string                   `json:"receiver_phone"`
	ReceiverAddress string                   `json:"receiver_address"`
	Items           []TransferOrderItemDTO   `json:"items" binding:"required,min=1"`
}

type TransferOrderItemDTO struct {
	IngredientID   uint            `json:"ingredient_id" binding:"required"`
	IngredientNo   string          `json:"ingredient_no"`
	IngredientName string          `json:"ingredient_name"`
	Unit           string          `json:"unit"`
	OutQty         decimal.Decimal `json:"out_qty" binding:"required"`
	UnitPrice      decimal.Decimal `json:"unit_price"`
	Remark         string          `json:"remark"`
}

type ConfirmOutboundDTO struct {
	OperatorID   uint   `json:"operator_id"`
	OperatorName string `json:"operator_name"`
	Remark       string `json:"remark"`
}

type ReceiveTransferDTO struct {
	OperatorID   uint                     `json:"operator_id"`
	OperatorName string                   `json:"operator_name"`
	Remark       string                   `json:"remark"`
	Items        []ReceiveTransferItemDTO `json:"items" binding:"required,min=1"`
}

type ReceiveTransferItemDTO struct {
	ItemID     uint            `json:"item_id" binding:"required"`
	InQty      decimal.Decimal `json:"in_qty" binding:"required"`
	Remark     string          `json:"remark"`
}

type UpdateLogisticsDTO struct {
	LogisticsCompany string `json:"logistics_company"`
	TrackingNo       string `json:"tracking_no"`
	LogisticsCode    string `json:"logistics_code"`
}

type TransferDiffDTO struct {
	DiffRemark string                   `json:"diff_remark"`
	Items      []TransferDiffItemDTO    `json:"items"`
}

type TransferDiffItemDTO struct {
	ItemID  uint            `json:"item_id"`
	InQty   decimal.Decimal `json:"in_qty"`
	Remark  string          `json:"remark"`
}
