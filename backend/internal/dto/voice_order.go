package dto

import (
	"github.com/shopspring/decimal"
)

type VoiceParseRequest struct {
	StoreID uint   `json:"store_id" binding:"required"`
	Text    string `json:"text" binding:"required"`
}

type VoiceMatchResult struct {
	ProductID   uint            `json:"product_id"`
	ProductName string          `json:"product_name"`
	SKUID       uint            `json:"sku_id"`
	SKUName     string          `json:"sku_name"`
	Price       decimal.Decimal `json:"price"`
	Quantity    int             `json:"quantity"`
	MatchScore  float64         `json:"match_score"`
	Image       string          `json:"image"`
}

type VoiceParseResponse struct {
	OriginalText string             `json:"original_text"`
	Items        []VoiceMatchResult `json:"items"`
	Unmatched    []string           `json:"unmatched"`
}
