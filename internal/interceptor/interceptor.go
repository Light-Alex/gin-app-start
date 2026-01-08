package interceptor

import (
	"gin-app-start/internal/common"

	"go.uber.org/zap"
)

var _ Interceptor = (*interceptor)(nil)

type Interceptor interface {
	// SessionAuth 验证用户会话是否有效
	SessionAuth() common.HandlerFunc

	// i 为了避免被其他包实现
	i()
}

type interceptor struct {
	logger *zap.Logger
}

func New(logger *zap.Logger) Interceptor {
	return &interceptor{
		logger: logger,
	}
}

func (i *interceptor) i() {}
