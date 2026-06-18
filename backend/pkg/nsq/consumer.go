package nsq

import (
	"log"
	"stalll-hub-pos/backend/config"
	"sync"

	"github.com/nsqio/go-nsq"
)

type Consumer struct {
	consumers map[string]*nsq.Consumer
	mu        sync.RWMutex
	cfg       *nsq.Config
}

var (
	defaultConsumer *Consumer
	consumerOnce    sync.Once
)

func GetConsumer() *Consumer {
	consumerOnce.Do(func() {
		defaultConsumer = &Consumer{
			consumers: make(map[string]*nsq.Consumer),
			cfg:       nsq.NewConfig(),
		}
	})
	return defaultConsumer
}

type MessageHandler func(message *nsq.Message) error

func (c *Consumer) Register(topic, channel string, handler MessageHandler) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	key := topic + ":" + channel
	if _, exists := c.consumers[key]; exists {
		log.Printf("Consumer already registered for topic=%s, channel=%s", topic, channel)
		return nil
	}

	consumer, err := nsq.NewConsumer(topic, channel, c.cfg)
	if err != nil {
		log.Printf("Failed to create consumer for topic=%s, channel=%s: %v", topic, channel, err)
		return err
	}

	consumer.AddHandler(nsq.HandlerFunc(handler))

	err = consumer.ConnectToNSQLookupd(config.AppConfig.NSQ.LookupdAddress)
	if err != nil {
		log.Printf("Failed to connect to nsqlookupd for topic=%s, channel=%s: %v", topic, channel, err)
		return err
	}

	c.consumers[key] = consumer
	log.Printf("NSQ consumer registered: topic=%s, channel=%s", topic, channel)
	return nil
}

func (c *Consumer) RegisterMultiple(configs []ConsumerRegistration) error {
	for _, cfg := range configs {
		if err := c.Register(cfg.Topic, cfg.Channel, cfg.Handler); err != nil {
			log.Printf("Failed to register consumer for topic=%s: %v", cfg.Topic, err)
		}
	}
	return nil
}

func (c *Consumer) Stop() {
	c.mu.Lock()
	defer c.mu.Unlock()

	for key, consumer := range c.consumers {
		consumer.Stop()
		log.Printf("NSQ consumer stopped: %s", key)
	}
	c.consumers = make(map[string]*nsq.Consumer)
}

func (c *Consumer) Wait() {
	for _, consumer := range c.consumers {
		<-consumer.StopChan
	}
}

type ConsumerRegistration struct {
	Topic   string
	Channel string
	Handler MessageHandler
}

func InitConsumers(configs []ConsumerRegistration) error {
	return GetConsumer().RegisterMultiple(configs)
}

func StopConsumers() {
	GetConsumer().Stop()
}
