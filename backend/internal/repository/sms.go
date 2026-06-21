package repository

import (
	"time"

	"gorm.io/gorm"
	"stalll-hub-pos/backend/internal/model"
)

type SmsRepository struct {
	db *gorm.DB
}

func NewSmsRepository(db *gorm.DB) *SmsRepository {
	return &SmsRepository{db: db}
}

func (r *SmsRepository) CreateTemplate(template *model.SmsTemplate) error {
	return r.db.Create(template).Error
}

func (r *SmsRepository) UpdateTemplate(template *model.SmsTemplate) error {
	return r.db.Save(template).Error
}

func (r *SmsRepository) DeleteTemplate(id uint) error {
	return r.db.Delete(&model.SmsTemplate{}, id).Error
}

func (r *SmsRepository) GetTemplateByID(id uint) (*model.SmsTemplate, error) {
	var template model.SmsTemplate
	err := r.db.First(&template, id).Error
	if err != nil {
		return nil, err
	}
	return &template, nil
}

func (r *SmsRepository) GetTemplateByCode(storeID uint, templateCode string) (*model.SmsTemplate, error) {
	var template model.SmsTemplate
	err := r.db.Where("store_id = ? AND template_code = ?", storeID, templateCode).First(&template).Error
	if err != nil {
		return nil, err
	}
	return &template, nil
}

func (r *SmsRepository) ListTemplates(storeID uint, templateType string, reviewStatus string, isActive *bool, keyword string, offset, limit int) ([]model.SmsTemplate, int64, error) {
	var templates []model.SmsTemplate
	var total int64

	query := r.db.Model(&model.SmsTemplate{}).Where("store_id = ?", storeID)

	if templateType != "" {
		query = query.Where("template_type = ?", templateType)
	}
	if reviewStatus != "" {
		query = query.Where("review_status = ?", reviewStatus)
	}
	if isActive != nil {
		query = query.Where("is_active = ?", *isActive)
	}
	if keyword != "" {
		query = query.Where("template_name LIKE ? OR template_code LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := query.Order("id DESC").Offset(offset).Limit(limit).Find(&templates).Error
	return templates, total, err
}

func (r *SmsRepository) UpdateTemplateReview(id uint, reviewStatus string, reviewRemark string, reviewerID uint, reviewerName string) error {
	now := time.Now()
	return r.db.Model(&model.SmsTemplate{}).Where("id = ?", id).Updates(map[string]interface{}{
		"review_status": reviewStatus,
		"review_remark": reviewRemark,
		"review_time":   &now,
		"reviewer_id":   reviewerID,
		"reviewer_name": reviewerName,
	}).Error
}

func (r *SmsRepository) GetActiveTemplateByCode(storeID uint, templateCode string) (*model.SmsTemplate, error) {
	var template model.SmsTemplate
	err := r.db.Where("store_id = ? AND template_code = ? AND is_active = ? AND review_status = ?",
		storeID, templateCode, true, "approved").First(&template).Error
	if err != nil {
		return nil, err
	}
	return &template, nil
}

func (r *SmsRepository) CreateTask(task *model.SmsTask) error {
	return r.db.Create(task).Error
}

func (r *SmsRepository) UpdateTask(task *model.SmsTask) error {
	return r.db.Save(task).Error
}

func (r *SmsRepository) DeleteTask(id uint) error {
	return r.db.Delete(&model.SmsTask{}, id).Error
}

func (r *SmsRepository) GetTaskByID(id uint) (*model.SmsTask, error) {
	var task model.SmsTask
	err := r.db.Preload("Template").First(&task, id).Error
	if err != nil {
		return nil, err
	}
	return &task, nil
}

func (r *SmsRepository) ListTasks(storeID uint, taskType string, status string, scheduleType string, keyword string, startDate, endDate string, offset, limit int) ([]model.SmsTask, int64, error) {
	var tasks []model.SmsTask
	var total int64

	query := r.db.Model(&model.SmsTask{}).Where("store_id = ?", storeID)

	if taskType != "" {
		query = query.Where("task_type = ?", taskType)
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}
	if scheduleType != "" {
		query = query.Where("schedule_type = ?", scheduleType)
	}
	if keyword != "" {
		query = query.Where("task_name LIKE ?", "%"+keyword+"%")
	}
	if startDate != "" {
		query = query.Where("created_at >= ?", startDate)
	}
	if endDate != "" {
		query = query.Where("created_at <= ?", endDate+" 23:59:59")
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := query.Preload("Template").Order("id DESC").Offset(offset).Limit(limit).Find(&tasks).Error
	return tasks, total, err
}

func (r *SmsRepository) GetPendingScheduleTasks(now time.Time) ([]model.SmsTask, error) {
	var tasks []model.SmsTask
	err := r.db.Where("status = ? AND schedule_type = ? AND scheduled_time <= ?",
		"pending", "scheduled", now).Find(&tasks).Error
	return tasks, err
}

func (r *SmsRepository) UpdateTaskStatus(id uint, status string) error {
	updates := map[string]interface{}{
		"status": status,
	}
	if status == "sending" {
		updates["start_time"] = time.Now()
	}
	if status == "completed" || status == "failed" {
		updates["end_time"] = time.Now()
	}
	return r.db.Model(&model.SmsTask{}).Where("id = ?", id).Updates(updates).Error
}

func (r *SmsRepository) UpdateTaskStats(id uint, successCount, failCount int) error {
	return r.db.Model(&model.SmsTask{}).Where("id = ?", id).Updates(map[string]interface{}{
		"success_count": successCount,
		"fail_count":    failCount,
	}).Error
}

func (r *SmsRepository) GetTaskStatistics(storeID uint, startDate, endDate string) (map[string]interface{}, error) {
	type StatsResult struct {
		TotalCount   int64 `json:"total_count"`
		SuccessCount int64 `json:"success_count"`
		FailCount    int64 `json:"fail_count"`
		TaskCount    int64 `json:"task_count"`
	}

	var stats StatsResult

	query := r.db.Model(&model.SmsRecord{}).Where("store_id = ?", storeID)

	if startDate != "" {
		query = query.Where("created_at >= ?", startDate)
	}
	if endDate != "" {
		query = query.Where("created_at <= ?", endDate+" 23:59:59")
	}

	if err := query.Select("COUNT(*) as total_count, SUM(CASE WHEN status = 'success' THEN 1 ELSE 0 END) as success_count, SUM(CASE WHEN status = 'failed' THEN 1 ELSE 0 END) as fail_count").
		Scan(&stats).Error; err != nil {
		return nil, err
	}

	taskQuery := r.db.Model(&model.SmsTask{}).Where("store_id = ?", storeID)
	if startDate != "" {
		taskQuery = taskQuery.Where("created_at >= ?", startDate)
	}
	if endDate != "" {
		taskQuery = taskQuery.Where("created_at <= ?", endDate+" 23:59:59")
	}

	if err := taskQuery.Count(&stats.TaskCount).Error; err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"total_count":   stats.TotalCount,
		"success_count": stats.SuccessCount,
		"fail_count":    stats.FailCount,
		"task_count":    stats.TaskCount,
	}, nil
}

func (r *SmsRepository) CreateRecord(record *model.SmsRecord) error {
	return r.db.Create(record).Error
}

func (r *SmsRepository) BulkCreateRecords(records []model.SmsRecord) error {
	if len(records) == 0 {
		return nil
	}
	return r.db.Create(&records).Error
}

func (r *SmsRepository) UpdateRecordStatus(id uint, status string, errorCode, errorMessage string) error {
	updates := map[string]interface{}{
		"status": status,
	}
	if status == "success" || status == "failed" {
		updates["send_time"] = time.Now()
	}
	if errorCode != "" {
		updates["error_code"] = errorCode
	}
	if errorMessage != "" {
		updates["error_message"] = errorMessage
	}
	return r.db.Model(&model.SmsRecord{}).Where("id = ?", id).Updates(updates).Error
}

func (r *SmsRepository) UpdateRecordByBizID(bizID string, updates map[string]interface{}) error {
	return r.db.Model(&model.SmsRecord{}).Where("biz_id = ?", bizID).Updates(updates).Error
}

func (r *SmsRepository) GetRecordByID(id uint) (*model.SmsRecord, error) {
	var record model.SmsRecord
	err := r.db.Preload("Task").Preload("Template").Preload("Member").First(&record, id).Error
	if err != nil {
		return nil, err
	}
	return &record, nil
}

func (r *SmsRepository) ListRecords(storeID uint, taskID uint, templateID uint, status string, phone string, sendType string, startDate, endDate string, offset, limit int) ([]model.SmsRecord, int64, error) {
	var records []model.SmsRecord
	var total int64

	query := r.db.Model(&model.SmsRecord{}).Where("store_id = ?", storeID)

	if taskID > 0 {
		query = query.Where("task_id = ?", taskID)
	}
	if templateID > 0 {
		query = query.Where("template_id = ?", templateID)
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}
	if phone != "" {
		query = query.Where("phone LIKE ?", "%"+phone+"%")
	}
	if sendType != "" {
		query = query.Where("send_type = ?", sendType)
	}
	if startDate != "" {
		query = query.Where("created_at >= ?", startDate)
	}
	if endDate != "" {
		query = query.Where("created_at <= ?", endDate+" 23:59:59")
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := query.Preload("Task").Preload("Template").Preload("Member").
		Order("id DESC").Offset(offset).Limit(limit).Find(&records).Error
	return records, total, err
}

func (r *SmsRepository) GetRecordsByTaskID(taskID uint) ([]model.SmsRecord, error) {
	var records []model.SmsRecord
	err := r.db.Where("task_id = ?", taskID).Find(&records).Error
	return records, err
}

func (r *SmsRepository) CountByTaskIDAndStatus(taskID uint, status string) (int64, error) {
	var count int64
	err := r.db.Model(&model.SmsRecord{}).Where("task_id = ? AND status = ?", taskID, status).Count(&count).Error
	return count, err
}
