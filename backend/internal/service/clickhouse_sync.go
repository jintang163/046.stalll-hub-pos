package service

import (
	"context"
	"fmt"
	"log"
	"math"
	"stalll-hub-pos/backend/internal/model"
	"stalll-hub-pos/backend/pkg/clickhouse"
	"stalll-hub-pos/backend/pkg/database"
	"time"

	"github.com/shopspring/decimal"
)

const (
	syncTypeOrder     = "order"
	syncTypeOrderItem = "order_item"
	batchSize         = 500
)

type ClickHouseSyncService struct{}

func NewClickHouseSyncService() *ClickHouseSyncService {
	return &ClickHouseSyncService{}
}

func (s *ClickHouseSyncService) StartSyncScheduler() {
	go s.runHourlySync()
	log.Println("[ClickHouse] Hourly sync scheduler started")
}

func (s *ClickHouseSyncService) runHourlySync() {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	s.SyncOrdersIncremental()

	for range ticker.C {
		s.SyncOrdersIncremental()
	}
}

func (s *ClickHouseSyncService) SyncOrdersIncremental() {
	if clickhouse.DB == nil {
		log.Println("[ClickHouse] DB not available, skip sync")
		return
	}

	stores := s.getAllStores()
	totalSynced := 0
	totalItems := 0

	for _, store := range stores {
		orderCount, itemCount, err := s.syncStoreOrders(store.ID)
		if err != nil {
			log.Printf("[ClickHouse] Store %d sync error: %v", store.ID, err)
			continue
		}
		totalSynced += orderCount
		totalItems += itemCount
	}

	if totalSynced > 0 || totalItems > 0 {
		log.Printf("[ClickHouse] Incremental sync done: %d orders, %d items", totalSynced, totalItems)
	}
}

func (s *ClickHouseSyncService) syncStoreOrders(storeID uint) (int, int, error) {
	lastOrderID, lastOrderTime, err := clickhouse.GetWatermark(syncTypeOrder, storeID)
	if err != nil {
		log.Printf("[ClickHouse] Store %d get watermark error: %v", storeID, err)
	}

	var maxID uint64
	var maxTime time.Time
	orderCount := 0
	itemCount := 0

	page := 1
	for {
		var orders []model.Order
		query := database.DB.Where("store_id = ? AND id > ?", storeID, lastOrderID).
			Order("id ASC").
			Limit(batchSize).
			Offset((page - 1) * batchSize)

		if !lastOrderTime.IsZero() {
			query = query.Where("created_at >= ?", lastOrderTime.Add(-5*time.Minute))
		}

		if err := query.Preload("Items").Find(&orders).Error; err != nil {
			return orderCount, itemCount, fmt.Errorf("fetch orders: %v", err)
		}

		if len(orders) == 0 {
			break
		}

		if err := s.insertOrdersBatch(orders); err != nil {
			log.Printf("[ClickHouse] Store %d insert batch error: %v", storeID, err)
		} else {
			orderCount += len(orders)
			for _, o := range orders {
				itemCount += len(o.Items)
				if uint64(o.ID) > maxID {
					maxID = uint64(o.ID)
				}
				if o.CreatedAt.After(maxTime) {
					maxTime = o.CreatedAt
				}
			}
		}

		if len(orders) < batchSize {
			break
		}
		page++
	}

	if maxID > lastOrderID {
		if err := clickhouse.UpdateWatermark(syncTypeOrder, storeID, maxID, maxTime, 1); err != nil {
			log.Printf("[ClickHouse] Store %d update watermark error: %v", storeID, err)
		}
	}

	return orderCount, itemCount, nil
}

func (s *ClickHouseSyncService) insertOrdersBatch(orders []model.Order) error {
	if len(orders) == 0 {
		return nil
	}

	ctx := context.Background()
	tx, err := clickhouse.DB.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %v", err)
	}

	orderStmt, err := tx.PrepareContext(ctx,
		`INSERT INTO stall_hub_pos.ch_orders (
			order_id, order_no, store_id, member_id, order_type,
			total_amount, discount_amount, coupon_amount, pay_amount,
			pay_method, order_status, pay_status, source,
			created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("prepare order stmt: %v", err)
	}
	defer orderStmt.Close()

	itemStmt, err := tx.PrepareContext(ctx,
		`INSERT INTO stall_hub_pos.ch_order_items (
			item_id, order_id, product_id, sku_id, stall_id, store_id,
			product_name, sku_name, price, quantity, subtotal,
			status, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("prepare item stmt: %v", err)
	}
	defer itemStmt.Close()

	for _, o := range orders {
		_, err := orderStmt.ExecContext(ctx,
			o.ID, o.OrderNo, o.StoreID, o.MemberID, o.OrderType,
			o.TotalAmount.String(), o.DiscountAmount.String(), o.CouponAmount.String(), o.PayAmount.String(),
			o.PayMethod, o.OrderStatus, o.PayStatus, o.Source,
			o.CreatedAt, o.UpdatedAt,
		)
		if err != nil {
			log.Printf("[ClickHouse] Insert order %d failed: %v", o.ID, err)
		}

		for _, item := range o.Items {
			_, err := itemStmt.ExecContext(ctx,
				item.ID, item.OrderID, item.ProductID, item.SKUID, item.StallID, o.StoreID,
				item.ProductName, item.SKUName, item.Price.String(), item.Quantity, item.Subtotal.String(),
				item.Status, item.CreatedAt, item.UpdatedAt,
			)
			if err != nil {
				log.Printf("[ClickHouse] Insert order item %d failed: %v", item.ID, err)
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit: %v", err)
	}
	return nil
}

func (s *ClickHouseSyncService) FullBackfill() (int, int, error) {
	if clickhouse.DB == nil {
		return 0, 0, fmt.Errorf("ClickHouse not connected")
	}

	log.Println("[ClickHouse] Starting full backfill...")

	stores := s.getAllStores()
	totalOrders := 0
	totalItems := 0

	for _, store := range stores {
		orderCount, itemCount, err := s.fullBackfillStore(store.ID)
		if err != nil {
			log.Printf("[ClickHouse] Store %d backfill error: %v", store.ID, err)
			continue
		}
		totalOrders += orderCount
		totalItems += itemCount
	}

	clickhouse.OptimizeTable("ch_orders")
	clickhouse.OptimizeTable("ch_order_items")

	log.Printf("[ClickHouse] Full backfill completed: %d orders, %d items", totalOrders, totalItems)
	return totalOrders, totalItems, nil
}

func (s *ClickHouseSyncService) fullBackfillStore(storeID uint) (int, int) {
	var totalOrders int64
	database.DB.Model(&model.Order{}).Where("store_id = ?", storeID).Count(&totalOrders)

	if totalOrders == 0 {
		return 0, 0
	}

	pages := int(math.Ceil(float64(totalOrders) / float64(batchSize)))
	orderCount := 0
	itemCount := 0
	var maxID uint64
	var maxTime time.Time

	for page := 1; page <= pages; page++ {
		var orders []model.Order
		if err := database.DB.Where("store_id = ?", storeID).
			Order("id ASC").
			Limit(batchSize).
			Offset((page - 1) * batchSize).
			Preload("Items").
			Find(&orders).Error; err != nil {
			log.Printf("[ClickHouse] Store %d page %d error: %v", storeID, page, err)
			continue
		}

		if err := s.insertOrdersBatch(orders); err != nil {
			log.Printf("[ClickHouse] Store %d page %d insert error: %v", storeID, page, err)
		} else {
			orderCount += len(orders)
			for _, o := range orders {
				itemCount += len(o.Items)
				if uint64(o.ID) > maxID {
					maxID = uint64(o.ID)
				}
				if o.CreatedAt.After(maxTime) {
					maxTime = o.CreatedAt
				}
			}
		}

		if page%10 == 0 {
			log.Printf("[ClickHouse] Store %d backfill progress: %d/%d", storeID, orderCount, totalOrders)
		}
	}

	if maxID > 0 {
		clickhouse.UpdateWatermark(syncTypeOrder, storeID, maxID, maxTime, 1)
	}

	return orderCount, itemCount
}

func (s *ClickHouseSyncService) getAllStores() []model.Store {
	var stores []model.Store
	if err := database.DB.Where("status = 1").Find(&stores).Error; err != nil {
		log.Printf("[ClickHouse] Failed to fetch stores: %v", err)
		return []model.Store{{ID: 1}}
	}
	return stores
}

type RevenueReport struct {
	StoreID        uint            `json:"store_id"`
	StoreName      string          `json:"store_name"`
	ReportDate     string          `json:"report_date,omitempty"`
	TotalRevenue   decimal.Decimal `json:"total_revenue"`
	OrderCount     int             `json:"order_count"`
	AvgOrderAmount decimal.Decimal `json:"avg_order_amount"`
}

type HourlyTrend struct {
	Hour       int             `json:"hour"`
	OrderCount int             `json:"order_count"`
	Revenue    decimal.Decimal `json:"revenue"`
}

type TopProduct struct {
	ProductID   uint            `json:"product_id"`
	ProductName string          `json:"product_name"`
	Quantity    int             `json:"quantity"`
	Revenue     decimal.Decimal `json:"revenue"`
}

func (s *ClickHouseSyncService) GetRevenueReport(storeID uint, startDate, endDate, reportType string) ([]RevenueReport, error) {
	if clickhouse.DB == nil {
		return nil, fmt.Errorf("ClickHouse not connected")
	}

	ctx := context.Background()
	var query string
	var args []interface{}

	if reportType == "monthly" {
		query = `SELECT
			store_id,
			formatDateTime(toStartOfMonth(created_at), '%Y-%m') as report_date,
			SUM(toDecimal64(pay_amount, 2)) as total_revenue,
			COUNT(*) as order_count
		FROM stall_hub_pos.ch_orders FINAL
		WHERE created_date >= ? AND created_date <= ? AND order_status != -1 AND pay_status = 1`
		args = []interface{}{startDate, endDate}
	} else {
		query = `SELECT
			store_id,
			toString(created_date) as report_date,
			SUM(toDecimal64(pay_amount, 2)) as total_revenue,
			COUNT(*) as order_count
		FROM stall_hub_pos.ch_orders FINAL
		WHERE created_date >= ? AND created_date <= ? AND order_status != -1 AND pay_status = 1`
		args = []interface{}{startDate, endDate}
	}

	if storeID > 0 {
		query += ` AND store_id = ?`
		args = append(args, storeID)
	}

	if reportType == "monthly" {
		query += ` GROUP BY store_id, toStartOfMonth(created_at), report_date ORDER BY report_date, store_id`
	} else {
		query += ` GROUP BY store_id, created_date, report_date ORDER BY report_date DESC, store_id`
	}

	rows, err := clickhouse.DB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query revenue: %v", err)
	}
	defer rows.Close()

	var reports []RevenueReport
	for rows.Next() {
		var r RevenueReport
		var totalRev string
		var storeIDVal uint
		var reportDate string

		if reportType == "monthly" || true {
			if err := rows.Scan(&storeIDVal, &reportDate, &totalRev, &r.OrderCount); err != nil {
				continue
			}
			r.StoreID = storeIDVal
			r.ReportDate = reportDate
		}

		r.TotalRevenue, _ = decimal.NewFromString(totalRev)
		if r.OrderCount > 0 {
			r.AvgOrderAmount = r.TotalRevenue.Div(decimal.NewFromInt(int64(r.OrderCount)))
		}
		reports = append(reports, r)
	}

	return reports, nil
}

func (s *ClickHouseSyncService) GetHourlyTrend(storeID uint, startDate, endDate string) ([]HourlyTrend, error) {
	if clickhouse.DB == nil {
		return nil, fmt.Errorf("ClickHouse not connected")
	}

	ctx := context.Background()
	query := `SELECT
		toHour(created_at) as hour,
		COUNT(*) as order_count,
		SUM(toDecimal64(pay_amount, 2)) as revenue
	FROM stall_hub_pos.ch_orders FINAL
	WHERE created_date >= ? AND created_date <= ? AND order_status != -1 AND pay_status = 1`

	args := []interface{}{startDate, endDate}
	if storeID > 0 {
		query += ` AND store_id = ?`
		args = append(args, storeID)
	}
	query += ` GROUP BY hour ORDER BY hour`

	rows, err := clickhouse.DB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query hourly: %v", err)
	}
	defer rows.Close()

	var trends []HourlyTrend
	for rows.Next() {
		var t HourlyTrend
		var rev string
		if err := rows.Scan(&t.Hour, &t.OrderCount, &rev); err != nil {
			continue
		}
		t.Revenue, _ = decimal.NewFromString(rev)
		trends = append(trends, t)
	}

	return trends, nil
}

func (s *ClickHouseSyncService) GetTopProducts(storeID uint, startDate, endDate string, topN int) ([]TopProduct, error) {
	if clickhouse.DB == nil {
		return nil, fmt.Errorf("ClickHouse not connected")
	}

	if topN <= 0 {
		topN = 10
	}

	ctx := context.Background()
	query := `SELECT
		product_id,
		product_name,
		SUM(quantity) as quantity,
		SUM(toDecimal64(subtotal, 2)) as revenue
	FROM stall_hub_pos.ch_order_items FINAL
	WHERE created_date >= ? AND created_date <= ?`

	args := []interface{}{startDate, endDate}
	if storeID > 0 {
		query += ` AND store_id = ?`
		args = append(args, storeID)
	}
	query += ` GROUP BY product_id, product_name ORDER BY revenue DESC LIMIT ?`
	args = append(args, topN)

	rows, err := clickhouse.DB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query top products: %v", err)
	}
	defer rows.Close()

	var products []TopProduct
	for rows.Next() {
		var p TopProduct
		var rev string
		if err := rows.Scan(&p.ProductID, &p.ProductName, &p.Quantity, &rev); err != nil {
			continue
		}
		p.Revenue, _ = decimal.NewFromString(rev)
		products = append(products, p)
	}

	return products, nil
}
