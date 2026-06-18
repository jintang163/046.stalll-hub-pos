package main

import (
	"fmt"
	"log"
	"stalll-hub-pos/backend/config"
	"stalll-hub-pos/backend/internal/api"
	"stalll-hub-pos/backend/internal/model"
	"stalll-hub-pos/backend/pkg/database"
	"stalll-hub-pos/backend/pkg/minio"
	"stalll-hub-pos/backend/pkg/nsq"
	"stalll-hub-pos/backend/pkg/redis"

	"golang.org/x/crypto/bcrypt"
)

func main() {
	config.LoadConfig()

	database.InitMySQL()
	redis.InitRedis()
	minio.InitMinIO()
	nsq.InitProducer()

	database.AutoMigrate(
		&model.Store{},
		&model.Printer{},
		&model.StoreUser{},
		&model.Category{},
		&model.Product{},
		&model.ProductSKU{},
		&model.ProductAttribute{},
		&model.AttributeValue{},
		&model.SKUAttributeValue{},
		&model.StockWarning{},
		&model.Order{},
		&model.OrderItem{},
		&model.OrderPayment{},
		&model.OrderRefund{},
		&model.RefundItem{},
		&model.SyncRecord{},
		&model.OrderQueue{},
		&model.Member{},
		&model.MemberLevel{},
		&model.MemberPointsRecord{},
		&model.Coupon{},
		&model.MemberCoupon{},
		&model.DailyReport{},
		&model.ProductSalesReport{},
		&model.CategorySalesReport{},
		&model.HourlyReport{},
		&model.PaymentReport{},
		&model.ReportTask{},
	)

	initDefaultData()

	initNSQConsumers()

	r := api.SetupRouter(database.DB, nsq.Producer)

	log.Printf("Server starting on port %s", config.AppConfig.Server.Port)
	if err := r.Run(fmt.Sprintf(":%s", config.AppConfig.Server.Port)); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}

	defer nsq.StopProducer()
}

func initDefaultData() {
	var storeCount int64
	database.DB.Model(&model.Store{}).Count(&storeCount)
	if storeCount == 0 {
		store := &model.Store{
			Name:          "大排档总店",
			Address:       "北京市朝阳区xx路xx号",
			Phone:         "13800138000",
			BusinessHours: "10:00-22:00",
			Status:        1,
			Description:   "正宗大排档，地道美食",
		}
		database.DB.Create(store)
		log.Println("Default store created")

		var userCount int64
		database.DB.Model(&model.StoreUser{}).Count(&userCount)
		if userCount == 0 {
			hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
			user := &model.StoreUser{
				StoreID:  store.ID,
				Username: "admin",
				Password: string(hashedPassword),
				RealName: "管理员",
				Phone:    "13800138000",
				Role:     "admin",
				Status:   1,
			}
			database.DB.Create(user)
			log.Println("Default admin user created: admin/admin123")
		}
	}
}

func initNSQConsumers() {
	consumerConfigs := []nsq.ConsumerConfig{
		{
			Topic:   nsq.TopicOrderCreated,
			Channel: "order_create_channel",
			Handler: &OrderCreateHandler{},
		},
		{
			Topic:   nsq.TopicOrderPaid,
			Channel: "order_pay_channel",
			Handler: &OrderPayHandler{},
		},
		{
			Topic:   nsq.TopicOrderRefund,
			Channel: "order_refund_channel",
			Handler: &OrderRefundHandler{},
		},
		{
			Topic:   nsq.TopicProductChanged,
			Channel: "product_change_channel",
			Handler: &ProductChangeHandler{},
		},
	}

	nsq.InitConsumer(consumerConfigs)
}

type OrderCreateHandler struct{}

func (h *OrderCreateHandler) HandleMessage(m *nsq.Message) error {
	log.Printf("[NSQ] Order created: %s", string(m.Body))
	return nil
}

type OrderPayHandler struct{}

func (h *OrderPayHandler) HandleMessage(m *nsq.Message) error {
	log.Printf("[NSQ] Order paid: %s", string(m.Body))
	return nil
}

type OrderRefundHandler struct{}

func (h *OrderRefundHandler) HandleMessage(m *nsq.Message) error {
	log.Printf("[NSQ] Order refund: %s", string(m.Body))
	return nil
}

type ProductChangeHandler struct{}

func (h *ProductChangeHandler) HandleMessage(m *nsq.Message) error {
	log.Printf("[NSQ] Product changed: %s", string(m.Body))
	return nil
}
