package handler

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"stalll-hub-pos/backend/internal/dto"
	"stalll-hub-pos/backend/internal/middleware"
	"stalll-hub-pos/backend/internal/service"
	"time"

	"github.com/gin-gonic/gin"
)

type AnalyticsHandler struct {
	chSyncService *service.ClickHouseSyncService
	costService   *service.CostService
}

func NewAnalyticsHandler() *AnalyticsHandler {
	return &AnalyticsHandler{
		chSyncService: service.NewClickHouseSyncService(),
		costService:   service.NewCostService(),
	}
}

func (h *AnalyticsHandler) GetRevenueReport(c *gin.Context) {
	var query dto.RevenueReportQueryDTO
	if err := c.ShouldBindQuery(&query); err != nil {
		middleware.Error(c, "参数错误: "+err.Error())
		return
	}

	reportType := query.ReportType
	if reportType == "" {
		reportType = "daily"
	}

	reports, err := h.chSyncService.GetRevenueReport(query.StoreID, query.StartDate, query.EndDate, reportType)
	if err != nil {
		middleware.Error(c, "获取营业报表失败: "+err.Error())
		return
	}

	middleware.Success(c, reports)
}

func (h *AnalyticsHandler) GetHourlyTrend(c *gin.Context) {
	var query dto.RevenueReportQueryDTO
	if err := c.ShouldBindQuery(&query); err != nil {
		middleware.Error(c, "参数错误: "+err.Error())
		return
	}

	trends, err := h.chSyncService.GetHourlyTrend(query.StoreID, query.StartDate, query.EndDate)
	if err != nil {
		middleware.Error(c, "获取时段趋势失败: "+err.Error())
		return
	}

	middleware.Success(c, trends)
}

func (h *AnalyticsHandler) GetTopProducts(c *gin.Context) {
	var query dto.RevenueReportQueryDTO
	if err := c.ShouldBindQuery(&query); err != nil {
		middleware.Error(c, "参数错误: "+err.Error())
		return
	}

	topN := 10
	if n := c.Query("top_n"); n != "" {
		if v, err := strconv.Atoi(n); err == nil && v > 0 {
			topN = v
		}
	}

	products, err := h.chSyncService.GetTopProducts(query.StoreID, query.StartDate, query.EndDate, topN)
	if err != nil {
		middleware.Error(c, "获取热门菜品失败: "+err.Error())
		return
	}

	middleware.Success(c, products)
}

func (h *AnalyticsHandler) ImportCostExcel(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		middleware.Error(c, "请上传文件: "+err.Error())
		return
	}

	effectiveDate := c.PostForm("effective_date")
	if effectiveDate == "" {
		effectiveDate = time.Now().Format("2006-01-02")
	}

	uploadDir := "./uploads/cost"
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		middleware.Error(c, "创建上传目录失败")
		return
	}

	fileName := fmt.Sprintf("%d_%s", time.Now().UnixNano(), filepath.Base(file.Filename))
	filePath := filepath.Join(uploadDir, fileName)

	if err := c.SaveUploadedFile(file, filePath); err != nil {
		middleware.Error(c, "保存文件失败: "+err.Error())
		return
	}

	batch, err := h.costService.ImportCostExcel(filePath, effectiveDate)
	if err != nil {
		middleware.Error(c, "导入成本数据失败: "+err.Error())
		return
	}

	middleware.Success(c, batch)
}

func (h *AnalyticsHandler) GetCostList(c *gin.Context) {
	var query dto.CostQueryDTO
	if err := c.ShouldBindQuery(&query); err != nil {
		middleware.Error(c, "参数错误: "+err.Error())
		return
	}

	if query.Page <= 0 {
		query.Page = 1
	}
	if query.PageSize <= 0 {
		query.PageSize = 20
	}

	costs, total, err := h.costService.GetCostList(&query)
	if err != nil {
		middleware.Error(c, "获取成本列表失败: "+err.Error())
		return
	}

	middleware.PageSuccess(c, costs, total, query.Page, query.PageSize)
}

func (h *AnalyticsHandler) GetProfitReport(c *gin.Context) {
	var query dto.ProfitReportQueryDTO
	if err := c.ShouldBindQuery(&query); err != nil {
		middleware.Error(c, "参数错误: "+err.Error())
		return
	}

	report, err := h.costService.GetProfitReport(&query)
	if err != nil {
		middleware.Error(c, "获取利润报表失败: "+err.Error())
		return
	}

	middleware.Success(c, report)
}

func (h *AnalyticsHandler) GetProfitSummary(c *gin.Context) {
	var query dto.ProfitReportQueryDTO
	if err := c.ShouldBindQuery(&query); err != nil {
		middleware.Error(c, "参数错误: "+err.Error())
		return
	}

	summary, err := h.costService.GetProfitSummary(&query)
	if err != nil {
		middleware.Error(c, "获取利润概览失败: "+err.Error())
		return
	}

	middleware.Success(c, summary)
}

func (h *AnalyticsHandler) TriggerFullBackfill(c *gin.Context) {
	go func() {
		if _, _, err := h.chSyncService.FullBackfill(); err != nil {
			fmt.Printf("Backfill error: %v\n", err)
		}
	}()
	middleware.Success(c, gin.H{"message": "全量回填任务已启动"})
}

func (h *AnalyticsHandler) GetSyncStatus(c *gin.Context) {
	middleware.Success(c, gin.H{"status": "running"})
}

func (h *AnalyticsHandler) GetProfitReportV2(c *gin.Context) {
	var query dto.ProfitReportQueryDTO
	if err := c.ShouldBindQuery(&query); err != nil {
		middleware.Error(c, "参数错误: "+err.Error())
		return
	}

	report, err := h.costService.GetProfitReportV2(&query)
	if err != nil {
		middleware.Error(c, "获取利润报表失败: "+err.Error())
		return
	}

	middleware.Success(c, report)
}

func (h *AnalyticsHandler) GetProfitSummaryV2(c *gin.Context) {
	var query dto.ProfitReportQueryDTO
	if err := c.ShouldBindQuery(&query); err != nil {
		middleware.Error(c, "参数错误: "+err.Error())
		return
	}

	summary, err := h.costService.GetProfitSummaryV2(&query)
	if err != nil {
		middleware.Error(c, "获取利润概览失败: "+err.Error())
		return
	}

	middleware.Success(c, summary)
}
