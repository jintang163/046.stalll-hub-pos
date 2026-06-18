package handler

import (
	"net/http"
	"strconv"
	"stalll-hub-pos/backend/internal/dto"
	"stalll-hub-pos/backend/internal/middleware"
	"stalll-hub-pos/backend/internal/service"

	"github.com/gin-gonic/gin"
)

type ReportHandler struct {
	reportService *service.ReportService
}

func NewReportHandler() *ReportHandler {
	return &ReportHandler{
		reportService: service.NewReportService(),
	}
}

func (h *ReportHandler) GetOverview(c *gin.Context) {
	storeID := middleware.GetStoreID(c)

	overview, err := h.reportService.GetOverview(storeID)
	if err != nil {
		middleware.Error(c, "获取概览数据失败: "+err.Error())
		return
	}

	middleware.Success(c, overview)
}

func (h *ReportHandler) GetDailyReports(c *gin.Context) {
	var query dto.DailyReportQueryDTO
	if err := c.ShouldBindQuery(&query); err != nil {
		middleware.Error(c, "参数错误: "+err.Error())
		return
	}

	if query.StoreID == 0 {
		query.StoreID = middleware.GetStoreID(c)
	}

	reports, err := h.reportService.GetDailyReports(&query)
	if err != nil {
		middleware.Error(c, "获取日报表失败: "+err.Error())
		return
	}

	middleware.Success(c, reports)
}

func (h *ReportHandler) GenerateDailyReport(c *gin.Context) {
	date := c.Query("date")
	if date == "" {
		middleware.Error(c, "日期参数不能为空")
		return
	}

	storeIDStr := c.Query("store_id")
	storeID, _ := strconv.ParseUint(storeIDStr, 10, 32)
	if storeID == 0 {
		storeID = uint(middleware.GetStoreID(c))
	}

	report, err := h.reportService.GenerateDailyReport(uint(storeID), date)
	if err != nil {
		middleware.Error(c, "生成日报表失败: "+err.Error())
		return
	}

	middleware.Success(c, report)
}

func (h *ReportHandler) GetProductSalesReport(c *gin.Context) {
	var query dto.ProductSalesReportDTO
	if err := c.ShouldBindQuery(&query); err != nil {
		middleware.Error(c, "参数错误: "+err.Error())
		return
	}

	if query.StoreID == 0 {
		query.StoreID = middleware.GetStoreID(c)
	}

	report, err := h.reportService.GetProductSalesReport(&query)
	if err != nil {
		middleware.Error(c, "获取商品销售报表失败: "+err.Error())
		return
	}

	middleware.Success(c, report)
}

func (h *ReportHandler) GetCategorySalesReport(c *gin.Context) {
	var query dto.ProductSalesReportDTO
	if err := c.ShouldBindQuery(&query); err != nil {
		middleware.Error(c, "参数错误: "+err.Error())
		return
	}

	if query.StoreID == 0 {
		query.StoreID = middleware.GetStoreID(c)
	}

	report, err := h.reportService.GetCategorySalesReport(&query)
	if err != nil {
		middleware.Error(c, "获取分类销售报表失败: "+err.Error())
		return
	}

	middleware.Success(c, report)
}

func (h *ReportHandler) GetHourlySalesReport(c *gin.Context) {
	var query dto.ProductSalesReportDTO
	if err := c.ShouldBindQuery(&query); err != nil {
		middleware.Error(c, "参数错误: "+err.Error())
		return
	}

	if query.StoreID == 0 {
		query.StoreID = middleware.GetStoreID(c)
	}

	report, err := h.reportService.GetHourlySalesReport(&query)
	if err != nil {
		middleware.Error(c, "获取时段销售报表失败: "+err.Error())
		return
	}

	middleware.Success(c, report)
}

func (h *ReportHandler) GetPaymentReport(c *gin.Context) {
	var query dto.PaymentReportDTO
	if err := c.ShouldBindQuery(&query); err != nil {
		middleware.Error(c, "参数错误: "+err.Error())
		return
	}

	if query.StoreID == 0 {
		query.StoreID = middleware.GetStoreID(c)
	}

	report, err := h.reportService.GetPaymentReport(&query)
	if err != nil {
		middleware.Error(c, "获取支付方式报表失败: "+err.Error())
		return
	}

	middleware.Success(c, report)
}

func (h *ReportHandler) ExportReport(c *gin.Context) {
	var req dto.ExportReportDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.Error(c, "参数错误: "+err.Error())
		return
	}

	if req.StoreID == 0 {
		req.StoreID = middleware.GetStoreID(c)
	}

	result, err := h.reportService.ExportReport(&req)
	if err != nil {
		middleware.Error(c, "导出报表失败: "+err.Error())
		return
	}

	middleware.Success(c, result)
}
