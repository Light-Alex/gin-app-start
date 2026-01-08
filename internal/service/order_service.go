package service

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"gin-app-start/internal/common"
	"gin-app-start/internal/dto"
	"gin-app-start/internal/model"
	"gin-app-start/internal/redis"
	"gin-app-start/internal/repository"
	"gin-app-start/pkg/errors"
	"gin-app-start/pkg/utils"

	"gorm.io/gorm"
)

var _ OrderService = (*orderService)(nil)

type OrderService interface {
	SaveOrderInCache(ctx common.Context, order *model.Order, expireTime time.Duration) error
	SaveOrderListInCache(ctx common.Context, orders []*model.Order, total int64, username string, page, pageSize int, expireTime time.Duration) error
	DeleteOrderListCache(ctx common.Context) error

	CreateOrder(ctx common.Context, req *dto.CreateOrderRequest) (*model.Order, error)
	GetOrderByOrderNumber(ctx common.Context, orderNumber string) (*model.Order, error)
	UpdateOrderByOrderNumber(ctx common.Context, req *dto.UpdateOrderRequest) (*model.Order, error)
	DeleteOrderByOrderNumber(ctx common.Context, orderNumber string) error
	GetOrderByID(ctx common.Context, id uint) (*model.Order, error)
	UpdateOrder(ctx common.Context, id uint, req *dto.UpdateOrderRequest) (*model.Order, error)
	DeleteOrder(ctx common.Context, id uint) error
	ListOrders(ctx common.Context, username string, page, pageSize int) ([]*model.Order, int64, error)
}

type orderService struct {
	orderRepo  repository.OrderRepository
	redisCache redis.RedisRepository
}

func NewOrderService(orderRepo repository.OrderRepository, redisCache redis.RedisRepository) OrderService {
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

func (s *orderService) SaveOrderInCache(ctx common.Context, order *model.Order, expireTime time.Duration) error {
	cacheKey := s.getOrderCacheKey(order.OrderNumber)

	data, err := json.MarshalIndent(order, "", "  ")
	if err != nil {
		return err
	}

	if err := s.redisCache.SetWithExpire(cacheKey, string(data), expireTime, redis.WithTrace(ctx.Trace())); err != nil {
		return err
	}
	return nil
}

// 保存订单列表到Redis缓存, 设置过期时间为expireTime
func (s *orderService) SaveOrderListInCache(ctx common.Context, orders []*model.Order, total int64, username string, page, pageSize int, expireTime time.Duration) error {
	cacheKey := s.getOrderListCacheKey(username, page, pageSize)

	data, err := json.MarshalIndent(orders, "", "  ")
	if err != nil {
		return err
	}

	params := redis.HashParams{
		Options: []redis.Option{redis.WithTrace(ctx.Trace())},
		Values: []interface{}{
			"orders", data,
			"total", total,
		},
	}

	if err := s.redisCache.HashSet(cacheKey, expireTime, params); err != nil {
		return err
	}

	return nil
}

// 删除订单列表缓存
func (s *orderService) DeleteOrderListCache(ctx common.Context) error {
	pattern := "order_list:*"
	redisCtx := s.redisCache.GetRedisContext()
	keys, err := s.redisCache.GetRedisClient().Keys(redisCtx, pattern).Result()
	if err != nil {
		return err
	}

	for _, key := range keys {
		if err := s.redisCache.Delete(key, redis.WithTrace(ctx.Trace())); err != nil {
			return err
		}
	}
	return nil
}

func (s *orderService) CreateOrder(ctx common.Context, req *dto.CreateOrderRequest) (*model.Order, error) {
	// 生成订单号
	orderNumber := utils.GenerateOrderNumberWithPrefix("EC")

	order, err := s.GetOrderByOrderNumber(ctx, orderNumber)
	// 如果订单号已存在, 则重新生成
	if err == nil && order != nil {
		return nil, errors.New("Order number already exists")
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
		return nil, err
	}

	// 保存订单到Redis, 设置订单缓存过期时间为30min
	if err := s.SaveOrderInCache(ctx, order, 30*time.Minute); err != nil {
		return nil, err
	}

	// 删除订单列表缓存
	if err := s.DeleteOrderListCache(ctx); err != nil {
		return nil, err
	}

	return order, nil
}

func (s *orderService) GetOrderByOrderNumber(ctx common.Context, orderNumber string) (*model.Order, error) {
	cacheKey := s.getOrderCacheKey(orderNumber)

	// 检查缓存中是否已存在该订单号
	orderStr, err := s.redisCache.Get(cacheKey)
	if err == nil && orderStr != "" {
		var order model.Order
		if err := json.Unmarshal([]byte(orderStr), &order); err == nil {
			return &order, nil
		}
	}

	if err == nil && orderStr == "" {
		return nil, nil
	}

	order, err := s.orderRepo.GetOrderByOrderNumber(ctx, orderNumber)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			// 缓存空值，防止缓存穿透
			if err := s.redisCache.SetWithExpire(cacheKey, "", 30*time.Minute); err != nil {
				return nil, err
			}
			return nil, err
		}
		return nil, err
	}

	// 保存订单到Redis, 设置订单缓存过期时间为30min
	if err := s.SaveOrderInCache(ctx, order, 30*time.Minute); err != nil {
		return nil, err
	}
	return order, nil
}

func (s *orderService) UpdateOrderByOrderNumber(ctx common.Context, req *dto.UpdateOrderRequest) (*model.Order, error) {
	orderNumber := req.OrderNumber
	order, err := s.GetOrderByOrderNumber(ctx, orderNumber)
	if err != nil || order == nil {
		return nil, err
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
		return nil, err
	}

	// 保存订单到Redis, 设置订单缓存过期时间为30min
	if err := s.SaveOrderInCache(ctx, order, 30*time.Minute); err != nil {
		return nil, err
	}

	// 删除订单列表缓存
	if err := s.DeleteOrderListCache(ctx); err != nil {
		return nil, err
	}
	return order, nil
}

func (s *orderService) DeleteOrderByOrderNumber(ctx common.Context, orderNumber string) error {
	order, err := s.GetOrderByOrderNumber(ctx, orderNumber)
	if err != nil || order == nil {
		return err
	}

	// 删除订单缓存
	if err := s.redisCache.Delete(s.getOrderCacheKey(orderNumber)); err != nil {
		return err
	}

	// 删除订单列表缓存
	if err := s.DeleteOrderListCache(ctx); err != nil {
		return err
	}

	if err := s.orderRepo.Delete(ctx, order.ID); err != nil {
		return err
	}
	return nil
}

func (s *orderService) GetOrderByID(ctx common.Context, id uint) (*model.Order, error) {
	order, err := s.orderRepo.GetByID(ctx, id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, err
		}
		return nil, err
	}
	return order, nil
}

func (s *orderService) UpdateOrder(ctx common.Context, id uint, req *dto.UpdateOrderRequest) (*model.Order, error) {
	order, err := s.orderRepo.GetByID(ctx, id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, err
		}
		return nil, err
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
		return nil, err
	}
	return order, nil
}

func (s *orderService) DeleteOrder(ctx common.Context, id uint) error {
	if err := s.orderRepo.Delete(ctx, id); err != nil {
		return err
	}
	return nil
}

func (s *orderService) ListOrders(ctx common.Context, username string, page, pageSize int) ([]*model.Order, int64, error) {
	// 从Redis缓存中获取订单列表
	cacheKey := s.getOrderListCacheKey(username, page, pageSize)
	cachedOrders, _ := s.redisCache.HashGet(cacheKey, "orders")
	cachedTotal, _ := s.redisCache.HashGet(cacheKey, "total")
	if cachedOrders != "" && cachedTotal != "" {
		total, err := strconv.ParseInt(cachedTotal, 10, 64)
		if err != nil {
			return nil, 0, err
		}

		var orders []*model.Order
		err = json.Unmarshal([]byte(cachedOrders), &orders)
		if err != nil {
			return nil, 0, err
		}

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
		return nil, 0, err
	}

	// 保存订单列表到Redis缓存, 设置过期时间为5min
	if err := s.SaveOrderListInCache(ctx, orders, total, username, page, pageSize, 30*time.Minute); err != nil {
		return nil, 0, err
	}

	return orders, total, nil
}
