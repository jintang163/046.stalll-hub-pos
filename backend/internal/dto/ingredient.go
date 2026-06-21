package dto

import "github.com/shopspring/decimal"

type IngredientQueryDTO struct {
	StoreID  uint   `form:"store_id"`
	Category string `form:"category"`
	Keyword  string `form:"keyword"`
	Status   int    `form:"status"`
	Page     int    `form:"page"`
	PageSize int    `form:"page_size"`
}

type BOMSaveDTO struct {
	StoreID   uint                `json:"store_id" binding:"required"`
	ProductID uint                `json:"product_id" binding:"required"`
	SKUID     uint                `json:"sku_id"`
	Items     []BOMItemSaveDTO    `json:"items"`
}

type BOMItemSaveDTO struct {
	IngredientID   uint            `json:"ingredient_id" binding:"required"`
	IngredientName string          `json:"ingredient_name"`
	Quantity       decimal.Decimal `json:"quantity" binding:"required"`
	Unit           string          `json:"unit"`
	WastageRate    decimal.Decimal `json:"wastage_rate"`
	SortOrder      int             `json:"sort_order"`
}

type ProductCostDetailDTO struct {
	ProductID       uint                  `json:"product_id"`
	SKUID           uint                  `json:"sku_id"`
	TotalCost       decimal.Decimal       `json:"total_cost"`
	IngredientCost  []BOMIngredientCostDTO `json:"ingredient_cost"`
	IngredientCount int                   `json:"ingredient_count"`
}

type BOMIngredientCostDTO struct {
	IngredientID   uint            `json:"ingredient_id"`
	IngredientName string          `json:"ingredient_name"`
	Unit           string          `json:"unit"`
	UnitPrice      decimal.Decimal `json:"unit_price"`
	Quantity       decimal.Decimal `json:"quantity"`
	WastageRate    decimal.Decimal `json:"wastage_rate"`
	ActualQty      decimal.Decimal `json:"actual_qty"`
	TotalCost      decimal.Decimal `json:"total_cost"`
}

type CostAlertQueryDTO struct {
	StoreID  uint `form:"store_id"`
	Status   int  `form:"status"`
	Page     int  `form:"page"`
	PageSize int  `form:"page_size"`
}

type CostAlertHandleDTO struct {
	AlertID uint   `json:"alert_id" binding:"required"`
	Handler string `json:"handler"`
	Remark  string `json:"remark"`
}

type IngredientPriceQueryDTO struct {
	IngredientID uint   `form:"ingredient_id" binding:"required"`
	StartDate    string `form:"start_date"`
	EndDate      string `form:"end_date"`
	Limit        int    `form:"limit"`
}

type ProfitSummaryV2Response struct {
	TotalRevenue    decimal.Decimal `json:"total_revenue"`
	TotalMaterialCost decimal.Decimal `json:"total_material_cost"`
	GrossProfit     decimal.Decimal `json:"gross_profit"`
	GrossMargin     decimal.Decimal `json:"gross_margin"`
	NetProfit       decimal.Decimal `json:"net_profit"`
	NetMargin       decimal.Decimal `json:"net_margin"`
	ProductCount    int             `json:"product_count"`
	OrderCount      int             `json:"order_count"`
}

type ProfitReportV2Response struct {
	ProductID       uint            `json:"product_id"`
	ProductName     string          `json:"product_name"`
	Quantity        int             `json:"quantity"`
	Revenue         decimal.Decimal `json:"revenue"`
	MaterialCost    decimal.Decimal `json:"material_cost"`
	GrossProfit     decimal.Decimal `json:"gross_profit"`
	GrossMargin     decimal.Decimal `json:"gross_margin"`
	UnitPrice       decimal.Decimal `json:"unit_price"`
	UnitCost        decimal.Decimal `json:"unit_cost"`
}
