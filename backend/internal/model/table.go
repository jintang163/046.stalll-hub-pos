package model

import "time"

type Table struct {
	BaseModel
	StoreID      uint      `gorm:"not null;index:idx_store_table" json:"store_id"`
	TableNo      string    `gorm:"size:20;not null;index:idx_store_table" json:"table_no"`
	Name         string    `gorm:"size:50" json:"name"`
	Type         string    `gorm:"size:20;default:normal" json:"type"`
	Capacity     int       `gorm:"default:4" json:"capacity"`
	Floor        int       `gorm:"default:1" json:"floor"`
	Area         string    `gorm:"size:50" json:"area"`
	QRCode       string    `gorm:"size:500" json:"qr_code"`
	QRCodeUrl    string    `gorm:"size:500" json:"qr_code_url"`
	Status       int       `gorm:"default:1" json:"status"`
	CurrentOrderID uint    `gorm:"index" json:"current_order_id"`
	CurrentCustomerCount int `gorm:"default:0" json:"current_customer_count"`
	CheckinTime  *time.Time `json:"checkin_time"`
	OccupiedDuration int  `gorm:"default:0" json:"occupied_duration"`
	TotalOrders  int       `gorm:"default:0" json:"total_orders"`
	TotalAmount  float64   `gorm:"default:0" json:"total_amount"`
	Store        Store     `gorm:"foreignKey:StoreID" json:"store,omitempty"`
}

type TableReservation struct {
	BaseModel
	StoreID      uint       `gorm:"not null;index" json:"store_id"`
	TableID      uint       `gorm:"index" json:"table_id"`
	MemberID     uint       `gorm:"index" json:"member_id"`
	MemberName   string     `gorm:"size:50" json:"member_name"`
	MemberPhone  string     `gorm:"size:20" json:"member_phone"`
	TableNo      string     `gorm:"size:20" json:"table_no"`
	ReserveDate  string     `gorm:"size:20;not null" json:"reserve_date"`
	ReserveTime  string     `gorm:"size:20;not null" json:"reserve_time"`
	PeopleCount  int        `gorm:"default:2" json:"people_count"`
	Status       int        `gorm:"default:1" json:"status"`
	CheckinStatus int       `gorm:"default:0" json:"checkin_status"`
	CheckinTime  *time.Time `json:"checkin_time"`
	CancelTime   *time.Time `json:"cancel_time"`
	Remark       string     `gorm:"size:500" json:"remark"`
	Source       string     `gorm:"size:20;default:wechat" json:"source"`
	OrderID      uint       `json:"order_id"`
	Table        Table      `gorm:"foreignKey:TableID" json:"table,omitempty"`
	Member       Member     `gorm:"foreignKey:MemberID" json:"member,omitempty"`
}

type QueueNumber struct {
	BaseModel
	StoreID      uint       `gorm:"not null;index" json:"store_id"`
	QueueType    string     `gorm:"size:20;default:small" json:"queue_type"`
	QueueNumber  string     `gorm:"size:20;not null" json:"queue_number"`
	Sequence     int        `gorm:"not null" json:"sequence"`
	MemberID     uint       `gorm:"index" json:"member_id"`
	MemberName   string     `gorm:"size:50" json:"member_name"`
	MemberPhone  string     `gorm:"size:20" json:"member_phone"`
	PeopleCount  int        `gorm:"default:2" json:"people_count"`
	Status       int        `gorm:"default:1" json:"status"`
	CallCount    int        `gorm:"default:0" json:"call_count"`
	LastCallTime *time.Time `json:"last_call_time"`
	CallTime     *time.Time `json:"call_time"`
	ArriveTime   *time.Time `json:"arrive_time"`
	CancelTime   *time.Time `json:"cancel_time"`
	AheadCount   int        `gorm:"default:0" json:"ahead_count"`
	WaitDuration int        `gorm:"default:0" json:"wait_duration"`
	Remark       string     `gorm:"size:200" json:"remark"`
	TableID      uint       `json:"table_id"`
	TableNo      string     `gorm:"size:20" json:"table_no"`
	Member       Member     `gorm:"foreignKey:MemberID" json:"member,omitempty"`
}

type QueueConfig struct {
	BaseModel
	StoreID       uint   `gorm:"not null;uniqueIndex" json:"store_id"`
	SmallPrefix   string `gorm:"size:10;default:A" json:"small_prefix"`
	SmallCapacity int    `gorm:"default:4" json:"small_capacity"`
	MediumPrefix  string `gorm:"size:10;default:B" json:"medium_prefix"`
	MediumCapacity int   `gorm:"default:6" json:"medium_capacity"`
	LargePrefix   string `gorm:"size:10;default:C" json:"large_prefix"`
	LargeCapacity  int   `gorm:"default:10" json:"large_capacity"`
	AutoCall      bool   `gorm:"default:true" json:"auto_call"`
	CallInterval  int    `gorm:"default:300" json:"call_interval"`
	MaxCallCount  int    `gorm:"default:3" json:"max_call_count"`
	AutoExpire    bool   `gorm:"default:true" json:"auto_expire"`
	ExpireMinutes int    `gorm:"default:15" json:"expire_minutes"`
	VoiceNotify   bool   `gorm:"default:true" json:"voice_notify"`
	SMSNotify     bool   `gorm:"default:false" json:"sms_notify"`
}

type TableArea struct {
	BaseModel
	StoreID      uint   `gorm:"not null;index" json:"store_id"`
	Name         string `gorm:"size:50;not null" json:"name"`
	SortOrder    int    `gorm:"default:0" json:"sort_order"`
	Status       int    `gorm:"default:1" json:"status"`
}
