package handler

import (
	"net/http"
	"strconv"
	"stalll-hub-pos/backend/internal/service"
	"stalll-hub-pos/backend/pkg/response"

	"github.com/gin-gonic/gin"
)

type StockCheckHandler struct {
	service *service.StockCheckService
}

var stockCheckHandlerInstance *StockCheckHandler

func NewStockCheckHandler() *StockCheckHandler {
	if stockCheckHandlerInstance == nil {
		stockCheckHandlerInstance = &StockCheckHandler{
			service: service.NewStockCheckService(),
		}
	}
	return stockCheckHandlerInstance
}

type CreateCheckReq struct {
	StoreID      uint     `json:"store_id"`
	Title        string   `json:"title"`
	CheckType    string   `json:"check_type"`
	CategoryIDs  []uint   `json:"category_ids"`
	OperatorID   uint     `json:"operator_id"`
	OperatorName string   `json:"operator_name"`
	Remark       string   `json:"remark"`
}

func (h *StockCheckHandler) Create(c *gin.Context) {
	var req CreateCheckReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	if req.StoreID == 0 {
		req.StoreID = 1
	}
	if req.Title == "" {
		response.Error(c, http.StatusBadRequest, "title required")
		return
	}

	serviceReq := &service.CreateStockCheckReq{
		StoreID:      req.StoreID,
		Title:        req.Title,
		CheckType:    req.CheckType,
		CategoryIDs:  req.CategoryIDs,
		OperatorID:   req.OperatorID,
		OperatorName: req.OperatorName,
		Remark:       req.Remark,
	}

	result, err := h.service.Create(serviceReq)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(c, result)
}

func (h *StockCheckHandler) List(c *gin.Context) {
	storeID, _ := strconv.ParseUint(c.Query("store_id"), 10, 32)
	status, _ := strconv.Atoi(c.DefaultQuery("status", "-1"))
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	list, total, err := h.service.List(uint(storeID), status, page, pageSize)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(c, gin.H{
		"list":      list,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

func (h *StockCheckHandler) GetByID(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid id")
		return
	}

	result, err := h.service.GetByID(uint(id))
	if err != nil {
		response.Error(c, http.StatusNotFound, "not found")
		return
	}

	response.Success(c, result)
}

func (h *StockCheckHandler) GetItems(c *gin.Context) {
	checkID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid id")
		return
	}

	status, _ := strconv.Atoi(c.DefaultQuery("status", "-1"))
	keyword := c.Query("keyword")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "50"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 200 {
		pageSize = 50
	}

	items, total, err := h.service.GetItems(uint(checkID), status, keyword, page, pageSize)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(c, gin.H{
		"list":      items,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

type UploadItemsReq struct {
	Items []service.StockCheckItemDTO `json:"items"`
}

func (h *StockCheckHandler) UploadItems(c *gin.Context) {
	checkID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid id")
		return
	}

	var req UploadItemsReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	err = h.service.UploadItems(uint(checkID), req.Items)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(c, nil)
}

type UpdateItemReq struct {
	ActualStock int    `json:"actual_stock"`
	Remark      string `json:"remark"`
}

func (h *StockCheckHandler) UpdateItem(c *gin.Context) {
	itemID, err := strconv.ParseUint(c.Param("item_id"), 10, 32)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid item id")
		return
	}

	var req UpdateItemReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	err = h.service.UpdateItem(uint(itemID), req.ActualStock, req.Remark)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(c, nil)
}

func (h *StockCheckHandler) Complete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid id")
		return
	}

	result, err := h.service.Complete(uint(id))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(c, result)
}

func (h *StockCheckHandler) DiffReport(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid id")
		return
	}

	result, err := h.service.GenerateDiffReport(uint(id))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(c, result)
}

type StockWarningHandler struct {
	dingTalk *service.DingTalkService
}

var stockWarningHandlerInstance *StockWarningHandler

func NewStockWarningHandler() *StockWarningHandler {
	if stockWarningHandlerInstance == nil {
		stockWarningHandlerInstance = &StockWarningHandler{
			dingTalk: service.NewDingTalkService(),
		}
	}
	return stockWarningHandlerInstance
}

func (h *StockWarningHandler) TestDingTalk(c *gin.Context) {
	content := c.Query("content")
	if content == "" {
		content = "钉钉机器人测试消息 - 库存预警系统"
	}

	err := h.dingTalk.SendText(content, false, nil)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(c, nil)
}
