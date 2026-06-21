package dto

import (
	"time"

	"github.com/shopspring/decimal"
)

type TimeSlotPricingCreateDTO struct {
	StoreID        uint            `json:"store_id" binding:"required"`
	Name           string          `json:"name" binding:"required,max=100"`
	StartTime      string          `json:"start_time" binding:"required,max=5"`
	EndTime        string          `json:"end_time" binding:"required,max=5"`
	Price          decimal.Decimal `json:"price" binding:"required,min=0"`
	OriginalPrice  decimal.Decimal `json:"original_price" binding:"min=0"`
	DiscountRate   decimal.Decimal `json:"discount_rate" binding:"min=0,max=100"`
	ApplicableType string          `json:"applicable_type" binding:"required,oneof=all category product"`
	ApplicableIDs  []uint          `json:"applicable_ids"`
	Weekdays       []int           `json:"weekdays"`
	MaxReservations int            `json:"max_reservations" binding:"min=0"`
	Status         int             `json:"status" binding:"oneof=0 1"`
	Description    string          `json:"description" binding:"max=500"`
}

type TimeSlotPricingUpdateDTO struct {
	Name           string          `json:"name" binding:"max=100"`
	StartTime      string          `json:"start_time" binding:"max=5"`
	EndTime        string          `json:"end_time" binding:"max=5"`
	Price          decimal.Decimal `json:"price" binding:"min=0"`
	OriginalPrice  decimal.Decimal `json:"original_price" binding:"min=0"`
	DiscountRate   decimal.Decimal `json:"discount_rate" binding:"min=0,max=100"`
	ApplicableType string          `json:"applicable_type" binding:"oneof=all category product"`
	ApplicableIDs  []uint          `json:"applicable_ids"`
	Weekdays       []int           `json:"weekdays"`
	MaxReservations *int           `json:"max_reservations"`
	Status         int             `json:"status" binding:"oneof=0 1"`
	Description    string          `json:"description" binding:"max=500"`
}

type TimeSlotPricingQueryDTO struct {
	PageQuery
	StoreID  uint   `form:"store_id"`
	Name     string `form:"name"`
	Status   int    `form:"status"`
	Weekday  int    `form:"weekday"`
	Date     string `form:"date"`
}

type TimeSlotPricingResponse struct {
	ID              uint            `json:"id"`
	StoreID         uint            `json:"store_id"`
	StoreName       string          `json:"store_name"`
	Name            string          `json:"name"`
	StartTime       string          `json:"start_time"`
	EndTime         string          `json:"end_time"`
	Price           decimal.Decimal `json:"price"`
	OriginalPrice   decimal.Decimal `json:"original_price"`
	DiscountRate    decimal.Decimal `json:"discount_rate"`
	ApplicableType  string          `json:"applicable_type"`
	ApplicableIDs   []uint          `json:"applicable_ids"`
	Weekdays        []int           `json:"weekdays"`
	MaxReservations int             `json:"max_reservations"`
	CurrentReservations int         `json:"current_reservations"`
	Status          int             `json:"status"`
	Description     string          `json:"description"`
	CreatedAt       time.Time       `json:"created_at"`
}

type StockReservationResponse struct {
	ID              uint            `json:"id"`
	OrderID         uint            `json:"order_id"`
	OrderNo         string          `json:"order_no"`
	MemberID        uint            `json:"member_id"`
	MemberName      string          `json:"member_name"`
	MemberPhone     string          `json:"member_phone"`
	TimeSlotID      uint            `json:"time_slot_id"`
	TimeSlotName    string          `json:"time_slot_name"`
	ReservationDate string          `json:"reservation_date"`
	StartTime       string          `json:"start_time"`
	EndTime         string          `json:"end_time"`
	PeopleCount     int             `json:"people_count"`
	TableNo         string          `json:"table_no"`
	DepositAmount   decimal.Decimal `json:"deposit_amount"`
	TotalAmount     decimal.Decimal `json:"total_amount"`
	Status          int             `json:"status"`
	CheckInTime     *time.Time      `json:"check_in_time"`
	Remark          string          `json:"remark"`
	CreatedAt       time.Time       `json:"created_at"`
}

type TimeSlotPriceCalculateRequest struct {
	StoreID         uint            `json:"store_id" binding:"required"`
	ReservationDate string          `json:"reservation_date" binding:"required"`
	TimeSlotID      uint            `json:"time_slot_id" binding:"required"`
	Items           []OrderItemDTO  `json:"items" binding:"required,min=1"`
	MemberID        uint            `json:"member_id"`
}

type TimeSlotPriceCalculateResponse struct {
	TimeSlotID      uint            `json:"time_slot_id"`
	TimeSlotName    string          `json:"time_slot_name"`
	StartTime       string          `json:"start_time"`
	EndTime         string          `json:"end_time"`
	OriginalAmount  decimal.Decimal `json:"original_amount"`
	DiscountAmount  decimal.Decimal `json:"discount_amount"`
	TimeSlotDiscount decimal.Decimal `json:"time_slot_discount"`
	FinalAmount     decimal.Decimal `json:"final_amount"`
	ApplicableItems []uint          `json:"applicable_items"`
	DepositRequired bool            `json:"deposit_required"`
	DepositAmount   decimal.Decimal `json:"deposit_amount"`
	Available       bool            `json:"available"`
	RemainingSlots  int             `json:"remaining_slots"`
}

type ReservationReminderResponse struct {
	ID              uint      `json:"id"`
	OrderID         uint      `json:"order_id"`
	OrderNo         string    `json:"order_no"`
	MemberID        uint      `json:"member_id"`
	MemberName      string    `json:"member_name"`
	MemberPhone     string    `json:"member_phone"`
	TimeSlotName    string    `json:"time_slot_name"`
	ReservationDate string    `json:"reservation_date"`
	StartTime       string    `json:"start_time"`
	EndTime         string    `json:"end_time"`
	TableNo         string    `json:"table_no"`
	PeopleCount     int       `json:"people_count"`
	ReminderType    string    `json:"reminder_type"`
	ReminderTime    time.Time `json:"reminder_time"`
	Status          int       `json:"status"`
}
