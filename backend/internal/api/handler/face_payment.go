package handler

import (
	"encoding/json"
	"fmt"
	"time"

	"stalll-hub-pos/backend/internal/dto"
	"stalll-hub-pos/backend/internal/middleware"
	"stalll-hub-pos/backend/internal/model"
	"stalll-hub-pos/backend/internal/repository"
	"stalll-hub-pos/backend/pkg/database"
	"stalll-hub-pos/backend/pkg/nsq"

	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

const (
	FacePaymentStatusInitialized = 0
	FacePaymentStatusProcessing  = 1
	FacePaymentStatusSuccess     = 2
	FacePaymentStatusFailed      = 3
	FacePaymentStatusCancelled   = 4

	FacePaymentProviderAlipay  = "alipay_face"
	FacePaymentProviderWechat  = "wechat_face"
)

type FacePaymentHandler struct {
	orderRepo *repository.OrderRepository
}

func NewFacePaymentHandler() *FacePaymentHandler {
	return &FacePaymentHandler{
		orderRepo: repository.NewOrderRepository(),
	}
}

func (h *FacePaymentHandler) FacePaymentInit(c *gin.Context) {
	var req dto.FacePaymentInitRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.Error(c, "参数错误: "+err.Error())
		return
	}

	if req.Provider != FacePaymentProviderAlipay && req.Provider != FacePaymentProviderWechat {
		middleware.Error(c, "不支持的刷脸支付提供商")
		return
	}

	order, err := h.orderRepo.GetByID(req.OrderID)
	if err != nil {
		middleware.Error(c, "订单不存在")
		return
	}

	if order.PayStatus == 1 {
		middleware.Error(c, "订单已支付")
		return
	}

	if order.OrderStatus == -1 {
		middleware.Error(c, "订单已取消")
		return
	}

	facePaymentID := fmt.Sprintf("FP%s%d", time.Now().Format("20060102150405"), order.ID)

	authInfo := ""
	if req.Provider == FacePaymentProviderAlipay {
		authInfo = h.buildAlipayAuthInfo(facePaymentID, order)
	} else {
		authInfo = h.buildWechatAuthInfo(facePaymentID, order)
	}

	record := model.FacePaymentRecord{
		FacePaymentID: facePaymentID,
		OrderID:       order.ID,
		OrderNo:       order.OrderNo,
		StoreID:       req.StoreID,
		DeviceID:      req.DeviceID,
		Provider:      req.Provider,
		Amount:        order.PayAmount,
		AuthInfo:      authInfo,
		Status:        FacePaymentStatusInitialized,
	}

	if err := database.DB.Create(&record).Error; err != nil {
		middleware.Error(c, "创建刷脸支付记录失败: "+err.Error())
		return
	}

	middleware.Success(c, gin.H{
		"face_payment_id": facePaymentID,
		"order_no":        order.OrderNo,
		"amount":          order.PayAmount,
		"provider":        req.Provider,
		"authinfo":        authInfo,
	})
}

func (h *FacePaymentHandler) FacePaymentConfirm(c *gin.Context) {
	var req dto.FacePaymentConfirmRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.Error(c, "参数错误: "+err.Error())
		return
	}

	var record model.FacePaymentRecord
	if err := database.DB.Where("face_payment_id = ?", req.FacePaymentID).First(&record).Error; err != nil {
		middleware.Error(c, "刷脸支付记录不存在")
		return
	}

	if record.Status == FacePaymentStatusSuccess {
		middleware.Error(c, "刷脸支付已完成")
		return
	}

	if record.Status == FacePaymentStatusCancelled {
		middleware.Error(c, "刷脸支付已取消")
		return
	}

	if err := database.DB.Model(&record).Updates(map[string]interface{}{
		"auth_code": req.AuthCode,
		"open_id":   req.OpenID,
		"status":    FacePaymentStatusProcessing,
	}).Error; err != nil {
		middleware.Error(c, "更新刷脸支付记录失败")
		return
	}

	var transactionID string
	var payTime time.Time

	switch req.Provider {
	case FacePaymentProviderAlipay:
		txID, pt, err := h.processAlipayFacePayment(&record, req.AuthCode)
		if err != nil {
			database.DB.Model(&record).Updates(map[string]interface{}{
				"status":  FacePaymentStatusFailed,
				"err_msg": err.Error(),
			})
			middleware.Error(c, "支付宝刷脸支付失败: "+err.Error())
			return
		}
		transactionID = txID
		payTime = pt
	case FacePaymentProviderWechat:
		txID, pt, err := h.processWechatFacePayment(&record, req.AuthCode)
		if err != nil {
			database.DB.Model(&record).Updates(map[string]interface{}{
				"status":  FacePaymentStatusFailed,
				"err_msg": err.Error(),
			})
			middleware.Error(c, "微信刷脸支付失败: "+err.Error())
			return
		}
		transactionID = txID
		payTime = pt
	default:
		middleware.Error(c, "不支持的刷脸支付提供商")
		return
	}

	err := database.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&record).Updates(map[string]interface{}{
			"status":         FacePaymentStatusSuccess,
			"transaction_id": transactionID,
			"pay_time":       payTime,
		}).Error; err != nil {
			return err
		}

		if err := tx.Model(&model.Order{}).Where("id = ?", record.OrderID).Updates(map[string]interface{}{
			"pay_status": 1,
			"pay_method": record.Provider,
			"pay_time":   payTime,
		}).Error; err != nil {
			return err
		}

		payment := model.OrderPayment{
			OrderID:       record.OrderID,
			PayMethod:     record.Provider,
			Amount:        record.Amount,
			TransactionID: transactionID,
			PayStatus:     1,
			PayTime:       &payTime,
		}
		if err := tx.Create(&payment).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		middleware.Error(c, "支付确认事务失败: "+err.Error())
		return
	}

	h.publishPaySuccessMessages(&record, transactionID)

	middleware.Success(c, gin.H{
		"success":        true,
		"order_no":       record.OrderNo,
		"transaction_id": transactionID,
		"pay_time":       payTime.Format("2006-01-02 15:04:05"),
	})
}

func (h *FacePaymentHandler) FacePaymentQuery(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		middleware.Error(c, "刷脸支付ID不能为空")
		return
	}

	var record model.FacePaymentRecord
	if err := database.DB.Where("face_payment_id = ?", id).First(&record).Error; err != nil {
		middleware.Error(c, "刷脸支付记录不存在")
		return
	}

	statusText := h.getFacePaymentStatusText(record.Status)
	middleware.Success(c, gin.H{
		"face_payment_id": record.FacePaymentID,
		"order_no":        record.OrderNo,
		"status":          record.Status,
		"status_text":     statusText,
		"provider":        record.Provider,
		"amount":          record.Amount,
		"transaction_id":  record.TransactionID,
		"pay_time":        record.PayTime,
	})
}

func (h *FacePaymentHandler) FacePaymentCancel(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		middleware.Error(c, "刷脸支付ID不能为空")
		return
	}

	var record model.FacePaymentRecord
	if err := database.DB.Where("face_payment_id = ?", id).First(&record).Error; err != nil {
		middleware.Error(c, "刷脸支付记录不存在")
		return
	}

	if record.Status == FacePaymentStatusSuccess {
		middleware.Error(c, "已完成的支付无法取消")
		return
	}

	if record.Status == FacePaymentStatusCancelled {
		middleware.Error(c, "支付已取消")
		return
	}

	if err := database.DB.Model(&record).Updates(map[string]interface{}{
		"status": FacePaymentStatusCancelled,
	}).Error; err != nil {
		middleware.Error(c, "取消刷脸支付失败: "+err.Error())
		return
	}

	middleware.Success(c, gin.H{
		"face_payment_id": record.FacePaymentID,
		"status":          FacePaymentStatusCancelled,
	})
}

func (h *FacePaymentHandler) VoiceBroadcast(c *gin.Context) {
	var req dto.VoiceBroadcastRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.Error(c, "参数错误: "+err.Error())
		return
	}

	msg := map[string]interface{}{
		"store_id":  req.StoreID,
		"text":      req.Text,
		"type":      req.Type,
		"timestamp": time.Now().Unix(),
	}

	if err := nsq.Publish("voice_broadcast", msg); err != nil {
		middleware.Error(c, "语音播报发送失败: "+err.Error())
		return
	}

	middleware.Success(c, gin.H{
		"success": true,
	})
}

func (h *FacePaymentHandler) buildAlipayAuthInfo(facePaymentID string, order *model.Order) string {
	authData := map[string]interface{}{
		"face_payment_id": facePaymentID,
		"order_no":        order.OrderNo,
		"amount":          order.PayAmount.String(),
		"scene":           "security_code",
	}
	data, _ := json.Marshal(authData)
	return string(data)
}

func (h *FacePaymentHandler) buildWechatAuthInfo(facePaymentID string, order *model.Order) string {
	authData := map[string]interface{}{
		"face_payment_id": facePaymentID,
		"order_no":        order.OrderNo,
		"amount":          order.PayAmount.String(),
	}
	data, _ := json.Marshal(authData)
	return string(data)
}

func (h *FacePaymentHandler) processAlipayFacePayment(record *model.FacePaymentRecord, authCode string) (string, time.Time, error) {
	transactionID := fmt.Sprintf("ALI%s", time.Now().Format("20060102150405"))
	payTime := time.Now()
	return transactionID, payTime, nil
}

func (h *FacePaymentHandler) processWechatFacePayment(record *model.FacePaymentRecord, authCode string) (string, time.Time, error) {
	transactionID := fmt.Sprintf("WX%s", time.Now().Format("20060102150405"))
	payTime := time.Now()
	return transactionID, payTime, nil
}

func (h *FacePaymentHandler) publishPaySuccessMessages(record *model.FacePaymentRecord, transactionID string) {
	nsq.PublishOrderPaySuccess(record.OrderNo, record.StoreID, map[string]interface{}{
		"order_id":       record.OrderID,
		"order_no":       record.OrderNo,
		"pay_method":     record.Provider,
		"amount":         record.Amount.String(),
		"transaction_id": transactionID,
	})

	nsq.PublishPrintOrder(record.OrderID, record.OrderNo, record.StoreID, 0, "receipt", nil)

	voiceText := fmt.Sprintf("刷脸支付成功，收款%s元", record.Amount.String())
	nsq.Publish("voice_broadcast", map[string]interface{}{
		"store_id":  record.StoreID,
		"text":      voiceText,
		"type":      "payment_success",
		"timestamp": time.Now().Unix(),
	})
}

func (h *FacePaymentHandler) getFacePaymentStatusText(status int) string {
	switch status {
	case FacePaymentStatusInitialized:
		return "已初始化"
	case FacePaymentStatusProcessing:
		return "处理中"
	case FacePaymentStatusSuccess:
		return "支付成功"
	case FacePaymentStatusFailed:
		return "支付失败"
	case FacePaymentStatusCancelled:
		return "已取消"
	default:
		return "未知状态"
	}
}
