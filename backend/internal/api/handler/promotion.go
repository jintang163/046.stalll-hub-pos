package handler

import (
	"net/http"
	"strconv"
	"stalll-hub-pos/backend/internal/dto"
	"stalll-hub-pos/backend/internal/middleware"
	"stalll-hub-pos/backend/internal/service"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
)

type PromotionHandler struct {
	promotionService *service.PromotionEngineService
}

func NewPromotionHandler() *PromotionHandler {
	return &PromotionHandler{
		promotionService: service.NewPromotionEngineService(),
	}
}

func (h *PromotionHandler) CreatePromotion(c *gin.Context) {
	var req dto.PromotionCreateDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.Error(c, "参数错误: "+err.Error())
		return
	}

	storeID := middleware.GetStoreID(c)
	promotion, err := h.promotionService.CreatePromotion(storeID, &req)
	if err != nil {
		middleware.Error(c, "创建活动失败: "+err.Error())
		return
	}

	middleware.Success(c, promotion)
}

func (h *PromotionHandler) UpdatePromotion(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		middleware.Error(c, "无效的活动ID")
		return
	}

	var req dto.PromotionUpdateDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.Error(c, "参数错误: "+err.Error())
		return
	}

	promotion, err := h.promotionService.UpdatePromotion(uint(id), &req)
	if err != nil {
		middleware.Error(c, "更新活动失败: "+err.Error())
		return
	}

	middleware.Success(c, promotion)
}

func (h *PromotionHandler) DeletePromotion(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		middleware.Error(c, "无效的活动ID")
		return
	}

	err = h.promotionService.DeletePromotion(uint(id))
	if err != nil {
		middleware.Error(c, "删除活动失败: "+err.Error())
		return
	}

	middleware.Success(c, nil)
}

func (h *PromotionHandler) GetPromotion(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		middleware.Error(c, "无效的活动ID")
		return
	}

	promotion, err := h.promotionService.GetPromotion(uint(id))
	if err != nil {
		middleware.Error(c, "获取活动失败: "+err.Error())
		return
	}

	middleware.Success(c, promotion)
}

func (h *PromotionHandler) ListPromotions(c *gin.Context) {
	var query dto.PromotionQueryDTO
	if err := c.ShouldBindQuery(&query); err != nil {
		middleware.Error(c, "参数错误: "+err.Error())
		return
	}

	result, err := h.promotionService.ListPromotions(&query)
	if err != nil {
		middleware.Error(c, "获取活动列表失败: "+err.Error())
		return
	}

	middleware.Success(c, result)
}

func (h *PromotionHandler) CalculateBestCombination(c *gin.Context) {
	var req dto.PromotionCalcRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.Error(c, "参数错误: "+err.Error())
		return
	}

	storeID := req.StoreID
	if storeID == 0 {
		storeID = middleware.GetStoreID(c)
	}

	memberID := middleware.GetMemberID(c)
	if memberID == 0 {
		memberID = req.MemberID
	}
	memberCouponID := req.MemberCouponID

	result, err := h.promotionService.CalculateBestCombination(storeID, req.Amount, req.ProductIDs, memberCouponID, memberID)
	if err != nil {
		middleware.Error(c, "计算优惠失败: "+err.Error())
		return
	}

	middleware.Success(c, result)
}

func (h *PromotionHandler) GetActivePromotions(c *gin.Context) {
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

	promotions, err := h.promotionService.GetActivePromotions(storeID)
	if err != nil {
		middleware.Error(c, "获取活动列表失败: "+err.Error())
		return
	}

	middleware.Success(c, promotions)
}

func (h *PromotionHandler) ClaimCoupon(c *gin.Context) {
	var req dto.MemberClaimCouponDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.Error(c, "参数错误: "+err.Error())
		return
	}

	memberID := middleware.GetMemberID(c)
	if memberID == 0 {
		middleware.ErrorWithCode(c, http.StatusUnauthorized, "请先登录")
		return
	}

	couponService := service.NewCouponService()
	result, err := couponService.ClaimCoupon(memberID, req.CouponID)
	if err != nil {
		middleware.Error(c, "领取优惠券失败: "+err.Error())
		return
	}

	middleware.Success(c, result)
}

func (h *PromotionHandler) GetAvailableCoupons(c *gin.Context) {
	memberID := middleware.GetMemberID(c)
	if memberID == 0 {
		middleware.Success(c, []interface{}{})
		return
	}

	storeID := uint(0)
	storeIDStr := c.Query("store_id")
	if storeIDStr != "" {
		id, err := strconv.ParseUint(storeIDStr, 10, 32)
		if err == nil {
			storeID = uint(id)
		}
	}

	amountStr := c.Query("amount")
	amount := decimal.Zero
	if amountStr != "" {
		if a, err := decimal.NewFromString(amountStr); err == nil {
			amount = a
		}
	}

	productIDsStr := c.Query("product_ids")
	var productIDs []uint
	if productIDsStr != "" {
		parts := strings.Split(productIDsStr, ",")
		for _, part := range parts {
			if id, err := strconv.ParseUint(part, 10, 32); err == nil {
				productIDs = append(productIDs, uint(id))
			}
		}
	}

	couponService := service.NewCouponService()
	coupons, err := couponService.GetAvailableCoupons(memberID, storeID, amount, productIDs)
	if err != nil {
		middleware.Error(c, "获取可用优惠券失败: "+err.Error())
		return
	}

	middleware.Success(c, coupons)
}

func (h *PromotionHandler) GetClaimableCoupons(c *gin.Context) {
	storeID := uint(0)
	storeIDStr := c.Query("store_id")
	if storeIDStr != "" {
		id, err := strconv.ParseUint(storeIDStr, 10, 32)
		if err == nil {
			storeID = uint(id)
		}
	}
	if storeID == 0 {
		storeID = middleware.GetStoreID(c)
	}

	memberID := middleware.GetMemberID(c)

	couponService := service.NewCouponService()
	coupons, err := couponService.GetClaimableCoupons(storeID, memberID)
	if err != nil {
		middleware.Error(c, "获取可领优惠券失败: "+err.Error())
		return
	}

	middleware.Success(c, coupons)
}

func (h *PromotionHandler) GetMyCoupons(c *gin.Context) {
	memberID := middleware.GetMemberID(c)
	if memberID == 0 {
		middleware.Success(c, gin.H{"list": []interface{}{}, "total": 0})
		return
	}

	status := 0
	statusStr := c.Query("status")
	if statusStr != "" {
		if s, err := strconv.Atoi(statusStr); err == nil {
			status = s
		}
	}

	page := 1
	pageSize := 20
	pageStr := c.Query("page")
	if pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil {
			page = p
		}
	}
	pageSizeStr := c.Query("page_size")
	if pageSizeStr != "" {
		if p, err := strconv.Atoi(pageSizeStr); err == nil {
			pageSize = p
		}
	}

	query := &dto.MemberCouponQueryDTO{
		MemberID: memberID,
		Status:   status,
	}
	query.Page = page
	query.PageSize = pageSize

	couponService := service.NewCouponService()
	result, err := couponService.GetMemberCoupons(query)
	if err != nil {
		middleware.Error(c, "获取我的优惠券失败: "+err.Error())
		return
	}

	middleware.Success(c, result)
}
