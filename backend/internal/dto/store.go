package dto

type StoreCreateDTO struct {
	Name          string `json:"name" binding:"required,max=100"`
	Address       string `json:"address" binding:"required,max=255"`
	Phone         string `json:"phone" binding:"required,max=20"`
	BusinessHours string `json:"business_hours" binding:"max=100"`
	Description   string `json:"description" binding:"max=500"`
	Logo          string `json:"logo" binding:"max=255"`
	Latitude      string `json:"latitude" binding:"max=50"`
	Longitude     string `json:"longitude" binding:"max=50"`
	Status        int    `json:"status" binding:"oneof=0 1"`
}

type StoreUpdateDTO struct {
	Name          string `json:"name" binding:"max=100"`
	Address       string `json:"address" binding:"max=255"`
	Phone         string `json:"phone" binding:"max=20"`
	BusinessHours string `json:"business_hours" binding:"max=100"`
	Description   string `json:"description" binding:"max=500"`
	Logo          string `json:"logo" binding:"max=255"`
	Latitude      string `json:"latitude" binding:"max=50"`
	Longitude     string `json:"longitude" binding:"max=50"`
	Status        int    `json:"status" binding:"oneof=0 1"`
}

type StoreQueryDTO struct {
	PageQuery
	Name   string `form:"name"`
	Status int    `form:"status"`
}

type StoreResponse struct {
	ID            uint   `json:"id"`
	Name          string `json:"name"`
	Address       string `json:"address"`
	Phone         string `json:"phone"`
	BusinessHours string `json:"business_hours"`
	Description   string `json:"description"`
	Logo          string `json:"logo"`
	Latitude      string `json:"latitude"`
	Longitude     string `json:"longitude"`
	Status        int    `json:"status"`
	CreatedAt     string `json:"created_at"`
}

type PrinterCreateDTO struct {
	StoreID     uint   `json:"store_id" binding:"required"`
	Name        string `json:"name" binding:"required,max=50"`
	Type        string `json:"type" binding:"required,oneof=kitchen bar receipt"`
	IPAddress   string `json:"ip_address" binding:"required,max=50"`
	Port        int    `json:"port" binding:"required,min=1,max=65535"`
	Width       int    `json:"width" binding:"required,oneof=58 80"`
	AutoCut     bool   `json:"auto_cut"`
	PrintHeader string `json:"print_header" binding:"max=200"`
	PrintFooter string `json:"print_footer" binding:"max=200"`
	Copies      int    `json:"copies" binding:"min=1,max=5"`
	Status      int    `json:"status" binding:"oneof=0 1"`
}

type PrinterUpdateDTO struct {
	Name        string `json:"name" binding:"max=50"`
	Type        string `json:"type" binding:"oneof=kitchen bar receipt"`
	IPAddress   string `json:"ip_address" binding:"max=50"`
	Port        int    `json:"port" binding:"min=1,max=65535"`
	Width       int    `json:"width" binding:"oneof=58 80"`
	AutoCut     *bool  `json:"auto_cut"`
	PrintHeader string `json:"print_header" binding:"max=200"`
	PrintFooter string `json:"print_footer" binding:"max=200"`
	Copies      int    `json:"copies" binding:"min=1,max=5"`
	Status      int    `json:"status" binding:"oneof=0 1"`
}

type PrinterQueryDTO struct {
	PageQuery
	StoreID uint   `form:"store_id"`
	Type    string `form:"type"`
	Status  int    `form:"status"`
}

type PrinterResponse struct {
	ID          uint   `json:"id"`
	StoreID     uint   `json:"store_id"`
	StoreName   string `json:"store_name"`
	Name        string `json:"name"`
	Type        string `json:"type"`
	IPAddress   string `json:"ip_address"`
	Port        int    `json:"port"`
	Width       int    `json:"width"`
	AutoCut     bool   `json:"auto_cut"`
	PrintHeader string `json:"print_header"`
	PrintFooter string `json:"print_footer"`
	Copies      int    `json:"copies"`
	Status      int    `json:"status"`
	CreatedAt   string `json:"created_at"`
}

type PrintTestDTO struct {
	PrinterID uint `json:"printer_id" binding:"required"`
}
