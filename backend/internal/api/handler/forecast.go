package handler

import (
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"stalll-hub-pos/backend/internal/dto"
	"stalll-hub-pos/backend/internal/middleware"
	"stalll-hub-pos/backend/internal/service"
)

type ForecastHandler struct {
	forecastService *service.ForecastService
	purchaseService *service.PurchaseService
}

func NewForecastHandler() *ForecastHandler {
	return &ForecastHandler{
		forecastService: service.NewForecastService(),
		purchaseService: service.NewPurchaseService(),
	}
}

func (h *ForecastHandler) GetForecast(c *gin.Context) {
	storeID, err := strconv.ParseUint(c.Param("storeId"), 10, 32)
	if err != nil {
		middleware.Error(c, http.StatusBadRequest, "Invalid store ID")
		return
	}

	forecastDays, _ := strconv.Atoi(c.DefaultQuery("forecast_days", "0"))
	historyDays, _ := strconv.Atoi(c.DefaultQuery("history_days", "0"))

	forecast, err := h.forecastService.GetStoreForecast(uint(storeID), forecastDays, historyDays)
	if err != nil {
		log.Printf("[ForecastHandler] Get forecast failed for store %d: %v", storeID, err)
		middleware.Error(c, http.StatusInternalServerError, "Failed to get forecast: "+err.Error())
		return
	}

	middleware.Success(c, forecast)
}

func (h *ForecastHandler) GetStockingSuggestion(c *gin.Context) {
	storeID, err := strconv.ParseUint(c.Param("storeId"), 10, 32)
	if err != nil {
		middleware.Error(c, http.StatusBadRequest, "Invalid store ID")
		return
	}

	forecastDays, _ := strconv.Atoi(c.DefaultQuery("forecast_days", "0"))
	historyDays, _ := strconv.Atoi(c.DefaultQuery("history_days", "0"))

	forecast, err := h.forecastService.GetStoreForecast(uint(storeID), forecastDays, historyDays)
	if err != nil {
		log.Printf("[ForecastHandler] Get forecast failed for store %d: %v", storeID, err)
		middleware.Error(c, http.StatusInternalServerError, "Failed to get forecast: "+err.Error())
		return
	}

	suggestions, err := h.forecastService.CalculateStockingSuggestion(uint(storeID), forecast)
	if err != nil {
		log.Printf("[ForecastHandler] Calculate stocking suggestion failed: %v", err)
		middleware.Error(c, http.StatusInternalServerError, "Failed to calculate stocking suggestion: "+err.Error())
		return
	}

	middleware.Success(c, suggestions)
}

func (h *ForecastHandler) GeneratePurchaseOrder(c *gin.Context) {
	storeID, err := strconv.ParseUint(c.Param("storeId"), 10, 32)
	if err != nil {
		middleware.Error(c, http.StatusBadRequest, "Invalid store ID")
		return
	}

	var req struct {
		ForecastDays int `json:"forecast_days"`
		HistoryDays  int `json:"history_days"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.Error(c, http.StatusBadRequest, "Invalid request: "+err.Error())
		return
	}

	forecast, err := h.forecastService.GetStoreForecast(uint(storeID), req.ForecastDays, req.HistoryDays)
	if err != nil {
		log.Printf("[ForecastHandler] Get forecast failed: %v", err)
		middleware.Error(c, http.StatusInternalServerError, "Failed to get forecast: "+err.Error())
		return
	}

	suggestions, err := h.forecastService.CalculateStockingSuggestion(uint(storeID), forecast)
	if err != nil {
		log.Printf("[ForecastHandler] Calculate stocking suggestion failed: %v", err)
		middleware.Error(c, http.StatusInternalServerError, "Failed to calculate stocking suggestion: "+err.Error())
		return
	}

	purchaseOrders, err := h.purchaseService.AutoGenerateFromForecast(
		uint(storeID), forecast, suggestions,
	)
	if err != nil {
		log.Printf("[ForecastHandler] Generate purchase order failed: %v", err)
		middleware.Error(c, http.StatusInternalServerError, "Failed to generate purchase order: "+err.Error())
		return
	}

	var respList []dto.PurchaseOrderResponse
	for _, po := range purchaseOrders {
		respList = append(respList, h.purchaseService.ConvertToResponse(po))
	}

	middleware.Success(c, gin.H{
		"count":   len(respList),
		"records": respList,
	})
}

func (h *ForecastHandler) GetPurchaseList(c *gin.Context) {
	var query dto.PurchaseOrderQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		middleware.Error(c, http.StatusBadRequest, "Invalid query: "+err.Error())
		return
	}

	orders, total, err := h.purchaseService.ListPurchaseOrders(&query)
	if err != nil {
		middleware.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	var respList []dto.PurchaseOrderResponse
	for _, o := range orders {
		respList = append(respList, h.purchaseService.ConvertToResponse(&o))
	}

	middleware.Success(c, dto.PurchaseOrderListResponse{
		List:  respList,
		Total: total,
		Page:  query.Page,
		Size:  query.Size,
	})
}

func (h *ForecastHandler) GetPurchaseDetail(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		middleware.Error(c, http.StatusBadRequest, "Invalid purchase order ID")
		return
	}

	purchase, err := h.purchaseService.GetPurchaseOrder(uint(id))
	if err != nil {
		middleware.Error(c, http.StatusNotFound, err.Error())
		return
	}

	resp := h.purchaseService.ConvertToResponse(purchase)
	middleware.Success(c, resp)
}

func (h *ForecastHandler) CreatePurchaseOrder(c *gin.Context) {
	var req dto.PurchaseOrderCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.Error(c, http.StatusBadRequest, "Invalid request: "+err.Error())
		return
	}

	purchase, err := h.purchaseService.GeneratePurchaseOrder(&req)
	if err != nil {
		log.Printf("[ForecastHandler] Create purchase order failed: %v", err)
		middleware.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	resp := h.purchaseService.ConvertToResponse(purchase)
	middleware.Success(c, resp)
}

func (h *ForecastHandler) UpdatePurchaseStatus(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		middleware.Error(c, http.StatusBadRequest, "Invalid purchase order ID")
		return
	}

	var req struct {
		Status int `json:"status" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.Error(c, http.StatusBadRequest, "Invalid request")
		return
	}

	if err := h.purchaseService.UpdateStatus(uint(id), req.Status); err != nil {
		middleware.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	middleware.Success(c, gin.H{"message": "Status updated successfully"})
}

func (h *ForecastHandler) SendPurchaseToSupplier(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		middleware.Error(c, http.StatusBadRequest, "Invalid purchase order ID")
		return
	}

	if err := h.purchaseService.SendToSupplier(uint(id)); err != nil {
		log.Printf("[ForecastHandler] Send purchase order failed: %v", err)
		middleware.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	middleware.Success(c, gin.H{"message": "Purchase order sent to supplier successfully"})
}

func (h *ForecastHandler) ExportPurchaseExcel(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		middleware.Error(c, http.StatusBadRequest, "Invalid purchase order ID")
		return
	}

	filePath, err := h.purchaseService.GenerateExcel(uint(id))
	if err != nil {
		log.Printf("[ForecastHandler] Export purchase excel failed: %v", err)
		middleware.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.FileAttachment(filePath, "purchase_order.xlsx")
}

func (h *ForecastHandler) TriggerForecastTask(c *gin.Context) {
	storeID, err := strconv.ParseUint(c.Param("storeId"), 10, 32)
	if err != nil {
		middleware.Error(c, http.StatusBadRequest, "Invalid store ID")
		return
	}

	go func() {
		forecast, err := h.forecastService.GetStoreForecast(uint(storeID), 0, 0)
		if err != nil {
			log.Printf("[ForecastHandler] Trigger forecast task failed: %v", err)
			return
		}

		suggestions, err := h.forecastService.CalculateStockingSuggestion(uint(storeID), forecast)
		if err != nil {
			log.Printf("[ForecastHandler] Calculate suggestions failed: %v", err)
			return
		}

		purchaseOrders, err := h.purchaseService.AutoGenerateFromForecast(
			uint(storeID), forecast, suggestions,
		)
		if err != nil {
			log.Printf("[ForecastHandler] Auto generate purchase failed: %v", err)
			return
		}

		for _, purchase := range purchaseOrders {
			if sendErr := h.purchaseService.SendToSupplier(purchase.ID); sendErr != nil {
				log.Printf("[ForecastHandler] Failed to send purchase order %s to supplier %s: %v",
					purchase.PurchaseNo, purchase.SupplierName, sendErr)
				continue
			}
			log.Printf("[ForecastHandler] Forecast task completed for store %d, purchase order: %s (supplier: %s)",
				storeID, purchase.PurchaseNo, purchase.SupplierName)
		}
	}()

	middleware.Success(c, gin.H{
		"message":    "Forecast task triggered successfully",
		"trigger_at": time.Now().Format("2006-01-02 15:04:05"),
	})
}
