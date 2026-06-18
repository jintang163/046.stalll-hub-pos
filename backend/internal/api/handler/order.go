package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
	"stalll-hub-pos/backend/internal/dto"
	"stalll-hub-pos/backend/internal/middleware"
	"stalll-hub-pos/backend/internal/service"
)

type OrderHandler struct {
	orderService *service.OrderService
}

func NewOrderHandler(orderService *service.OrderService) *OrderHandler {
	return &OrderHandler{orderService: orderService}
}

func (h *OrderHandler) Create(c *gin.Context) {
	var req dto.CreateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.Error(c, http.StatusBadRequest, "Invalid request: "+err.Error())
		return
	}

	resp, err := h.orderService.Create(&req)
	if err != nil {
		middleware.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	middleware.Success(c, resp)
}

func (h *OrderHandler) GetByID(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		middleware.Error(c, http.StatusBadRequest, "Invalid order ID")
		return
	}

	order, err := h.orderService.GetByID(uint(id))
	if err != nil {
		middleware.Error(c, http.StatusNotFound, "Order not found")
		return
	}

	middleware.Success(c, order)
}

func (h *OrderHandler) GetByOrderNo(c *gin.Context) {
	orderNo := c.Param("orderNo")
	if orderNo == "" {
		middleware.Error(c, http.StatusBadRequest, "Order number is required")
		return
	}

	order, err := h.orderService.GetByOrderNo(orderNo)
	if err != nil {
		middleware.Error(c, http.StatusNotFound, "Order not found")
		return
	}

	middleware.Success(c, order)
}

func (h *OrderHandler) List(c *gin.Context) {
	var query dto.OrderQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		middleware.Error(c, http.StatusBadRequest, "Invalid query parameters: "+err.Error())
		return
	}

	if query.Page <= 0 {
		query.Page = 1
	}
	if query.PageSize <= 0 {
		query.PageSize = 10
	}

	pageResp, err := h.orderService.List(&query)
	if err != nil {
		middleware.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	middleware.Success(c, pageResp)
}

func (h *OrderHandler) UpdateStatus(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		middleware.Error(c, http.StatusBadRequest, "Invalid order ID")
		return
	}

	var req dto.UpdateOrderStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.Error(c, http.StatusBadRequest, "Invalid request: "+err.Error())
		return
	}

	err = h.orderService.UpdateStatus(uint(id), req.OrderStatus)
	if err != nil {
		middleware.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	middleware.Success(c, gin.H{"message": "Status updated successfully"})
}

func (h *OrderHandler) Cancel(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		middleware.Error(c, http.StatusBadRequest, "Invalid order ID")
		return
	}

	var req dto.CancelOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.Error(c, http.StatusBadRequest, "Invalid request: "+err.Error())
		return
	}

	err = h.orderService.Cancel(uint(id), req.Reason)
	if err != nil {
		middleware.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	middleware.Success(c, gin.H{"message": "Order cancelled successfully"})
}

func (h *OrderHandler) Refund(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		middleware.Error(c, http.StatusBadRequest, "Invalid order ID")
		return
	}

	var req dto.RefundOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.Error(c, http.StatusBadRequest, "Invalid request: "+err.Error())
		return
	}

	refundID, err := h.orderService.Refund(uint(id), &req)
	if err != nil {
		middleware.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	middleware.Success(c, gin.H{
		"message":   "Refund request submitted successfully",
		"refund_id": refundID,
	})
}

func (h *OrderHandler) GetPaymentParams(c *gin.Context) {
	var req dto.PaymentParamsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.Error(c, http.StatusBadRequest, "Invalid request: "+err.Error())
		return
	}

	params, err := h.orderService.GetPaymentParams(&req)
	if err != nil {
		middleware.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	middleware.Success(c, params)
}

func (h *OrderHandler) WechatNotify(c *gin.Context) {
	var body map[string]interface{}
	if err := c.ShouldBindJSON(&body); err != nil {
		middleware.XMLResponse(c, http.StatusBadRequest, "FAIL", "Invalid request")
		return
	}

	orderNo, ok := body["out_trade_no"].(string)
	if !ok {
		middleware.XMLResponse(c, http.StatusBadRequest, "FAIL", "Missing order number")
		return
	}

	transactionID, _ := body["transaction_id"].(string)
	totalFee, _ := body["total_fee"].(float64)

	amount := decimal.NewFromFloat(totalFee / 100)

	err := h.orderService.NotifyPayment(orderNo, "wechat", transactionID, amount)
	if err != nil {
		middleware.XMLResponse(c, http.StatusInternalServerError, "FAIL", err.Error())
		return
	}

	middleware.XMLResponse(c, http.StatusOK, "SUCCESS", "OK")
}

func (h *OrderHandler) AlipayNotify(c *gin.Context) {
	orderNo := c.PostForm("out_trade_no")
	if orderNo == "" {
		c.String(http.StatusBadRequest, "fail")
		return
	}

	transactionID := c.PostForm("trade_no")
	totalAmount := c.PostForm("total_amount")

	amount, err := decimal.NewFromString(totalAmount)
	if err != nil {
		c.String(http.StatusBadRequest, "fail")
		return
	}

	err = h.orderService.NotifyPayment(orderNo, "alipay", transactionID, amount)
	if err != nil {
		c.String(http.StatusInternalServerError, "fail")
		return
	}

	c.String(http.StatusOK, "success")
}

func (h *OrderHandler) BatchCreate(c *gin.Context) {
	var req dto.BatchOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.Error(c, http.StatusBadRequest, "Invalid request: "+err.Error())
		return
	}

	resp, err := h.orderService.BatchCreate(&req)
	if err != nil {
		middleware.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	middleware.Success(c, resp)
}

func (h *OrderHandler) GetIncremental(c *gin.Context) {
	lastID, err := strconv.ParseUint(c.DefaultQuery("last_id", "0"), 10, 32)
	if err != nil {
		middleware.Error(c, http.StatusBadRequest, "Invalid last ID")
		return
	}

	limit, err := strconv.Atoi(c.DefaultQuery("limit", "100"))
	if err != nil {
		limit = 100
	}

	orders, err := h.orderService.GetIncrementalOrders(uint(lastID), limit)
	if err != nil {
		middleware.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	middleware.Success(c, gin.H{
		"orders":   orders,
		"count":    len(orders),
		"last_id":  lastID,
		"has_more": len(orders) == limit,
	})
}

func (h *OrderHandler) GetForSync(c *gin.Context) {
	storeID, err := strconv.ParseUint(c.DefaultQuery("store_id", "0"), 10, 32)
	if err != nil {
		middleware.Error(c, http.StatusBadRequest, "Invalid store ID")
		return
	}

	lastSyncID, err := strconv.ParseUint(c.DefaultQuery("last_sync_id", "0"), 10, 32)
	if err != nil {
		middleware.Error(c, http.StatusBadRequest, "Invalid last sync ID")
		return
	}

	limit, err := strconv.Atoi(c.DefaultQuery("limit", "100"))
	if err != nil {
		limit = 100
	}

	orders, err := h.orderService.GetOrdersForSync(uint(storeID), uint(lastSyncID), limit)
	if err != nil {
		middleware.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	middleware.Success(c, gin.H{
		"orders":        orders,
		"count":         len(orders),
		"last_sync_id":  lastSyncID,
		"has_more":      len(orders) == limit,
	})
}

func (h *OrderHandler) CreateQueue(c *gin.Context) {
	var body map[string]interface{}
	if err := c.ShouldBindJSON(&body); err != nil {
		middleware.Error(c, http.StatusBadRequest, "Invalid request: "+err.Error())
		return
	}

	storeID, ok := body["store_id"].(float64)
	if !ok {
		middleware.Error(c, http.StatusBadRequest, "Store ID is required")
		return
	}

	orderData, err := json.Marshal(body["order"])
	if err != nil {
		middleware.Error(c, http.StatusBadRequest, "Invalid order data")
		return
	}

	err = h.orderService.CreateOrderQueue(uint(storeID), string(orderData))
	if err != nil {
		middleware.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	middleware.Success(c, gin.H{"message": "Order queued successfully"})
}

func (h *OrderHandler) ProcessQueues(c *gin.Context) {
	storeID, err := strconv.ParseUint(c.Param("storeId"), 10, 32)
	if err != nil {
		middleware.Error(c, http.StatusBadRequest, "Invalid store ID")
		return
	}

	limit, err := strconv.Atoi(c.DefaultQuery("limit", "10"))
	if err != nil {
		limit = 10
	}

	err = h.orderService.ProcessPendingQueues(uint(storeID), limit)
	if err != nil {
		middleware.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	middleware.Success(c, gin.H{"message": "Queues processed successfully"})
}
