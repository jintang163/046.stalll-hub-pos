package handler

import (
	"stalll-hub-pos/backend/internal/dto"
	"stalll-hub-pos/backend/internal/middleware"
	"stalll-hub-pos/backend/internal/service"

	"github.com/gin-gonic/gin"
)

type VoiceOrderHandler struct {
	voiceService *service.VoiceOrderService
}

func NewVoiceOrderHandler() *VoiceOrderHandler {
	return &VoiceOrderHandler{
		voiceService: service.NewVoiceOrderService(),
	}
}

func (h *VoiceOrderHandler) ParseVoiceText(c *gin.Context) {
	var req dto.VoiceParseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.Error(c, "参数错误: "+err.Error())
		return
	}

	result, err := h.voiceService.ParseVoiceText(req.StoreID, req.Text)
	if err != nil {
		middleware.Error(c, "语音解析失败: "+err.Error())
		return
	}

	middleware.Success(c, result)
}
