package dto

import (
	"time"

	"github.com/shopspring/decimal"
)

type SmsTemplateCreateDTO struct {
	StoreID       uint     `json:"store_id" binding:"required"`
	TemplateCode  string   `json:"template_code" binding:"required,max=50"`
	TemplateName  string   `json:"template_name" binding:"required,max=100"`
	TemplateType  string   `json:"template_type" binding:"required,max=20"`
	Content       string   `json:"content" binding:"required"`
	SignName      string   `json:"sign_name" binding:"max=50"`
	VariableNames []string `json:"variable_names"`
	Description   string   `json:"description" binding:"max=500"`
}

type SmsTemplateUpdateDTO struct {
	TemplateName  *string   `json:"template_name" binding:"omitempty,max=100"`
	Content       *string   `json:"content"`
	SignName      *string   `json:"sign_name" binding:"omitempty,max=50"`
	VariableNames *[]string `json:"variable_names"`
	Description   *string   `json:"description" binding:"omitempty,max=500"`
}

type SmsTemplateQueryDTO struct {
	PageQuery
	StoreID      uint   `form:"store_id"`
	TemplateType string `form:"template_type"`
	ReviewStatus string `form:"review_status"`
	IsActive     *bool  `form:"is_active"`
	Keyword      string `form:"keyword"`
}

type SmsTemplateResponse struct {
	ID                   uint       `json:"id"`
	StoreID              uint       `json:"store_id"`
	StoreName            string     `json:"store_name"`
	TemplateCode         string     `json:"template_code"`
	TemplateName         string     `json:"template_name"`
	TemplateType         string     `json:"template_type"`
	Content              string     `json:"content"`
	SignName             string     `json:"sign_name"`
	VariableCount        int        `json:"variable_count"`
	VariableNames        string     `json:"variable_names"`
	ReviewStatus         string     `json:"review_status"`
	ReviewRemark         string     `json:"review_remark"`
	ReviewTime           *time.Time `json:"review_time"`
	ReviewerID           uint       `json:"reviewer_id"`
	ReviewerName         string     `json:"reviewer_name"`
	PlatformTemplateCode string     `json:"platform_template_code"`
	IsActive             bool       `json:"is_active"`
	UsedCount            int        `json:"used_count"`
	Description          string     `json:"description"`
	CreatedAt            time.Time  `json:"created_at"`
	UpdatedAt            time.Time  `json:"updated_at"`
}

type SmsTemplateReviewDTO struct {
	ReviewStatus string `json:"review_status" binding:"required,max=20"`
	ReviewRemark string `json:"review_remark" binding:"max=500"`
	ReviewerID   uint   `json:"reviewer_id"`
	ReviewerName string `json:"reviewer_name" binding:"max=50"`
}

type SmsTaskCreateDTO struct {
	TaskName         string          `json:"task_name" binding:"required,max=100"`
	TaskType         string          `json:"task_type" binding:"max=20"`
	TemplateID       uint            `json:"template_id" binding:"required"`
	SignName         string          `json:"sign_name" binding:"max=50"`
	TargetType       string          `json:"target_type" binding:"max=20"`
	MemberLevelIDs   []uint          `json:"member_level_ids"`
	MinConsumeCount  int             `json:"min_consume_count" binding:"min=0"`
	MaxConsumeCount  int             `json:"max_consume_count" binding:"min=0"`
	MinConsumeAmount decimal.Decimal `json:"min_consume_amount" binding:"min=0"`
	MaxConsumeAmount decimal.Decimal `json:"max_consume_amount" binding:"min=0"`
	MinPoints        int             `json:"min_points" binding:"min=0"`
	MaxPoints        int             `json:"max_points" binding:"min=0"`
	ScheduleType     string          `json:"schedule_type" binding:"max=20"`
	ScheduledTime    *time.Time      `json:"scheduled_time"`
	Remark           string          `json:"remark" binding:"max=500"`
}

type SmsTaskQueryDTO struct {
	PageQuery
	StoreID      uint   `form:"store_id"`
	TaskType     string `form:"task_type"`
	Status       string `form:"status"`
	ScheduleType string `form:"schedule_type"`
	Keyword      string `form:"keyword"`
	StartDate    string `form:"start_date"`
	EndDate      string `form:"end_date"`
}

type SmsTaskResponse struct {
	ID               uint            `json:"id"`
	StoreID          uint            `json:"store_id"`
	StoreName        string          `json:"store_name"`
	TaskName         string          `json:"task_name"`
	TaskType         string          `json:"task_type"`
	TemplateID       uint            `json:"template_id"`
	TemplateName     string          `json:"template_name"`
	TemplateCode     string          `json:"template_code"`
	SignName         string          `json:"sign_name"`
	Content          string          `json:"content"`
	TargetType       string          `json:"target_type"`
	TargetFilters    string          `json:"target_filters"`
	MemberLevelIDs   string          `json:"member_level_ids"`
	MinConsumeCount  int             `json:"min_consume_count"`
	MaxConsumeCount  int             `json:"max_consume_count"`
	MinConsumeAmount decimal.Decimal `json:"min_consume_amount"`
	MaxConsumeAmount decimal.Decimal `json:"max_consume_amount"`
	MinPoints        int             `json:"min_points"`
	MaxPoints        int             `json:"max_points"`
	TargetCount      int             `json:"target_count"`
	SuccessCount     int             `json:"success_count"`
	FailCount        int             `json:"fail_count"`
	ReadCount        int             `json:"read_count"`
	ConversionCount  int             `json:"conversion_count"`
	ConversionAmount decimal.Decimal `json:"conversion_amount"`
	ConversionRate   decimal.Decimal `json:"conversion_rate"`
	SuccessRate      decimal.Decimal `json:"success_rate"`
	ScheduleType     string          `json:"schedule_type"`
	ScheduledTime    *time.Time      `json:"scheduled_time"`
	StartTime        *time.Time      `json:"start_time"`
	EndTime          *time.Time      `json:"end_time"`
	Status           string          `json:"status"`
	CreatorID        uint            `json:"creator_id"`
	CreatorName      string          `json:"creator_name"`
	Remark           string          `json:"remark"`
	CreatedAt        time.Time       `json:"created_at"`
	UpdatedAt        time.Time       `json:"updated_at"`
}

type SmsTaskStatisticsResponse struct {
	TotalTasks       int             `json:"total_tasks"`
	TotalSent        int             `json:"total_sent"`
	SuccessCount     int             `json:"success_count"`
	FailCount        int             `json:"fail_count"`
	SuccessRate      decimal.Decimal `json:"success_rate"`
	ReadCount        int             `json:"read_count"`
	ConversionCount  int             `json:"conversion_count"`
	ConversionRate   decimal.Decimal `json:"conversion_rate"`
	ConversionAmount decimal.Decimal `json:"conversion_amount"`
}

type SmsRecordQueryDTO struct {
	PageQuery
	StoreID   uint   `form:"store_id"`
	TaskID    uint   `form:"task_id"`
	TemplateID uint  `form:"template_id"`
	Status    string `form:"status"`
	Phone     string `form:"phone"`
	SendType  string `form:"send_type"`
	StartDate string `form:"start_date"`
	EndDate   string `form:"end_date"`
}

type SmsRecordResponse struct {
	ID                uint            `json:"id"`
	StoreID           uint            `json:"store_id"`
	StoreName         string          `json:"store_name"`
	TaskID            uint            `json:"task_id"`
	TaskName          string          `json:"task_name"`
	TemplateID        uint            `json:"template_id"`
	TemplateName      string          `json:"template_name"`
	TemplateCode      string          `json:"template_code"`
	SignName          string          `json:"sign_name"`
	Content           string          `json:"content"`
	MemberID          uint            `json:"member_id"`
	MemberName        string          `json:"member_name"`
	Phone             string          `json:"phone"`
	SendType          string          `json:"send_type"`
	Status            string          `json:"status"`
	ErrorCode         string          `json:"error_code"`
	ErrorMessage      string          `json:"error_message"`
	RequestID         string          `json:"request_id"`
	BizID             string          `json:"biz_id"`
	SendTime          *time.Time      `json:"send_time"`
	DeliverTime       *time.Time      `json:"deliver_time"`
	IsRead            bool            `json:"is_read"`
	ReadTime          *time.Time      `json:"read_time"`
	IsConverted       bool            `json:"is_converted"`
	ConversionOrderID uint            `json:"conversion_order_id"`
	ConversionAmount  decimal.Decimal `json:"conversion_amount"`
	ConversionTime    *time.Time      `json:"conversion_time"`
	PricePer          decimal.Decimal `json:"price_per"`
	CostAmount        decimal.Decimal `json:"cost_amount"`
	CreatedAt         time.Time       `json:"created_at"`
	UpdatedAt         time.Time       `json:"updated_at"`
}

type SmsTargetCountDTO struct {
	StoreID          uint            `json:"store_id" binding:"required"`
	TargetType       string          `json:"target_type" binding:"max=20"`
	MemberLevelIDs   []uint          `json:"member_level_ids"`
	MinConsumeCount  int             `json:"min_consume_count" binding:"min=0"`
	MaxConsumeCount  int             `json:"max_consume_count" binding:"min=0"`
	MinConsumeAmount decimal.Decimal `json:"min_consume_amount" binding:"min=0"`
	MaxConsumeAmount decimal.Decimal `json:"max_consume_amount" binding:"min=0"`
	MinPoints        int             `json:"min_points" binding:"min=0"`
	MaxPoints        int             `json:"max_points" binding:"min=0"`
}

type SmsTargetCountResponse struct {
	TargetCount int `json:"target_count"`
}

type SmsSendTestDTO struct {
	Phone      string `json:"phone" binding:"required,max=20"`
	TemplateID uint   `json:"template_id" binding:"required"`
	SignName   string `json:"sign_name" binding:"max=50"`
	Content    string `json:"content"`
}
