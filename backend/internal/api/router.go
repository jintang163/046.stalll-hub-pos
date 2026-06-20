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
		}

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
	}

	return r
}
