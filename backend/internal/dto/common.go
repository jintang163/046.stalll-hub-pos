package dto

type Pagination struct {
	Page     int `form:"page" json:"page" binding:"min=1"`
	PageSize int `form:"page_size" json:"page_size" binding:"min=1,max=100"`
}

func (p *Pagination) GetOffset() int {
	if p.Page == 0 {
		p.Page = 1
	}
	if p.PageSize == 0 {
		p.PageSize = 10
	}
	return (p.Page - 1) * p.PageSize
}

func (p *Pagination) GetLimit() int {
	if p.PageSize == 0 {
		p.PageSize = 10
	}
	return p.PageSize
}

type IDRequest struct {
	ID uint `uri:"id" binding:"required,min=1"`
}

type IDsRequest struct {
	IDs []uint `json:"ids" binding:"required,min=1"`
}

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type LoginResponse struct {
	Token string      `json:"token"`
	User  interface{} `json:"user"`
}

type StatusUpdateRequest struct {
	Status int `json:"status" binding:"required"`
}
