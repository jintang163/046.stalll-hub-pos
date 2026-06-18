package handler

import (
	"net/http"
	"strconv"
	"stalll-hub-pos/backend/internal/dto"
	"stalll-hub-pos/backend/internal/middleware"
	"stalll-hub-pos/backend/internal/service"

	"github.com/gin-gonic/gin"
)

type CouponHandler struct {
	couponService *service.CouponService
}

func NewCouponHandler() *CouponHandler {
	return &CouponHandler{
		couponService: service.NewCouponService(),
	}
}

func (h *CouponHandler) CreateCoupon(c *gin.Context) {
	var req dto.CouponCreateDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.Error(c, "参数错误: "+err.Error())
		return
	}

	coupon, err := h.couponService.CreateCoupon(&req)
	if err != nil {
		middleware.Error(c, "创建优惠券失败: "+err.Error())
		return
	}

	middleware.Success(c, coupon)
}

func (h *CouponHandler) UpdateCoupon(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		middleware.Error(c, "无效的优惠券ID")
		return
	}

	var req dto.CouponUpdateDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.Error(c, "参数错误: "+err.Error())
		return
	}

	coupon, err := h.couponService.UpdateCoupon(uint(id), &req)
	if err != nil {
		middleware.Error(c, "更新优惠券失败: "+err.Error())
		return
	}

	middleware.Success(c, coupon)
}

func (h *CouponHandler) DeleteCoupon(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		middleware.Error(c, "无效的优惠券ID")
		return
	}

	err = h.couponService.DeleteCoupon(uint(id))
	if err != nil {
		middleware.Error(c, "删除优惠券失败: "+err.Error())
		return
	}

	middleware.Success(c, nil)
}

func (h *CouponHandler) GetCoupon(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		middleware.Error(c, "无效的优惠券ID")
		return
	}

	coupon, err := h.couponService.GetCoupon(uint(id))
	if err != nil {
		middleware.Error(c, "获取优惠券失败: "+err.Error())
		return
	}

	middleware.Success(c, coupon)
}

func (h *CouponHandler) ListCoupons(c *gin.Context) {
	var query dto.CouponQueryDTO
	if err := c.ShouldBindQuery(&query); err != nil {
		middleware.Error(c, "参数错误: "+err.Error())
		return
	}

	coupons, total, err := h.couponService.ListCoupons(&query)
	if err != nil {
		middleware.Error(c, "获取优惠券列表失败: "+err.Error())
		return
	}

	middleware.Success(c, gin.H{
		"list":  coupons,
		"total": total,
		"page":  query.Page,
		"size":  query.PageSize,
	})
}

func (h *CouponHandler) IssueCoupon(c *gin.Context) {
	var req dto.IssueCouponDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.Error(c, "参数错误: "+err.Error())
		return
	}

	result, err := h.couponService.IssueCoupon(&req)
	if err != nil {
		middleware.Error(c, "发放优惠券失败: "+err.Error())
		return
	}

	middleware.Success(c, result)
}

func (h *CouponHandler) ListMemberCoupons(c *gin.Context) {
	var query dto.MemberCouponQueryDTO
	if err := c.ShouldBindQuery(&query); err != nil {
		middleware.Error(c, "参数错误: "+err.Error())
		return
	}

	coupons, total, err := h.couponService.ListMemberCoupons(&query)
	if err != nil {
		middleware.Error(c, "获取会员优惠券失败: "+err.Error())
		return
	}

	middleware.Success(c, gin.H{
		"list":  coupons,
		"total": total,
		"page":  query.Page,
		"size":  query.PageSize,
	})
}

func (h *CouponHandler) GetMemberCoupon(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		middleware.Error(c, "无效的会员优惠券ID")
		return
	}

	coupon, err := h.couponService.GetMemberCoupon(uint(id))
	if err != nil {
		middleware.Error(c, "获取会员优惠券失败: "+err.Error())
		return
	}

	middleware.Success(c, coupon)
}

func (h *CouponHandler) VerifyCoupon(c *gin.Context) {
	var req dto.VerifyCouponDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.Error(c, "参数错误: "+err.Error())
		return
	}

	if req.StoreID == 0 {
		req.StoreID = middleware.GetStoreID(c)
	}

	result, err := h.couponService.VerifyCoupon(&req)
	if err != nil {
		middleware.Error(c, "核销优惠券失败: "+err.Error())
		return
	}

	middleware.Success(c, result)
}
