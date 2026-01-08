package interceptor

import (
	"net/http"

	"gin-app-start/internal/code"
	"gin-app-start/internal/common"
	"gin-app-start/pkg/errors"
)

func (i *interceptor) SessionAuth() common.HandlerFunc {
	return func(c common.Context) {
		// 从服务端中获取session
		session := c.GetSession()
		sessionData := session.Get(common.SESSION_KEY)
		if sessionData == nil {
			c.AbortWithError(common.Error(
				http.StatusUnauthorized,
				code.AuthorizationError,
				code.Text(code.AuthorizationError)).WithError(errors.New("server session not found")),
			)
			return
		}
		c.SetSessionUserInfo(sessionData)
	}
}
