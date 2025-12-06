package utils

import (
	"fmt"
	"io"
	"math/rand"
	"mime/multipart"
	"os"
	"path"
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

func SaveToFile(file *multipart.FileHeader, dst string) (string, error) {
	// 获取后缀名
	ext := path.Ext(file.Filename)
	fileName := GenerateUUID() + ext
	src, err := file.Open()
	if err != nil {
		return fileName, err
	}
	defer src.Close()

	// 判断dst目录是否存在, 如果目录不存在，创建目录
	if _, err := os.Stat(dst); os.IsNotExist(err) {
		err := os.MkdirAll(dst, 0755)
		if err != nil {
			return fileName, err
		}
	}

	// 创建 dst 文件
	out, err := os.Create(path.Join(dst, fileName))
	if err != nil {
		return fileName, err
	}
	defer out.Close()
	_, err = io.Copy(out, src)
	if err != nil {
		return fileName, err
	}
	return fileName, nil
}
