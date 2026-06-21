package model

import (
	"time"

	"github.com/shopspring/decimal"
)

type Category struct {
	BaseModel
	StoreID     uint      `gorm:"not null;index" json:"store_id"`
	Name        string    `gorm:"size:50;not null" json:"name"`
	SortOrder   int       `gorm:"default:0" json:"sort_order"`
	Status      int       `gorm:"default:1" json:"status"`
	Description string    `gorm:"size:255" json:"description"`
	Products    []Product `gorm:"foreignKey:CategoryID" json:"products,omitempty"`
}

type Product struct {
	BaseModel
	StoreID         uint           `gorm:"not null;index" json:"store_id"`
	CategoryID      uint           `gorm:"not null;index" json:"category_id"`
	Name            string         `gorm:"size:100;not null" json:"name"`
	Description     string         `gorm:"size:500" json:"description"`
	MainImage       string         `gorm:"size:255" json:"main_image"`
	Images          string         `gorm:"size:1000" json:"images"`
	SortOrder       int            `gorm:"default:0" json:"sort_order"`
	Status          int            `gorm:"default:1" json:"status"`
	IsHot           bool           `gorm:"default:false" json:"is_hot"`
	IsRecommend     bool           `gorm:"default:false" json:"is_recommend"`
	StockWarningThreshold int      `gorm:"default:10" json:"stock_warning_threshold"`
	Category        Category       `gorm:"foreignKey:CategoryID" json:"category,omitempty"`
	Stall           *Stall         `gorm:"foreignKey:StallID" json:"stall,omitempty"`
	SKUs            []ProductSKU   `gorm:"foreignKey:ProductID" json:"skus,omitempty"`
	Attributes      []ProductAttribute `gorm:"foreignKey:ProductID" json:"attributes,omitempty"`
}

type ProductSKU struct {
	BaseModel
	ProductID    uint            `gorm:"not null;index" json:"product_id"`
	StoreID      uint            `gorm:"not null;index" json:"store_id"`
	SKUCode      string          `gorm:"size:50;unique;not null" json:"sku_code"`
	SpecName     string          `gorm:"size:50;not null" json:"spec_name"`
	Price        decimal.Decimal `gorm:"type:decimal(10,2);not null" json:"price"`
	OriginalPrice decimal.Decimal `gorm:"type:decimal(10,2)" json:"original_price"`
	Stock        int             `gorm:"default:0" json:"stock"`
	SoldCount    int             `gorm:"default:0" json:"sold_count"`
	Image        string          `gorm:"size:255" json:"image"`
	Status       int             `gorm:"default:1" json:"status"`
	IsSoldOut    bool            `gorm:"default:false;index" json:"is_sold_out"`
	Product      Product         `gorm:"foreignKey:ProductID" json:"product,omitempty"`
	AttributeValues []SKUAttributeValue `gorm:"foreignKey:SKUID" json:"attribute_values,omitempty"`
}

type ProductAttribute struct {
	BaseModel
	ProductID uint      `gorm:"not null;index" json:"product_id"`
	Name      string    `gorm:"size:50;not null" json:"name"`
	SortOrder int       `gorm:"default:0" json:"sort_order"`
	Status    int       `gorm:"default:1" json:"status"`
	Product   Product   `gorm:"foreignKey:ProductID" json:"product,omitempty"`
	Values    []AttributeValue `gorm:"foreignKey:AttributeID" json:"values,omitempty"`
}

type AttributeValue struct {
	BaseModel
	AttributeID uint              `gorm:"not null;index" json:"attribute_id"`
	Value       string            `gorm:"size:50;not null" json:"value"`
	SortOrder   int               `gorm:"default:0" json:"sort_order"`
	Status      int               `gorm:"default:1" json:"status"`
	ExtraPrice  decimal.Decimal   `gorm:"type:decimal(10,2);default:0" json:"extra_price"`
	Stock       int               `gorm:"default:-1" json:"stock"`
	Attribute   ProductAttribute  `gorm:"foreignKey:AttributeID" json:"attribute,omitempty"`
}

type SKUAttributeValue struct {
	BaseModel
	SKUID       uint           `gorm:"not null;index" json:"sku_id"`
	AttributeID uint           `gorm:"not null;index" json:"attribute_id"`
	ValueID     uint           `gorm:"not null;index" json:"value_id"`
	SKU         ProductSKU     `gorm:"foreignKey:SKUID" json:"sku,omitempty"`
	Attribute   ProductAttribute `gorm:"foreignKey:AttributeID" json:"attribute,omitempty"`
	Value       AttributeValue `gorm:"foreignKey:ValueID" json:"value,omitempty"`
}

type StockWarning struct {
	BaseModel
	StoreID    uint       `gorm:"not null;index" json:"store_id"`
	SKUID      uint       `gorm:"not null;index" json:"sku_id"`
	ProductID  uint       `gorm:"not null;index" json:"product_id"`
	CurrentStock int      `gorm:"not null" json:"current_stock"`
	Threshold  int        `gorm:"not null" json:"threshold"`
	Status     int        `gorm:"default:0" json:"status"`
	Store      Store      `gorm:"foreignKey:StoreID" json:"store,omitempty"`
	SKU        ProductSKU `gorm:"foreignKey:SKUID" json:"sku,omitempty"`
	Product    Product    `gorm:"foreignKey:ProductID" json:"product,omitempty"`
}

type StockCheck struct {
	BaseModel
	StoreID       uint          `gorm:"not null;index" json:"store_id"`
	CheckNo       string        `gorm:"size:50;unique;not null" json:"check_no"`
	Title         string        `gorm:"size:100;not null" json:"title"`
	CheckType     string        `gorm:"size:20;default:all" json:"check_type"`
	Status        int           `gorm:"default:0" json:"status"`
	TotalSKU      int           `gorm:"default:0" json:"total_sku"`
	CheckedSKU    int           `gorm:"default:0" json:"checked_sku"`
	TotalDiffQty  int           `gorm:"default:0" json:"total_diff_qty"`
	TotalDiffAmount float64     `gorm:"type:decimal(12,2);default:0" json:"total_diff_amount"`
	OperatorID    uint          `json:"operator_id"`
	OperatorName  string        `gorm:"size:50" json:"operator_name"`
	Remark        string        `gorm:"size:500" json:"remark"`
	StartTime     *time.Time    `json:"start_time"`
	EndTime       *time.Time    `json:"end_time"`
	Store         Store         `gorm:"foreignKey:StoreID" json:"store,omitempty"`
	Items         []StockCheckItem `gorm:"foreignKey:CheckID" json:"items,omitempty"`
}

type StockCheckItem struct {
	BaseModel
	CheckID      uint           `gorm:"not null;index" json:"check_id"`
	ProductID    uint           `gorm:"not null;index" json:"product_id"`
	SKUID        uint           `gorm:"not null;index" json:"sku_id"`
	SKUCode      string         `gorm:"size:50;index" json:"sku_code"`
	ProductName  string         `gorm:"size:100" json:"product_name"`
	SpecName     string         `gorm:"size:50" json:"spec_name"`
	CategoryID   uint           `json:"category_id"`
	CategoryName string         `gorm:"size:50" json:"category_name"`
	SystemStock  int            `gorm:"default:0" json:"system_stock"`
	ActualStock  int            `gorm:"default:0" json:"actual_stock"`
	DiffQty      int            `gorm:"default:0" json:"diff_qty"`
	CostPrice    float64        `gorm:"type:decimal(10,2);default:0" json:"cost_price"`
	DiffAmount   float64        `gorm:"type:decimal(10,2);default:0" json:"diff_amount"`
	Status       int            `gorm:"default:0" json:"status"`
	Remark       string         `gorm:"size:200" json:"remark"`
	Check        StockCheck     `gorm:"foreignKey:CheckID" json:"check,omitempty"`
}

type SoldOutRecord struct {
	BaseModel
	StoreID      uint      `gorm:"not null;index" json:"store_id"`
	ProductID    uint      `gorm:"not null;index" json:"product_id"`
	SKUID        uint      `gorm:"not null;index" json:"sku_id"`
	SKUCode      string    `gorm:"size:50" json:"sku_code"`
	ProductName  string    `gorm:"size:100" json:"product_name"`
	SpecName     string    `gorm:"size:50" json:"spec_name"`
	CategoryID   uint      `json:"category_id"`
	CategoryName string    `gorm:"size:50" json:"category_name"`
	ActionType   string    `gorm:"size:20;index" json:"action_type"`
	OperatorID   uint      `json:"operator_id"`
	OperatorName string    `gorm:"size:50" json:"operator_name"`
	Source       string    `gorm:"size:20" json:"source"`
	Remark       string    `gorm:"size:500" json:"remark"`
	StockAtAction int      `json:"stock_at_action"`
	Store        Store     `gorm:"foreignKey:StoreID" json:"store,omitempty"`
	Product      Product   `gorm:"foreignKey:ProductID" json:"product,omitempty"`
	SKU          ProductSKU `gorm:"foreignKey:SKUID" json:"sku,omitempty"`
}
