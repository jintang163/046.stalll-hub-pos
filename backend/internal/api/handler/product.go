package handler

import (
	"net/http"
	"strconv"
	"stalll-hub-pos/backend/internal/database"
	"stalll-hub-pos/backend/internal/dto"
	"stalll-hub-pos/backend/internal/middleware"
	"stalll-hub-pos/backend/internal/model"
	"stalll-hub-pos/backend/internal/service"

	"github.com/gin-gonic/gin"
)

type ProductHandler struct {
	productService *service.ProductService
}

func NewProductHandler() *ProductHandler {
	return &ProductHandler{
		productService: service.NewProductService(),
	}
}

func (h *ProductHandler) Create(c *gin.Context) {
	var req dto.ProductCreateDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.Error(c, "参数错误: "+err.Error())
		return
	}

	if req.StoreID == 0 {
		req.StoreID = middleware.GetStoreID(c)
	}

	product, err := h.productService.CreateProduct(&req)
	if err != nil {
		middleware.Error(c, "创建商品失败: "+err.Error())
		return
	}

	middleware.Success(c, product)
}

func (h *ProductHandler) Update(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		middleware.Error(c, "无效的商品ID")
		return
	}

	var req dto.ProductUpdateDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.Error(c, "参数错误: "+err.Error())
		return
	}

	product, err := h.productService.UpdateProduct(uint(id), &req)
	if err != nil {
		middleware.Error(c, "更新商品失败: "+err.Error())
		return
	}

	middleware.Success(c, product)
}

func (h *ProductHandler) Delete(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		middleware.Error(c, "无效的商品ID")
		return
	}

	err = h.productService.DeleteProduct(uint(id))
	if err != nil {
		middleware.Error(c, "删除商品失败: "+err.Error())
		return
	}

	middleware.Success(c, nil)
}

func (h *ProductHandler) Get(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		middleware.Error(c, "无效的商品ID")
		return
	}

	product, err := h.productService.GetProduct(uint(id))
	if err != nil {
		middleware.Error(c, "获取商品失败: "+err.Error())
		return
	}

	middleware.Success(c, product)
}

func (h *ProductHandler) List(c *gin.Context) {
	var query dto.ProductQueryDTO
	if err := c.ShouldBindQuery(&query); err != nil {
		middleware.Error(c, "参数错误: "+err.Error())
		return
	}

	if query.StoreID == 0 {
		query.StoreID = middleware.GetStoreID(c)
	}

	products, total, err := h.productService.ListProducts(&query)
	if err != nil {
		middleware.Error(c, "获取商品列表失败: "+err.Error())
		return
	}

	middleware.Success(c, gin.H{
		"list":  products,
		"total": total,
		"page":  query.Page,
		"size":  query.PageSize,
	})
}

func (h *ProductHandler) UpdateStock(c *gin.Context) {
	var req dto.SKUStockUpdateDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.Error(c, "参数错误: "+err.Error())
		return
	}

	if req.StoreID == 0 {
		req.StoreID = middleware.GetStoreID(c)
	}

	err := h.productService.UpdateSKUStock(req.StoreID, req.Items)
	if err != nil {
		middleware.Error(c, "更新库存失败: "+err.Error())
		return
	}

	middleware.Success(c, nil)
}

func (h *ProductHandler) BatchUpdatePrice(c *gin.Context) {
	var req dto.BatchPriceUpdateDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.Error(c, "参数错误: "+err.Error())
		return
	}

	if req.StoreID == 0 {
		req.StoreID = middleware.GetStoreID(c)
	}

	err := h.productService.BatchUpdatePrice(&req)
	if err != nil {
		middleware.Error(c, "批量改价失败: "+err.Error())
		return
	}

	middleware.Success(c, nil)
}

func (h *ProductHandler) Copy(c *gin.Context) {
	var req dto.ProductCopyDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.Error(c, "参数错误: "+err.Error())
		return
	}

	if req.StoreID == 0 {
		req.StoreID = middleware.GetStoreID(c)
	}

	product, err := h.productService.CopyProduct(&req)
	if err != nil {
		middleware.Error(c, "复制商品失败: "+err.Error())
		return
	}

	middleware.Success(c, product)
}

func (h *ProductHandler) Sync(c *gin.Context) {
	storeIDStr := c.Query("store_id")
	lastSyncIDStr := c.Query("last_sync_id")
	limitStr := c.Query("limit")

	storeID, _ := strconv.ParseUint(storeIDStr, 10, 32)
	lastSyncID, _ := strconv.ParseUint(lastSyncIDStr, 10, 32)
	limit, _ := strconv.Atoi(limitStr)

	if limit == 0 || limit > 500 {
		limit = 100
	}

	result, err := h.productService.SyncProducts(uint(storeID), uint(lastSyncID), limit)
	if err != nil {
		middleware.Error(c, "同步商品失败: "+err.Error())
		return
	}

	middleware.Success(c, result)
}

func (h *ProductHandler) GetStockWarnings(c *gin.Context) {
	storeID := middleware.GetStoreID(c)
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	warnings, total, err := h.productService.GetStockWarnings(storeID, page, pageSize)
	if err != nil {
		middleware.Error(c, "获取库存预警失败: "+err.Error())
		return
	}

	middleware.Success(c, gin.H{
		"list":  warnings,
		"total": total,
		"page":  page,
		"size":  pageSize,
	})
}

func (h *ProductHandler) ListCategories(c *gin.Context) {
	storeIDStr := c.Query("store_id")
	storeID, _ := strconv.ParseUint(storeIDStr, 10, 32)
	if storeID == 0 {
		storeID = uint64(middleware.GetStoreID(c))
	}

	var categories []model.Category
	err := database.DB.Where("store_id = ? AND status = ?", storeID, 1).
		Order("sort_order ASC, id ASC").
		Find(&categories).Error
	if err != nil {
		middleware.Error(c, "获取分类列表失败: "+err.Error())
		return
	}

	middleware.Success(c, categories)
}
