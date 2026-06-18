package consumer

import (
	"encoding/json"
	"fmt"
	"log"
	"stalll-hub-pos/backend/internal/model"
	"stalll-hub-pos/backend/internal/repository"
	"stalll-hub-pos/backend/pkg/nsq"
	"stalll-hub-pos/backend/pkg/redis"
	"time"

	"github.com/nsqio/go-nsq"
)

type ProductChangeConsumer struct {
	productRepo *repository.ProductRepository
}

func NewProductChangeConsumer() *ProductChangeConsumer {
	return &ProductChangeConsumer{
		productRepo: repository.NewProductRepository(),
	}
}

type ProductChangeMessage struct {
	Action    string      `json:"action"`
	StoreID   uint        `json:"store_id"`
	ProductID uint        `json:"product_id"`
	Data      interface{} `json:"data"`
	Timestamp int64       `json:"timestamp"`
}

func (c *ProductChangeConsumer) HandleMessage(m *nsq.Message) error {
	var msg ProductChangeMessage
	if err := json.Unmarshal(m.Body, &msg); err != nil {
		log.Printf("[ProductConsumer] Failed to unmarshal message: %v, body: %s", err, string(m.Body))
		return err
	}

	log.Printf("[ProductConsumer] Processing product change: action=%s, product_id=%d, store_id=%d", msg.Action, msg.ProductID, msg.StoreID)

	if err := c.updateCache(&msg); err != nil {
		log.Printf("[ProductConsumer] Failed to update cache for product %d: %v", msg.ProductID, err)
	}

	if err := c.notifyFrontend(&msg); err != nil {
		log.Printf("[ProductConsumer] Failed to notify frontend for product %d: %v", msg.ProductID, err)
	}

	log.Printf("[ProductConsumer] Product change processed: product_id=%d, action=%s", msg.ProductID, msg.Action)
	return nil
}

func (c *ProductChangeConsumer) updateCache(msg *ProductChangeMessage) error {
	product, err := c.productRepo.GetByID(msg.ProductID)
	if err != nil {
		log.Printf("[ProductConsumer] Failed to get product %d: %v", msg.ProductID, err)
		return c.deleteProductCache(msg.StoreID, msg.ProductID)
	}

	if msg.Action == "delete" {
		return c.deleteProductCache(msg.StoreID, msg.ProductID)
	}

	productKey := c.getProductKey(msg.StoreID, msg.ProductID)
	productData, err := json.Marshal(product)
	if err != nil {
		return fmt.Errorf("failed to marshal product: %w", err)
	}
	if err := redis.Set(productKey, productData, 30*time.Minute); err != nil {
		log.Printf("[ProductConsumer] Failed to set product cache: %v", err)
	}

	productsListKey := c.getProductsListKey(msg.StoreID)
	if err := redis.Del(productsListKey); err != nil {
		log.Printf("[ProductConsumer] Failed to delete products list cache: %v", err)
	}

	categoryKey := c.getCategoryProductsKey(msg.StoreID, product.CategoryID)
	if err := redis.Del(categoryKey); err != nil {
		log.Printf("[ProductConsumer] Failed to delete category products cache: %v", err)
	}

	return nil
}

func (c *ProductChangeConsumer) deleteProductCache(storeID, productID uint) error {
	productKey := c.getProductKey(storeID, productID)
	if err := redis.Del(productKey); err != nil {
		log.Printf("[ProductConsumer] Failed to delete product cache: %v", err)
	}

	productsListKey := c.getProductsListKey(storeID)
	if err := redis.Del(productsListKey); err != nil {
		log.Printf("[ProductConsumer] Failed to delete products list cache: %v", err)
	}

	return nil
}

func (c *ProductChangeConsumer) notifyFrontend(msg *ProductChangeMessage) error {
	notification := map[string]interface{}{
		"type":       "product_change",
		"action":     msg.Action,
		"product_id": msg.ProductID,
		"store_id":   msg.StoreID,
		"timestamp":  msg.Timestamp,
	}

	channel := fmt.Sprintf("store:%d:products", msg.StoreID)
	notificationData, err := json.Marshal(notification)
	if err != nil {
		return fmt.Errorf("failed to marshal notification: %w", err)
	}

	return redis.Client.Publish(redis.Ctx, channel, notificationData).Err()
}

func (c *ProductChangeConsumer) getProductKey(storeID, productID uint) string {
	return fmt.Sprintf("product:%d:%d", storeID, productID)
}

func (c *ProductChangeConsumer) getProductsListKey(storeID uint) string {
	return fmt.Sprintf("products:list:%d", storeID)
}

func (c *ProductChangeConsumer) getCategoryProductsKey(storeID, categoryID uint) string {
	return fmt.Sprintf("products:category:%d:%d", storeID, categoryID)
}

func (c *ProductChangeConsumer) getSKUKey(storeID, skuID uint) string {
	return fmt.Sprintf("sku:%d:%d", storeID, skuID)
}

func (c *ProductChangeConsumer) updateSKUCache(storeID uint, sku *model.ProductSKU) error {
	skuKey := c.getSKUKey(storeID, sku.ID)
	skuData, err := json.Marshal(sku)
	if err != nil {
		return fmt.Errorf("failed to marshal SKU: %w", err)
	}
	return redis.Set(skuKey, skuData, 30*time.Minute)
}
