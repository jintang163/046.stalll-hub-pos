package handler

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"sync"
	"stalll-hub-pos/backend/internal/service"
	"stalll-hub-pos/backend/pkg/redis"
	"stalll-hub-pos/backend/pkg/response"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

type QueueWSClient struct {
	conn    *websocket.Conn
	storeID uint
}

type QueueHandler struct {
	queueService *service.QueueService
	wsClients    map[uint][]*websocket.Conn
	wsMutex      sync.RWMutex
}

var queueHandlerInstance *QueueHandler

func NewQueueHandler() *QueueHandler {
	if queueHandlerInstance == nil {
		queueHandlerInstance = &QueueHandler{
			queueService: service.NewQueueService(),
			wsClients:    make(map[uint][]*websocket.Conn),
		}
		go queueHandlerInstance.listenRedisPubSub()
	}
	return queueHandlerInstance
}

type TakeQueueReq struct {
	StoreID     uint   `json:"store_id"`
	MemberID    uint   `json:"member_id"`
	MemberName  string `json:"member_name"`
	MemberPhone string `json:"member_phone"`
	PeopleCount int    `json:"people_count"`
	Remark      string `json:"remark"`
}

func (h *QueueHandler) TakeNumber(c *gin.Context) {
	var req TakeQueueReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	if req.StoreID == 0 {
		response.Error(c, http.StatusBadRequest, "store_id required")
		return
	}

	result, err := h.queueService.TakeNumber(
		req.StoreID, req.MemberID, req.MemberName, req.MemberPhone,
		req.PeopleCount, req.Remark,
	)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(c, result)
}

func (h *QueueHandler) GetQueueInfo(c *gin.Context) {
	storeID, _ := strconv.ParseUint(c.Query("store_id"), 10, 32)
	queueType := c.DefaultQuery("queue_type", "small")
	queueID := c.Query("queue_id")

	if storeID == 0 {
		response.Error(c, http.StatusBadRequest, "store_id required")
		return
	}

	info, err := h.queueService.GetQueueInfo(uint(storeID), queueType, queueID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(c, info)
}

func (h *QueueHandler) GetAllWaiting(c *gin.Context) {
	storeID, _ := strconv.ParseUint(c.Query("store_id"), 10, 32)
	if storeID == 0 {
		response.Error(c, http.StatusBadRequest, "store_id required")
		return
	}

	result, err := h.queueService.GetAllWaiting(uint(storeID))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(c, result)
}

type CallNumberReq struct {
	StoreID   uint   `json:"store_id"`
	QueueType string `json:"queue_type"`
}

func (h *QueueHandler) CallNumber(c *gin.Context) {
	var req CallNumberReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	if req.StoreID == 0 || req.QueueType == "" {
		response.Error(c, http.StatusBadRequest, "store_id and queue_type required")
		return
	}

	result, err := h.queueService.CallNumber(req.StoreID, req.QueueType)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(c, result)
}

type ArriveReq struct {
	StoreID uint   `json:"store_id"`
	QueueID string `json:"queue_id"`
	TableNo string `json:"table_no"`
}

func (h *QueueHandler) Arrive(c *gin.Context) {
	var req ArriveReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	if req.StoreID == 0 || req.QueueID == "" {
		response.Error(c, http.StatusBadRequest, "store_id and queue_id required")
		return
	}

	err := h.queueService.Arrive(req.StoreID, req.QueueID, req.TableNo)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(c, nil)
}

type CancelReq struct {
	StoreID uint   `json:"store_id"`
	QueueID string `json:"queue_id"`
}

func (h *QueueHandler) Cancel(c *gin.Context) {
	var req CancelReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	if req.StoreID == 0 || req.QueueID == "" {
		response.Error(c, http.StatusBadRequest, "store_id and queue_id required")
		return
	}

	err := h.queueService.Cancel(req.StoreID, req.QueueID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(c, nil)
}

type SavePreOrderReq struct {
	QueueID     string                 `json:"queue_id"`
	StoreID     uint                   `json:"store_id"`
	MemberID    uint                   `json:"member_id"`
	Items       []service.PreOrderItem `json:"items"`
	TotalAmount float64                `json:"total_amount"`
	Remark      string                 `json:"remark"`
}

func (h *QueueHandler) SavePreOrder(c *gin.Context) {
	var req SavePreOrderReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	if req.QueueID == "" {
		response.Error(c, http.StatusBadRequest, "queue_id required")
		return
	}

	err := h.queueService.SavePreOrder(req.QueueID, req.StoreID, req.MemberID, req.Items, req.TotalAmount, req.Remark)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(c, nil)
}

func (h *QueueHandler) GetPreOrder(c *gin.Context) {
	queueID := c.Query("queue_id")
	if queueID == "" {
		response.Error(c, http.StatusBadRequest, "queue_id required")
		return
	}

	preOrder, err := h.queueService.GetPreOrder(queueID)
	if err != nil {
		response.Error(c, http.StatusNotFound, err.Error())
		return
	}

	response.Success(c, preOrder)
}

func (h *QueueHandler) WebSocket(c *gin.Context) {
	storeIDStr := c.Query("store_id")
	storeID, err := strconv.ParseUint(storeIDStr, 10, 32)
	if err != nil || storeID == 0 {
		response.Error(c, http.StatusBadRequest, "store_id required")
		return
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("[WS] upgrade error: %v", err)
		return
	}

	h.addClient(uint(storeID), conn)
	defer func() {
		h.removeClient(uint(storeID), conn)
		conn.Close()
	}()

	conn.WriteJSON(map[string]interface{}{
		"type":    "connected",
		"message": "queue websocket connected",
	})

	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			break
		}
	}
}

func (h *QueueHandler) addClient(storeID uint, conn *websocket.Conn) {
	h.wsMutex.Lock()
	defer h.wsMutex.Unlock()
	h.wsClients[storeID] = append(h.wsClients[storeID], conn)
	log.Printf("[WS] client added, store=%d, total=%d", storeID, len(h.wsClients[storeID]))
}

func (h *QueueHandler) removeClient(storeID uint, conn *websocket.Conn) {
	h.wsMutex.Lock()
	defer h.wsMutex.Unlock()

	clients, ok := h.wsClients[storeID]
	if !ok {
		return
	}

	for i, c := range clients {
		if c == conn {
			h.wsClients[storeID] = append(clients[:i], clients[i+1:]...)
			break
		}
	}
	log.Printf("[WS] client removed, store=%d, total=%d", storeID, len(h.wsClients[storeID]))
}

func (h *QueueHandler) broadcast(storeID uint, msg interface{}) {
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

	for _, conn := range clients {
		if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
			log.Printf("[WS] write error: %v", err)
		}
	}
}

func (h *QueueHandler) listenRedisPubSub() {
	pubsub := redis.Subscribe("queue:call:*")
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
			log.Printf("[WS] broadcast to store %d: %v", storeID, data["type"])
		}
	}
}

func (h *QueueHandler) GetQueueConfig(c *gin.Context) {
	storeID, _ := strconv.ParseUint(c.Query("store_id"), 10, 32)
	if storeID == 0 {
		response.Error(c, http.StatusBadRequest, "store_id required")
		return
	}

	cfg, err := h.queueService.GetQueueConfig(uint(storeID))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(c, cfg)
}
