package router

import (
	"gin-app-start/internal/config"
	"gin-app-start/internal/controller"
	"gin-app-start/internal/middleware"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func SetupRouter(
	healthCtrl *controller.HealthController,
	userCtrl *controller.UserController,
	orderCtrl *controller.OrderController,
	cfg *config.Config,
) *gin.Engine {
	gin.SetMode(cfg.Server.Mode)

	router := gin.New()

	router.Use(middleware.Recovery())
	router.Use(middleware.Logger())
	router.Use(middleware.CORS())

	if cfg.Server.LimitNum > 0 {
		router.Use(middleware.RateLimit(cfg.Server.LimitNum))
	}

	router.GET("/health", healthCtrl.HealthCheck)

	// Swagger documentation
	// 注册 Swagger 路由
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	apiV1 := router.Group("/api/v1")
	{
		users := apiV1.Group("/users")
		{
			users.POST("", userCtrl.CreateUser)
			users.GET("/:id", userCtrl.GetUser)
			users.PUT("/:id", userCtrl.UpdateUser)
			users.DELETE("/:id", userCtrl.DeleteUser)
			users.GET("", userCtrl.ListUsers)
		}

		orders := apiV1.Group("/orders")
		{
			orders.POST("", orderCtrl.CreateOrder)
			orders.GET("/:order_number", orderCtrl.GetOrderByOrderNumber)
			orders.PUT("/:order_number", orderCtrl.UpdateOrderByOrderNumber)
			orders.DELETE("/:order_number", orderCtrl.DeleteOrderByOrderNumber)
			orders.GET("", orderCtrl.ListOrders)
		}
	}

	return router
}
