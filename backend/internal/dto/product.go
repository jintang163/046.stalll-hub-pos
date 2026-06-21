package dto

import "github.com/shopspring/decimal"

type ProductCreateDTO struct {
	StoreID               uint                   `json:"store_id" binding:"required"`
	CategoryID            uint                   `json:"category_id" binding:"required"`
	Name                  string                 `json:"name" binding:"required,max=100"`
	Description           string                 `json:"description"`
	MainImage             string                 `json:"main_image"`
	Images                string                 `json:"images"`
	SortOrder             int                    `json:"sort_order"`
	Status                int                    `json:"status"`
	IsHot                 bool                   `json:"is_hot"`
	IsRecommend           bool                   `json:"is_recommend"`
	StockWarningThreshold int                    `json:"stock_warning_threshold"`
	SKUs                  []ProductSKUCreateDTO  `json:"skus" binding:"required,min=1"`
	Attributes            []AttributeCreateDTO   `json:"attributes"`
}

type ProductUpdateDTO struct {
	CategoryID            uint                   `json:"category_id"`
	Name                  string                 `json:"name" binding:"max=100"`
	Description           string                 `json:"description"`
	MainImage             string                 `json:"main_image"`
	Images                string                 `json:"images"`
	SortOrder             int                    `json:"sort_order"`
	Status                int                    `json:"status"`
	IsHot                 bool                   `json:"is_hot"`
	IsRecommend           bool                   `json:"is_recommend"`
	StockWarningThreshold int                    `json:"stock_warning_threshold"`
	SKUs                  []ProductSKUUpdateDTO  `json:"skus"`
	Attributes            []AttributeUpdateDTO   `json:"attributes"`
}

type ProductSKUCreateDTO struct {
	SKUCode      string          `json:"sku_code" binding:"required,max=50"`
	SpecName     string          `json:"spec_name" binding:"required,max=50"`
	Price        decimal.Decimal `json:"price" binding:"required"`
	OriginalPrice decimal.Decimal `json:"original_price"`
	Stock        int             `json:"stock"`
	Image        string          `json:"image"`
	Status       int             `json:"status"`
	AttributeValues []SKUAttributeValueDTO `json:"attribute_values"`
}

type ProductSKUUpdateDTO struct {
	ID           uint            `json:"id"`
	SKUCode      string          `json:"sku_code" binding:"max=50"`
	SpecName     string          `json:"spec_name" binding:"max=50"`
	Price        decimal.Decimal `json:"price"`
	OriginalPrice decimal.Decimal `json:"original_price"`
	Stock        int             `json:"stock"`
	Image        string          `json:"image"`
	Status       int             `json:"status"`
	AttributeValues []SKUAttributeValueDTO `json:"attribute_values"`
}

type SKUAttributeValueDTO struct {
	AttributeID uint `json:"attribute_id" binding:"required"`
	ValueID     uint `json:"value_id" binding:"required"`
}

type AttributeCreateDTO struct {
	Name      string               `json:"name" binding:"required,max=50"`
	SortOrder int                  `json:"sort_order"`
	Status    int                  `json:"status"`
	Values    []AttributeValueDTO `json:"values" binding:"required,min=1"`
}

type AttributeUpdateDTO struct {
	ID        uint                 `json:"id"`
	Name      string               `json:"name" binding:"max=50"`
	SortOrder int                  `json:"sort_order"`
	Status    int                  `json:"status"`
	Values    []AttributeValueDTO `json:"values"`
}

type AttributeValueDTO struct {
	ID         uint            `json:"id"`
	Value      string          `json:"value" binding:"required,max=50"`
	SortOrder  int             `json:"sort_order"`
	Status     int             `json:"status"`
	ExtraPrice decimal.Decimal `json:"extra_price"`
	Stock      int             `json:"stock"`
}

type ProductQueryDTO struct {
	StoreID    uint   `form:"store_id"`
	CategoryID uint   `form:"category_id"`
	Name       string `form:"name"`
	Status     *int   `form:"status"`
	IsHot      *bool  `form:"is_hot"`
	IsRecommend *bool `form:"is_recommend"`
	Pagination
}

type SKUStockUpdateDTO struct {
	StoreID uint `json:"store_id" binding:"required"`
	Items   []SKUStockItem `json:"items" binding:"required,min=1"`
}

type SKUStockItem struct {
	SKUID  uint `json:"sku_id" binding:"required"`
	Stock  int  `json:"stock" binding:"required"`
}

type BatchPriceUpdateDTO struct {
	StoreID    uint            `json:"store_id" binding:"required"`
	ProductIDs []uint          `json:"product_ids" binding:"required,min=1"`
	Price      decimal.Decimal `json:"price"`
	PriceType  string          `json:"price_type" binding:"required,oneof=fixed percentage"`
}

type ProductCopyDTO struct {
	StoreID      uint   `json:"store_id" binding:"required"`
	ProductID    uint   `json:"product_id" binding:"required"`
	NewName      string `json:"new_name"`
	CategoryID   uint   `json:"category_id"`
}

type ProductListResponse struct {
	ID          uint            `json:"id"`
	CategoryID  uint            `json:"category_id"`
	Name        string          `json:"name"`
	Description string          `json:"description"`
	MainImage   string          `json:"main_image"`
	Status      int             `json:"status"`
	IsHot       bool            `json:"is_hot"`
	IsRecommend bool            `json:"is_recommend"`
	SortOrder   int             `json:"sort_order"`
	MinPrice    decimal.Decimal `json:"min_price"`
	MaxPrice    decimal.Decimal `json:"max_price"`
	TotalStock  int             `json:"total_stock"`
	SKUCount    int             `json:"sku_count"`
	CreatedAt   string          `json:"created_at"`
}

type ProductDetailResponse struct {
	ID                    uint                   `json:"id"`
	StoreID               uint                   `json:"store_id"`
	CategoryID            uint                   `json:"category_id"`
	Name                  string                 `json:"name"`
	Description           string                 `json:"description"`
	MainImage             string                 `json:"main_image"`
	Images                string                 `json:"images"`
	SortOrder             int                    `json:"sort_order"`
	Status                int                    `json:"status"`
	IsHot                 bool                   `json:"is_hot"`
	IsRecommend           bool                   `json:"is_recommend"`
	StockWarningThreshold int                    `json:"stock_warning_threshold"`
	Category              *CategorySimpleDTO     `json:"category"`
	SKUs                  []ProductSKUResponse   `json:"skus"`
	Attributes            []AttributeResponse    `json:"attributes"`
	CreatedAt             string                 `json:"created_at"`
	UpdatedAt             string                 `json:"updated_at"`
}

type CategorySimpleDTO struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
}

type ProductSKUResponse struct {
	ID              uint                  `json:"id"`
	ProductID       uint                  `json:"product_id"`
	SKUCode         string                `json:"sku_code"`
	SpecName        string                `json:"spec_name"`
	Price           decimal.Decimal       `json:"price"`
	OriginalPrice   decimal.Decimal       `json:"original_price"`
	Stock           int                   `json:"stock"`
	SoldCount       int                   `json:"sold_count"`
	Image           string                `json:"image"`
	Status          int                   `json:"status"`
	IsSoldOut       bool                  `json:"is_sold_out"`
	AttributeValues []SKUAttributeResponse `json:"attribute_values"`
}

type SKUAttributeResponse struct {
	AttributeID   uint   `json:"attribute_id"`
	AttributeName string `json:"attribute_name"`
	ValueID       uint   `json:"value_id"`
	ValueName     string `json:"value_name"`
}

type AttributeResponse struct {
	ID        uint                   `json:"id"`
	ProductID uint                   `json:"product_id"`
	Name      string                 `json:"name"`
	SortOrder int                    `json:"sort_order"`
	Status    int                    `json:"status"`
	Values    []AttributeValueResponse `json:"values"`
}

type AttributeValueResponse struct {
	ID         uint            `json:"id"`
	Value      string          `json:"value"`
	SortOrder  int             `json:"sort_order"`
	Status     int             `json:"status"`
	ExtraPrice decimal.Decimal `json:"extra_price"`
	Stock      int             `json:"stock"`
}

type SyncProductResponse struct {
	LastSyncID uint                    `json:"last_sync_id"`
	Total      int64                   `json:"total"`
	Products   []ProductDetailResponse `json:"products"`
}

type SoldOutBatchDTO struct {
	StoreID      uint   `json:"store_id"`
	SKUIds       []uint `json:"sku_ids" binding:"required,min=1"`
	OperatorID   uint   `json:"operator_id"`
	OperatorName string `json:"operator_name"`
	Source       string `json:"source"`
	Remark       string `json:"remark"`
}

type SoldOutRecordQueryDTO struct {
	StoreID    uint   `form:"store_id"`
	ProductID  uint   `form:"product_id"`
	SKUID      uint   `form:"sku_id"`
	ActionType string `form:"action_type"`
	StartDate  string `form:"start_date"`
	EndDate    string `form:"end_date"`
	Pagination
}

type SoldOutRecordResponse struct {
	ID            uint   `json:"id"`
	StoreID       uint   `json:"store_id"`
	ProductID     uint   `json:"product_id"`
	SKUID         uint   `json:"sku_id"`
	SKUCode       string `json:"sku_code"`
	ProductName   string `json:"product_name"`
	SpecName      string `json:"spec_name"`
	CategoryID    uint   `json:"category_id"`
	CategoryName  string `json:"category_name"`
	ActionType    string `json:"action_type"`
	OperatorID    uint   `json:"operator_id"`
	OperatorName  string `json:"operator_name"`
	Source        string `json:"source"`
	Remark        string `json:"remark"`
	StockAtAction int    `json:"stock_at_action"`
	CreatedAt     string `json:"created_at"`
}
