package router

import (
	"fmt"

	"gin-app-start/internal/config"
	"gin-app-start/internal/controller"
	"gin-app-start/internal/middleware"
	"gin-app-start/pkg/response"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-contrib/sessions/redis"
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
	// 设置最大文件上传大小
	router.MaxMultipartMemory = int64(cfg.File.MaxSize)
	// 404处理
	router.NoRoute(func(c *gin.Context) {
		path := c.Request.URL.Path
		method := c.Request.Method
		response.Error(c, 404, fmt.Sprintf("%s %s not found", method, path))
	})

	router.Use(middleware.Recovery())
	router.Use(middleware.Logger())
	router.Use(middleware.CORS())

	if cfg.Server.LimitNum > 0 {
		router.Use(middleware.RateLimit(cfg.Server.LimitNum))
	}

	// sessions.Store: 会话存储接口，用于存储会话数据
	var store sessions.Store
	if cfg.Session.UseRedis {
		store, _ = redis.NewStore(cfg.Session.Size, "tcp", cfg.Redis.Addr, "", cfg.Redis.Password, []byte(cfg.Session.Key))
	} else {
		store = cookie.NewStore([]byte(cfg.Session.Key))
	}

	store.Options(sessions.Options{
		Path:     cfg.Session.Path,
		MaxAge:   cfg.Session.MaxAge,
		HttpOnly: cfg.Session.HttpOnly,
	})

	// sessions.Sessions功能：创建Session对象并关联到当前请求
	router.Use(sessions.Sessions(cfg.Session.Name, store))

	router.GET("/health", healthCtrl.HealthCheck)

	// Swagger documentation
	// 注册 Swagger 路由
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	apiV1 := router.Group("/api/v1")
	{
		users := apiV1.Group("/users")
		{
			users.POST("", userCtrl.CreateUser)
			users.POST("/login", userCtrl.Login)
		}

		authUsers := apiV1.Group("/users").Use(middleware.SessionAuth())
		{
			authUsers.GET("/:id", userCtrl.GetUser)
			authUsers.PUT("/:id", userCtrl.UpdateUser)
			authUsers.POST("/change_pwd", userCtrl.ChangePassword)
			authUsers.POST("/upload_avatar", userCtrl.UploadImage)
			authUsers.GET("/file", userCtrl.GetImage)
			authUsers.DELETE("/:id", userCtrl.DeleteUser)
			authUsers.GET("", userCtrl.ListUsers)
			authUsers.POST("/logout", userCtrl.Logout)
		}

		orders := apiV1.Group("/orders").Use(middleware.SessionAuth())
		{
			orders.POST("", orderCtrl.CreateOrder)
			orders.GET("/search", orderCtrl.GetOrderByOrderNumber)
			orders.PUT("", orderCtrl.UpdateOrderByOrderNumber)
			orders.DELETE("", orderCtrl.DeleteOrderByOrderNumber)
			orders.GET("", orderCtrl.ListOrders)
		}
	}

	return router
}
