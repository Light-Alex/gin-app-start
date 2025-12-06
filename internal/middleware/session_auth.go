package middleware

import (
	"gin-app-start/internal/common"
	"gin-app-start/pkg/errors"
	"gin-app-start/pkg/logger"
	"gin-app-start/pkg/response"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

func SessionAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从服务端中获取session
		session := sessions.Default(c)
		sessionData := session.Get(common.SESSION_KEY)
		if sessionData == nil {
			logger.Error("Session not found")
			response.Error(c, errors.ErrUnauthorized.Code, errors.ErrUnauthorized.Message)
			c.Abort()
			return
		}
		c.Set(common.SESSION_KEY, sessionData)
		c.Next()
	}
}
