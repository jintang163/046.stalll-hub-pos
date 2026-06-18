package repository

import (
	"fmt"
	"time"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
	"stalll-hub-pos/backend/internal/dto"
	"stalll-hub-pos/backend/internal/model"
	"stalll-hub-pos/backend/pkg/database"
)

type ReportRepository struct {
	db *gorm.DB
}

func NewReportRepository(db *gorm.DB) *ReportRepository {
	return &ReportRepository{db: db}
}

func (r *ReportRepository) GetDailyReport(storeID uint, startDate, endDate string) ([]model.DailyReport, error) {
	var reports []model.DailyReport
	db := database.DB.Model(&model.DailyReport{})
	if storeID > 0 {
		db = db.Where("store_id = ?", storeID)
	}
	if startDate != "" {
		db = db.Where("report_date >= ?", startDate)
	}
	if endDate != "" {
		db = db.Where("report_date <= ?", endDate)
	}
	err := db.Preload("Store").Order("report_date DESC").Find(&reports).Error
	return reports, err
}

func (r *ReportRepository) GenerateDailyReport(storeID uint, reportDate string) (*model.DailyReport, error) {
	nextDate := time.Now().AddDate(0, 0, 1).Format("2006-01-02")

	var totalOrders int64
	var totalAmount, payAmount, refundAmount decimal.Decimal
	var canceledOrders, refundedOrders int64

	database.DB.Model(&model.Order{}).
		Where("store_id = ? AND created_at >= ? AND created_at < ?", storeID, reportDate, nextDate).
		Where("order_status != ?", -1).
		Count(&totalOrders).
		Select("COALESCE(SUM(total_amount), 0), COALESCE(SUM(pay_amount), 0)").
		Row().Scan(&totalAmount, &payAmount)

	database.DB.Model(&model.Order{}).
		Where("store_id = ? AND created_at >= ? AND created_at < ?", storeID, reportDate, nextDate).
		Where("order_status = ?", -1).
		Count(&canceledOrders)

	database.DB.Model(&model.OrderRefund{}).
		Where("store_id = ? AND created_at >= ? AND created_at < ?", storeID, reportDate, nextDate).
		Where("refund_status = 1").
		Select("COALESCE(SUM(refund_amount), 0), COUNT(*)").
		Row().Scan(&refundAmount, &refundedOrders)

	netAmount := payAmount.Sub(refundAmount)

	var discountAmount, couponAmount decimal.Decimal
	var pointsUsed, pointsEarned int64

	database.DB.Model(&model.Order{}).
		Where("store_id = ? AND created_at >= ? AND created_at < ?", storeID, reportDate, nextDate).
		Where("order_status != ?", -1).
		Select("COALESCE(SUM(discount_amount), 0), COALESCE(SUM(coupon_amount), 0), COALESCE(SUM(points_used), 0), COALESCE(SUM(points_earned), 0)").
		Row().Scan(&discountAmount, &couponAmount, &pointsUsed, &pointsEarned)

	var newMembers int64
	database.DB.Model(&model.Member{}).
		Where("store_id = ? AND created_at >= ? AND created_at < ?", storeID, reportDate, nextDate).
		Count(&newMembers)

	var activeMembers int64
	database.DB.Model(&model.Order{}).
		Where("store_id = ? AND created_at >= ? AND created_at < ?", storeID, reportDate, nextDate).
		Where("member_id > 0").
		Distinct("member_id").
		Count(&activeMembers)

	var avgOrderAmount decimal.Decimal
	if totalOrders > 0 {
		avgOrderAmount = payAmount.Div(decimal.NewFromInt(totalOrders))
	}

	peakHour, peakHourOrders := r.getPeakHour(storeID, reportDate, nextDate)

	report := &model.DailyReport{
		StoreID:            storeID,
		ReportDate:         reportDate,
		TotalOrders:        int(totalOrders),
		TotalAmount:        totalAmount,
		PayAmount:          payAmount,
		RefundAmount:       refundAmount,
		NetAmount:          netAmount,
		DiscountAmount:     discountAmount,
		CouponAmount:       couponAmount,
		PointsUsed:         int(pointsUsed),
		PointsEarned:       int(pointsEarned),
		NewMembers:         int(newMembers),
		ActiveMembers:      int(activeMembers),
		AverageOrderAmount: avgOrderAmount,
		PeakHourOrders:     peakHourOrders,
		PeakHour:           peakHour,
		CanceledOrders:     int(canceledOrders),
		RefundedOrders:     int(refundedOrders),
	}

	var existingReport model.DailyReport
	err := database.DB.Where("store_id = ? AND report_date = ?", storeID, reportDate).
		First(&existingReport).Error

	if err == nil {
		report.ID = existingReport.ID
		err = database.DB.Save(report).Error
	} else {
		err = database.DB.Create(report).Error
	}

	return report, err
}

func (r *ReportRepository) getPeakHour(storeID uint, startDate, nextDate string) (string, int) {
	type Result struct {
		Hour       int
		OrderCount int
	}
	var results []Result

	database.DB.Model(&model.Order{}).
		Select("HOUR(created_at) as hour, COUNT(*) as order_count").
		Where("store_id = ? AND created_at >= ? AND created_at < ?", storeID, startDate, nextDate).
		Where("order_status != ?", -1).
		Group("HOUR(created_at)").
		Order("order_count DESC").
		Limit(1).
		Scan(&results)

	if len(results) > 0 {
		return fmt.Sprintf("%02d:00", results[0].Hour), results[0].OrderCount
	}
	return "00:00", 0
}

func (r *ReportRepository) GetProductSales(storeID uint, startDate, endDate string, categoryID uint, topN int) ([]dto.ProductSalesResponse, error) {
	var results []struct {
		ProductID      uint
		ProductName    string
		ProductImage   string
		CategoryID     uint
		CategoryName   string
		TotalQuantity  int
		TotalAmount    decimal.Decimal
		RefundQuantity int
		RefundAmount   decimal.Decimal
	}

	db := database.DB.Table("order_items oi").
		Select(`oi.product_id, 
				oi.product_name, 
				COALESCE(p.image, '') as product_image,
				COALESCE(p.category_id, 0) as category_id,
				COALESCE(c.name, '') as category_name,
				SUM(oi.quantity) as total_quantity,
				SUM(oi.subtotal) as total_amount,
				SUM(CASE WHEN oi.status = 0 THEN oi.quantity ELSE 0 END) as refund_quantity,
				SUM(CASE WHEN oi.status = 0 THEN oi.subtotal ELSE 0 END) as refund_amount`).
		Joins("LEFT JOIN products p ON oi.product_id = p.id").
		Joins("LEFT JOIN categories c ON p.category_id = c.id").
		Joins("LEFT JOIN orders o ON oi.order_id = o.id").
		Where("o.store_id = ? AND o.created_at >= ? AND o.created_at <= ?", storeID, startDate, endDate+" 23:59:59").
		Where("o.order_status != ?", -1)

	if categoryID > 0 {
		db = db.Where("p.category_id = ?", categoryID)
	}

	db = db.Group("oi.product_id, oi.product_name, p.image, p.category_id, c.name").
		Order("total_amount DESC")

	if topN > 0 {
		db = db.Limit(topN)
	}

	err := db.Scan(&results).Error
	if err != nil {
		return nil, err
	}

	var response []dto.ProductSalesResponse
	for i, r := range results {
		response = append(response, dto.ProductSalesResponse{
			ProductID:      r.ProductID,
			ProductName:    r.ProductName,
			ProductImage:   r.ProductImage,
			CategoryID:     r.CategoryID,
			CategoryName:   r.CategoryName,
			TotalQuantity:  r.TotalQuantity,
			TotalAmount:    r.TotalAmount,
			RefundQuantity: r.RefundQuantity,
			RefundAmount:   r.RefundAmount,
			NetAmount:      r.TotalAmount.Sub(r.RefundAmount),
			Rank:           i + 1,
		})
	}

	return response, nil
}

func (r *ReportRepository) GetCategorySales(storeID uint, startDate, endDate string) ([]dto.CategorySalesResponse, error) {
	var results []struct {
		CategoryID    uint
		CategoryName  string
		ProductCount  int
		TotalQuantity int
		TotalAmount   decimal.Decimal
	}

	database.DB.Table("order_items oi").
		Select(`COALESCE(p.category_id, 0) as category_id,
				COALESCE(c.name, '未分类') as category_name,
				COUNT(DISTINCT oi.product_id) as product_count,
				SUM(oi.quantity) as total_quantity,
				SUM(oi.subtotal) as total_amount`).
		Joins("LEFT JOIN products p ON oi.product_id = p.id").
		Joins("LEFT JOIN categories c ON p.category_id = c.id").
		Joins("LEFT JOIN orders o ON oi.order_id = o.id").
		Where("o.store_id = ? AND o.created_at >= ? AND o.created_at <= ?", storeID, startDate, endDate+" 23:59:59").
		Where("o.order_status != ? AND oi.status = 1", -1).
		Group("p.category_id, c.name").
		Order("total_amount DESC").
		Scan(&results)

	var totalAmount decimal.Decimal
	for _, r := range results {
		totalAmount = totalAmount.Add(r.TotalAmount)
	}

	var response []dto.CategorySalesResponse
	for _, r := range results {
		percentage := decimal.Zero
		if totalAmount.GreaterThan(decimal.Zero) {
			percentage = r.TotalAmount.Div(totalAmount).Mul(decimal.NewFromInt(100))
		}
		response = append(response, dto.CategorySalesResponse{
			CategoryID:    r.CategoryID,
			CategoryName:  r.CategoryName,
			ProductCount:  r.ProductCount,
			TotalQuantity: r.TotalQuantity,
			TotalAmount:   r.TotalAmount,
			Percentage:    percentage,
		})
	}

	return response, nil
}

func (r *ReportRepository) GetHourlySales(storeID uint, startDate, endDate string) ([]dto.HourlySalesResponse, error) {
	var results []struct {
		Hour        int
		TotalOrders int
		TotalAmount decimal.Decimal
	}

	database.DB.Table("orders o").
		Select(`HOUR(o.created_at) as hour,
				COUNT(*) as total_orders,
				SUM(o.pay_amount) as total_amount`).
		Where("o.store_id = ? AND o.created_at >= ? AND o.created_at <= ?", storeID, startDate, endDate+" 23:59:59").
		Where("o.order_status != ? AND o.pay_status = 1", -1).
		Group("HOUR(o.created_at)").
		Order("hour ASC").
		Scan(&results)

	var totalAmount decimal.Decimal
	for _, r := range results {
		totalAmount = totalAmount.Add(r.TotalAmount)
	}

	var response []dto.HourlySalesResponse
	for _, r := range results {
		percentage := decimal.Zero
		if totalAmount.GreaterThan(decimal.Zero) {
			percentage = r.TotalAmount.Div(totalAmount).Mul(decimal.NewFromInt(100))
		}
		response = append(response, dto.HourlySalesResponse{
			Hour:        r.Hour,
			TotalOrders: r.TotalOrders,
			TotalAmount: r.TotalAmount,
			Percentage:  percentage,
		})
	}

	return response, nil
}

func (r *ReportRepository) GetPaymentSales(storeID uint, startDate, endDate string) ([]dto.PaymentSalesResponse, error) {
	paymentNames := map[string]string{
		"wechat":  "微信支付",
		"alipay":  "支付宝",
		"cash":    "现金",
		"card":    "银行卡",
		"balance": "余额",
	}

	var results []struct {
		PayMethod   string
		TotalOrders int
		TotalAmount decimal.Decimal
	}

	database.DB.Table("order_payments op").
		Select(`op.pay_method,
				COUNT(*) as total_orders,
				SUM(op.amount) as total_amount`).
		Joins("LEFT JOIN orders o ON op.order_id = o.id").
		Where("o.store_id = ? AND op.created_at >= ? AND op.created_at <= ?", storeID, startDate, endDate+" 23:59:59").
		Where("op.pay_status = 1").
		Group("op.pay_method").
		Order("total_amount DESC").
		Scan(&results)

	var totalAmount decimal.Decimal
	for _, r := range results {
		totalAmount = totalAmount.Add(r.TotalAmount)
	}

	var response []dto.PaymentSalesResponse
	for _, r := range results {
		percentage := decimal.Zero
		averageAmount := decimal.Zero
		if totalAmount.GreaterThan(decimal.Zero) {
			percentage = r.TotalAmount.Div(totalAmount).Mul(decimal.NewFromInt(100))
		}
		if r.TotalOrders > 0 {
			averageAmount = r.TotalAmount.Div(decimal.NewFromInt(int64(r.TotalOrders)))
		}
		response = append(response, dto.PaymentSalesResponse{
			PayMethod:     r.PayMethod,
			PayMethodName: paymentNames[r.PayMethod],
			TotalOrders:   r.TotalOrders,
			TotalAmount:   r.TotalAmount,
			Percentage:    percentage,
			AverageAmount: averageAmount,
		})
	}

	return response, nil
}
