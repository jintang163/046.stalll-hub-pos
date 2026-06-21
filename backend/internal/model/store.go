package model

import "time"

type Store struct {
	BaseModel
	Name          string     `gorm:"size:100;not null" json:"name"`
	Address       string     `gorm:"size:255" json:"address"`
	Phone         string     `gorm:"size:20" json:"phone"`
	BusinessHours string     `gorm:"size:100" json:"business_hours"`
	Status        int        `gorm:"default:1" json:"status"`
	Description   string     `gorm:"size:500" json:"description"`
	Logo          string     `gorm:"size:255" json:"logo"`
	Latitude      string     `gorm:"size:50" json:"latitude"`
	Longitude     string     `gorm:"size:50" json:"longitude"`
	OpenTime      string     `gorm:"size:10" json:"open_time"`
	CloseTime     string     `gorm:"size:10" json:"close_time"`
	Province      string     `gorm:"size:50" json:"province"`
	City          string     `gorm:"size:50" json:"city"`
	District      string     `gorm:"size:50" json:"district"`
	AreaCode      string     `gorm:"size:20" json:"area_code"`
	TablesCount   int        `gorm:"default:0" json:"tables_count"`
	QueueEnabled  bool       `gorm:"default:true" json:"queue_enabled"`
	ReserveEnabled bool      `gorm:"default:true" json:"reserve_enabled"`
}

type Printer struct {
	BaseModel
	StoreID     uint   `gorm:"not null;index" json:"store_id"`
	Name        string `gorm:"size:50;not null" json:"name"`
	Type        string `gorm:"size:20;not null" json:"type"`
	IPAddress   string `gorm:"size:50" json:"ip_address"`
	PrintType   string `gorm:"size:20;default:kitchen" json:"print_type"`
	Status      int    `gorm:"default:1" json:"status"`
	IsDefault   bool   `gorm:"default:false" json:"is_default"`
	Store       Store  `gorm:"foreignKey:StoreID" json:"store,omitempty"`
}

type StoreUser struct {
	BaseModel
	StoreID  uint   `gorm:"not null;index" json:"store_id"`
	Username string `gorm:"size:50;not null;unique" json:"username"`
	Password string `gorm:"size:255;not null" json:"-"`
	RealName string `gorm:"size:50" json:"real_name"`
	Phone    string `gorm:"size:20" json:"phone"`
	Role     string `gorm:"size:20;default:staff" json:"role"`
	Status   int    `gorm:"default:1" json:"status"`
	Store    Store  `gorm:"foreignKey:StoreID" json:"store,omitempty"`
}

type WaiterCall struct {
	BaseModel
	StoreID   uint   `gorm:"not null;index" json:"store_id"`
	TableID   uint   `gorm:"not null;index" json:"table_id"`
	TableNo   string `gorm:"size:20;not null" json:"table_no"`
	Content   string `gorm:"size:500" json:"content"`
	CallType  string `gorm:"size:20;default:service" json:"call_type"`
	Status    int    `gorm:"default:1" json:"status"`
	HandlerID uint   `gorm:"index" json:"handler_id"`
	HandleTime *time.Time `json:"handle_time"`
	Table     Table  `gorm:"foreignKey:TableID" json:"table,omitempty"`
}
