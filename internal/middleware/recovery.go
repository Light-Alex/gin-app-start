package middleware

import (
	"fmt"
	"runtime/debug"

	"gin-app-start/internal/code"
	"gin-app-start/pkg/response"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func Recovery(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				logger.Error("HTTP Panic",
					zap.String("panic", fmt.Sprintf("%+v", err)),
					zap.String("path", c.Request.URL.Path),
					zap.String("method", c.Request.Method),
					zap.String("ip", c.ClientIP()),
					zap.String("stack", string(debug.Stack())),
				)

				response.Error(c, code.ServerError, code.Text(code.ServerError))
				c.Abort() // 终止当前请求的后续处理，防止 panic 后的代码继续执行导致更多问题
			}
		}()
		c.Next()
	}
}
