package main

import (
	"fmt"
	"log"
	"stalll-hub-pos/backend/config"
	"stalll-hub-pos/backend/internal/api"
	"stalll-hub-pos/backend/internal/consumer"
	"stalll-hub-pos/backend/internal/model"
	"stalll-hub-pos/backend/internal/service"
	"stalll-hub-pos/backend/pkg/clickhouse"
	"stalll-hub-pos/backend/pkg/database"
	"stalll-hub-pos/backend/pkg/minio"
	"stalll-hub-pos/backend/pkg/nsq"
	"stalll-hub-pos/backend/pkg/redis"

	"github.com/shopspring/decimal"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	config.LoadConfig()

	database.InitMySQL()
	redis.InitRedis()
	minio.InitMinIO()
	nsq.InitProducer()
	clickhouse.InitClickHouse()

	database.AutoMigrate(
		&model.Store{},
		&model.Printer{},
		&model.StoreUser{},
		&model.Category{},
		&model.Product{},
		&model.ProductSKU{},
		&model.ProductAttribute{},
		&model.AttributeValue{},
		&model.SKUAttributeValue{},
		&model.StockWarning{},
		&model.Order{},
		&model.OrderItem{},
		&model.OrderPayment{},
		&model.OrderRefund{},
		&model.RefundItem{},
		&model.SyncRecord{},
		&model.OrderQueue{},
		&model.Member{},
		&model.MemberLevel{},
		&model.MemberPointsRecord{},
		&model.Coupon{},
		&model.MemberCoupon{},
		&model.Promotion{},
		&model.PromotionTier{},
		&model.DailyReport{},
		&model.ProductSalesReport{},
		&model.CategorySalesReport{},
		&model.HourlyReport{},
		&model.PaymentReport{},
		&model.ReportTask{},
		&model.RecommendConfig{},
		&model.RecommendResult{},
		&model.PointsRule{},
		&model.RechargeActivity{},
		&model.MemberRecharge{},
		&model.Stall{},
		&model.StallDevice{},
		&model.StallUser{},
		&model.StallSettlement{},
		&model.StallSettlementItem{},
		&model.StallDailyReport{},
		&model.StockCheck{},
		&model.StockCheckItem{},
		&model.ProductCost{},
		&model.CostImportBatch{},
		&model.ProfitReport{},
		&model.Ingredient{},
		&model.ProductBOM{},
		&model.IngredientPrice{},
		&model.CostAlert{},
		&model.DeliveryOrder{},
		&model.PickupCode{},
		&model.Rider{},
		&model.DeliveryTracking{},
	)

	initDefaultData()

	initNSQConsumers()

	initRecommendScheduler()

	initMemberScheduler()

	initClickHouseSync()

	initInventorySync()

	initCostAlert()

	r := api.SetupRouter(database.DB, nsq.Producer)

	defer nsq.StopProducer()
	defer nsq.StopConsumers()

	log.Printf("Server starting on port %s", config.AppConfig.Server.Port)
	if err := r.Run(fmt.Sprintf(":%s", config.AppConfig.Server.Port)); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func initDefaultData() {
	var storeCount int64
	database.DB.Model(&model.Store{}).Count(&storeCount)
	if storeCount == 0 {
		store := &model.Store{
			Name:          "大排档总店",
			Address:       "北京市朝阳区xx路xx号",
			Phone:         "13800138000",
			BusinessHours: "10:00-22:00",
			Status:        1,
			Description:   "正宗大排档，地道美食",
		}
		database.DB.Create(store)
		log.Println("Default store created")

		var userCount int64
		database.DB.Model(&model.StoreUser{}).Count(&userCount)
		if userCount == 0 {
			hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
			user := &model.StoreUser{
				StoreID:  store.ID,
				Username: "admin",
				Password: string(hashedPassword),
				RealName: "管理员",
				Phone:    "13800138000",
				Role:     "admin",
				Status:   1,
			}
			database.DB.Create(user)
			log.Println("Default admin user created: admin/admin123")
		}

		initDefaultStallData(store.ID)
	}
}

func initDefaultStallData(storeID uint) {
	var stallCount int64
	database.DB.Model(&model.Stall{}).Where("store_id = ?", storeID).Count(&stallCount)
	if stallCount == 0 {
		stalls := []model.Stall{
			{
				StoreID:       storeID,
				StallNo:       "ST001",
				Name:          "烧烤档",
				RevenueRatio:  decimal.RequireFromString("0.7000"),
				PlatformRatio: decimal.RequireFromString("0.3000"),
				ContactName:   "张师傅",
				ContactPhone:  "13800000001",
				Status:        1,
				Sort:          1,
				Remark:        "主营各类烧烤、烤串",
			},
			{
				StoreID:       storeID,
				StallNo:       "ST002",
				Name:          "饮品档",
				RevenueRatio:  decimal.RequireFromString("0.7500"),
				PlatformRatio: decimal.RequireFromString("0.2500"),
				ContactName:   "李小姐",
				ContactPhone:  "13800000002",
				Status:        1,
				Sort:          2,
				Remark:        "主营饮品、奶茶、咖啡",
			},
			{
				StoreID:       storeID,
				StallNo:       "ST003",
				Name:          "凉菜档",
				RevenueRatio:  decimal.RequireFromString("0.6500"),
				PlatformRatio: decimal.RequireFromString("0.3500"),
				ContactName:   "王师傅",
				ContactPhone:  "13800000003",
				Status:        1,
				Sort:          3,
				Remark:        "主营各类凉菜、卤味",
			},
		}
		for i := range stalls {
			database.DB.Create(&stalls[i])
		}
		log.Printf("Created %d default stalls", len(stalls))

		var stallUserCount int64
		database.DB.Model(&model.StallUser{}).Count(&stallUserCount)
		if stallUserCount == 0 {
			for i, stall := range stalls {
				usernames := []string{"shaokao", "yinpin", "liangcai"}
				names := []string{"张师傅", "李小姐", "王师傅"}
				hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("stall123"), bcrypt.DefaultCost)
				stallUser := &model.StallUser{
					StoreID:    storeID,
					StallID:    stall.ID,
					Username:   usernames[i],
					Password:   string(hashedPassword),
					RealName:   names[i],
					Phone:      stall.ContactPhone,
					Status:     1,
				}
				database.DB.Create(stallUser)
			}
			log.Println("Created default stall users: shaokao/staff123, yinpin/staff123, liangcai/staff123")
		}
	}
}

func initNSQConsumers() {
	if err := consumer.InitAllConsumers(); err != nil {
		log.Fatalf("Failed to initialize NSQ consumers: %v", err)
	}
	log.Println("All NSQ consumers initialized successfully")
}

func initRecommendScheduler() {
	recService := service.NewRecommendService()
	recService.StartAutoRefreshScheduler()
	log.Println("Recommendation auto-refresh scheduler started")
}

func initMemberScheduler() {
	schedulerService := service.NewSchedulerService()
	schedulerService.StartAllSchedulers()
	log.Println("Member system scheduler started (birthday coupons, recharge activities)")
}

func initClickHouseSync() {
	chSyncService := service.NewClickHouseSyncService()
	chSyncService.StartSyncScheduler()
}

func initInventorySync() {
	if !config.AppConfig.Inventory.Enabled {
		log.Println("[Config] inventory.enabled=false, skipping inventory sync. " +
			"Set INVENTORY_ENABLED=true or inventory.enabled=true in config.yaml to enable.")
		return
	}
	if config.AppConfig.Inventory.BaseURL == "" {
		log.Println("[Config] inventory.base_url is empty, skipping inventory sync.")
		return
	}
	inventoryService := service.NewInventorySyncService()
	inventoryService.StartSyncScheduler()
}

func initCostAlert() {
	if !config.AppConfig.CostAlert.Enabled {
		log.Println("[Config] cost_alert.enabled=false, skipping cost alert. " +
			"Set COST_ALERT_ENABLED=true or cost_alert.enabled=true in config.yaml to enable.")
		return
	}
	if config.AppConfig.DingTalk.Webhook == "" {
		log.Println("[Config] dingtalk.webhook is empty, cost alert DingTalk notifications will be disabled. " +
			"Set DINGTALK_WEBHOOK or dingtalk.webhook in config.yaml to enable notifications.")
	}
	log.Printf("[Config] Cost alert enabled: threshold=%.1f%%, cooldown=%dh, operating_expense_rate=%.1f%%",
		config.AppConfig.CostAlert.PriceChangeThreshold,
		config.AppConfig.CostAlert.CooldownHours,
		config.AppConfig.CostAlert.OperatingExpenseRate)
}
