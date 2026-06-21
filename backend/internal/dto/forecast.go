package dto

import (
	"time"

	"github.com/shopspring/decimal"
)

type DailyForecastItem struct {
	Date     string          `json:"date"`
	Forecast decimal.Decimal `json:"forecast"`
	Lower    decimal.Decimal `json:"lower"`
	Upper    decimal.Decimal `json:"upper"`
}

type SKUForecastItem struct {
	SKUID           uint              `json:"sku_id"`
	SKUName         string            `json:"sku_name"`
	ProductID       uint              `json:"product_id"`
	ProductName     string            `json:"product_name"`
	DailyForecast   []DailyForecastItem `json:"daily_forecast"`
	TotalForecast   decimal.Decimal   `json:"total_forecast"`
	AvgDaily        decimal.Decimal   `json:"avg_daily"`
	ConfidenceLower decimal.Decimal   `json:"confidence_lower"`
	ConfidenceUpper decimal.Decimal   `json:"confidence_upper"`
	Trend           string            `json:"trend"`
}

type StoreForecastResponse struct {
	StoreID             uint              `json:"store_id"`
	StoreName           string            `json:"store_name"`
	ForecastDate        string            `json:"forecast_date"`
	ForecastDays        int               `json:"forecast_days"`
	GeneratedAt         string            `json:"generated_at"`
	SKUForecasts        []SKUForecastItem `json:"sku_forecasts"`
	TotalForecastSKUCount int             `json:"total_forecast_sku_count"`
	DataQualityScore    float64           `json:"data_quality_score"`
}

type ForecastRequest struct {
	StoreID      uint   `json:"store_id" binding:"required"`
	ForecastDays int    `json:"forecast_days"`
	HistoryDays  int    `json:"history_days"`
}

type ForecastQuery struct {
	StoreID uint   `form:"store_id"`
	Date    string `form:"date"`
	Page    int    `form:"page,default=1"`
	Size    int    `form:"size,default=20"`
}

type PurchaseOrderCreateRequest struct {
	StoreID       uint                    `json:"store_id" binding:"required"`
	ForecastDate  string                  `json:"forecast_date"`
	ForecastDays  int                     `json:"forecast_days"`
	SupplierName  string                  `json:"supplier_name"`
	SupplierPhone string                  `json:"supplier_phone"`
	SupplierEmail string                  `json:"supplier_email"`
	Items         []PurchaseItemCreate    `json:"items" binding:"required"`
	Remark        string                  `json:"remark"`
}

type PurchaseItemCreate struct {
	IngredientID   uint            `json:"ingredient_id" binding:"required"`
	IngredientName string          `json:"ingredient_name"`
	Category       string          `json:"category"`
	Unit           string          `json:"unit"`
	ForecastQty    decimal.Decimal `json:"forecast_qty"`
	SafetyStockQty decimal.Decimal `json:"safety_stock_qty"`
	CurrentStock   decimal.Decimal `json:"current_stock"`
	PurchaseQty    decimal.Decimal `json:"purchase_qty" binding:"required"`
	UnitPrice      decimal.Decimal `json:"unit_price"`
}

type PurchaseOrderResponse struct {
	ID            uint                    `json:"id"`
	StoreID       uint                    `json:"store_id"`
	StoreName     string                  `json:"store_name"`
	PurchaseNo    string                  `json:"purchase_no"`
	SupplierName  string                  `json:"supplier_name"`
	SupplierPhone string                  `json:"supplier_phone"`
	SupplierEmail string                  `json:"supplier_email"`
	TotalAmount   decimal.Decimal         `json:"total_amount"`
	TotalQuantity int                     `json:"total_quantity"`
	ItemCount     int                     `json:"item_count"`
	Status        int                     `json:"status"`
	StatusText    string                  `json:"status_text"`
	ForecastDate  string                  `json:"forecast_date"`
	ForecastDays  int                     `json:"forecast_days"`
	Remark        string                  `json:"remark"`
	SentAt        *time.Time              `json:"sent_at"`
	CreatedAt     time.Time               `json:"created_at"`
	Items         []PurchaseItemResponse  `json:"items,omitempty"`
}

type PurchaseItemResponse struct {
	ID             uint            `json:"id"`
	IngredientID   uint            `json:"ingredient_id"`
	IngredientName string          `json:"ingredient_name"`
	Category       string          `json:"category"`
	Unit           string          `json:"unit"`
	ForecastQty    decimal.Decimal `json:"forecast_qty"`
	SafetyStockQty decimal.Decimal `json:"safety_stock_qty"`
	CurrentStock   decimal.Decimal `json:"current_stock"`
	PurchaseQty    decimal.Decimal `json:"purchase_qty"`
	UnitPrice      decimal.Decimal `json:"unit_price"`
	Subtotal       decimal.Decimal `json:"subtotal"`
}

type PurchaseOrderQuery struct {
	StoreID uint   `form:"store_id"`
	Status  int    `form:"status"`
	Date    string `form:"date"`
	Keyword string `form:"keyword"`
	Page    int    `form:"page,default=1"`
	Size    int    `form:"size,default=20"`
}

type PurchaseOrderListResponse struct {
	List  []PurchaseOrderResponse `json:"list"`
	Total int64                   `json:"total"`
	Page  int                     `json:"page"`
	Size  int                     `json:"size"`
}

type StockingSuggestionItem struct {
	IngredientID   uint            `json:"ingredient_id"`
	IngredientName string          `json:"ingredient_name"`
	Category       string          `json:"category"`
	Unit           string          `json:"unit"`
	CurrentStock   decimal.Decimal `json:"current_stock"`
	ForecastUsage  decimal.Decimal `json:"forecast_usage"`
	SafetyStock    decimal.Decimal `json:"safety_stock"`
	SuggestedQty   decimal.Decimal `json:"suggested_qty"`
	UnitPrice      decimal.Decimal `json:"unit_price"`
	EstimatedCost  decimal.Decimal `json:"estimated_cost"`
	Supplier       string          `json:"supplier"`
}

type StockingSuggestionResponse struct {
	StoreID       uint                     `json:"store_id"`
	ForecastDate  string                   `json:"forecast_date"`
	ForecastDays  int                      `json:"forecast_days"`
	TotalItems    int                      `json:"total_items"`
	TotalEstCost  decimal.Decimal          `json:"total_estimated_cost"`
	Suggestions   []StockingSuggestionItem `json:"suggestions"`
}
