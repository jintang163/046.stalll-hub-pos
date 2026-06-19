package service

import (
	"log"
	"stalll-hub-pos/backend/internal/model"
	"stalll-hub-pos/backend/pkg/database"
	"time"
)

type SchedulerService struct {
	rechargeService *RechargeActivityService
}

func NewSchedulerService() *SchedulerService {
	return &SchedulerService{
		rechargeService: NewRechargeActivityService(),
	}
}

func (s *SchedulerService) StartAllSchedulers() {
	go s.runBirthdayCouponScheduler()
	go s.runRechargeActivityScheduler()
	log.Println("[Scheduler] All schedulers started")
}

func (s *SchedulerService) runBirthdayCouponScheduler() {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	s.checkBirthdayCoupons()

	for range ticker.C {
		s.checkBirthdayCoupons()
	}
}

func (s *SchedulerService) runRechargeActivityScheduler() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	s.checkRechargeActivities()

	for range ticker.C {
		s.checkRechargeActivities()
	}
}

func (s *SchedulerService) checkBirthdayCoupons() {
	now := time.Now()
	todayStr := now.Format("01-02")

	var birthdayCoupons []model.Coupon
	if err := database.DB.Where("rule_key = ? AND status = 1", "birthday").Find(&birthdayCoupons).Error; err != nil {
		log.Printf("[Scheduler] Failed to find birthday coupons: %v", err)
		return
	}

	if len(birthdayCoupons) == 0 {
		var generalCoupons []model.Coupon
		if err := database.DB.Where("name LIKE ? AND status = 1", "%生日%").Find(&generalCoupons).Error; err == nil {
			birthdayCoupons = generalCoupons
		}
	}

	if len(birthdayCoupons) == 0 {
		return
	}

	var members []model.Member
	if err := database.DB.Where("DATE_FORMAT(birthday, '%m-%d') = ? AND status = 1", todayStr).
		Preload("Store").Find(&members).Error; err != nil {
		log.Printf("[Scheduler] Failed to find birthday members: %v", err)
		return
	}

	if len(members) == 0 {
		return
	}

	log.Printf("[Scheduler] Found %d members with birthday today (%s)", len(members), todayStr)

	for _, member := range members {
		for _, coupon := range birthdayCoupons {
			var existingCount int64
			database.DB.Model(&model.MemberCoupon{}).
				Where("coupon_id = ? AND member_id = ? AND DATE(created_at) = ?", coupon.ID, member.ID, now.Format("2006-01-02")).
				Count(&existingCount)

			if existingCount > 0 {
				continue
			}

			var expireAt *time.Time
			if coupon.ValidType == "fixed" && coupon.EndTime != nil {
				expireAt = coupon.EndTime
			} else if coupon.ValidType == "relative" && coupon.ValidDays > 0 {
				t := now.AddDate(0, 0, coupon.ValidDays)
				expireAt = &t
			} else {
				t := now.AddDate(0, 0, 30)
				expireAt = &t
			}

			mc := &model.MemberCoupon{
				StoreID:    member.StoreID,
				MemberID:   member.ID,
				CouponID:   coupon.ID,
				Status:     1,
				ExpireTime: expireAt,
			}

			if err := database.DB.Create(mc).Error; err != nil {
				log.Printf("[Scheduler] Failed to issue birthday coupon to member %d: %v", member.ID, err)
			} else {
				log.Printf("[Scheduler] Issued birthday coupon %d to member %d", coupon.ID, member.ID)
			}
		}
	}
}

func (s *SchedulerService) checkRechargeActivities() {
	activated, err := s.rechargeService.ActivatePendingActivities()
	if err != nil {
		log.Printf("[Scheduler] Failed to activate pending recharge activities: %v", err)
	} else if activated > 0 {
		log.Printf("[Scheduler] Activated %d recharge activities", activated)
	}

	deactivated, err := s.rechargeService.DeactivateExpiredActivities()
	if err != nil {
		log.Printf("[Scheduler] Failed to deactivate expired recharge activities: %v", err)
	} else if deactivated > 0 {
		log.Printf("[Scheduler] Deactivated %d expired recharge activities", deactivated)
	}
}
