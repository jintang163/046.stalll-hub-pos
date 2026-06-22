package dto

type ReceiptAdDTO struct {
	ID            uint   `json:"id"`
	Title         string `json:"title" binding:"required,max=100"`
	AdType        string `json:"ad_type" binding:"required,oneof=image qrcode text"`
	ImageURL      string `json:"image_url"`
	QRCodeContent string `json:"qr_code_content"`
	LinkURL       string `json:"link_url"`
	Content       string `json:"content"`
	Subtitle      string `json:"subtitle"`
	Position      string `json:"position"`
	SortOrder     int    `json:"sort_order"`
	Status        int    `json:"status"`
	StartDate     string `json:"start_date"`
	EndDate       string `json:"end_date"`
	StartTime     string `json:"start_time"`
	EndTime       string `json:"end_time"`
	Remark        string `json:"remark"`
}

type ReceiptAdListRequest struct {
	Page     int    `form:"page,default=1"`
	PageSize int    `form:"page_size,default=20"`
	Status   int    `form:"status"`
	Position string `form:"position"`
	AdType   string `form:"ad_type"`
	Keyword  string `form:"keyword"`
}

type ReceiptAdResponse struct {
	ID            uint   `json:"id"`
	StoreID       uint   `json:"store_id"`
	Title         string `json:"title"`
	AdType        string `json:"ad_type"`
	AdTypeText    string `json:"ad_type_text"`
	ImageURL      string `json:"image_url"`
	QRCodeContent string `json:"qr_code_content"`
	LinkURL       string `json:"link_url"`
	Content       string `json:"content"`
	Subtitle      string `json:"subtitle"`
	Position      string `json:"position"`
	PositionText  string `json:"position_text"`
	SortOrder     int    `json:"sort_order"`
	Status        int    `json:"status"`
	StatusText    string `json:"status_text"`
	ViewCount     int    `json:"view_count"`
	ClickCount    int    `json:"click_count"`
	StartDate     string `json:"start_date"`
	EndDate       string `json:"end_date"`
	StartTime     string `json:"start_time"`
	EndTime       string `json:"end_time"`
	Remark        string `json:"remark"`
	CreatedAt     string `json:"created_at"`
}

type ReceiptAdClickRequest struct {
	AdID      uint   `json:"ad_id" binding:"required"`
	OrderID   uint   `json:"order_id"`
	OrderNo   string `json:"order_no"`
	ClickType string `json:"click_type"`
	IP        string `json:"ip"`
	UserAgent string `json:"user_agent"`
}

type ReceiptAdStatsRequest struct {
	StartDate string `form:"start_date" binding:"required"`
	EndDate   string `form:"end_date" binding:"required"`
	AdID      uint   `form:"ad_id"`
}
