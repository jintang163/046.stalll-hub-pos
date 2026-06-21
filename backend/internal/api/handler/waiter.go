package handler

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
	"stalll-hub-pos/backend/internal/dto"
	"stalll-hub-pos/backend/internal/middleware"
	"stalll-hub-pos/backend/internal/model"
	"stalll-hub-pos/backend/internal/service"
	"stalll-hub-pos/backend/pkg/database"
	"stalll-hub-pos/backend/pkg/redis"
)

var waiterUpgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

type WaiterWSClient struct {
	conn    *websocket.Conn
	storeID uint
	userID  uint
}

type WaiterHandler struct {
	orderService *service.OrderService
	wsClients    map[uint][]*WaiterWSClient
	wsMutex      sync.RWMutex
}

var waiterHandlerInstance *WaiterHandler

func NewWaiterHandler() *WaiterHandler {
	if waiterHandlerInstance == nil {
		waiterHandlerInstance = &WaiterHandler{
			orderService: service.NewOrderService(),
			wsClients:    make(map[uint][]*WaiterWSClient),
		}
		go waiterHandlerInstance.listenRedisPubSub()
	}
	return waiterHandlerInstance
}

type UpdateCookStatusRequest struct {
	OrderItemIDs []uint `json:"order_item_ids" binding:"required"`
	CookStatus   int    `json:"cook_status" binding:"required,oneof=0 1 2 3"`
}

func (h *WaiterHandler) UpdateItemCookStatus(c *gin.Context) {
	var req UpdateCookStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.Error(c, http.StatusBadRequest, "Invalid request: "+err.Error())
		return
	}

	if err := database.DB.Model(&model.OrderItem{}).
		Where("id IN ?", req.OrderItemIDs).
		Update("cook_status", req.CookStatus).Error; err != nil {
		middleware.Error(c, http.StatusInternalServerError, "Failed to update cook status: "+err.Error())
		return
	}

	for _, itemID := range req.OrderItemIDs {
		var item model.OrderItem
		database.DB.Where("id = ?", itemID).First(&item)
		if item.OrderID > 0 {
			h.publishOrderUpdate(item.OrderID)
		}
	}

	middleware.Success(c, gin.H{"message": "Cook status updated successfully"})
}

type ServeItemRequest struct {
	OrderItemIDs []uint `json:"order_item_ids" binding:"required"`
}

func (h *WaiterHandler) MarkItemsServed(c *gin.Context) {
	var req ServeItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.Error(c, http.StatusBadRequest, "Invalid request: "+err.Error())
		return
	}

	if err := database.DB.Model(&model.OrderItem{}).
		Where("id IN ?", req.OrderItemIDs).
		Update("cook_status", 3).Error; err != nil {
		middleware.Error(c, http.StatusInternalServerError, "Failed to mark items as served: "+err.Error())
		return
	}

	for _, itemID := range req.OrderItemIDs {
		var item model.OrderItem
		database.DB.Where("id = ?", itemID).First(&item)
		if item.OrderID > 0 {
			h.publishOrderUpdate(item.OrderID)
		}
	}

	middleware.Success(c, gin.H{"message": "Items marked as served successfully"})
}

type CallWaiterRequest struct {
	StoreID   uint   `json:"store_id" binding:"required"`
	TableID   uint   `json:"table_id" binding:"required"`
	TableNo   string `json:"table_no" binding:"required"`
	Content   string `json:"content"`
	CallType  string `json:"call_type" binding:"oneof=service water pay other"`
}

func (h *WaiterHandler) CallWaiter(c *gin.Context) {
	var req CallWaiterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.Error(c, http.StatusBadRequest, "Invalid request: "+err.Error())
		return
	}

	callRecord := &model.WaiterCall{
		StoreID:  req.StoreID,
		TableID:  req.TableID,
		TableNo:  req.TableNo,
		Content:  req.Content,
		CallType: req.CallType,
		Status:   1,
	}
	if err := database.DB.Create(callRecord).Error; err != nil {
		middleware.Error(c, http.StatusInternalServerError, "Failed to create call record: "+err.Error())
		return
	}

	msg := map[string]interface{}{
		"type":       "call_waiter",
		"call_id":    callRecord.ID,
		"store_id":   req.StoreID,
		"table_id":   req.TableID,
		"table_no":   req.TableNo,
		"content":    req.Content,
		"call_type":  req.CallType,
		"created_at": callRecord.CreatedAt,
	}

	msgData, _ := json.Marshal(msg)
	_ = redis.Publish("waiter:call:"+strconv.FormatUint(uint64(req.StoreID), 10), string(msgData))

	h.broadcast(req.StoreID, msg)

	middleware.Success(c, gin.H{
		"message": "Call sent successfully",
		"call_id": callRecord.ID,
	})
}

func (h *WaiterHandler) ListCalls(c *gin.Context) {
	storeID, _ := strconv.ParseUint(c.Query("store_id"), 10, 32)
	status, _ := strconv.Atoi(c.DefaultQuery("status", "1"))

	var calls []model.WaiterCall
	db := database.DB.Where("store_id = ?", storeID)
	if status > 0 {
		db = db.Where("status = ?", status)
	}
	db.Order("created_at DESC").Limit(50).Find(&calls)

	middleware.Success(c, calls)
}

func (h *WaiterHandler) HandleCall(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	userID := middleware.GetUserID(c)

	if err := database.DB.Model(&model.WaiterCall{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":      2,
			"handler_id":  userID,
		}).Error; err != nil {
		middleware.Error(c, http.StatusInternalServerError, "Failed to handle call: "+err.Error())
		return
	}

	middleware.Success(c, gin.H{"message": "Call handled successfully"})
}

func (h *WaiterHandler) WebSocket(c *gin.Context) {
	storeIDStr := c.Query("store_id")
	storeID, err := strconv.ParseUint(storeIDStr, 10, 32)
	if err != nil || storeID == 0 {
		middleware.Error(c, http.StatusBadRequest, "store_id required")
		return
	}

	userIDStr := c.Query("user_id")
	userID, _ := strconv.ParseUint(userIDStr, 10, 32)

	conn, err := waiterUpgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("[WaiterWS] upgrade error: %v", err)
		return
	}

	client := &WaiterWSClient{
		conn:    conn,
		storeID: uint(storeID),
		userID:  uint(userID),
	}

	h.addClient(client)
	defer func() {
		h.removeClient(client)
		conn.Close()
	}()

	conn.WriteJSON(map[string]interface{}{
		"type":    "connected",
		"message": "waiter websocket connected",
	})

	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			break
		}
	}
}

func (h *WaiterHandler) addClient(client *WaiterWSClient) {
	h.wsMutex.Lock()
	defer h.wsMutex.Unlock()
	h.wsClients[client.storeID] = append(h.wsClients[client.storeID], client)
	log.Printf("[WaiterWS] client added, store=%d, user=%d, total=%d",
		client.storeID, client.userID, len(h.wsClients[client.storeID]))
}

func (h *WaiterHandler) removeClient(client *WaiterWSClient) {
	h.wsMutex.Lock()
	defer h.wsMutex.Unlock()

	clients, ok := h.wsClients[client.storeID]
	if !ok {
		return
	}

	for i, c := range clients {
		if c == client {
			h.wsClients[client.storeID] = append(clients[:i], clients[i+1:]...)
			break
		}
	}
	log.Printf("[WaiterWS] client removed, store=%d, total=%d",
		client.storeID, len(h.wsClients[client.storeID]))
}

func (h *WaiterHandler) broadcast(storeID uint, msg interface{}) {
	h.wsMutex.RLock()
	defer h.wsMutex.RUnlock()

	clients, ok := h.wsClients[storeID]
	if !ok || len(clients) == 0 {
		return
	}

	data, err := json.Marshal(msg)
	if err != nil {
		return
	}

	for _, client := range clients {
		if err := client.conn.WriteMessage(websocket.TextMessage, data); err != nil {
			log.Printf("[WaiterWS] write error: %v", err)
		}
	}
}

func (h *WaiterHandler) publishOrderUpdate(orderID uint) {
	var order model.Order
	database.DB.Preload("Items").Where("id = ?", orderID).First(&order)
	if order.ID == 0 {
		return
	}

	msg := map[string]interface{}{
		"type":     "order_update",
		"order_id": orderID,
		"order_no": order.OrderNo,
	}
	msgData, _ := json.Marshal(msg)
	_ = redis.Publish("waiter:order:"+strconv.FormatUint(uint64(order.StoreID), 10), string(msgData))
}

func (h *WaiterHandler) listenRedisPubSub() {
	pubsub := redis.Subscribe("waiter:call:*", "waiter:order:*")
	defer pubsub.Close()

	ch := pubsub.Channel()
	for msg := range ch {
		var data map[string]interface{}
		if err := json.Unmarshal([]byte(msg.Payload), &data); err != nil {
			continue
		}

		storeIDVal, ok := data["store_id"]
		if !ok {
			continue
		}

		var storeID uint
		switch v := storeIDVal.(type) {
		case float64:
			storeID = uint(v)
		case string:
			if n, err := strconv.ParseUint(v, 10, 32); err == nil {
				storeID = uint(n)
			}
		}

		if storeID > 0 {
			h.broadcast(storeID, data)
		}
	}
}

func (h *WaiterHandler) GetTablesWithStatus(c *gin.Context) {
	storeID, _ := strconv.ParseUint(c.Query("store_id"), 10, 32)
	floor, _ := strconv.Atoi(c.DefaultQuery("floor", "0"))
	area := c.Query("area")

	var tables []model.Table
	db := database.DB.Where("store_id = ?", storeID)
	if floor > 0 {
		db = db.Where("floor = ?", floor)
	}
	if area != "" {
		db = db.Where("area = ?", area)
	}
	db.Find(&tables)

	type TableWithOrder struct {
		model.Table
		OrderNo       string  `json:"order_no"`
		OrderAmount   float64 `json:"order_amount"`
		ItemCount     int     `json:"item_count"`
		ServedCount   int     `json:"served_count"`
		DisplayStatus string  `json:"display_status"`
	}

	var result []TableWithOrder
	for _, t := range tables {
		tw := TableWithOrder{
			Table:         t,
			DisplayStatus: "idle",
		}

		if t.CurrentOrderID > 0 {
			var order model.Order
			database.DB.Preload("Items").Where("id = ?", t.CurrentOrderID).First(&order)
			if order.ID > 0 {
				tw.OrderNo = order.OrderNo
				amount, _ := order.TotalAmount.Float64()
				tw.OrderAmount = amount
				tw.ItemCount = len(order.Items)

				servedCount := 0
				for _, item := range order.Items {
					if item.CookStatus >= 3 {
						servedCount++
					}
				}
				tw.ServedCount = servedCount

				if order.PayStatus == 1 || order.OrderStatus >= 4 {
					tw.DisplayStatus = "paid"
				} else if tw.ServedCount == tw.ItemCount && tw.ItemCount > 0 {
					tw.DisplayStatus = "all_served"
				} else {
					tw.DisplayStatus = "ordered"
				}
			} else {
				tw.DisplayStatus = "occupied"
			}
		} else if t.Status == 1 && t.CurrentCustomerCount > 0 {
			tw.DisplayStatus = "occupied"
		}

		result = append(result, tw)
	}

	middleware.Success(c, result)
}

func (h *WaiterHandler) GetWaiterStats(c *gin.Context) {
	storeID, _ := strconv.ParseUint(c.Query("store_id"), 10, 32)

	var totalTables int64
	var idleTables int64
	var occupiedTables int64
	var orderedTables int64
	var pendingCalls int64
	var pendingOrders int64

	database.DB.Model(&model.Table{}).Where("store_id = ? AND status = 1", storeID).Count(&totalTables)
	database.DB.Model(&model.Table{}).Where("store_id = ? AND status = 1 AND current_order_id = 0 AND (current_customer_count = 0 OR current_customer_count IS NULL)", storeID).Count(&idleTables)
	database.DB.Model(&model.Table{}).Where("store_id = ? AND status = 1 AND current_order_id = 0 AND current_customer_count > 0", storeID).Count(&occupiedTables)
	database.DB.Model(&model.Table{}).Where("store_id = ? AND status = 1 AND current_order_id > 0", storeID).Count(&orderedTables)
	database.DB.Model(&model.WaiterCall{}).Where("store_id = ? AND status = 1", storeID).Count(&pendingCalls)
	database.DB.Model(&model.Order{}).Where("store_id = ? AND order_status IN (1,2,3)", storeID).Count(&pendingOrders)

	middleware.Success(c, gin.H{
		"total_tables":    totalTables,
		"idle_tables":     idleTables,
		"occupied_tables": occupiedTables,
		"ordered_tables":  orderedTables,
		"pending_calls":   pendingCalls,
		"pending_orders":  pendingOrders,
	})
}

type AddOrderItemsRequest struct {
	OrderID uint                `json:"order_id" binding:"required"`
	Items   []dto.OrderItemDTO  `json:"items" binding:"required,min=1"`
}

func (h *WaiterHandler) AddOrderItems(c *gin.Context) {
	var req AddOrderItemsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.Error(c, http.StatusBadRequest, "Invalid request: "+err.Error())
		return
	}

	order, err := h.orderService.GetByID(req.OrderID)
	if err != nil {
		middleware.Error(c, http.StatusNotFound, "Order not found")
		return
	}

	var orderModel model.Order
	database.DB.Where("id = ?", req.OrderID).First(&orderModel)
	if orderModel.OrderStatus >= 4 || orderModel.OrderStatus == -1 {
		middleware.Error(c, http.StatusBadRequest, "Cannot add items to current order status")
		return
	}

	var newItems []model.OrderItem
	for _, item := range req.Items {
		var product model.Product
		database.DB.Where("id = ?", item.ProductID).First(&product)
		var sku model.ProductSKU
		database.DB.Where("id = ?", item.SKUID).First(&sku)

		subtotal := item.Price.Mul(decimal.NewFromInt(int64(item.Quantity)))

		newItems = append(newItems, model.OrderItem{
			OrderID:       req.OrderID,
			ProductID:     item.ProductID,
			SKUID:         item.SKUID,
			ProductName:   product.Name,
			SKUName:       sku.Name,
			Image:         sku.Image,
			Price:         item.Price,
			Quantity:      item.Quantity,
			Subtotal:      subtotal,
			Status:        1,
			PrintStatus:   0,
			CookStatus:    0,
		})
	}

	if err := database.DB.Create(&newItems).Error; err != nil {
		middleware.Error(c, http.StatusInternalServerError, "Failed to add items: "+err.Error())
		return
	}

	var totalNewAmount decimal.Decimal
	for _, item := range newItems {
		totalNewAmount = totalNewAmount.Add(item.Subtotal)
	}
	database.DB.Model(&orderModel).UpdateColumn("total_amount", gorm.Expr("total_amount + ?", totalNewAmount))
	database.DB.Model(&orderModel).UpdateColumn("pay_amount", gorm.Expr("pay_amount + ?", totalNewAmount))

	h.publishOrderUpdate(req.OrderID)

	middleware.Success(c, gin.H{"message": "Items added successfully"})
}
