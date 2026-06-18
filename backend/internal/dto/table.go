package dto

type TableCreateDTO struct {
	StoreID   uint   `json:"store_id" binding:"required"`
	TableNo   string `json:"table_no" binding:"required,max=20"`
	Name      string `json:"name" binding:"max=50"`
	Type      string `json:"type" binding:"oneof=normal booth round square private"`
	Capacity  int    `json:"capacity" binding:"min=1,max=30"`
	Floor     int    `json:"floor" binding:"min=1"`
	Area      string `json:"area" binding:"max=50"`
	QRCode    string `json:"qr_code"`
	QRCodeUrl string `json:"qr_code_url"`
	Status    int    `json:"status" binding:"oneof=0 1"`
}

type TableUpdateDTO struct {
	TableNo   string `json:"table_no" binding:"max=20"`
	Name      string `json:"name" binding:"max=50"`
	Type      string `json:"type" binding:"omitempty,oneof=normal booth round square private"`
	Capacity  int    `json:"capacity" binding:"omitempty,min=1,max=30"`
	Floor     int    `json:"floor" binding:"omitempty,min=1"`
	Area      string `json:"area" binding:"max=50"`
	Status    int    `json:"status" binding:"omitempty,oneof=0 1"`
}

type TableQueryDTO struct {
	StoreID  uint   `form:"store_id"`
	Status   int    `form:"status"`
	Floor    int    `form:"floor"`
	Area     string `form:"area"`
	Type     string `form:"type"`
	Keyword  string `form:"keyword"`
	PageNum  int    `form:"page_num,default=1"`
	PageSize int    `form:"page_size,default=20"`
}

type TableOccupiedInfo struct {
	ID                  uint    `json:"id"`
	TableNo             string  `json:"table_no"`
	Name                string  `json:"name"`
	Capacity            int     `json:"capacity"`
	CurrentCustomerCount int    `json:"current_customer_count"`
	CheckinTime         string  `json:"checkin_time"`
	OccupiedDuration    int     `json:"occupied_duration"`
	CurrentOrderID      uint    `json:"current_order_id"`
	CurrentOrderAmount  float64 `json:"current_order_amount"`
	Status              int     `json:"status"`
}

type TableQRCodeDTO struct {
	TableID uint `json:"table_id" binding:"required"`
	Scene   string `json:"scene"`
	Page    string `json:"page"`
	Width   int    `json:"width,default=430"`
}

type ReservationCreateDTO struct {
	StoreID     uint   `json:"store_id" binding:"required"`
	TableID     uint   `json:"table_id"`
	MemberID    uint   `json:"member_id"`
	MemberName  string `json:"member_name" binding:"required,max=50"`
	MemberPhone string `json:"member_phone" binding:"required,max=20"`
	TableNo     string `json:"table_no"`
	ReserveDate string `json:"reserve_date" binding:"required"`
	ReserveTime string `json:"reserve_time" binding:"required"`
	PeopleCount int    `json:"people_count" binding:"min=1,max=30"`
	Remark      string `json:"remark" binding:"max=500"`
	Source      string `json:"source"`
}

type ReservationUpdateDTO struct {
	Status      int    `json:"status" binding:"oneof=1 2 3 4"`
	TableID     uint   `json:"table_id"`
	TableNo     string `json:"table_no"`
	ReserveDate string `json:"reserve_date"`
	ReserveTime string `json:"reserve_time"`
	PeopleCount int    `json:"people_count" binding:"omitempty,min=1,max=30"`
	Remark      string `json:"remark" binding:"max=500"`
}

type ReservationQueryDTO struct {
	StoreID     uint   `form:"store_id"`
	MemberID    uint   `form:"member_id"`
	Status      int    `form:"status"`
	ReserveDate string `form:"reserve_date"`
	CheckinStatus int  `form:"checkin_status"`
	Keyword     string `form:"keyword"`
	PageNum     int    `form:"page_num,default=1"`
	PageSize    int    `form:"page_size,default=20"`
}

type ReservationTimeSlotDTO struct {
	StoreID     uint   `form:"store_id" binding:"required"`
	ReserveDate string `form:"reserve_date" binding:"required"`
	PeopleCount int    `form:"people_count,default=2"`
}

type TimeSlotInfo struct {
	Time      string `json:"time"`
	Available int    `json:"available"`
	Total     int    `json:"total"`
	Status    int    `json:"status"`
}

type QueueCreateDTO struct {
	StoreID     uint   `json:"store_id" binding:"required"`
	QueueType   string `json:"queue_type" binding:"oneof=small medium large"`
	MemberID    uint   `json:"member_id"`
	MemberName  string `json:"member_name" binding:"required,max=50"`
	MemberPhone string `json:"member_phone" binding:"required,max=20"`
	PeopleCount int    `json:"people_count" binding:"min=1,max=30"`
	Remark      string `json:"remark" binding:"max=200"`
}

type QueueCallDTO struct {
	QueueID uint `json:"queue_id" binding:"required"`
}

type QueueCancelDTO struct {
	QueueID uint   `json:"queue_id" binding:"required"`
	Reason  string `json:"reason" binding:"max=200"`
}

type QueueQueryDTO struct {
	StoreID   uint   `form:"store_id"`
	QueueType string `form:"queue_type"`
	Status    int    `form:"status"`
	MemberID  uint   `form:"member_id"`
	Keyword   string `form:"keyword"`
	PageNum   int    `form:"page_num,default=1"`
	PageSize  int    `form:"page_size,default=20"`
}

type QueueStatusDTO struct {
	StoreID     uint   `form:"store_id" binding:"required"`
	QueueType   string `form:"queue_type"`
	MemberID    uint   `form:"member_id"`
	QueueNumber string `form:"queue_number"`
}

type QueueInfoDTO struct {
	QueueNumber string `json:"queue_number"`
	QueueType   string `json:"queue_type"`
	Status      int    `json:"status"`
	Sequence    int    `json:"sequence"`
	AheadCount  int    `json:"ahead_count"`
	WaitTime    int    `json:"wait_time"`
	PeopleCount int    `json:"people_count"`
	CreatedAt   string `json:"created_at"`
}

type QueueConfigDTO struct {
	StoreID        uint `json:"store_id" binding:"required"`
	SmallPrefix    string `json:"small_prefix" binding:"max=10"`
	SmallCapacity  int    `json:"small_capacity" binding:"min=1"`
	MediumPrefix   string `json:"medium_prefix" binding:"max=10"`
	MediumCapacity int    `json:"medium_capacity" binding:"min=1"`
	LargePrefix    string `json:"large_prefix" binding:"max=10"`
	LargeCapacity  int    `json:"large_capacity" binding:"min=1"`
	AutoCall       bool   `json:"auto_call"`
	CallInterval   int    `json:"call_interval" binding:"min=60"`
	MaxCallCount   int    `json:"max_call_count" binding:"min=1"`
	AutoExpire     bool   `json:"auto_expire"`
	ExpireMinutes  int    `json:"expire_minutes" binding:"min=1"`
	VoiceNotify    bool   `json:"voice_notify"`
	SMSNotify      bool   `json:"sms_notify"`
}

type TableAreaCreateDTO struct {
	StoreID   uint   `json:"store_id" binding:"required"`
	Name      string `json:"name" binding:"required,max=50"`
	SortOrder int    `json:"sort_order"`
	Status    int    `json:"status" binding:"oneof=0 1"`
}

type TableAreaUpdateDTO struct {
	Name      string `json:"name" binding:"max=50"`
	SortOrder int    `json:"sort_order"`
	Status    int    `json:"status" binding:"omitempty,oneof=0 1"`
}

type TableCheckinDTO struct {
	TableID         uint `json:"table_id" binding:"required"`
	PeopleCount     int  `json:"people_count" binding:"min=1"`
	ReservationID   uint `json:"reservation_id"`
	MemberID        uint `json:"member_id"`
}

type TableCheckoutDTO struct {
	TableID uint `json:"table_id" binding:"required"`
	OrderID uint `json:"order_id"`
}

type TableScanDTO struct {
	Scene string `json:"scene" binding:"required"`
}

type ScanResultDTO struct {
	StoreID  uint   `json:"store_id"`
	StoreName string `json:"store_name"`
	TableID  uint   `json:"table_id"`
	TableNo  string `json:"table_no"`
	TableType string `json:"table_type"`
	Capacity int    `json:"capacity"`
	Area     string `json:"area"`
	Floor    int    `json:"floor"`
	Status   int    `json:"status"`
	Message  string `json:"message"`
}

type StoreMapDTO struct {
	ID        uint    `json:"id"`
	Name      string  `json:"name"`
	Address   string  `json:"address"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Status    int     `json:"status"`
	TablesCount int   `json:"tables_count"`
	OpenTime  string  `json:"open_time"`
	CloseTime string  `json:"close_time"`
	Phone     string  `json:"phone"`
}
