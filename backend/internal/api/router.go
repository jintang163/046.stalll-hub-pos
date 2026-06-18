package api

import (
	"stalll-hub-pos/backend/internal/api/handler"
	"stalll-hub-pos/backend/internal/middleware"
	"stalll-hub-pos/backend/internal/repository"
	"stalll-hub-pos/backend/internal/service"
	"stalll-hub-pos/backend/config"
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

	orderRepo := repository.NewOrderRepository(db)
	productRepo := repository.NewProductRepository(db)
	orderService := service.NewOrderService(orderRepo, productRepo, nsqProducer, config.AppConfig)
	orderHandler := handler.NewOrderHandler(orderService)

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
	}

	return r
}
