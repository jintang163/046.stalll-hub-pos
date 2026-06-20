package service

import (
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/shopspring/decimal"
	"stalll-hub-pos/backend/internal/dto"
	"stalll-hub-pos/backend/internal/model"
	"stalll-hub-pos/backend/internal/repository"
)

type PromotionEngineService struct {
	promotionRepo *repository.PromotionRepository
	couponRepo    *repository.CouponRepository
}

func NewPromotionEngineService() *PromotionEngineService {
	return &PromotionEngineService{
		promotionRepo: repository.NewPromotionRepository(nil),
		couponRepo:    repository.NewCouponRepository(nil),
	}
}

func (s *PromotionEngineService) CreatePromotion(storeID uint, req *dto.PromotionCreateDTO) (*dto.PromotionResponse, error) {
	applicableIDs := ""
	if len(req.ApplicableIDs) > 0 {
		ids := make([]string, len(req.ApplicableIDs))
		for i, id := range req.ApplicableIDs {
			ids[i] = fmt.Sprintf("%d", id)
		}
		applicableIDs = strings.Join(ids, ",")
	}

	promotion := &model.Promotion{
		StoreID:        storeID,
		RuleKey:        req.RuleKey,
		Name:           req.Name,
		Type:           req.Type,
		MinAmount:      req.MinAmount,
		DiscountAmount: req.DiscountAmount,
		DiscountRate:   req.DiscountRate,
		MaxDiscount:    req.MaxDiscount,
		ApplicableType: req.ApplicableType,
		ApplicableIDs:  applicableIDs,
		StartTime:      req.StartTime,
		EndTime:        req.EndTime,
		Priority:       req.Priority,
		Stackable:      req.Stackable,
		Status:         req.Status,
		Description:    req.Description,
	}

	if promotion.Status == 0 {
		promotion.Status = 1
	}

	var tiers []model.PromotionTier
	for _, t := range req.Tiers {
		tiers = append(tiers, model.PromotionTier{
			MinAmount:      t.MinAmount,
			DiscountAmount: t.DiscountAmount,
		})
	}

	err := s.promotionRepo.Create(promotion, tiers)
	if err != nil {
		return nil, fmt.Errorf("create promotion failed: %w", err)
	}

	return s.GetPromotion(promotion.ID)
}

func (s *PromotionEngineService) GetPromotion(id uint) (*dto.PromotionResponse, error) {
	promotion, err := s.promotionRepo.GetByID(id)
	if err != nil {
		return nil, err
	}
	return s.convertToPromotionResponse(promotion), nil
}

func (s *PromotionEngineService) UpdatePromotion(id uint, req *dto.PromotionUpdateDTO) (*dto.PromotionResponse, error) {
	promotion, err := s.promotionRepo.GetByID(id)
	if err != nil {
		return nil, errors.New("promotion not found")
	}

	if req.RuleKey != "" {
		promotion.RuleKey = req.RuleKey
	}
	if req.Name != "" {
		promotion.Name = req.Name
	}
	if !req.MinAmount.IsZero() {
		promotion.MinAmount = req.MinAmount
	}
	if !req.DiscountAmount.IsZero() {
		promotion.DiscountAmount = req.DiscountAmount
	}
	if !req.DiscountRate.IsZero() {
		promotion.DiscountRate = req.DiscountRate
	}
	if !req.MaxDiscount.IsZero() {
		promotion.MaxDiscount = req.MaxDiscount
	}
	if req.ApplicableType != "" {
		promotion.ApplicableType = req.ApplicableType
	}
	if len(req.ApplicableIDs) > 0 {
		ids := make([]string, len(req.ApplicableIDs))
		for i, id := range req.ApplicableIDs {
			ids[i] = fmt.Sprintf("%d", id)
		}
		promotion.ApplicableIDs = strings.Join(ids, ",")
	}
	if req.StartTime != nil {
		promotion.StartTime = req.StartTime
	}
	if req.EndTime != nil {
		promotion.EndTime = req.EndTime
	}
	if req.Priority != nil {
		promotion.Priority = *req.Priority
	}
	if req.Stackable != nil {
		promotion.Stackable = *req.Stackable
	}
	promotion.Description = req.Description
	if req.Status != 0 {
		promotion.Status = req.Status
	}

	var tiers []model.PromotionTier
	for _, t := range req.Tiers {
		tiers = append(tiers, model.PromotionTier{
			MinAmount:      t.MinAmount,
			DiscountAmount: t.DiscountAmount,
		})
	}

	err = s.promotionRepo.Update(promotion, tiers)
	if err != nil {
		return nil, err
	}

	return s.GetPromotion(id)
}

func (s *PromotionEngineService) DeletePromotion(id uint) error {
	_, err := s.promotionRepo.GetByID(id)
	if err != nil {
		return errors.New("promotion not found")
	}
	return s.promotionRepo.Delete(id)
}

func (s *PromotionEngineService) ListPromotions(query *dto.PromotionQueryDTO) (*dto.PageResponse, error) {
	if query.Page == 0 {
		query.Page = 1
	}
	if query.PageSize == 0 {
		query.PageSize = 10
	}

	promotions, total, err := s.promotionRepo.List(
		query.Name,
		query.Type,
		query.Status,
		query.Page,
		query.PageSize,
	)
	if err != nil {
		return nil, err
	}

	var list []dto.PromotionResponse
	for _, p := range promotions {
		list = append(list, *s.convertToPromotionResponse(&p))
	}

	return &dto.PageResponse{
		List:  list,
		Total: total,
		Page:  query.Page,
		Size:  query.PageSize,
	}, nil
}

func (s *PromotionEngineService) CalculatePromotion(promotion model.Promotion, amount decimal.Decimal, productIDs []uint) (decimal.Decimal, bool) {
	now := time.Now()
	if promotion.Status != 1 {
		return decimal.Zero, false
	}
	if promotion.StartTime != nil && promotion.StartTime.After(now) {
		return decimal.Zero, false
	}
	if promotion.EndTime != nil && promotion.EndTime.Before(now) {
		return decimal.Zero, false
	}

	if promotion.ApplicableType != "all" && len(productIDs) > 0 {
		applicableIDs := s.parseIDs(promotion.ApplicableIDs)
		if len(applicableIDs) > 0 {
			found := false
			for _, pid := range productIDs {
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
				return decimal.Zero, false
			}
		}
	}

	switch promotion.Type {
	case "full_reduction":
		if amount.LessThan(promotion.MinAmount) {
			return decimal.Zero, false
		}
		return promotion.DiscountAmount, true

	case "discount":
		if amount.LessThan(promotion.MinAmount) {
			return decimal.Zero, false
		}
		discount := amount.Mul(promotion.DiscountRate).Div(decimal.NewFromInt(100))
		if !promotion.MaxDiscount.IsZero() && discount.GreaterThan(promotion.MaxDiscount) {
			return promotion.MaxDiscount, true
		}
		return discount, true

	case "tiered":
		tiers, err := s.promotionRepo.GetTiers(promotion.ID)
		if err != nil || len(tiers) == 0 {
			return decimal.Zero, false
		}
		sort.Slice(tiers, func(i, j int) bool {
			return tiers[i].MinAmount.LessThan(tiers[j].MinAmount)
		})
		var maxDiscount decimal.Decimal
		for _, tier := range tiers {
			if amount.GreaterThanOrEqual(tier.MinAmount) {
				maxDiscount = tier.DiscountAmount
			}
		}
		if maxDiscount.IsZero() {
			return decimal.Zero, false
		}
		return maxDiscount, true

	default:
		return decimal.Zero, false
	}
}

func (s *PromotionEngineService) CalculateBestCombination(storeID uint, amount decimal.Decimal, productIDs []uint, memberCouponID uint, memberID uint) (*dto.BestPromotionResponse, error) {
	promotions, err := s.promotionRepo.GetActivePromotions(storeID)
	if err != nil {
		return nil, err
	}

	result := &dto.BestPromotionResponse{
		Promotions:    []dto.PromotionCalcResult{},
		TotalDiscount: decimal.Zero,
		FinalAmount:   amount,
	}

	var applicablePromos []struct {
		promotion model.Promotion
		discount  decimal.Decimal
	}

	for _, p := range promotions {
		discount, valid := s.CalculatePromotion(p, amount, productIDs)
		if valid && discount.GreaterThan(decimal.Zero) {
			applicablePromos = append(applicablePromos, struct {
				promotion model.Promotion
				discount  decimal.Decimal
			}{promotion: p, discount: discount})
		}
	}

	sort.Slice(applicablePromos, func(i, j int) bool {
		if applicablePromos[i].promotion.Priority != applicablePromos[j].promotion.Priority {
			return applicablePromos[i].promotion.Priority > applicablePromos[j].promotion.Priority
		}
		return applicablePromos[i].discount.GreaterThan(applicablePromos[j].discount)
	})

	totalDiscount := decimal.Zero
	remainingAmount := amount

	for _, ap := range applicablePromos {
		if ap.promotion.Stackable || len(result.Promotions) == 0 {
			discount := ap.discount
			if discount.GreaterThan(remainingAmount) {
				discount = remainingAmount
			}
			result.Promotions = append(result.Promotions, dto.PromotionCalcResult{
				PromotionID:    ap.promotion.ID,
				PromotionName:  ap.promotion.Name,
				PromotionType:  ap.promotion.Type,
				DiscountAmount: discount,
			})
			totalDiscount = totalDiscount.Add(discount)
			remainingAmount = remainingAmount.Sub(discount)
			if remainingAmount.LessThanOrEqual(decimal.Zero) {
				break
			}
		}
	}

	if memberCouponID > 0 && memberID > 0 {
		mc, err := s.couponRepo.GetMemberCoupon(memberCouponID, memberID)
		if err == nil && mc.Status == 1 && (mc.ExpireAt == nil || mc.ExpireAt.After(time.Now())) {
			coupon := mc.Coupon
			if coupon.Status == 1 {
				couponDiscount := s.calculateCouponDiscount(coupon, remainingAmount)
				if couponDiscount.GreaterThan(decimal.Zero) {
					if couponDiscount.GreaterThan(remainingAmount) {
						couponDiscount = remainingAmount
					}
					result.Promotions = append(result.Promotions, dto.PromotionCalcResult{
						PromotionID:    coupon.ID,
						PromotionName:  coupon.Name,
						PromotionType:  "coupon",
						DiscountAmount: couponDiscount,
					})
					totalDiscount = totalDiscount.Add(couponDiscount)
					remainingAmount = remainingAmount.Sub(couponDiscount)
				}
			}
		}
	}

	if totalDiscount.GreaterThan(amount) {
		totalDiscount = amount
	}

	result.TotalDiscount = totalDiscount
	result.FinalAmount = amount.Sub(totalDiscount)
	if result.FinalAmount.LessThan(decimal.Zero) {
		result.FinalAmount = decimal.Zero
	}

	return result, nil
}

func (s *PromotionEngineService) calculateCouponDiscount(coupon model.Coupon, amount decimal.Decimal) decimal.Decimal {
	if amount.LessThan(coupon.MinAmount) {
		return decimal.Zero
	}
	switch coupon.Type {
	case "fixed":
		return coupon.Value
	case "percentage":
		discount := amount.Mul(coupon.DiscountRate).Div(decimal.NewFromInt(100))
		if !coupon.MaxDiscount.IsZero() && discount.GreaterThan(coupon.MaxDiscount) {
			return coupon.MaxDiscount
		}
		return discount
	case "exchange":
		return coupon.Value
	default:
		return coupon.Value
	}
}

func (s *PromotionEngineService) parseIDs(idsStr string) []uint {
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

func (s *PromotionEngineService) convertToPromotionResponse(p *model.Promotion) *dto.PromotionResponse {
	applicableIDs := s.parseIDs(p.ApplicableIDs)
	tiers, _ := s.promotionRepo.GetTiers(p.ID)
	var tierDTOs []dto.PromotionTierDTO
	for _, t := range tiers {
		tierDTOs = append(tierDTOs, dto.PromotionTierDTO{
			MinAmount:      t.MinAmount,
			DiscountAmount: t.DiscountAmount,
		})
	}

	return &dto.PromotionResponse{
		ID:             p.ID,
		StoreID:        p.StoreID,
		RuleKey:        p.RuleKey,
		Name:           p.Name,
		Type:           p.Type,
		MinAmount:      p.MinAmount,
		DiscountAmount: p.DiscountAmount,
		DiscountRate:   p.DiscountRate,
		MaxDiscount:    p.MaxDiscount,
		ApplicableType: p.ApplicableType,
		ApplicableIDs:  applicableIDs,
		StartTime:      p.StartTime,
		EndTime:        p.EndTime,
		Tiers:          tierDTOs,
		Priority:       p.Priority,
		Stackable:      p.Stackable,
		Status:         p.Status,
		Description:    p.Description,
		CreatedAt:      p.CreatedAt,
	}
}

func (s *PromotionEngineService) ActivatePendingPromotions() (int64, error) {
	return s.promotionRepo.ActivatePendingPromotions()
}

func (s *PromotionEngineService) DeactivateExpiredPromotions() (int64, error) {
	return s.promotionRepo.DeactivateExpiredPromotions()
}

func (s *PromotionEngineService) GetActivePromotions(storeID uint) ([]dto.PromotionResponse, error) {
	promotions, err := s.promotionRepo.GetActivePromotions(storeID)
	if err != nil {
		return nil, err
	}

	var list []dto.PromotionResponse
	for _, p := range promotions {
		list = append(list, *s.convertToPromotionResponse(&p))
	}
	return list, nil
}
