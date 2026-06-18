package handler

import (
	"net/http"
	"strconv"

	"stalll-hub-pos/backend/internal/dto"
	"stalll-hub-pos/backend/internal/service"
	"stalll-hub-pos/backend/pkg/response"

	"github.com/gin-gonic/gin"
)

type TableHandler struct {
	tableService *service.TableService
}

func NewTableHandler() *TableHandler {
	return &TableHandler{
		tableService: service.NewTableService(),
	}
}

func (h *TableHandler) CreateTable(c *gin.Context) {
	var req dto.TableCreateDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	table, err := h.tableService.CreateTable(&req)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.Success(c, table)
}

func (h *TableHandler) UpdateTable(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	var req dto.TableUpdateDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	err := h.tableService.UpdateTable(uint(id), &req)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.Success(c, nil)
}

func (h *TableHandler) DeleteTable(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	err := h.tableService.DeleteTable(uint(id))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.Success(c, nil)
}

func (h *TableHandler) GetTable(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	table, err := h.tableService.GetTable(uint(id))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.Success(c, table)
}

func (h *TableHandler) ListTables(c *gin.Context) {
	var req dto.TableQueryDTO
	if err := c.ShouldBindQuery(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	tables, total, err := h.tableService.ListTables(&req)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.Page(c, tables, total, req.PageNum, req.PageSize)
}

func (h *TableHandler) GetOccupiedTables(c *gin.Context) {
	storeID, _ := strconv.ParseUint(c.Query("store_id"), 10, 32)
	tables, err := h.tableService.GetOccupiedTables(uint(storeID))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.Success(c, tables)
}

func (h *TableHandler) Checkin(c *gin.Context) {
	var req dto.TableCheckinDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	err := h.tableService.Checkin(&req)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.Success(c, nil)
}

func (h *TableHandler) Checkout(c *gin.Context) {
	var req dto.TableCheckoutDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	err := h.tableService.Checkout(&req)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.Success(c, nil)
}

func (h *TableHandler) GenerateQRCode(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	qrCode, qrCodeUrl, err := h.tableService.GenerateTableQRCode(uint(id))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.Success(c, gin.H{
		"qr_code":      qrCode,
		"qr_code_url":  qrCodeUrl,
	})
}

func (h *TableHandler) ScanQRCode(c *gin.Context) {
	var req dto.TableScanDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	result, err := h.tableService.ScanQRCode(req.Scene)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.Success(c, result)
}

func (h *TableHandler) GetAvailableTables(c *gin.Context) {
	storeID, _ := strconv.ParseUint(c.Query("store_id"), 10, 32)
	peopleCount, _ := strconv.Atoi(c.DefaultQuery("people_count", "2"))
	tables, err := h.tableService.GetAvailableTables(uint(storeID), peopleCount)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.Success(c, tables)
}

func (h *TableHandler) GetStoreMap(c *gin.Context) {
	stores, err := h.tableService.GetStoreMap()
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.Success(c, stores)
}

func (h *TableHandler) BatchCreateTables(c *gin.Context) {
	var req struct {
		StoreID   uint   `json:"store_id" binding:"required"`
		Count     int    `json:"count" binding:"required,min=1,max=100"`
		Prefix    string `json:"prefix" binding:"required,max=10"`
		StartNo   int    `json:"start_no" binding:"min=1"`
		Capacity  int    `json:"capacity" binding:"min=1"`
		Floor     int    `json:"floor" binding:"min=1"`
		Area      string `json:"area"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	tables, err := h.tableService.BatchCreateTables(req.StoreID, req.Count, req.Prefix, req.StartNo, req.Capacity, req.Floor, req.Area)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.Success(c, tables)
}

func (h *TableHandler) CreateArea(c *gin.Context) {
	var req dto.TableAreaCreateDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	area, err := h.tableService.CreateArea(&req)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.Success(c, area)
}

func (h *TableHandler) UpdateArea(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	var req dto.TableAreaUpdateDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	err := h.tableService.UpdateArea(uint(id), &req)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.Success(c, nil)
}

func (h *TableHandler) DeleteArea(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	err := h.tableService.DeleteArea(uint(id))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.Success(c, nil)
}

func (h *TableHandler) ListAreas(c *gin.Context) {
	storeID, _ := strconv.ParseUint(c.Query("store_id"), 10, 32)
	areas, err := h.tableService.ListAreas(uint(storeID))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.Success(c, areas)
}

func (h *TableHandler) CreateReservation(c *gin.Context) {
	var req dto.ReservationCreateDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	reservation, err := h.tableService.CreateReservation(&req)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.Success(c, reservation)
}

func (h *TableHandler) UpdateReservation(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	var req dto.ReservationUpdateDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	err := h.tableService.UpdateReservation(uint(id), &req)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.Success(c, nil)
}

func (h *TableHandler) CancelReservation(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	err := h.tableService.CancelReservation(uint(id))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.Success(c, nil)
}

func (h *TableHandler) CheckinReservation(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	err := h.tableService.CheckinReservation(uint(id))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.Success(c, nil)
}

func (h *TableHandler) GetReservation(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	reservation, err := h.tableService.GetReservation(uint(id))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.Success(c, reservation)
}

func (h *TableHandler) ListReservations(c *gin.Context) {
	var req dto.ReservationQueryDTO
	if err := c.ShouldBindQuery(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	reservations, total, err := h.tableService.ListReservations(&req)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.Page(c, reservations, total, req.PageNum, req.PageSize)
}

func (h *TableHandler) GetTimeSlots(c *gin.Context) {
	var req dto.ReservationTimeSlotDTO
	if err := c.ShouldBindQuery(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	slots, err := h.tableService.GetTimeSlots(&req)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.Success(c, slots)
}

func (h *TableHandler) CreateQueue(c *gin.Context) {
	var req dto.QueueCreateDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	queue, err := h.tableService.CreateQueue(&req)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.Success(c, queue)
}

func (h *TableHandler) CallQueue(c *gin.Context) {
	var req dto.QueueCallDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	queue, err := h.tableService.CallQueue(&req)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.Success(c, queue)
}

func (h *TableHandler) CallNextQueue(c *gin.Context) {
	storeID, _ := strconv.ParseUint(c.Param("store_id"), 10, 32)
	queueType := c.Query("queue_type")
	queue, err := h.tableService.CallNextQueue(uint(storeID), queueType)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.Success(c, queue)
}

func (h *TableHandler) CancelQueue(c *gin.Context) {
	var req dto.QueueCancelDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	err := h.tableService.CancelQueue(&req)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.Success(c, nil)
}

func (h *TableHandler) ArriveQueue(c *gin.Context) {
	var req struct {
		QueueID uint   `json:"queue_id" binding:"required"`
		TableID uint   `json:"table_id" binding:"required"`
		TableNo string `json:"table_no" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	err := h.tableService.ArriveQueue(req.QueueID, req.TableID, req.TableNo)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.Success(c, nil)
}

func (h *TableHandler) GetQueue(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	queue, err := h.tableService.GetQueue(uint(id))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.Success(c, queue)
}

func (h *TableHandler) ListQueues(c *gin.Context) {
	var req dto.QueueQueryDTO
	if err := c.ShouldBindQuery(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	queues, total, err := h.tableService.ListQueues(&req)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.Page(c, queues, total, req.PageNum, req.PageSize)
}

func (h *TableHandler) GetQueueStatus(c *gin.Context) {
	var req dto.QueueStatusDTO
	if err := c.ShouldBindQuery(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	status, err := h.tableService.GetQueueStatus(&req)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.Success(c, status)
}

func (h *TableHandler) GetQueueConfig(c *gin.Context) {
	storeID, _ := strconv.ParseUint(c.Query("store_id"), 10, 32)
	config, err := h.tableService.GetQueueConfig(uint(storeID))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.Success(c, config)
}

func (h *TableHandler) SaveQueueConfig(c *gin.Context) {
	var req dto.QueueConfigDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	err := h.tableService.SaveQueueConfig(&req)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.Success(c, nil)
}

func (h *TableHandler) GetMyQueue(c *gin.Context) {
	memberID, _ := strconv.ParseUint(c.Query("member_id"), 10, 32)
	storeID, _ := strconv.ParseUint(c.Query("store_id"), 10, 32)
	queues, err := h.tableService.GetMyQueue(uint(memberID), uint(storeID))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.Success(c, queues)
}

func (h *TableHandler) GetWaitingCount(c *gin.Context) {
	storeID, _ := strconv.ParseUint(c.Query("store_id"), 10, 32)
	counts, err := h.tableService.GetWaitingCount(uint(storeID))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.Success(c, counts)
}
