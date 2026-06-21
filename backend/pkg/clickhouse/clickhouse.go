package clickhouse

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"stalll-hub-pos/backend/config"
	"time"

	_ "github.com/ClickHouse/clickhouse-go/v2"
)

var DB *sql.DB

func InitClickHouse() {
	cfg := config.AppConfig.ClickHouse
	dsn := fmt.Sprintf("clickhouse://%s:%s@%s:%s/%s",
		cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.DBName)

	var err error
	DB, err = sql.Open("clickhouse", dsn)
	if err != nil {
		log.Fatalf("Failed to open ClickHouse connection: %v", err)
	}

	DB.SetMaxOpenConns(10)
	DB.SetMaxIdleConns(3)
	DB.SetConnMaxLifetime(time.Hour)

	if err := DB.Ping(); err != nil {
		log.Printf("Warning: ClickHouse ping failed (will retry on queries): %v", err)
	} else {
		log.Println("ClickHouse connected successfully")
	}

	initSchema()
}

func initSchema() {
	ctx := context.Background()
	queries := []string{
		`CREATE DATABASE IF NOT EXISTS stall_hub_pos`,
		`CREATE TABLE IF NOT EXISTS stall_hub_pos.ch_orders (
			order_id        UInt64,
			order_no        String,
			store_id        UInt64,
			member_id       UInt64,
			order_type      String,
			total_amount    Decimal(12,2),
			discount_amount Decimal(12,2),
			coupon_amount   Decimal(12,2),
			pay_amount      Decimal(12,2),
			pay_method      String,
			order_status    Int32,
			pay_status      Int32,
			source          String,
			created_at      DateTime,
			created_date    Date MATERIALIZED toDate(created_at),
			created_hour    UInt8 MATERIALIZED toHour(created_at),
			updated_at      DateTime DEFAULT now()
		) ENGINE = ReplacingMergeTree(updated_at)
		PARTITION BY toYYYYMM(created_at)
		ORDER BY (store_id, order_id)`,
		`CREATE TABLE IF NOT EXISTS stall_hub_pos.ch_order_items (
			item_id         UInt64,
			order_id        UInt64,
			product_id      UInt64,
			sku_id          UInt64,
			stall_id        UInt64,
			store_id        UInt64,
			product_name    String,
			sku_name        String,
			price           Decimal(10,2),
			quantity        Int32,
			subtotal        Decimal(10,2),
			status          Int32,
			created_at      DateTime,
			created_date    Date MATERIALIZED toDate(created_at),
			updated_at      DateTime DEFAULT now()
		) ENGINE = ReplacingMergeTree(updated_at)
		PARTITION BY toYYYYMM(created_at)
		ORDER BY (store_id, item_id)`,
		`CREATE TABLE IF NOT EXISTS stall_hub_pos.ch_sync_watermark (
			sync_type   String,
			store_id    UInt64,
			last_sync_id UInt64,
			last_sync_time DateTime,
			sync_status Int32,
			updated_at  DateTime DEFAULT now()
		) ENGINE = ReplacingMergeTree(updated_at)
		ORDER BY (sync_type, store_id)`,
	}

	for _, q := range queries {
		if err := DB.ExecContext(ctx, q); err != nil {
			log.Printf("ClickHouse schema init (may already exist): %v", err)
		}
	}
	log.Println("ClickHouse schema initialized (ReplacingMergeTree)")
}

func OptimizeTable(table string) {
	ctx := context.Background()
	query := fmt.Sprintf("OPTIMIZE TABLE stall_hub_pos.%s FINAL", table)
	if err := DB.ExecContext(ctx, query); err != nil {
		log.Printf("ClickHouse optimize %s: %v", table, err)
	}
}

func GetWatermark(syncType string, storeID uint) (uint64, time.Time, error) {
	ctx := context.Background()
	var lastID uint64
	var lastTime time.Time

	row := DB.QueryRowContext(ctx,
		`SELECT last_sync_id, last_sync_time FROM stall_hub_pos.ch_sync_watermark 
		 WHERE sync_type = ? AND store_id = ? ORDER BY updated_at DESC LIMIT 1`,
		syncType, storeID)

	err := row.Scan(&lastID, &lastTime)
	if err == sql.ErrNoRows {
		return 0, time.Time{}, nil
	}
	return lastID, lastTime, err
}

func UpdateWatermark(syncType string, storeID uint, lastID uint64, lastTime time.Time, status int32) error {
	ctx := context.Background()
	_, err := DB.ExecContext(ctx,
		`INSERT INTO stall_hub_pos.ch_sync_watermark 
		 (sync_type, store_id, last_sync_id, last_sync_time, sync_status) 
		 VALUES (?, ?, ?, ?, ?)`,
		syncType, storeID, lastID, lastTime, status)
	return err
}
