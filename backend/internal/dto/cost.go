package dto

import (
	"github.com/shopspring/decimal"
)

type CostQueryDTO struct {
	ProductID     uint   `form:"product_id"`
	EffectiveDate string `form:"effective_date"`
	Page          int    `form:"page"`
	PageSize      int    `form:"page_size"`
}

type CostImportDTO struct {
	EffectiveDate string `json:"effective_date" binding:"required"`
}

type ProfitReportQueryDTO struct {
	StoreID   uint   `form:"store_id"`
	StartDate string `form:"start_date" binding:"required"`
	EndDate   string `form:"end_date" binding:"required"`
}

type ProfitReportResponse struct {
	ProductID    uint            `json:"product_id"`
	ProductName  string          `json:"product_name"`
	Quantity     int             `json:"quantity"`
	Revenue      decimal.Decimal `json:"revenue"`
	UnitCost     decimal.Decimal `json:"unit_cost"`
	TotalCost    decimal.Decimal `json:"total_cost"`
	GrossProfit  decimal.Decimal `json:"gross_profit"`
	GrossMargin  decimal.Decimal `json:"gross_margin"`
}

type ProfitSummaryResponse struct {
	TotalRevenue decimal.Decimal `json:"total_revenue"`
	TotalCost    decimal.Decimal `json:"total_cost"`
	GrossProfit  decimal.Decimal `json:"gross_profit"`
	GrossMargin  decimal.Decimal `json:"gross_margin"`
	ProductCount int             `json:"product_count"`
}

type RevenueReportQueryDTO struct {
	StoreID    uint   `form:"store_id"`
	StartDate  string `form:"start_date" binding:"required"`
	EndDate    string `form:"end_date" binding:"required"`
	ReportType string `form:"report_type"`
}
