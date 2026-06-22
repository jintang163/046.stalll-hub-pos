package handler

import (
	"net/http"
	"strconv"
	"stalll-hub-pos/backend/internal/dto"
	"stalll-hub-pos/backend/internal/middleware"
	"stalll-hub-pos/backend/internal/service"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
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

func (h *RecommendHandler) GetConfigMeta(c *gin.Context) {
	meta := h.recommendService.GetConfigMeta()
	middleware.Success(c, meta)
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

	memberID := req.MemberID
	userID := req.UserID

	if memberID == 0 && userID == 0 {
		claimsVal, exists := c.Get("jwt_claims")
		if exists {
			if claims, ok := claimsVal.(jwt.MapClaims); ok {
				if mid, ok := claims["member_id"].(float64); ok && mid > 0 {
					memberID = uint(mid)
				}
				if uid, ok := claims["user_id"].(float64); ok && uid > 0 {
					userID = uint(uid)
				}
			}
		}
		if memberID == 0 && userID == 0 {
			authHeader := c.GetHeader("X-Member-ID")
			if authHeader != "" {
				if id, err := strconv.ParseUint(authHeader, 10, 32); err == nil {
					memberID = uint(id)
				}
			}
			authHeader = c.GetHeader("X-User-ID")
			if authHeader != "" && userID == 0 {
				if id, err := strconv.ParseUint(authHeader, 10, 32); err == nil {
					userID = uint(id)
				}
			}
		}
	}

	list, err := h.recommendService.GetCartRecommendations(storeID, req.ProductIDs, memberID, userID, req.Count)
	if err != nil {
		middleware.Error(c, "获取推荐失败: "+err.Error())
		return
	}
	middleware.Success(c, list)
}

func (h *RecommendHandler) GetScanOrderRecommendations(c *gin.Context) {
	storeIDStr := c.Query("store_id")
	tableNo := c.Query("table_no")
	countStr := c.Query("count")

	if storeIDStr == "" {
		middleware.Error(c, "缺少门店ID")
		return
	}
	if tableNo == "" {
		middleware.Error(c, "缺少桌号")
		return
	}

	storeID64, err := strconv.ParseUint(storeIDStr, 10, 32)
	if err != nil {
		middleware.Error(c, "门店ID格式错误")
		return
	}
	storeID := uint(storeID64)

	count := 4
	if countStr != "" {
		if cnt, err := strconv.Atoi(countStr); err == nil && cnt > 0 && cnt <= 10 {
			count = cnt
		}
	}

	list, err := h.recommendService.GetScanOrderRecommendations(storeID, tableNo, count)
	if err != nil {
		middleware.Error(c, "获取推荐失败: "+err.Error())
		return
	}
	middleware.Success(c, gin.H{
		"items":     list,
		"table_no":  tableNo,
		"store_id":  storeID,
		"count":     len(list),
		"timestamp": time.Now().Unix(),
	})
}
