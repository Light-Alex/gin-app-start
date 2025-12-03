package repository

import (
	"context"
	"gin-app-start/internal/model"

	"gorm.io/gorm"
)

type OrderRepository interface {
	Create(ctx context.Context, order *model.Order) error
	GetByID(ctx context.Context, id uint) (*model.Order, error)
	GetOrderByOrderNumber(ctx context.Context, orderNumber string) (*model.Order, error)
	DeleteOrderByOrderNumber(ctx context.Context, orderNumber string) error
	Update(ctx context.Context, user *model.Order) error
	Delete(ctx context.Context, id uint) error
	List(ctx context.Context, offset, limit int) ([]*model.Order, int64, error)
	Count(ctx context.Context) (int64, error)
}

type orderRepository struct {
	*BaseRepository[model.Order]
}

func NewOrderRepository(db *gorm.DB) OrderRepository {
	return &orderRepository{
		BaseRepository: NewBaseRepository[model.Order](db),
	}
}

func (r *orderRepository) GetOrderByOrderNumber(ctx context.Context, orderNumber string) (*model.Order, error) {
	var order model.Order
	err := r.db.WithContext(ctx).Where("order_number = ?", orderNumber).First(&order).Error
	if err != nil {
		return nil, err
	}
	return &order, nil
}

func (r *orderRepository) DeleteOrderByOrderNumber(ctx context.Context, orderNumber string) error {
	return r.db.WithContext(ctx).Where("order_number = ?", orderNumber).Delete(&model.Order{}).Error
}
