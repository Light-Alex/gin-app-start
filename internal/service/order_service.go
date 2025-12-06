package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"gin-app-start/internal/dto"
	"gin-app-start/internal/model"
	"gin-app-start/internal/repository"
	"gin-app-start/pkg/errors"
	"gin-app-start/pkg/logger"
	"gin-app-start/pkg/utils"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

var _ OrderService = (*orderService)(nil)

type OrderService interface {
	CreateOrder(ctx context.Context, req *dto.CreateOrderRequest) (*model.Order, error)
	GetOrderByOrderNumber(ctx context.Context, orderNumber string) (*model.Order, error)
	UpdateOrderByOrderNumber(ctx context.Context, req *dto.UpdateOrderRequest) (*model.Order, error)
	DeleteOrderByOrderNumber(ctx context.Context, orderNumber string) error
	ListOrders(ctx context.Context, username string, page, pageSize int) ([]*model.Order, int64, error)

	GetOrderByID(ctx context.Context, id uint) (*model.Order, error)
	UpdateOrder(ctx context.Context, id uint, req *dto.UpdateOrderRequest) (*model.Order, error)
	DeleteOrder(ctx context.Context, id uint) error
}

type orderService struct {
	orderRepo  repository.OrderRepository
	redisCache repository.RedisRepository
}

func NewOrderService(orderRepo repository.OrderRepository, redisCache repository.RedisRepository) OrderService {
	return &orderService{
		orderRepo:  orderRepo,
		redisCache: redisCache,
	}
}

func (s *orderService) getOrderCacheKey(orderNumber string) string {
	return fmt.Sprintf("order:%s", orderNumber)
}

func (s *orderService) getOrderListCacheKey(username string, page, pageSize int) string {
	return fmt.Sprintf("order_list:%s:%d:%d", username, page, pageSize)
}

func (s *orderService) saveOrderInCache(order *model.Order, expireTime time.Duration) error {
	cacheKey := s.getOrderCacheKey(order.OrderNumber)

	data, err := json.MarshalIndent(order, "", "  ")
	if err != nil {
		logger.Error("Failed to marshal order", zap.Error(err), zap.String("order_number", order.OrderNumber))
		return errors.ErrOrderMarshalFailed
	}

	if err := s.redisCache.SetWithExpire(cacheKey, string(data), expireTime); err != nil {
		logger.Error("Failed to set order cache", zap.Error(err), zap.String("order_number", order.OrderNumber))
		return errors.ErrOrderCacheFailed
	}
	logger.Info("Order cached successfully", zap.String("order_number", order.OrderNumber))
	return nil
}

// 保存订单列表到Redis缓存, 设置过期时间为expireTime
func (s *orderService) saveOrderListInCache(orders []*model.Order, total int64, username string, page, pageSize int, expireTime time.Duration) error {
	cacheKey := s.getOrderListCacheKey(username, page, pageSize)

	data, err := json.MarshalIndent(orders, "", "  ")
	if err != nil {
		logger.Error("Failed to marshal order list", zap.Error(err), zap.Int("page", page), zap.Int("page_size", pageSize))
		return errors.ErrOrderMarshalFailed
	}

	s.redisCache.HashSet(cacheKey, expireTime, map[string]interface{}{
		"orders": data,
		"total":  total,
	})
	if err != nil {
		logger.Error("Failed to set order list cache", zap.Error(err), zap.Int("page", page), zap.Int("page_size", pageSize))
		return errors.ErrOrderCacheFailed
	}

	logger.Info("Order list cached successfully", zap.Int("page", page), zap.Int("page_size", pageSize))
	return nil
}

// 删除订单列表缓存
func (s *orderService) deleteOrderListCache() error {
	pattern := "order_list:*"
	redisCtx := s.redisCache.GetRedisContext()
	keys, err := s.redisCache.GetRedisClient().Keys(redisCtx, pattern).Result()
	if err != nil {
		logger.Error("Failed to scan keys", zap.Error(err), zap.String("pattern", pattern))
		return errors.ErrRedisScanKeysFailed
	}

	for _, key := range keys {
		if err := s.redisCache.Delete(key); err != nil {
			logger.Error("Failed to delete order list cache", zap.Error(err), zap.String("key", key))
			return errors.ErrOrderListCacheDeleteFailed
		}
	}
	logger.Info("Order list cache deleted successfully", zap.String("pattern", pattern))
	return nil
}

func (s *orderService) CreateOrder(ctx context.Context, req *dto.CreateOrderRequest) (*model.Order, error) {
	// 生成订单号
	orderNumber := utils.GenerateOrderNumberWithPrefix("EC")

	order, err := s.GetOrderByOrderNumber(ctx, orderNumber)
	// 如果订单号已存在, 则重新生成
	if err == nil && order != nil {
		logger.Error("Order already exists", zap.String("order_number", orderNumber))
		return nil, errors.ErrOrderExists
	}

	order = &model.Order{
		OrderNumber: orderNumber,
		Username:    req.Username,
		UserID:      req.UserId,
		TotalPrice:  req.TotalPrice,
		Description: req.Description,
		Status:      1,
	}

	// 保存订单到数据库
	if err := s.orderRepo.Create(ctx, order); err != nil {
		logger.Error("Failed to create order", zap.Error(err), zap.String("order_number", orderNumber))
		return nil, errors.ErrOrderCreateFailed
	}

	// 保存订单到Redis, 设置订单缓存过期时间为30min
	if err := s.saveOrderInCache(order, 30*time.Minute); err != nil {
		logger.Error("Failed to save order cache", zap.Error(err), zap.String("order_number", orderNumber))
		return nil, errors.ErrOrderCacheFailed
	}

	// 删除订单列表缓存
	if err := s.deleteOrderListCache(); err != nil {
		logger.Error("Failed to delete order list cache", zap.Error(err))
		return nil, errors.ErrOrderListCacheDeleteFailed
	}

	logger.Info("Order created successfully",
		zap.String("order_number", order.OrderNumber),
		zap.Uint("order_id", order.ID),
	)

	return order, nil
}

func (s *orderService) GetOrderByOrderNumber(ctx context.Context, orderNumber string) (*model.Order, error) {
	cacheKey := s.getOrderCacheKey(orderNumber)

	// 检查缓存中是否已存在该订单号
	orderStr, err := s.redisCache.Get(cacheKey)
	if err == nil && orderStr != "" {
		var order model.Order
		if err := json.Unmarshal([]byte(orderStr), &order); err == nil {
			logger.Info("Order retrieved from cache", zap.String("order_number", orderNumber))
			return &order, nil
		}
	}

	if err == nil && orderStr == "" {
		logger.Warn("Query too frequently", zap.String("order_number", orderNumber))
		return nil, nil
	}

	order, err := s.orderRepo.GetOrderByOrderNumber(ctx, orderNumber)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			// 缓存空值，防止缓存穿透
			if err := s.redisCache.SetWithExpire(cacheKey, "", 30*time.Minute); err != nil {
				logger.Error("Failed to set empty cache", zap.Error(err), zap.String("order_number", orderNumber))
			}
			return nil, errors.ErrEmptyCache
		}
		logger.Error("Failed to query order", zap.Error(err), zap.String("order_number", orderNumber))
		return nil, errors.ErrOrderFailed
	}

	// 保存订单到Redis, 设置订单缓存过期时间为30min
	if err := s.saveOrderInCache(order, 30*time.Minute); err != nil {
		logger.Error("Failed to save order cache", zap.Error(err), zap.String("order_number", orderNumber))
		return nil, errors.ErrOrderCacheFailed
	}
	return order, nil
}

func (s *orderService) UpdateOrderByOrderNumber(ctx context.Context, req *dto.UpdateOrderRequest) (*model.Order, error) {
	orderNumber := req.OrderNumber
	order, err := s.GetOrderByOrderNumber(ctx, orderNumber)
	if err != nil || order == nil {
		logger.Error("Order not found", zap.String("order_number", orderNumber))
		return nil, errors.ErrOrderNotFound
	}

	// 更新订单字段
	if req.TotalPrice != 0 {
		order.TotalPrice = req.TotalPrice
	}
	if req.Description != "" {
		order.Description = req.Description
	}
	if req.Status != 0 {
		order.Status = req.Status
	}

	if err := s.orderRepo.Update(ctx, order); err != nil {
		logger.Error("Failed to update order", zap.Error(err), zap.String("order_number", orderNumber))
		return nil, errors.ErrOrderUpdateFailed
	}

	// 保存订单到Redis, 设置订单缓存过期时间为30min
	if err := s.saveOrderInCache(order, 30*time.Minute); err != nil {
		logger.Error("Failed to save order cache", zap.Error(err), zap.String("order_number", orderNumber))
		return nil, errors.ErrOrderCacheFailed
	}

	// 删除订单列表缓存
	if err := s.deleteOrderListCache(); err != nil {
		logger.Error("Failed to delete order list cache", zap.Error(err))
		return nil, errors.ErrOrderListCacheDeleteFailed
	}

	logger.Info("Order updated successfully", zap.String("order_number", orderNumber))
	return order, nil
}

func (s *orderService) DeleteOrderByOrderNumber(ctx context.Context, orderNumber string) error {
	order, err := s.GetOrderByOrderNumber(ctx, orderNumber)
	if err != nil || order == nil {
		logger.Error("Order not found", zap.String("order_number", orderNumber))
		return errors.ErrOrderNotFound
	}

	// 删除订单缓存
	if err := s.redisCache.Delete(s.getOrderCacheKey(orderNumber)); err != nil {
		logger.Error("Failed to delete order cache", zap.Error(err), zap.String("order_number", orderNumber))
		return errors.ErrOrderCacheDeleteFailed
	}

	// 删除订单列表缓存
	if err := s.deleteOrderListCache(); err != nil {
		logger.Error("Failed to delete order list cache", zap.Error(err))
		return errors.ErrOrderListCacheDeleteFailed
	}

	if err := s.orderRepo.Delete(ctx, order.ID); err != nil {
		logger.Error("Failed to delete order", zap.Error(err), zap.String("order_number", orderNumber))
		return errors.ErrOrderDeleteFailed
	}

	logger.Info("Order deleted successfully", zap.String("order_number", orderNumber))
	return nil
}

func (s *orderService) GetOrderByID(ctx context.Context, id uint) (*model.Order, error) {
	order, err := s.orderRepo.GetByID(ctx, id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			logger.Error("Order not found", zap.Uint("order_id", id))
			return nil, errors.ErrOrderNotFound
		}
		logger.Error("Failed to query order", zap.Error(err), zap.Uint("order_id", id))
		return nil, errors.ErrOrderFailed
	}
	return order, nil
}

func (s *orderService) UpdateOrder(ctx context.Context, id uint, req *dto.UpdateOrderRequest) (*model.Order, error) {
	order, err := s.orderRepo.GetByID(ctx, id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			logger.Error("Order not found", zap.Uint("order_id", id))
			return nil, errors.ErrOrderNotFound
		}
		logger.Error("Failed to query order", zap.Error(err), zap.Uint("order_id", id))
		return nil, errors.ErrOrderFailed
	}

	// 更新订单字段
	if req.TotalPrice != 0 {
		order.TotalPrice = req.TotalPrice
	}
	if req.Description != "" {
		order.Description = req.Description
	}
	if req.Status != 0 {
		order.Status = req.Status
	}

	if err := s.orderRepo.Update(ctx, order); err != nil {
		logger.Error("Failed to update order", zap.Error(err), zap.Uint("order_id", id))
		return nil, errors.ErrOrderUpdateFailed
	}

	logger.Info("Order updated successfully", zap.Uint("order_id", id))
	return order, nil
}

func (s *orderService) DeleteOrder(ctx context.Context, id uint) error {
	if err := s.orderRepo.Delete(ctx, id); err != nil {
		logger.Error("Failed to delete order", zap.Error(err), zap.Uint("order_id", id))
		return errors.ErrOrderDeleteFailed
	}
	logger.Info("Order deleted successfully", zap.Uint("order_id", id))
	return nil
}

func (s *orderService) ListOrders(ctx context.Context, username string, page, pageSize int) ([]*model.Order, int64, error) {
	// 从Redis缓存中获取订单列表
	cacheKey := s.getOrderListCacheKey(username, page, pageSize)
	cachedOrders, _ := s.redisCache.HashGet(cacheKey, "orders")
	cachedTotal, _ := s.redisCache.HashGet(cacheKey, "total")
	if cachedOrders != "" && cachedTotal != "" {
		total, err := strconv.ParseInt(cachedTotal, 10, 64)
		if err != nil {
			logger.Error("Failed to parse total from cache", zap.Error(err), zap.String("total", cachedTotal))
			return nil, 0, errors.ErrOrderCacheParseTotalFailed
		}

		var orders []*model.Order
		err = json.Unmarshal([]byte(cachedOrders), &orders)
		if err != nil {
			logger.Error("Failed to unmarshal orders from cache", zap.Error(err), zap.String("orders", cachedOrders))
			return nil, 0, errors.ErrOrderCacheUnmarshalFailed
		}

		logger.Info("Orders retrieved from cache", zap.Int("page", page), zap.Int("page_size", pageSize), zap.Int64("total", total))
		return orders, total, nil
	}

	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}
	if pageSize > 100 {
		pageSize = 100
	}

	offset := (page - 1) * pageSize
	orders, total, err := s.orderRepo.List(ctx, username, offset, pageSize)
	if err != nil {
		logger.Error("Failed to list orders", zap.Error(err), zap.Int("page", page), zap.Int("page_size", pageSize))
		return nil, 0, errors.ErrOrderListFailed
	}

	// 保存订单列表到Redis缓存, 设置过期时间为5min
	if err := s.saveOrderListInCache(orders, total, username, page, pageSize, 30*time.Minute); err != nil {
		logger.Error("Failed to save order list cache", zap.Error(err), zap.Int("page", page), zap.Int("page_size", pageSize))
		return nil, 0, errors.ErrOrderCacheFailed
	}

	return orders, total, nil
}
