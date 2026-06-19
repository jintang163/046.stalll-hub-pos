package service

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/shopspring/decimal"
	"stalll-hub-pos/backend/internal/dto"
	"stalll-hub-pos/backend/internal/model"
	"stalll-hub-pos/backend/internal/repository"
)

type CouponService struct {
	couponRepo *repository.CouponRepository
	memberRepo *repository.MemberRepository
}

func NewCouponService() *CouponService {
	return &CouponService{
		couponRepo: repository.NewCouponRepository(nil),
		memberRepo: repository.NewMemberRepository(nil),
	}
}

func (s *CouponService) CreateCoupon(storeID uint, req *dto.CouponCreateDTO) (*model.Coupon, error) {
	applicableIDs := ""
	if len(req.ApplicableIDs) > 0 {
		ids := make([]string, len(req.ApplicableIDs))
		for i, id := range req.ApplicableIDs {
			ids[i] = fmt.Sprintf("%d", id)
		}
		applicableIDs = strings.Join(ids, ",")
	}

	excludeProducts := ""
	if len(req.ExcludeProducts) > 0 {
		ids := make([]string, len(req.ExcludeProducts))
		for i, id := range req.ExcludeProducts {
			ids[i] = fmt.Sprintf("%d", id)
		}
		excludeProducts = strings.Join(ids, ",")
	}

	coupon := &model.Coupon{
		StoreID:         storeID,
		RuleKey:         req.RuleKey,
		Name:            req.Name,
		Type:            req.Type,
		Value:           req.Value,
		MinConsume:      req.MinAmount,
		DiscountRate:    req.Value,
		MaxDiscount:     req.MaxDiscount,
		TotalCount:      req.TotalCount,
		UsedCount:       0,
		PerUserLimit:    req.PerUserLimit,
		ValidType:       req.ValidityType,
		ValidDays:       req.ValidityDays,
		StartTime:       req.StartTime,
		EndTime:         req.EndTime,
		ApplyScope:      req.ApplicableType,
		ProductIDs:      applicableIDs,
		ExcludeProducts: excludeProducts,
		Status:          req.Status,
		Description:     req.Description,
	}

	if coupon.Status == 0 {
		coupon.Status = 1
	}

	if coupon.Type == "percentage" && coupon.MaxDiscount.IsZero() {
		coupon.MaxDiscount = decimal.NewFromInt(100)
	}

	err := s.couponRepo.Create(coupon)
	if err != nil {
		return nil, fmt.Errorf("create coupon failed: %w", err)
	}

	return s.couponRepo.GetByID(coupon.ID)
}

func (s *CouponService) GetCoupon(id uint) (*dto.CouponResponse, error) {
	coupon, err := s.couponRepo.GetByID(id)
	if err != nil {
		return nil, err
	}
	return s.convertToCouponResponse(coupon), nil
}

func (s *CouponService) UpdateCoupon(id uint, req *dto.CouponUpdateDTO) (*dto.CouponResponse, error) {
	coupon, err := s.couponRepo.GetByID(id)
	if err != nil {
		return nil, errors.New("coupon not found")
	}

	if req.RuleKey != "" {
		coupon.RuleKey = req.RuleKey
	}
	if req.Name != "" {
		coupon.Name = req.Name
	}
	if !req.Value.IsZero() {
		coupon.Value = req.Value
		if coupon.Type == "percentage" {
			coupon.DiscountRate = req.Value
		}
	}
	if !req.MinAmount.IsZero() {
		coupon.MinConsume = req.MinAmount
	}
	if !req.MaxDiscount.IsZero() {
		coupon.MaxDiscount = req.MaxDiscount
	}
	if req.TotalCount > 0 {
		coupon.TotalCount = req.TotalCount
	}
	if req.PerUserLimit > 0 {
		coupon.PerUserLimit = req.PerUserLimit
	}
	if req.ValidityType != "" {
		coupon.ValidType = req.ValidityType
	}
	if req.ValidityDays > 0 {
		coupon.ValidDays = req.ValidityDays
	}
	if req.StartTime != nil {
		coupon.StartTime = req.StartTime
	}
	if req.EndTime != nil {
		coupon.EndTime = req.EndTime
	}
	if req.ApplicableType != "" {
		coupon.ApplyScope = req.ApplicableType
	}
	if len(req.ApplicableIDs) > 0 {
		ids := make([]string, len(req.ApplicableIDs))
		for i, id := range req.ApplicableIDs {
			ids[i] = fmt.Sprintf("%d", id)
		}
		coupon.ProductIDs = strings.Join(ids, ",")
	}
	if len(req.ExcludeProducts) > 0 {
		ids := make([]string, len(req.ExcludeProducts))
		for i, id := range req.ExcludeProducts {
			ids[i] = fmt.Sprintf("%d", id)
		}
		coupon.ExcludeProducts = strings.Join(ids, ",")
	}
	if req.Stackable != nil {
	}
	coupon.Description = req.Description
	if req.Status != 0 {
		coupon.Status = req.Status
	}

	err = s.couponRepo.Update(coupon)
	if err != nil {
		return nil, err
	}

	coupon, err = s.couponRepo.GetByID(id)
	if err != nil {
		return nil, err
	}

	return s.convertToCouponResponse(coupon), nil
}

func (s *CouponService) DeleteCoupon(id uint) error {
	_, err := s.couponRepo.GetByID(id)
	if err != nil {
		return errors.New("coupon not found")
	}
	return s.couponRepo.Delete(id)
}

func (s *CouponService) ListCoupons(query *dto.CouponQueryDTO) (*dto.PageResponse, error) {
	coupons, total, err := s.couponRepo.List(
		query.Name,
		query.Type,
		query.Status,
		query.Page,
		query.PageSize,
	)
	if err != nil {
		return nil, err
	}

	var list []dto.CouponResponse
	for _, c := range coupons {
		list = append(list, *s.convertToCouponResponse(&c))
	}

	return &dto.PageResponse{
		List:  list,
		Total: total,
		Page:  query.Page,
		Size:  query.PageSize,
	}, nil
}

func (s *CouponService) IssueCoupon(req *dto.IssueCouponDTO) (*dto.IssueCouponResult, error) {
	coupon, err := s.couponRepo.GetByID(req.CouponID)
	if err != nil {
		return nil, errors.New("coupon not found")
	}

	if coupon.Status != 1 {
		return nil, errors.New("coupon is not active")
	}

	if coupon.UsedCount+len(req.MemberIDs) > coupon.TotalCount {
		return nil, errors.New("coupon stock insufficient")
	}

	memberCoupons, err := s.couponRepo.IssueToMembers(req.CouponID, req.MemberIDs)
	if err != nil {
		return nil, fmt.Errorf("issue coupon failed: %w", err)
	}

	successCount := len(memberCoupons)
	failCount := len(req.MemberIDs) - successCount

	var errorsList []string
	if failCount > 0 {
		errorsList = append(errorsList, fmt.Sprintf("%d members already reached the limit", failCount))
	}

	return &dto.IssueCouponResult{
		SuccessCount: successCount,
		FailCount:    failCount,
		Errors:       errorsList,
	}, nil
}

func (s *CouponService) VerifyCoupon(req *dto.VerifyCouponDTO) (*dto.VerifyCouponResponse, error) {
	mc, err := s.couponRepo.GetMemberCoupon(req.CouponID, req.MemberID)
	if err != nil {
		return &dto.VerifyCouponResponse{
			Valid:   false,
			Message: "coupon not found",
		}, nil
	}

	if mc.Status != 1 {
		return &dto.VerifyCouponResponse{
			Valid:   false,
			Message: "coupon already used or expired",
		}, nil
	}

	if mc.ExpireAt != nil && mc.ExpireAt.Before(time.Now()) {
		return &dto.VerifyCouponResponse{
			Valid:   false,
			Message: "coupon expired",
		}, nil
	}

	coupon := mc.Coupon
	if coupon.Status != 1 {
		return &dto.VerifyCouponResponse{
			Valid:   false,
			Message: "coupon is not active",
		}, nil
	}

	if req.Amount.LessThan(coupon.MinConsume) {
		return &dto.VerifyCouponResponse{
			Valid:   false,
			Message: fmt.Sprintf("minimum consumption %s required", coupon.MinConsume.String()),
		}, nil
	}

	if coupon.ApplyScope != "all" && len(req.ProductIDs) > 0 {
		applicableIDs := s.parseIDs(coupon.ProductIDs)
		if len(applicableIDs) > 0 {
			found := false
			for _, pid := range req.ProductIDs {
				for _, aid := range applicableIDs {
					if pid == aid {
						found = true
						break
					}
				}
				if found {
					break
				}
			}
			if !found {
				return &dto.VerifyCouponResponse{
					Valid:   false,
					Message: "coupon not applicable to these products",
				}, nil
			}
		}

		excludeIDs := s.parseIDs(coupon.ExcludeProducts)
		if len(excludeIDs) > 0 {
			for _, pid := range req.ProductIDs {
				for _, eid := range excludeIDs {
					if pid == eid {
						return &dto.VerifyCouponResponse{
							Valid:   false,
							Message: "coupon excludes some products in the order",
						}, nil
					}
				}
			}
		}
	}

	discountAmount := s.calculateDiscount(coupon, req.Amount)

	return &dto.VerifyCouponResponse{
		Valid:          true,
		DiscountAmount: discountAmount,
		Message:        "coupon is valid",
	}, nil
}

func (s *CouponService) UseCoupon(id uint, orderID uint) error {
	return s.couponRepo.UseCoupon(id, orderID)
}

func (s *CouponService) ReturnCoupon(id uint) error {
	return s.couponRepo.ReturnCoupon(id)
}

func (s *CouponService) GetMemberCoupons(query *dto.MemberCouponQueryDTO) (*dto.PageResponse, error) {
	memberCoupons, total, err := s.couponRepo.GetMemberCoupons(
		query.MemberID,
		query.Status,
		query.Page,
		query.PageSize,
	)
	if err != nil {
		return nil, err
	}

	var list []dto.MemberCouponResponse
	for _, mc := range memberCoupons {
		list = append(list, dto.MemberCouponResponse{
			ID:        mc.ID,
			MemberID:  mc.MemberID,
			CouponID:  mc.CouponID,
			Coupon:    *s.convertToCouponResponse(&mc.Coupon),
			Code:      mc.Code,
			Status:    mc.Status,
			UsedAt:    mc.UsedAt,
			ExpireAt:  mc.ExpireAt,
			OrderID:   mc.OrderID,
			CreatedAt: mc.CreatedAt,
		})
	}

	return &dto.PageResponse{
		List:  list,
		Total: total,
		Page:  query.Page,
		Size:  query.PageSize,
	}, nil
}

func (s *CouponService) GetAvailableCoupons(memberID, storeID uint, amount decimal.Decimal, productIDs []uint) ([]dto.MemberCouponResponse, error) {
	memberCoupons, err := s.couponRepo.GetAvailableCoupons(memberID, storeID, amount.InexactFloat64(), productIDs)
	if err != nil {
		return nil, err
	}

	var list []dto.MemberCouponResponse
	for _, mc := range memberCoupons {
		list = append(list, dto.MemberCouponResponse{
			ID:        mc.ID,
			MemberID:  mc.MemberID,
			CouponID:  mc.CouponID,
			Coupon:    *s.convertToCouponResponse(&mc.Coupon),
			Code:      mc.Code,
			Status:    mc.Status,
			UsedAt:    mc.UsedAt,
			ExpireAt:  mc.ExpireAt,
			OrderID:   mc.OrderID,
			CreatedAt: mc.CreatedAt,
		})
	}

	return list, nil
}

func (s *CouponService) calculateDiscount(coupon model.Coupon, amount decimal.Decimal) decimal.Decimal {
	switch coupon.Type {
	case "fixed":
		return coupon.Value
	case "percentage":
		discount := amount.Mul(coupon.DiscountRate).Div(decimal.NewFromInt(100))
		if !coupon.MaxDiscount.IsZero() && discount.GreaterThan(coupon.MaxDiscount) {
			return coupon.MaxDiscount
		}
		return discount
	default:
		return coupon.Value
	}
}

func (s *CouponService) parseIDs(idsStr string) []uint {
	if idsStr == "" {
		return nil
	}
	parts := strings.Split(idsStr, ",")
	var ids []uint
	for _, part := range parts {
		var id uint
		if _, err := fmt.Sscanf(part, "%d", &id); err == nil {
			ids = append(ids, id)
		}
	}
	return ids
}

func (s *CouponService) convertToCouponResponse(c *model.Coupon) *dto.CouponResponse {
	applicableIDs := s.parseIDs(c.ProductIDs)
	excludeProducts := s.parseIDs(c.ExcludeProducts)

	return &dto.CouponResponse{
		ID:              c.ID,
		RuleKey:         c.RuleKey,
		Name:            c.Name,
		Type:            c.Type,
		Value:           c.Value,
		MinAmount:       c.MinConsume,
		MaxDiscount:     c.MaxDiscount,
		TotalCount:      c.TotalCount,
		UsedCount:       c.UsedCount,
		PerUserLimit:    c.PerUserLimit,
		ValidityType:    c.ValidType,
		ValidityDays:    c.ValidDays,
		StartTime:       c.StartTime,
		EndTime:         c.EndTime,
		ApplicableType:  c.ApplyScope,
		ApplicableIDs:   applicableIDs,
		ExcludeProducts: excludeProducts,
		Description:     c.Description,
		Status:          c.Status,
		CreatedAt:       c.CreatedAt,
	}
}
