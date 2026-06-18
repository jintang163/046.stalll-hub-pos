package consumer

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"stalll-hub-pos/backend/internal/consumer/escpos"
	"stalll-hub-pos/backend/internal/model"
	"stalll-hub-pos/backend/internal/repository"
	"stalll-hub-pos/backend/pkg/database"
	"stalll-hub-pos/backend/pkg/nsq"

	"github.com/nsqio/go-nsq"
)

type PrintConsumer struct {
	printerRepo *repository.PrinterRepository
	storeRepo   *repository.StoreRepository
	orderRepo   *repository.OrderRepository
}

func NewPrintConsumer() *PrintConsumer {
	return &PrintConsumer{
		printerRepo: repository.NewPrinterRepository(database.DB),
		storeRepo:   repository.NewStoreRepository(database.DB),
		orderRepo:   repository.NewOrderRepository(database.DB),
	}
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

type OrderPrintData struct {
	OrderID     uint                   `json:"order_id"`
	OrderNo     string                 `json:"order_no"`
	StoreID     uint                   `json:"store_id"`
	TableNo     string                 `json:"table_no"`
	OrderType   string                 `json:"order_type"`
	TotalAmount string                 `json:"total_amount"`
	PayAmount   string                 `json:"pay_amount"`
	PayMethod   string                 `json:"pay_method"`
	PayStatus   int                    `json:"pay_status"`
	PointsEarned int                 `json:"points_earned"`
	Items       []OrderPrintItem       `json:"items"`
	CreatedAt   interface{}          `json:"created_at"`
	PaidAt      interface{}          `json:"paid_at"`
}

type OrderPrintItem struct {
	ProductName string `json:"product_name"`
	SKUName     string `json:"sku_name"`
	Quantity     int    `json:"quantity"`
	Price        string `json:"price"`
	Subtotal     string `json:"subtotal"`
}

func (c *PrintConsumer) HandleMessage(m *nsq.Message) error {
	var msg PrintMessage
	if err := json.Unmarshal(m.Body, &msg); err != nil {
		log.Printf("[PrintConsumer] Failed to unmarshal message: %v, body: %s", err, string(m.Body))
		return err
	}

	log.Printf("[PrintConsumer] Processing print: order_no=%s, type=%s, store_id=%d", msg.OrderNo, msg.PrintType, msg.StoreID)

	store, err := c.storeRepo.GetByID(msg.StoreID)
	if err != nil {
		log.Printf("[PrintConsumer] Failed to get store %d: %v", msg.StoreID, err)
		return err
	}

	printers, err := c.getPrinters(msg)
	if err != nil {
		log.Printf("[PrintConsumer] Failed to get printers for store %d: %v", msg.StoreID, err)
		return err
	}

	printData, err := c.generatePrintCommands(store.Name, msg)
	if err != nil {
		log.Printf("[PrintConsumer] Failed to generate print commands: %v", err)
		return err
	}

	for _, printer := range printers {
		if printer.Status != 1 {
			log.Printf("[PrintConsumer] Printer %d is offline, skipping", printer.ID)
			continue
		}

		if err := c.printToNetworkPrinter(printer, printData); err != nil {
			log.Printf("[PrintConsumer] Failed to print to printer %s: %v", printer.IPAddress, err)
			continue
		}
		log.Printf("[PrintConsumer] Successfully printed to printer %s for order %s", printer.IPAddress, msg.OrderNo)
	}

	c.updateOrderPrintStatus(msg.OrderID)

	log.Printf("[PrintConsumer] Print processed for order %s", msg.OrderNo)
	return nil
}

func (c *PrintConsumer) getPrinters(msg PrintMessage) ([]model.Printer, error) {
	if msg.PrinterID > 0 {
		printer, err := c.printerRepo.GetByID(msg.PrinterID)
		if err != nil {
			return nil, err
		}
		return []model.Printer{*printer}, nil
	}

	printType := "kitchen"
	if msg.PrintType == "receipt" {
		printType = "receipt"
	}

	printers, err := c.printerRepo.GetByStoreAndType(msg.StoreID, printType)
	if err != nil {
		return nil, err
	}

	if len(printers) == 0 {
		printers, err = c.printerRepo.GetByStore(msg.StoreID)
		if err != nil {
			return nil, err
		}
	}

	return printers, nil
}

func (c *PrintConsumer) generatePrintCommands(storeName string, msg PrintMessage) ([]byte, error) {
	printDataBytes, err := json.Marshal(msg.PrintData)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal print data: %w", err)
	}

	var orderData OrderPrintData
	if err := json.Unmarshal(printDataBytes, &orderData); err != nil {
		return nil, fmt.Errorf("failed to unmarshal order data: %w", err)
	}

	printer := escpos.NewPrinter()

	if msg.PrintType == "order" {
		return c.generateKitchenPrint(storeName, &orderData)
	}

	return c.generateReceiptPrint(storeName, &orderData)
}

func (c *PrintConsumer) generateKitchenPrint(storeName string, data *OrderPrintData) ([]byte, error) {
	printer := escpos.NewPrinter()

	printer.PrintHeader(storeName, data.OrderNo)

	printer.SetTextBold(true)
	printer.PrintLine("后厨单")
	printer.SetTextBold(false)
	printer.Feed(1)

	if data.TableNo != "" {
		printer.PrintLine(fmt.Sprintf("桌号: %s", data.TableNo))
	}
	printer.PrintLine(fmt.Sprintf("订单类型: %s", c.getOrderTypeName(data.OrderType)))
	printer.PrintSeparator()

	printer.SetTextBold(true)
	printer.PrintLine(fmt.Sprintf("%-20s %3s", "菜品", "数量"))
	printer.PrintSeparator()
	printer.SetTextBold(false)

	for _, item := range data.Items {
		name := item.ProductName
		if item.SKUName != "" && item.SKUName != item.ProductName {
			name = fmt.Sprintf("%s(%s)", item.ProductName, item.SKUName)
		}
		line := fmt.Sprintf("%-20s %3d", truncateString(name, 20), item.Quantity)
		printer.PrintLine(line)
	}

	printer.PrintFooter("请及时备菜")

	return printer.Bytes(), nil
}

func (c *PrintConsumer) generateReceiptPrint(storeName string, data *OrderPrintData) ([]byte, error) {
	printer := escpos.NewPrinter()

	printer.PrintHeader(storeName, data.OrderNo)

	printer.SetTextBold(true)
	printer.PrintLine("结账单")
	printer.SetTextBold(false)
	printer.Feed(1)

	if data.TableNo != "" {
		printer.PrintLine(fmt.Sprintf("桌号: %s", data.TableNo))
	}
	printer.PrintLine(fmt.Sprintf("订单类型: %s", c.getOrderTypeName(data.OrderType)))
	printer.PrintSeparator()

	printer.SetTextBold(true)
	printer.PrintLine(fmt.Sprintf("%-18s %3s %6s %8s", "菜品", "数量", "单价", "金额"))
	printer.PrintSeparator()
	printer.SetTextBold(false)

	for _, item := range data.Items {
		name := item.ProductName
		if item.SKUName != "" && item.SKUName != item.ProductName {
			name = fmt.Sprintf("%s(%s)", item.ProductName, item.SKUName)
		}
		price, _ := strconv.ParseFloat(item.Price, 64)
		subtotal, _ := strconv.ParseFloat(item.Subtotal, 64)
		printer.PrintItemWithWidth(name, item.Quantity, fmt.Sprintf("%.2f", price), fmt.Sprintf("%.2f", subtotal), 18)
	}

	printer.PrintSeparator()

	printer.PrintSummary("合计:", data.TotalAmount)
	printer.PrintSummary("应收:", data.PayAmount)

	if data.PayStatus == 1 {
		printer.PrintLine(fmt.Sprintf("支付方式: %s", c.getPayMethodName(data.PayMethod)))
	}

	if data.PointsEarned > 0 {
		printer.PrintLine(fmt.Sprintf("赠送积分: %d", data.PointsEarned))
	}

	printer.PrintFooter("欢迎下次光临")

	return printer.Bytes(), nil
}

func (c *PrintConsumer) printToNetworkPrinter(printer model.Printer, data []byte) error {
	networkPrinter := escpos.NewNetworkPrinter(printer.IPAddress, 9100)
	return networkPrinter.Print(data)
}

func (c *PrintConsumer) updateOrderPrintStatus(orderID uint) {
	if orderID == 0 {
		return
	}
	_ = database.DB.Model(&model.Order{}).Where("id = ?", orderID).Update("print_status", 1)
}

func (c *PrintConsumer) getOrderTypeName(orderType string) string {
	switch orderType {
	case "dine_in":
		return "堂食"
	case "takeaway":
		return "外卖"
	case "delivery":
		return "配送"
	default:
		return orderType
	}
}

func (c *PrintConsumer) getPayMethodName(payMethod string) string {
	switch payMethod {
	case "wechat":
		return "微信支付"
	case "alipay":
		return "支付宝"
	case "cash":
		return "现金"
	case "card":
		return "刷卡"
	default:
		return payMethod
	}
}

func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	runes := []rune(s)
	if len(runes) <= maxLen {
		return s
	}
	return string(runes[:maxLen-3]) + "..."
}
