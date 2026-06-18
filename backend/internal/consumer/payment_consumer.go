package consumer

import (
	"encoding/json"
	"log"
	"stalll-hub-pos/backend/internal/model"
	"stalll-hub-pos/backend/internal/repository"
	"stalll-hub-pos/backend/pkg/database"
	"stalll-hub-pos/backend/pkg/nsq"

	"github.com/nsqio/go-nsq"
	"gorm.io/gorm"
)

type PaymentSuccessConsumer struct {
	orderRepo  *repository.OrderRepository
	memberRepo *repository.MemberRepository
}

func NewPaymentSuccessConsumer() *PaymentSuccessConsumer {
	return &PaymentSuccessConsumer{
		orderRepo:  repository.NewOrderRepository(database.DB),
		memberRepo: repository.NewMemberRepository(database.DB),
	}
}

type PaymentSuccessMessage struct {
	OrderID       uint   `json:"order_id"`
	OrderNo       string `json:"order_no"`
	PayMethod     string `json:"pay_method"`
	Amount        string `json:"amount"`
	TransactionID string `json:"transaction_id"`
	PayTime       string `json:"pay_time"`
	StoreID       uint   `json:"store_id"`
}

func (c *PaymentSuccessConsumer) HandleMessage(m *nsq.Message) error {
	var msg PaymentSuccessMessage
	if err := json.Unmarshal(m.Body, &msg); err != nil {
		log.Printf("[PaymentConsumer] Failed to unmarshal message: %v, body: %s", err, string(m.Body))
		return err
	}

	log.Printf("[PaymentConsumer] Processing payment success: order_no=%s, store_id=%d", msg.OrderNo, msg.StoreID)

	order, err := c.orderRepo.GetByID(msg.OrderID)
	if err != nil {
		log.Printf("[PaymentConsumer] Failed to get order %d: %v", msg.OrderID, err)
		return err
	}

	if order.PayStatus == 1 {
		log.Printf("[PaymentConsumer] Order %s already paid, skipping", msg.OrderNo)
		return nil
	}

	if err := c.updateOrderStatus(order); err != nil {
		log.Printf("[PaymentConsumer] Failed to update order status for %s: %v", msg.OrderNo, err)
		return err
	}

	if order.MemberID > 0 && order.PointsEarned > 0 {
		if err := c.addMemberPoints(order); err != nil {
			log.Printf("[PaymentConsumer] Failed to add member points for order %s: %v", msg.OrderNo, err)
		}
	}

	if err := c.sendPrintMessage(order); err != nil {
		log.Printf("[PaymentConsumer] Failed to send print message for order %s: %v", msg.OrderNo, err)
	}

	if err := c.sendOrderStatusNotification(order); err != nil {
		log.Printf("[PaymentConsumer] Failed to send order status notification for %s: %v", msg.OrderNo, err)
	}

	log.Printf("[PaymentConsumer] Payment success processed for order %s", msg.OrderNo)
	return nil
}

func (c *PaymentSuccessConsumer) updateOrderStatus(order *model.Order) error {
	return database.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&model.Order{}).Where("id = ?", order.ID).
			Updates(map[string]interface{}{
				"pay_status": 1,
				"pay_method": order.PayMethod,
				"pay_time":   order.PayTime,
			}).Error; err != nil {
			return err
		}

		return tx.Model(&model.Order{}).Where("id = ?", order.ID).
			Update("order_status", 2).Error
	})
}

func (c *PaymentSuccessConsumer) addMemberPoints(order *model.Order) error {
	remark := "消费赠送积分"
	if order.PointsEarned > 0 {
		remark = "消费赠送积分"
	}
	return c.memberRepo.AdjustPoints(order.MemberID, order.PointsEarned, remark, order.ID)
}

func (c *PaymentSuccessConsumer) sendPrintMessage(order *model.Order) error {
	items := make([]map[string]interface{}, len(order.Items))
	for i, item := range order.Items {
		items[i] = map[string]interface{}{
			"product_name": item.ProductName,
			"sku_name":     item.SKUName,
			"quantity":     item.Quantity,
			"price":        item.Price.String(),
			"subtotal":     item.Subtotal.String(),
		}
	}

	printData := map[string]interface{}{
		"order_id":      order.ID,
		"order_no":      order.OrderNo,
		"store_id":      order.StoreID,
		"table_no":      order.TableNo,
		"order_type":    order.OrderType,
		"total_amount":  order.TotalAmount.String(),
		"pay_amount":    order.PayAmount.String(),
		"pay_method":    order.PayMethod,
		"pay_status":    order.PayStatus,
		"points_earned": order.PointsEarned,
		"items":         items,
		"paid_at":       order.PayTime,
	}

	return nsq.PublishPrintOrder(order.ID, order.OrderNo, order.StoreID, 0, "receipt", printData)
}

func (c *PaymentSuccessConsumer) sendOrderStatusNotification(order *model.Order) error {
	statusData := map[string]interface{}{
		"order_id":    order.ID,
		"order_no":    order.OrderNo,
		"store_id":    order.StoreID,
		"status":      2,
		"status_text": "已支付",
		"timestamp":   nsq.GetCurrentTimestamp(),
	}
	return nsq.Publish(nsq.TopicOrderStatus, statusData)
}
