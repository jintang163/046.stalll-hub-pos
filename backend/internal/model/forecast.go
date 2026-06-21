package model

import (
	"time"

	"github.com/shopspring/decimal"
)

type PurchaseOrder struct {
	BaseModel
	StoreID       uint            `gorm:"not null;index" json:"store_id"`
	PurchaseNo    string          `gorm:"size:50;not null;uniqueIndex" json:"purchase_no"`
	SupplierName  string          `gorm:"size:100" json:"supplier_name"`
	SupplierPhone string          `gorm:"size:20" json:"supplier_phone"`
	SupplierEmail string          `gorm:"size:100" json:"supplier_email"`
	TotalAmount   decimal.Decimal `gorm:"type:decimal(12,2);default:0" json:"total_amount"`
	TotalQuantity int             `gorm:"default:0" json:"total_quantity"`
	ItemCount     int             `gorm:"default:0" json:"item_count"`
	Status        int             `gorm:"default:0;index" json:"status"`
	ForecastDate  string          `gorm:"size:10;index" json:"forecast_date"`
	ForecastDays  int             `gorm:"default:3" json:"forecast_days"`
	Remark        string          `gorm:"size:255" json:"remark"`
	SentAt        *time.Time      `json:"sent_at"`
	Items         []PurchaseOrderItem `gorm:"foreignKey:PurchaseID" json:"items,omitempty"`
	Store         Store           `gorm:"foreignKey:StoreID" json:"store,omitempty"`
}

type PurchaseOrderItem struct {
	BaseModel
	PurchaseID     uint            `gorm:"not null;index" json:"purchase_id"`
	IngredientID   uint            `gorm:"not null;index" json:"ingredient_id"`
	IngredientName string          `gorm:"size:100;not null" json:"ingredient_name"`
	Category       string          `gorm:"size:50" json:"category"`
	Unit           string          `gorm:"size:20" json:"unit"`
	ForecastQty    decimal.Decimal `gorm:"type:decimal(10,2)" json:"forecast_qty"`
	SafetyStockQty decimal.Decimal `gorm:"type:decimal(10,2)" json:"safety_stock_qty"`
	CurrentStock   decimal.Decimal `gorm:"type:decimal(10,2)" json:"current_stock"`
	PurchaseQty    decimal.Decimal `gorm:"type:decimal(10,2)" json:"purchase_qty"`
	UnitPrice      decimal.Decimal `gorm:"type:decimal(10,2)" json:"unit_price"`
	Subtotal       decimal.Decimal `gorm:"type:decimal(10,2)" json:"subtotal"`
	SortOrder      int             `gorm:"default:0" json:"sort_order"`
	Purchase       PurchaseOrder   `gorm:"foreignKey:PurchaseID" json:"purchase,omitempty"`
}

type SalesForecast struct {
	BaseModel
	StoreID           uint            `gorm:"not null;index" json:"store_id"`
	SKUID             uint            `gorm:"not null;index" json:"sku_id"`
	SKUName           string          `gorm:"size:100" json:"sku_name"`
	ProductID         uint            `gorm:"index" json:"product_id"`
	ProductName       string          `gorm:"size:100" json:"product_name"`
	ForecastDate      string          `gorm:"size:10;index" json:"forecast_date"`
	ForecastDays      int             `gorm:"default:3" json:"forecast_days"`
	DailyForecastJSON string          `gorm:"type:text" json:"daily_forecast_json"`
	TotalForecast     decimal.Decimal `gorm:"type:decimal(10,2)" json:"total_forecast"`
	AvgDailyForecast  decimal.Decimal `gorm:"type:decimal(10,2)" json:"avg_daily_forecast"`
	ConfidenceLower   decimal.Decimal `gorm:"type:decimal(10,2)" json:"confidence_lower"`
	ConfidenceUpper   decimal.Decimal `gorm:"type:decimal(10,2)" json:"confidence_upper"`
	Trend             string          `gorm:"size:20" json:"trend"`
	QualityScore      float64         `gorm:"type:decimal(5,2)" json:"quality_score"`
	GeneratedAt       *time.Time      `json:"generated_at"`
}
