package service

import (
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
	"stalll-hub-pos/backend/config"
	"stalll-hub-pos/backend/internal/dto"
	"stalll-hub-pos/backend/internal/model"
	"stalll-hub-pos/backend/internal/repository"
	"stalll-hub-pos/backend/pkg/nsq"
)

type OrderService struct {
	orderRepo   *repository.OrderRepository
	productRepo *repository.ProductRepository
	nsqProducer *nsq.Producer
	cfg         *config.Config
}

func NewOrderService(
	orderRepo *repository.OrderRepository,
	productRepo *repository.ProductRepository,
	nsqProducer *nsq.Producer,
	cfg *config.Config,
) *OrderService {
	return &OrderService{
		orderRepo:   orderRepo,
		productRepo: productRepo,
		nsqProducer: nsqProducer,
		cfg:         cfg,
	}
}

func (s *OrderService) generateOrderNo() string {
	now := time.Now()
	n, _ := rand.Int(rand.Reader, big.NewInt(1000000))
	return fmt.Sprintf("DD%s%06d", now.Format("20060102150405"), n.Int64())
}

func (s *OrderService) generateRefundNo() string {
	now := time.Now()
	n, _ := rand.Int(rand.Reader, big.NewInt(1000000))
	return fmt.Sprintf("RF%s%06d", now.Format("20060102150405"), n.Int64())
}

func (s *OrderService) attributeValuesToString(attrs []dto.AttributeValue) string {
	if len(attrs) == 0 {
		return ""
	}
	var parts []string
	for _, attr := range attrs {
		parts = append(parts, fmt.Sprintf("%s:%s", attr.Name, attr.Value))
	}
	return strings.Join(parts, ", ")
}

func (s *OrderService) Create(req *dto.CreateOrderRequest) (*dto.CreateOrderResponse, error) {
	orderNo := s.generateOrderNo()

	totalAmount := decimal.Zero
	var orderItems []model.OrderItem

	for _, item := range req.Items {
		sku, err := s.productRepo.GetSKUByID(item.SKUID)
		if err != nil {
			return nil, fmt.Errorf("SKU not found: %w", err)
		}
		if sku.Status != 1 {
			return nil, errors.New("SKU is offline")
		}
		if sku.Stock < item.Quantity {
			return nil, fmt.Errorf("insufficient stock for SKU %s", sku.Name)
		}

		product, err := s.productRepo.GetByID(item.ProductID)
		if err != nil {
			return nil, fmt.Errorf("product not found: %w", err)
		}

		subtotal := item.Price.Mul(decimal.NewFromInt(int64(item.Quantity)))
		totalAmount = totalAmount.Add(subtotal)

		orderItems = append(orderItems, model.OrderItem{
			ProductID:       item.ProductID,
			SKUID:           item.SKUID,
			ProductName:     product.Name,
			SKUName:         sku.Name,
			AttributeValues: s.attributeValuesToString(item.AttributeValues),
			Image:           sku.Image,
			Price:           item.Price,
			Quantity:        item.Quantity,
			Subtotal:        subtotal,
			Status:          1,
			PrintStatus:     0,
			CookStatus:      0,
		})
	}

	discountAmount := decimal.Zero
	couponAmount := decimal.Zero
	pointsUsed := req.PointsUsed
	var pointsEarned int

	if pointsUsed > 0 {
		discountAmount = discountAmount.Add(decimal.NewFromInt(int64(pointsUsed)).Div(decimal.NewFromInt(100)))
	}

	payAmount := totalAmount.Sub(discountAmount).Sub(couponAmount)
	if payAmount.LessThan(decimal.Zero) {
		payAmount = decimal.Zero
	}

	if payAmount.GreaterThan(decimal.Zero) {
		pointsEarned = int(payAmount.IntPart())
	}

	order := &model.Order{
		OrderNo:        orderNo,
		StoreID:        req.StoreID,
		MemberID:       req.MemberID,
		TableNo:        req.TableNo,
		OrderType:      req.OrderType,
		TotalAmount:    totalAmount,
		DiscountAmount: discountAmount,
		CouponAmount:   couponAmount,
		PayAmount:      payAmount,
		PayStatus:      0,
		OrderStatus:    1,
		PrintStatus:    0,
		PointsEarned:   pointsEarned,
		PointsUsed:     pointsUsed,
		CouponID:       req.CouponID,
		Remark:         req.Remark,
		Source:         req.Source,
		Items:          orderItems,
	}

	err := s.orderRepo.CreateWithItems(order)
	if err != nil {
		return nil, fmt.Errorf("create order failed: %w", err)
	}

	if s.nsqProducer != nil {
		orderData, _ := json.Marshal(map[string]interface{}{
			"order_id":    order.ID,
			"order_no":    order.OrderNo,
			"store_id":    order.StoreID,
			"order_type":  order.OrderType,
			"pay_amount":  order.PayAmount.String(),
			"item_count":  len(order.Items),
			"created_at":  order.CreatedAt,
		})
		_ = s.nsqProducer.Publish(nsq.TopicOrderCreated, orderData)
	}

	return &dto.CreateOrderResponse{
		OrderID:   order.ID,
		OrderNo:   order.OrderNo,
		PayAmount: order.PayAmount,
	}, nil
}

func (s *OrderService) GetByID(id uint) (*dto.OrderDetailResponse, error) {
	order, err := s.orderRepo.GetByID(id)
	if err != nil {
		return nil, err
	}
	return s.convertToDetailResponse(order), nil
}

func (s *OrderService) GetByOrderNo(orderNo string) (*dto.OrderDetailResponse, error) {
	order, err := s.orderRepo.GetByOrderNo(orderNo)
	if err != nil {
		return nil, err
	}
	return s.convertToDetailResponse(order), nil
}

func (s *OrderService) List(query *dto.OrderQuery) (*dto.PageResponse, error) {
	orders, total, err := s.orderRepo.List(query)
	if err != nil {
		return nil, err
	}

	var list []dto.OrderListResponse
	for _, order := range orders {
		list = append(list, s.convertToListResponse(&order))
	}

	return &dto.PageResponse{
		List:  list,
		Total: total,
		Page:  query.Page,
		Size:  query.PageSize,
	}, nil
}

func (s *OrderService) UpdateStatus(id uint, status int) error {
	order, err := s.orderRepo.GetByID(id)
	if err != nil {
		return err
	}

	if order.OrderStatus == -1 {
		return errors.New("order already cancelled")
	}

	validTransitions := map[int][]int{
		1: {2, -1},
		2: {3, -1},
		3: {4, -1},
		4: {5},
		5: {},
	}

	valid := false
	for _, s := range validTransitions[order.OrderStatus] {
		if s == status {
			valid = true
			break
		}
	}
	if !valid {
		return fmt.Errorf("invalid status transition from %d to %d", order.OrderStatus, status)
	}

	return s.orderRepo.UpdateStatus(id, status)
}

func (s *OrderService) Cancel(id uint, reason string) error {
	order, err := s.orderRepo.GetByID(id)
	if err != nil {
		return err
	}

	if order.PayStatus == 1 {
		return errors.New("paid order cannot be cancelled, please apply for refund")
	}

	if order.OrderStatus >= 3 {
		return errors.New("order already in production, cannot cancel")
	}

	return s.orderRepo.Cancel(id, reason)
}

func (s *OrderService) Refund(id uint, req *dto.RefundOrderRequest) (uint, error) {
	order, err := s.orderRepo.GetByID(id)
	if err != nil {
		return 0, err
	}

	if order.PayStatus != 1 {
		return 0, errors.New("unpaid order cannot be refunded")
	}

	if order.OrderStatus == -1 {
		return 0, errors.New("order already cancelled")
	}

	refund := &model.OrderRefund{
		OrderID:      id,
		RefundNo:     s.generateRefundNo(),
		RefundAmount: req.RefundAmount,
		RefundReason: req.RefundReason,
		RefundType:   req.RefundType,
		RefundStatus: 0,
		Remark:       req.RefundReason,
	}

	for _, item := range req.Items {
		refund.Items = append(refund.Items, model.RefundItem{
			OrderItemID: item.OrderItemID,
			Quantity:    item.Quantity,
		})
	}

	err = s.orderRepo.CreateRefund(refund)
	if err != nil {
		return 0, fmt.Errorf("create refund failed: %w", err)
	}

	if s.nsqProducer != nil {
		refundData, _ := json.Marshal(map[string]interface{}{
			"refund_id":     refund.ID,
			"refund_no":     refund.RefundNo,
			"order_id":      id,
			"refund_amount": refund.RefundAmount.String(),
			"refund_type":   refund.RefundType,
		})
		_ = s.nsqProducer.Publish(nsq.TopicOrderRefund, refundData)
	}

	return refund.ID, nil
}

func (s *OrderService) GetPaymentParams(req *dto.PaymentParamsRequest) (*dto.PaymentParamsResponse, error) {
	order, err := s.orderRepo.GetByID(req.OrderID)
	if err != nil {
		return nil, err
	}

	if order.PayStatus == 1 {
		return nil, errors.New("order already paid")
	}

	if order.OrderStatus == -1 {
		return nil, errors.New("order already cancelled")
	}

	params := make(map[string]string)

	switch req.PayType {
	case "wechat":
		params["appId"] = s.cfg.Wechat.AppID
		params["timeStamp"] = fmt.Sprintf("%d", time.Now().Unix())
		params["nonceStr"] = s.generateNonceStr()
		params["package"] = fmt.Sprintf("prepay_id=%s", s.generatePrepayID(order.OrderNo))
		params["signType"] = "MD5"
		params["paySign"] = s.generatePaySign(params)

	case "alipay":
		params["partner"] = s.cfg.Alipay.AppID
		params["out_trade_no"] = order.OrderNo
		params["subject"] = fmt.Sprintf("大排档订单-%s", order.OrderNo)
		params["total_fee"] = order.PayAmount.String()
		params["notify_url"] = s.cfg.Alipay.NotifyURL

	case "cash":
		params["cashier"] = "system"
		params["amount"] = order.PayAmount.String()
	}

	return &dto.PaymentParamsResponse{
		PayType: req.PayType,
		OrderID: order.ID,
		OrderNo: order.OrderNo,
		Amount:  order.PayAmount,
		Params:  params,
	}, nil
}

func (s *OrderService) NotifyPayment(orderNo string, payMethod string, transactionID string, amount decimal.Decimal) error {
	order, err := s.orderRepo.GetByOrderNo(orderNo)
	if err != nil {
		return err
	}

	if order.PayStatus == 1 {
		return nil
	}

	if !amount.Equal(order.PayAmount) {
		return fmt.Errorf("payment amount mismatch: expected %s, got %s", order.PayAmount, amount)
	}

	payTime := time.Now()
	err = s.orderRepo.UpdatePayStatus(order.ID, 1, payMethod, &payTime)
	if err != nil {
		return err
	}

	payment := &model.OrderPayment{
		OrderID:       order.ID,
		PayMethod:     payMethod,
		Amount:        amount,
		TransactionID: transactionID,
		PayStatus:     1,
		PayTime:       &payTime,
	}
	err = s.orderRepo.CreatePayment(payment)
	if err != nil {
		return err
	}

	err = s.orderRepo.UpdateStatus(order.ID, 2)
	if err != nil {
		return err
	}

	if s.nsqProducer != nil {
		payData, _ := json.Marshal(map[string]interface{}{
			"order_id":       order.ID,
			"order_no":       order.OrderNo,
			"pay_method":     payMethod,
			"amount":         amount.String(),
			"transaction_id": transactionID,
			"pay_time":       payTime,
		})
		_ = s.nsqProducer.Publish(nsq.TopicOrderPaid, payData)
	}

	return nil
}

func (s *OrderService) BatchCreate(req *dto.BatchOrderRequest) (*dto.BatchOrderResponse, error) {
	var orderIDs []uint
	var errorsList []string
	successCount := 0

	for i, orderReq := range req.Orders {
		resp, err := s.Create(&orderReq)
		if err != nil {
			errorsList = append(errorsList, fmt.Sprintf("order %d: %v", i+1, err))
			continue
		}
		orderIDs = append(orderIDs, resp.OrderID)
		successCount++
	}

	return &dto.BatchOrderResponse{
		SuccessCount: successCount,
		FailCount:    len(req.Orders) - successCount,
		OrderIDs:     orderIDs,
		Errors:       errorsList,
	}, nil
}

func (s *OrderService) GetIncrementalOrders(lastID uint, limit int) ([]model.Order, error) {
	return s.orderRepo.GetIncrementalOrders(lastID, limit)
}

func (s *OrderService) GetOrdersForSync(storeID uint, lastSyncID uint, limit int) ([]model.Order, error) {
	return s.orderRepo.GetOrdersForSync(storeID, lastSyncID, limit)
}

func (s *OrderService) CreateOrderQueue(storeID uint, orderData string) error {
	queue := &model.OrderQueue{
		StoreID:   storeID,
		OrderData: orderData,
		Status:    0,
	}
	return s.orderRepo.CreateOrderQueue(queue)
}

func (s *OrderService) ProcessPendingQueues(storeID uint, limit int) error {
	queues, err := s.orderRepo.GetPendingOrderQueues(storeID, limit)
	if err != nil {
		return err
	}

	for _, queue := range queues {
		var req dto.CreateOrderRequest
		err := json.Unmarshal([]byte(queue.OrderData), &req)
		if err != nil {
			_ = s.orderRepo.UpdateOrderQueueStatus(queue.ID, 2, err.Error())
			continue
		}

		_, err = s.Create(&req)
		if err != nil {
			retryCount := queue.RetryCount + 1
			if retryCount >= 3 {
				_ = s.orderRepo.UpdateOrderQueueStatus(queue.ID, 2, err.Error())
			} else {
				_ = s.orderRepo.UpdateOrderQueueStatus(queue.ID, 0, err.Error())
			}
			continue
		}

		_ = s.orderRepo.UpdateOrderQueueStatus(queue.ID, 1, "")
	}

	return nil
}

func (s *OrderService) convertToListResponse(order *model.Order) dto.OrderListResponse {
	itemCount := 0
	for _, item := range order.Items {
		itemCount += item.Quantity
	}

	storeName := ""
	if order.Store.Name != "" {
		storeName = order.Store.Name
	}

	memberName := ""
	if order.Member != nil {
		memberName = order.Member.Name
	}

	return dto.OrderListResponse{
		ID:             order.ID,
		OrderNo:        order.OrderNo,
		StoreID:        order.StoreID,
		StoreName:      storeName,
		MemberID:       order.MemberID,
		MemberName:     memberName,
		TableNo:        order.TableNo,
		OrderType:      order.OrderType,
		TotalAmount:    order.TotalAmount,
		DiscountAmount: order.DiscountAmount,
		CouponAmount:   order.CouponAmount,
		PayAmount:      order.PayAmount,
		PayStatus:      order.PayStatus,
		OrderStatus:    order.OrderStatus,
		ItemCount:      itemCount,
		Remark:         order.Remark,
		Source:         order.Source,
		CreatedAt:      order.CreatedAt,
	}
}

func (s *OrderService) convertToDetailResponse(order *model.Order) *dto.OrderDetailResponse {
	items := make([]dto.OrderItemDetail, len(order.Items))
	for i, item := range order.Items {
		items[i] = dto.OrderItemDetail{
			ID:              item.ID,
			ProductID:       item.ProductID,
			SKUID:           item.SKUID,
			ProductName:     item.ProductName,
			SKUName:         item.SKUName,
			AttributeValues: item.AttributeValues,
			Image:           item.Image,
			Price:           item.Price,
			Quantity:        item.Quantity,
			Subtotal:        item.Subtotal,
			Status:          item.Status,
			PrintStatus:     item.PrintStatus,
			CookStatus:      item.CookStatus,
		}
	}

	refunds := make([]dto.OrderRefundDetail, len(order.Refunds))
	for i, refund := range order.Refunds {
		refunds[i] = dto.OrderRefundDetail{
			ID:           refund.ID,
			RefundNo:     refund.RefundNo,
			RefundAmount: refund.RefundAmount,
			RefundReason: refund.RefundReason,
			RefundType:   refund.RefundType,
			RefundStatus: refund.RefundStatus,
			RefundTime:   refund.RefundTime,
			Remark:       refund.Remark,
		}
	}

	storeName := ""
	if order.Store.Name != "" {
		storeName = order.Store.Name
	}

	memberName := ""
	if order.Member != nil {
		memberName = order.Member.Name
	}

	return &dto.OrderDetailResponse{
		ID:              order.ID,
		OrderNo:         order.OrderNo,
		StoreID:         order.StoreID,
		StoreName:       storeName,
		MemberID:        order.MemberID,
		MemberName:      memberName,
		TableNo:         order.TableNo,
		OrderType:       order.OrderType,
		TotalAmount:     order.TotalAmount,
		DiscountAmount:  order.DiscountAmount,
		CouponAmount:    order.CouponAmount,
		PayAmount:       order.PayAmount,
		PayMethod:       order.PayMethod,
		PayStatus:       order.PayStatus,
		PayTime:         order.PayTime,
		OrderStatus:     order.OrderStatus,
		PrintStatus:     order.PrintStatus,
		PointsEarned:    order.PointsEarned,
		PointsUsed:      order.PointsUsed,
		CouponID:        order.CouponID,
		Remark:          order.Remark,
		Source:          order.Source,
		Items:           items,
		Refunds:         refunds,
		CreatedAt:       order.CreatedAt,
	}
}

func (s *OrderService) generateNonceStr() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return fmt.Sprintf("%x", b)
}

func (s *OrderService) generatePrepayID(orderNo string) string {
	return fmt.Sprintf("wx%s%d", orderNo, time.Now().Unix())
}

func (s *OrderService) generatePaySign(params map[string]string) string {
	return fmt.Sprintf("%x", params["nonceStr"])
}

var _ = errors.New
var _ = gorm.ErrRecordNotFound
