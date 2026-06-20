package consumer

import (
	"encoding/json"
	"log"
	"stalll-hub-pos/backend/pkg/nsq"
)

type StallDeviceAlertConsumer struct{}

func NewStallDeviceAlertConsumer() *StallDeviceAlertConsumer {
	return &StallDeviceAlertConsumer{}
}

func (c *StallDeviceAlertConsumer) HandleMessage(topic, channel string, message []byte) error {
	var msg nsq.StallDeviceAlertMessage
	if err := json.Unmarshal(message, &msg); err != nil {
		log.Printf("[StallDeviceAlertConsumer] 解析消息失败: %v", err)
		return err
	}

	log.Printf("[StallDeviceAlertConsumer] 收到设备告警: 设备ID=%d, 设备名=%s, 告警类型=%s, 离线时长=%d分钟",
		msg.DeviceID, msg.DeviceName, msg.AlertType, msg.OfflineMinutes)

	c.processAlert(&msg)

	return nil
}

func (c *StallDeviceAlertConsumer) processAlert(msg *nsq.StallDeviceAlertMessage) {
	switch msg.AlertType {
	case "offline":
		c.handleOfflineAlert(msg)
	case "reconnect":
		c.handleReconnectAlert(msg)
	case "low_battery":
		c.handleLowBatteryAlert(msg)
	default:
		log.Printf("[StallDeviceAlertConsumer] 未知告警类型: %s", msg.AlertType)
	}
}

func (c *StallDeviceAlertConsumer) handleOfflineAlert(msg *nsq.StallDeviceAlertMessage) {
	log.Printf("[设备告警] 设备离线告警: 摊位=%s, 设备=%s, 已离线%d分钟",
		msg.StallName, msg.DeviceName, msg.OfflineMinutes)
}

func (c *StallDeviceAlertConsumer) handleReconnectAlert(msg *nsq.StallDeviceAlertMessage) {
	log.Printf("[设备告警] 设备恢复在线: 摊位=%s, 设备=%s",
		msg.StallName, msg.DeviceName)
}

func (c *StallDeviceAlertConsumer) handleLowBatteryAlert(msg *nsq.StallDeviceAlertMessage) {
	log.Printf("[设备告警] 设备低电量: 摊位=%s, 设备=%s, 电量=%d%%",
		msg.StallName, msg.DeviceName, msg.BatteryLevel)
}
