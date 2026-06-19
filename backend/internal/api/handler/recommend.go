package handler

import (
	"net/http"
	"strconv"
	"stalll-hub-pos/backend/internal/dto"
	"stalll-hub-pos/backend/internal/middleware"
	"stalll-hub-pos/backend/internal/service"

	"github.com/gin-gonic/gin"
)

type RecommendHandler struct {
	recommendService *service.RecommendService
}

func NewRecommendHandler() *RecommendHandler {
	return &RecommendHandler{
		recommendService: service.NewRecommendService(),
	}
}

func (h *RecommendHandler) GetConfig(c *gin.Context) {
	storeID := middleware.GetStoreID(c)
	if storeID == 0 {
		idStr := c.Query("store_id")
		id, err := strconv.ParseUint(idStr, 10, 32)
		if err != nil {
			middleware.Error(c, "缺少门店ID")
			return
		}
		storeID = uint(id)
	}

	cfg, err := h.recommendService.GetOrCreateConfig(storeID)
	if err != nil {
		middleware.Error(c, "获取推荐配置失败: "+err.Error())
		return
	}
	middleware.Success(c, cfg)
}

func (h *RecommendHandler) UpdateConfig(c *gin.Context) {
	storeID := middleware.GetStoreID(c)
	if storeID == 0 {
		middleware.Error(c, http.StatusUnauthorized, "未授权")
		return
	}

	var req dto.UpdateRecommendConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.Error(c, "参数错误: "+err.Error())
		return
	}

	cfg, err := h.recommendService.UpdateConfig(storeID, &req)
	if err != nil {
		middleware.Error(c, "更新推荐配置失败: "+err.Error())
		return
	}
	middleware.Success(c, cfg)
}

func (h *RecommendHandler) TriggerRefresh(c *gin.Context) {
	storeID := middleware.GetStoreID(c)
	if storeID == 0 {
		var req dto.TriggerRefreshRequest
		if err := c.ShouldBindJSON(&req); err == nil && req.StoreID > 0 {
			storeID = req.StoreID
		}
	}
	if storeID == 0 {
		middleware.Error(c, "缺少门店ID")
		return
	}

	go func() {
		_ = h.recommendService.TriggerRefresh(storeID)
	}()

	middleware.Success(c, gin.H{
		"store_id": storeID,
		"message":  "推荐刷新任务已启动，请稍后查看结果",
	})
}

func (h *RecommendHandler) GetRefreshStatus(c *gin.Context) {
	storeID := middleware.GetStoreID(c)
	if storeID == 0 {
		idStr := c.Query("store_id")
		id, err := strconv.ParseUint(idStr, 10, 32)
		if err != nil {
			middleware.Error(c, "缺少门店ID")
			return
		}
		storeID = uint(id)
	}

	status, err := h.recommendService.GetRefreshStatus(storeID)
	if err != nil {
		middleware.Error(c, "获取刷新状态失败: "+err.Error())
		return
	}
	middleware.Success(c, status)
}

func (h *RecommendHandler) GetCartRecommendations(c *gin.Context) {
	var req dto.GetRecommendRequest
	if err := c.ShouldBind(&req); err != nil {
		middleware.Error(c, "参数错误: "+err.Error())
		return
	}

	storeID := req.StoreID
	if storeID == 0 {
		idStr := c.Query("store_id")
		id, err := strconv.ParseUint(idStr, 10, 32)
		if err == nil {
			storeID = uint(id)
		}
	}
	if storeID == 0 {
		middleware.Error(c, "缺少门店ID")
		return
	}

	list, err := h.recommendService.GetCartRecommendations(storeID, req.ProductIDs, req.Count)
	if err != nil {
		middleware.Error(c, "获取推荐失败: "+err.Error())
		return
	}
	middleware.Success(c, list)
}
