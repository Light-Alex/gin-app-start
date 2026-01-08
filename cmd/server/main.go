package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "gin-app-start/docs"
	"gin-app-start/internal/common"
	"gin-app-start/internal/config"
	"gin-app-start/internal/controller"
	"gin-app-start/internal/model"
	"gin-app-start/internal/redis"
	"gin-app-start/internal/repository"
	"gin-app-start/internal/router"
	"gin-app-start/internal/service"
	"gin-app-start/pkg/database"
	"gin-app-start/pkg/logger"
	"gin-app-start/pkg/timeutil"

	"go.uber.org/zap"
)

//	@title			Gin App API
//	@version		1.0
//	@description	This is a RESTful API server built with Gin framework.
//	@termsOfService	http://swagger.io/terms/

//	@contact.name	API Support
//	@contact.url	http://www.swagger.io/support
//	@contact.email	support@swagger.io

//	@license.name	Apache 2.0
//	@license.url	http://www.apache.org/licenses/LICENSE-2.0.html

//	@host		localhost:9060
//	@BasePath	/

//	@schemes	http https

var Version string

func main() {
	log.Printf("Version: %s\n", Version)

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	accessLogger, err := logger.Init(
		cfg,

		// 禁用控制台输出
		logger.WithDisableConsole(),
		// 添加自定义字段 "domain"，格式为 "项目名[环境]"，例如：go-gin-api[fat]，便于区分不同环境和项目的日志
		logger.WithField("domain", fmt.Sprintf("%s[%s]", common.ProjectName, cfg.Server.Mode)),
		// 设置时间格式为 "2006-01-02 15:04:05"
		logger.WithTimeLayout(timeutil.CSTLayout),
		// 日志输出到文件 cfg.Log.FilePath
		logger.WithFileP(cfg.Log.FilePath),
	)

	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}

	defer accessLogger.Sync()

	accessLogger.Info("Application starting", zap.String("version", Version), zap.String("mode", cfg.Server.Mode))

	db, err := database.NewPostgresDB(&database.PostgresConfig{
		Host:         cfg.Database.Host,
		Port:         cfg.Database.Port,
		User:         cfg.Database.User,
		Password:     cfg.Database.Password,
		DBName:       cfg.Database.DBName,
		SSLMode:      cfg.Database.SSLMode,
		MaxIdleConns: cfg.Database.MaxIdleConns,
		MaxOpenConns: cfg.Database.MaxOpenConns,
		MaxLifetime:  cfg.Database.MaxLifetime,
		LogLevel:     cfg.Database.LogLevel,
	})
	if err != nil {
		accessLogger.Fatal("Failed to initialize database", zap.Error(err))
	}
	defer database.DBRepo.DbClose()

	accessLogger.Info("Database connected successfully")

	if cfg.Database.AutoMigrate {
		if err := db.AutoMigrate(&model.User{}, &model.Order{}); err != nil {
			accessLogger.Fatal("Database migration failed", zap.Error(err))
		}
		accessLogger.Info("Database migration completed")
	}

	redisClient, err := database.NewRedisClient(&database.RedisConfig{
		Addr:         cfg.Redis.Addr,
		Password:     cfg.Redis.Password,
		DB:           cfg.Redis.DB,
		PoolSize:     cfg.Redis.PoolSize,
		MinIdleConns: cfg.Redis.MinIdleConns,
		MaxRetries:   cfg.Redis.MaxRetries,
	})
	if err != nil {
		accessLogger.Warn("Failed to initialize Redis", zap.Error(err))
	} else {
		defer redisClient.Close()
		accessLogger.Info("Redis connected successfully")
	}

	userRepo := repository.NewUserRepository(db)
	userService := service.NewUserService(userRepo)
	userController := controller.NewUserController(userService)
	healthController := controller.NewHealthController()

	redisRepo := redis.NewRedisRepository(redisClient, context.Background())
	orderRepo := repository.NewOrderRepository(db)
	orderService := service.NewOrderService(orderRepo, redisRepo)
	orderController := controller.NewOrderController(orderService)

	s, err := router.SetupRouter(accessLogger, healthController, userController, orderController, cfg)
	if err != nil {
		accessLogger.Fatal("Failed to initialize router", zap.Error(err))
	}

	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Server.Port),
		Handler:      s.Mux,
		ReadTimeout:  time.Duration(cfg.Server.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(cfg.Server.WriteTimeout) * time.Second,
	}

	go func() {
		appURL := fmt.Sprintf("http://localhost:%d", cfg.Server.Port)
		swaggerURL := fmt.Sprintf("http://localhost:%d/swagger/index.html", cfg.Server.Port)

		accessLogger.Info("Server started", zap.String("url", appURL))
		accessLogger.Info("Swagger documentation", zap.String("url", swaggerURL))

		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			accessLogger.Fatal("Server failed to start", zap.Error(err))
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	accessLogger.Info("Server shutting down...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		accessLogger.Error("Server shutdown failed", zap.Error(err))
	}

	accessLogger.Info("Server stopped")
}
