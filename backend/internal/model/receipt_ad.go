package model

import (
	"time"
)

type ReceiptAd struct {
	BaseModel
	StoreID       uint      `gorm:"not null;index" json:"store_id"`
	Title         string    `gorm:"size:100;not null" json:"title"`
	AdType        string    `gorm:"size:20;default:image" json:"ad_type"`
	ImageURL      string    `gorm:"size:255" json:"image_url"`
	QRCodeContent string    `gorm:"size:500" json:"qr_code_content"`
	LinkURL       string    `gorm:"size:255" json:"link_url"`
	Content       string    `gorm:"size:500" json:"content"`
	Subtitle      string    `gorm:"size:100" json:"subtitle"`
	Position      string    `gorm:"size:20;default:footer" json:"position"`
	SortOrder     int       `gorm:"default:0;index" json:"sort_order"`
	Status        int       `gorm:"default:1;index" json:"status"`
	ViewCount     int       `gorm:"default:0" json:"view_count"`
	ClickCount    int       `gorm:"default:0" json:"click_count"`
	StartDate     string    `gorm:"size:10;index" json:"start_date"`
	EndDate       string    `gorm:"size:10;index" json:"end_date"`
	StartTime     string    `gorm:"size:5" json:"start_time"`
	EndTime       string    `gorm:"size:5" json:"end_time"`
	Remark        string    `gorm:"size:255" json:"remark"`
	Store         Store     `gorm:"foreignKey:StoreID" json:"store,omitempty"`
}

type ReceiptAdClick struct {
	BaseModel
	StoreID   uint      `gorm:"not null;index" json:"store_id"`
	AdID      uint      `gorm:"not null;index" json:"ad_id"`
	OrderID   uint      `gorm:"index" json:"order_id"`
	OrderNo   string    `gorm:"size:50;index" json:"order_no"`
	ClickType string    `gorm:"size:20;default:scan" json:"click_type"`
	IP        string    `gorm:"size:50" json:"ip"`
	UserAgent string    `gorm:"size:255" json:"user_agent"`
	Ad        ReceiptAd `gorm:"foreignKey:AdID" json:"ad,omitempty"`
}

type ReceiptAdStats struct {
	Date       string `json:"date"`
	ViewCount  int    `json:"view_count"`
	ClickCount int    `json:"click_count"`
}
