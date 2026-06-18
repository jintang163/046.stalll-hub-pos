package nsq

import (
	"encoding/json"
	"log"
	"stalll-hub-pos/backend/config"
	"time"

	"github.com/nsqio/go-nsq"
)

var Producer *nsq.Producer

func InitProducer() {
	cfg := nsq.NewConfig()
	var err error
	Producer, err = nsq.NewProducer(config.AppConfig.NSQ.NSQDAddress, cfg)
	if err != nil {
		log.Fatalf("Failed to create NSQ producer: %v", err)
	}

	if err := Producer.Ping(); err != nil {
		log.Fatalf("Failed to ping NSQ: %v", err)
	}

	log.Println("NSQ producer initialized successfully")
}

func Publish(topic string, message interface{}) error {
	data, err := json.Marshal(message)
	if err != nil {
		return err
	}
	return Producer.Publish(topic, data)
}

func PublishDeferred(topic string, message interface{}, delay time.Duration) error {
	data, err := json.Marshal(message)
	if err != nil {
		return err
	}
	return Producer.DeferredPublish(topic, delay, data)
}

type ConsumerConfig struct {
	Topic   string
	Channel string
	Handler nsq.Handler
}

func InitConsumer(configs []ConsumerConfig) {
	for _, cfg := range configs {
		go startConsumer(cfg)
	}
}

func startConsumer(cfg ConsumerConfig) {
	nsqCfg := nsq.NewConfig()
	consumer, err := nsq.NewConsumer(cfg.Topic, cfg.Channel, nsqCfg)
	if err != nil {
		log.Printf("Failed to create consumer for topic %s: %v", cfg.Topic, err)
		return
	}

	consumer.AddHandler(cfg.Handler)

	err = consumer.ConnectToNSQLookupd(config.AppConfig.NSQ.LookupdAddress)
	if err != nil {
		log.Printf("Failed to connect to nsqlookupd for topic %s: %v", cfg.Topic, err)
		return
	}

	log.Printf("NSQ consumer started for topic: %s, channel: %s", cfg.Topic, cfg.Channel)

	select {}
}

func StopProducer() {
	if Producer != nil {
		Producer.Stop()
	}
}
