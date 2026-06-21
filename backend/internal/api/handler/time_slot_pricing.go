package handler

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"stalll-hub-pos/backend/internal/dto"
	"stalll-hub-pos/backend/internal/middleware"
	"stalll-hub-pos/backend/internal/service"
)

type TimeSlotPricingHandler struct {
	timeSlotPricingService *service.TimeSlotPricingService
}

func NewTimeSlotPricingHandler() *TimeSlotPricingHandler {
	return &TimeSlotPricingHandler{
		timeSlotPricingService: service.NewTimeSlotPricingService(),
	}
}

func (h *TimeSlotPricingHandler) Create(c *gin.Context) {
	var req dto.TimeSlotPricingCreateDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.Error(c, "参数错误: "+err.Error())
		return
	}

	storeID := middleware.GetStoreID(c)
	if storeID == 0 {
		storeID = req.StoreID
	}

	result, err := h.timeSlotPricingService.CreateTimeSlotPricing(storeID, &req)
	if err != nil {
		middleware.Error(c, "创建时段定价失败: "+err.Error())
		return
	}

	middleware.Success(c, result)
}

func (h *TimeSlotPricingHandler) Update(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		middleware.Error(c, "无效的时段定价ID")
		return
	}

	var req dto.TimeSlotPricingUpdateDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.Error(c, "参数错误: "+err.Error())
		return
	}

	result, err := h.timeSlotPricingService.UpdateTimeSlotPricing(uint(id), &req)
	if err != nil {
		middleware.Error(c, "更新时段定价失败: "+err.Error())
		return
	}

	middleware.Success(c, result)
}

func (h *TimeSlotPricingHandler) Delete(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		middleware.Error(c, "无效的时段定价ID")
		return
	}

	err = h.timeSlotPricingService.DeleteTimeSlotPricing(uint(id))
	if err != nil {
		middleware.Error(c, "删除时段定价失败: "+err.Error())
		return
	}

	middleware.Success(c, nil)
}

func (h *TimeSlotPricingHandler) GetByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		middleware.Error(c, "无效的时段定价ID")
		return
	}

	result, err := h.timeSlotPricingService.GetTimeSlotPricing(uint(id))
	if err != nil {
		middleware.Error(c, "获取时段定价失败: "+err.Error())
		return
	}

	middleware.Success(c, result)
}

func (h *TimeSlotPricingHandler) List(c *gin.Context) {
	var query dto.TimeSlotPricingQueryDTO
	if err := c.ShouldBindQuery(&query); err != nil {
		middleware.Error(c, "参数错误: "+err.Error())
		return
	}

	if query.StoreID == 0 {
		query.StoreID = middleware.GetStoreID(c)
	}

	result, err := h.timeSlotPricingService.ListTimeSlotPricings(&query)
	if err != nil {
		middleware.Error(c, "获取时段定价列表失败: "+err.Error())
		return
	}

	middleware.Success(c, result)
}

func (h *TimeSlotPricingHandler) GetActive(c *gin.Context) {
	storeIDStr := c.Query("store_id")
	storeID := uint(0)
	if storeIDStr != "" {
		id, err := strconv.ParseUint(storeIDStr, 10, 32)
		if err == nil {
			storeID = uint(id)
		}
	}

	if storeID == 0 {
		storeID = middleware.GetStoreID(c)
	}

	result, err := h.timeSlotPricingService.GetActiveTimeSlots(storeID)
	if err != nil {
		middleware.Error(c, "获取活动时段定价失败: "+err.Error())
		return
	}

	middleware.Success(c, result)
}

func (h *TimeSlotPricingHandler) CalculatePrice(c *gin.Context) {
	storeIDStr := c.Query("store_id")
	storeID, err := strconv.ParseUint(storeIDStr, 10, 32)
	if err != nil {
		middleware.Error(c, "无效的门店ID")
		return
	}

	skuIDStr := c.Query("sku_id")
	skuID, err := strconv.ParseUint(skuIDStr, 10, 32)
	if err != nil {
		middleware.Error(c, "无效的SKU ID")
		return
	}

	quantityStr := c.Query("quantity")
	quantity, err := strconv.Atoi(quantityStr)
	if err != nil {
		quantity = 1
	}

	checkTime := time.Now()
	timeStr := c.Query("check_time")
	if timeStr != "" {
		if t, err := time.Parse("2006-01-02 15:04:05", timeStr); err == nil {
			checkTime = t
		}
	}

	result, err := h.timeSlotPricingService.CalculatePrice(uint(storeID), uint(skuID), quantity, checkTime)
	if err != nil {
		middleware.Error(c, "计算时段价格失败: "+err.Error())
		return
	}

	middleware.Success(c, result)
}

func (h *TimeSlotPricingHandler) CalculateOrderPrices(c *gin.Context) {
	storeIDStr := c.Query("store_id")
	storeID, err := strconv.ParseUint(storeIDStr, 10, 32)
	if err != nil {
		middleware.Error(c, "无效的门店ID")
		return
	}

	var items []dto.OrderItemDTO
	if err := c.ShouldBindJSON(&items); err != nil {
		middleware.Error(c, "参数错误: "+err.Error())
		return
	}

	checkTime := time.Now()
	timeStr := c.Query("check_time")
	if timeStr != "" {
		if t, err := time.Parse("2006-01-02 15:04:05", timeStr); err == nil {
			checkTime = t
		}
	}

	totalFinal, updatedItems, err := h.timeSlotPricingService.CalculateOrderPrices(uint(storeID), items, checkTime)
	if err != nil {
		middleware.Error(c, "计算订单时段价格失败: "+err.Error())
		return
	}

	middleware.Success(c, gin.H{
		"total_final":  totalFinal,
		"updated_items": updatedItems,
	})
}
