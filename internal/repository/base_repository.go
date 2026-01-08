package repository

import (
	"gin-app-start/internal/common"

	"gorm.io/gorm"
)

type BaseRepository[T any] struct {
	db *gorm.DB
}

func NewBaseRepository[T any](db *gorm.DB) *BaseRepository[T] {
	return &BaseRepository[T]{db: db}
}

func (r *BaseRepository[T]) Create(ctx common.Context, entity *T) error {
	return r.db.WithContext(ctx.RequestContext()).Create(entity).Error
}

func (r *BaseRepository[T]) GetByID(ctx common.Context, id uint) (*T, error) {
	var entity T
	err := r.db.WithContext(ctx.RequestContext()).First(&entity, id).Error
	if err != nil {
		return nil, err
	}
	return &entity, nil
}

func (r *BaseRepository[T]) Update(ctx common.Context, entity *T) error {
	return r.db.WithContext(ctx.RequestContext()).Save(entity).Error
}

func (r *BaseRepository[T]) Delete(ctx common.Context, id uint) error {
	// 软删除
	return r.db.WithContext(ctx.RequestContext()).Delete(new(T), id).Error

	// 硬删除（谨慎使用）
	// return r.db.WithContext(ctx).Unscoped().Delete(new(T), id).Error
}

func (r *BaseRepository[T]) List(ctx common.Context, offset, limit int) ([]*T, int64, error) {
	var entities []*T
	total, err := r.Count(ctx)
	if err != nil {
		return nil, 0, err
	}

	err = r.db.WithContext(ctx.RequestContext()).Offset(offset).Limit(limit).Find(&entities).Error
	return entities, total, err
}

func (r *BaseRepository[T]) Count(ctx common.Context) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx.RequestContext()).Model(new(T)).Count(&count).Error
	return count, err
}

func (r *BaseRepository[T]) GetDB() *gorm.DB {
	return r.db
}
