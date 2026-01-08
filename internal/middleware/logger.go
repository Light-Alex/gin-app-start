package middleware

import (
	"fmt"
	"gin-app-start/pkg/trace"
	"net/http"
	"net/url"
	"runtime/debug"
	"time"

	"gin-app-start/internal/code"
	"gin-app-start/internal/common"

	"github.com/gin-gonic/gin"
	"go.uber.org/multierr"
	"go.uber.org/zap"
)

func Logger(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Writer.Status() == http.StatusNotFound {
			return
		}

		start := time.Now()

		context := common.NewContext(c)
		defer common.ReleaseContext(context)

		context.Init()
		context.SetLogger(logger)

		if traceId := context.GetHeader(trace.Header); traceId != "" {
			context.SetTrace(trace.New(traceId))
		} else {
			context.SetTrace(trace.New(""))
		}

		defer func() {
			var (
				response        interface{}
				businessCode    int
				businessCodeMsg string
				abortErr        error
				// traceId         string
			)

			if ct := context.Trace(); ct != nil {
				context.SetHeader(trace.Header, ct.ID())
				// traceId = ct.ID()
			}

			// region 发生 panic 时，记录日志并返回服务器错误
			if err := recover(); err != nil {
				stackInfo := string(debug.Stack())
				logger.Error("got panic", zap.String("panic", fmt.Sprintf("%+v", err)), zap.String("stack", stackInfo))
				context.AbortWithError(common.Error(
					http.StatusInternalServerError,
					code.ServerError,
					code.Text(code.ServerError)),
				)

				// todo: 发送告警通知给相关人员
			}
			// endregion

			// region 发生错误时，记录日志并返回错误信息
			if c.IsAborted() {
				for i := range c.Errors {
					multierr.AppendInto(&abortErr, c.Errors[i])
				}

				if err := context.AbortError(); err != nil {

					// todo: 发送告警通知给相关人员

					multierr.AppendInto(&abortErr, err.StackError())
					businessCode = err.BusinessCode()
					businessCodeMsg = err.Message()
					response = &code.Failure{
						Code:    businessCode,
						Message: businessCodeMsg,
					}
					c.JSON(err.HTTPCode(), response)
				}
			}
			// endregion

			// region 正确返回
			response = context.GetPayload()
			if response != nil {
				c.JSON(http.StatusOK, response)
			}
			// endregion

			// region 记录日志
			var t *trace.Trace
			if x := context.Trace(); x != nil {
				t = x.(*trace.Trace)
			} else {
				return
			}

			// 获取访问路径
			decodedURL, _ := url.QueryUnescape(c.Request.URL.RequestURI())

			// ctx.Request.Header，精简 Header 参数
			traceHeader := map[string]string{
				"Content-Type": c.GetHeader("Content-Type"),
			}

			t.WithRequest(&trace.Request{
				TTL:        "un-limit",
				Method:     c.Request.Method,
				DecodedURL: decodedURL,
				Header:     traceHeader,
				Body:       string(context.RawData()),
			})

			var responseBody interface{}

			if response != nil {
				responseBody = response
			}

			t.WithResponse(&trace.Response{
				Header:          c.Writer.Header(),
				HttpCode:        c.Writer.Status(),
				HttpCodeMsg:     http.StatusText(c.Writer.Status()),
				BusinessCode:    businessCode,
				BusinessCodeMsg: businessCodeMsg,
				Body:            responseBody,
				CostSeconds:     time.Since(start).Seconds(),
			})

			t.Success = !c.IsAborted() && (c.Writer.Status() == http.StatusOK)
			t.CostSeconds = time.Since(start).Seconds()

			logger.Info("trace-log",
				zap.String("user_agent", c.Request.UserAgent()),
				zap.String("ip", c.ClientIP()),
				zap.Any("method", c.Request.Method),
				zap.Any("path", decodedURL),
				zap.Any("http_code", c.Writer.Status()),
				zap.Any("business_code", businessCode),
				zap.Any("success", t.Success),
				zap.Any("cost_seconds", t.CostSeconds),
				zap.Any("trace_id", t.Identifier),
				zap.Any("trace_info", t),
				zap.Error(abortErr),
			)
			// endregion
		}()

		c.Next()
	}
}
