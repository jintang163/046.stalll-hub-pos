package handler

import (
	"fmt"
	"strconv"
	"stalll-hub-pos/backend/internal/dto"
	"stalll-hub-pos/backend/internal/middleware"
	"stalll-hub-pos/backend/internal/service"

	"github.com/gin-gonic/gin"
)

type ReviewHandler struct {
	reviewService *service.ReviewService
}

func NewReviewHandler() *ReviewHandler {
	return &ReviewHandler{
		reviewService: service.NewReviewService(),
	}
}

func (h *ReviewHandler) SaveAuth(c *gin.Context) {
	var req dto.PlatformAuthDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.Error(c, "参数错误: "+err.Error())
		return
	}

	err := h.reviewService.SaveAuth(&req)
	if err != nil {
		middleware.Error(c, "保存平台授权失败: "+err.Error())
		return
	}

	middleware.Success(c, gin.H{"message": "授权保存成功"})
}

func (h *ReviewHandler) GetAuth(c *gin.Context) {
	storeIDStr := c.Query("store_id")
	storeID, err := strconv.ParseUint(storeIDStr, 10, 32)
	if err != nil {
		middleware.Error(c, "无效的门店ID")
		return
	}

	platform := c.Query("platform")
	if platform == "" {
		middleware.Error(c, "平台参数不能为空")
		return
	}

	auth, err := h.reviewService.GetAuth(uint(storeID), platform)
	if err != nil {
		middleware.Error(c, "获取平台授权失败: "+err.Error())
		return
	}

	middleware.Success(c, auth)
}

func (h *ReviewHandler) ListAuths(c *gin.Context) {
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

	auths, err := h.reviewService.ListAuths(storeID)
	if err != nil {
		middleware.Error(c, "获取平台授权列表失败: "+err.Error())
		return
	}

	middleware.Success(c, auths)
}

func (h *ReviewHandler) SyncReviews(c *gin.Context) {
	var req dto.SyncRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.Error(c, "参数错误: "+err.Error())
		return
	}

	go func() {
		if _, err := h.reviewService.SyncStoreReviews(req.StoreID, req.Platform); err != nil {
			fmt.Printf("[Review] Sync store %d platform %s error: %v\n", req.StoreID, req.Platform, err)
		}
	}()

	middleware.Success(c, gin.H{"message": "同步任务已启动"})
}

func (h *ReviewHandler) SyncAll(c *gin.Context) {
	go func() {
		if _, err := h.reviewService.SyncAllStores(); err != nil {
			fmt.Printf("[Review] Sync all stores error: %v\n", err)
		}
	}()

	middleware.Success(c, gin.H{"message": "全门店同步任务已启动"})
}

func (h *ReviewHandler) ListRatings(c *gin.Context) {
	var query dto.ReviewRatingQueryDTO
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

	if query.StoreID == 0 {
		query.StoreID = middleware.GetStoreID(c)
	}

	result, err := h.reviewService.ListRatings(&query)
	if err != nil {
		middleware.Error(c, "获取评分列表失败: "+err.Error())
		return
	}

	middleware.PageSuccess(c, result.List, result.Total, result.Page, result.Size)
}

func (h *ReviewHandler) GetRatingTrend(c *gin.Context) {
	var query dto.ReviewRatingQueryDTO
	if err := c.ShouldBindQuery(&query); err != nil {
		middleware.Error(c, "参数错误: "+err.Error())
		return
	}

	if query.StoreID == 0 {
		query.StoreID = middleware.GetStoreID(c)
	}

	trends, err := h.reviewService.GetRatingTrend(&query)
	if err != nil {
		middleware.Error(c, "获取评分趋势失败: "+err.Error())
		return
	}

	middleware.Success(c, trends)
}

func (h *ReviewHandler) ListReviews(c *gin.Context) {
	var query dto.ReviewQueryDTO
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

	if query.StoreID == 0 {
		query.StoreID = middleware.GetStoreID(c)
	}

	result, err := h.reviewService.ListReviews(&query)
	if err != nil {
		middleware.Error(c, "获取评价列表失败: "+err.Error())
		return
	}

	middleware.PageSuccess(c, result.List, result.Total, result.Page, result.Size)
}

func (h *ReviewHandler) GetReview(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		middleware.Error(c, "无效的评价ID")
		return
	}

	review, err := h.reviewService.GetReview(uint(id))
	if err != nil {
		middleware.Error(c, "获取评价详情失败: "+err.Error())
		return
	}

	middleware.Success(c, review)
}

func (h *ReviewHandler) ReplyReview(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		middleware.Error(c, "无效的评价ID")
		return
	}

	var req dto.ReviewReplyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.Error(c, "参数错误: "+err.Error())
		return
	}

	err = h.reviewService.ReplyReview(uint(id), &req)
	if err != nil {
		middleware.Error(c, "回复评价失败: "+err.Error())
		return
	}

	middleware.Success(c, gin.H{"message": "回复成功"})
}

func (h *ReviewHandler) CreateWorkOrder(c *gin.Context) {
	var req dto.WorkOrderCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.Error(c, "参数错误: "+err.Error())
		return
	}

	workOrder, err := h.reviewService.CreateWorkOrder(&req)
	if err != nil {
		middleware.Error(c, "创建工单失败: "+err.Error())
		return
	}

	middleware.Success(c, workOrder)
}

func (h *ReviewHandler) ListWorkOrders(c *gin.Context) {
	var query dto.WorkOrderQueryDTO
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

	if query.StoreID == 0 {
		query.StoreID = middleware.GetStoreID(c)
	}

	result, err := h.reviewService.ListWorkOrders(&query)
	if err != nil {
		middleware.Error(c, "获取工单列表失败: "+err.Error())
		return
	}

	middleware.PageSuccess(c, result.List, result.Total, result.Page, result.Size)
}

func (h *ReviewHandler) GetWorkOrder(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		middleware.Error(c, "无效的工单ID")
		return
	}

	workOrder, err := h.reviewService.GetWorkOrder(uint(id))
	if err != nil {
		middleware.Error(c, "获取工单详情失败: "+err.Error())
		return
	}

	middleware.Success(c, workOrder)
}

func (h *ReviewHandler) HandleWorkOrder(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		middleware.Error(c, "无效的工单ID")
		return
	}

	var req dto.WorkOrderHandleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.Error(c, "参数错误: "+err.Error())
		return
	}

	handlerID := middleware.GetUserID(c)

	err = h.reviewService.HandleWorkOrder(uint(id), handlerID, &req)
	if err != nil {
		middleware.Error(c, "处理工单失败: "+err.Error())
		return
	}

	middleware.Success(c, gin.H{"message": "工单处理成功"})
}

func (h *ReviewHandler) ListAlerts(c *gin.Context) {
	var query dto.AlertQueryDTO
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

	if query.StoreID == 0 {
		query.StoreID = middleware.GetStoreID(c)
	}

	result, err := h.reviewService.ListAlerts(&query)
	if err != nil {
		middleware.Error(c, "获取告警列表失败: "+err.Error())
		return
	}

	middleware.PageSuccess(c, result.List, result.Total, result.Page, result.Size)
}

func (h *ReviewHandler) HandleAlert(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		middleware.Error(c, "无效的告警ID")
		return
	}

	var req dto.AlertHandleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.Error(c, "参数错误: "+err.Error())
		return
	}

	handlerID := middleware.GetUserID(c)

	err = h.reviewService.HandleAlert(uint(id), handlerID, &req)
	if err != nil {
		middleware.Error(c, "处理告警失败: "+err.Error())
		return
	}

	middleware.Success(c, gin.H{"message": "告警处理成功"})
}

func (h *ReviewHandler) CheckAlerts(c *gin.Context) {
	go func() {
		if _, err := h.reviewService.CheckRatingAlerts(); err != nil {
			fmt.Printf("[Review] Check alerts error: %v\n", err)
		}
	}()

	middleware.Success(c, gin.H{"message": "告警检查任务已启动"})
}
