package errors

import "fmt"

type BusinessError struct {
	Code    int
	Message string
	Cause   error // 导致错误的原始错误
}

func (e *BusinessError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("code: %d, message: %s, cause: %v", e.Code, e.Message, e.Cause)
	}
	return fmt.Sprintf("code: %d, message: %s", e.Code, e.Message)
}

func NewBusinessError(code int, message string) *BusinessError {
	return &BusinessError{
		Code:    code,
		Message: message,
	}
}

func WrapBusinessError(code int, message string, cause error) *BusinessError {
	return &BusinessError{
		Code:    code,
		Message: message,
		Cause:   cause,
	}
}

var (
	ErrInvalidParams              = NewBusinessError(10001, "Invalid parameters")
	ErrUserNotFound               = NewBusinessError(10002, "User not found")
	ErrUnauthorized               = NewBusinessError(10003, "Unauthorized access")
	ErrUserExists                 = NewBusinessError(10004, "User already exists")
	ErrDatabaseError              = NewBusinessError(10005, "Database error")
	ErrOrderFailed                = NewBusinessError(10020, "Failed to query order")
	ErrOrderExists                = NewBusinessError(10021, "Order already exists")
	ErrOrderCreateFailed          = NewBusinessError(10022, "Failed to create order")
	ErrOrderNotFound              = NewBusinessError(10023, "Order not found")
	ErrOrderUpdateFailed          = NewBusinessError(10024, "Failed to update order")
	ErrOrderDeleteFailed          = NewBusinessError(10025, "Failed to delete order")
	ErrOrderListFailed            = NewBusinessError(10026, "Failed to list orders")
	ErrOrderMarshalFailed         = NewBusinessError(10027, "Failed to marshal order")
	ErrOrderCacheFailed           = NewBusinessError(10028, "Failed to cache order")
	ErrEmptyCache                 = NewBusinessError(10029, "Set empty cache")
	ErrOrderCacheDeleteFailed     = NewBusinessError(10030, "Failed to delete order cache")
	ErrOrderCacheParseTotalFailed = NewBusinessError(10031, "Failed to parse total from cache")
	ErrOrderCacheUnmarshalFailed  = NewBusinessError(10032, "Failed to unmarshal orders from cache")
	ErrRedisScanKeysFailed        = NewBusinessError(10033, "Failed to scan keys")
	ErrOrderListCacheDeleteFailed = NewBusinessError(10034, "Failed to delete order list cache")
	ErrInternalError              = NewBusinessError(50000, "Internal server error")
)
