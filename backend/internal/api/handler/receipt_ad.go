package handler

import (
	"net/http"
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

func (h *ReceiptAdHandler) GetActiveAdsInternal(c *gin.Context) {
	storeIDStr := c.Query("store_id")
	if storeIDStr == "" {
		middleware.Error(c, "缺少 store_id 参数")
		return
	}
	storeID, err := strconv.ParseUint(storeIDStr, 10, 32)
	if err != nil {
		middleware.Error(c, "store_id 格式错误")
		return
	}
	position := c.Query("position")
	if position == "" {
		position = "footer"
	}

	ads, err := h.adService.GetActiveAds(uint(storeID), position)
	if err != nil {
		middleware.Error(c, err.Error())
		return
	}

	middleware.Success(c, ads)
}

func (h *ReceiptAdHandler) IncrementViewCountInternal(c *gin.Context) {
	var body struct {
		AdID uint `json:"ad_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		middleware.Error(c, "参数错误")
		return
	}
	h.adService.IncrementViewCount(body.AdID)
	middleware.Success(c, nil)
}

func (h *ReceiptAdHandler) RecordClickPublic(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.Redirect(http.StatusFound, "/")
		return
	}

	orderID := c.Query("order_id")
	orderNo := c.Query("order_no")

	storeID, ad, err := h.adService.GetAdStoreID(uint(id))
	if err != nil {
		c.Redirect(http.StatusFound, "/")
		return
	}

	orderIDUint := uint(0)
	if orderID != "" {
		if oid, err := strconv.ParseUint(orderID, 10, 32); err == nil {
			orderIDUint = uint(oid)
		}
	}

	ip := c.ClientIP()
	userAgent := c.GetHeader("User-Agent")

	req := &dto.ReceiptAdClickRequest{
		AdID:      uint(id),
		OrderID:   orderIDUint,
		OrderNo:   orderNo,
		ClickType: "scan",
		IP:        ip,
		UserAgent: userAgent,
	}
	_ = h.adService.RecordClick(storeID, req)

	if ad.LinkURL != "" {
		c.Redirect(http.StatusFound, ad.LinkURL)
		return
	}

	if ad.QRCodeContent != "" {
		c.Redirect(http.StatusFound, ad.QRCodeContent)
		return
	}

	c.Redirect(http.StatusFound, "/")
}

func (h *ReceiptAdHandler) GetAdDetailPublic(c *gin.Context) {
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

func (h *ReceiptAdHandler) RecordClickView(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		middleware.Error(c, "无效的ID")
		return
	}

	storeID, _, err := h.adService.GetAdStoreID(uint(id))
	if err != nil {
		middleware.Error(c, err.Error())
		return
	}

	orderID := c.Query("order_id")
	orderNo := c.Query("order_no")

	orderIDUint := uint(0)
	if orderID != "" {
		if oid, err := strconv.ParseUint(orderID, 10, 32); err == nil {
			orderIDUint = uint(oid)
		}
	}

	ip := c.ClientIP()
	userAgent := c.GetHeader("User-Agent")

	req := &dto.ReceiptAdClickRequest{
		AdID:      uint(id),
		OrderID:   orderIDUint,
		OrderNo:   orderNo,
		ClickType: "view",
		IP:        ip,
		UserAgent: userAgent,
	}
	_ = h.adService.RecordClick(storeID, req)
	middleware.Success(c, nil)
}
