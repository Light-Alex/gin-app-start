package repository

import (
	"gin-app-start/internal/common"
	"gin-app-start/internal/model"

	"gorm.io/gorm"
)

type UserRepository interface {
	Create(ctx common.Context, user *model.User) error
	GetByID(ctx common.Context, id uint) (*model.User, error)
	GetByUsername(ctx common.Context, username string) (*model.User, error)
	GetByEmail(ctx common.Context, email string) (*model.User, error)
	GetByPhone(ctx common.Context, phone string) (*model.User, error)
	Update(ctx common.Context, user *model.User) error
	Delete(ctx common.Context, id uint) error
	List(ctx common.Context, offset, limit int) ([]*model.User, int64, error)
}

type userRepository struct {
	*BaseRepository[model.User]
}

func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{
		BaseRepository: NewBaseRepository[model.User](db),
	}
}

func (r *userRepository) GetByUsername(ctx common.Context, username string) (*model.User, error) {
	var user model.User
	err := r.db.WithContext(ctx.RequestContext()).Where("username = ?", username).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) GetByEmail(ctx common.Context, email string) (*model.User, error) {
	var user model.User
	err := r.db.WithContext(ctx.RequestContext()).Where("email = ?", email).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) GetByPhone(ctx common.Context, phone string) (*model.User, error) {
	var user model.User
	err := r.db.WithContext(ctx.RequestContext()).Where("phone = ?", phone).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// List 分页查询用户列表
//
// 该方法实现了用户数据的分页查询功能，支持分页参数和总数统计，
// 适用于前端表格展示、数据导出等需要分页的场景。
//
// 参数:
//   - ctx: 上下文，用于超时控制、取消操作等
//   - offset: 偏移量，表示跳过的记录数（从0开始）
//   - limit: 每页记录数，控制返回的用户数量
//
// 返回值:
//   - []*model.User: 用户列表切片，包含查询到的用户数据
//   - int64: 用户总数，用于前端分页组件计算总页数
//   - error: 错误信息，成功时为nil
func (r *userRepository) List(ctx common.Context, offset, limit int) ([]*model.User, int64, error) {
	var users []*model.User
	var total int64

	if err := r.db.WithContext(ctx.RequestContext()).Model(&model.User{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := r.db.WithContext(ctx.RequestContext()).Offset(offset).Limit(limit).Find(&users).Error
	return users, total, err
}
