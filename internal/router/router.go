package router

import (
	"errors"
	"fmt"
	"net/http"

	"gin-app-start/internal/common"
	"gin-app-start/internal/config"
	"gin-app-start/internal/controller"
	"gin-app-start/internal/interceptor"
	"gin-app-start/internal/middleware"
	"gin-app-start/pkg/color"
	"gin-app-start/pkg/response"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-contrib/sessions/redis"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"go.uber.org/zap"
)

const _UI = `
 ██████╗ ██╗███╗   ██╗       █████╗ ██████╗ ██████╗       ███████╗████████╗ █████╗ ██████╗ ████████╗
██╔════╝ ██║████╗  ██║      ██╔══██╗██╔══██╗██╔══██╗      ██╔════╝╚══██╔══╝██╔══██╗██╔══██╗╚══██╔══╝
██║  ███╗██║██╔██╗ ██║█████╗███████║██████╔╝██████╔╝█████╗███████╗   ██║   ███████║██████╔╝   ██║   
██║   ██║██║██║╚██╗██║╚════╝██╔══██║██╔═══╝ ██╔═══╝ ╚════╝╚════██║   ██║   ██╔══██║██╔══██╗   ██║   
╚██████╔╝██║██║ ╚████║      ██║  ██║██║     ██║           ███████║   ██║   ██║  ██║██║  ██║   ██║   
 ╚═════╝ ╚═╝╚═╝  ╚═══╝      ╚═╝  ╚═╝╚═╝     ╚═╝           ╚══════╝   ╚═╝   ╚═╝  ╚═╝╚═╝  ╚═╝   ╚═╝   
                                                                                                    
`

// DisableTraceLog 禁止记录日志
func DisableTraceLog(ctx common.Context) {
	ctx.DisableTrace()
}

// RouterGroup 包装gin的RouterGroup
type RouterGroup interface {
	Group(string, ...common.HandlerFunc) RouterGroup
	IRoutes
}

var _ IRoutes = (*router)(nil)

// IRoutes 包装gin的IRoutes
type IRoutes interface {
	Any(string, ...common.HandlerFunc)
	GET(string, ...common.HandlerFunc)
	POST(string, ...common.HandlerFunc)
	DELETE(string, ...common.HandlerFunc)
	PATCH(string, ...common.HandlerFunc)
	PUT(string, ...common.HandlerFunc)
	OPTIONS(string, ...common.HandlerFunc)
	HEAD(string, ...common.HandlerFunc)
}

type router struct {
	group *gin.RouterGroup
}

func (r *router) Group(relativePath string, handlers ...common.HandlerFunc) RouterGroup {
	group := r.group.Group(relativePath, wrapHandlers(handlers...)...)
	return &router{group: group}
}

func (r *router) Any(relativePath string, handlers ...common.HandlerFunc) {
	r.group.Any(relativePath, wrapHandlers(handlers...)...)
}

func (r *router) GET(relativePath string, handlers ...common.HandlerFunc) {
	r.group.GET(relativePath, wrapHandlers(handlers...)...)
}

func (r *router) POST(relativePath string, handlers ...common.HandlerFunc) {
	r.group.POST(relativePath, wrapHandlers(handlers...)...)
}

func (r *router) DELETE(relativePath string, handlers ...common.HandlerFunc) {
	r.group.DELETE(relativePath, wrapHandlers(handlers...)...)
}

func (r *router) PATCH(relativePath string, handlers ...common.HandlerFunc) {
	r.group.PATCH(relativePath, wrapHandlers(handlers...)...)
}

func (r *router) PUT(relativePath string, handlers ...common.HandlerFunc) {
	r.group.PUT(relativePath, wrapHandlers(handlers...)...)
}

func (r *router) OPTIONS(relativePath string, handlers ...common.HandlerFunc) {
	r.group.OPTIONS(relativePath, wrapHandlers(handlers...)...)
}

func (r *router) HEAD(relativePath string, handlers ...common.HandlerFunc) {
	r.group.HEAD(relativePath, wrapHandlers(handlers...)...)
}

// wrapHandlers 包装 gin.HandlerFunc
func wrapHandlers(handlers ...common.HandlerFunc) []gin.HandlerFunc {
	funcs := make([]gin.HandlerFunc, len(handlers))
	for i, handler := range handlers {
		handler := handler
		funcs[i] = func(c *gin.Context) {
			ctx := common.NewContext(c)
			defer common.ReleaseContext(ctx)

			handler(ctx)
		}
	}

	return funcs
}

var _ Mux = (*mux)(nil)

// Mux http mux
type Mux interface {
	http.Handler                                                           // 实现 http.Handler 接口
	Group(relativePath string, handlers ...common.HandlerFunc) RouterGroup // 创建路由组
}

type mux struct {
	engine *gin.Engine
}

func (m *mux) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	m.engine.ServeHTTP(w, req)
}

func (m *mux) Group(relativePath string, handlers ...common.HandlerFunc) RouterGroup {
	return &router{
		group: m.engine.Group(relativePath, wrapHandlers(handlers...)...),
	}
}

type resource struct {
	mux          Mux                     // HTTP 路由
	logger       *zap.Logger             // HTTP 路由日志
	interceptors interceptor.Interceptor // HTTP 路由拦截器
}

type Server struct {
	Mux Mux
}

func SetupRouter(
	logger *zap.Logger,
	healthCtrl *controller.HealthController,
	userCtrl *controller.UserController,
	orderCtrl *controller.OrderController,
	cfg *config.Config,
) (*Server, error) {
	if logger == nil {
		return nil, errors.New("logger required")
	}

	r := new(resource)
	r.logger = logger

	gin.SetMode(cfg.Server.Mode)
	mux := &mux{
		engine: gin.New(),
	}

	fmt.Println(color.Blue(_UI))

	// 设置最大文件上传大小
	mux.engine.MaxMultipartMemory = int64(cfg.File.MaxSize)
	// 404处理
	mux.engine.NoRoute(func(c *gin.Context) {
		path := c.Request.URL.Path
		method := c.Request.Method
		response.Error(c, http.StatusNotFound, fmt.Sprintf("%s %s not found", method, path))
	})

	mux.engine.Use(middleware.CORS())
	mux.engine.Use(middleware.Recovery(logger))
	mux.engine.Use(middleware.Logger(logger))

	if cfg.Server.LimitNum > 0 {
		mux.engine.Use(middleware.RateLimit(cfg.Server.LimitNum))
	}

	// sessions.Store: 会话存储接口，用于存储会话数据
	var store sessions.Store
	if cfg.Session.UseRedis {
		store, _ = redis.NewStore(cfg.Session.Size, "tcp", cfg.Redis.Addr, "", cfg.Redis.Password, []byte(cfg.Session.Key))
	} else {
		store = cookie.NewStore([]byte(cfg.Session.Key))
	}

	// 设置session的选项
	// Path: session的路径作用域
	// HttpOnly: session是否只能通过HTTP(S)协议访问，不能通过JavaScript等客户端脚本访问
	// MaxAge: session的过期时间，单位为秒
	store.Options(sessions.Options{
		Path:     cfg.Session.Path,
		MaxAge:   cfg.Session.MaxAge,
		HttpOnly: cfg.Session.HttpOnly,
	})

	// sessions.Sessions功能：创建Session对象并关联到当前请求
	mux.engine.Use(sessions.Sessions(cfg.Session.Name, store))

	// Swagger documentation
	// 注册 Swagger 路由
	mux.engine.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	r.mux = mux
	r.interceptors = interceptor.New(logger)

	root := mux.Group("")
	{
		root.GET("/health", healthCtrl.HealthCheck())
	}

	apiV1 := mux.Group("/api/v1")
	{
		users := apiV1.Group("/users")
		{
			users.POST("", userCtrl.CreateUser())
			users.POST("/login", userCtrl.Login())
		}

		authUsers := apiV1.Group("/users", r.interceptors.SessionAuth())
		{
			authUsers.GET("/:id", userCtrl.GetUser())
			authUsers.PUT("/:id", userCtrl.UpdateUser())
			authUsers.POST("/change_pwd", userCtrl.ChangePassword())
			authUsers.POST("/upload_avatar", userCtrl.UploadImage())
			authUsers.GET("/file", userCtrl.GetImage())
			authUsers.DELETE("/:id", userCtrl.DeleteUser())
			authUsers.GET("", userCtrl.ListUsers())
			authUsers.POST("/logout", userCtrl.Logout())
		}

		orders := apiV1.Group("/orders", r.interceptors.SessionAuth())
		{
			orders.POST("", orderCtrl.CreateOrder())
			orders.GET("/search", orderCtrl.GetOrderByOrderNumber())
			orders.PUT("", orderCtrl.UpdateOrderByOrderNumber())
			orders.DELETE("", orderCtrl.DeleteOrderByOrderNumber())
			orders.GET("", orderCtrl.ListOrders())
		}
	}

	s := new(Server)
	s.Mux = r.mux

	return s, nil
}
