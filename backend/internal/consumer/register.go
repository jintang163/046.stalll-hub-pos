package consumer

import (
	"log"
	"stalll-hub-pos/backend/pkg/nsq"
)

func InitAllConsumers() error {
	orderConsumer := NewOrderCreateConsumer()
	paymentConsumer := NewPaymentSuccessConsumer()
	productConsumer := NewProductChangeConsumer()
	stallChangeConsumer := NewStallChangeConsumer()
	stallDeviceAlertConsumer := NewStallDeviceAlertConsumer()

	configs := []nsq.ConsumerRegistration{
		{
			Topic:   nsq.TopicOrderCreated,
			Channel: "order_create_channel",
			Handler: orderConsumer.HandleMessage,
		},
		{
			Topic:   nsq.TopicOrderPaid,
			Channel: "order_pay_channel",
			Handler: paymentConsumer.HandleMessage,
		},
		{
			Topic:   nsq.TopicProductChanged,
			Channel: "product_change_channel",
			Handler: productConsumer.HandleMessage,
		},
		{
			Topic:   nsq.TopicStallChanged,
			Channel: "stall_change_channel",
			Handler: stallChangeConsumer.HandleMessage,
		},
		{
			Topic:   nsq.TopicStallDeviceAlert,
			Channel: "stall_device_alert_channel",
			Handler: stallDeviceAlertConsumer.HandleMessage,
		},
	}

	log.Println("[PrintConsumer] 打印任务已迁移到 Node.js printer-service，Go端不再消费 print_order 主题")

	return nsq.InitConsumers(configs)
}

