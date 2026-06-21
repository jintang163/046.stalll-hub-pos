package handler

import (
	"strconv"
	"stalll-hub-pos/backend/internal/dto"
	"stalll-hub-pos/backend/internal/middleware"
	"stalll-hub-pos/backend/internal/model"
	"stalll-hub-pos/backend/internal/service"

	"github.com/gin-gonic/gin"
)

type IngredientHandler struct {
	bomService   *service.BOMService
	alertService *service.CostAlertService
	inventorySync *service.InventorySyncService
}

func NewIngredientHandler() *IngredientHandler {
	return &IngredientHandler{
		bomService:    service.NewBOMService(),
		alertService:  service.NewCostAlertService(),
		inventorySync: service.NewInventorySyncService(),
	}
}

func (h *IngredientHandler) GetIngredients(c *gin.Context) {
	var query dto.IngredientQueryDTO
	if err := c.ShouldBindQuery(&query); err != nil {
		middleware.Error(c, "参数错误: "+err.Error())
		return
	}

	if query.Page <= 0 {
		query.Page = 1
	}
	if query.PageSize <= 0 {
		query.PageSize = 20
	}

	ingredients, total, err := h.bomService.GetIngredients(&query)
	if err != nil {
		middleware.Error(c, "获取食材列表失败: "+err.Error())
		return
	}

	middleware.PageSuccess(c, ingredients, total, query.Page, query.PageSize)
}

func (h *IngredientHandler) GetIngredient(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		middleware.Error(c, "无效的ID")
		return
	}

	ingredient, err := h.bomService.GetIngredientByID(uint(id))
	if err != nil {
		middleware.Error(c, "食材不存在: "+err.Error())
		return
	}

	middleware.Success(c, ingredient)
}

func (h *IngredientHandler) CreateIngredient(c *gin.Context) {
	var ingredient model.Ingredient
	if err := c.ShouldBindJSON(&ingredient); err != nil {
		middleware.Error(c, "参数错误: "+err.Error())
		return
	}

	if err := h.bomService.CreateIngredient(&ingredient); err != nil {
		middleware.Error(c, "创建食材失败: "+err.Error())
		return
	}

	middleware.Success(c, ingredient)
}

func (h *IngredientHandler) UpdateIngredient(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		middleware.Error(c, "无效的ID")
		return
	}

	var ingredient model.Ingredient
	if err := c.ShouldBindJSON(&ingredient); err != nil {
		middleware.Error(c, "参数错误: "+err.Error())
		return
	}
	ingredient.ID = uint(id)

	if err := h.bomService.UpdateIngredient(&ingredient); err != nil {
		middleware.Error(c, "更新食材失败: "+err.Error())
		return
	}

	middleware.Success(c, ingredient)
}

func (h *IngredientHandler) DeleteIngredient(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		middleware.Error(c, "无效的ID")
		return
	}

	if err := h.bomService.DeleteIngredient(uint(id)); err != nil {
		middleware.Error(c, "删除食材失败: "+err.Error())
		return
	}

	middleware.Success(c, gin.H{"message": "删除成功"})
}

func (h *IngredientHandler) GetIngredientCategories(c *gin.Context) {
	storeIDStr := c.Query("store_id")
	var storeID uint
	if storeIDStr != "" {
		if id, err := strconv.ParseUint(storeIDStr, 10, 32); err == nil {
			storeID = uint(id)
		}
	}

	categories, err := h.bomService.GetIngredientCategories(storeID)
	if err != nil {
		middleware.Error(c, "获取食材分类失败: "+err.Error())
		return
	}

	middleware.Success(c, categories)
}

func (h *IngredientHandler) GetPriceHistory(c *gin.Context) {
	var query dto.IngredientPriceQueryDTO
	if err := c.ShouldBindQuery(&query); err != nil {
		middleware.Error(c, "参数错误: "+err.Error())
		return
	}

	if query.Limit <= 0 {
		query.Limit = 30
	}

	prices, err := h.alertService.GetPriceHistory(query.IngredientID, query.StartDate, query.EndDate, query.Limit)
	if err != nil {
		middleware.Error(c, "获取价格历史失败: "+err.Error())
		return
	}

	middleware.Success(c, prices)
}

func (h *IngredientHandler) GetProductBOM(c *gin.Context) {
	productIDStr := c.Param("product_id")
	productID, err := strconv.ParseUint(productIDStr, 10, 32)
	if err != nil {
		middleware.Error(c, "无效的产品ID")
		return
	}

	skuIDStr := c.Query("sku_id")
	var skuID uint
	if skuIDStr != "" {
		if id, err := strconv.ParseUint(skuIDStr, 10, 32); err == nil {
			skuID = uint(id)
		}
	}

	items, err := h.bomService.GetProductBOM(uint(productID), skuID)
	if err != nil {
		middleware.Error(c, "获取BOM失败: "+err.Error())
		return
	}

	middleware.Success(c, items)
}

func (h *IngredientHandler) SaveProductBOM(c *gin.Context) {
	var body dto.BOMSaveDTO
	if err := c.ShouldBindJSON(&body); err != nil {
		middleware.Error(c, "参数错误: "+err.Error())
		return
	}

	if err := h.bomService.SaveProductBOMList(body.StoreID, body.ProductID, body.SKUID, body.Items); err != nil {
		middleware.Error(c, "保存BOM失败: "+err.Error())
		return
	}

	middleware.Success(c, gin.H{"message": "保存成功"})
}

func (h *IngredientHandler) GetProductCostDetail(c *gin.Context) {
	productIDStr := c.Param("product_id")
	productID, err := strconv.ParseUint(productIDStr, 10, 32)
	if err != nil {
		middleware.Error(c, "无效的产品ID")
		return
	}

	skuIDStr := c.Query("sku_id")
	var skuID uint
	if skuIDStr != "" {
		if id, err := strconv.ParseUint(skuIDStr, 10, 32); err == nil {
			skuID = uint(id)
		}
	}

	detail, err := h.bomService.GetProductCostDetail(uint(productID), skuID)
	if err != nil {
		middleware.Error(c, "获取产品成本详情失败: "+err.Error())
		return
	}

	middleware.Success(c, detail)
}

func (h *IngredientHandler) GetCostAlerts(c *gin.Context) {
	var query dto.CostAlertQueryDTO
	if err := c.ShouldBindQuery(&query); err != nil {
		middleware.Error(c, "参数错误: "+err.Error())
		return
	}

	if query.Page <= 0 {
		query.Page = 1
	}
	if query.PageSize <= 0 {
		query.PageSize = 20
	}

	alerts, total, err := h.alertService.GetAlerts(query.StoreID, query.Status, query.Page, query.PageSize)
	if err != nil {
		middleware.Error(c, "获取告警列表失败: "+err.Error())
		return
	}

	middleware.PageSuccess(c, alerts, total, query.Page, query.PageSize)
}

func (h *IngredientHandler) HandleAlert(c *gin.Context) {
	var body dto.CostAlertHandleDTO
	if err := c.ShouldBindJSON(&body); err != nil {
		middleware.Error(c, "参数错误: "+err.Error())
		return
	}

	if err := h.alertService.HandleAlert(body.AlertID, body.Handler, body.Remark); err != nil {
		middleware.Error(c, "处理告警失败: "+err.Error())
		return
	}

	middleware.Success(c, gin.H{"message": "处理成功"})
}

func (h *IngredientHandler) TriggerInventorySync(c *gin.Context) {
	go func() {
		h.inventorySync.SyncAllStores()
	}()
	middleware.Success(c, gin.H{"message": "进销存同步任务已启动"})
}
