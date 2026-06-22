package handler

import (
	"strconv"
	"stalll-hub-pos/backend/internal/dto"
	"stalll-hub-pos/backend/internal/middleware"
	"stalll-hub-pos/backend/internal/service"

	"github.com/gin-gonic/gin"
)

type TransferHandler struct {
	transferService *service.TransferService
	kdnService      *service.KuaiDiNiaoService
}

var transferHandlerInstance *TransferHandler

func NewTransferHandler() *TransferHandler {
	if transferHandlerInstance == nil {
		transferHandlerInstance = &TransferHandler{
			transferService: service.NewTransferService(),
			kdnService:      service.NewKuaiDiNiaoService(),
		}
	}
	return transferHandlerInstance
}

func (h *TransferHandler) CreateTransfer(c *gin.Context) {
	var req dto.CreateTransferOrderDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.Error(c, "参数错误: "+err.Error())
		return
	}

	result, err := h.transferService.CreateTransfer(&req)
	if err != nil {
		middleware.Error(c, "创建调拨单失败: "+err.Error())
		return
	}

	middleware.Success(c, result)
}

func (h *TransferHandler) GetTransfer(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		middleware.Error(c, "无效的ID")
		return
	}

	result, err := h.transferService.GetTransferByID(uint(id))
	if err != nil {
		middleware.Error(c, "调拨单不存在: "+err.Error())
		return
	}

	middleware.Success(c, result)
}

func (h *TransferHandler) ListTransfers(c *gin.Context) {
	var query dto.TransferOrderQueryDTO
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

	list, total, err := h.transferService.ListTransfers(&query)
	if err != nil {
		middleware.Error(c, "获取调拨单列表失败: "+err.Error())
		return
	}

	middleware.PageSuccess(c, list, total, query.Page, query.PageSize)
}

func (h *TransferHandler) AcceptTransfer(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		middleware.Error(c, "无效的ID")
		return
	}

	storeID := middleware.GetStoreID(c)

	var body struct {
		OperatorID   uint   `json:"operator_id"`
		OperatorName string `json:"operator_name"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		middleware.Error(c, "参数错误: "+err.Error())
		return
	}

	result, err := h.transferService.AcceptTransfer(uint(id), body.OperatorID, body.OperatorName, storeID)
	if err != nil {
		middleware.Error(c, "确认接单失败: "+err.Error())
		return
	}

	middleware.Success(c, result)
}

func (h *TransferHandler) RejectTransfer(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		middleware.Error(c, "无效的ID")
		return
	}

	storeID := middleware.GetStoreID(c)

	var body struct {
		OperatorID   uint   `json:"operator_id"`
		OperatorName string `json:"operator_name"`
		Reason       string `json:"reason"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		middleware.Error(c, "参数错误: "+err.Error())
		return
	}

	result, err := h.transferService.RejectTransfer(uint(id), body.OperatorID, body.OperatorName, storeID, body.Reason)
	if err != nil {
		middleware.Error(c, "拒单失败: "+err.Error())
		return
	}

	middleware.Success(c, result)
}

func (h *TransferHandler) ConfirmOutbound(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		middleware.Error(c, "无效的ID")
		return
	}

	storeID := middleware.GetStoreID(c)

	var req dto.ConfirmOutboundDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.Error(c, "参数错误: "+err.Error())
		return
	}

	result, err := h.transferService.ConfirmOutbound(uint(id), req.OperatorID, req.OperatorName, storeID, req.Remark)
	if err != nil {
		middleware.Error(c, "确认出库失败: "+err.Error())
		return
	}

	middleware.Success(c, result)
}

func (h *TransferHandler) StartShipping(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		middleware.Error(c, "无效的ID")
		return
	}

	storeID := middleware.GetStoreID(c)

	var req dto.UpdateLogisticsDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.Error(c, "参数错误: "+err.Error())
		return
	}

	result, err := h.transferService.StartShipping(uint(id), req.LogisticsCompany, req.TrackingNo, req.LogisticsCode, storeID)
	if err != nil {
		middleware.Error(c, "发货失败: "+err.Error())
		return
	}

	middleware.Success(c, result)
}

func (h *TransferHandler) ReceiveTransfer(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		middleware.Error(c, "无效的ID")
		return
	}

	storeID := middleware.GetStoreID(c)

	var req dto.ReceiveTransferDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.Error(c, "参数错误: "+err.Error())
		return
	}

	result, err := h.transferService.ReceiveTransfer(uint(id), &req, storeID)
	if err != nil {
		middleware.Error(c, "收货失败: "+err.Error())
		return
	}

	middleware.Success(c, result)
}

func (h *TransferHandler) CompleteTransfer(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		middleware.Error(c, "无效的ID")
		return
	}

	storeID := middleware.GetStoreID(c)

	var body struct {
		DiffRemark string `json:"diff_remark"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		middleware.Error(c, "参数错误: "+err.Error())
		return
	}

	result, err := h.transferService.CompleteTransfer(uint(id), body.DiffRemark, storeID)
	if err != nil {
		middleware.Error(c, "完成调拨失败: "+err.Error())
		return
	}

	middleware.Success(c, result)
}

func (h *TransferHandler) CancelTransfer(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		middleware.Error(c, "无效的ID")
		return
	}

	storeID := middleware.GetStoreID(c)

	var body struct {
		Remark string `json:"remark"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		middleware.Error(c, "参数错误: "+err.Error())
		return
	}

	result, err := h.transferService.CancelTransfer(uint(id), body.Remark, storeID)
	if err != nil {
		middleware.Error(c, "取消调拨失败: "+err.Error())
		return
	}

	middleware.Success(c, result)
}

func (h *TransferHandler) GetLogisticsTrack(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		middleware.Error(c, "无效的ID")
		return
	}

	transfer, err := h.transferService.GetTransferByID(uint(id))
	if err != nil {
		middleware.Error(c, "调拨单不存在: "+err.Error())
		return
	}

	if transfer.TrackingNo == "" {
		tracks, err := h.transferService.GetLogisticsTracks(uint(id))
		if err != nil {
			middleware.Error(c, "获取物流轨迹失败: "+err.Error())
			return
		}
		middleware.Success(c, map[string]interface{}{
			"tracking_no": "",
			"status":      "",
			"tracks":      tracks,
		})
		return
	}

	tracks, err := h.transferService.RefreshLogisticsTrack(uint(id))
	if err != nil {
		localTracks, _ := h.transferService.GetLogisticsTracks(uint(id))
		middleware.Success(c, map[string]interface{}{
			"tracking_no": transfer.TrackingNo,
			"status":      "",
			"tracks":      localTracks,
			"error":       err.Error(),
		})
		return
	}

	middleware.Success(c, map[string]interface{}{
		"tracking_no": transfer.TrackingNo,
		"status":      "",
		"tracks":      tracks,
	})
}

func (h *TransferHandler) RefreshLogistics(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		middleware.Error(c, "无效的ID")
		return
	}

	tracks, err := h.transferService.RefreshLogisticsTrack(uint(id))
	if err != nil {
		middleware.Error(c, "刷新物流信息失败: "+err.Error())
		return
	}

	middleware.Success(c, tracks)
}

func (h *TransferHandler) GetTransferItems(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		middleware.Error(c, "无效的ID")
		return
	}

	items, err := h.transferService.GetItems(uint(id))
	if err != nil {
		middleware.Error(c, "获取调拨明细失败: "+err.Error())
		return
	}

	middleware.Success(c, items)
}
