package service

import (
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"time"

	"gorm.io/gorm"
	"stalll-hub-pos/backend/internal/dto"
	"stalll-hub-pos/backend/internal/model"
	"stalll-hub-pos/backend/pkg/database"
	"stalll-hub-pos/backend/pkg/redis"
)

type DeliveryService struct {
	amapService *AmapService
}

func NewDeliveryService() *DeliveryService {
	return &DeliveryService{
		amapService: NewAmapService(),
	}
}

func (s *DeliveryService) CreateDeliveryOrder(req *dto.DeliveryOrderCreateRequest) (*dto.DeliveryOrderResponse, error) {
	var order model.Order
	if err := database.DB.First(&order, req.OrderID).Error; err != nil {
		return nil, fmt.Errorf("order not found: %w", err)
	}

	if order.OrderType != "delivery" {
		return nil, errors.New("order is not a delivery order")
	}

	if order.PayStatus != 1 {
		return nil, errors.New("order is not paid")
	}

	deliveryOrder := &model.DeliveryOrder{
		OrderID:         req.OrderID,
		StoreID:         order.StoreID,
		DeliveryType:    req.DeliveryType,
		DeliveryStatus:  0,
		SenderName:      req.SenderName,
		SenderPhone:     req.SenderPhone,
		SenderAddress:   req.SenderAddress,
		SenderLng:       req.SenderLng,
		SenderLat:       req.SenderLat,
		ReceiverName:    req.ReceiverName,
		ReceiverPhone:   req.ReceiverPhone,
		ReceiverAddress: req.ReceiverAddress,
		ReceiverLng:     req.ReceiverLng,
		ReceiverLat:     req.ReceiverLat,
	}

	if req.DeliveryType == "self" {
		route, err := s.amapService.PlanRoute(
			req.SenderLng, req.SenderLat,
			req.ReceiverLng, req.ReceiverLat,
		)
		if err == nil {
			deliveryOrder.Distance = route.Distance
			deliveryOrder.Duration = route.Duration
			deliveryOrder.DeliveryFee = route.Fee
			deliveryOrder.RouteData = route.Route
		}

		estTime := time.Now().Add(time.Duration(deliveryOrder.Duration) * time.Minute)
		deliveryOrder.EstimatedTime = &estTime
	}

	if err := database.DB.Create(deliveryOrder).Error; err != nil {
		return nil, fmt.Errorf("create delivery order failed: %w", err)
	}

	if req.DeliveryType == "meituan" {
		s.createMeituanDelivery(deliveryOrder)
	} else if req.DeliveryType == "eleme" {
		s.createElemeDelivery(deliveryOrder)
	}

	return s.convertToResponse(deliveryOrder, &order), nil
}

func (s *DeliveryService) createMeituanDelivery(deliveryOrder *model.DeliveryOrder) {
	fmt.Printf("[Meituan SDK] Creating delivery for order %d, would call Meituan API here\n", deliveryOrder.OrderID)
	deliveryOrder.PlatformType = "meituan"
	deliveryOrder.PlatformOrderID = fmt.Sprintf("MT%d", deliveryOrder.ID)
	database.DB.Save(deliveryOrder)
}

func (s *DeliveryService) createElemeDelivery(deliveryOrder *model.DeliveryOrder) {
	fmt.Printf("[Eleme SDK] Creating delivery for order %d, would call Eleme API here\n", deliveryOrder.OrderID)
	deliveryOrder.PlatformType = "eleme"
	deliveryOrder.PlatformOrderID = fmt.Sprintf("ELM%d", deliveryOrder.ID)
	database.DB.Save(deliveryOrder)
}

func (s *DeliveryService) GetDeliveryOrder(id uint) (*dto.DeliveryOrderResponse, error) {
	var deliveryOrder model.DeliveryOrder
	if err := database.DB.Preload("Order").First(&deliveryOrder, id).Error; err != nil {
		return nil, err
	}
	return s.convertToResponse(&deliveryOrder, &deliveryOrder.Order), nil
}

func (s *DeliveryService) GetDeliveryOrderByOrderID(orderID uint) (*dto.DeliveryOrderResponse, error) {
	var deliveryOrder model.DeliveryOrder
	if err := database.DB.Preload("Order").Where("order_id = ?", orderID).First(&deliveryOrder).Error; err != nil {
		return nil, err
	}
	return s.convertToResponse(&deliveryOrder, &deliveryOrder.Order), nil
}

func (s *DeliveryService) ListDeliveryOrders(query *dto.DeliveryOrderQuery) ([]dto.DeliveryOrderResponse, int64, error) {
	var orders []model.DeliveryOrder
	var total int64

	db := database.DB.Model(&model.DeliveryOrder{})

	if query.StoreID > 0 {
		db = db.Where("store_id = ?", query.StoreID)
	}
	if query.DeliveryType != "" {
		db = db.Where("delivery_type = ?", query.DeliveryType)
	}
	if query.DeliveryStatus > 0 {
		db = db.Where("delivery_status = ?", query.DeliveryStatus)
	}
	if query.OrderID > 0 {
		db = db.Where("order_id = ?", query.OrderID)
	}

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (query.Page - 1) * query.PageSize
	if err := db.Preload("Order").Order("id DESC").Offset(offset).Limit(query.PageSize).Find(&orders).Error; err != nil {
		return nil, 0, err
	}

	var list []dto.DeliveryOrderResponse
	for _, o := range orders {
		list = append(list, *s.convertToResponse(&o, &o.Order))
	}

	return list, total, nil
}

func (s *DeliveryService) UpdateDeliveryStatus(id uint, status int) error {
	var deliveryOrder model.DeliveryOrder
	if err := database.DB.First(&deliveryOrder, id).Error; err != nil {
		return err
	}

	now := time.Now()
	updates := map[string]interface{}{
		"delivery_status": status,
	}

	switch status {
	case 1:
		updates["picked_up_at"] = &now
	case 2:
		updates["delivered_at"] = &now
	case 3:
		updates["delivered_at"] = &now
	}

	return database.DB.Model(&deliveryOrder).Updates(updates).Error
}

func (s *DeliveryService) AssignRider(id uint, riderID uint) error {
	var deliveryOrder model.DeliveryOrder
	if err := database.DB.First(&deliveryOrder, id).Error; err != nil {
		return err
	}

	var rider model.Rider
	if err := database.DB.First(&rider, riderID).Error; err != nil {
		return fmt.Errorf("rider not found: %w", err)
	}

	return database.DB.Model(&deliveryOrder).Updates(map[string]interface{}{
		"rider_id":    riderID,
		"rider_name":  rider.Name,
		"rider_phone": rider.Phone,
		"rider_lng":   rider.CurrentLng,
		"rider_lat":   rider.CurrentLat,
	}).Error
}

func (s *DeliveryService) UpdateRiderLocation(req *dto.RiderLocationUpdate) error {
	var rider model.Rider
	if err := database.DB.First(&rider, req.RiderID).Error; err != nil {
		return fmt.Errorf("rider not found: %w", err)
	}

	if err := database.DB.Model(&rider).Updates(map[string]interface{}{
		"current_lng": req.Lng,
		"current_lat": req.Lat,
	}).Error; err != nil {
		return err
	}

	tracking := &model.DeliveryTracking{
		RiderID:   req.RiderID,
		Lng:       req.Lng,
		Lat:       req.Lat,
		Speed:     req.Speed,
		Heading:   req.Heading,
		Timestamp: time.Now().Unix(),
	}

	var deliveryOrder model.DeliveryOrder
	if err := database.DB.Where("rider_id = ? AND delivery_status IN ?", req.RiderID, []int{0, 1}).First(&deliveryOrder).Error; err == nil {
		tracking.DeliveryOrderID = deliveryOrder.ID
		database.DB.Model(&deliveryOrder).Updates(map[string]interface{}{
			"rider_lng": req.Lng,
			"rider_lat": req.Lat,
		})
	}

	if err := database.DB.Create(tracking).Error; err != nil {
		return err
	}

	cacheKey := fmt.Sprintf("rider:location:%d", req.RiderID)
	locationData := fmt.Sprintf(`{"lng":%f,"lat":%f,"speed":%f,"heading":%f,"ts":%d}`,
		req.Lng, req.Lat, req.Speed, req.Heading, time.Now().Unix())
	redis.Set(cacheKey, locationData, 5*time.Minute)

	redis.Publish(fmt.Sprintf("rider:track:%d", req.RiderID), locationData)

	return nil
}

func (s *DeliveryService) GetRiderLocation(riderID uint) (*dto.RiderLocationResponse, error) {
	var rider model.Rider
	if err := database.DB.First(&rider, riderID).Error; err != nil {
		return nil, fmt.Errorf("rider not found: %w", err)
	}

	cacheKey := fmt.Sprintf("rider:location:%d", riderID)
	data, err := redis.Get(cacheKey)
	if err == nil && data != "" {
		var loc struct {
			Lng     float64 `json:"lng"`
			Lat     float64 `json:"lat"`
			Speed   float64 `json:"speed"`
			Heading float64 `json:"heading"`
			Ts      int64   `json:"ts"`
		}
		if err := parseJSON(data, &loc); err == nil {
			return &dto.RiderLocationResponse{
				RiderID:   riderID,
				RiderName: rider.Name,
				Lng:       loc.Lng,
				Lat:       loc.Lat,
				Speed:     loc.Speed,
				Heading:   loc.Heading,
				UpdatedAt: time.Unix(loc.Ts, 0).Format("2006-01-02 15:04:05"),
			}, nil
		}
	}

	return &dto.RiderLocationResponse{
		RiderID:   riderID,
		RiderName: rider.Name,
		Lng:       rider.CurrentLng,
		Lat:       rider.CurrentLat,
		UpdatedAt: rider.UpdatedAt.Format("2006-01-02 15:04:05"),
	}, nil
}

func (s *DeliveryService) GetDeliveryTracking(orderID uint) (*dto.DeliveryTrackingResponse, error) {
	var deliveryOrder model.DeliveryOrder
	if err := database.DB.Preload("Order").Where("order_id = ?", orderID).First(&deliveryOrder).Error; err != nil {
		return nil, fmt.Errorf("delivery order not found: %w", err)
	}

	var trackings []model.DeliveryTracking
	database.DB.Where("delivery_order_id = ?", deliveryOrder.ID).
		Order("timestamp DESC").Limit(50).Find(&trackings)

	var trackPoints []dto.TrackingPoint
	for _, t := range trackings {
		trackPoints = append(trackPoints, dto.TrackingPoint{
			Lng:       t.Lng,
			Lat:       t.Lat,
			Speed:     t.Speed,
			Timestamp: t.Timestamp,
		})
	}

	return &dto.DeliveryTrackingResponse{
		OrderNo:         deliveryOrder.Order.OrderNo,
		DeliveryStatus:  deliveryOrder.DeliveryStatus,
		RiderID:         deliveryOrder.RiderID,
		RiderName:       deliveryOrder.RiderName,
		RiderPhone:      deliveryOrder.RiderPhone,
		RiderLng:        deliveryOrder.RiderLng,
		RiderLat:        deliveryOrder.RiderLat,
		Distance:        deliveryOrder.Distance,
		Duration:        deliveryOrder.Duration,
		ReceiverAddress: deliveryOrder.ReceiverAddress,
		SenderAddress:   deliveryOrder.SenderAddress,
		EstimatedTime:   deliveryOrder.EstimatedTime,
		Trackings:       trackPoints,
	}, nil
}

func (s *DeliveryService) GeneratePickupCode(orderID uint, storeID uint) (*dto.PickupCodeResponse, error) {
	var order model.Order
	if err := database.DB.First(&order, orderID).Error; err != nil {
		return nil, fmt.Errorf("order not found: %w", err)
	}

	if order.OrderType != "pickup" && order.OrderType != "takeout" {
		return nil, errors.New("order is not a pickup/takeout order")
	}

	if order.PayStatus != 1 {
		return nil, errors.New("order is not paid")
	}

	code := generatePickupCode()

	expiredAt := time.Now().Add(2 * time.Hour)

	pickupCode := &model.PickupCode{
		OrderID:   orderID,
		StoreID:   storeID,
		Code:      code,
		Status:    0,
		ExpiredAt: expiredAt,
	}

	if err := database.DB.Create(pickupCode).Error; err != nil {
		return nil, fmt.Errorf("create pickup code failed: %w", err)
	}

	redisKey := fmt.Sprintf("pickup:code:%s", code)
	redis.Set(redisKey, fmt.Sprintf("%d", orderID), 2*time.Hour)

	database.DB.Model(&order).Update("pickup_code", code)

	return &dto.PickupCodeResponse{
		OrderID: orderID,
		Code:    code,
		Status:  0,
	}, nil
}

func (s *DeliveryService) VerifyPickupCode(code string, storeID uint) (*dto.PickupCodeResponse, error) {
	redisKey := fmt.Sprintf("pickup:code:%s", code)
	data, err := redis.Get(redisKey)
	if err == nil && data != "" {
		var pickupCode model.PickupCode
		if err := database.DB.Where("code = ? AND store_id = ?", code, storeID).First(&pickupCode).Error; err != nil {
			return nil, errors.New("invalid pickup code")
		}

		if pickupCode.Status == 1 {
			return nil, errors.New("pickup code already used")
		}

		if time.Now().After(pickupCode.ExpiredAt) {
			return nil, errors.New("pickup code expired")
		}

		now := time.Now()
		database.DB.Model(&pickupCode).Updates(map[string]interface{}{
			"status":  1,
			"used_at": &now,
		})

		redis.Del(redisKey)

		return &dto.PickupCodeResponse{
			OrderID: pickupCode.OrderID,
			Code:    pickupCode.Code,
			Status:  1,
		}, nil
	}

	var pickupCode model.PickupCode
	if err := database.DB.Where("code = ? AND store_id = ?", code, storeID).First(&pickupCode).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("invalid pickup code")
		}
		return nil, err
	}

	if pickupCode.Status == 1 {
		return nil, errors.New("pickup code already used")
	}

	if time.Now().After(pickupCode.ExpiredAt) {
		return nil, errors.New("pickup code expired")
	}

	now := time.Now()
	database.DB.Model(&pickupCode).Updates(map[string]interface{}{
		"status":  1,
		"used_at": &now,
	})

	return &dto.PickupCodeResponse{
		OrderID: pickupCode.OrderID,
		Code:    pickupCode.Code,
		Status:  1,
	}, nil
}

func (s *DeliveryService) GetPickupCodeByOrderID(orderID uint) (*dto.PickupCodeResponse, error) {
	var pickupCode model.PickupCode
	if err := database.DB.Where("order_id = ?", orderID).First(&pickupCode).Error; err != nil {
		return nil, fmt.Errorf("pickup code not found: %w", err)
	}
	return &dto.PickupCodeResponse{
		OrderID: pickupCode.OrderID,
		Code:    pickupCode.Code,
		Status:  pickupCode.Status,
	}, nil
}

func (s *DeliveryService) CreateRider(storeID uint, req *dto.CreateRiderRequest) (*model.Rider, error) {
	rider := &model.Rider{
		StoreID: storeID,
		Name:    req.Name,
		Phone:   req.Phone,
		Status:  1,
	}
	if err := database.DB.Create(rider).Error; err != nil {
		return nil, fmt.Errorf("create rider failed: %w", err)
	}
	return rider, nil
}

func (s *DeliveryService) ListRiders(storeID uint) ([]model.Rider, error) {
	var riders []model.Rider
	err := database.DB.Where("store_id = ?", storeID).Find(&riders).Error
	return riders, err
}

func (s *DeliveryService) DeleteRider(id uint) error {
	return database.DB.Delete(&model.Rider{}, id).Error
}

func (s *DeliveryService) PlanRoute(req *dto.RoutePlanRequest) (*dto.RoutePlanResponse, error) {
	return s.amapService.PlanRoute(req.OriginLng, req.OriginLat, req.DestLng, req.DestLat)
}

func (s *DeliveryService) Geocode(req *dto.GeocodeRequest) (*dto.GeocodeResponse, error) {
	return s.amapService.Geocode(req.Address, req.City)
}

func (s *DeliveryService) convertToResponse(d *model.DeliveryOrder, order *model.Order) *dto.DeliveryOrderResponse {
	orderNo := ""
	if order != nil {
		orderNo = order.OrderNo
	}
	return &dto.DeliveryOrderResponse{
		ID:              d.ID,
		OrderID:         d.OrderID,
		OrderNo:         orderNo,
		StoreID:         d.StoreID,
		DeliveryType:    d.DeliveryType,
		DeliveryStatus:  d.DeliveryStatus,
		RiderID:         d.RiderID,
		RiderName:       d.RiderName,
		RiderPhone:      d.RiderPhone,
		RiderLng:        d.RiderLng,
		RiderLat:        d.RiderLat,
		DeliveryFee:     d.DeliveryFee,
		Distance:        d.Distance,
		Duration:        d.Duration,
		SenderName:      d.SenderName,
		SenderPhone:     d.SenderPhone,
		SenderAddress:   d.SenderAddress,
		ReceiverName:    d.ReceiverName,
		ReceiverPhone:   d.ReceiverPhone,
		ReceiverAddress: d.ReceiverAddress,
		PlatformOrderID: d.PlatformOrderID,
		PlatformType:    d.PlatformType,
		EstimatedTime:   d.EstimatedTime,
		PickedUpAt:      d.PickedUpAt,
		DeliveredAt:     d.DeliveredAt,
		CreatedAt:       d.CreatedAt,
	}
}

func generatePickupCode() string {
	n, _ := rand.Int(rand.Reader, big.NewInt(100000000))
	return fmt.Sprintf("%08d", n.Int64())
}

func parseJSON(data string, v interface{}) error {
	return json.Unmarshal([]byte(data), v)
}

var _ = json.Unmarshal
