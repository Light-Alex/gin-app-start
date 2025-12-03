package utils

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/google/uuid"
)

func GenerateUUID() string {
	return uuid.New().String()
}

func RandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	seededRand := rand.New(rand.NewSource(time.Now().UnixNano()))

	b := make([]byte, length)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(b)
}

func Contains[T comparable](slice []T, target T) bool {
	for _, item := range slice {
		if item == target {
			return true
		}
	}
	return false
}

func Pointer[T any](v T) *T {
	return &v
}

// GenerateOrderNumberWithPrefix 生成带业务前缀的订单号
// 格式: 前缀 + 年月日 + 6位随机数 (示例: EC20231215123456)
func GenerateOrderNumberWithPrefix(prefix string) string {
	now := time.Now()

	// 格式化时间部分: 年月日
	datePart := now.Format("20060102")

	// 生成6位随机数
	randomPart := fmt.Sprintf("%06d", rand.Intn(1000000))

	return fmt.Sprintf("%s%s%s", prefix, datePart, randomPart)
}
