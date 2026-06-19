package handler

import (
	"net/http"
	"strconv"
	"stalll-hub-pos/backend/internal/dto"
	"stalll-hub-pos/backend/internal/middleware"
	"stalll-hub-pos/backend/internal/service"

	"github.com/gin-gonic/gin"
)

type RechargeActivityHandler struct {
	rechargeService *service.RechargeActivityService
}

func NewRechargeActivityHandler() *RechargeActivityHandler {
	return &RechargeActivityHandler{
		rechargeService: service.NewRechargeActivityService(),
	}
}

func (h *RechargeActivityHandler) CreateActivity(c *gin.Context) {
	var req dto.RechargeActivityCreateDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.Error(c, "参数错误: "+err.Error())
		return
	}

	if req.StoreID == 0 {
		req.StoreID = middleware.GetStoreID(c)
	}

	activity, err := h.rechargeService.CreateActivity(&req)
	if err != nil {
		middleware.Error(c, "创建充值活动失败: "+err.Error())
		return
	}

	middleware.Success(c, activity)
}

func (h *RechargeActivityHandler) GetActivity(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		middleware.Error(c, "无效的活动ID")
		return
	}

	activity, err := h.rechargeService.GetActivity(uint(id))
	if err != nil {
		middleware.Error(c, "获取充值活动失败: "+err.Error())
		return
	}

	middleware.Success(c, activity)
}

func (h *RechargeActivityHandler) UpdateActivity(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		middleware.Error(c, "无效的活动ID")
		return
	}

	var req dto.RechargeActivityUpdateDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.Error(c, "参数错误: "+err.Error())
		return
	}

	activity, err := h.rechargeService.UpdateActivity(uint(id), &req)
	if err != nil {
		middleware.Error(c, "更新充值活动失败: "+err.Error())
		return
	}

	middleware.Success(c, activity)
}

func (h *RechargeActivityHandler) DeleteActivity(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		middleware.Error(c, "无效的活动ID")
		return
	}

	err = h.rechargeService.DeleteActivity(uint(id))
	if err != nil {
		middleware.Error(c, "删除充值活动失败: "+err.Error())
		return
	}

	middleware.Success(c, nil)
}

func (h *RechargeActivityHandler) ListActivities(c *gin.Context) {
	var query dto.RechargeActivityQueryDTO
	if err := c.ShouldBindQuery(&query); err != nil {
		middleware.Error(c, "参数错误: "+err.Error())
		return
	}

	if query.StoreID == 0 {
		query.StoreID = middleware.GetStoreID(c)
	}

	result, err := h.rechargeService.ListActivities(&query)
	if err != nil {
		middleware.Error(c, "获取充值活动列表失败: "+err.Error())
		return
	}

	middleware.Success(c, result)
}

func (h *RechargeActivityHandler) ProcessRecharge(c *gin.Context) {
	var req dto.MemberRechargeDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.Error(c, "参数错误: "+err.Error())
		return
	}

	result, err := h.rechargeService.ProcessRecharge(&req)
	if err != nil {
		middleware.Error(c, "充值处理失败: "+err.Error())
		return
	}

	middleware.Success(c, result)
}

func (h *RechargeActivityHandler) ListRecharges(c *gin.Context) {
	memberIDStr := c.Query("member_id")
	storeIDStr := c.Query("store_id")
	pageStr := c.DefaultQuery("page", "1")
	pageSizeStr := c.DefaultQuery("page_size", "10")

	memberID, _ := strconv.ParseUint(memberIDStr, 10, 32)
	storeID, _ := strconv.ParseUint(storeIDStr, 10, 32)
	page, _ := strconv.Atoi(pageStr)
	pageSize, _ := strconv.Atoi(pageSizeStr)

	if storeID == 0 {
		sid := middleware.GetStoreID(c)
		storeID = uint(sid)
	}

	result, err := h.rechargeService.ListRecharges(uint(memberID), uint(storeID), page, pageSize)
	if err != nil {
		middleware.Error(c, "获取充值记录失败: "+err.Error())
		return
	}

	middleware.Success(c, result)
}

func init() {
	_ = http.StatusOK
}
