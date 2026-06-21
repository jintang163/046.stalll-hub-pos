package service

import (
	"crypto/rand"
	"errors"
	"fmt"
	"log"
	"math/big"
	"strconv"
	"time"

	"github.com/shopspring/decimal"
	"stalll-hub-pos/backend/internal/dto"
	"stalll-hub-pos/backend/internal/model"
	"stalll-hub-pos/backend/internal/repository"
	"stalll-hub-pos/backend/pkg/database"
)

const (
	badReviewRating     = 3.0
	alertThresholdDrop  = 0.2
	workOrderDueHours   = 24
	workOrderDefaultType = "bad_review"
)

type ReviewService struct {
	reviewRepo      *repository.ReviewRepository
	dingtalkService *DingTalkService
}

func NewReviewService() *ReviewService {
	return &ReviewService{
		reviewRepo:      repository.NewReviewRepository(database.DB),
		dingtalkService: NewDingTalkService(),
	}
}

func (s *ReviewService) SaveAuth(req *dto.PlatformAuthDTO) error {
	auth := &model.StorePlatformAuth{
		StoreID:      req.StoreID,
		Platform:     req.Platform,
		StoreUrl:     req.StoreUrl,
		ShopID:       req.ShopID,
		AuthToken:    req.AuthToken,
		RefreshToken: req.RefreshToken,
		Status:       1,
		SyncStatus:   "pending",
	}
	return s.reviewRepo.UpsertAuth(auth)
}

func (s *ReviewService) GetAuth(storeID uint, platform string) (*dto.PlatformAuthResponse, error) {
	auth, err := s.reviewRepo.GetAuth(storeID, platform)
	if err != nil {
		return nil, err
	}
	return s.convertToAuthResponse(auth), nil
}

func (s *ReviewService) ListAuths(storeID uint) ([]dto.PlatformAuthResponse, error) {
	auths, err := s.reviewRepo.ListAuths(storeID)
	if err != nil {
		return nil, err
	}
	var result []dto.PlatformAuthResponse
	for _, auth := range auths {
		result = append(result, *s.convertToAuthResponse(&auth))
	}
	return result, nil
}

func (s *ReviewService) convertToAuthResponse(auth *model.StorePlatformAuth) *dto.PlatformAuthResponse {
	return &dto.PlatformAuthResponse{
		ID:           auth.ID,
		StoreID:      auth.StoreID,
		Platform:     auth.Platform,
		StoreUrl:     auth.StoreUrl,
		AuthToken:    auth.AuthToken,
		RefreshToken: auth.RefreshToken,
		ExpireTime:   auth.ExpireTime,
		ShopID:       auth.ShopID,
		Status:       auth.Status,
		LastSyncTime: auth.LastSyncTime,
		SyncStatus:   auth.SyncStatus,
		SyncError:    auth.SyncError,
		CreatedAt:    auth.CreatedAt,
		UpdatedAt:    auth.UpdatedAt,
	}
}

func (s *ReviewService) SyncStoreReviews(storeID uint, platform string) (*dto.SyncStatusResponse, error) {
	auth, err := s.reviewRepo.GetAuth(storeID, platform)
	if err != nil {
		return nil, fmt.Errorf("auth not found: %w", err)
	}
	if auth.Status != 1 {
		return nil, errors.New("auth is disabled")
	}

	now := time.Now()
	err = s.reviewRepo.UpdateSyncStatus(storeID, platform, "syncing", "", nil)
	if err != nil {
		return nil, fmt.Errorf("update sync status failed: %w", err)
	}

	rating, err := s.fetchRatingFromPlatform(storeID, platform, auth)
	if err != nil {
		_ = s.reviewRepo.UpdateSyncStatus(storeID, platform, "failed", err.Error(), nil)
		return nil, fmt.Errorf("fetch rating failed: %w", err)
	}
	err = s.reviewRepo.CreateRating(rating)
	if err != nil {
		log.Printf("create rating failed: %v", err)
	}

	reviews, err := s.fetchReviewsFromPlatform(storeID, platform, auth)
	if err != nil {
		_ = s.reviewRepo.UpdateSyncStatus(storeID, platform, "failed", err.Error(), nil)
		return nil, fmt.Errorf("fetch reviews failed: %w", err)
	}
	for _, review := range reviews {
		err := s.reviewRepo.UpsertReview(&review)
		if err != nil {
			log.Printf("upsert review failed: %v", err)
		}
	}

	err = s.reviewRepo.UpdateSyncStatus(storeID, platform, "success", "", &now)
	if err != nil {
		log.Printf("update sync status failed: %v", err)
	}

	return &dto.SyncStatusResponse{
		StoreID:      storeID,
		Platform:     platform,
		SyncStatus:   "success",
		SyncError:    "",
		LastSyncTime: &now,
	}, nil
}

func (s *ReviewService) fetchRatingFromPlatform(storeID uint, platform string, auth *model.StorePlatformAuth) (*model.PlatformReviewRating, error) {
	ratingDate := time.Now().Format("2006-01-02")
	overall := decimal.NewFromFloat(4.5 + float64(storeID%10)/10.0)
	if overall.GreaterThan(decimal.NewFromInt(5)) {
		overall = decimal.NewFromInt(5)
	}
	taste := overall.Sub(decimal.NewFromFloat(0.1))
	env := overall.Sub(decimal.NewFromFloat(0.2))
	service := overall

	reviewCount := 100 + int(storeID)
	goodCount := int(float64(reviewCount) * 0.85)
	midCount := int(float64(reviewCount) * 0.1)
	badCount := reviewCount - goodCount - midCount
	goodRate := decimal.NewFromInt(int64(goodCount)).Mul(decimal.NewFromInt(100)).Div(decimal.NewFromInt(int64(reviewCount)))

	return &model.PlatformReviewRating{
		StoreID:           storeID,
		Platform:          platform,
		OverallRating:     overall,
		TasteRating:       taste,
		EnvironmentRating: env,
		ServiceRating:     service,
		ReviewCount:       reviewCount,
		GoodReviewCount:   goodCount,
		MidReviewCount:    midCount,
		BadReviewCount:    badCount,
		GoodReviewRate:    goodRate,
		RatingDate:        ratingDate,
		SnapshotTime:      time.Now(),
	}, nil
}

func (s *ReviewService) fetchReviewsFromPlatform(storeID uint, platform string, auth *model.StorePlatformAuth) ([]model.PlatformReview, error) {
	var reviews []model.PlatformReview
	for i := 1; i <= 5; i++ {
		rating := decimal.NewFromFloat(float64(6-i) - 0.1)
		if rating.LessThan(decimal.NewFromFloat(1.0)) {
			rating = decimal.NewFromFloat(1.0)
		}
		reviewTime := time.Now().Add(-time.Duration(i) * 24 * time.Hour)
		isBad := rating.LessThanOrEqual(decimal.NewFromFloat(badReviewRating))

		reviews = append(reviews, model.PlatformReview{
			StoreID:            storeID,
			Platform:           platform,
			PlatformID:         fmt.Sprintf("%s_%d_%d", platform, storeID, reviewTime.Unix()),
			UserNickname:       fmt.Sprintf("用户%d", i),
			UserAvatar:         "",
			UserLevel:          "Vip" + strconv.Itoa(i),
			Rating:             rating,
			TasteRating:        rating,
			EnvironmentRating:  rating.Sub(decimal.NewFromFloat(0.1)),
			ServiceRating:      rating,
			Content:            fmt.Sprintf("这是第%d条模拟评价内容，口味不错，服务也很好。", i),
			Images:             "",
			ReviewTime:         reviewTime,
			ReplyContent:       "",
			ReplyTime:          nil,
			IsBadReview:        isBad,
			IsReplied:          false,
			IsWorkOrderCreated: false,
			OrderNo:            "",
			PerCapita:          decimal.NewFromFloat(50.0 + float64(i)*10),
		})
	}
	return reviews, nil
}

func (s *ReviewService) SyncAllStores() (int, error) {
	auths, err := s.reviewRepo.ListAuths(0)
	if err != nil {
		return 0, err
	}
	successCount := 0
	for _, auth := range auths {
		if auth.Status != 1 {
			continue
		}
		_, err := s.SyncStoreReviews(auth.StoreID, auth.Platform)
		if err != nil {
			log.Printf("sync store %d platform %s failed: %v", auth.StoreID, auth.Platform, err)
			continue
		}
		successCount++
	}
	return successCount, nil
}

func (s *ReviewService) ListRatings(query *dto.ReviewRatingQueryDTO) (*dto.PageResponse, error) {
	if query.Page == 0 {
		query.Page = 1
	}
	if query.PageSize == 0 {
		query.PageSize = 10
	}
	offset := (query.Page - 1) * query.PageSize

	ratings, total, err := s.reviewRepo.ListRatings(
		query.StoreID,
		query.Platform,
		query.StartDate,
		query.EndDate,
		offset,
		query.PageSize,
	)
	if err != nil {
		return nil, err
	}

	var list []dto.ReviewRatingResponse
	for _, r := range ratings {
		list = append(list, *s.convertToRatingResponse(&r))
	}

	return &dto.PageResponse{
		List:  list,
		Total: total,
		Page:  query.Page,
		Size:  query.PageSize,
	}, nil
}

func (s *ReviewService) GetRatingTrend(query *dto.ReviewRatingQueryDTO) ([]dto.RatingTrendResponse, error) {
	ratings, err := s.reviewRepo.GetRatingTrend(
		query.StoreID,
		query.Platform,
		query.StartDate,
		query.EndDate,
	)
	if err != nil {
		return nil, err
	}

	var result []dto.RatingTrendResponse
	for _, r := range ratings {
		result = append(result, dto.RatingTrendResponse{
			Date:              r.RatingDate,
			OverallRating:     r.OverallRating,
			TasteRating:       r.TasteRating,
			EnvironmentRating: r.EnvironmentRating,
			ServiceRating:     r.ServiceRating,
			ReviewCount:       r.ReviewCount,
		})
	}
	return result, nil
}

func (s *ReviewService) AnalyzeRatingChange(storeID uint, platform string) (decimal.Decimal, decimal.Decimal, decimal.Decimal, error) {
	currRating, err := s.reviewRepo.GetLatestRating(storeID, platform)
	if err != nil {
		return decimal.Zero, decimal.Zero, decimal.Zero, fmt.Errorf("get current rating failed: %w", err)
	}

	prevRating, err := s.reviewRepo.GetPreviousDayRating(storeID, platform, currRating.RatingDate)
	if err != nil {
		return decimal.Zero, currRating.OverallRating, decimal.Zero, nil
	}

	drop := prevRating.OverallRating.Sub(currRating.OverallRating)
	if drop.LessThan(decimal.Zero) {
		drop = decimal.Zero
	}

	return prevRating.OverallRating, currRating.OverallRating, drop, nil
}

func (s *ReviewService) convertToRatingResponse(r *model.PlatformReviewRating) *dto.ReviewRatingResponse {
	storeName := ""
	if r.Store.Name != "" {
		storeName = r.Store.Name
	}
	return &dto.ReviewRatingResponse{
		ID:                r.ID,
		StoreID:           r.StoreID,
		StoreName:         storeName,
		Platform:          r.Platform,
		OverallRating:     r.OverallRating,
		TasteRating:       r.TasteRating,
		EnvironmentRating: r.EnvironmentRating,
		ServiceRating:     r.ServiceRating,
		ReviewCount:       r.ReviewCount,
		GoodReviewCount:   r.GoodReviewCount,
		MidReviewCount:    r.MidReviewCount,
		BadReviewCount:    r.BadReviewCount,
		GoodReviewRate:    r.GoodReviewRate,
		RatingDate:        r.RatingDate,
		SnapshotTime:      r.SnapshotTime,
		CreatedAt:         r.CreatedAt,
		UpdatedAt:         r.UpdatedAt,
	}
}

func (s *ReviewService) ListReviews(query *dto.ReviewQueryDTO) (*dto.PageResponse, error) {
	if query.Page == 0 {
		query.Page = 1
	}
	if query.PageSize == 0 {
		query.PageSize = 10
	}
	offset := (query.Page - 1) * query.PageSize

	var ratingMin, ratingMax *float64
	if query.RatingMin != "" {
		if v, err := strconv.ParseFloat(query.RatingMin, 64); err == nil {
			ratingMin = &v
		}
	}
	if query.RatingMax != "" {
		if v, err := strconv.ParseFloat(query.RatingMax, 64); err == nil {
			ratingMax = &v
		}
	}

	reviews, total, err := s.reviewRepo.ListReviews(
		query.StoreID,
		query.Platform,
		ratingMin,
		ratingMax,
		query.IsBadReview,
		query.IsReplied,
		query.Keyword,
		query.StartDate,
		query.EndDate,
		offset,
		query.PageSize,
	)
	if err != nil {
		return nil, err
	}

	var list []dto.ReviewResponse
	for _, r := range reviews {
		list = append(list, *s.convertToReviewResponse(&r))
	}

	return &dto.PageResponse{
		List:  list,
		Total: total,
		Page:  query.Page,
		Size:  query.PageSize,
	}, nil
}

func (s *ReviewService) GetReview(id uint) (*dto.ReviewResponse, error) {
	review, err := s.reviewRepo.GetReviewByID(id)
	if err != nil {
		return nil, err
	}
	return s.convertToReviewResponse(review), nil
}

func (s *ReviewService) ReplyReview(id uint, req *dto.ReviewReplyRequest) error {
	_, err := s.reviewRepo.GetReviewByID(id)
	if err != nil {
		return errors.New("review not found")
	}
	return s.reviewRepo.UpdateReviewReply(id, req.ReplyContent)
}

func (s *ReviewService) convertToReviewResponse(r *model.PlatformReview) *dto.ReviewResponse {
	storeName := ""
	if r.Store.Name != "" {
		storeName = r.Store.Name
	}
	return &dto.ReviewResponse{
		ID:                 r.ID,
		StoreID:            r.StoreID,
		StoreName:          storeName,
		Platform:           r.Platform,
		PlatformID:         r.PlatformID,
		UserNickname:       r.UserNickname,
		UserAvatar:         r.UserAvatar,
		UserLevel:          r.UserLevel,
		Rating:             r.Rating,
		TasteRating:        r.TasteRating,
		EnvironmentRating:  r.EnvironmentRating,
		ServiceRating:      r.ServiceRating,
		Content:            r.Content,
		Images:             r.Images,
		ReviewTime:         r.ReviewTime,
		ReplyContent:       r.ReplyContent,
		ReplyTime:          r.ReplyTime,
		IsBadReview:        r.IsBadReview,
		IsReplied:          r.IsReplied,
		IsWorkOrderCreated: r.IsWorkOrderCreated,
		OrderNo:            r.OrderNo,
		PerCapita:          r.PerCapita,
		CreatedAt:          r.CreatedAt,
		UpdatedAt:          r.UpdatedAt,
	}
}

func (s *ReviewService) AutoCreateWorkOrders() (int, error) {
	trueVal := true
	falseVal := false
	badRating := badReviewRating
	ratingMinStr := fmt.Sprintf("%.1f", 1.0)
	ratingMaxStr := fmt.Sprintf("%.1f", badRating)

	var ratingMin, ratingMax *float64
	if v, err := strconv.ParseFloat(ratingMinStr, 64); err == nil {
		ratingMin = &v
	}
	if v, err := strconv.ParseFloat(ratingMaxStr, 64); err == nil {
		ratingMax = &v
	}

	reviews, _, err := s.reviewRepo.ListReviews(
		0,
		"",
		ratingMin,
		ratingMax,
		&trueVal,
		&falseVal,
		"",
		"",
		"",
		0,
		1000,
	)
	if err != nil {
		return 0, err
	}

	count := 0
	for _, review := range reviews {
		if review.IsWorkOrderCreated {
			continue
		}
		existing, _ := s.reviewRepo.GetWorkOrderByReviewID(review.ID)
		if existing != nil {
			_ = s.reviewRepo.MarkWorkOrderCreated(review.ID)
			continue
		}

		dueTime := time.Now().Add(workOrderDueHours * time.Hour)
		workOrder := &model.ReviewWorkOrder{
			StoreID:      review.StoreID,
			ReviewID:     review.ID,
			WorkOrderNo:  s.generateWorkOrderNo(),
			Type:         workOrderDefaultType,
			Title:        fmt.Sprintf("差评处理 - %s评分", review.Platform),
			Description:  review.Content,
			Priority:     "high",
			Status:       "pending",
			AssigneeID:   0,
			AssigneeName: "店长",
			DueTime:      &dueTime,
		}

		err := s.reviewRepo.CreateWorkOrder(workOrder)
		if err != nil {
			log.Printf("create work order for review %d failed: %v", review.ID, err)
			continue
		}
		_ = s.reviewRepo.MarkWorkOrderCreated(review.ID)
		count++
	}
	return count, nil
}

func (s *ReviewService) CreateWorkOrder(req *dto.WorkOrderCreateRequest) (*dto.WorkOrderResponse, error) {
	review, err := s.reviewRepo.GetReviewByID(req.ReviewID)
	if err != nil {
		return nil, errors.New("review not found")
	}

	existing, _ := s.reviewRepo.GetWorkOrderByReviewID(req.ReviewID)
	if existing != nil {
		return nil, errors.New("work order already exists for this review")
	}

	assigneeName := req.AssigneeName
	if assigneeName == "" {
		assigneeName = "店长"
	}
	dueTime := time.Now().Add(workOrderDueHours * time.Hour)

	workOrder := &model.ReviewWorkOrder{
		StoreID:      review.StoreID,
		ReviewID:     req.ReviewID,
		WorkOrderNo:  s.generateWorkOrderNo(),
		Type:         workOrderDefaultType,
		Title:        fmt.Sprintf("差评处理 - %s评分", review.Platform),
		Description:  review.Content,
		Priority:     req.Priority,
		Status:       "pending",
		AssigneeID:   req.AssigneeID,
		AssigneeName: assigneeName,
		DueTime:      &dueTime,
	}

	err = s.reviewRepo.CreateWorkOrder(workOrder)
	if err != nil {
		return nil, fmt.Errorf("create work order failed: %w", err)
	}
	_ = s.reviewRepo.MarkWorkOrderCreated(req.ReviewID)

	return s.GetWorkOrder(workOrder.ID)
}

func (s *ReviewService) ListWorkOrders(query *dto.WorkOrderQueryDTO) (*dto.PageResponse, error) {
	if query.Page == 0 {
		query.Page = 1
	}
	if query.PageSize == 0 {
		query.PageSize = 10
	}
	offset := (query.Page - 1) * query.PageSize

	var assigneeID *uint
	if query.AssigneeID > 0 {
		assigneeID = &query.AssigneeID
	}

	orders, total, err := s.reviewRepo.ListWorkOrders(
		query.StoreID,
		query.Status,
		query.Priority,
		assigneeID,
		query.Keyword,
		offset,
		query.PageSize,
	)
	if err != nil {
		return nil, err
	}

	var list []dto.WorkOrderResponse
	for _, o := range orders {
		list = append(list, *s.convertToWorkOrderResponse(&o))
	}

	return &dto.PageResponse{
		List:  list,
		Total: total,
		Page:  query.Page,
		Size:  query.PageSize,
	}, nil
}

func (s *ReviewService) GetWorkOrder(id uint) (*dto.WorkOrderResponse, error) {
	order, err := s.reviewRepo.GetWorkOrderByID(id)
	if err != nil {
		return nil, err
	}
	return s.convertToWorkOrderResponse(order), nil
}

func (s *ReviewService) HandleWorkOrder(id uint, req *dto.WorkOrderHandleRequest) error {
	_, err := s.reviewRepo.GetWorkOrderByID(id)
	if err != nil {
		return errors.New("work order not found")
	}

	now := time.Now()
	updates := map[string]interface{}{
		"status":        req.Status,
		"handle_result": req.HandleResult,
		"handler_name":  req.HandlerName,
		"handle_time":   &now,
	}
	return s.reviewRepo.UpdateWorkOrder(id, updates)
}

func (s *ReviewService) generateWorkOrderNo() string {
	now := time.Now()
	n, _ := rand.Int(rand.Reader, big.NewInt(1000000))
	return fmt.Sprintf("WO%s%06d", now.Format("20060102150405"), n.Int64())
}

func (s *ReviewService) convertToWorkOrderResponse(o *model.ReviewWorkOrder) *dto.WorkOrderResponse {
	storeName := ""
	if o.Store.Name != "" {
		storeName = o.Store.Name
	}
	reviewContent := ""
	if o.Review.Content != "" {
		reviewContent = o.Review.Content
	}
	return &dto.WorkOrderResponse{
		ID:            o.ID,
		StoreID:       o.StoreID,
		StoreName:     storeName,
		ReviewID:      o.ReviewID,
		ReviewContent: reviewContent,
		WorkOrderNo:   o.WorkOrderNo,
		Type:          o.Type,
		Title:         o.Title,
		Description:   o.Description,
		Priority:      o.Priority,
		Status:        o.Status,
		AssigneeID:    o.AssigneeID,
		AssigneeName:  o.AssigneeName,
		HandlerID:     o.HandlerID,
		HandlerName:   o.HandlerName,
		HandleTime:    o.HandleTime,
		HandleResult:  o.HandleResult,
		DueTime:       o.DueTime,
		CreatedAt:     o.CreatedAt,
		UpdatedAt:     o.UpdatedAt,
	}
}

func (s *ReviewService) CheckRatingAlerts() (int, error) {
	auths, err := s.reviewRepo.ListAuths(0)
	if err != nil {
		return 0, err
	}

	threshold := decimal.NewFromFloat(alertThresholdDrop)
	count := 0

	for _, auth := range auths {
		if auth.Status != 1 {
			continue
		}
		prev, curr, drop, err := s.AnalyzeRatingChange(auth.StoreID, auth.Platform)
		if err != nil {
			log.Printf("analyze rating change for store %d platform %s failed: %v", auth.StoreID, auth.Platform, err)
			continue
		}
		if drop.LessThan(threshold) {
			continue
		}

		alertType := "rating_drop"
		hasAlert, err := s.reviewRepo.HasAlertToday(auth.StoreID, auth.Platform, alertType)
		if err != nil {
			log.Printf("check alert exists failed: %v", err)
			continue
		}
		if hasAlert {
			continue
		}

		storeName := ""
		if auth.Store.Name != "" {
			storeName = auth.Store.Name
		}

		alert := &model.RatingAlert{
			StoreID:    auth.StoreID,
			Platform:   auth.Platform,
			AlertType:  alertType,
			Title:      fmt.Sprintf("【评分下降告警】%s - %s", storeName, auth.Platform),
			Content:    fmt.Sprintf("评分从%s下降到%s，下降了%s，超过阈值%s", prev.String(), curr.String(), drop.String(), threshold.String()),
			PrevRating: prev,
			CurrRating: curr,
			RatingDrop: drop,
			Threshold:  threshold,
			Status:     "unhandled",
			AlertTime:  time.Now(),
		}

		err = s.reviewRepo.CreateAlert(alert)
		if err != nil {
			log.Printf("create alert failed: %v", err)
			continue
		}

		_ = s.SendRatingAlert(alert, storeName)
		count++
	}
	return count, nil
}

func (s *ReviewService) SendRatingAlert(alert *model.RatingAlert, storeName string) error {
	title := fmt.Sprintf("【评分告警】%s - %s评分下降", storeName, alert.Platform)
	text := fmt.Sprintf("## 📉 评分下降告警\n\n"+
		"- **门店**：%s\n"+
		"- **平台**：%s\n"+
		"- **告警时间**：%s\n"+
		"- **上次评分**：%s\n"+
		"- **当前评分**：%s\n"+
		"- **下降幅度**：%s\n"+
		"- **告警阈值**：%s\n\n"+
		"⚠️ 评分下降超过阈值，请及时关注差评处理情况！\n",
		storeName,
		alert.Platform,
		time.Now().Format("2006-01-02 15:04:05"),
		alert.PrevRating.String(),
		alert.CurrRating.String(),
		alert.RatingDrop.String(),
		alert.Threshold.String(),
	)

	return s.dingtalkService.SendMarkdown(title, text, false, nil)
}

func (s *ReviewService) ListAlerts(query *dto.AlertQueryDTO) (*dto.PageResponse, error) {
	if query.Page == 0 {
		query.Page = 1
	}
	if query.PageSize == 0 {
		query.PageSize = 10
	}
	offset := (query.Page - 1) * query.PageSize

	alerts, total, err := s.reviewRepo.ListAlerts(
		query.StoreID,
		query.Status,
		query.AlertType,
		query.StartDate,
		query.EndDate,
		offset,
		query.PageSize,
	)
	if err != nil {
		return nil, err
	}

	var list []dto.AlertResponse
	for _, a := range alerts {
		list = append(list, *s.convertToAlertResponse(&a))
	}

	return &dto.PageResponse{
		List:  list,
		Total: total,
		Page:  query.Page,
		Size:  query.PageSize,
	}, nil
}

func (s *ReviewService) HandleAlert(id uint, req *dto.AlertHandleRequest) error {
	_, err := s.reviewRepo.GetAlertByID(id)
	if err != nil {
		return errors.New("alert not found")
	}

	now := time.Now()
	updates := map[string]interface{}{
		"status":        req.Status,
		"handle_remark": req.HandleRemark,
		"handler_name":  req.HandlerName,
		"handle_time":   &now,
	}
	return s.reviewRepo.UpdateAlert(id, updates)
}

func (s *ReviewService) convertToAlertResponse(a *model.RatingAlert) *dto.AlertResponse {
	storeName := ""
	if a.Store.Name != "" {
		storeName = a.Store.Name
	}
	return &dto.AlertResponse{
		ID:           a.ID,
		StoreID:      a.StoreID,
		StoreName:    storeName,
		Platform:     a.Platform,
		AlertType:    a.AlertType,
		Title:        a.Title,
		Content:      a.Content,
		PrevRating:   a.PrevRating,
		CurrRating:   a.CurrRating,
		RatingDrop:   a.RatingDrop,
		Threshold:    a.Threshold,
		Status:       a.Status,
		AlertTime:    a.AlertTime,
		HandlerID:    a.HandlerID,
		HandlerName:  a.HandlerName,
		HandleTime:   a.HandleTime,
		HandleRemark: a.HandleRemark,
		CreatedAt:    a.CreatedAt,
		UpdatedAt:    a.UpdatedAt,
	}
}
