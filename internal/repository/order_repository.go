package repository

import (
	"gin-app-start/internal/common"
	"gin-app-start/internal/model"

	"gorm.io/gorm"
)

type OrderRepository interface {
	Create(ctx common.Context, order *model.Order) error
	GetByID(ctx common.Context, id uint) (*model.Order, error)
	GetOrderByOrderNumber(ctx common.Context, orderNumber string) (*model.Order, error)
	DeleteOrderByOrderNumber(ctx common.Context, orderNumber string) error
	Update(ctx common.Context, user *model.Order) error
	Delete(ctx common.Context, id uint) error
	List(ctx common.Context, username string, offset, limit int) ([]*model.Order, int64, error)
	Count(ctx common.Context) (int64, error)
}

type orderRepository struct {
	*BaseRepository[model.Order]
}

func NewOrderRepository(db *gorm.DB) OrderRepository {
	return &orderRepository{
		BaseRepository: NewBaseRepository[model.Order](db),
	}
}

func (r *orderRepository) GetOrderByOrderNumber(ctx common.Context, orderNumber string) (*model.Order, error) {
	var order model.Order
	err := r.db.WithContext(ctx.RequestContext()).Where("order_number = ?", orderNumber).First(&order).Error
	if err != nil {
		return nil, err
	}
	return &order, nil
}

func (r *orderRepository) DeleteOrderByOrderNumber(ctx common.Context, orderNumber string) error {
	return r.db.WithContext(ctx.RequestContext()).Where("order_number = ?", orderNumber).Delete(&model.Order{}).Error
}

func (r *orderRepository) List(ctx common.Context, username string, offset, limit int) ([]*model.Order, int64, error) {
	var orders []*model.Order
	var err error
	if username == common.ADMIN_NAME {
		err = r.db.WithContext(ctx.RequestContext()).Offset(offset).Limit(limit).Find(&orders).Error
	} else {
		err = r.db.WithContext(ctx.RequestContext()).Offset(offset).Limit(limit).Where("username = ?", username).Find(&orders).Error
	}
	total := int64(len(orders))
	return orders, total, err
}
