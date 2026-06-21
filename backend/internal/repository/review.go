package repository

import (
	"time"

	"gorm.io/gorm"
	"stalll-hub-pos/backend/internal/model"
)

type ReviewRepository struct {
	db *gorm.DB
}

func NewReviewRepository(database *gorm.DB) *ReviewRepository {
	return &ReviewRepository{db: database}
}

func (r *ReviewRepository) CreateRating(rating *model.PlatformReviewRating) error {
	return r.db.Create(rating).Error
}

func (r *ReviewRepository) GetRatingByDate(storeID uint, platform string, ratingDate string) (*model.PlatformReviewRating, error) {
	var rating model.PlatformReviewRating
	err := r.db.Where("store_id = ? AND platform = ? AND rating_date = ?", storeID, platform, ratingDate).
		First(&rating).Error
	if err != nil {
		return nil, err
	}
	return &rating, nil
}

func (r *ReviewRepository) ListRatings(storeID uint, platform string, startDate, endDate string, offset, limit int) ([]model.PlatformReviewRating, int64, error) {
	var ratings []model.PlatformReviewRating
	var total int64

	db := r.db.Model(&model.PlatformReviewRating{})
	if storeID > 0 {
		db = db.Where("store_id = ?", storeID)
	}
	if platform != "" {
		db = db.Where("platform = ?", platform)
	}
	if startDate != "" {
		db = db.Where("rating_date >= ?", startDate)
	}
	if endDate != "" {
		db = db.Where("rating_date <= ?", endDate)
	}

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := db.Order("rating_date DESC").Offset(offset).Limit(limit).Find(&ratings).Error
	return ratings, total, err
}

func (r *ReviewRepository) GetRatingTrend(storeID uint, platform string, startDate, endDate string) ([]model.PlatformReviewRating, error) {
	var ratings []model.PlatformReviewRating
	db := r.db.Model(&model.PlatformReviewRating{})
	if storeID > 0 {
		db = db.Where("store_id = ?", storeID)
	}
	if platform != "" {
		db = db.Where("platform = ?", platform)
	}
	if startDate != "" {
		db = db.Where("rating_date >= ?", startDate)
	}
	if endDate != "" {
		db = db.Where("rating_date <= ?", endDate)
	}
	err := db.Order("rating_date ASC").Find(&ratings).Error
	return ratings, err
}

func (r *ReviewRepository) GetLatestRating(storeID uint, platform string) (*model.PlatformReviewRating, error) {
	var rating model.PlatformReviewRating
	err := r.db.Where("store_id = ? AND platform = ?", storeID, platform).
		Order("rating_date DESC").
		First(&rating).Error
	if err != nil {
		return nil, err
	}
	return &rating, nil
}

func (r *ReviewRepository) GetPreviousDayRating(storeID uint, platform string, currentDate string) (*model.PlatformReviewRating, error) {
	var rating model.PlatformReviewRating
	err := r.db.Where("store_id = ? AND platform = ? AND rating_date < ?", storeID, platform, currentDate).
		Order("rating_date DESC").
		First(&rating).Error
	if err != nil {
		return nil, err
	}
	return &rating, nil
}

func (r *ReviewRepository) CreateReview(review *model.PlatformReview) error {
	return r.db.Create(review).Error
}

func (r *ReviewRepository) UpsertReview(review *model.PlatformReview) error {
	var existing model.PlatformReview
	err := r.db.Where("platform_id = ?", review.PlatformID).First(&existing).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return r.db.Create(review).Error
		}
		return err
	}
	review.ID = existing.ID
	review.CreatedAt = existing.CreatedAt
	return r.db.Save(review).Error
}

func (r *ReviewRepository) GetReviewByPlatformID(platform string, platformID string) (*model.PlatformReview, error) {
	var review model.PlatformReview
	err := r.db.Where("platform = ? AND platform_id = ?", platform, platformID).
		First(&review).Error
	if err != nil {
		return nil, err
	}
	return &review, nil
}

func (r *ReviewRepository) GetReviewByID(id uint) (*model.PlatformReview, error) {
	var review model.PlatformReview
	err := r.db.Preload("Store").First(&review, id).Error
	if err != nil {
		return nil, err
	}
	return &review, nil
}

func (r *ReviewRepository) ListReviews(storeID uint, platform string, ratingMin, ratingMax *float64, isBadReview *bool, isReplied *bool, keyword string, startDate, endDate string, offset, limit int) ([]model.PlatformReview, int64, error) {
	var reviews []model.PlatformReview
	var total int64

	db := r.db.Model(&model.PlatformReview{})
	if storeID > 0 {
		db = db.Where("store_id = ?", storeID)
	}
	if platform != "" {
		db = db.Where("platform = ?", platform)
	}
	if ratingMin != nil {
		db = db.Where("rating >= ?", *ratingMin)
	}
	if ratingMax != nil {
		db = db.Where("rating <= ?", *ratingMax)
	}
	if isBadReview != nil {
		db = db.Where("is_bad_review = ?", *isBadReview)
	}
	if isReplied != nil {
		db = db.Where("is_replied = ?", *isReplied)
	}
	if keyword != "" {
		db = db.Where("content LIKE ? OR user_nickname LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}
	if startDate != "" {
		db = db.Where("review_time >= ?", startDate)
	}
	if endDate != "" {
		db = db.Where("review_time <= ?", endDate+" 23:59:59")
	}

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := db.Preload("Store").Order("review_time DESC").Offset(offset).Limit(limit).Find(&reviews).Error
	return reviews, total, err
}

func (r *ReviewRepository) UpdateReviewReply(id uint, replyContent string) error {
	now := time.Now()
	return r.db.Model(&model.PlatformReview{}).Where("id = ?", id).
		Updates(map[string]interface{}{
			"reply_content": replyContent,
			"reply_time":    &now,
			"is_replied":    true,
		}).Error
}

func (r *ReviewRepository) MarkWorkOrderCreated(id uint) error {
	return r.db.Model(&model.PlatformReview{}).Where("id = ?", id).
		Update("is_work_order_created", true).Error
}

func (r *ReviewRepository) CreateWorkOrder(order *model.ReviewWorkOrder) error {
	return r.db.Create(order).Error
}

func (r *ReviewRepository) GetWorkOrderByID(id uint) (*model.ReviewWorkOrder, error) {
	var order model.ReviewWorkOrder
	err := r.db.Preload("Review").Preload("Store").First(&order, id).Error
	if err != nil {
		return nil, err
	}
	return &order, nil
}

func (r *ReviewRepository) GetWorkOrderByReviewID(reviewID uint) (*model.ReviewWorkOrder, error) {
	var order model.ReviewWorkOrder
	err := r.db.Where("review_id = ?", reviewID).
		Preload("Review").Preload("Store").
		First(&order).Error
	if err != nil {
		return nil, err
	}
	return &order, nil
}

func (r *ReviewRepository) ListWorkOrders(storeID uint, status string, priority string, assigneeID *uint, keyword string, offset, limit int) ([]model.ReviewWorkOrder, int64, error) {
	var orders []model.ReviewWorkOrder
	var total int64

	db := r.db.Model(&model.ReviewWorkOrder{})
	if storeID > 0 {
		db = db.Where("store_id = ?", storeID)
	}
	if status != "" {
		db = db.Where("status = ?", status)
	}
	if priority != "" {
		db = db.Where("priority = ?", priority)
	}
	if assigneeID != nil {
		db = db.Where("assignee_id = ?", *assigneeID)
	}
	if keyword != "" {
		db = db.Where("title LIKE ? OR description LIKE ? OR work_order_no LIKE ?", "%"+keyword+"%", "%"+keyword+"%", "%"+keyword+"%")
	}

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := db.Preload("Review").Preload("Store").
		Order("id DESC").Offset(offset).Limit(limit).Find(&orders).Error
	return orders, total, err
}

func (r *ReviewRepository) UpdateWorkOrder(id uint, updates map[string]interface{}) error {
	return r.db.Model(&model.ReviewWorkOrder{}).Where("id = ?", id).Updates(updates).Error
}

func (r *ReviewRepository) GetPendingWorkOrderCount(storeID uint) (int64, error) {
	var count int64
	db := r.db.Model(&model.ReviewWorkOrder{}).Where("status = ?", "pending")
	if storeID > 0 {
		db = db.Where("store_id = ?", storeID)
	}
	err := db.Count(&count).Error
	return count, err
}

func (r *ReviewRepository) CreateAlert(alert *model.RatingAlert) error {
	return r.db.Create(alert).Error
}

func (r *ReviewRepository) GetAlertByID(id uint) (*model.RatingAlert, error) {
	var alert model.RatingAlert
	err := r.db.Preload("Store").First(&alert, id).Error
	if err != nil {
		return nil, err
	}
	return &alert, nil
}

func (r *ReviewRepository) ListAlerts(storeID uint, status string, alertType string, startDate, endDate string, offset, limit int) ([]model.RatingAlert, int64, error) {
	var alerts []model.RatingAlert
	var total int64

	db := r.db.Model(&model.RatingAlert{})
	if storeID > 0 {
		db = db.Where("store_id = ?", storeID)
	}
	if status != "" {
		db = db.Where("status = ?", status)
	}
	if alertType != "" {
		db = db.Where("alert_type = ?", alertType)
	}
	if startDate != "" {
		db = db.Where("alert_time >= ?", startDate)
	}
	if endDate != "" {
		db = db.Where("alert_time <= ?", endDate+" 23:59:59")
	}

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := db.Preload("Store").Order("alert_time DESC").Offset(offset).Limit(limit).Find(&alerts).Error
	return alerts, total, err
}

func (r *ReviewRepository) UpdateAlert(id uint, updates map[string]interface{}) error {
	return r.db.Model(&model.RatingAlert{}).Where("id = ?", id).Updates(updates).Error
}

func (r *ReviewRepository) HasAlertToday(storeID uint, platform string, alertType string) (bool, error) {
	var count int64
	today := time.Now().Format("2006-01-02")
	err := r.db.Model(&model.RatingAlert{}).
		Where("store_id = ? AND platform = ? AND alert_type = ? AND DATE(alert_time) = ?", storeID, platform, alertType, today).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *ReviewRepository) UpsertAuth(auth *model.StorePlatformAuth) error {
	var existing model.StorePlatformAuth
	err := r.db.Where("store_id = ? AND platform = ?", auth.StoreID, auth.Platform).First(&existing).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return r.db.Create(auth).Error
		}
		return err
	}
	auth.ID = existing.ID
	auth.CreatedAt = existing.CreatedAt
	return r.db.Save(auth).Error
}

func (r *ReviewRepository) GetAuth(storeID uint, platform string) (*model.StorePlatformAuth, error) {
	var auth model.StorePlatformAuth
	err := r.db.Where("store_id = ? AND platform = ?", storeID, platform).
		First(&auth).Error
	if err != nil {
		return nil, err
	}
	return &auth, nil
}

func (r *ReviewRepository) ListAuths(storeID uint) ([]model.StorePlatformAuth, error) {
	var auths []model.StorePlatformAuth
	db := r.db.Model(&model.StorePlatformAuth{})
	if storeID > 0 {
		db = db.Where("store_id = ?", storeID)
	}
	err := db.Order("id ASC").Find(&auths).Error
	return auths, err
}

func (r *ReviewRepository) UpdateSyncStatus(storeID uint, platform string, syncStatus string, syncError string, lastSyncTime *time.Time) error {
	updates := map[string]interface{}{
		"sync_status": syncStatus,
		"sync_error":  syncError,
	}
	if lastSyncTime != nil {
		updates["last_sync_time"] = lastSyncTime
	}
	return r.db.Model(&model.StorePlatformAuth{}).
		Where("store_id = ? AND platform = ?", storeID, platform).
		Updates(updates).Error
}
