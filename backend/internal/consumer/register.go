package consumer

import (
	"stalll-hub-pos/backend/pkg/nsq"
)

func InitAllConsumers() error {
	orderConsumer := NewOrderCreateConsumer()
	paymentConsumer := NewPaymentSuccessConsumer()
	printConsumer := NewPrintConsumer()
	productConsumer := NewProductChangeConsumer()

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
			Topic:   nsq.TopicPrintOrder,
			Channel: "print_order_channel",
			Handler: printConsumer.HandleMessage,
		},
		{
			Topic:   nsq.TopicProductChanged,
			Channel: "product_change_channel",
			Handler: productConsumer.HandleMessage,
		},
	}

	return nsq.InitConsumers(configs)
}
