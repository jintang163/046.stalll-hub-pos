package consumer

import (
	"encoding/json"
	"fmt"
	"log"
	"stalll-hub-pos/backend/internal/model"
	"stalll-hub-pos/backend/internal/repository"
	"stalll-hub-pos/backend/pkg/database"
	"stalll-hub-pos/backend/pkg/nsq"

	"github.com/nsqio/go-nsq"
	"github.com/shopspring/decimal"
)

type OrderCreateConsumer struct {
	orderRepo   *repository.OrderRepository
	productRepo *repository.ProductRepository
	memberRepo  *repository.MemberRepository
}

func NewOrderCreateConsumer() *OrderCreateConsumer {
	return &OrderCreateConsumer{
		orderRepo:   repository.NewOrderRepository(database.DB),
		productRepo: repository.NewProductRepository(),
		memberRepo:  repository.NewMemberRepository(database.DB),
	}
}

type OrderCreateMessage struct {
	OrderID    uint            `json:"order_id"`
	OrderNo    string          `json:"order_no"`
	StoreID    uint            `json:"store_id"`
	OrderType  string          `json:"order_type"`
	PayAmount  string          `json:"pay_amount"`
	ItemCount  int             `json:"item_count"`
	MemberID   uint            `json:"member_id"`
	CreatedAt  string          `json:"created_at"`
}

func (c *OrderCreateConsumer) HandleMessage(m *nsq.Message) error {
	var msg OrderCreateMessage
	if err := json.Unmarshal(m.Body, &msg); err != nil {
		log.Printf("[OrderConsumer] Failed to unmarshal message: %v, body: %s", err, string(m.Body))
		return err
	}

	log.Printf("[OrderConsumer] Processing order created: order_no=%s, store_id=%d", msg.OrderNo, msg.StoreID)

	order, err := c.orderRepo.GetByID(msg.OrderID)
	if err != nil {
		log.Printf("[OrderConsumer] Failed to get order %d: %v", msg.OrderID, err)
		return err
	}

	if err := c.processStockDeduction(order); err != nil {
		log.Printf("[OrderConsumer] Stock deduction failed for order %s: %v", msg.OrderNo, err)
		return err
	}

	if order.MemberID > 0 {
		if err := c.updateMemberStats(order); err != nil {
			log.Printf("[OrderConsumer] Failed to update member stats for order %s: %v", msg.OrderNo, err)
		}
	}

	if err := c.sendPrintMessage(order); err != nil {
		log.Printf("[OrderConsumer] Failed to send print message for order %s: %v", msg.OrderNo, err)
	}

	log.Printf("[OrderConsumer] Order %s processed successfully", msg.OrderNo)
	return nil
}

func (c *OrderCreateConsumer) processStockDeduction(order *model.Order) error {
	for _, item := range order.Items {
		success := false
		maxRetries := 3

		for retry := 0; retry < maxRetries && !success; retry++ {
			oldStock, newStock, err := c.productRepo.DecreaseStockWithOptimisticLock(item.SKUID, item.Quantity)
			if err != nil {
				log.Printf("[OrderConsumer] Failed to decrease stock for SKU %d (attempt %d): %v", item.SKUID, retry+1, err)
				continue
			}

			if oldStock == newStock {
				if oldStock < item.Quantity {
					log.Printf("[OrderConsumer] Insufficient stock for SKU %d: have %d, need %d", item.SKUID, oldStock, item.Quantity)
					c.sendStockWarning(order.StoreID, item.SKUID, item.ProductID, oldStock, item.Quantity)
					return fmt.Errorf("insufficient stock for SKU %d", item.SKUID)
				}
				log.Printf("[OrderConsumer] Optimistic lock conflict for SKU %d, retrying... (attempt %d)", item.SKUID, retry+1)
				continue
			}

			success = true
			log.Printf("[OrderConsumer] Stock decreased for SKU %d: %d -> %d", item.SKUID, oldStock, newStock)

			sku, _ := c.productRepo.GetSKUByID(item.SKUID)
			product, _ := c.productRepo.GetByID(item.ProductID)
			threshold := 10
			if product != nil {
				threshold = product.StockWarningThreshold
			}
			_ = c.productRepo.CheckStockWarning(order.StoreID, item.SKUID, item.ProductID, newStock, threshold)

			_ = nsq.PublishStockChange(order.StoreID, item.SKUID, item.ProductID, oldStock, newStock, "decrease")
		}

		if !success {
			return fmt.Errorf("failed to decrease stock for SKU %d after %d retries", item.SKUID, maxRetries)
		}
	}
	return nil
}

func (c *OrderCreateConsumer) sendStockWarning(storeID, skuID, productID uint, currentStock, neededQty int) {
	warningData := map[string]interface{}{
		"store_id":     storeID,
		"sku_id":       skuID,
		"product_id":   productID,
		"current_stock": currentStock,
		"needed_qty":   neededQty,
		"timestamp":    nsq.GetCurrentTimestamp(),
	}
	_ = nsq.Publish(nsq.TopicStockWarning, warningData)
	log.Printf("[OrderConsumer] Stock warning sent for SKU %d", skuID)
}

func (c *OrderCreateConsumer) updateMemberStats(order *model.Order) error {
	payAmount, _ := decimal.NewFromString(order.PayAmount.String())
	amount, _ := payAmount.Float64()
	return c.memberRepo.UpdateStats(order.MemberID, amount)
}

func (c *OrderCreateConsumer) sendPrintMessage(order *model.Order) error {
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
		"order_id":     order.ID,
		"order_no":     order.OrderNo,
		"store_id":     order.StoreID,
		"table_no":     order.TableNo,
		"order_type":   order.OrderType,
		"total_amount": order.TotalAmount.String(),
		"pay_amount":   order.PayAmount.String(),
		"items":        items,
		"created_at":   order.CreatedAt,
	}

	return nsq.PublishPrintOrder(order.ID, order.OrderNo, order.StoreID, 0, "order", printData)
}
