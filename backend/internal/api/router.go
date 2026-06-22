package api

import (
	"stalll-hub-pos/backend/internal/api/handler"
	"stalll-hub-pos/backend/internal/middleware"
	"stalll-hub-pos/backend/internal/repository"
	"stalll-hub-pos/backend/internal/service"
	"stalll-hub-pos/backend/pkg/config"
	"stalll-hub-pos/backend/pkg/nsq"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func SetupRouter(db *gorm.DB, nsqProducer *nsq.Producer) *gin.Engine {
	r := gin.Default()

	r.Use(middleware.CORS())
	r.Use(gin.Logger())
	r.Use(gin.Recovery())

	authHandler := handler.NewAuthHandler()
	productHandler := handler.NewProductHandler()
	storeHandler := handler.NewStoreHandler()
	memberHandler := handler.NewMemberHandler()
	couponHandler := handler.NewCouponHandler()
	promotionHandler := handler.NewPromotionHandler()
	voiceOrderHandler := handler.NewVoiceOrderHandler()
	reportHandler := handler.NewReportHandler()
	paymentHandler := handler.NewPaymentHandler()
	tableHandler := handler.NewTableHandler()
	recommendHandler := handler.NewRecommendHandler()
	pointsRuleHandler := handler.NewPointsRuleHandler()
	rechargeActivityHandler := handler.NewRechargeActivityHandler()
	stallHandler := handler.NewStallHandler()
	queueHandler := handler.NewQueueHandler()
	stockCheckHandler := handler.NewStockCheckHandler()
	stockWarningHandler := handler.NewStockWarningHandler()
	analyticsHandler := handler.NewAnalyticsHandler()
	ingredientHandler := handler.NewIngredientHandler()
	deliveryHandler := handler.NewDeliveryHandler()
	forecastHandler := handler.NewForecastHandler()
	waiterHandler := handler.NewWaiterHandler()
	facePaymentHandler := handler.NewFacePaymentHandler()
	timeSlotPricingHandler := handler.NewTimeSlotPricingHandler()
	reviewHandler := handler.NewReviewHandler()
	smsHandler := handler.NewSmsHandler()
	transferHandler := handler.NewTransferHandler()

	orderHandler := handler.NewOrderHandler(nil)

	api := r.Group("/api/v1")
	{
		auth := api.Group("/auth")
		{
			auth.POST("/login", authHandler.Login)
			auth.POST("/logout", middleware.JWTAuth(), authHandler.Logout)
			auth.GET("/user", middleware.JWTAuth(), authHandler.GetCurrentUser)
		}

		products := api.Group("/products")
		products.Use(middleware.JWTAuth())
		{
			products.POST("", productHandler.Create)
			products.GET("", productHandler.List)
			products.GET("/:id", productHandler.Get)
			products.PUT("/:id", productHandler.Update)
			products.DELETE("/:id", productHandler.Delete)
			products.POST("/copy", productHandler.Copy)
			products.POST("/batch-price", productHandler.BatchUpdatePrice)
			products.PUT("/stock", productHandler.UpdateStock)
			products.GET("/sync", productHandler.Sync)
			products.GET("/stock-warnings", productHandler.GetStockWarnings)
			products.POST("/sold-out", productHandler.BatchSoldOut)
			products.POST("/sold-out/restore", productHandler.BatchRestoreSoldOut)
			products.GET("/sold-out/records", productHandler.ListSoldOutRecords)
		}
		api.GET("/products/categories", productHandler.ListCategories)

		orders := api.Group("/orders")
		{
			orders.POST("", orderHandler.Create)
			orders.GET("", middleware.JWTAuth(), orderHandler.List)
			orders.GET("/:id", orderHandler.GetByID)
			orders.GET("/no/:orderNo", orderHandler.GetByOrderNo)
			orders.GET("/no/:orderNo/print", orderHandler.GetForPrint)
			orders.PUT("/:id/status", middleware.JWTAuth(), orderHandler.UpdateStatus)
			orders.POST("/:id/cancel", orderHandler.Cancel)
			orders.POST("/:id/refund", orderHandler.Refund)
			orders.POST("/payment", orderHandler.GetPaymentParams)
			orders.POST("/batch", middleware.JWTAuth(), orderHandler.BatchCreate)
			orders.GET("/sync/incremental", orderHandler.GetIncremental)
			orders.GET("/sync/store", orderHandler.GetForSync)
		}

		api.POST("/payment/wechat/notify", orderHandler.WechatNotify)
		api.POST("/payment/alipay/notify", orderHandler.AlipayNotify)

		queue := api.Group("/queue")
		{
			queue.POST("", orderHandler.CreateQueue)
			queue.POST("/:storeId/process", orderHandler.ProcessQueues)
		}

		sync := api.Group("/sync")
		{
			sync.GET("/products", productHandler.Sync)
		}

		stores := api.Group("/stores")
		stores.Use(middleware.JWTAuth())
		{
			stores.POST("", storeHandler.CreateStore)
			stores.GET("", storeHandler.ListStores)
			stores.GET("/:id", storeHandler.GetStore)
			stores.PUT("/:id", storeHandler.UpdateStore)
			stores.DELETE("/:id", storeHandler.DeleteStore)
		}

		printers := api.Group("/printers")
		printers.Use(middleware.JWTAuth())
		{
			printers.POST("", storeHandler.CreatePrinter)
			printers.GET("", storeHandler.ListPrinters)
			printers.GET("/:id", storeHandler.GetPrinter)
			printers.PUT("/:id", storeHandler.UpdatePrinter)
			printers.DELETE("/:id", storeHandler.DeletePrinter)
			printers.POST("/test", storeHandler.PrintTest)
		}

		internalPrinters := api.Group("/internal/printers")
		{
			internalPrinters.GET("", storeHandler.ListPrinters)
			internalPrinters.GET("/:id", storeHandler.GetPrinter)
		}

		members := api.Group("/members")
		members.Use(middleware.JWTAuth())
		{
			members.POST("", memberHandler.CreateMember)
			members.GET("", memberHandler.ListMembers)
			members.GET("/:id", memberHandler.GetMember)
			members.PUT("/:id", memberHandler.UpdateMember)
			members.DELETE("/:id", memberHandler.DeleteMember)
			members.POST("/:id/points", memberHandler.AdjustPoints)
			members.POST("/login", memberHandler.MemberLogin)
		}

		memberLevels := api.Group("/member-levels")
		memberLevels.Use(middleware.JWTAuth())
		{
			memberLevels.POST("", memberHandler.CreateMemberLevel)
			memberLevels.GET("", memberHandler.ListMemberLevels)
			memberLevels.GET("/:id", memberHandler.GetMemberLevel)
			memberLevels.PUT("/:id", memberHandler.UpdateMemberLevel)
			memberLevels.DELETE("/:id", memberHandler.DeleteMemberLevel)
		}

		pointsRecords := api.Group("/points-records")
		pointsRecords.Use(middleware.JWTAuth())
		{
			pointsRecords.GET("", memberHandler.ListPointsRecords)
		}

		coupons := api.Group("/coupons")
		coupons.Use(middleware.JWTAuth())
		{
			coupons.POST("", couponHandler.CreateCoupon)
			coupons.GET("", couponHandler.ListCoupons)
			coupons.GET("/:id", couponHandler.GetCoupon)
			coupons.PUT("/:id", couponHandler.UpdateCoupon)
			coupons.DELETE("/:id", couponHandler.DeleteCoupon)
			coupons.POST("/issue", couponHandler.IssueCoupon)
			coupons.POST("/verify", couponHandler.VerifyCoupon)
		}

		memberCoupons := api.Group("/member-coupons")
		memberCoupons.Use(middleware.JWTAuth())
		{
			memberCoupons.GET("", couponHandler.ListMemberCoupons)
			memberCoupons.GET("/:id", couponHandler.GetMemberCoupon)
		}

		promotions := api.Group("/promotions")
		promotions.Use(middleware.JWTAuth())
		{
			promotions.POST("", promotionHandler.CreatePromotion)
			promotions.GET("", promotionHandler.ListPromotions)
			promotions.GET("/:id", promotionHandler.GetPromotion)
			promotions.PUT("/:id", promotionHandler.UpdatePromotion)
			promotions.DELETE("/:id", promotionHandler.DeletePromotion)
			promotions.POST("/calculate", promotionHandler.CalculateBestCombination)
		}

		miniPromotions := api.Group("/mini/promotions")
		miniPromotions.Use(middleware.MemberAuth())
		{
			miniPromotions.GET("/active", promotionHandler.GetActivePromotions)
			miniPromotions.POST("/calculate", promotionHandler.CalculateBestCombination)
		}

		miniVoice := api.Group("/mini/voice")
		miniVoice.Use(middleware.MemberAuth())
		{
			miniVoice.POST("/parse", voiceOrderHandler.ParseVoiceText)
		}

		miniCoupons := api.Group("/mini/coupons")
		miniCoupons.Use(middleware.MemberAuth())
		{
			miniCoupons.GET("/claimable", promotionHandler.GetClaimableCoupons)
			miniCoupons.GET("/available", promotionHandler.GetAvailableCoupons)
			miniCoupons.GET("/my", promotionHandler.GetMyCoupons)
			miniCoupons.POST("/claim", promotionHandler.ClaimCoupon)
		}

		reports := api.Group("/reports")
		reports.Use(middleware.JWTAuth())
		{
			reports.GET("/overview", reportHandler.GetOverview)
			reports.GET("/daily", reportHandler.GetDailyReports)
			reports.POST("/daily/generate", reportHandler.GenerateDailyReport)
			reports.GET("/product-sales", reportHandler.GetProductSalesReport)
			reports.GET("/category-sales", reportHandler.GetCategorySalesReport)
			reports.GET("/hourly-sales", reportHandler.GetHourlySalesReport)
			reports.GET("/payment", reportHandler.GetPaymentReport)
			reports.POST("/export", reportHandler.ExportReport)
		}

		wechat := api.Group("/payment/wechat")
		{
			wechat.POST("/unified-order", middleware.JWTAuth(), paymentHandler.WechatUnifiedOrder)
			wechat.GET("/order/:orderNo", middleware.JWTAuth(), paymentHandler.WechatQueryOrder)
			wechat.POST("/refund", middleware.JWTAuth(), paymentHandler.WechatRefund)
			wechat.POST("/notify", paymentHandler.WechatNotify)
			wechat.POST("/refund/notify", paymentHandler.WechatRefundNotify)
		}

		tables := api.Group("/tables")
		{
			tables.POST("", middleware.JWTAuth(), tableHandler.CreateTable)
			tables.GET("", middleware.JWTAuth(), tableHandler.ListTables)
			tables.GET("/:id", middleware.JWTAuth(), tableHandler.GetTable)
			tables.PUT("/:id", middleware.JWTAuth(), tableHandler.UpdateTable)
			tables.DELETE("/:id", middleware.JWTAuth(), tableHandler.DeleteTable)
			tables.POST("/batch", middleware.JWTAuth(), tableHandler.BatchCreateTables)
			tables.POST("/:id/qrcode", middleware.JWTAuth(), tableHandler.GenerateQRCode)
			tables.GET("/occupied", middleware.JWTAuth(), tableHandler.GetOccupiedTables)
			tables.GET("/available", tableHandler.GetAvailableTables)
			tables.POST("/checkin", middleware.JWTAuth(), tableHandler.Checkin)
			tables.POST("/checkout", middleware.JWTAuth(), tableHandler.Checkout)
			tables.POST("/scan", tableHandler.ScanQRCode)
		}

		tableAreas := api.Group("/table-areas")
		tableAreas.Use(middleware.JWTAuth())
		{
			tableAreas.POST("", tableHandler.CreateArea)
			tableAreas.GET("", tableHandler.ListAreas)
			tableAreas.PUT("/:id", tableHandler.UpdateArea)
			tableAreas.DELETE("/:id", tableHandler.DeleteArea)
		}

		reservations := api.Group("/reservations")
		{
			reservations.POST("", tableHandler.CreateReservation)
			reservations.GET("", middleware.JWTAuth(), tableHandler.ListReservations)
			reservations.GET("/:id", middleware.JWTAuth(), tableHandler.GetReservation)
			reservations.PUT("/:id", middleware.JWTAuth(), tableHandler.UpdateReservation)
			reservations.POST("/:id/cancel", tableHandler.CancelReservation)
			reservations.POST("/:id/checkin", middleware.JWTAuth(), tableHandler.CheckinReservation)
			reservations.GET("/timeslots", tableHandler.GetTimeSlots)
		}

		queues := api.Group("/queues")
		{
			queues.POST("", tableHandler.CreateQueue)
			queues.GET("", middleware.JWTAuth(), tableHandler.ListQueues)
			queues.GET("/:id", middleware.JWTAuth(), tableHandler.GetQueue)
			queues.POST("/call", middleware.JWTAuth(), tableHandler.CallQueue)
			queues.POST("/call-next/:store_id", middleware.JWTAuth(), tableHandler.CallNextQueue)
			queues.POST("/cancel", tableHandler.CancelQueue)
			queues.POST("/arrive", middleware.JWTAuth(), tableHandler.ArriveQueue)
			queues.GET("/status", tableHandler.GetQueueStatus)
			queues.GET("/my", tableHandler.GetMyQueue)
			queues.GET("/waiting-count", tableHandler.GetWaitingCount)
		}

		queueConfigs := api.Group("/queue-config")
		queueConfigs.Use(middleware.JWTAuth())
		{
			queueConfigs.GET("", tableHandler.GetQueueConfig)
			queueConfigs.POST("", tableHandler.SaveQueueConfig)
		}

		storesMap := api.Group("/store-map")
		{
			storesMap.GET("", tableHandler.GetStoreMap)
		}

		pointsRules := api.Group("/points-rules")
		pointsRules.Use(middleware.JWTAuth())
		{
			pointsRules.POST("", pointsRuleHandler.CreateRule)
			pointsRules.GET("", pointsRuleHandler.ListRules)
			pointsRules.GET("/:id", pointsRuleHandler.GetRule)
			pointsRules.PUT("/:id", pointsRuleHandler.UpdateRule)
			pointsRules.DELETE("/:id", pointsRuleHandler.DeleteRule)
			pointsRules.POST("/calculate-earn", pointsRuleHandler.CalculateEarnedPoints)
			pointsRules.POST("/calculate-redeem", pointsRuleHandler.CalculateRedemptionDiscount)
		}

		rechargeActivities := api.Group("/recharge-activities")
		rechargeActivities.Use(middleware.JWTAuth())
		{
			rechargeActivities.POST("", rechargeActivityHandler.CreateActivity)
			rechargeActivities.GET("", rechargeActivityHandler.ListActivities)
			rechargeActivities.GET("/:id", rechargeActivityHandler.GetActivity)
			rechargeActivities.PUT("/:id", rechargeActivityHandler.UpdateActivity)
			rechargeActivities.DELETE("/:id", rechargeActivityHandler.DeleteActivity)
		}

		memberRecharges := api.Group("/member-recharges")
		memberRecharges.Use(middleware.JWTAuth())
		{
			memberRecharges.POST("", rechargeActivityHandler.ProcessRecharge)
			memberRecharges.GET("", rechargeActivityHandler.ListRecharges)
		}

		recommendations := api.Group("/recommendations")
		{
			recommendations.GET("/cart", recommendHandler.GetCartRecommendations)
			recommendations.GET("/config/meta", middleware.JWTAuth(), recommendHandler.GetConfigMeta)
			recommendations.GET("/config", middleware.JWTAuth(), recommendHandler.GetConfig)
			recommendations.PUT("/config", middleware.JWTAuth(), recommendHandler.UpdateConfig)
			recommendations.POST("/refresh", middleware.JWTAuth(), recommendHandler.TriggerRefresh)
			recommendations.GET("/refresh/status", middleware.JWTAuth(), recommendHandler.GetRefreshStatus)
		}

		stalls := api.Group("/stalls")
		stalls.Use(middleware.JWTAuth())
		{
			stalls.POST("", stallHandler.CreateStall)
			stalls.GET("", stallHandler.ListStalls)
			stalls.GET("/all", stallHandler.GetAllStalls)
			stalls.GET("/:id", stallHandler.GetStall)
			stalls.PUT("/:id", stallHandler.UpdateStall)
			stalls.DELETE("/:id", stallHandler.DeleteStall)
		}

		stallDevices := api.Group("/stall-devices")
		stallDevices.Use(middleware.JWTAuth())
		{
			stallDevices.POST("", stallHandler.RegisterDevice)
			stallDevices.GET("", stallHandler.ListDevices)
			stallDevices.GET("/:id", stallHandler.GetDevice)
			stallDevices.DELETE("/:id", stallHandler.DeleteDevice)
		}

		stallHeartbeat := api.Group("/stall/heartbeat")
		{
			stallHeartbeat.POST("", stallHandler.Heartbeat)
		}

		stallUsers := api.Group("/stall-users")
		stallUsers.Use(middleware.JWTAuth())
		{
			stallUsers.POST("", stallHandler.CreateStallUser)
			stallUsers.GET("", stallHandler.ListStallUsers)
			stallUsers.PUT("/:id", stallHandler.UpdateStallUser)
			stallUsers.DELETE("/:id", stallHandler.DeleteStallUser)
		}

		stallAuth := api.Group("/stall/auth")
		{
			stallAuth.POST("/login", stallHandler.StallLogin)
		}

		stallSettlements := api.Group("/stall-settlements")
		stallSettlements.Use(middleware.JWTAuth())
		{
			stallSettlements.POST("", stallHandler.CreateSettlement)
			stallSettlements.GET("", stallHandler.ListSettlements)
		}

		stallReports := api.Group("/stall-reports")
		stallReports.Use(middleware.JWTAuth())
		{
			stallReports.GET("/daily", stallHandler.GetDailyReport)
			stallReports.POST("/daily/generate", stallHandler.GenerateDailyReport)
		}

		stockChecks := api.Group("/stock-checks")
		stockChecks.Use(middleware.JWTAuth())
		{
			stockChecks.POST("", stockCheckHandler.Create)
			stockChecks.GET("", stockCheckHandler.List)
			stockChecks.GET("/:id", stockCheckHandler.GetByID)
			stockChecks.GET("/:id/items", stockCheckHandler.GetItems)
			stockChecks.POST("/:id/upload", stockCheckHandler.UploadItems)
			stockChecks.PUT("/items/:item_id", stockCheckHandler.UpdateItem)
			stockChecks.POST("/:id/complete", stockCheckHandler.Complete)
			stockChecks.GET("/:id/diff", stockCheckHandler.DiffReport)
		}

		stockWarnings := api.Group("/stock-warnings")
		{
			stockWarnings.GET("/dingtalk/test", stockWarningHandler.TestDingTalk)
		}

		queue2 := api.Group("/queue2")
		{
			queue2.POST("/take", queueHandler.TakeNumber)
			queue2.GET("/info", queueHandler.GetQueueInfo)
			queue2.GET("/all-waiting", middleware.JWTAuth(), queueHandler.GetAllWaiting)
			queue2.POST("/call", middleware.JWTAuth(), queueHandler.CallNumber)
			queue2.POST("/arrive", middleware.JWTAuth(), queueHandler.Arrive)
			queue2.POST("/cancel", queueHandler.Cancel)
			queue2.GET("/config", queueHandler.GetQueueConfig)
			queue2.POST("/preorder", queueHandler.SavePreOrder)
			queue2.GET("/preorder", queueHandler.GetPreOrder)
		}

		api.GET("/queue/ws", queueHandler.WebSocket)

		analytics := api.Group("/analytics")
		analytics.Use(middleware.JWTAuth())
		{
			analytics.GET("/revenue", analyticsHandler.GetRevenueReport)
			analytics.GET("/hourly-trend", analyticsHandler.GetHourlyTrend)
			analytics.GET("/top-products", analyticsHandler.GetTopProducts)
			analytics.POST("/sync/backfill", analyticsHandler.TriggerFullBackfill)
			analytics.GET("/sync/status", analyticsHandler.GetSyncStatus)
			analytics.POST("/cost/import", analyticsHandler.ImportCostExcel)
			analytics.GET("/cost/list", analyticsHandler.GetCostList)
			analytics.GET("/profit/report", analyticsHandler.GetProfitReport)
			analytics.GET("/profit/summary", analyticsHandler.GetProfitSummary)
			analytics.GET("/profit/report/v2", analyticsHandler.GetProfitReportV2)
			analytics.GET("/profit/summary/v2", analyticsHandler.GetProfitSummaryV2)
		}

		ingredients := api.Group("/ingredients")
		ingredients.Use(middleware.JWTAuth())
		{
			ingredients.GET("", ingredientHandler.GetIngredients)
			ingredients.GET("/categories", ingredientHandler.GetIngredientCategories)
			ingredients.GET("/:id", ingredientHandler.GetIngredient)
			ingredients.POST("", ingredientHandler.CreateIngredient)
			ingredients.PUT("/:id", ingredientHandler.UpdateIngredient)
			ingredients.DELETE("/:id", ingredientHandler.DeleteIngredient)
			ingredients.GET("/:id/price-history", ingredientHandler.GetPriceHistory)
		}

		bom := api.Group("/bom")
		bom.Use(middleware.JWTAuth())
		{
			bom.GET("/:product_id", ingredientHandler.GetProductBOM)
			bom.POST("/save", ingredientHandler.SaveProductBOM)
			bom.GET("/:product_id/cost-detail", ingredientHandler.GetProductCostDetail)
		}

		costAlerts := api.Group("/cost-alerts")
		costAlerts.Use(middleware.JWTAuth())
		{
			costAlerts.GET("", ingredientHandler.GetCostAlerts)
			costAlerts.POST("/handle", ingredientHandler.HandleAlert)
		}

		inventory := api.Group("/inventory")
		inventory.Use(middleware.JWTAuth())
		{
			inventory.POST("/sync", ingredientHandler.TriggerInventorySync)
		}

		delivery := api.Group("/delivery")
		{
			delivery.POST("", middleware.JWTAuth(), deliveryHandler.CreateDeliveryOrder)
			delivery.GET("", middleware.JWTAuth(), deliveryHandler.ListDeliveryOrders)
			delivery.GET("/:id", middleware.JWTAuth(), deliveryHandler.GetDeliveryOrder)
			delivery.PUT("/:id/status", middleware.JWTAuth(), deliveryHandler.UpdateDeliveryStatus)
			delivery.POST("/:id/assign-rider", middleware.JWTAuth(), deliveryHandler.AssignRider)
			delivery.POST("/:id/simulate-location", deliveryHandler.SimulateRiderLocation)
			delivery.GET("/order/:orderId", deliveryHandler.GetDeliveryOrderByOrderID)
			delivery.GET("/tracking/:orderId", deliveryHandler.GetDeliveryTracking)
		}

		riders := api.Group("/riders")
		riders.Use(middleware.JWTAuth())
		{
			riders.POST("/store/:storeId", deliveryHandler.CreateRider)
			riders.GET("/store/:storeId", deliveryHandler.ListRiders)
			riders.DELETE("/:id", deliveryHandler.DeleteRider)
			riders.POST("/location", deliveryHandler.UpdateRiderLocation)
			riders.GET("/:riderId/location", deliveryHandler.GetRiderLocation)
		}

		pickup := api.Group("/pickup")
		{
			pickup.POST("/code", deliveryHandler.GeneratePickupCode)
			pickup.POST("/verify", deliveryHandler.VerifyPickupCode)
			pickup.GET("/order/:orderId", deliveryHandler.GetPickupCodeByOrder)
		}

		amap := api.Group("/amap")
		amap.Use(middleware.JWTAuth())
		{
			amap.POST("/route", deliveryHandler.PlanRoute)
			amap.POST("/geocode", deliveryHandler.Geocode)
		}

		forecast := api.Group("/forecast")
		forecast.Use(middleware.JWTAuth())
		{
			forecast.GET("/store/:storeId", forecastHandler.GetForecast)
			forecast.GET("/store/:storeId/stocking", forecastHandler.GetStockingSuggestion)
			forecast.POST("/store/:storeId/purchase", forecastHandler.GeneratePurchaseOrder)
			forecast.POST("/store/:storeId/trigger", forecastHandler.TriggerForecastTask)
		}

		purchases := api.Group("/purchases")
		purchases.Use(middleware.JWTAuth())
		{
			purchases.GET("", forecastHandler.GetPurchaseList)
			purchases.GET("/:id", forecastHandler.GetPurchaseDetail)
			purchases.POST("", forecastHandler.CreatePurchaseOrder)
			purchases.PUT("/:id/status", forecastHandler.UpdatePurchaseStatus)
			purchases.POST("/:id/send", forecastHandler.SendPurchaseToSupplier)
			purchases.GET("/:id/export", forecastHandler.ExportPurchaseExcel)
		}

		waiter := api.Group("/waiter")
		waiter.Use(middleware.JWTAuth())
		{
			waiter.GET("/stats", waiterHandler.GetWaiterStats)
			waiter.GET("/tables", waiterHandler.GetTablesWithStatus)
			waiter.GET("/order-items/by-cook-status", waiterHandler.GetPendingCookItems)
			waiter.PUT("/order-items/cook-status", waiterHandler.UpdateItemCookStatus)
			waiter.POST("/order-items/serve", waiterHandler.MarkItemsServed)
			waiter.POST("/orders/:id/items", waiterHandler.AddOrderItems)
			waiter.GET("/calls", waiterHandler.ListCalls)
			waiter.POST("/calls/:id/handle", waiterHandler.HandleCall)
		}

		api.POST("/waiter/call", waiterHandler.CallWaiter)
		api.GET("/waiter/ws", waiterHandler.WebSocket)

		facePayment := api.Group("/face-payment")
		facePayment.Use(middleware.JWTAuth())
		{
			facePayment.POST("/init", facePaymentHandler.FacePaymentInit)
			facePayment.POST("/confirm", facePaymentHandler.FacePaymentConfirm)
			facePayment.GET("/:id/status", facePaymentHandler.FacePaymentQuery)
			facePayment.POST("/:id/cancel", facePaymentHandler.FacePaymentCancel)
		}

		voice := api.Group("/voice")
		voice.Use(middleware.JWTAuth())
		{
			voice.POST("/broadcast", facePaymentHandler.VoiceBroadcast)
		}

		timeSlotPricing := api.Group("/time-slot-pricing")
		timeSlotPricing.Use(middleware.JWTAuth())
		{
			timeSlotPricing.POST("", timeSlotPricingHandler.Create)
			timeSlotPricing.GET("", timeSlotPricingHandler.List)
			timeSlotPricing.GET("/:id", timeSlotPricingHandler.Get)
			timeSlotPricing.PUT("/:id", timeSlotPricingHandler.Update)
			timeSlotPricing.DELETE("/:id", timeSlotPricingHandler.Delete)
		}

		miniTimeSlotPricing := api.Group("/mini/time-slot-pricing")
		miniTimeSlotPricing.Use(middleware.MemberAuth())
		{
			miniTimeSlotPricing.GET("/active", timeSlotPricingHandler.GetActiveTimeSlots)
			miniTimeSlotPricing.POST("/calculate", timeSlotPricingHandler.CalculatePrice)
		}

		review := api.Group("/review")
		review.Use(middleware.JWTAuth())
		{
			review.POST("/auth", reviewHandler.SaveAuth)
			review.GET("/auth", reviewHandler.GetAuth)
			review.GET("/auths", reviewHandler.ListAuths)

			review.POST("/sync", reviewHandler.SyncReviews)
			review.POST("/sync-all", reviewHandler.SyncAll)

			review.GET("/ratings", reviewHandler.ListRatings)
			review.GET("/ratings/trend", reviewHandler.GetRatingTrend)

			review.GET("/reviews", reviewHandler.ListReviews)
			review.GET("/reviews/:id", reviewHandler.GetReview)
			review.POST("/reviews/:id/reply", reviewHandler.ReplyReview)

			review.POST("/work-orders", reviewHandler.CreateWorkOrder)
			review.GET("/work-orders", reviewHandler.ListWorkOrders)
			review.GET("/work-orders/:id", reviewHandler.GetWorkOrder)
			review.POST("/work-orders/:id/handle", reviewHandler.HandleWorkOrder)

			review.GET("/alerts", reviewHandler.ListAlerts)
			review.POST("/alerts/:id/handle", reviewHandler.HandleAlert)
			review.POST("/alerts/check", reviewHandler.CheckAlerts)
		}

		sms := api.Group("/sms")
		sms.Use(middleware.JWTAuth())
		{
			sms.POST("/templates", smsHandler.CreateTemplate)
			sms.PUT("/templates/:id", smsHandler.UpdateTemplate)
			sms.DELETE("/templates/:id", smsHandler.DeleteTemplate)
			sms.GET("/templates/:id", smsHandler.GetTemplate)
			sms.GET("/templates", smsHandler.ListTemplates)
			sms.POST("/templates/:id/review", smsHandler.ReviewTemplate)
			sms.GET("/templates/active/list", smsHandler.ListActiveTemplates)

			sms.POST("/tasks", smsHandler.CreateTask)
			sms.PUT("/tasks/:id", smsHandler.UpdateTask)
			sms.DELETE("/tasks/:id", smsHandler.DeleteTask)
			sms.GET("/tasks/:id", smsHandler.GetTask)
			sms.GET("/tasks", smsHandler.ListTasks)
			sms.POST("/tasks/:id/start", smsHandler.StartTask)
			sms.POST("/tasks/:id/pause", smsHandler.PauseTask)
			sms.GET("/tasks/:id/statistics", smsHandler.GetTaskStatistics)
			sms.POST("/tasks/target-count", smsHandler.CalculateTargetCount)

			sms.GET("/records", smsHandler.ListRecords)
			sms.GET("/records/:id", smsHandler.GetRecord)

			sms.POST("/test-send", smsHandler.SendTestSms)
		}

		transfers := api.Group("/transfers")
		transfers.Use(middleware.JWTAuth())
		{
			transfers.POST("", transferHandler.CreateTransfer)
			transfers.GET("", transferHandler.ListTransfers)
			transfers.GET("/:id", transferHandler.GetTransfer)
			transfers.GET("/:id/items", transferHandler.GetTransferItems)
			transfers.POST("/:id/confirm-outbound", transferHandler.ConfirmOutbound)
			transfers.POST("/:id/ship", transferHandler.StartShipping)
			transfers.POST("/:id/receive", transferHandler.ReceiveTransfer)
			transfers.POST("/:id/complete", transferHandler.CompleteTransfer)
			transfers.POST("/:id/cancel", transferHandler.CancelTransfer)
			transfers.GET("/:id/logistics", transferHandler.GetLogisticsTrack)
		}
	}

	return r
}
