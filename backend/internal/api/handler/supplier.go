package handler

import (
	"strconv"
	"stalll-hub-pos/backend/internal/dto"
	"stalll-hub-pos/backend/internal/middleware"
	"stalll-hub-pos/backend/internal/service"

	"github.com/gin-gonic/gin"
)

type SupplierHandler struct {
	supplierService *service.SupplierService
}

func NewSupplierHandler() *SupplierHandler {
	return &SupplierHandler{
		supplierService: service.NewSupplierService(),
	}
}

func (h *SupplierHandler) CreateSupplier(c *gin.Context) {
	var req dto.SupplierCreateDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.Error(c, "参数错误: "+err.Error())
		return
	}

	resp, err := h.supplierService.CreateSupplier(&req)
	if err != nil {
		middleware.Error(c, "创建供应商失败: "+err.Error())
		return
	}

	middleware.Success(c, resp)
}

func (h *SupplierHandler) UpdateSupplier(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		middleware.Error(c, "无效的ID")
		return
	}

	var req dto.SupplierUpdateDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.Error(c, "参数错误: "+err.Error())
		return
	}

	resp, err := h.supplierService.UpdateSupplier(uint(id), &req)
	if err != nil {
		middleware.Error(c, "更新供应商失败: "+err.Error())
		return
	}

	middleware.Success(c, resp)
}

func (h *SupplierHandler) DeleteSupplier(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		middleware.Error(c, "无效的ID")
		return
	}

	if err := h.supplierService.DeleteSupplier(uint(id)); err != nil {
		middleware.Error(c, "删除供应商失败: "+err.Error())
		return
	}

	middleware.Success(c, gin.H{"message": "删除成功"})
}

func (h *SupplierHandler) GetSupplier(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		middleware.Error(c, "无效的ID")
		return
	}

	resp, err := h.supplierService.GetSupplier(uint(id))
	if err != nil {
		middleware.Error(c, "获取供应商失败: "+err.Error())
		return
	}

	middleware.Success(c, resp)
}

func (h *SupplierHandler) ListSuppliers(c *gin.Context) {
	var query dto.SupplierQueryDTO
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

	resp, err := h.supplierService.ListSuppliers(&query)
	if err != nil {
		middleware.Error(c, "获取供应商列表失败: "+err.Error())
		return
	}

	middleware.PageSuccess(c, resp.List, resp.Total, resp.Page, resp.Size)
}

func (h *SupplierHandler) GetSupplierCategories(c *gin.Context) {
	storeIDStr := c.Query("store_id")
	var storeID uint
	if storeIDStr != "" {
		if id, err := strconv.ParseUint(storeIDStr, 10, 32); err == nil {
			storeID = uint(id)
		}
	}

	categories, err := h.supplierService.GetSupplierCategories(storeID)
	if err != nil {
		middleware.Error(c, "获取供应商分类失败: "+err.Error())
		return
	}

	middleware.Success(c, categories)
}

func (h *SupplierHandler) GetSupplierStats(c *gin.Context) {
	storeIDStr := c.Query("store_id")
	var storeID uint
	if storeIDStr != "" {
		if id, err := strconv.ParseUint(storeIDStr, 10, 32); err == nil {
			storeID = uint(id)
		}
	}

	stats, err := h.supplierService.GetSupplierStats(storeID)
	if err != nil {
		middleware.Error(c, "获取供应商统计失败: "+err.Error())
		return
	}

	middleware.Success(c, stats)
}

func (h *SupplierHandler) NotifySupplier(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		middleware.Error(c, "无效的ID")
		return
	}

	var req dto.NotifySupplierDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.Error(c, "参数错误: "+err.Error())
		return
	}

	if len(req.NotifyType) == 0 {
		req.NotifyType = []string{"sms"}
	}

	if err := h.supplierService.NotifySupplier(uint(id), req.NotifyType, req.Content); err != nil {
		middleware.Error(c, "通知供应商失败: "+err.Error())
		return
	}

	middleware.Success(c, gin.H{"message": "通知已发送"})
}

func (h *SupplierHandler) GetEnums(c *gin.Context) {
	middleware.Success(c, gin.H{
		"payment_terms":      service.GetPaymentTermMap(),
		"settlement_methods": service.GetSettlementMethodMap(),
		"payable_statuses":   service.GetPayableStatusMap(),
		"reconcile_statuses": service.GetReconcileStatusMap(),
	})
}

type PurchaseReceiveHandler struct {
	receiveService *service.PurchaseReceiveService
}

func NewPurchaseReceiveHandler() *PurchaseReceiveHandler {
	return &PurchaseReceiveHandler{
		receiveService: service.NewPurchaseReceiveService(),
	}
}

func (h *PurchaseReceiveHandler) CreateReceive(c *gin.Context) {
	var req dto.PurchaseReceiveCreateDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.Error(c, "参数错误: "+err.Error())
		return
	}

	resp, err := h.receiveService.CreateReceive(&req)
	if err != nil {
		middleware.Error(c, "创建收货单失败: "+err.Error())
		return
	}

	middleware.Success(c, resp)
}

func (h *PurchaseReceiveHandler) GetReceive(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		middleware.Error(c, "无效的ID")
		return
	}

	resp, err := h.receiveService.GetReceive(uint(id))
	if err != nil {
		middleware.Error(c, "获取收货单失败: "+err.Error())
		return
	}

	middleware.Success(c, resp)
}

func (h *PurchaseReceiveHandler) ListReceives(c *gin.Context) {
	var query dto.PurchaseReceiveQueryDTO
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

	resp, err := h.receiveService.ListReceives(&query)
	if err != nil {
		middleware.Error(c, "获取收货单列表失败: "+err.Error())
		return
	}

	middleware.PageSuccess(c, resp.List, resp.Total, resp.Page, resp.Size)
}

type PurchaseOrderV2Handler struct {
	purchaseService *service.PurchaseOrderV2Service
}

func NewPurchaseOrderV2Handler() *PurchaseOrderV2Handler {
	return &PurchaseOrderV2Handler{
		purchaseService: service.NewPurchaseOrderV2Service(),
	}
}

func (h *PurchaseOrderV2Handler) CreatePurchaseOrder(c *gin.Context) {
	var req dto.PurchaseOrderCreateV2DTO
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.Error(c, "参数错误: "+err.Error())
		return
	}

	resp, err := h.purchaseService.CreatePurchaseOrder(&req)
	if err != nil {
		middleware.Error(c, "创建采购订单失败: "+err.Error())
		return
	}

	middleware.Success(c, resp)
}

func (h *PurchaseOrderV2Handler) GetPurchaseOrder(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		middleware.Error(c, "无效的ID")
		return
	}

	resp, err := h.purchaseService.GetPurchaseOrder(uint(id))
	if err != nil {
		middleware.Error(c, "获取采购订单失败: "+err.Error())
		return
	}

	middleware.Success(c, resp)
}

func (h *PurchaseOrderV2Handler) ListPurchaseOrders(c *gin.Context) {
	var query dto.PurchaseOrderQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		middleware.Error(c, "参数错误: "+err.Error())
		return
	}

	if query.Page <= 0 {
		query.Page = 1
	}
	if query.Size <= 0 {
		query.Size = 20
	}

	resp, err := h.purchaseService.ListPurchaseOrders(&query)
	if err != nil {
		middleware.Error(c, "获取采购订单列表失败: "+err.Error())
		return
	}

	middleware.PageSuccess(c, resp.List, resp.Total, resp.Page, resp.Size)
}

func (h *PurchaseOrderV2Handler) SendToSupplier(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		middleware.Error(c, "无效的ID")
		return
	}

	var req dto.NotifySupplierDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		req.NotifyType = []string{"sms", "email"}
		req.Content = ""
	}

	if err := h.purchaseService.SendToSupplier(uint(id), req.NotifyType, req.Content); err != nil {
		middleware.Error(c, "发送采购订单失败: "+err.Error())
		return
	}

	middleware.Success(c, gin.H{"message": "已发送给供应商"})
}

func (h *PurchaseOrderV2Handler) CompletePurchase(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		middleware.Error(c, "无效的ID")
		return
	}

	if err := h.purchaseService.CompletePurchase(uint(id)); err != nil {
		middleware.Error(c, "完成采购订单失败: "+err.Error())
		return
	}

	middleware.Success(c, gin.H{"message": "采购订单已完成，应付账款已生成"})
}

func (h *PurchaseOrderV2Handler) CancelPurchase(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		middleware.Error(c, "无效的ID")
		return
	}

	var body struct {
		Remark string `json:"remark"`
	}
	c.ShouldBindJSON(&body)

	if err := h.purchaseService.CancelPurchase(uint(id), body.Remark); err != nil {
		middleware.Error(c, "取消采购订单失败: "+err.Error())
		return
	}

	middleware.Success(c, gin.H{"message": "采购订单已取消"})
}

type AccountsPayableHandler struct {
	payableService *service.AccountsPayableService
}

func NewAccountsPayableHandler() *AccountsPayableHandler {
	return &AccountsPayableHandler{
		payableService: service.NewAccountsPayableService(),
	}
}

func (h *AccountsPayableHandler) GetPayable(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		middleware.Error(c, "无效的ID")
		return
	}

	resp, err := h.payableService.GetPayable(uint(id))
	if err != nil {
		middleware.Error(c, "获取应付账款失败: "+err.Error())
		return
	}

	middleware.Success(c, resp)
}

func (h *AccountsPayableHandler) ListPayables(c *gin.Context) {
	var query dto.PayableQueryDTO
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

	resp, err := h.payableService.ListPayables(&query)
	if err != nil {
		middleware.Error(c, "获取应付账款列表失败: "+err.Error())
		return
	}

	middleware.PageSuccess(c, resp.List, resp.Total, resp.Page, resp.Size)
}

func (h *AccountsPayableHandler) GetPayableStats(c *gin.Context) {
	var query dto.PayableQueryDTO
	if err := c.ShouldBindQuery(&query); err != nil {
		middleware.Error(c, "参数错误: "+err.Error())
		return
	}

	stats, err := h.payableService.GetPayableStats(&query)
	if err != nil {
		middleware.Error(c, "获取应付统计失败: "+err.Error())
		return
	}

	middleware.Success(c, stats)
}

func (h *AccountsPayableHandler) CreatePayment(c *gin.Context) {
	var req dto.PayablePaymentCreateDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.Error(c, "参数错误: "+err.Error())
		return
	}

	resp, err := h.payableService.CreatePayment(&req)
	if err != nil {
		middleware.Error(c, "创建付款记录失败: "+err.Error())
		return
	}

	middleware.Success(c, resp)
}

func (h *AccountsPayableHandler) ListPayments(c *gin.Context) {
	var query dto.PayablePaymentQueryDTO
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

	resp, err := h.payableService.ListPayments(&query)
	if err != nil {
		middleware.Error(c, "获取付款记录列表失败: "+err.Error())
		return
	}

	middleware.PageSuccess(c, resp.List, resp.Total, resp.Page, resp.Size)
}

func (h *AccountsPayableHandler) CreateReconciliation(c *gin.Context) {
	var req dto.ReconciliationCreateDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.Error(c, "参数错误: "+err.Error())
		return
	}

	resp, err := h.payableService.CreateReconciliation(&req)
	if err != nil {
		middleware.Error(c, "创建对账单失败: "+err.Error())
		return
	}

	middleware.Success(c, resp)
}

func (h *AccountsPayableHandler) GetReconciliation(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		middleware.Error(c, "无效的ID")
		return
	}

	resp, err := h.payableService.GetReconciliation(uint(id))
	if err != nil {
		middleware.Error(c, "获取对账单失败: "+err.Error())
		return
	}

	middleware.Success(c, resp)
}

func (h *AccountsPayableHandler) ListReconciliations(c *gin.Context) {
	var query dto.ReconciliationQueryDTO
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

	resp, err := h.payableService.ListReconciliations(&query)
	if err != nil {
		middleware.Error(c, "获取对账单列表失败: "+err.Error())
		return
	}

	middleware.PageSuccess(c, resp.List, resp.Total, resp.Page, resp.Size)
}

func (h *AccountsPayableHandler) ConfirmReconciliation(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		middleware.Error(c, "无效的ID")
		return
	}

	var req dto.ReconciliationConfirmDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.Error(c, "参数错误: "+err.Error())
		return
	}

	resp, err := h.payableService.ConfirmReconciliation(uint(id), &req)
	if err != nil {
		middleware.Error(c, "确认对账单失败: "+err.Error())
		return
	}

	middleware.Success(c, resp)
}

func (h *AccountsPayableHandler) UpdateOverdueStatus(c *gin.Context) {
	var req struct {
		StoreID uint `json:"store_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.ShouldBindQuery(&req)
	}

	h.payableService.UpdateOverdueStatus()

	middleware.Success(c, gin.H{"message": "逾期状态更新成功"})
}

func (h *AccountsPayableHandler) InputSupplierAmount(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		middleware.Error(c, "无效的ID")
		return
	}

	var req dto.ReconciliationSupplierAmountDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.Error(c, "参数错误: "+err.Error())
		return
	}

	resp, err := h.payableService.InputSupplierAmount(uint(id), &req)
	if err != nil {
		middleware.Error(c, "录入供应商金额失败: "+err.Error())
		return
	}

	middleware.Success(c, resp)
}
