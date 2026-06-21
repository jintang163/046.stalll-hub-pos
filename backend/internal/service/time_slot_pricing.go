package service

import (
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
	"stalll-hub-pos/backend/internal/dto"
	"stalll-hub-pos/backend/internal/model"
	"stalll-hub-pos/backend/internal/repository"
	"stalll-hub-pos/backend/pkg/database"
)

type TimeSlotPricingService struct {
	timeSlotRepo *repository.TimeSlotPricingRepository
	stockResRepo *repository.StockReservationRepository
	reminderRepo *repository.ReservationReminderRepository
	productRepo  *repository.ProductRepository
	orderRepo    *repository.OrderRepository
}

func NewTimeSlotPricingService() *TimeSlotPricingService {
	return &TimeSlotPricingService{
		timeSlotRepo: repository.NewTimeSlotPricingRepository(),
		stockResRepo: repository.NewStockReservationRepository(),
		reminderRepo: repository.NewReservationReminderRepository(),
		productRepo:  repository.NewProductRepository(),
		orderRepo:    repository.NewOrderRepository(),
	}
}

func (s *TimeSlotPricingService) CreateTimeSlotPricing(storeID uint, req *dto.TimeSlotPricingCreateDTO) (*dto.TimeSlotPricingResponse, error) {
	applicableIDs := ""
	if len(req.ApplicableIDs) > 0 {
		ids := make([]string, len(req.ApplicableIDs))
		for i, id := range req.ApplicableIDs {
			ids[i] = fmt.Sprintf("%d", id)
		}
		applicableIDs = strings.Join(ids, ",")
	}

	weekdays := ""
	if len(req.Weekdays) > 0 {
		days := make([]string, len(req.Weekdays))
		for i, day := range req.Weekdays {
			days[i] = fmt.Sprintf("%d", day)
		}
		weekdays = strings.Join(days, ",")
	} else {
		weekdays = "1,2,3,4,5,6,7"
	}

	pricing := &model.TimeSlotPricing{
		StoreID:        storeID,
		Name:           req.Name,
		StartTime:      req.StartTime,
		EndTime:        req.EndTime,
		PricingType:    "discount",
		DiscountRate:   req.DiscountRate,
		DiscountAmount: decimal.Zero,
		MinAmount:      decimal.Zero,
		ApplicableType: req.ApplicableType,
		ApplicableIDs:  applicableIDs,
		Weekdays:       weekdays,
		Priority:       100,
		Status:         req.Status,
		Description:    req.Description,
	}

	if pricing.Status == 0 {
		pricing.Status = 1
	}

	err := s.timeSlotRepo.Create(pricing)
	if err != nil {
		return nil, fmt.Errorf("create time slot pricing failed: %w", err)
	}

	return s.GetTimeSlotPricing(pricing.ID)
}

func (s *TimeSlotPricingService) UpdateTimeSlotPricing(id uint, req *dto.TimeSlotPricingUpdateDTO) (*dto.TimeSlotPricingResponse, error) {
	pricing, err := s.timeSlotRepo.GetByID(id)
	if err != nil {
		return nil, errors.New("time slot pricing not found")
	}

	if req.Name != "" {
		pricing.Name = req.Name
	}
	if req.StartTime != "" {
		pricing.StartTime = req.StartTime
	}
	if req.EndTime != "" {
		pricing.EndTime = req.EndTime
	}
	if !req.DiscountRate.IsZero() {
		pricing.DiscountRate = req.DiscountRate
	}
	if req.ApplicableType != "" {
		pricing.ApplicableType = req.ApplicableType
	}
	if len(req.ApplicableIDs) > 0 {
		ids := make([]string, len(req.ApplicableIDs))
		for i, id := range req.ApplicableIDs {
			ids[i] = fmt.Sprintf("%d", id)
		}
		pricing.ApplicableIDs = strings.Join(ids, ",")
	}
	if len(req.Weekdays) > 0 {
		days := make([]string, len(req.Weekdays))
		for i, day := range req.Weekdays {
			days[i] = fmt.Sprintf("%d", day)
		}
		pricing.Weekdays = strings.Join(days, ",")
	}
	if req.Status != 0 {
		pricing.Status = req.Status
	}
	pricing.Description = req.Description

	err = s.timeSlotRepo.Update(pricing)
	if err != nil {
		return nil, err
	}

	return s.GetTimeSlotPricing(id)
}

func (s *TimeSlotPricingService) DeleteTimeSlotPricing(id uint) error {
	_, err := s.timeSlotRepo.GetByID(id)
	if err != nil {
		return errors.New("time slot pricing not found")
	}
	return s.timeSlotRepo.Delete(id)
}

func (s *TimeSlotPricingService) GetTimeSlotPricing(id uint) (*dto.TimeSlotPricingResponse, error) {
	pricing, err := s.timeSlotRepo.GetByID(id)
	if err != nil {
		return nil, err
	}
	return s.convertToTimeSlotPricingResponse(pricing), nil
}

func (s *TimeSlotPricingService) ListTimeSlotPricings(query *dto.TimeSlotPricingQueryDTO) (*dto.PageResponse, error) {
	if query.Page == 0 {
		query.Page = 1
	}
	if query.PageSize == 0 {
		query.PageSize = 10
	}

	offset := (query.Page - 1) * query.PageSize

	var status *int
	if query.Status > 0 {
		s := query.Status
		status = &s
	}

	pricings, total, err := s.timeSlotRepo.List(
		query.StoreID,
		query.Name,
		status,
		offset,
		query.PageSize,
	)
	if err != nil {
		return nil, err
	}

	var list []dto.TimeSlotPricingResponse
	for _, p := range pricings {
		list = append(list, *s.convertToTimeSlotPricingResponse(&p))
	}

	return &dto.PageResponse{
		List:  list,
		Total: total,
		Page:  query.Page,
		Size:  query.PageSize,
	}, nil
}

func (s *TimeSlotPricingService) GetActiveTimeSlots(storeID uint) ([]dto.TimeSlotPricingResponse, error) {
	pricings, err := s.timeSlotRepo.GetActivePricings(storeID)
	if err != nil {
		return nil, err
	}

	var list []dto.TimeSlotPricingResponse
	for _, p := range pricings {
		list = append(list, *s.convertToTimeSlotPricingResponse(&p))
	}
	return list, nil
}

func (s *TimeSlotPricingService) CalculatePrice(storeID uint, skuID uint, quantity int, checkTime time.Time) (*dto.TimeSlotPriceCalculateResponse, error) {
	if quantity <= 0 {
		return nil, errors.New("quantity must be greater than 0")
	}

	sku, err := s.productRepo.GetSKUByID(skuID)
	if err != nil {
		return nil, fmt.Errorf("SKU not found: %w", err)
	}

	timeSlot, err := s.GetApplicableTimeSlot(storeID, checkTime)
	if err != nil {
		return &dto.TimeSlotPriceCalculateResponse{
			OriginalAmount: sku.Price.Mul(decimal.NewFromInt(int64(quantity))),
			DiscountAmount: decimal.Zero,
			FinalAmount:    sku.Price.Mul(decimal.NewFromInt(int64(quantity))),
			Available:      true,
		}, nil
	}

	if !s.isSKUApplicable(timeSlot, sku.ProductID) {
		return &dto.TimeSlotPriceCalculateResponse{
			OriginalAmount: sku.Price.Mul(decimal.NewFromInt(int64(quantity))),
			DiscountAmount: decimal.Zero,
			FinalAmount:    sku.Price.Mul(decimal.NewFromInt(int64(quantity))),
			Available:      true,
		}, nil
	}

	originalAmount := sku.Price.Mul(decimal.NewFromInt(int64(quantity)))
	discountAmount := s.calculateTimeSlotDiscount(timeSlot, originalAmount)
	finalAmount := originalAmount.Sub(discountAmount)

	if finalAmount.LessThan(decimal.Zero) {
		finalAmount = decimal.Zero
	}

	return &dto.TimeSlotPriceCalculateResponse{
		TimeSlotID:       timeSlot.ID,
		TimeSlotName:     timeSlot.Name,
		StartTime:        timeSlot.StartTime,
		EndTime:          timeSlot.EndTime,
		OriginalAmount:   originalAmount,
		DiscountAmount:   discountAmount,
		TimeSlotDiscount: discountAmount,
		FinalAmount:      finalAmount,
		ApplicableItems:  []uint{skuID},
		Available:        true,
	}, nil
}

func (s *TimeSlotPricingService) CalculateOrderPrices(storeID uint, items []dto.OrderItemDTO, checkTime time.Time) (decimal.Decimal, []dto.OrderItemDTO, error) {
	if len(items) == 0 {
		return decimal.Zero, items, nil
	}

	timeSlot, err := s.GetApplicableTimeSlot(storeID, checkTime)
	if err != nil {
		total := decimal.Zero
		for i := range items {
			subtotal := items[i].Price.Mul(decimal.NewFromInt(int64(items[i].Quantity)))
			total = total.Add(subtotal)
		}
		return total, items, nil
	}

	totalOriginal := decimal.Zero
	totalDiscount := decimal.Zero
	updatedItems := make([]dto.OrderItemDTO, len(items))

	for i, item := range items {
		updatedItems[i] = item
		subtotal := item.Price.Mul(decimal.NewFromInt(int64(item.Quantity)))
		totalOriginal = totalOriginal.Add(subtotal)

		if s.isSKUApplicable(timeSlot, item.ProductID) {
			discount := s.calculateTimeSlotDiscount(timeSlot, subtotal)
			totalDiscount = totalDiscount.Add(discount)
		}
	}

	totalFinal := totalOriginal.Sub(totalDiscount)
	if totalFinal.LessThan(decimal.Zero) {
		totalFinal = decimal.Zero
	}

	return totalFinal, updatedItems, nil
}

func (s *TimeSlotPricingService) GetApplicableTimeSlot(storeID uint, checkTime time.Time) (*model.TimeSlotPricing, error) {
	pricings, err := s.timeSlotRepo.GetActiveByStoreAndTime(storeID, checkTime)
	if err != nil {
		return nil, err
	}

	if len(pricings) == 0 {
		return nil, errors.New("no applicable time slot found")
	}

	sort.Slice(pricings, func(i, j int) bool {
		if pricings[i].Priority != pricings[j].Priority {
			return pricings[i].Priority > pricings[j].Priority
		}
		return pricings[i].ID < pricings[j].ID
	})

	return &pricings[0], nil
}

func (s *TimeSlotPricingService) ReserveStock(order *model.Order) error {
	if order == nil {
		return errors.New("order cannot be nil")
	}

	if order.StockReserved {
		return nil
	}

	now := time.Now()
	expireAt := now.Add(30 * time.Minute)

	return database.DB.Transaction(func(tx *gorm.DB) error {
		for _, item := range order.Items {
			reservation := &model.StockReservation{
				StoreID:    order.StoreID,
				OrderID:    order.ID,
				SKUID:      item.SKUID,
				ProductID:  item.ProductID,
				Quantity:   item.Quantity,
				ReservedAt: now,
				ExpireAt:   expireAt,
				IsReleased: false,
			}
			if err := tx.Create(reservation).Error; err != nil {
				return fmt.Errorf("create stock reservation failed: %w", err)
			}
		}

		if err := tx.Model(&model.Order{}).
			Where("id = ?", order.ID).
			Update("stock_reserved", true).Error; err != nil {
			return fmt.Errorf("update order stock reserved status failed: %w", err)
		}

		return nil
	})
}

func (s *TimeSlotPricingService) ReleaseStock(orderID uint) error {
	if orderID == 0 {
		return errors.New("order ID cannot be 0")
	}

	err := s.stockResRepo.ReleaseByOrderID(orderID)
	if err != nil {
		return fmt.Errorf("release stock failed: %w", err)
	}

	return s.orderRepo.db.Model(&model.Order{}).
		Where("id = ?", orderID).
		Update("stock_reserved", false).Error
}

func (s *TimeSlotPricingService) CleanExpiredReservations() (int64, error) {
	now := time.Now()
	count, err := s.stockResRepo.CleanExpiredReservations(now)
	if err != nil {
		return 0, fmt.Errorf("clean expired reservations failed: %w", err)
	}
	return count, nil
}

func (s *TimeSlotPricingService) CreateReminder(order *model.Order) error {
	if order == nil {
		return errors.New("order cannot be nil")
	}

	if !order.IsReservation || order.ReservationTime == nil {
		return errors.New("order is not a reservation")
	}

	remindAt := order.ReservationTime.Add(-1 * time.Hour)
	if remindAt.Before(time.Now()) {
		remindAt = time.Now().Add(5 * time.Minute)
	}

	reminder := &model.ReservationReminder{
		StoreID:  order.StoreID,
		OrderID:  order.ID,
		RemindAt: remindAt,
		IsSent:   false,
	}

	return s.reminderRepo.Create(reminder)
}

func (s *TimeSlotPricingService) ProcessPendingReminders() (int, error) {
	now := time.Now()
	reminders, err := s.reminderRepo.GetPendingReminders(now)
	if err != nil {
		return 0, fmt.Errorf("get pending reminders failed: %w", err)
	}

	sentCount := 0
	for _, reminder := range reminders {
		if err := s.SendReservationReminder(&reminder); err != nil {
			continue
		}
		if err := s.reminderRepo.MarkAsSent(reminder.ID, now); err == nil {
			sentCount++
		}
	}

	return sentCount, nil
}

func (s *TimeSlotPricingService) SendReservationReminder(reminder *model.ReservationReminder) error {
	if reminder == nil {
		return errors.New("reminder cannot be nil")
	}

	order, err := s.orderRepo.GetByID(reminder.OrderID)
	if err != nil {
		return fmt.Errorf("get order failed: %w", err)
	}

	fmt.Printf("Sending reservation reminder: order_no=%s, member_id=%d\n", order.OrderNo, order.MemberID)

	return nil
}

func (s *TimeSlotPricingService) convertToTimeSlotPricingResponse(p *model.TimeSlotPricing) *dto.TimeSlotPricingResponse {
	applicableIDs := s.parseIDs(p.ApplicableIDs)
	weekdays := s.parseWeekdays(p.Weekdays)

	storeName := ""
	if p.Store.Name != "" {
		storeName = p.Store.Name
	}

	return &dto.TimeSlotPricingResponse{
		ID:                 p.ID,
		StoreID:            p.StoreID,
		StoreName:          storeName,
		Name:               p.Name,
		StartTime:          p.StartTime,
		EndTime:            p.EndTime,
		Price:              decimal.Zero,
		OriginalPrice:      decimal.Zero,
		DiscountRate:       p.DiscountRate,
		ApplicableType:     p.ApplicableType,
		ApplicableIDs:      applicableIDs,
		Weekdays:           weekdays,
		MaxReservations:    0,
		CurrentReservations: 0,
		Status:             p.Status,
		Description:        p.Description,
		CreatedAt:          p.CreatedAt,
	}
}

func (s *TimeSlotPricingService) parseIDs(idsStr string) []uint {
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

func (s *TimeSlotPricingService) parseWeekdays(weekdaysStr string) []int {
	if weekdaysStr == "" {
		return []int{1, 2, 3, 4, 5, 6, 7}
	}
	parts := strings.Split(weekdaysStr, ",")
	var days []int
	for _, part := range parts {
		if day, err := strconv.Atoi(strings.TrimSpace(part)); err == nil {
			days = append(days, day)
		}
	}
	return days
}

func (s *TimeSlotPricingService) isSKUApplicable(timeSlot *model.TimeSlotPricing, productID uint) bool {
	if timeSlot.ApplicableType == "all" {
		return true
	}

	applicableIDs := s.parseIDs(timeSlot.ApplicableIDs)
	if len(applicableIDs) == 0 {
		return true
	}

	for _, id := range applicableIDs {
		if id == productID {
			return true
		}
	}

	return false
}

func (s *TimeSlotPricingService) calculateTimeSlotDiscount(timeSlot *model.TimeSlotPricing, amount decimal.Decimal) decimal.Decimal {
	if timeSlot.DiscountRate.LessThanOrEqual(decimal.Zero) || timeSlot.DiscountRate.GreaterThanOrEqual(decimal.NewFromInt(100)) {
		return decimal.Zero
	}

	discount := amount.Mul(decimal.NewFromInt(100).Sub(timeSlot.DiscountRate)).Div(decimal.NewFromInt(100))

	if !timeSlot.MinAmount.IsZero() && amount.LessThan(timeSlot.MinAmount) {
		return decimal.Zero
	}

	if !timeSlot.DiscountAmount.IsZero() && discount.GreaterThan(timeSlot.DiscountAmount) {
		return timeSlot.DiscountAmount
	}

	return discount
}
