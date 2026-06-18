package service

import (
	"fmt"
	"time"

	"github.com/shopspring/decimal"
	"stalll-hub-pos/backend/internal/dto"
	"stalll-hub-pos/backend/internal/model"
	"stalll-hub-pos/backend/internal/repository"
)

type ReportService struct {
	reportRepo  *repository.ReportRepository
	orderRepo   *repository.OrderRepository
	productRepo *repository.ProductRepository
	couponRepo  *repository.CouponRepository
	memberRepo  *repository.MemberRepository
}

func NewReportService() *ReportService {
	return &ReportService{
		reportRepo:  repository.NewReportRepository(nil),
		orderRepo:   repository.NewOrderRepository(nil),
		productRepo: repository.NewProductRepository(),
		couponRepo:  repository.NewCouponRepository(nil),
		memberRepo:  repository.NewMemberRepository(nil),
	}
}

func (s *ReportService) GetDailyReport(query *dto.DailyReportQueryDTO) ([]dto.DailyReportResponse, error) {
	reports, err := s.reportRepo.GetDailyReport(query.StoreID, query.StartDate, query.EndDate)
	if err != nil {
		return nil, err
	}

	var response []dto.DailyReportResponse
	for _, r := range reports {
		storeName := ""
		if r.Store.Name != "" {
			storeName = r.Store.Name
		}

		response = append(response, dto.DailyReportResponse{
			ID:                 r.ID,
			StoreID:            r.StoreID,
			StoreName:          storeName,
			ReportDate:         r.ReportDate,
			TotalOrders:        r.OrderCount,
			TotalAmount:        r.TotalAmount,
			PayAmount:          r.NetAmount.Add(r.RefundAmount),
			RefundAmount:       r.RefundAmount,
			NetAmount:          r.NetAmount,
			DiscountAmount:     r.DiscountAmount,
			CouponAmount:       r.CouponAmount,
			PointsUsed:         r.PointsUsed,
			PointsEarned:       r.PointsEarned,
			NewMembers:         r.NewMemberCount,
			ActiveMembers:      r.MemberCount,
			AverageOrderAmount: r.NetAmount.Div(decimal.NewFromInt(int64(r.OrderCount))),
			PeakHourOrders:     0,
			PeakHour:           "",
			CanceledOrders:     0,
			RefundedOrders:     0,
			CreatedAt:          r.CreatedAt,
		})
	}

	return response, nil
}

func (s *ReportService) GenerateDailyReport(storeID uint, reportDate string) (*model.DailyReport, error) {
	if reportDate == "" {
		reportDate = time.Now().Format("2006-01-02")
	}
	return s.reportRepo.GenerateDailyReport(storeID, reportDate)
}

func (s *ReportService) GetProductSalesReport(query *dto.ProductSalesReportDTO) ([]dto.ProductSalesResponse, error) {
	return s.reportRepo.GetProductSales(
		query.StoreID,
		query.StartDate,
		query.EndDate,
		query.CategoryID,
		query.TopN,
	)
}

func (s *ReportService) GetCategorySalesReport(query *dto.ProductSalesReportDTO) ([]dto.CategorySalesResponse, error) {
	return s.reportRepo.GetCategorySales(
		query.StoreID,
		query.StartDate,
		query.EndDate,
	)
}

func (s *ReportService) GetHourlySalesReport(query *dto.DailyReportQueryDTO) ([]dto.HourlySalesResponse, error) {
	return s.reportRepo.GetHourlySales(
		query.StoreID,
		query.StartDate,
		query.EndDate,
	)
}

func (s *ReportService) GetPaymentReport(query *dto.PaymentReportDTO) ([]dto.PaymentSalesResponse, error) {
	return s.reportRepo.GetPaymentSales(
		query.StoreID,
		query.StartDate,
		query.EndDate,
	)
}

func (s *ReportService) GetOverview(storeID uint) (*dto.OverviewResponse, error) {
	today := time.Now().Format("2006-01-02")
	yesterday := time.Now().AddDate(0, 0, -1).Format("2006-01-02")
	monthStart := time.Now().Format("2006-01") + "-01"

	todayReport, _ := s.reportRepo.GenerateDailyReport(storeID, today)
	yesterdayReport, _ := s.reportRepo.GenerateDailyReport(storeID, yesterday)

	todayOrders := 0
	todayAmount := decimal.Zero
	todayCustomers := 0
	if todayReport != nil {
		todayOrders = todayReport.OrderCount
		todayAmount = todayReport.NetAmount
		todayCustomers = todayReport.MemberCount + todayReport.NewMemberCount
	}

	yesterdayOrders := 0
	yesterdayAmount := decimal.Zero
	yesterdayCustomers := 0
	if yesterdayReport != nil {
		yesterdayOrders = yesterdayReport.OrderCount
		yesterdayAmount = yesterdayReport.NetAmount
		yesterdayCustomers = yesterdayReport.MemberCount + yesterdayReport.NewMemberCount
	}

	monthReports, _ := s.reportRepo.GetDailyReport(storeID, monthStart, today)
	monthOrders := 0
	monthAmount := decimal.Zero
	monthCustomers := 0
	for _, r := range monthReports {
		monthOrders += r.OrderCount
		monthAmount = monthAmount.Add(r.NetAmount)
		monthCustomers += r.MemberCount + r.NewMemberCount
	}

	stockWarnings, _ := s.productRepo.GetStockWarnings(storeID)
	stockWarningCount := len(stockWarnings)

	pendingOrders := s.getPendingOrdersCount(storeID)

	activeMembers := s.getActiveMembersCount(storeID)

	availableCoupons := s.getAvailableCouponsCount(storeID)

	todayHourlySales, _ := s.reportRepo.GetHourlySales(storeID, today, today)

	topProducts, _ := s.reportRepo.GetProductSales(storeID, today, today, 0, 5)

	return &dto.OverviewResponse{
		StoreID:            storeID,
		TodayOrders:        todayOrders,
		TodayAmount:        todayAmount,
		TodayCustomers:     todayCustomers,
		YesterdayOrders:    yesterdayOrders,
		YesterdayAmount:    yesterdayAmount,
		YesterdayCustomers: yesterdayCustomers,
		MonthOrders:        monthOrders,
		MonthAmount:        monthAmount,
		MonthCustomers:     monthCustomers,
		StockWarningCount:  stockWarningCount,
		PendingOrders:      pendingOrders,
		ActiveMembers:      activeMembers,
		AvailableCoupons:   availableCoupons,
		TodayHourlySales:   todayHourlySales,
		TopProducts:        topProducts,
	}, nil
}

func (s *ReportService) getPendingOrdersCount(storeID uint) int {
	query := &dto.OrderQuery{
		StoreID:     storeID,
		OrderStatus: 1,
		Page:        1,
		PageSize:    1,
	}
	_, total, _ := s.orderRepo.List(query)
	return int(total)
}

func (s *ReportService) getActiveMembersCount(storeID uint) int {
	thirtyDaysAgo := time.Now().AddDate(0, 0, -30).Format("2006-01-02")
	members, _, _ := s.memberRepo.List(storeID, "", "", 0, 1, 1, 1000)
	count := 0
	for _, m := range members {
		if m.Status == 1 {
			count++
		}
	}
	return count
}

func (s *ReportService) getAvailableCouponsCount(storeID uint) int {
	query := &dto.CouponQueryDTO{
		StoreID: storeID,
		Status:  1,
		Page:    1,
		PageSize: 1000,
	}
	_, total, _ := s.couponRepo.List("", "", 1, 1, 1000)
	return int(total)
}

func (s *ReportService) ExportReport(req *dto.ExportReportDTO) (*dto.ExportResponse, error) {
	var fileName string
	var downloadURL string

	switch req.ReportType {
	case "daily":
		fileName = fmt.Sprintf("daily_report_%s_%s.%s", req.StartDate, req.EndDate, req.Format)
	case "product":
		fileName = fmt.Sprintf("product_sales_%s_%s.%s", req.StartDate, req.EndDate, req.Format)
	case "category":
		fileName = fmt.Sprintf("category_sales_%s_%s.%s", req.StartDate, req.EndDate, req.Format)
	case "payment":
		fileName = fmt.Sprintf("payment_report_%s_%s.%s", req.StartDate, req.EndDate, req.Format)
	default:
		fileName = fmt.Sprintf("report_%s.%s", time.Now().Format("20060102150405"), req.Format)
	}

	downloadURL = fmt.Sprintf("/api/reports/export/%s", fileName)

	return &dto.ExportResponse{
		DownloadURL: downloadURL,
		FileName:    fileName,
	}, nil
}
