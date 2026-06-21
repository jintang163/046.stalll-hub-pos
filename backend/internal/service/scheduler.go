package service

import (
	"log"
	"stalll-hub-pos/backend/internal/model"
	"stalll-hub-pos/backend/pkg/database"
	"time"
)

type SchedulerService struct {
	rechargeService  *RechargeActivityService
	promotionService *PromotionEngineService
	stallService     *StallService
	dingTalk         *DingTalkService
	forecastService  *ForecastService
	purchaseService  *PurchaseService
	timeSlotService  *TimeSlotPricingService
}

func NewSchedulerService() *SchedulerService {
	return &SchedulerService{
		rechargeService:  NewRechargeActivityService(),
		promotionService: NewPromotionEngineService(),
		stallService:     NewStallService(),
		dingTalk:         NewDingTalkService(),
		forecastService:  NewForecastService(),
		purchaseService:  NewPurchaseService(),
		timeSlotService:  NewTimeSlotPricingService(),
	}
}

func (s *SchedulerService) StartAllSchedulers() {
	go s.runBirthdayCouponScheduler()
	go s.runRechargeActivityScheduler()
	go s.runPromotionScheduler()
	go s.runCouponStatusScheduler()
	go s.runStallDeviceCheckScheduler()
	go s.runStallDailyReportScheduler()
	go s.runStockWarningScheduler()
	go s.runStockReservationCleaner()
	go s.runReservationReminder()
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

func (s *SchedulerService) runPromotionScheduler() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	s.checkPromotions()

	for range ticker.C {
		s.checkPromotions()
	}
}

func (s *SchedulerService) runCouponStatusScheduler() {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	s.checkCouponStatus()

	for range ticker.C {
		s.checkCouponStatus()
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
			if coupon.ValidityType == "fixed" && coupon.EndTime != nil {
				expireAt = coupon.EndTime
			} else if coupon.ValidityType == "relative" && coupon.ValidityDays > 0 {
				t := now.AddDate(0, 0, coupon.ValidityDays)
				expireAt = &t
			} else {
				t := now.AddDate(0, 0, 30)
				expireAt = &t
			}

			mc := &model.MemberCoupon{
				StoreID:  member.StoreID,
				MemberID: member.ID,
				CouponID: coupon.ID,
				Status:   1,
				ExpireAt: expireAt,
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

func (s *SchedulerService) checkPromotions() {
	activated, err := s.promotionService.ActivatePendingPromotions()
	if err != nil {
		log.Printf("[Scheduler] Failed to activate pending promotions: %v", err)
	} else if activated > 0 {
		log.Printf("[Scheduler] Activated %d promotions", activated)
	}

	deactivated, err := s.promotionService.DeactivateExpiredPromotions()
	if err != nil {
		log.Printf("[Scheduler] Failed to deactivate expired promotions: %v", err)
	} else if deactivated > 0 {
		log.Printf("[Scheduler] Deactivated %d expired promotions", deactivated)
	}
}

func (s *SchedulerService) checkCouponStatus() {
	now := time.Now()

	result := database.DB.Model(&model.Coupon{}).
		Where("status = 1 AND end_time IS NOT NULL AND end_time < ?", now).
		Update("status", 2)
	if result.Error != nil {
		log.Printf("[Scheduler] Failed to deactivate expired coupons: %v", result.Error)
	} else if result.RowsAffected > 0 {
		log.Printf("[Scheduler] Deactivated %d expired coupons", result.RowsAffected)
	}

	result2 := database.DB.Model(&model.MemberCoupon{}).
		Where("status = 1 AND expire_at IS NOT NULL AND expire_at < ?", now).
		Update("status", 3)
	if result2.Error != nil {
		log.Printf("[Scheduler] Failed to expire member coupons: %v", result2.Error)
	} else if result2.RowsAffected > 0 {
		log.Printf("[Scheduler] Expired %d member coupons", result2.RowsAffected)
	}
}

func (s *SchedulerService) runStallDeviceCheckScheduler() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		s.checkStallDevicesOffline()
	}
}

func (s *SchedulerService) checkStallDevicesOffline() {
	alertCount, err := s.stallService.CheckOfflineDevices()
	if err != nil {
		log.Printf("[Scheduler] Failed to check offline stall devices: %v", err)
	} else if alertCount > 0 {
		log.Printf("[Scheduler] Found %d offline stall devices", alertCount)
	}
}

func (s *SchedulerService) runStallDailyReportScheduler() {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for range ticker.C {
		now := time.Now()
		if now.Hour() == 2 {
			s.generateStallDailyReports()
		}
	}
}

func (s *SchedulerService) generateStallDailyReports() {
	yesterday := time.Now().AddDate(0, 0, -1).Format("2006-01-02")
	log.Printf("[Scheduler] Generating stall daily reports for %s", yesterday)
}

func (s *SchedulerService) runStockWarningScheduler() {
	ticker := time.NewTicker(6 * time.Hour)
	defer ticker.Stop()

	s.checkStockWarnings()

	for range ticker.C {
		now := time.Now()
		if now.Hour() == 9 || now.Hour() == 15 {
			s.checkStockWarnings()
		}
	}
}

func (s *SchedulerService) checkStockWarnings() {
	log.Println("[Scheduler] Checking stock warnings...")

	var stores []model.Store
	if err := database.DB.Where("status = 1").Find(&stores).Error; err != nil {
		log.Printf("[Scheduler] Failed to find stores: %v", err)
		return
	}

	for _, store := range stores {
		var warnings []model.StockWarning

		var skus []model.ProductSKU
		if err := database.DB.Model(&model.ProductSKU{}).
			Where("store_id = ? AND status = 1", store.ID).
			Preload("Product").
			Find(&skus).Error; err != nil {
			log.Printf("[Scheduler] Failed to find SKUs for store %d: %v", store.ID, err)
			continue
		}

		for _, sku := range skus {
			threshold := sku.Product.StockWarningThreshold
			if threshold <= 0 {
				threshold = 10
			}

			if sku.Stock < threshold {
				warnings = append(warnings, model.StockWarning{
					StoreID:      store.ID,
					SKUID:        sku.ID,
					ProductID:    sku.ProductID,
					CurrentStock: sku.Stock,
					Threshold:    threshold,
					Status:       1,
				})
			}
		}

		if len(warnings) > 0 {
			log.Printf("[Scheduler] Store %s: %d SKUs below warning threshold", store.Name, len(warnings))
			go s.dingTalk.SendStockWarning(warnings, store.Name)
		}
	}
}

func (s *SchedulerService) StartForecastScheduler() {
	go s.runForecastScheduler()
	log.Println("[Scheduler] Forecast scheduler started")
}

func (s *SchedulerService) runForecastScheduler() {
	ticker := time.NewTicker(24 * time.Hour)
	defer ticker.Stop()

	go s.runDailyForecast()

	for range ticker.C {
		now := time.Now()
		if now.Hour() == 3 {
			s.runDailyForecast()
		}
	}
}

func (s *SchedulerService) runDailyForecast() {
	log.Println("[Forecast] Starting daily forecast task...")

	var stores []model.Store
	if err := database.DB.Where("status = 1").Find(&stores).Error; err != nil {
		log.Printf("[Forecast] Failed to fetch stores: %v", err)
		return
	}

	for _, store := range stores {
		go func(store model.Store) {
			defer func() {
				if r := recover(); r != nil {
					log.Printf("[Forecast] Panic in store %d forecast: %v", store.ID, r)
				}
			}()

			forecast, err := s.forecastService.GetStoreForecast(store.ID, 0, 0)
			if err != nil {
				log.Printf("[Forecast] Store %d forecast failed: %v", store.ID, err)
				return
			}

			suggestions, err := s.forecastService.CalculateStockingSuggestion(store.ID, forecast)
			if err != nil {
				log.Printf("[Forecast] Store %d stocking suggestion failed: %v", store.ID, err)
				return
			}

			purchaseOrders, err := s.purchaseService.AutoGenerateFromForecast(
				store.ID, forecast, suggestions,
			)
			if err != nil {
				log.Printf("[Forecast] Store %d auto generate purchase order failed: %v", store.ID, err)
				return
			}

			for _, purchase := range purchaseOrders {
				if sendErr := s.purchaseService.SendToSupplier(purchase.ID); sendErr != nil {
					log.Printf("[Forecast] Store %d failed to send purchase order %s to supplier %s: %v",
						store.ID, purchase.PurchaseNo, purchase.SupplierName, sendErr)
					continue
				}
				log.Printf("[Forecast] Store %s: generated and sent purchase order %s (supplier: %s), %d items, total: %s",
					store.Name, purchase.PurchaseNo, purchase.SupplierName, purchase.ItemCount, purchase.TotalAmount.String())
			}
		}(store)
	}

	log.Println("[Forecast] Daily forecast tasks submitted")
}

func (s *SchedulerService) runStockReservationCleaner() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	s.checkStockReservations()

	for range ticker.C {
		s.checkStockReservations()
	}
}

func (s *SchedulerService) runReservationReminder() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		s.checkReservationReminders()
	}
}

func (s *SchedulerService) checkStockReservations() {
	count, err := s.timeSlotService.CleanExpiredReservations()
	if err != nil {
		log.Printf("[Scheduler] Failed to clean expired stock reservations: %v", err)
	} else if count > 0 {
		log.Printf("[Scheduler] Cleaned %d expired stock reservations", count)
	}
}

func (s *SchedulerService) checkReservationReminders() {
	sentCount, err := s.timeSlotService.ProcessPendingReminders()
	if err != nil {
		log.Printf("[Scheduler] Failed to process reservation reminders: %v", err)
	} else if sentCount > 0 {
		log.Printf("[Scheduler] Sent %d reservation reminders", sentCount)
	}
}
