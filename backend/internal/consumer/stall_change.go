package consumer

import (
	"encoding/json"
	"log"
	"stalll-hub-pos/backend/pkg/nsq"
)

type StallChangeConsumer struct{}

func NewStallChangeConsumer() *StallChangeConsumer {
	return &StallChangeConsumer{}
}

func (c *StallChangeConsumer) HandleMessage(topic, channel string, message []byte) error {
	var msg nsq.StallChangeMessage
	if err := json.Unmarshal(message, &msg); err != nil {
		log.Printf("[StallChangeConsumer] 解析消息失败: %v", err)
		return err
	}

	log.Printf("[StallChangeConsumer] 收到摊位变更消息: 摊位ID=%d, 变更类型=%s", msg.StallID, msg.ChangeType)

	switch msg.ChangeType {
	case "create":
		c.handleStallCreate(&msg)
	case "update":
		c.handleStallUpdate(&msg)
	case "delete":
		c.handleStallDelete(&msg)
	case "status_change":
		c.handleStatusChange(&msg)
	default:
		log.Printf("[StallChangeConsumer] 未知变更类型: %s", msg.ChangeType)
	}

	return nil
}

func (c *StallChangeConsumer) handleStallCreate(msg *nsq.StallChangeMessage) {
	log.Printf("[摊位变更] 新增摊位: %s (编号: %s)", msg.StallName, msg.StallNo)
}

func (c *StallChangeConsumer) handleStallUpdate(msg *nsq.StallChangeMessage) {
	log.Printf("[摊位变更] 更新摊位: %s (编号: %s)", msg.StallName, msg.StallNo)
}

func (c *StallChangeConsumer) handleStallDelete(msg *nsq.StallChangeMessage) {
	log.Printf("[摊位变更] 删除摊位: ID=%d", msg.StallID)
}

func (c *StallChangeConsumer) handleStatusChange(msg *nsq.StallChangeMessage) {
	log.Printf("[摊位变更] 摊位状态变更: %s -> 状态=%d", msg.StallName, msg.Status)
}
