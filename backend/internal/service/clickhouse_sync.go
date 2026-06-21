package service

import (
	"context"
	"fmt"
	"log"
	"stalll-hub-pos/backend/pkg/clickhouse"
	"stalll-hub-pos/backend/pkg/database"
	"time"

	"github.com/shopspring/decimal"
	"stalll-hub-pos/backend/internal/model"
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

	s.SyncOrders()

	for range ticker.C {
		s.SyncOrders()
	}
}

func (s *ClickHouseSyncService) SyncOrders() {
	if clickhouse.DB == nil {
		return
	}

	now := time.Now()
	startTime := now.Add(-2 * time.Hour).Format("2006-01-02 15:04:05")
	endTime := now.Format("2006-01-02 15:04:05")

	log.Printf("[ClickHouse] Syncing orders from %s to %s", startTime, endTime)

	var orders []model.Order
	if err := database.DB.Where("created_at >= ? AND created_at < ?", startTime, endTime).
		Find(&orders).Error; err != nil {
		log.Printf("[ClickHouse] Failed to fetch orders: %v", err)
		return
	}

	if len(orders) == 0 {
		log.Println("[ClickHouse] No new orders to sync")
		return
	}

	ctx := context.Background()
	tx, err := clickhouse.DB.BeginTx(ctx, nil)
	if err != nil {
		log.Printf("[ClickHouse] Failed to begin transaction: %v", err)
		return
	}

	orderStmt, err := tx.PrepareContext(ctx, `INSERT INTO stall_hub_pos.ch_orders (
		order_id, order_no, store_id, member_id, order_type,
		total_amount, discount_amount, coupon_amount, pay_amount,
		pay_method, order_status, pay_status, source, created_at
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`)
	if err != nil {
		tx.Rollback()
		log.Printf("[ClickHouse] Failed to prepare order stmt: %v", err)
		return
	}
	defer orderStmt.Close()

	var orderIDs []uint
	for _, o := range orders {
		orderIDs = append(orderIDs, o.ID)
		_, err := orderStmt.ExecContext(ctx,
			o.ID, o.OrderNo, o.StoreID, o.MemberID, o.OrderType,
			o.TotalAmount.String(), o.DiscountAmount.String(), o.CouponAmount.String(), o.PayAmount.String(),
			o.PayMethod, o.OrderStatus, o.PayStatus, o.Source, o.CreatedAt,
		)
		if err != nil {
			log.Printf("[ClickHouse] Failed to insert order %d: %v", o.ID, err)
		}
	}

	var items []model.OrderItem
	if err := database.DB.Where("order_id IN ?", orderIDs).Find(&items).Error; err != nil {
		log.Printf("[ClickHouse] Failed to fetch order items: %v", err)
	} else if len(items) > 0 {
		itemStmt, err := tx.PrepareContext(ctx, `INSERT INTO stall_hub_pos.ch_order_items (
			item_id, order_id, product_id, sku_id, stall_id,
			product_name, sku_name, price, quantity, subtotal, created_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`)
		if err != nil {
			log.Printf("[ClickHouse] Failed to prepare item stmt: %v", err)
		} else {
			defer itemStmt.Close()
			for _, item := range items {
				_, err := itemStmt.ExecContext(ctx,
					item.ID, item.OrderID, item.ProductID, item.SKUID, item.StallID,
					item.ProductName, item.SKUName, item.Price.String(), item.Quantity, item.Subtotal.String(), item.CreatedAt,
				)
				if err != nil {
					log.Printf("[ClickHouse] Failed to insert item %d: %v", item.ID, err)
				}
			}
		}
	}

	if err := tx.Commit(); err != nil {
		log.Printf("[ClickHouse] Failed to commit transaction: %v", err)
		return
	}

	log.Printf("[ClickHouse] Synced %d orders and their items", len(orders))
}

type RevenueReport struct {
	StoreID       uint            `json:"store_id"`
	StoreName     string          `json:"store_name"`
	TotalRevenue  decimal.Decimal `json:"total_revenue"`
	OrderCount    int             `json:"order_count"`
	AvgOrderAmount decimal.Decimal `json:"avg_order_amount"`
}

type HourlyTrend struct {
	Hour        int             `json:"hour"`
	OrderCount  int             `json:"order_count"`
	Revenue     decimal.Decimal `json:"revenue"`
}

type TopProduct struct {
	ProductID   uint            `json:"product_id"`
	ProductName string          `json:"product_name"`
	Quantity    int             `json:"quantity"`
	Revenue     decimal.Decimal `json:"revenue"`
}

func (s *ClickHouseSyncService) GetRevenueReport(storeID uint, startDate, endDate string, reportType string) ([]RevenueReport, error) {
	if clickhouse.DB == nil {
		return nil, fmt.Errorf("ClickHouse not connected")
	}

	ctx := context.Background()
	query := `SELECT
		store_id,
		SUM(toDecimal64(pay_amount, 2)) as total_revenue,
		COUNT(*) as order_count
	FROM stall_hub_pos.ch_orders
	WHERE created_date >= ? AND created_date <= ? AND order_status != -1`

	args := []interface{}{startDate, endDate}
	if storeID > 0 {
		query += ` AND store_id = ?`
		args = append(args, storeID)
	}

	if reportType == "monthly" {
		query += ` GROUP BY store_id, toYYYYMM(created_at) ORDER BY toYYYYMM(created_at)`
	} else {
		query += ` GROUP BY store_id ORDER BY store_id`
	}

	rows, err := clickhouse.DB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var reports []RevenueReport
	for rows.Next() {
		var r RevenueReport
		var totalRev string
		if err := rows.Scan(&r.StoreID, &totalRev, &r.OrderCount); err != nil {
			continue
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
		created_hour,
		COUNT(*) as order_count,
		SUM(toDecimal64(pay_amount, 2)) as revenue
	FROM stall_hub_pos.ch_orders
	WHERE created_date >= ? AND created_date <= ? AND order_status != -1`

	args := []interface{}{startDate, endDate}
	if storeID > 0 {
		query += ` AND store_id = ?`
		args = append(args, storeID)
	}
	query += ` GROUP BY created_hour ORDER BY created_hour`

	rows, err := clickhouse.DB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
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
	FROM stall_hub_pos.ch_order_items
	WHERE created_date >= ? AND created_date <= ?`

	args := []interface{}{startDate, endDate}
	if storeID > 0 {
		query += ` AND product_id IN (SELECT product_id FROM stall_hub_pos.ch_orders WHERE store_id = ? AND order_status != -1 AND created_date >= ? AND created_date <= ?)`
		args = append(args, storeID, startDate, endDate)
	}
	query += ` GROUP BY product_id, product_name ORDER BY revenue DESC LIMIT ?`
	args = append(args, topN)

	rows, err := clickhouse.DB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
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
