package service

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"sort"
	"time"

	"github.com/shopspring/decimal"
	"stalll-hub-pos/backend/config"
	"stalll-hub-pos/backend/internal/dto"
	"stalll-hub-pos/backend/internal/model"
	"stalll-hub-pos/backend/pkg/database"
)

type ForecastService struct {
	client    *http.Client
	baseURL   string
}

func NewForecastService() *ForecastService {
	cfg := config.AppConfig.Forecast
	if cfg.BaseURL == "" {
		cfg.BaseURL = "http://127.0.0.1:8010"
	}
	return &ForecastService{
		client:  &http.Client{Timeout: 60 * time.Second},
		baseURL: cfg.BaseURL,
	}
}

func (s *ForecastService) GetStoreForecast(storeID uint, forecastDays, historyDays int) (*dto.StoreForecastResponse, error) {
	if forecastDays <= 0 {
		forecastDays = config.AppConfig.Forecast.ForecastDays
	}
	if historyDays <= 0 {
		historyDays = config.AppConfig.Forecast.HistoryDays
	}

	url := fmt.Sprintf("%s/forecast/store/%d?forecast_days=%d&history_days=%d",
		s.baseURL, storeID, forecastDays, historyDays)

	resp, err := s.client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("forecast service call failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read forecast response failed: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("forecast service error: status=%d, body=%s", resp.StatusCode, string(body))
	}

	var result dto.StoreForecastResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("parse forecast response failed: %w", err)
	}

	log.Printf("[ForecastService] Store %d forecast: %d SKUs, quality=%.2f",
		storeID, result.TotalForecastSKUCount, result.DataQualityScore)

	return &result, nil
}

func (s *ForecastService) HealthCheck() (bool, error) {
	url := fmt.Sprintf("%s/health", s.baseURL)
	resp, err := s.client.Get(url)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	return resp.StatusCode == http.StatusOK, nil
}

func (s *ForecastService) CalculateStockingSuggestion(
	storeID uint,
	forecast *dto.StoreForecastResponse,
) (*dto.StockingSuggestionResponse, error) {
	safetyRatio := config.AppConfig.Forecast.SafetyStockRatio
	if safetyRatio <= 0 {
		safetyRatio = 1.2
	}

	suggestions := make([]dto.StockingSuggestionItem, 0)
	totalCost := decimal.Zero

	for _, sku := range forecast.SKUForecasts {
		ingredients := s.getSKUIngredients(storeID, sku.ProductID, sku.SKUID)
		for _, ing := range ingredients {
			usage := sku.TotalForecast.Mul(ing.Quantity)
			safetyStock := usage.Mul(decimal.NewFromFloat(safetyRatio))
			suggestedQty := safetyStock.Sub(ing.CurrentStock)
			if suggestedQty.LessThan(decimal.Zero) {
				suggestedQty = decimal.Zero
			}

			estimatedCost := suggestedQty.Mul(ing.UnitPrice)
			totalCost = totalCost.Add(estimatedCost)

			suggestions = append(suggestions, dto.StockingSuggestionItem{
				IngredientID:   ing.IngredientID,
				IngredientName: ing.IngredientName,
				Category:       ing.Category,
				Unit:           ing.Unit,
				CurrentStock:   ing.CurrentStock,
				ForecastUsage:  usage,
				SafetyStock:    safetyStock,
				SuggestedQty:   suggestedQty,
				UnitPrice:      ing.UnitPrice,
				EstimatedCost:  estimatedCost,
				Supplier:       ing.Supplier,
			})
		}
	}

	suggestions = mergeSuggestions(suggestions)

	return &dto.StockingSuggestionResponse{
		StoreID:      storeID,
		ForecastDate: forecast.ForecastDate,
		ForecastDays: forecast.ForecastDays,
		TotalItems:   len(suggestions),
		TotalEstCost: totalCost,
		Suggestions:  suggestions,
	}, nil
}

type skuIngredient struct {
	IngredientID   uint
	IngredientName string
	Category       string
	Unit           string
	Quantity       decimal.Decimal
	CurrentStock   decimal.Decimal
	UnitPrice      decimal.Decimal
	Supplier       string
}

func (s *ForecastService) getSKUIngredients(storeID, productID, skuID uint) []skuIngredient {
	var result []skuIngredient

	var boms []model.ProductBOM
	query := database.DB.Where("store_id = ? AND product_id = ?", storeID, productID)
	if skuID > 0 {
		query = query.Where("sku_id = ? OR sku_id = 0", skuID)
	}
	query.Preload("Ingredient").Find(&boms)

	for _, bom := range boms {
		if bom.IngredientID == 0 || bom.Ingredient.Status != 1 {
			continue
		}

		currentStock := decimal.Zero
		if skuStock, ok := s.getIngredientStock(bom.IngredientID); ok {
			currentStock = skuStock
		}

		result = append(result, skuIngredient{
			IngredientID:   bom.IngredientID,
			IngredientName: bom.IngredientName,
			Category:       bom.Ingredient.Category,
			Unit:           bom.Ingredient.Unit,
			Quantity:       bom.Quantity,
			CurrentStock:   currentStock,
			UnitPrice:      bom.Ingredient.CurrentPrice,
			Supplier:       bom.Ingredient.Supplier,
		})
	}

	return result
}

func (s *ForecastService) getIngredientStock(ingredientID uint) (decimal.Decimal, bool) {
	var ingredient model.Ingredient
	if err := database.DB.Select("current_price").First(&ingredient, ingredientID).Error; err != nil {
		return decimal.Zero, false
	}
	return decimal.Zero, true
}

func mergeSuggestions(items []dto.StockingSuggestionItem) []dto.StockingSuggestionItem {
	merged := make(map[uint]dto.StockingSuggestionItem)

	for _, item := range items {
		if existing, ok := merged[item.IngredientID]; ok {
			existing.ForecastUsage = existing.ForecastUsage.Add(item.ForecastUsage)
			existing.SafetyStock = existing.SafetyStock.Add(item.SafetyStock)
			existing.SuggestedQty = existing.SuggestedQty.Add(item.SuggestedQty)
			existing.EstimatedCost = existing.EstimatedCost.Add(item.EstimatedCost)
			merged[item.IngredientID] = existing
		} else {
			merged[item.IngredientID] = item
		}
	}

	result := make([]dto.StockingSuggestionItem, 0, len(merged))
	for _, v := range merged {
		result = append(result, v)
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].SuggestedQty.GreaterThan(result[j].SuggestedQty)
	})

	return result
}
