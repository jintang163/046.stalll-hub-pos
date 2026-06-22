package handler

import (
	"strconv"
	"stalll-hub-pos/backend/internal/dto"
	"stalll-hub-pos/backend/internal/middleware"
	"stalll-hub-pos/backend/internal/service"

	"github.com/gin-gonic/gin"
)

type ReceiptAdHandler struct {
	adService *service.ReceiptAdService
}

func NewReceiptAdHandler() *ReceiptAdHandler {
	return &ReceiptAdHandler{
		adService: service.NewReceiptAdService(),
	}
}

func (h *ReceiptAdHandler) Create(c *gin.Context) {
	storeID := middleware.GetStoreID(c)
	var req dto.ReceiptAdDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.Error(c, "参数错误: "+err.Error())
		return
	}

	ad, err := h.adService.CreateReceiptAd(storeID, &req)
	if err != nil {
		middleware.Error(c, err.Error())
		return
	}

	middleware.Success(c, ad)
}

func (h *ReceiptAdHandler) Update(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		middleware.Error(c, "无效的ID")
		return
	}

	var req dto.ReceiptAdDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.Error(c, "参数错误: "+err.Error())
		return
	}

	ad, err := h.adService.UpdateReceiptAd(uint(id), &req)
	if err != nil {
		middleware.Error(c, err.Error())
		return
	}

	middleware.Success(c, ad)
}

func (h *ReceiptAdHandler) Delete(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		middleware.Error(c, "无效的ID")
		return
	}

	err = h.adService.DeleteReceiptAd(uint(id))
	if err != nil {
		middleware.Error(c, err.Error())
		return
	}

	middleware.Success(c, nil)
}

func (h *ReceiptAdHandler) Get(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		middleware.Error(c, "无效的ID")
		return
	}

	ad, err := h.adService.GetReceiptAd(uint(id))
	if err != nil {
		middleware.Error(c, err.Error())
		return
	}

	middleware.Success(c, ad)
}

func (h *ReceiptAdHandler) List(c *gin.Context) {
	storeID := middleware.GetStoreID(c)
	var req dto.ReceiptAdListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		middleware.Error(c, "参数错误: "+err.Error())
		return
	}

	list, total, err := h.adService.ListReceiptAds(storeID, &req)
	if err != nil {
		middleware.Error(c, err.Error())
		return
	}

	middleware.PageSuccess(c, list, total, req.Page, req.PageSize)
}

func (h *ReceiptAdHandler) UpdateStatus(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		middleware.Error(c, "无效的ID")
		return
	}

	var body struct {
		Status int `json:"status"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		middleware.Error(c, "参数错误")
		return
	}

	err = h.adService.UpdateStatus(uint(id), body.Status)
	if err != nil {
		middleware.Error(c, err.Error())
		return
	}

	middleware.Success(c, nil)
}

func (h *ReceiptAdHandler) RecordClick(c *gin.Context) {
	storeID := middleware.GetStoreID(c)
	var req dto.ReceiptAdClickRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.Error(c, "参数错误: "+err.Error())
		return
	}

	err := h.adService.RecordClick(storeID, &req)
	if err != nil {
		middleware.Error(c, err.Error())
		return
	}

	middleware.Success(c, nil)
}

func (h *ReceiptAdHandler) GetStats(c *gin.Context) {
	storeID := middleware.GetStoreID(c)
	var req dto.ReceiptAdStatsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		middleware.Error(c, "参数错误: "+err.Error())
		return
	}

	stats, err := h.adService.GetStats(storeID, &req)
	if err != nil {
		middleware.Error(c, err.Error())
		return
	}

	middleware.Success(c, stats)
}

func (h *ReceiptAdHandler) GetActiveAds(c *gin.Context) {
	storeID := middleware.GetStoreID(c)
	position := c.Query("position")
	if position == "" {
		position = "footer"
	}

	ads, err := h.adService.GetActiveAds(storeID, position)
	if err != nil {
		middleware.Error(c, err.Error())
		return
	}

	middleware.Success(c, ads)
}
