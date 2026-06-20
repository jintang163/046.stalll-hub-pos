package repository

import (
	"time"

	"gorm.io/gorm"
	"stalll-hub-pos/backend/internal/dto"
	"stalll-hub-pos/backend/internal/model"
	"stalll-hub-pos/backend/pkg/database"
)

type OrderRepository struct {
	db *gorm.DB
}

func NewOrderRepository() *OrderRepository {
	return &OrderRepository{db: database.DB}
}

func (r *OrderRepository) Create(order *model.Order) error {
	return r.db.Create(order).Error
}

func (r *OrderRepository) CreateWithItems(order *model.Order) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(order).Error; err != nil {
			return err
		}
		return nil
	})
}

func (r *OrderRepository) GetByID(id uint) (*model.Order, error) {
	var order model.Order
	err := r.db.Preload("Items").Preload("Refunds").
		Preload("Store").Preload("Member").
		First(&order, id).Error
	if err != nil {
		return nil, err
	}
	return &order, nil
}

func (r *OrderRepository) GetByOrderNo(orderNo string) (*model.Order, error) {
	var order model.Order
	err := r.db.Preload("Items").Preload("Refunds").
		Preload("Store").Preload("Member").
		Where("order_no = ?", orderNo).First(&order).Error
	if err != nil {
		return nil, err
	}
	return &order, nil
}

type ProductCategoryInfo struct {
	ProductID    uint   `gorm:"column:product_id"`
	CategoryID   uint   `gorm:"column:category_id"`
	CategoryName string `gorm:"column:category_name"`
}

func (r *OrderRepository) GetProductCategoryMap(productIDs []uint) (map[uint]ProductCategoryInfo, error) {
	if len(productIDs) == 0 {
		return make(map[uint]ProductCategoryInfo), nil
	}
	var infos []ProductCategoryInfo
	err := r.db.Table("products p").
		Select("p.id as product_id, p.category_id as category_id, c.name as category_name").
		Joins("LEFT JOIN categories c ON p.category_id = c.id").
		Where("p.id IN ?", productIDs).
		Scan(&infos).Error
	if err != nil {
		return nil, err
	}
	categoryMap := make(map[uint]ProductCategoryInfo, len(infos))
	for _, info := range infos {
		categoryMap[info.ProductID] = info
	}
	return categoryMap, nil
}

func (r *OrderRepository) List(query *dto.OrderQuery) ([]model.Order, int64, error) {
	var orders []model.Order
	var total int64

	db := r.db.Model(&model.Order{})

	if query.StoreID > 0 {
		db = db.Where("store_id = ?", query.StoreID)
	}
	if query.MemberID > 0 {
		db = db.Where("member_id = ?", query.MemberID)
	}
	if query.OrderStatus > 0 {
		db = db.Where("order_status = ?", query.OrderStatus)
	}
	if query.PayStatus > 0 {
		db = db.Where("pay_status = ?", query.PayStatus)
	}
	if query.OrderType != "" {
		db = db.Where("order_type = ?", query.OrderType)
	}
	if query.OrderNo != "" {
		db = db.Where("order_no LIKE ?", "%"+query.OrderNo+"%")
	}
	if query.StartDate != "" {
		db = db.Where("created_at >= ?", query.StartDate)
	}
	if query.EndDate != "" {
		db = db.Where("created_at <= ?", query.EndDate+" 23:59:59")
	}

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (query.Page - 1) * query.PageSize
	err := db.Preload("Items").Preload("Store").Preload("Member").
		Order("id DESC").
		Offset(offset).Limit(query.PageSize).
		Find(&orders).Error

	return orders, total, err
}

func (r *OrderRepository) Update(order *model.Order) error {
	return r.db.Save(order).Error
}

func (r *OrderRepository) UpdateStatus(id uint, status int) error {
	return r.db.Model(&model.Order{}).Where("id = ?", id).
		Update("order_status", status).Error
}

func (r *OrderRepository) UpdatePayStatus(id uint, payStatus int, payMethod string, payTime *time.Time) error {
	updates := map[string]interface{}{
		"pay_status": payStatus,
		"pay_method": payMethod,
	}
	if payTime != nil {
		updates["pay_time"] = payTime
	}
	return r.db.Model(&model.Order{}).Where("id = ?", id).Updates(updates).Error
}

func (r *OrderRepository) Cancel(id uint, reason string) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&model.Order{}).Where("id = ?", id).
			Updates(map[string]interface{}{
				"order_status": -1,
				"remark":       reason,
			}).Error; err != nil {
			return err
		}

		var items []model.OrderItem
		if err := tx.Where("order_id = ?", id).Find(&items).Error; err != nil {
			return err
		}

		for _, item := range items {
			if err := tx.Model(&model.SKU{}).Where("id = ?", item.SKUID).
				UpdateColumn("stock", gorm.Expr("stock + ?", item.Quantity)).Error; err != nil {
				return err
			}
		}

		return nil
	})
}

func (r *OrderRepository) CreateRefund(refund *model.OrderRefund) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(refund).Error; err != nil {
			return err
		}

		for _, item := range refund.Items {
			if err := tx.Model(&model.OrderItem{}).Where("id = ?", item.OrderItemID).
				UpdateColumn("status", 0).Error; err != nil {
				return err
			}

			if err := tx.Model(&model.SKU{}).Where("id = ?",
				tx.Model(&model.OrderItem{}).Select("sku_id").Where("id = ?", item.OrderItemID)).
				UpdateColumn("stock", gorm.Expr("stock + ?", item.Quantity)).Error; err != nil {
				return err
			}
		}

		return nil
	})
}

func (r *OrderRepository) GetRefundByID(id uint) (*model.OrderRefund, error) {
	var refund model.OrderRefund
	err := r.db.Preload("Items").Preload("Items.OrderItem").
		Preload("Order").First(&refund, id).Error
	if err != nil {
		return nil, err
	}
	return &refund, nil
}

func (r *OrderRepository) UpdateRefundStatus(id uint, status int) error {
	return r.db.Model(&model.OrderRefund{}).Where("id = ?", id).
		Updates(map[string]interface{}{
			"refund_status": status,
			"refund_time":   time.Now(),
		}).Error
}

func (r *OrderRepository) CreatePayment(payment *model.OrderPayment) error {
	return r.db.Create(payment).Error
}

func (r *OrderRepository) GetPaymentsByOrderID(orderID uint) ([]model.OrderPayment, error) {
	var payments []model.OrderPayment
	err := r.db.Where("order_id = ?", orderID).Find(&payments).Error
	return payments, err
}

func (r *OrderRepository) CreateSyncRecord(record *model.SyncRecord) error {
	return r.db.Create(record).Error
}

func (r *OrderRepository) UpdateSyncRecord(record *model.SyncRecord) error {
	return r.db.Save(record).Error
}

func (r *OrderRepository) GetLastSyncRecord(storeID uint, syncType string) (*model.SyncRecord, error) {
	var record model.SyncRecord
	err := r.db.Where("store_id = ? AND sync_type = ?", storeID, syncType).
		Order("id DESC").First(&record).Error
	if err != nil {
		return nil, err
	}
	return &record, nil
}

func (r *OrderRepository) CreateOrderQueue(queue *model.OrderQueue) error {
	return r.db.Create(queue).Error
}

func (r *OrderRepository) GetPendingOrderQueues(storeID uint, limit int) ([]model.OrderQueue, error) {
	var queues []model.OrderQueue
	err := r.db.Where("store_id = ? AND status = 0", storeID).
		Order("id ASC").Limit(limit).Find(&queues).Error
	return queues, err
}

func (r *OrderRepository) UpdateOrderQueueStatus(id uint, status int, errorMsg string) error {
	updates := map[string]interface{}{
		"status": status,
	}
	if errorMsg != "" {
		updates["error_msg"] = errorMsg
		updates["retry_count"] = gorm.Expr("retry_count + 1")
	}
	return r.db.Model(&model.OrderQueue{}).Where("id = ?", id).Updates(updates).Error
}

func (r *OrderRepository) GetOrdersForSync(storeID uint, lastSyncID uint, limit int) ([]model.Order, error) {
	var orders []model.Order
	err := r.db.Preload("Items").Preload("Payments").Preload("Refunds").
		Where("store_id = ? AND id > ?", storeID, lastSyncID).
		Order("id ASC").Limit(limit).Find(&orders).Error
	return orders, err
}

func (r *OrderRepository) GetIncrementalOrders(lastID uint, limit int) ([]model.Order, error) {
	var orders []model.Order
	err := r.db.Preload("Items").
		Where("id > ?", lastID).
		Order("id ASC").Limit(limit).Find(&orders).Error
	return orders, err
}

func (r *OrderRepository) BatchCreateOrders(orders []model.Order) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		for i := range orders {
			if err := tx.Create(&orders[i]).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *OrderRepository) GetPaidOrdersByStallAndDate(stallID uint, date string) ([]model.Order, error) {
	var orders []model.Order
	err := r.db.Preload("Items").
		Where("pay_status = 1 AND DATE(created_at) = ?", date).
		Where("id IN (SELECT DISTINCT order_id FROM order_items WHERE stall_id = ?)", stallID).
		Order("id ASC").
		Find(&orders).Error
	return orders, err
}

func (r *OrderRepository) GetOrdersByStall(stallID uint, page, pageSize int) ([]model.Order, int64, error) {
	var orders []model.Order
	var total int64

	db := r.db.Model(&model.Order{}).
		Where("id IN (SELECT DISTINCT order_id FROM order_items WHERE stall_id = ?)", stallID)

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	err := db.Preload("Items").Preload("Store").
		Order("id DESC").
		Offset(offset).Limit(pageSize).
		Find(&orders).Error

	return orders, total, err
}
