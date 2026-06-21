package service

import (
	"errors"
	"fmt"
	"log"
	"math/rand"
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

const (
	batchSendSize = 100
	workerCount   = 5
)

type SmsService struct {
	smsRepo    *repository.SmsRepository
	memberRepo *repository.MemberRepository
}

func NewSmsService() *SmsService {
	return &SmsService{
		smsRepo:    repository.NewSmsRepository(nil),
		memberRepo: repository.NewMemberRepository(nil),
	}
}

func (s *SmsService) sendAliSms(phone, signName, templateCode, templateParam string) (bizID string, err error) {
	if rand.Float64() < 0.8 {
		bizID = fmt.Sprintf("ALI%s%d", time.Now().Format("20060102150405"), rand.Intn(100000))
		return bizID, nil
	}

	errorCodes := []string{"isv.MOBILE_NUMBER_ILLEGAL", "isv.BUSINESS_LIMIT_CONTROL", "isv.TEMPLATE_MISSING_PARAMETERS", "isv.OUT_OF_SERVICE"}
	errorMessages := []string{"手机号非法", "业务限流", "模板缺少参数", "服务不可用"}
	idx := rand.Intn(len(errorCodes))
	return "", fmt.Errorf("error_code: %s, error_message: %s", errorCodes[idx], errorMessages[idx])
}

func (s *SmsService) CreateTemplate(storeID uint, req *dto.SmsTemplateCreateDTO) (*dto.SmsTemplateResponse, error) {
	existing, _ := s.smsRepo.GetTemplateByCode(storeID, req.TemplateCode)
	if existing != nil {
		return nil, errors.New("template code already exists")
	}

	variableNames := ""
	variableCount := 0
	if len(req.VariableNames) > 0 {
		variableNames = strings.Join(req.VariableNames, ",")
		variableCount = len(req.VariableNames)
	}

	template := &model.SmsTemplate{
		StoreID:       storeID,
		TemplateCode:  req.TemplateCode,
		TemplateName:  req.TemplateName,
		TemplateType:  req.TemplateType,
		Content:       req.Content,
		SignName:      req.SignName,
		VariableCount: variableCount,
		VariableNames: variableNames,
		ReviewStatus:  "pending",
		IsActive:      false,
		UsedCount:     0,
		Description:   req.Description,
	}

	err := s.smsRepo.CreateTemplate(template)
	if err != nil {
		return nil, fmt.Errorf("create template failed: %w", err)
	}

	return s.GetTemplate(template.ID)
}

func (s *SmsService) UpdateTemplate(id uint, req *dto.SmsTemplateUpdateDTO) (*dto.SmsTemplateResponse, error) {
	template, err := s.smsRepo.GetTemplateByID(id)
	if err != nil {
		return nil, errors.New("template not found")
	}

	if req.TemplateName != nil {
		template.TemplateName = *req.TemplateName
	}
	if req.Content != nil {
		template.Content = *req.Content
	}
	if req.SignName != nil {
		template.SignName = *req.SignName
	}
	if req.VariableNames != nil {
		template.VariableNames = strings.Join(*req.VariableNames, ",")
		template.VariableCount = len(*req.VariableNames)
	}
	if req.Description != nil {
		template.Description = *req.Description
	}

	template.ReviewStatus = "pending"
	template.IsActive = false

	err = s.smsRepo.UpdateTemplate(template)
	if err != nil {
		return nil, err
	}

	return s.GetTemplate(id)
}

func (s *SmsService) DeleteTemplate(id uint) error {
	_, err := s.smsRepo.GetTemplateByID(id)
	if err != nil {
		return errors.New("template not found")
	}
	return s.smsRepo.DeleteTemplate(id)
}

func (s *SmsService) GetTemplate(id uint) (*dto.SmsTemplateResponse, error) {
	template, err := s.smsRepo.GetTemplateByID(id)
	if err != nil {
		return nil, err
	}
	return s.convertToTemplateResponse(template), nil
}

func (s *SmsService) ListTemplates(query *dto.SmsTemplateQueryDTO) (*dto.PageResponse, error) {
	offset := (query.Page - 1) * query.PageSize
	templates, total, err := s.smsRepo.ListTemplates(
		query.StoreID,
		query.TemplateType,
		query.ReviewStatus,
		query.IsActive,
		query.Keyword,
		offset,
		query.PageSize,
	)
	if err != nil {
		return nil, err
	}

	var list []dto.SmsTemplateResponse
	for _, t := range templates {
		list = append(list, *s.convertToTemplateResponse(&t))
	}

	return &dto.PageResponse{
		List:  list,
		Total: total,
		Page:  query.Page,
		Size:  query.PageSize,
	}, nil
}

func (s *SmsService) ReviewTemplate(id uint, req *dto.SmsTemplateReviewDTO) error {
	template, err := s.smsRepo.GetTemplateByID(id)
	if err != nil {
		return errors.New("template not found")
	}

	if template.ReviewStatus != "pending" {
		return errors.New("template already reviewed")
	}

	err = s.smsRepo.UpdateTemplateReview(id, req.ReviewStatus, req.ReviewRemark, req.ReviewerID, req.ReviewerName)
	if err != nil {
		return err
	}

	if req.ReviewStatus == "approved" {
		template.IsActive = true
		template.ReviewStatus = "approved"
		template.PlatformTemplateCode = template.TemplateCode
		return s.smsRepo.UpdateTemplate(template)
	}

	return nil
}

func (s *SmsService) CalculateTargetCount(storeID uint, query *dto.SmsTargetCountDTO) (int64, error) {
	db := database.DB.Model(&model.Member{})
	db = s.buildMemberQuery(db, storeID, query)
	var count int64
	err := db.Count(&count).Error
	return count, err
}

func (s *SmsService) buildMemberQuery(db *gorm.DB, storeID uint, query *dto.SmsTargetCountDTO) *gorm.DB {
	db = db.Where("store_id = ? AND status = 1", storeID)

	if len(query.MemberLevelIDs) > 0 {
		db = db.Where("level_id IN ?", query.MemberLevelIDs)
	}

	if query.MinConsumeCount > 0 {
		db = db.Where("order_count >= ?", query.MinConsumeCount)
	}
	if query.MaxConsumeCount > 0 {
		db = db.Where("order_count <= ?", query.MaxConsumeCount)
	}

	if !query.MinConsumeAmount.IsZero() {
		db = db.Where("total_consume >= ?", query.MinConsumeAmount)
	}
	if !query.MaxConsumeAmount.IsZero() {
		db = db.Where("total_consume <= ?", query.MaxConsumeAmount)
	}

	if query.MinPoints > 0 {
		db = db.Where("points >= ?", query.MinPoints)
	}
	if query.MaxPoints > 0 {
		db = db.Where("points <= ?", query.MaxPoints)
	}

	return db
}

func (s *SmsService) GetTargetMembers(storeID uint, query *dto.SmsTargetCountDTO) ([]model.Member, error) {
	db := database.DB.Model(&model.Member{})
	db = s.buildMemberQuery(db, storeID, query)

	var members []model.Member
	err := db.Preload("Level").Find(&members).Error
	return members, err
}

func (s *SmsService) CreateTask(storeID uint, req *dto.SmsTaskCreateDTO, creatorID uint, creatorName string) (*dto.SmsTaskResponse, error) {
	template, err := s.smsRepo.GetTemplateByID(req.TemplateID)
	if err != nil {
		return nil, errors.New("template not found")
	}

	if !template.IsActive || template.ReviewStatus != "approved" {
		return nil, errors.New("template is not active or not approved")
	}

	memberLevelIDs := ""
	if len(req.MemberLevelIDs) > 0 {
		ids := make([]string, len(req.MemberLevelIDs))
		for i, id := range req.MemberLevelIDs {
			ids[i] = fmt.Sprintf("%d", id)
		}
		memberLevelIDs = strings.Join(ids, ",")
	}

	targetQuery := &dto.SmsTargetCountDTO{
		StoreID:          storeID,
		MemberLevelIDs:   req.MemberLevelIDs,
		MinConsumeCount:  req.MinConsumeCount,
		MaxConsumeCount:  req.MaxConsumeCount,
		MinConsumeAmount: req.MinConsumeAmount,
		MaxConsumeAmount: req.MaxConsumeAmount,
		MinPoints:        req.MinPoints,
		MaxPoints:        req.MaxPoints,
	}
	targetCount, _ := s.CalculateTargetCount(storeID, targetQuery)

	status := "draft"
	if req.ScheduleType == "immediately" {
		status = "pending"
	}

	task := &model.SmsTask{
		StoreID:          storeID,
		TaskName:         req.TaskName,
		TaskType:         req.TaskType,
		TemplateID:       req.TemplateID,
		TemplateCode:     template.TemplateCode,
		SignName:         req.SignName,
		Content:          template.Content,
		TargetType:       req.TargetType,
		MemberLevelIDs:   memberLevelIDs,
		MinConsumeCount:  req.MinConsumeCount,
		MaxConsumeCount:  req.MaxConsumeCount,
		MinConsumeAmount: req.MinConsumeAmount,
		MaxConsumeAmount: req.MaxConsumeAmount,
		MinPoints:        req.MinPoints,
		MaxPoints:        req.MaxPoints,
		TargetCount:      int(targetCount),
		SuccessCount:     0,
		FailCount:        0,
		ReadCount:        0,
		ConversionCount:  0,
		ConversionAmount: decimal.Zero,
		ConversionRate:   decimal.Zero,
		SuccessRate:      decimal.Zero,
		ScheduleType:     req.ScheduleType,
		ScheduledTime:    req.ScheduledTime,
		Status:           status,
		CreatorID:        creatorID,
		CreatorName:      creatorName,
		Remark:           req.Remark,
	}

	if task.TaskType == "" {
		task.TaskType = "marketing"
	}
	if task.TargetType == "" {
		task.TargetType = "custom"
	}
	if task.ScheduleType == "" {
		task.ScheduleType = "immediately"
	}

	err = s.smsRepo.CreateTask(task)
	if err != nil {
		return nil, fmt.Errorf("create task failed: %w", err)
	}

	return s.GetTask(task.ID)
}

func (s *SmsService) UpdateTask(id uint, req *dto.SmsTaskUpdateDTO) error {
	_, err := s.smsRepo.GetTaskByID(id)
	if err != nil {
		return errors.New("task not found")
	}
	return nil
}

func (s *SmsService) DeleteTask(id uint) error {
	_, err := s.smsRepo.GetTaskByID(id)
	if err != nil {
		return errors.New("task not found")
	}
	return s.smsRepo.DeleteTask(id)
}

func (s *SmsService) GetTask(id uint) (*dto.SmsTaskResponse, error) {
	task, err := s.smsRepo.GetTaskByID(id)
	if err != nil {
		return nil, err
	}
	return s.convertToTaskResponse(task), nil
}

func (s *SmsService) ListTasks(query *dto.SmsTaskQueryDTO) (*dto.PageResponse, error) {
	offset := (query.Page - 1) * query.PageSize
	tasks, total, err := s.smsRepo.ListTasks(
		query.StoreID,
		query.TaskType,
		query.Status,
		query.ScheduleType,
		query.Keyword,
		query.StartDate,
		query.EndDate,
		offset,
		query.PageSize,
	)
	if err != nil {
		return nil, err
	}

	var list []dto.SmsTaskResponse
	for _, t := range tasks {
		list = append(list, *s.convertToTaskResponse(&t))
	}

	return &dto.PageResponse{
		List:  list,
		Total: total,
		Page:  query.Page,
		Size:  query.PageSize,
	}, nil
}

func (s *SmsService) StartTask(id uint) error {
	task, err := s.smsRepo.GetTaskByID(id)
	if err != nil {
		return errors.New("task not found")
	}

	if task.Status == "sending" {
		return errors.New("task is already sending")
	}
	if task.Status == "completed" {
		return errors.New("task already completed")
	}

	err = s.smsRepo.UpdateTaskStatus(id, "sending")
	if err != nil {
		return err
	}

	go s.ExecuteTask(task)

	return nil
}

func (s *SmsService) PauseTask(id uint) error {
	task, err := s.smsRepo.GetTaskByID(id)
	if err != nil {
		return errors.New("task not found")
	}

	if task.Status != "sending" && task.Status != "pending" {
		return errors.New("task cannot be paused")
	}

	return s.smsRepo.UpdateTaskStatus(id, "paused")
}

func (s *SmsService) ExecuteTask(task *model.SmsTask) error {
	targetQuery := &dto.SmsTargetCountDTO{
		StoreID:          task.StoreID,
		MemberLevelIDs:   s.parseUintIDs(task.MemberLevelIDs),
		MinConsumeCount:  task.MinConsumeCount,
		MaxConsumeCount:  task.MaxConsumeCount,
		MinConsumeAmount: task.MinConsumeAmount,
		MaxConsumeAmount: task.MaxConsumeAmount,
		MinPoints:        task.MinPoints,
		MaxPoints:        task.MaxPoints,
	}

	members, err := s.GetTargetMembers(task.StoreID, targetQuery)
	if err != nil {
		log.Printf("get target members failed: %v", err)
		_ = s.smsRepo.UpdateTaskStatus(task.ID, "failed")
		return err
	}

	var records []model.SmsRecord
	for _, member := range members {
		record := model.SmsRecord{
			StoreID:      task.StoreID,
			TaskID:       task.ID,
			TemplateID:   task.TemplateID,
			TemplateCode: task.TemplateCode,
			SignName:     task.SignName,
			Content:      task.Content,
			MemberID:     member.ID,
			MemberName:   member.Name,
			Phone:        member.Phone,
			SendType:     task.TaskType,
			Status:       "pending",
			PricePer:     decimal.NewFromFloat(0.045),
			CostAmount:   decimal.Zero,
		}
		records = append(records, record)
	}

	err = s.smsRepo.BulkCreateRecords(records)
	if err != nil {
		log.Printf("bulk create records failed: %v", err)
		_ = s.smsRepo.UpdateTaskStatus(task.ID, "failed")
		return err
	}

	allRecords, _ := s.smsRepo.GetRecordsByTaskID(task.ID)

	successCount := 0
	failCount := 0
	for i := 0; i < len(allRecords); i += batchSendSize {
		end := i + batchSendSize
		if end > len(allRecords) {
			end = len(allRecords)
		}
		batch := allRecords[i:end]

		results := s.batchSend(batch)
		for _, result := range results {
			if result {
				successCount++
			} else {
				failCount++
			}
		}

		_ = s.smsRepo.UpdateTaskStats(task.ID, successCount, failCount)
	}

	successRate := decimal.Zero
	if len(allRecords) > 0 {
		successRate = decimal.NewFromInt(int64(successCount)).Mul(decimal.NewFromInt(100)).Div(decimal.NewFromInt(int64(len(allRecords))))
	}

	database.DB.Model(&model.SmsTask{}).Where("id = ?", task.ID).Updates(map[string]interface{}{
		"success_rate": successRate,
		"target_count": len(allRecords),
	})

	_ = s.smsRepo.UpdateTaskStatus(task.ID, "completed")

	return nil
}

func (s *SmsService) SendSms(storeID uint, memberID uint, phone, signName, templateCode, content string, taskID uint) error {
	template, err := s.smsRepo.GetActiveTemplateByCode(storeID, templateCode)
	if err != nil {
		return errors.New("template not found or not active")
	}

	record := &model.SmsRecord{
		StoreID:      storeID,
		TaskID:       taskID,
		TemplateID:   template.ID,
		TemplateCode: templateCode,
		SignName:     signName,
		Content:      content,
		MemberID:     memberID,
		Phone:        phone,
		SendType:     "marketing",
		Status:       "pending",
		PricePer:     decimal.NewFromFloat(0.045),
	}

	err = s.smsRepo.CreateRecord(record)
	if err != nil {
		return err
	}

	bizID, err := s.sendAliSms(phone, signName, templateCode, "")
	if err != nil {
		errorCode := ""
		errorMessage := ""
		errParts := strings.SplitN(err.Error(), ", ", 2)
		if len(errParts) == 2 {
			errorCode = strings.TrimPrefix(errParts[0], "error_code: ")
			errorMessage = strings.TrimPrefix(errParts[1], "error_message: ")
		}
		_ = s.smsRepo.UpdateRecordStatus(record.ID, "failed", errorCode, errorMessage)
		return err
	}

	_ = s.smsRepo.UpdateRecordStatus(record.ID, "success", "", "")
	database.DB.Model(&model.SmsRecord{}).Where("id = ?", record.ID).Update("biz_id", bizID)

	return nil
}

func (s *SmsService) SendTestSms(storeID uint, req *dto.SmsSendTestDTO) error {
	template, err := s.smsRepo.GetTemplateByID(req.TemplateID)
	if err != nil {
		return errors.New("template not found")
	}

	content := req.Content
	if content == "" {
		content = template.Content
	}

	signName := req.SignName
	if signName == "" {
		signName = template.SignName
	}

	record := &model.SmsRecord{
		StoreID:      storeID,
		TemplateID:   req.TemplateID,
		TemplateCode: template.TemplateCode,
		SignName:     signName,
		Content:      content,
		Phone:        req.Phone,
		SendType:     "test",
		Status:       "pending",
		PricePer:     decimal.NewFromFloat(0.045),
	}

	err = s.smsRepo.CreateRecord(record)
	if err != nil {
		return err
	}

	bizID, err := s.sendAliSms(req.Phone, signName, template.TemplateCode, "")
	if err != nil {
		errorCode := ""
		errorMessage := ""
		errParts := strings.SplitN(err.Error(), ", ", 2)
		if len(errParts) == 2 {
			errorCode = strings.TrimPrefix(errParts[0], "error_code: ")
			errorMessage = strings.TrimPrefix(errParts[1], "error_message: ")
		}
		_ = s.smsRepo.UpdateRecordStatus(record.ID, "failed", errorCode, errorMessage)
		return err
	}

	_ = s.smsRepo.UpdateRecordStatus(record.ID, "success", "", "")
	database.DB.Model(&model.SmsRecord{}).Where("id = ?", record.ID).Update("biz_id", bizID)

	return nil
}

func (s *SmsService) batchSend(records []model.SmsRecord) []bool {
	results := make([]bool, len(records))
	jobs := make(chan int, len(records))
	resultsChan := make(chan struct {
		index  int
		success bool
	}, len(records))

	for w := 0; w < workerCount; w++ {
		go func() {
			for idx := range jobs {
				record := records[idx]
				bizID, err := s.sendAliSms(record.Phone, record.SignName, record.TemplateCode, "")
				if err != nil {
					errorCode := ""
					errorMessage := ""
					errParts := strings.SplitN(err.Error(), ", ", 2)
					if len(errParts) == 2 {
						errorCode = strings.TrimPrefix(errParts[0], "error_code: ")
						errorMessage = strings.TrimPrefix(errParts[1], "error_message: ")
					}
					_ = s.smsRepo.UpdateRecordStatus(record.ID, "failed", errorCode, errorMessage)
					resultsChan <- struct {
						index  int
						success bool
					}{idx, false}
				} else {
					_ = s.smsRepo.UpdateRecordStatus(record.ID, "success", "", "")
					database.DB.Model(&model.SmsRecord{}).Where("id = ?", record.ID).Update("biz_id", bizID)
					resultsChan <- struct {
						index  int
						success bool
					}{idx, true}
				}
			}
		}()
	}

	for i := range records {
		jobs <- i
	}
	close(jobs)

	for i := 0; i < len(records); i++ {
		result := <-resultsChan
		results[result.index] = result.success
	}

	return results
}

func (s *SmsService) ListRecords(query *dto.SmsRecordQueryDTO) (*dto.PageResponse, error) {
	offset := (query.Page - 1) * query.PageSize
	records, total, err := s.smsRepo.ListRecords(
		query.StoreID,
		query.TaskID,
		query.TemplateID,
		query.Status,
		query.Phone,
		query.SendType,
		query.StartDate,
		query.EndDate,
		offset,
		query.PageSize,
	)
	if err != nil {
		return nil, err
	}

	var list []dto.SmsRecordResponse
	for _, r := range records {
		list = append(list, *s.convertToRecordResponse(&r))
	}

	return &dto.PageResponse{
		List:  list,
		Total: total,
		Page:  query.Page,
		Size:  query.PageSize,
	}, nil
}

func (s *SmsService) GetRecord(id uint) (*dto.SmsRecordResponse, error) {
	record, err := s.smsRepo.GetRecordByID(id)
	if err != nil {
		return nil, err
	}
	return s.convertToRecordResponse(record), nil
}

func (s *SmsService) GetTaskStatistics(query *dto.SmsTaskQueryDTO) (*dto.SmsTaskStatisticsResponse, error) {
	stats, err := s.smsRepo.GetTaskStatistics(query.StoreID, query.StartDate, query.EndDate)
	if err != nil {
		return nil, err
	}

	totalTasks, _ := stats["task_count"].(int64)
	totalSent, _ := stats["total_count"].(int64)
	successCount, _ := stats["success_count"].(int64)
	failCount, _ := stats["fail_count"].(int64)

	successRate := decimal.Zero
	if totalSent > 0 {
		successRate = decimal.NewFromInt(successCount).Mul(decimal.NewFromInt(100)).Div(decimal.NewFromInt(totalSent))
	}

	conversionRate := decimal.Zero
	if successCount > 0 {
		conversionRate = decimal.NewFromInt(0).Mul(decimal.NewFromInt(100)).Div(decimal.NewFromInt(successCount))
	}

	return &dto.SmsTaskStatisticsResponse{
		TotalTasks:       int(totalTasks),
		TotalSent:        int(totalSent),
		SuccessCount:     int(successCount),
		FailCount:        int(failCount),
		SuccessRate:      successRate,
		ReadCount:        0,
		ConversionCount:  0,
		ConversionRate:   conversionRate,
		ConversionAmount: decimal.Zero,
	}, nil
}

func (s *SmsService) UpdateTaskConversionStats(taskID uint) error {
	task, err := s.smsRepo.GetTaskByID(taskID)
	if err != nil {
		return errors.New("task not found")
	}

	successCount, _ := s.smsRepo.CountByTaskIDAndStatus(taskID, "success")

	type ConversionResult struct {
		Count  int64
		Amount decimal.Decimal
	}
	var result ConversionResult

	database.DB.Model(&model.SmsRecord{}).
		Select("COUNT(*) as count, COALESCE(SUM(conversion_amount), 0) as amount").
		Where("task_id = ? AND is_converted = ?", taskID, true).
		Scan(&result)

	conversionRate := decimal.Zero
	if successCount > 0 {
		conversionRate = decimal.NewFromInt(result.Count).Mul(decimal.NewFromInt(100)).Div(decimal.NewFromInt(successCount))
	}

	successRate := decimal.Zero
	if task.TargetCount > 0 {
		successRate = decimal.NewFromInt(successCount).Mul(decimal.NewFromInt(100)).Div(decimal.NewFromInt(int64(task.TargetCount)))
	}

	return database.DB.Model(&model.SmsTask{}).Where("id = ?", taskID).Updates(map[string]interface{}{
		"conversion_count":  result.Count,
		"conversion_amount": result.Amount,
		"conversion_rate":   conversionRate,
		"success_rate":      successRate,
	}).Error
}

func (s *SmsService) convertToTemplateResponse(t *model.SmsTemplate) *dto.SmsTemplateResponse {
	storeName := ""
	if t.Store.Name != "" {
		storeName = t.Store.Name
	}

	return &dto.SmsTemplateResponse{
		ID:                   t.ID,
		StoreID:              t.StoreID,
		StoreName:            storeName,
		TemplateCode:         t.TemplateCode,
		TemplateName:         t.TemplateName,
		TemplateType:         t.TemplateType,
		Content:              t.Content,
		SignName:             t.SignName,
		VariableCount:        t.VariableCount,
		VariableNames:        t.VariableNames,
		ReviewStatus:         t.ReviewStatus,
		ReviewRemark:         t.ReviewRemark,
		ReviewTime:           t.ReviewTime,
		ReviewerID:           t.ReviewerID,
		ReviewerName:         t.ReviewerName,
		PlatformTemplateCode: t.PlatformTemplateCode,
		IsActive:             t.IsActive,
		UsedCount:            t.UsedCount,
		Description:          t.Description,
		CreatedAt:            t.CreatedAt,
		UpdatedAt:            t.UpdatedAt,
	}
}

func (s *SmsService) convertToTaskResponse(t *model.SmsTask) *dto.SmsTaskResponse {
	storeName := ""
	if t.Store.Name != "" {
		storeName = t.Store.Name
	}

	templateName := ""
	if t.Template != nil {
		templateName = t.Template.TemplateName
	}

	return &dto.SmsTaskResponse{
		ID:               t.ID,
		StoreID:          t.StoreID,
		StoreName:        storeName,
		TaskName:         t.TaskName,
		TaskType:         t.TaskType,
		TemplateID:       t.TemplateID,
		TemplateName:     templateName,
		TemplateCode:     t.TemplateCode,
		SignName:         t.SignName,
		Content:          t.Content,
		TargetType:       t.TargetType,
		TargetFilters:    t.TargetFilters,
		MemberLevelIDs:   t.MemberLevelIDs,
		MinConsumeCount:  t.MinConsumeCount,
		MaxConsumeCount:  t.MaxConsumeCount,
		MinConsumeAmount: t.MinConsumeAmount,
		MaxConsumeAmount: t.MaxConsumeAmount,
		MinPoints:        t.MinPoints,
		MaxPoints:        t.MaxPoints,
		TargetCount:      t.TargetCount,
		SuccessCount:     t.SuccessCount,
		FailCount:        t.FailCount,
		ReadCount:        t.ReadCount,
		ConversionCount:  t.ConversionCount,
		ConversionAmount: t.ConversionAmount,
		ConversionRate:   t.ConversionRate,
		SuccessRate:      t.SuccessRate,
		ScheduleType:     t.ScheduleType,
		ScheduledTime:    t.ScheduledTime,
		StartTime:        t.StartTime,
		EndTime:          t.EndTime,
		Status:           t.Status,
		CreatorID:        t.CreatorID,
		CreatorName:      t.CreatorName,
		Remark:           t.Remark,
		CreatedAt:        t.CreatedAt,
		UpdatedAt:        t.UpdatedAt,
	}
}

func (s *SmsService) convertToRecordResponse(r *model.SmsRecord) *dto.SmsRecordResponse {
	storeName := ""
	if r.Store.Name != "" {
		storeName = r.Store.Name
	}

	taskName := ""
	if r.Task != nil {
		taskName = r.Task.TaskName
	}

	templateName := ""
	if r.Template != nil {
		templateName = r.Template.TemplateName
	}

	memberName := ""
	if r.Member != nil {
		memberName = r.Member.Name
	}

	return &dto.SmsRecordResponse{
		ID:                r.ID,
		StoreID:           r.StoreID,
		StoreName:         storeName,
		TaskID:            r.TaskID,
		TaskName:          taskName,
		TemplateID:        r.TemplateID,
		TemplateName:      templateName,
		TemplateCode:      r.TemplateCode,
		SignName:          r.SignName,
		Content:           r.Content,
		MemberID:          r.MemberID,
		MemberName:        memberName,
		Phone:             r.Phone,
		SendType:          r.SendType,
		Status:            r.Status,
		ErrorCode:         r.ErrorCode,
		ErrorMessage:      r.ErrorMessage,
		RequestID:         r.RequestID,
		BizID:             r.BizID,
		SendTime:          r.SendTime,
		DeliverTime:       r.DeliverTime,
		IsRead:            r.IsRead,
		ReadTime:          r.ReadTime,
		IsConverted:       r.IsConverted,
		ConversionOrderID: r.ConversionOrderID,
		ConversionAmount:  r.ConversionAmount,
		ConversionTime:    r.ConversionTime,
		PricePer:          r.PricePer,
		CostAmount:        r.CostAmount,
		CreatedAt:         r.CreatedAt,
		UpdatedAt:         r.UpdatedAt,
	}
}

func (s *SmsService) parseUintIDs(idsStr string) []uint {
	if idsStr == "" {
		return nil
	}
	parts := strings.Split(idsStr, ",")
	var ids []uint
	for _, part := range parts {
		id, err := strconv.ParseUint(part, 10, 32)
		if err == nil {
			ids = append(ids, uint(id))
		}
	}
	return ids
}
