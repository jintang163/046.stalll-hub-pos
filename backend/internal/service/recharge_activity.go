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

type RechargeActivityService struct {
	activityRepo *repository.RechargeActivityRepository
	rechargeRepo *repository.MemberRechargeRepository
	memberRepo   *repository.MemberRepository
}

func NewRechargeActivityService() *RechargeActivityService {
	return &RechargeActivityService{
		activityRepo: repository.NewRechargeActivityRepository(nil),
		rechargeRepo: repository.NewMemberRechargeRepository(nil),
		memberRepo:   repository.NewMemberRepository(nil),
	}
}

func (s *RechargeActivityService) CreateActivity(req *dto.RechargeActivityCreateDTO) (*dto.RechargeActivityResponse, error) {
	if req.EndTime.Before(*req.StartTime) {
		return nil, errors.New("end time must be after start time")
	}

	activity := &model.RechargeActivity{
		StoreID:       req.StoreID,
		Name:          req.Name,
		MinAmount:     req.MinAmount,
		BonusAmount:   req.BonusAmount,
		BonusPoints:   req.BonusPoints,
		BonusCouponID: req.BonusCouponID,
		StartTime:     req.StartTime,
		EndTime:       req.EndTime,
		AutoActivate:  req.AutoActivate,
		Status:        req.Status,
		Description:   req.Description,
	}
	if activity.Status == 0 && !activity.AutoActivate {
		activity.Status = 0
	}

	err := s.activityRepo.Create(activity)
	if err != nil {
		return nil, fmt.Errorf("create recharge activity failed: %w", err)
	}
	activity, _ = s.activityRepo.GetByID(activity.ID)
	return s.convertToActivityResponse(activity), nil
}

func (s *RechargeActivityService) GetActivity(id uint) (*dto.RechargeActivityResponse, error) {
	activity, err := s.activityRepo.GetByID(id)
	if err != nil {
		return nil, err
	}
	return s.convertToActivityResponse(activity), nil
}

func (s *RechargeActivityService) UpdateActivity(id uint, req *dto.RechargeActivityUpdateDTO) (*dto.RechargeActivityResponse, error) {
	activity, err := s.activityRepo.GetByID(id)
	if err != nil {
		return nil, errors.New("activity not found")
	}

	if req.Name != "" {
		activity.Name = req.Name
	}
	if !req.MinAmount.IsZero() {
		activity.MinAmount = req.MinAmount
	}
	activity.BonusAmount = req.BonusAmount
	if req.BonusPoints > 0 {
		activity.BonusPoints = req.BonusPoints
	}
	activity.BonusCouponID = req.BonusCouponID
	if req.StartTime != nil {
		activity.StartTime = req.StartTime
	}
	if req.EndTime != nil {
		activity.EndTime = req.EndTime
	}
	if req.AutoActivate != nil {
		activity.AutoActivate = *req.AutoActivate
	}
	if req.Status != 0 {
		activity.Status = req.Status
	}
	activity.Description = req.Description

	err = s.activityRepo.Update(activity)
	if err != nil {
		return nil, err
	}
	activity, _ = s.activityRepo.GetByID(id)
	return s.convertToActivityResponse(activity), nil
}

func (s *RechargeActivityService) DeleteActivity(id uint) error {
	return s.activityRepo.Delete(id)
}

func (s *RechargeActivityService) ListActivities(query *dto.RechargeActivityQueryDTO) (*dto.PageResponse, error) {
	if query.Page == 0 {
		query.Page = 1
	}
	if query.PageSize == 0 {
		query.PageSize = 10
	}

	activities, total, err := s.activityRepo.List(query.StoreID, query.Status, query.Page, query.PageSize)
	if err != nil {
		return nil, err
	}

	var list []dto.RechargeActivityResponse
	for _, a := range activities {
		list = append(list, *s.convertToActivityResponse(&a))
	}

	return &dto.PageResponse{
		List:  list,
		Total: total,
		Page:  query.Page,
		Size:  query.PageSize,
	}, nil
}

func (s *RechargeActivityService) MatchActivity(storeID uint, amount decimal.Decimal) (*dto.RechargeActivityMatchResult, error) {
	activities, err := s.activityRepo.GetActiveActivities(storeID)
	if err != nil {
		return nil, err
	}

	var matched *model.RechargeActivity
	for _, a := range activities {
		if amount.GreaterThanOrEqual(a.MinAmount) {
			if matched == nil || a.MinAmount.GreaterThan(matched.MinAmount) {
				matched = &a
			}
		}
	}

	if matched == nil {
		return &dto.RechargeActivityMatchResult{}, nil
	}

	return &dto.RechargeActivityMatchResult{
		ActivityID:    matched.ID,
		ActivityName:  matched.Name,
		BonusAmount:   matched.BonusAmount,
		BonusPoints:   matched.BonusPoints,
		BonusCouponID: matched.BonusCouponID,
	}, nil
}

func (s *RechargeActivityService) ProcessRecharge(req *dto.MemberRechargeDTO) (*dto.MemberRechargeResponse, error) {
	member, err := s.memberRepo.GetByID(req.MemberID)
	if err != nil {
		return nil, errors.New("member not found")
	}

	var activityID uint
	var bonusAmount decimal.Decimal
	var bonusPoints int

	matchResult, err := s.MatchActivity(member.StoreID, req.Amount)
	if err == nil && matchResult.ActivityID > 0 {
		activityID = matchResult.ActivityID
		bonusAmount = matchResult.BonusAmount
		bonusPoints = matchResult.BonusPoints
	}

	if req.ActivityID > 0 {
		activity, err := s.activityRepo.GetByID(req.ActivityID)
		if err == nil && req.Amount.GreaterThanOrEqual(activity.MinAmount) {
			activityID = activity.ID
			bonusAmount = activity.BonusAmount
			bonusPoints = activity.BonusPoints
		}
	}

	recharge := &model.MemberRecharge{
		StoreID:     member.StoreID,
		MemberID:    req.MemberID,
		Amount:      req.Amount,
		BonusAmount: bonusAmount,
		BonusPoints: bonusPoints,
		ActivityID:  activityID,
		PayMethod:   req.PayMethod,
		Status:      1,
	}

	err = database.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(recharge).Error; err != nil {
			return err
		}

		totalAdd := req.Amount.Add(bonusAmount)
		if err := tx.Model(&model.Member{}).Where("id = ?", req.MemberID).
			Update("balance", gorm.Expr("balance + ?", totalAdd)).Error; err != nil {
			return err
		}

		if bonusPoints > 0 {
			if err := tx.Model(&model.Member{}).Where("id = ?", req.MemberID).
				Updates(map[string]interface{}{
					"points":       gorm.Expr("points + ?", bonusPoints),
					"total_points": gorm.Expr("total_points + ?", bonusPoints),
				}).Error; err != nil {
				return err
			}

			memberAfter, _ := s.memberRepo.GetByID(req.MemberID)
			record := &model.MemberPointsRecord{
				MemberID: req.MemberID,
				StoreID:  member.StoreID,
				Type:     "earn",
				Points:   bonusPoints,
				Balance:  memberAfter.Points,
				Remark:   fmt.Sprintf("充值活动赠送%d积分", bonusPoints),
			}
			if err := tx.Create(record).Error; err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("process recharge failed: %w", err)
	}

	recharge, _ = s.rechargeRepo.GetByID(recharge.ID)
	return s.convertToRechargeResponse(recharge), nil
}

func (s *RechargeActivityService) ListRecharges(memberID, storeID uint, page, pageSize int) (*dto.PageResponse, error) {
	if page == 0 {
		page = 1
	}
	if pageSize == 0 {
		pageSize = 10
	}

	recharges, total, err := s.rechargeRepo.List(memberID, storeID, page, pageSize)
	if err != nil {
		return nil, err
	}

	var list []dto.MemberRechargeResponse
	for _, r := range recharges {
		list = append(list, *s.convertToRechargeResponse(&r))
	}

	return &dto.PageResponse{
		List:  list,
		Total: total,
		Page:  page,
		Size:  pageSize,
	}, nil
}

func (s *RechargeActivityService) ActivatePendingActivities() (int, error) {
	activities, err := s.activityRepo.GetPendingAutoActivate()
	if err != nil {
		return 0, err
	}

	count := 0
	for _, a := range activities {
		a.Status = 1
		if err := s.activityRepo.Update(&a); err == nil {
			count++
		}
	}
	return count, nil
}

func (s *RechargeActivityService) DeactivateExpiredActivities() (int, error) {
	activities, err := s.activityRepo.GetExpiredActive()
	if err != nil {
		return 0, err
	}

	count := 0
	for _, a := range activities {
		a.Status = 2
		if err := s.activityRepo.Update(&a); err == nil {
			count++
		}
	}
	return count, nil
}

func (s *RechargeActivityService) convertToActivityResponse(a *model.RechargeActivity) *dto.RechargeActivityResponse {
	return &dto.RechargeActivityResponse{
		ID:            a.ID,
		StoreID:       a.StoreID,
		Name:          a.Name,
		MinAmount:     a.MinAmount,
		BonusAmount:   a.BonusAmount,
		BonusPoints:   a.BonusPoints,
		BonusCouponID: a.BonusCouponID,
		StartTime:     a.StartTime,
		EndTime:       a.EndTime,
		AutoActivate:  a.AutoActivate,
		Status:        a.Status,
		Description:   a.Description,
		CreatedAt:     a.CreatedAt,
		UpdatedAt:     a.UpdatedAt,
	}
}

func (s *RechargeActivityService) convertToRechargeResponse(r *model.MemberRecharge) *dto.MemberRechargeResponse {
	return &dto.MemberRechargeResponse{
		ID:          r.ID,
		StoreID:     r.StoreID,
		MemberID:    r.MemberID,
		Amount:      r.Amount,
		BonusAmount: r.BonusAmount,
		BonusPoints: r.BonusPoints,
		ActivityID:  r.ActivityID,
		PayMethod:   r.PayMethod,
		Status:      r.Status,
		CreatedAt:   r.CreatedAt,
	}
}
