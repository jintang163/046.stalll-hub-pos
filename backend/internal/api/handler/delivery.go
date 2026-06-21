package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"stalll-hub-pos/backend/internal/dto"
	"stalll-hub-pos/backend/internal/middleware"
	"stalll-hub-pos/backend/internal/service"
)

type DeliveryHandler struct {
	deliveryService *service.DeliveryService
}

func NewDeliveryHandler() *DeliveryHandler {
	return &DeliveryHandler{
		deliveryService: service.NewDeliveryService(),
	}
}

func (h *DeliveryHandler) CreateDeliveryOrder(c *gin.Context) {
	var req dto.DeliveryOrderCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.Error(c, http.StatusBadRequest, "Invalid request: "+err.Error())
		return
	}

	resp, err := h.deliveryService.CreateDeliveryOrder(&req)
	if err != nil {
		middleware.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	middleware.Success(c, resp)
}

func (h *DeliveryHandler) GetDeliveryOrder(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		middleware.Error(c, http.StatusBadRequest, "Invalid delivery order ID")
		return
	}

	resp, err := h.deliveryService.GetDeliveryOrder(uint(id))
	if err != nil {
		middleware.Error(c, http.StatusNotFound, "Delivery order not found")
		return
	}

	middleware.Success(c, resp)
}

func (h *DeliveryHandler) GetDeliveryOrderByOrderID(c *gin.Context) {
	orderID, err := strconv.ParseUint(c.Param("orderId"), 10, 32)
	if err != nil {
		middleware.Error(c, http.StatusBadRequest, "Invalid order ID")
		return
	}

	resp, err := h.deliveryService.GetDeliveryOrderByOrderID(uint(orderID))
	if err != nil {
		middleware.Error(c, http.StatusNotFound, "Delivery order not found")
		return
	}

	middleware.Success(c, resp)
}

func (h *DeliveryHandler) ListDeliveryOrders(c *gin.Context) {
	var query dto.DeliveryOrderQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		middleware.Error(c, http.StatusBadRequest, "Invalid query parameters: "+err.Error())
		return
	}

	if query.Page <= 0 {
		query.Page = 1
	}
	if query.PageSize <= 0 {
		query.PageSize = 10
	}

	list, total, err := h.deliveryService.ListDeliveryOrders(&query)
	if err != nil {
		middleware.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	middleware.PageSuccess(c, list, total, query.Page, query.PageSize)
}

func (h *DeliveryHandler) UpdateDeliveryStatus(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		middleware.Error(c, http.StatusBadRequest, "Invalid delivery order ID")
		return
	}

	var req dto.DeliveryStatusUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.Error(c, http.StatusBadRequest, "Invalid request: "+err.Error())
		return
	}

	if err := h.deliveryService.UpdateDeliveryStatus(uint(id), req.DeliveryStatus); err != nil {
		middleware.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	middleware.Success(c, gin.H{"message": "Delivery status updated successfully"})
}

func (h *DeliveryHandler) AssignRider(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		middleware.Error(c, http.StatusBadRequest, "Invalid delivery order ID")
		return
	}

	var req dto.AssignRiderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.Error(c, http.StatusBadRequest, "Invalid request: "+err.Error())
		return
	}

	if err := h.deliveryService.AssignRider(uint(id), req.RiderID); err != nil {
		middleware.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	middleware.Success(c, gin.H{"message": "Rider assigned successfully"})
}

func (h *DeliveryHandler) UpdateRiderLocation(c *gin.Context) {
	var req dto.RiderLocationUpdate
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.Error(c, http.StatusBadRequest, "Invalid request: "+err.Error())
		return
	}

	if err := h.deliveryService.UpdateRiderLocation(&req); err != nil {
		middleware.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	middleware.Success(c, gin.H{"message": "Rider location updated successfully"})
}

func (h *DeliveryHandler) GetRiderLocation(c *gin.Context) {
	riderID, err := strconv.ParseUint(c.Param("riderId"), 10, 32)
	if err != nil {
		middleware.Error(c, http.StatusBadRequest, "Invalid rider ID")
		return
	}

	resp, err := h.deliveryService.GetRiderLocation(uint(riderID))
	if err != nil {
		middleware.Error(c, http.StatusNotFound, err.Error())
		return
	}

	middleware.Success(c, resp)
}

func (h *DeliveryHandler) GetDeliveryTracking(c *gin.Context) {
	orderID, err := strconv.ParseUint(c.Param("orderId"), 10, 32)
	if err != nil {
		middleware.Error(c, http.StatusBadRequest, "Invalid order ID")
		return
	}

	resp, err := h.deliveryService.GetDeliveryTracking(uint(orderID))
	if err != nil {
		middleware.Error(c, http.StatusNotFound, err.Error())
		return
	}

	middleware.Success(c, resp)
}

func (h *DeliveryHandler) GeneratePickupCode(c *gin.Context) {
	var req struct {
		OrderID uint `json:"order_id" binding:"required"`
		StoreID uint `json:"store_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.Error(c, http.StatusBadRequest, "Invalid request: "+err.Error())
		return
	}

	resp, err := h.deliveryService.GeneratePickupCode(req.OrderID, req.StoreID)
	if err != nil {
		middleware.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	middleware.Success(c, resp)
}

func (h *DeliveryHandler) VerifyPickupCode(c *gin.Context) {
	var req dto.VerifyPickupCodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.Error(c, http.StatusBadRequest, "Invalid request: "+err.Error())
		return
	}

	resp, err := h.deliveryService.VerifyPickupCode(req.Code, req.StoreID)
	if err != nil {
		middleware.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	middleware.Success(c, resp)
}

func (h *DeliveryHandler) GetPickupCodeByOrder(c *gin.Context) {
	orderID, err := strconv.ParseUint(c.Param("orderId"), 10, 32)
	if err != nil {
		middleware.Error(c, http.StatusBadRequest, "Invalid order ID")
		return
	}

	resp, err := h.deliveryService.GetPickupCodeByOrderID(uint(orderID))
	if err != nil {
		middleware.Error(c, http.StatusNotFound, err.Error())
		return
	}

	middleware.Success(c, resp)
}

func (h *DeliveryHandler) CreateRider(c *gin.Context) {
	storeID, err := strconv.ParseUint(c.Param("storeId"), 10, 32)
	if err != nil {
		middleware.Error(c, http.StatusBadRequest, "Invalid store ID")
		return
	}

	var req dto.CreateRiderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.Error(c, http.StatusBadRequest, "Invalid request: "+err.Error())
		return
	}

	rider, err := h.deliveryService.CreateRider(uint(storeID), &req)
	if err != nil {
		middleware.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	middleware.Success(c, rider)
}

func (h *DeliveryHandler) ListRiders(c *gin.Context) {
	storeID, err := strconv.ParseUint(c.Param("storeId"), 10, 32)
	if err != nil {
		middleware.Error(c, http.StatusBadRequest, "Invalid store ID")
		return
	}

	riders, err := h.deliveryService.ListRiders(uint(storeID))
	if err != nil {
		middleware.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	middleware.Success(c, riders)
}

func (h *DeliveryHandler) DeleteRider(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		middleware.Error(c, http.StatusBadRequest, "Invalid rider ID")
		return
	}

	if err := h.deliveryService.DeleteRider(uint(id)); err != nil {
		middleware.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	middleware.Success(c, gin.H{"message": "Rider deleted successfully"})
}

func (h *DeliveryHandler) PlanRoute(c *gin.Context) {
	var req dto.RoutePlanRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.Error(c, http.StatusBadRequest, "Invalid request: "+err.Error())
		return
	}

	resp, err := h.deliveryService.PlanRoute(&req)
	if err != nil {
		middleware.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	middleware.Success(c, resp)
}

func (h *DeliveryHandler) Geocode(c *gin.Context) {
	var req dto.GeocodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.Error(c, http.StatusBadRequest, "Invalid request: "+err.Error())
		return
	}

	resp, err := h.deliveryService.Geocode(&req)
	if err != nil {
		middleware.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	middleware.Success(c, resp)
}
