package handler

import (
	"net/http"
	"strconv"
	"stalll-hub-pos/backend/internal/dto"
	"stalll-hub-pos/backend/internal/middleware"
	"stalll-hub-pos/backend/internal/service"

	"github.com/gin-gonic/gin"
)

type PaymentHandler struct {
	paymentService *service.PaymentService
}

func NewPaymentHandler() *PaymentHandler {
	return &PaymentHandler{
		paymentService: service.NewPaymentService(),
	}
}

func (h *PaymentHandler) WechatUnifiedOrder(c *gin.Context) {
	var req dto.WechatUnifiedOrderDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.Error(c, "参数错误: "+err.Error())
		return
	}

	if req.StoreID == 0 {
		req.StoreID = middleware.GetStoreID(c)
	}

	resp, err := h.paymentService.WechatUnifiedOrder(&req)
	if err != nil {
		middleware.Error(c, "微信统一下单失败: "+err.Error())
		return
	}

	middleware.Success(c, resp)
}

func (h *PaymentHandler) WechatQueryOrder(c *gin.Context) {
	orderNo := c.Param("orderNo")
	if orderNo == "" {
		middleware.Error(c, "订单号不能为空")
		return
	}

	resp, err := h.paymentService.WechatQueryOrder(orderNo)
	if err != nil {
		middleware.Error(c, "查询微信订单失败: "+err.Error())
		return
	}

	middleware.Success(c, resp)
}

func (h *PaymentHandler) WechatRefund(c *gin.Context) {
	var req dto.WechatRefundDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.Error(c, "参数错误: "+err.Error())
		return
	}

	resp, err := h.paymentService.WechatRefund(&req)
	if err != nil {
		middleware.Error(c, "微信退款失败: "+err.Error())
		return
	}

	middleware.Success(c, resp)
}

func (h *PaymentHandler) WechatNotify(c *gin.Context) {
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

	amount := totalFee / 100

	err := h.paymentService.WechatNotify(orderNo, transactionID, amount, body)
	if err != nil {
		middleware.XMLResponse(c, http.StatusInternalServerError, "FAIL", err.Error())
		return
	}

	middleware.XMLResponse(c, http.StatusOK, "SUCCESS", "OK")
}

func (h *PaymentHandler) WechatRefundNotify(c *gin.Context) {
	var body map[string]interface{}
	if err := c.ShouldBindJSON(&body); err != nil {
		middleware.XMLResponse(c, http.StatusBadRequest, "FAIL", "Invalid request")
		return
	}

	refundID, ok := body["out_refund_no"].(string)
	if !ok {
		middleware.XMLResponse(c, http.StatusBadRequest, "FAIL", "Missing refund number")
		return
	}

	refundFee, _ := body["refund_fee"].(float64)
	refundAmount := refundFee / 100

	err := h.paymentService.WechatRefundNotify(refundID, refundAmount, body)
	if err != nil {
		middleware.XMLResponse(c, http.StatusInternalServerError, "FAIL", err.Error())
		return
	}

	middleware.XMLResponse(c, http.StatusOK, "SUCCESS", "OK")
}
