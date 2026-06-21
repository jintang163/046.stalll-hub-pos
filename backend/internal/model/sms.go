package model

import (
	"time"

	"github.com/shopspring/decimal"
)

type SmsTemplate struct {
	BaseModel
	StoreID        uint   `gorm:"not null;index" json:"store_id"`
	TemplateCode   string `gorm:"size:50;not null;index" json:"template_code"`
	TemplateName   string `gorm:"size:100;not null" json:"template_name"`
	TemplateType   string `gorm:"size:20;default:marketing" json:"template_type"`
	Content        string `gorm:"type:text;not null" json:"content"`
	SignName       string `gorm:"size:50" json:"sign_name"`
	VariableCount  int    `gorm:"default:0" json:"variable_count"`
	VariableNames  string `gorm:"size:500" json:"variable_names"`
	ReviewStatus   string `gorm:"size:20;default:pending;index" json:"review_status"`
	ReviewRemark   string `gorm:"size:500" json:"review_remark"`
	ReviewTime     *time.Time `json:"review_time"`
	ReviewerID     uint   `json:"reviewer_id"`
	ReviewerName   string `gorm:"size:50" json:"reviewer_name"`
	PlatformTemplateCode string `gorm:"size:50" json:"platform_template_code"`
	IsActive       bool   `gorm:"default:false" json:"is_active"`
	UsedCount      int    `gorm:"default:0" json:"used_count"`
	Description    string `gorm:"size:500" json:"description"`
	Store          Store  `gorm:"foreignKey:StoreID" json:"store,omitempty"`
}

type SmsTask struct {
	BaseModel
	StoreID        uint      `gorm:"not null;index" json:"store_id"`
	TaskName       string    `gorm:"size:100;not null" json:"task_name"`
	TaskType       string    `gorm:"size:20;default:marketing" json:"task_type"`
	TemplateID     uint      `gorm:"not null" json:"template_id"`
	TemplateCode   string    `gorm:"size:50" json:"template_code"`
	SignName       string    `gorm:"size:50" json:"sign_name"`
	Content        string    `gorm:"type:text" json:"content"`
	TargetType     string    `gorm:"size:20;default:custom" json:"target_type"`
	TargetFilters  string    `gorm:"type:text" json:"target_filters"`
	MemberLevelIDs string    `gorm:"size:200" json:"member_level_ids"`
	MinConsumeCount int     `gorm:"default:0" json:"min_consume_count"`
	MaxConsumeCount int     `gorm:"default:0" json:"max_consume_count"`
	MinConsumeAmount decimal.Decimal `gorm:"type:decimal(10,2);default:0" json:"min_consume_amount"`
	MaxConsumeAmount decimal.Decimal `gorm:"type:decimal(10,2);default:0" json:"max_consume_amount"`
	MinPoints      int       `gorm:"default:0" json:"min_points"`
	MaxPoints      int       `gorm:"default:0" json:"max_points"`
	TargetCount    int       `gorm:"default:0" json:"target_count"`
	SuccessCount   int       `gorm:"default:0" json:"success_count"`
	FailCount      int       `gorm:"default:0" json:"fail_count"`
	ReadCount      int       `gorm:"default:0" json:"read_count"`
	ConversionCount int      `gorm:"default:0" json:"conversion_count"`
	ConversionAmount decimal.Decimal `gorm:"type:decimal(10,2);default:0" json:"conversion_amount"`
	ConversionRate decimal.Decimal `gorm:"type:decimal(5,2);default:0" json:"conversion_rate"`
	SuccessRate    decimal.Decimal `gorm:"type:decimal(5,2);default:0" json:"success_rate"`
	ScheduleType   string    `gorm:"size:20;default:immediately" json:"schedule_type"`
	ScheduledTime  *time.Time `json:"scheduled_time"`
	StartTime      *time.Time `json:"start_time"`
	EndTime        *time.Time `json:"end_time"`
	Status         string    `gorm:"size:20;default:draft;index" json:"status"`
	CreatorID      uint      `json:"creator_id"`
	CreatorName    string    `gorm:"size:50" json:"creator_name"`
	Remark         string    `gorm:"size:500" json:"remark"`
	Store          Store     `gorm:"foreignKey:StoreID" json:"store,omitempty"`
	Template       *SmsTemplate `gorm:"foreignKey:TemplateID" json:"template,omitempty"`
}

type SmsRecord struct {
	BaseModel
	StoreID       uint      `gorm:"not null;index" json:"store_id"`
	TaskID        uint      `gorm:"index" json:"task_id"`
	TemplateID    uint      `json:"template_id"`
	TemplateCode  string    `gorm:"size:50" json:"template_code"`
	SignName      string    `gorm:"size:50" json:"sign_name"`
	Content       string    `gorm:"type:text" json:"content"`
	MemberID      uint      `gorm:"index" json:"member_id"`
	MemberName    string    `gorm:"size:50" json:"member_name"`
	Phone         string    `gorm:"size:20;not null;index" json:"phone"`
	SendType      string    `gorm:"size:20;default:marketing" json:"send_type"`
	Status        string    `gorm:"size:20;default:pending;index" json:"status"`
	ErrorCode     string    `gorm:"size:50" json:"error_code"`
	ErrorMessage  string    `gorm:"size:200" json:"error_message"`
	RequestID     string    `gorm:"size:100" json:"request_id"`
	BizID         string    `gorm:"size:100" json:"biz_id"`
	SendTime      *time.Time `json:"send_time"`
	DeliverTime   *time.Time `json:"deliver_time"`
	IsRead        bool      `gorm:"default:false" json:"is_read"`
	ReadTime      *time.Time `json:"read_time"`
	IsConverted   bool      `gorm:"default:false" json:"is_converted"`
	ConversionOrderID uint   `json:"conversion_order_id"`
	ConversionAmount decimal.Decimal `gorm:"type:decimal(10,2);default:0" json:"conversion_amount"`
	ConversionTime *time.Time `json:"conversion_time"`
	PricePer      decimal.Decimal `gorm:"type:decimal(10,4);default:0" json:"price_per"`
	CostAmount    decimal.Decimal `gorm:"type:decimal(10,2);default:0" json:"cost_amount"`
	Store         Store     `gorm:"foreignKey:StoreID" json:"store,omitempty"`
	Task          *SmsTask  `gorm:"foreignKey:TaskID" json:"task,omitempty"`
	Template      *SmsTemplate `gorm:"foreignKey:TemplateID" json:"template,omitempty"`
	Member        *Member   `gorm:"foreignKey:MemberID" json:"member,omitempty"`
}
