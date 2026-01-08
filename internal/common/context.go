package common

import (
	"bytes"
	stdctx "context"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/url"
	"sync"

	"gin-app-start/pkg/trace"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"go.uber.org/zap"
)

type HandlerFunc func(c Context)

type Trace = trace.T

const (
	_Alias            = "_alias_"
	_TraceName        = "_trace_"
	_LoggerName       = "_logger_"
	_BodyName         = "_body_"
	_PayloadName      = "_payload_"
	_GraphPayloadName = "_graph_payload_"
	_SessionUserInfo  = "_session_user_info"
	_AbortErrorName   = "_abort_error_"
	_IsRecordMetrics  = "_is_record_metrics_"
)

type context struct {
	ctx *gin.Context
}

type StdContext struct {
	stdctx.Context
	Trace
	*zap.Logger
}

var contextPool = &sync.Pool{
	New: func() interface{} {
		return new(context)
	},
}

func NewContext(ctx *gin.Context) Context {
	context := contextPool.Get().(*context)
	context.ctx = ctx
	return context
}

func ReleaseContext(ctx Context) {
	c := ctx.(*context)
	c.ctx = nil
	contextPool.Put(c)
}

var _ Context = (*context)(nil)

type Context interface {
	Init()

	// GetGinContext 获取gin.Context对象
	GetGinContext() *gin.Context

	// Query 获取key对应的query参数值
	Query(key string) string

	// DefaultQuery 获取key对应的query参数值, 若不存在则返回 defaultValue
	DefaultQuery(key, defaultValue string) string

	// Param 获取key对应的path参数值
	Param(key string) string

	// PostForm 获取key对应的postform参数值
	PostForm(key string) string

	// FormFile 获取key对应的formfile参数值
	FormFile(key string) (*multipart.FileHeader, error)

	// File 响应文件
	File(filepath string)

	// ShouldBindQuery 反序列化 querystring
	// tag: `form:"xxx"` (注：不要写成 query)
	ShouldBindQuery(obj interface{}) error

	// ShouldBindPostForm 反序列化 postform (querystring会被忽略)
	// tag: `form:"xxx"`
	ShouldBindPostForm(obj interface{}) error

	// ShouldBindForm 同时反序列化 querystring 和 postform;
	// 当 querystring 和 postform 存在相同字段时，postform 优先使用。
	// tag: `form:"xxx"`
	ShouldBindForm(obj interface{}) error

	// ShouldBindJSON 反序列化 postjson
	// tag: `json:"xxx"`
	ShouldBindJSON(obj interface{}) error

	// ShouldBindURI 反序列化 path 参数(如路由路径为 /user/:name)
	// tag: `uri:"xxx"`
	ShouldBindURI(obj interface{}) error

	// Trace 获取 Trace 对象
	Trace() Trace
	SetTrace(trace Trace)
	DisableTrace()

	// Logger 获取 Logger 对象
	Logger() *zap.Logger
	SetLogger(logger *zap.Logger)

	// Payload 正确返回
	Payload(payload interface{})
	GetPayload() interface{}

	// AbortWithError 错误返回
	AbortWithError(err BusinessError)
	AbortError() BusinessError

	// Header 获取 Header 对象
	Header() http.Header
	// GetHeader 获取 Header
	GetHeader(key string) string
	// SetHeader 设置 Header
	SetHeader(key, value string)

	// GetSession 获取 Session 对象
	GetSession() sessions.Session

	// SessionUserInfo 当前用户信息
	SessionUserInfo() interface{}
	SetSessionUserInfo(value interface{})

	// Request 获取 Request 对象
	Request() *http.Request
	// RawData 获取 Request.Body
	RawData() []byte
	// Method 获取 Request.Method
	Method() string
	// RequestContext 获取请求的 context (当 client 关闭后，会自动 canceled)
	Host() string
	// Path 获取 请求的路径 Request.URL.Path (不附带 querystring)
	Path() string
	// URI 获取 unescape 后的 Request.URL.RequestURI()
	URI() string
	RequestContext() StdContext

	// ResponseWriter 获取 ResponseWriter 对象
	ResponseWriter() gin.ResponseWriter
}

func (c *context) Init() {
	// 从Gin上下文中读取HTTP请求的原始字节数据
	body, err := c.ctx.GetRawData()
	if err != nil {
		panic(err)
	}

	// 将请求体数据存储在Gin上下文中供后续使用
	c.ctx.Set(_BodyName, body)

	// GetRawData() 消耗了原始请求体，需要重新构造
	c.ctx.Request.Body = ioutil.NopCloser(bytes.NewBuffer(body)) // re-construct req body
}

// GetGinContext 获取gin.Context对象
func (c *context) GetGinContext() *gin.Context {
	return c.ctx
}

// Query 获取key对应的query参数值
func (c *context) Query(key string) string {
	return c.ctx.Query(key)
}

// DefaultQuery 获取key对应的query参数值, 若不存在则返回 defaultValue
func (c *context) DefaultQuery(key, defaultValue string) string {
	return c.ctx.DefaultQuery(key, defaultValue)
}

// Param 获取key对应的path参数值
func (c *context) Param(key string) string {
	return c.ctx.Param(key)
}

// PostForm 获取key对应的postform参数值
func (c *context) PostForm(key string) string {
	return c.ctx.PostForm(key)
}

// FormFile 获取key对应的formfile参数值
func (c *context) FormFile(key string) (*multipart.FileHeader, error) {
	file, err := c.ctx.FormFile(key)
	return file, err
}

// File 响应文件
func (c *context) File(filepath string) {
	c.ctx.File(filepath)
}

// ShouldBindQuery 反序列化querystring
// tag: `form:"xxx"` (注：不要写成query)
func (c *context) ShouldBindQuery(obj interface{}) error {
	return c.ctx.ShouldBindWith(obj, binding.Query)
}

// ShouldBindPostForm 反序列化 postform (querystring 会被忽略)
// tag: `form:"xxx"`
func (c *context) ShouldBindPostForm(obj interface{}) error {
	return c.ctx.ShouldBindWith(obj, binding.FormPost)
}

// ShouldBindForm 同时反序列化querystring和postform;
// 当querystring和postform存在相同字段时，postform优先使用。
// tag: `form:"xxx"`
func (c *context) ShouldBindForm(obj interface{}) error {
	return c.ctx.ShouldBindWith(obj, binding.Form)
}

// ShouldBindJSON 反序列化postjson
// tag: `json:"xxx"`
func (c *context) ShouldBindJSON(obj interface{}) error {
	return c.ctx.ShouldBindWith(obj, binding.JSON)
}

// ShouldBindURI 反序列化path参数(如路由路径为 /user/:name)
// tag: `uri:"xxx"`
func (c *context) ShouldBindURI(obj interface{}) error {
	return c.ctx.ShouldBindUri(obj)
}

func (c *context) Trace() Trace {
	t, ok := c.ctx.Get(_TraceName)
	if !ok || t == nil {
		return nil
	}

	return t.(Trace)
}

func (c *context) SetTrace(trace Trace) {
	c.ctx.Set(_TraceName, trace)
}

func (c *context) DisableTrace() {
	c.SetTrace(nil)
}

func (c *context) Logger() *zap.Logger {
	logger, ok := c.ctx.Get(_LoggerName)
	if !ok {
		return nil
	}

	return logger.(*zap.Logger)
}

func (c *context) SetLogger(logger *zap.Logger) {
	c.ctx.Set(_LoggerName, logger)
}

func (c *context) GetPayload() interface{} {
	if payload, ok := c.ctx.Get(_PayloadName); ok != false {
		return payload
	}
	return nil
}

func (c *context) Payload(payload interface{}) {
	c.ctx.Set(_PayloadName, payload)
}

func (c *context) Header() http.Header {
	header := c.ctx.Request.Header

	clone := make(http.Header, len(header))
	for k, v := range header {
		value := make([]string, len(v))
		copy(value, v)

		clone[k] = value
	}
	return clone
}

func (c *context) AbortWithError(err BusinessError) {
	if err != nil {
		httpCode := err.HTTPCode()
		if httpCode == 0 {
			httpCode = http.StatusInternalServerError
		}

		c.ctx.AbortWithStatus(httpCode)
		c.ctx.Set(_AbortErrorName, err)
	}
}

func (c *context) AbortError() BusinessError {
	err, _ := c.ctx.Get(_AbortErrorName)
	return err.(BusinessError)
}

func (c *context) GetSession() sessions.Session {
	return sessions.Default(c.ctx)
}

func (c *context) GetHeader(key string) string {
	return c.ctx.GetHeader(key)
}

func (c *context) SetHeader(key, value string) {
	c.ctx.Header(key, value)
}

func (c *context) SessionUserInfo() interface{} {
	val, ok := c.ctx.Get(_SessionUserInfo)
	if !ok {
		return nil
	}

	return val
}

func (c *context) SetSessionUserInfo(value interface{}) {
	c.ctx.Set(_SessionUserInfo, value)
}

// Request 获取 Request
func (c *context) Request() *http.Request {
	return c.ctx.Request
}

func (c *context) RawData() []byte {
	body, ok := c.ctx.Get(_BodyName)
	if !ok {
		return nil
	}

	return body.([]byte)
}

// Method 请求的method
func (c *context) Method() string {
	return c.ctx.Request.Method
}

// Host 请求的host
func (c *context) Host() string {
	return c.ctx.Request.Host
}

// Path 请求的路径(不附带querystring)
func (c *context) Path() string {
	return c.ctx.Request.URL.Path
}

// URI unescape后的uri
func (c *context) URI() string {
	uri, _ := url.QueryUnescape(c.ctx.Request.URL.RequestURI())
	return uri
}

// RequestContext (包装 Trace + Logger) 获取请求的 context (当client关闭后，会自动canceled)
func (c *context) RequestContext() StdContext {
	return StdContext{
		//c.ctx.Request.Context(),
		stdctx.Background(),
		c.Trace(),
		c.Logger(),
	}
}

// ResponseWriter 获取 ResponseWriter
func (c *context) ResponseWriter() gin.ResponseWriter {
	return c.ctx.Writer
}
