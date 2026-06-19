package service

import (
	"errors"
	"fmt"
	"stalll-hub-pos/backend/internal/dto"
	"stalll-hub-pos/backend/internal/model"
	"stalll-hub-pos/backend/internal/repository"
	"stalll-hub-pos/backend/pkg/database"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type PointsEngineService struct {
	ruleRepo   *repository.PointsRuleRepository
	memberRepo *repository.MemberRepository
}

func NewPointsEngineService() *PointsEngineService {
	return &PointsEngineService{
		ruleRepo:   repository.NewPointsRuleRepository(nil),
		memberRepo: repository.NewMemberRepository(nil),
	}
}

func (s *PointsEngineService) CreateRule(req *dto.PointsRuleCreateDTO) (*dto.PointsRuleResponse, error) {
	rule := &model.PointsRule{
		StoreID:          req.StoreID,
		RuleKey:          req.RuleKey,
		RuleName:         req.RuleName,
		RuleType:         req.RuleType,
		PointsPerYuan:    req.PointsPerYuan,
		RedeemRate:       req.RedeemRate,
		MinRedeemPoints:  req.MinRedeemPoints,
		BonusPoints:      req.BonusPoints,
		MinConsumeAmount: req.MinConsumeAmount,
		Priority:         req.Priority,
		Status:           req.Status,
	}
	if rule.Status == 0 {
		rule.Status = 1
	}
	if rule.PointsPerYuan.IsZero() {
		rule.PointsPerYuan = decimal.NewFromInt(1)
	}
	if rule.RedeemRate.IsZero() && rule.RuleType == "redeem" {
		rule.RedeemRate = decimal.NewFromFloat(0.01)
	}

	err := s.ruleRepo.Create(rule)
	if err != nil {
		return nil, fmt.Errorf("create points rule failed: %w", err)
	}
	rule, _ = s.ruleRepo.GetByID(rule.ID)
	return s.convertToRuleResponse(rule), nil
}

func (s *PointsEngineService) GetRule(id uint) (*dto.PointsRuleResponse, error) {
	rule, err := s.ruleRepo.GetByID(id)
	if err != nil {
		return nil, err
	}
	return s.convertToRuleResponse(rule), nil
}

func (s *PointsEngineService) UpdateRule(id uint, req *dto.PointsRuleUpdateDTO) (*dto.PointsRuleResponse, error) {
	rule, err := s.ruleRepo.GetByID(id)
	if err != nil {
		return nil, errors.New("rule not found")
	}

	if req.RuleName != "" {
		rule.RuleName = req.RuleName
	}
	if !req.PointsPerYuan.IsZero() {
		rule.PointsPerYuan = req.PointsPerYuan
	}
	if !req.RedeemRate.IsZero() {
		rule.RedeemRate = req.RedeemRate
	}
	if req.MinRedeemPoints > 0 {
		rule.MinRedeemPoints = req.MinRedeemPoints
	}
	if req.BonusPoints > 0 {
		rule.BonusPoints = req.BonusPoints
	}
	rule.MinConsumeAmount = req.MinConsumeAmount
	rule.Priority = req.Priority
	if req.Status != 0 {
		rule.Status = req.Status
	}

	err = s.ruleRepo.Update(rule)
	if err != nil {
		return nil, err
	}
	rule, _ = s.ruleRepo.GetByID(id)
	return s.convertToRuleResponse(rule), nil
}

func (s *PointsEngineService) DeleteRule(id uint) error {
	return s.ruleRepo.Delete(id)
}

func (s *PointsEngineService) ListRules(query *dto.PointsRuleQueryDTO) (*dto.PageResponse, error) {
	if query.Page == 0 {
		query.Page = 1
	}
	if query.PageSize == 0 {
		query.PageSize = 10
	}

	rules, total, err := s.ruleRepo.List(query.StoreID, query.RuleType, query.Status, query.Page, query.PageSize)
	if err != nil {
		return nil, err
	}

	var list []dto.PointsRuleResponse
	for _, r := range rules {
		list = append(list, *s.convertToRuleResponse(&r))
	}

	return &dto.PageResponse{
		List:  list,
		Total: total,
		Page:  query.Page,
		Size:  query.PageSize,
	}, nil
}

func (s *PointsEngineService) CalculateEarnedPoints(storeID uint, consumeAmount decimal.Decimal, levelPointsRate decimal.Decimal) int {
	rules, err := s.ruleRepo.GetActiveRulesByType(storeID, "earn")
	if err != nil || len(rules) == 0 {
		earned := consumeAmount.Mul(decimal.NewFromInt(1)).Mul(levelPointsRate)
		return int(earned.Floor().IntPart())
	}

	rule := rules[0]
	if !rule.MinConsumeAmount.IsZero() && consumeAmount.LessThan(rule.MinConsumeAmount) {
		return 0
	}

	earned := consumeAmount.Mul(rule.PointsPerYuan).Mul(levelPointsRate)
	return int(earned.Floor().IntPart())
}

func (s *PointsEngineService) CalculateRedemptionDiscount(storeID uint, points int) (decimal.Decimal, int, error) {
	if points <= 0 {
		return decimal.Zero, 0, nil
	}

	rules, err := s.ruleRepo.GetActiveRulesByType(storeID, "redeem")
	if err != nil || len(rules) == 0 {
		if points < 100 {
			return decimal.Zero, 0, nil
		}
		usablePoints := (points / 100) * 100
		discount := decimal.NewFromInt(int64(usablePoints)).Mul(decimal.NewFromFloat(0.01))
		return discount, usablePoints, nil
	}

	rule := rules[0]
	if points < rule.MinRedeemPoints {
		return decimal.Zero, 0, errors.New(fmt.Sprintf("minimum %d points required for redemption", rule.MinRedeemPoints))
	}

	unitPoints := rule.MinRedeemPoints
	usablePoints := (points / unitPoints) * unitPoints
	discount := decimal.NewFromInt(int64(usablePoints)).Mul(rule.RedeemRate)
	return discount, usablePoints, nil
}

func (s *PointsEngineService) CalculateRegisterBonus(storeID uint) int {
	rules, err := s.ruleRepo.GetActiveRulesByType(storeID, "register")
	if err != nil || len(rules) == 0 {
		return 0
	}
	return rules[0].BonusPoints
}

func (s *PointsEngineService) ProcessOrderPoints(memberID, storeID, orderID uint, consumeAmount decimal.Decimal, usePoints int, levelPointsRate decimal.Decimal) (*dto.PointsCalcResult, error) {
	member, err := s.memberRepo.GetByID(memberID)
	if err != nil {
		return nil, errors.New("member not found")
	}

	result := &dto.PointsCalcResult{}

	if usePoints > 0 {
		if member.Points < usePoints {
			return nil, errors.New("insufficient points")
		}

		discount, usablePoints, err := s.CalculateRedemptionDiscount(storeID, usePoints)
		if err != nil {
			return nil, err
		}

		result.DiscountAmount = discount
		result.PointsUsed = usablePoints
	}

	earnedPoints := s.CalculateEarnedPoints(storeID, consumeAmount, levelPointsRate)
	result.PointsEarned = earnedPoints

	return s.applyPointsCalc(member, storeID, orderID, result)
}

func (s *PointsEngineService) applyPointsCalc(member *model.Member, storeID, orderID uint, result *dto.PointsCalcResult) (*dto.PointsCalcResult, error) {
	err := database.DB.Transaction(func(tx *gorm.DB) error {
		if result.PointsUsed > 0 {
			newPoints := member.Points - result.PointsUsed
			if newPoints < 0 {
				newPoints = 0
			}
			if err := tx.Model(&model.Member{}).Where("id = ?", member.ID).Update("points", newPoints).Error; err != nil {
				return err
			}
			record := &model.MemberPointsRecord{
				MemberID: member.ID,
				StoreID:  storeID,
				Type:     "spend",
				Points:   -result.PointsUsed,
				Balance:  newPoints,
				OrderID:  orderID,
				Remark:   fmt.Sprintf("积分抵扣%.2f元", result.DiscountAmount.InexactFloat64()),
			}
			if err := tx.Create(record).Error; err != nil {
				return err
			}
			member.Points = newPoints
		}

		if result.PointsEarned > 0 {
			newPoints := member.Points + result.PointsEarned
			newTotalPoints := member.TotalPoints + result.PointsEarned
			if err := tx.Model(&model.Member{}).Where("id = ?", member.ID).
				Updates(map[string]interface{}{"points": newPoints, "total_points": newTotalPoints}).Error; err != nil {
				return err
			}
			record := &model.MemberPointsRecord{
				MemberID: member.ID,
				StoreID:  storeID,
				Type:     "earn",
				Points:   result.PointsEarned,
				Balance:  newPoints,
				OrderID:  orderID,
				Remark:   "消费赠送积分",
			}
			if err := tx.Create(record).Error; err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}
	return result, nil
}

func (s *PointsEngineService) convertToRuleResponse(r *model.PointsRule) *dto.PointsRuleResponse {
	return &dto.PointsRuleResponse{
		ID:               r.ID,
		StoreID:          r.StoreID,
		RuleKey:          r.RuleKey,
		RuleName:         r.RuleName,
		RuleType:         r.RuleType,
		PointsPerYuan:    r.PointsPerYuan,
		RedeemRate:       r.RedeemRate,
		MinRedeemPoints:  r.MinRedeemPoints,
		BonusPoints:      r.BonusPoints,
		MinConsumeAmount: r.MinConsumeAmount,
		Priority:         r.Priority,
		Status:           r.Status,
		CreatedAt:        r.CreatedAt,
		UpdatedAt:        r.UpdatedAt,
	}
}
