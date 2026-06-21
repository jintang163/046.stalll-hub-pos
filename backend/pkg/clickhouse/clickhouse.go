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

	DB.SetMaxOpenConns(5)
	DB.SetMaxIdleConns(2)
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
			created_hour    UInt8 MATERIALIZED toHour(created_at)
		) ENGINE = MergeTree()
		PARTITION BY toYYYYMM(created_at)
		ORDER BY (store_id, created_at)`,
		`CREATE TABLE IF NOT EXISTS stall_hub_pos.ch_order_items (
			item_id         UInt64,
			order_id        UInt64,
			product_id      UInt64,
			sku_id          UInt64,
			stall_id        UInt64,
			product_name    String,
			sku_name        String,
			price           Decimal(10,2),
			quantity        Int32,
			subtotal        Decimal(10,2),
			created_at      DateTime,
			created_date    Date MATERIALIZED toDate(created_at)
		) ENGINE = MergeTree()
		PARTITION BY toYYYYMM(created_at)
		ORDER BY (product_id, created_at)`,
	}

	for _, q := range queries {
		if err := DB.ExecContext(ctx, q); err != nil {
			log.Printf("ClickHouse schema init (may already exist): %v", err)
		}
	}
	log.Println("ClickHouse schema initialized")
}
