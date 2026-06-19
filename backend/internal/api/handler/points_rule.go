package handler

import (
	"strconv"
	"stalll-hub-pos/backend/internal/dto"
	"stalll-hub-pos/backend/internal/middleware"
	"stalll-hub-pos/backend/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
)

type PointsRuleHandler struct {
	pointsEngineService *service.PointsEngineService
}

func NewPointsRuleHandler() *PointsRuleHandler {
	return &PointsRuleHandler{
		pointsEngineService: service.NewPointsEngineService(),
	}
}

func (h *PointsRuleHandler) CreateRule(c *gin.Context) {
	var req dto.PointsRuleCreateDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.Error(c, "参数错误: "+err.Error())
		return
	}

	if req.StoreID == 0 {
		req.StoreID = middleware.GetStoreID(c)
	}

	rule, err := h.pointsEngineService.CreateRule(&req)
	if err != nil {
		middleware.Error(c, "创建积分规则失败: "+err.Error())
		return
	}

	middleware.Success(c, rule)
}

func (h *PointsRuleHandler) GetRule(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		middleware.Error(c, "无效的规则ID")
		return
	}

	rule, err := h.pointsEngineService.GetRule(uint(id))
	if err != nil {
		middleware.Error(c, "获取积分规则失败: "+err.Error())
		return
	}

	middleware.Success(c, rule)
}

func (h *PointsRuleHandler) UpdateRule(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		middleware.Error(c, "无效的规则ID")
		return
	}

	var req dto.PointsRuleUpdateDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.Error(c, "参数错误: "+err.Error())
		return
	}

	rule, err := h.pointsEngineService.UpdateRule(uint(id), &req)
	if err != nil {
		middleware.Error(c, "更新积分规则失败: "+err.Error())
		return
	}

	middleware.Success(c, rule)
}

func (h *PointsRuleHandler) DeleteRule(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		middleware.Error(c, "无效的规则ID")
		return
	}

	err = h.pointsEngineService.DeleteRule(uint(id))
	if err != nil {
		middleware.Error(c, "删除积分规则失败: "+err.Error())
		return
	}

	middleware.Success(c, nil)
}

func (h *PointsRuleHandler) ListRules(c *gin.Context) {
	var query dto.PointsRuleQueryDTO
	if err := c.ShouldBindQuery(&query); err != nil {
		middleware.Error(c, "参数错误: "+err.Error())
		return
	}

	if query.StoreID == 0 {
		query.StoreID = middleware.GetStoreID(c)
	}

	result, err := h.pointsEngineService.ListRules(&query)
	if err != nil {
		middleware.Error(c, "获取积分规则列表失败: "+err.Error())
		return
	}

	middleware.Success(c, result)
}

func (h *PointsRuleHandler) CalculateEarnedPoints(c *gin.Context) {
	var req dto.PointsEarnRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.Error(c, "参数错误: "+err.Error())
		return
	}

	storeID := middleware.GetStoreID(c)
	pointsRateStr := c.DefaultQuery("points_rate", "1")
	pointsRate, _ := decimal.NewFromString(pointsRateStr)
	if pointsRate.IsZero() {
		pointsRate = decimal.NewFromInt(1)
	}

	points := h.pointsEngineService.CalculateEarnedPoints(storeID, req.ConsumeAmount, pointsRate)

	middleware.Success(c, gin.H{"points_earned": points})
}

func (h *PointsRuleHandler) CalculateRedemptionDiscount(c *gin.Context) {
	var req dto.PointsRedeemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.Error(c, "参数错误: "+err.Error())
		return
	}

	storeID := middleware.GetStoreID(c)
	discount, usablePoints, err := h.pointsEngineService.CalculateRedemptionDiscount(storeID, req.Points)
	if err != nil {
		middleware.Error(c, "计算积分抵扣失败: "+err.Error())
		return
	}

	middleware.Success(c, gin.H{
		"discount_amount": discount,
		"usable_points":   usablePoints,
	})
}
