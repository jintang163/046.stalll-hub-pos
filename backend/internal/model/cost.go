package model

import (
	"time"

	"github.com/shopspring/decimal"
)

type ProductCost struct {
	BaseModel
	ProductID   uint            `gorm:"not null;index" json:"product_id"`
	ProductName string          `gorm:"size:100;not null" json:"product_name"`
	SKUID       uint            `gorm:"index" json:"sku_id"`
	SKUName     string          `gorm:"size:100" json:"sku_name"`
	UnitCost    decimal.Decimal `gorm:"type:decimal(10,2);not null" json:"unit_cost"`
	Price       decimal.Decimal `gorm:"type:decimal(10,2)" json:"price"`
	GrossProfit decimal.Decimal `gorm:"type:decimal(10,2)" json:"gross_profit"`
	GrossMargin decimal.Decimal `gorm:"type:decimal(5,2)" json:"gross_margin"`
	EffectiveDate string        `gorm:"size:10;not null" json:"effective_date"`
	BatchNo      string         `gorm:"size:32;index" json:"batch_no"`
	Product      Product         `gorm:"foreignKey:ProductID" json:"product,omitempty"`
}

type CostImportBatch struct {
	BaseModel
	BatchNo       string     `gorm:"size:32;unique;not null" json:"batch_no"`
	FileName      string     `gorm:"size:255" json:"file_name"`
	TotalRows     int        `gorm:"default:0" json:"total_rows"`
	SuccessCount  int        `gorm:"default:0" json:"success_count"`
	FailCount     int        `gorm:"default:0" json:"fail_count"`
	Status        int        `gorm:"default:0" json:"status"`
	EffectiveDate string     `gorm:"size:10;not null" json:"effective_date"`
	CompletedAt   *time.Time `json:"completed_at"`
}

type ProfitReport struct {
	BaseModel
	StoreID         uint            `gorm:"not null;index" json:"store_id"`
	ReportDate      string          `gorm:"size:10;not null;index" json:"report_date"`
	ReportType      string          `gorm:"size:10;not null;default:daily" json:"report_type"`
	TotalRevenue    decimal.Decimal `gorm:"type:decimal(12,2);default:0" json:"total_revenue"`
	TotalCost       decimal.Decimal `gorm:"type:decimal(12,2);default:0" json:"total_cost"`
	GrossProfit     decimal.Decimal `gorm:"type:decimal(12,2);default:0" json:"gross_profit"`
	GrossMargin     decimal.Decimal `gorm:"type:decimal(5,2);default:0" json:"gross_margin"`
	OrderCount      int             `gorm:"default:0" json:"order_count"`
	ProductCount    int             `gorm:"default:0" json:"product_count"`
	Store           Store           `gorm:"foreignKey:StoreID" json:"store,omitempty"`
}
