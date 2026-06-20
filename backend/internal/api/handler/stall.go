package handler

import (
	"net/http"
	"strconv"
	"stalll-hub-pos/backend/internal/dto"
	"stalll-hub-pos/backend/internal/middleware"
	"stalll-hub-pos/backend/internal/service"

	"github.com/gin-gonic/gin"
)

type StallHandler struct {
	stallService *service.StallService
}

func NewStallHandler() *StallHandler {
	return &StallHandler{
		stallService: service.NewStallService(),
	}
}

func (h *StallHandler) CreateStall(c *gin.Context) {
	var req dto.StallCreateDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.Error(c, "参数错误: "+err.Error())
		return
	}

	stall, err := h.stallService.CreateStall(&req)
	if err != nil {
		middleware.Error(c, "创建摊位失败: "+err.Error())
		return
	}

	middleware.Success(c, stall)
}

func (h *StallHandler) GetStall(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		middleware.Error(c, "无效的摊位ID")
		return
	}

	stall, err := h.stallService.GetStall(uint(id))
	if err != nil {
		middleware.Error(c, "获取摊位失败: "+err.Error())
		return
	}

	middleware.Success(c, stall)
}

func (h *StallHandler) UpdateStall(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		middleware.Error(c, "无效的摊位ID")
		return
	}

	var req dto.StallUpdateDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.Error(c, "参数错误: "+err.Error())
		return
	}

	stall, err := h.stallService.UpdateStall(uint(id), &req)
	if err != nil {
		middleware.Error(c, "更新摊位失败: "+err.Error())
		return
	}

	middleware.Success(c, stall)
}

func (h *StallHandler) DeleteStall(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		middleware.Error(c, "无效的摊位ID")
		return
	}

	err = h.stallService.DeleteStall(uint(id))
	if err != nil {
		middleware.Error(c, "删除摊位失败: "+err.Error())
		return
	}

	middleware.Success(c, nil)
}

func (h *StallHandler) ListStalls(c *gin.Context) {
	var query dto.StallQueryDTO
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

	stalls, total, err := h.stallService.ListStalls(&query)
	if err != nil {
		middleware.Error(c, "获取摊位列表失败: "+err.Error())
		return
	}

	middleware.PageSuccess(c, stalls, total, query.Page, query.PageSize)
}

func (h *StallHandler) GetAllStalls(c *gin.Context) {
	storeIDStr := c.Query("store_id")
	storeID := uint(0)
	if storeIDStr != "" {
		if id, err := strconv.ParseUint(storeIDStr, 10, 32); err == nil {
			storeID = uint(id)
		}
	}

	stalls, err := h.stallService.GetAllStalls(storeID)
	if err != nil {
		middleware.Error(c, "获取摊位列表失败: "+err.Error())
		return
	}

	middleware.Success(c, stalls)
}

func (h *StallHandler) RegisterDevice(c *gin.Context) {
	var req dto.StallDeviceRegisterDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.Error(c, "参数错误: "+err.Error())
		return
	}

	device, err := h.stallService.RegisterDevice(&req)
	if err != nil {
		middleware.Error(c, "注册设备失败: "+err.Error())
		return
	}

	middleware.Success(c, device)
}

func (h *StallHandler) Heartbeat(c *gin.Context) {
	var req dto.StallDeviceHeartbeatDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.Error(c, "参数错误: "+err.Error())
		return
	}

	err := h.stallService.Heartbeat(req.DeviceID, req.AppVersion)
	if err != nil {
		middleware.Error(c, "心跳失败: "+err.Error())
		return
	}

	middleware.Success(c, gin.H{"status": "ok"})
}

func (h *StallHandler) GetDevice(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		middleware.Error(c, "无效的设备ID")
		return
	}

	device, err := h.stallService.GetDevice(uint(id))
	if err != nil {
		middleware.Error(c, "获取设备失败: "+err.Error())
		return
	}

	middleware.Success(c, device)
}

func (h *StallHandler) ListDevices(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	storeID, _ := strconv.ParseUint(c.Query("store_id"), 10, 32)
	stallID, _ := strconv.ParseUint(c.Query("stall_id"), 10, 32)

	devices, total, err := h.stallService.ListDevices(uint(storeID), uint(stallID), page, pageSize)
	if err != nil {
		middleware.Error(c, "获取设备列表失败: "+err.Error())
		return
	}

	middleware.PageSuccess(c, devices, total, page, pageSize)
}

func (h *StallHandler) DeleteDevice(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		middleware.Error(c, "无效的设备ID")
		return
	}

	err = h.stallService.DeleteDevice(uint(id))
	if err != nil {
		middleware.Error(c, "删除设备失败: "+err.Error())
		return
	}

	middleware.Success(c, nil)
}

func (h *StallHandler) CreateStallUser(c *gin.Context) {
	var req dto.StallUserCreateDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.Error(c, "参数错误: "+err.Error())
		return
	}

	user, err := h.stallService.CreateStallUser(&req)
	if err != nil {
		middleware.Error(c, "创建摊位用户失败: "+err.Error())
		return
	}

	middleware.Success(c, user)
}

func (h *StallHandler) UpdateStallUser(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		middleware.Error(c, "无效的用户ID")
		return
	}

	var req dto.StallUserUpdateDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.Error(c, "参数错误: "+err.Error())
		return
	}

	user, err := h.stallService.UpdateStallUser(uint(id), &req)
	if err != nil {
		middleware.Error(c, "更新摊位用户失败: "+err.Error())
		return
	}

	middleware.Success(c, user)
}

func (h *StallHandler) ListStallUsers(c *gin.Context) {
	var query dto.StallUserQueryDTO
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

	users, total, err := h.stallService.ListStallUsers(&query)
	if err != nil {
		middleware.Error(c, "获取摊位用户列表失败: "+err.Error())
		return
	}

	middleware.PageSuccess(c, users, total, query.Page, query.PageSize)
}

func (h *StallHandler) DeleteStallUser(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		middleware.Error(c, "无效的用户ID")
		return
	}

	err = h.stallService.DeleteStallUser(uint(id))
	if err != nil {
		middleware.Error(c, "删除摊位用户失败: "+err.Error())
		return
	}

	middleware.Success(c, nil)
}

func (h *StallHandler) StallLogin(c *gin.Context) {
	var req dto.StallLoginDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.Error(c, "参数错误: "+err.Error())
		return
	}

	result, err := h.stallService.StallLogin(&req)
	if err != nil {
		middleware.Error(c, "登录失败: "+err.Error())
		return
	}

	middleware.Success(c, result)
}

func (h *StallHandler) CreateSettlement(c *gin.Context) {
	var req dto.StallSettlementCreateDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.Error(c, "参数错误: "+err.Error())
		return
	}

	settlement, err := h.stallService.CreateSettlement(&req)
	if err != nil {
		middleware.Error(c, "创建结算单失败: "+err.Error())
		return
	}

	middleware.Success(c, settlement)
}

func (h *StallHandler) ListSettlements(c *gin.Context) {
	var query dto.StallSettlementQueryDTO
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

	settlements, total, err := h.stallService.ListSettlements(&query)
	if err != nil {
		middleware.Error(c, "获取结算列表失败: "+err.Error())
		return
	}

	middleware.PageSuccess(c, settlements, total, query.Page, query.PageSize)
}

func (h *StallHandler) GetDailyReport(c *gin.Context) {
	var query dto.StallDailyReportQueryDTO
	if err := c.ShouldBindQuery(&query); err != nil {
		middleware.Error(c, "参数错误: "+err.Error())
		return
	}

	reports, err := h.stallService.GetDailyReport(&query)
	if err != nil {
		middleware.Error(c, "获取摊位日报失败: "+err.Error())
		return
	}

	middleware.Success(c, reports)
}

func (h *StallHandler) GenerateDailyReport(c *gin.Context) {
	storeID, _ := strconv.ParseUint(c.Query("store_id"), 10, 32)
	stallID, _ := strconv.ParseUint(c.Query("stall_id"), 10, 32)
	reportDate := c.Query("report_date")

	if stallID == 0 {
		middleware.Error(c, "摊位ID不能为空")
		return
	}

	report, err := h.stallService.GenerateDailyReport(uint(storeID), uint(stallID), reportDate)
	if err != nil {
		middleware.Error(c, "生成日报失败: "+err.Error())
		return
	}

	middleware.Success(c, report)
}
