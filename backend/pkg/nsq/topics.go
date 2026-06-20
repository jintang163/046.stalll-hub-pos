package nsq

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
	Action    string      `json:"action"`
	StoreID   uint        `json:"store_id"`
	StallID   uint        `json:"stall_id"`
	Data      interface{} `json:"data"`
	Timestamp int64       `json:"timestamp"`
}

type StallOrderMessage struct {
	OrderNo    string      `json:"order_no"`
	StoreID    uint        `json:"store_id"`
	StallID    uint        `json:"stall_id"`
	OrderData  interface{} `json:"order_data"`
	Action     string      `json:"action"`
	Timestamp  int64       `json:"timestamp"`
}

type StallDeviceAlertMessage struct {
	DeviceID   string `json:"device_id"`
	StallID    uint   `json:"stall_id"`
	StoreID    uint   `json:"store_id"`
	AlertType  string `json:"alert_type"`
	Timestamp  int64  `json:"timestamp"`
}

func PublishStallChange(action string, storeID, stallID uint, data interface{}) error {
	msg := StallChangeMessage{
		Action:    action,
		StoreID:   storeID,
		StallID:   stallID,
		Data:      data,
		Timestamp: GetCurrentTimestamp(),
	}
	return Publish(TopicStallChanged, msg)
}

func PublishStallOrder(orderNo string, storeID, stallID uint, orderData interface{}, action string) error {
	msg := StallOrderMessage{
		OrderNo:   orderNo,
		StoreID:   storeID,
		StallID:   stallID,
		OrderData: orderData,
		Action:    action,
		Timestamp: GetCurrentTimestamp(),
	}
	return Publish(TopicStallOrder, msg)
}

func PublishStallDeviceAlert(deviceID string, storeID, stallID uint, alertType string) error {
	msg := StallDeviceAlertMessage{
		DeviceID:  deviceID,
		StallID:   stallID,
		StoreID:   storeID,
		AlertType: alertType,
		Timestamp: GetCurrentTimestamp(),
	}
	return Publish(TopicStallDeviceAlert, msg)
}

func GetCurrentTimestamp() int64 {
	return GetCurrentTime().Unix()
}
