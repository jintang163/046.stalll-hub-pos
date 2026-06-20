package nsq

import "fmt"

const (
	TopicProductChanged  = "product_change"
	TopicStockChange     = "stock_change"
	TopicOrderCreated    = "order_create"
	TopicOrderStatus     = "order_status"
	TopicOrderPaid       = "order_pay_success"
	TopicOrderRefund     = "order_refund"
	TopicPrintOrder      = "print_order"
	TopicSyncProduct     = "sync_product"
	TopicMemberPoints    = "member_points"
	TopicStockWarning    = "stock_warning"
	TopicOrderUpdate     = "order_update"
	TopicStallChanged    = "stall_change"
	TopicStallOrder      = "stall_order"
	TopicStallDeviceAlert = "stall_device_alert"
)

type ProductChangeMessage struct {
	Action    string      `json:"action"`
	StoreID   uint        `json:"store_id"`
	ProductID uint        `json:"product_id"`
	Data      interface{} `json:"data"`
	Timestamp int64       `json:"timestamp"`
}

type StockChangeMessage struct {
	StoreID   uint `json:"store_id"`
	SKUID     uint `json:"sku_id"`
	ProductID uint `json:"product_id"`
	OldStock  int  `json:"old_stock"`
	NewStock  int  `json:"new_stock"`
	ChangeType string `json:"change_type"`
	Timestamp int64 `json:"timestamp"`
}

type OrderMessage struct {
	OrderNo   string      `json:"order_no"`
	StoreID   uint        `json:"store_id"`
	OrderData interface{} `json:"order_data"`
	Timestamp int64       `json:"timestamp"`
}

type PrintMessage struct {
	OrderID   uint        `json:"order_id"`
	OrderNo   string      `json:"order_no"`
	StoreID   uint        `json:"store_id"`
	PrinterID uint        `json:"printer_id"`
	PrintType string      `json:"print_type"`
	PrintData interface{} `json:"print_data"`
	Timestamp int64       `json:"timestamp"`
}

func PublishProductChange(action string, storeID, productID uint, data interface{}) error {
	msg := ProductChangeMessage{
		Action:    action,
		StoreID:   storeID,
		ProductID: productID,
		Data:      data,
		Timestamp: GetCurrentTimestamp(),
	}
	return Publish(TopicProductChanged, msg)
}

func PublishStockChange(storeID, skuID, productID uint, oldStock, newStock int, changeType string) error {
	msg := StockChangeMessage{
		StoreID:    storeID,
		SKUID:      skuID,
		ProductID:  productID,
		OldStock:   oldStock,
		NewStock:   newStock,
		ChangeType: changeType,
		Timestamp:  GetCurrentTimestamp(),
	}
	return Publish(TopicStockChange, msg)
}

func PublishOrderCreate(orderNo string, storeID uint, orderData interface{}) error {
	msg := OrderMessage{
		OrderNo:   orderNo,
		StoreID:   storeID,
		OrderData: orderData,
		Timestamp: GetCurrentTimestamp(),
	}
	return Publish(TopicOrderCreated, msg)
}

func PublishOrderPaySuccess(orderNo string, storeID uint, orderData interface{}) error {
	msg := OrderMessage{
		OrderNo:   orderNo,
		StoreID:   storeID,
		OrderData: orderData,
		Timestamp: GetCurrentTimestamp(),
	}
	return Publish(TopicOrderPaid, msg)
}

func PublishPrintOrder(orderID uint, orderNo string, storeID uint, printerID uint, printType string, printData interface{}) error {
	msg := PrintMessage{
		OrderID:   orderID,
		OrderNo:   orderNo,
		StoreID:   storeID,
		PrinterID: printerID,
		PrintType: printType,
		PrintData: printData,
		Timestamp: GetCurrentTimestamp(),
	}
	return Publish(TopicPrintOrder, msg)
}

type StallChangeMessage struct {
	ChangeType string      `json:"change_type"`
	StoreID    uint        `json:"store_id"`
	StallID    uint        `json:"stall_id"`
	StallNo    string      `json:"stall_no"`
	StallName  string      `json:"stall_name"`
	Status     int         `json:"status"`
	Data       interface{} `json:"data"`
	Timestamp  int64       `json:"timestamp"`
}

type StallOrderMessage struct {
	OrderNo   string      `json:"order_no"`
	StoreID   uint        `json:"store_id"`
	StallID   uint        `json:"stall_id"`
	Items     interface{} `json:"items"`
	Order     interface{} `json:"order"`
	Action    string      `json:"action"`
	Timestamp int64       `json:"timestamp"`
}

type StallDeviceAlertMessage struct {
	DeviceID       uint   `json:"device_id"`
	DeviceName     string `json:"device_name"`
	DeviceNo       string `json:"device_no"`
	StallID        uint   `json:"stall_id"`
	StallName      string `json:"stall_name"`
	StoreID        uint   `json:"store_id"`
	AlertType      string `json:"alert_type"`
	OfflineMinutes int    `json:"offline_minutes"`
	BatteryLevel   int    `json:"battery_level"`
	Timestamp      int64  `json:"timestamp"`
}

func FormatStallOrderTopic(stallID uint) string {
	return fmt.Sprintf("stall_order_stall_%d", stallID)
}

func PublishStallChange(changeType string, storeID, stallID uint, stallNo, stallName string, status int, data interface{}) error {
	msg := StallChangeMessage{
		ChangeType: changeType,
		StoreID:    storeID,
		StallID:    stallID,
		StallNo:    stallNo,
		StallName:  stallName,
		Status:     status,
		Data:       data,
		Timestamp:  GetCurrentTimestamp(),
	}
	return Publish(TopicStallChanged, msg)
}

func PublishStallOrder(orderNo string, storeID, stallID uint, items, order interface{}, action string) error {
	msg := StallOrderMessage{
		OrderNo:   orderNo,
		StoreID:   storeID,
		StallID:   stallID,
		Items:     items,
		Order:     order,
		Action:    action,
		Timestamp: GetCurrentTimestamp(),
	}
	return Publish(FormatStallOrderTopic(stallID), msg)
}

func PublishStallDeviceAlert(deviceID uint, deviceName, deviceNo, stallName string, storeID, stallID uint, alertType string, offlineMinutes, batteryLevel int) error {
	msg := StallDeviceAlertMessage{
		DeviceID:       deviceID,
		DeviceName:     deviceName,
		DeviceNo:       deviceNo,
		StallID:        stallID,
		StallName:      stallName,
		StoreID:        storeID,
		AlertType:      alertType,
		OfflineMinutes: offlineMinutes,
		BatteryLevel:   batteryLevel,
		Timestamp:      GetCurrentTimestamp(),
	}
	return Publish(TopicStallDeviceAlert, msg)
}

func GetCurrentTimestamp() int64 {
	return GetCurrentTime().Unix()
}
