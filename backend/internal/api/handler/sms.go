package handler

import (
	"strconv"
	"stalll-hub-pos/backend/internal/dto"
	"stalll-hub-pos/backend/internal/middleware"
	"stalll-hub-pos/backend/internal/service"

	"github.com/gin-gonic/gin"
)

type SmsHandler struct {
	smsService *service.SmsService
}

func NewSmsHandler() *SmsHandler {
	return &SmsHandler{
		smsService: service.NewSmsService(),
	}
}

func (h *SmsHandler) CreateTemplate(c *gin.Context) {
	var req dto.SmsTemplateCreateDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.Error(c, "参数错误: "+err.Error())
		return
	}

	creatorID := middleware.GetUserID(c)

	template, err := h.smsService.CreateTemplate(creatorID, &req)
	if err != nil {
		middleware.Error(c, "创建短信模板失败: "+err.Error())
		return
	}

	middleware.Success(c, template)
}

func (h *SmsHandler) UpdateTemplate(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		middleware.Error(c, "无效的模板ID")
		return
	}

	var req dto.SmsTemplateUpdateDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.Error(c, "参数错误: "+err.Error())
		return
	}

	template, err := h.smsService.UpdateTemplate(uint(id), &req)
	if err != nil {
		middleware.Error(c, "更新短信模板失败: "+err.Error())
		return
	}

	middleware.Success(c, template)
}

func (h *SmsHandler) DeleteTemplate(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		middleware.Error(c, "无效的模板ID")
		return
	}

	err = h.smsService.DeleteTemplate(uint(id))
	if err != nil {
		middleware.Error(c, "删除短信模板失败: "+err.Error())
		return
	}

	middleware.Success(c, nil)
}

func (h *SmsHandler) GetTemplate(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		middleware.Error(c, "无效的模板ID")
		return
	}

	template, err := h.smsService.GetTemplate(uint(id))
	if err != nil {
		middleware.Error(c, "获取模板详情失败: "+err.Error())
		return
	}

	middleware.Success(c, template)
}

func (h *SmsHandler) ListTemplates(c *gin.Context) {
	var query dto.SmsTemplateQueryDTO
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

	templates, total, err := h.smsService.ListTemplates(&query)
	if err != nil {
		middleware.Error(c, "获取模板列表失败: "+err.Error())
		return
	}

	middleware.PageSuccess(c, templates, total, query.Page, query.PageSize)
}

func (h *SmsHandler) ReviewTemplate(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		middleware.Error(c, "无效的模板ID")
		return
	}

	var req dto.SmsTemplateReviewDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.Error(c, "参数错误: "+err.Error())
		return
	}

	reviewerID := middleware.GetUserID(c)
	req.ReviewerID = reviewerID

	err = h.smsService.ReviewTemplate(uint(id), &req)
	if err != nil {
		middleware.Error(c, "审核模板失败: "+err.Error())
		return
	}

	middleware.Success(c, gin.H{"message": "审核成功"})
}

func (h *SmsHandler) ListActiveTemplates(c *gin.Context) {
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

	templateType := c.Query("template_type")

	templates, err := h.smsService.ListActiveTemplates(storeID, templateType)
	if err != nil {
		middleware.Error(c, "获取已激活模板列表失败: "+err.Error())
		return
	}

	middleware.Success(c, templates)
}

func (h *SmsHandler) CreateTask(c *gin.Context) {
	var req dto.SmsTaskCreateDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.Error(c, "参数错误: "+err.Error())
		return
	}

	creatorID := middleware.GetUserID(c)

	task, err := h.smsService.CreateTask(creatorID, &req)
	if err != nil {
		middleware.Error(c, "创建短信任务失败: "+err.Error())
		return
	}

	middleware.Success(c, task)
}

func (h *SmsHandler) UpdateTask(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		middleware.Error(c, "无效的任务ID")
		return
	}

	var req dto.SmsTaskCreateDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.Error(c, "参数错误: "+err.Error())
		return
	}

	task, err := h.smsService.UpdateTask(uint(id), &req)
	if err != nil {
		middleware.Error(c, "更新短信任务失败: "+err.Error())
		return
	}

	middleware.Success(c, task)
}

func (h *SmsHandler) DeleteTask(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		middleware.Error(c, "无效的任务ID")
		return
	}

	err = h.smsService.DeleteTask(uint(id))
	if err != nil {
		middleware.Error(c, "删除短信任务失败: "+err.Error())
		return
	}

	middleware.Success(c, nil)
}

func (h *SmsHandler) GetTask(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		middleware.Error(c, "无效的任务ID")
		return
	}

	task, err := h.smsService.GetTask(uint(id))
	if err != nil {
		middleware.Error(c, "获取任务详情失败: "+err.Error())
		return
	}

	middleware.Success(c, task)
}

func (h *SmsHandler) ListTasks(c *gin.Context) {
	var query dto.SmsTaskQueryDTO
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

	tasks, total, err := h.smsService.ListTasks(&query)
	if err != nil {
		middleware.Error(c, "获取任务列表失败: "+err.Error())
		return
	}

	middleware.PageSuccess(c, tasks, total, query.Page, query.PageSize)
}

func (h *SmsHandler) StartTask(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		middleware.Error(c, "无效的任务ID")
		return
	}

	err = h.smsService.StartTask(uint(id))
	if err != nil {
		middleware.Error(c, "启动任务失败: "+err.Error())
		return
	}

	middleware.Success(c, gin.H{"message": "任务已启动"})
}

func (h *SmsHandler) PauseTask(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		middleware.Error(c, "无效的任务ID")
		return
	}

	err = h.smsService.PauseTask(uint(id))
	if err != nil {
		middleware.Error(c, "暂停任务失败: "+err.Error())
		return
	}

	middleware.Success(c, gin.H{"message": "任务已暂停"})
}

func (h *SmsHandler) GetTaskStatistics(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		middleware.Error(c, "无效的任务ID")
		return
	}

	statistics, err := h.smsService.GetTaskStatistics(uint(id))
	if err != nil {
		middleware.Error(c, "获取任务统计失败: "+err.Error())
		return
	}

	middleware.Success(c, statistics)
}

func (h *SmsHandler) CalculateTargetCount(c *gin.Context) {
	var req dto.SmsTargetCountDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.Error(c, "参数错误: "+err.Error())
		return
	}

	if req.StoreID == 0 {
		req.StoreID = middleware.GetStoreID(c)
	}

	result, err := h.smsService.CalculateTargetCount(&req)
	if err != nil {
		middleware.Error(c, "计算目标用户数量失败: "+err.Error())
		return
	}

	middleware.Success(c, result)
}

func (h *SmsHandler) ListRecords(c *gin.Context) {
	var query dto.SmsRecordQueryDTO
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

	records, total, err := h.smsService.ListRecords(&query)
	if err != nil {
		middleware.Error(c, "获取发送记录列表失败: "+err.Error())
		return
	}

	middleware.PageSuccess(c, records, total, query.Page, query.PageSize)
}

func (h *SmsHandler) GetRecord(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		middleware.Error(c, "无效的记录ID")
		return
	}

	record, err := h.smsService.GetRecord(uint(id))
	if err != nil {
		middleware.Error(c, "获取发送记录详情失败: "+err.Error())
		return
	}

	middleware.Success(c, record)
}

func (h *SmsHandler) SendTestSms(c *gin.Context) {
	var req dto.SmsSendTestDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.Error(c, "参数错误: "+err.Error())
		return
	}

	err := h.smsService.SendTestSms(&req)
	if err != nil {
		middleware.Error(c, "发送测试短信失败: "+err.Error())
		return
	}

	middleware.Success(c, gin.H{"message": "测试短信发送成功"})
}
