package middleware

import (
	"gin-app-start/pkg/logger"
	"gin-app-start/pkg/response"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				logger.Error("HTTP Panic",
					zap.Any("error", err),
					zap.String("path", c.Request.URL.Path),
					zap.String("method", c.Request.Method),
					zap.String("ip", c.ClientIP()),
				)

				response.Error(c, 50000, "Internal server error")
				c.Abort() // 终止当前请求的后续处理，防止 panic 后的代码继续执行导致更多问题
			}
		}()
		c.Next()
	}
}
