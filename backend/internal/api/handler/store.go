package handler

import (
	"net/http"
	"strconv"
	"stalll-hub-pos/backend/internal/dto"
	"stalll-hub-pos/backend/internal/middleware"
	"stalll-hub-pos/backend/internal/service"

	"github.com/gin-gonic/gin"
)

type StoreHandler struct {
	storeService *service.StoreService
}

func NewStoreHandler() *StoreHandler {
	return &StoreHandler{
		storeService: service.NewStoreService(),
	}
}

func (h *StoreHandler) CreateStore(c *gin.Context) {
	var req dto.StoreCreateDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.Error(c, "参数错误: "+err.Error())
		return
	}

	store, err := h.storeService.CreateStore(&req)
	if err != nil {
		middleware.Error(c, "创建门店失败: "+err.Error())
		return
	}

	middleware.Success(c, store)
}

func (h *StoreHandler) UpdateStore(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		middleware.Error(c, "无效的门店ID")
		return
	}

	var req dto.StoreUpdateDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.Error(c, "参数错误: "+err.Error())
		return
	}

	store, err := h.storeService.UpdateStore(uint(id), &req)
	if err != nil {
		middleware.Error(c, "更新门店失败: "+err.Error())
		return
	}

	middleware.Success(c, store)
}

func (h *StoreHandler) DeleteStore(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		middleware.Error(c, "无效的门店ID")
		return
	}

	err = h.storeService.DeleteStore(uint(id))
	if err != nil {
		middleware.Error(c, "删除门店失败: "+err.Error())
		return
	}

	middleware.Success(c, nil)
}

func (h *StoreHandler) GetStore(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		middleware.Error(c, "无效的门店ID")
		return
	}

	store, err := h.storeService.GetStore(uint(id))
	if err != nil {
		middleware.Error(c, "获取门店失败: "+err.Error())
		return
	}

	middleware.Success(c, store)
}

func (h *StoreHandler) ListStores(c *gin.Context) {
	var query dto.StoreQueryDTO
	if err := c.ShouldBindQuery(&query); err != nil {
		middleware.Error(c, "参数错误: "+err.Error())
		return
	}

	stores, total, err := h.storeService.ListStores(&query)
	if err != nil {
		middleware.Error(c, "获取门店列表失败: "+err.Error())
		return
	}

	middleware.Success(c, gin.H{
		"list":  stores,
		"total": total,
		"page":  query.Page,
		"size":  query.PageSize,
	})
}

func (h *StoreHandler) CreatePrinter(c *gin.Context) {
	var req dto.PrinterCreateDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.Error(c, "参数错误: "+err.Error())
		return
	}

	if req.StoreID == 0 {
		req.StoreID = middleware.GetStoreID(c)
	}

	printer, err := h.storeService.CreatePrinter(&req)
	if err != nil {
		middleware.Error(c, "创建打印机失败: "+err.Error())
		return
	}

	middleware.Success(c, printer)
}

func (h *StoreHandler) UpdatePrinter(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		middleware.Error(c, "无效的打印机ID")
		return
	}

	var req dto.PrinterUpdateDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.Error(c, "参数错误: "+err.Error())
		return
	}

	printer, err := h.storeService.UpdatePrinter(uint(id), &req)
	if err != nil {
		middleware.Error(c, "更新打印机失败: "+err.Error())
		return
	}

	middleware.Success(c, printer)
}

func (h *StoreHandler) DeletePrinter(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		middleware.Error(c, "无效的打印机ID")
		return
	}

	err = h.storeService.DeletePrinter(uint(id))
	if err != nil {
		middleware.Error(c, "删除打印机失败: "+err.Error())
		return
	}

	middleware.Success(c, nil)
}

func (h *StoreHandler) GetPrinter(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		middleware.Error(c, "无效的打印机ID")
		return
	}

	printer, err := h.storeService.GetPrinter(uint(id))
	if err != nil {
		middleware.Error(c, "获取打印机失败: "+err.Error())
		return
	}

	middleware.Success(c, printer)
}

func (h *StoreHandler) ListPrinters(c *gin.Context) {
	var query dto.PrinterQueryDTO
	if err := c.ShouldBindQuery(&query); err != nil {
		middleware.Error(c, "参数错误: "+err.Error())
		return
	}

	if query.StoreID == 0 {
		query.StoreID = middleware.GetStoreID(c)
	}

	printers, total, err := h.storeService.ListPrinters(&query)
	if err != nil {
		middleware.Error(c, "获取打印机列表失败: "+err.Error())
		return
	}

	middleware.Success(c, gin.H{
		"list":  printers,
		"total": total,
		"page":  query.Page,
		"size":  query.PageSize,
	})
}

func (h *StoreHandler) PrintTest(c *gin.Context) {
	var req dto.PrintTestDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.Error(c, "参数错误: "+err.Error())
		return
	}

	err := h.storeService.PrintTest(req.PrinterID)
	if err != nil {
		middleware.Error(c, "打印测试失败: "+err.Error())
		return
	}

	middleware.Success(c, gin.H{"message": "打印测试成功"})
}
