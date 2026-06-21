package model

import (
	"time"

	"github.com/shopspring/decimal"
)

type Ingredient struct {
	BaseModel
	StoreID       uint            `gorm:"not null;index" json:"store_id"`
	IngredientNo  string          `gorm:"size:50;not null;index" json:"ingredient_no"`
	Name          string          `gorm:"size:100;not null" json:"name"`
	Category      string          `gorm:"size:50;index" json:"category"`
	Unit          string          `gorm:"size:20;not null" json:"unit"`
	CurrentPrice  decimal.Decimal `gorm:"type:decimal(10,2);default:0" json:"current_price"`
	CurrentStock  decimal.Decimal `gorm:"type:decimal(12,2);default:0" json:"current_stock"`
	StockUnit     string          `gorm:"size:20" json:"stock_unit"`
	Supplier      string          `gorm:"size:100" json:"supplier"`
	SupplierPhone string          `gorm:"size:20" json:"supplier_phone"`
	SupplierEmail string          `gorm:"size:100" json:"supplier_email"`
	Status        int             `gorm:"default:1" json:"status"`
	Remark        string          `gorm:"size:255" json:"remark"`
	Store         Store           `gorm:"foreignKey:StoreID" json:"store,omitempty"`
}

type ProductBOM struct {
	BaseModel
	StoreID      uint            `gorm:"not null;index" json:"store_id"`
	ProductID    uint            `gorm:"not null;index" json:"product_id"`
	SKUID        uint            `gorm:"index" json:"sku_id"`
	IngredientID uint            `gorm:"not null;index" json:"ingredient_id"`
	IngredientName string        `gorm:"size:100;not null" json:"ingredient_name"`
	Quantity     decimal.Decimal `gorm:"type:decimal(10,3);not null" json:"quantity"`
	Unit         string          `gorm:"size:20" json:"unit"`
	WastageRate  decimal.Decimal `gorm:"type:decimal(5,2);default:0" json:"wastage_rate"`
	SortOrder    int             `gorm:"default:0" json:"sort_order"`
	Product      Product         `gorm:"foreignKey:ProductID" json:"product,omitempty"`
	Ingredient   Ingredient      `gorm:"foreignKey:IngredientID" json:"ingredient,omitempty"`
}

type IngredientPrice struct {
	BaseModel
	StoreID       uint            `gorm:"not null;index" json:"store_id"`
	IngredientID  uint            `gorm:"not null;index" json:"ingredient_id"`
	IngredientNo  string          `gorm:"size:50;index" json:"ingredient_no"`
	IngredientName string         `gorm:"size:100" json:"ingredient_name"`
	Price         decimal.Decimal `gorm:"type:decimal(10,2);not null" json:"price"`
	PreviousPrice decimal.Decimal `gorm:"type:decimal(10,2);default:0" json:"previous_price"`
	PriceChange   decimal.Decimal `gorm:"type:decimal(5,2);default:0" json:"price_change"`
	Supplier      string          `gorm:"size:100" json:"supplier"`
	EffectiveDate string          `gorm:"size:10;index;not null" json:"effective_date"`
	Source        string          `gorm:"size:20;default:manual" json:"source"`
	BatchNo       string          `gorm:"size:32;index" json:"batch_no"`
	Ingredient    Ingredient      `gorm:"foreignKey:IngredientID" json:"ingredient,omitempty"`
}

type CostAlert struct {
	BaseModel
	StoreID       uint            `gorm:"not null;index" json:"store_id"`
	IngredientID  uint            `gorm:"not null;index" json:"ingredient_id"`
	IngredientName string         `gorm:"size:100;not null" json:"ingredient_name"`
	AlertType     string          `gorm:"size:20;not null" json:"alert_type"`
	PreviousPrice decimal.Decimal `gorm:"type:decimal(10,2)" json:"previous_price"`
	CurrentPrice  decimal.Decimal `gorm:"type:decimal(10,2)" json:"current_price"`
	ChangeRate    decimal.Decimal `gorm:"type:decimal(5,2)" json:"change_rate"`
	Threshold     decimal.Decimal `gorm:"type:decimal(5,2)" json:"threshold"`
	Status        int             `gorm:"default:0" json:"status"`
	NotifiedAt    *time.Time      `json:"notified_at"`
	HandledAt     *time.Time      `json:"handled_at"`
	Handler       string          `gorm:"size:50" json:"handler"`
	Remark        string          `gorm:"size:255" json:"remark"`
	Ingredient    Ingredient      `gorm:"foreignKey:IngredientID" json:"ingredient,omitempty"`
}
