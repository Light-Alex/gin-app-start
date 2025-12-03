package model

import (
	"time"

	"gorm.io/gorm"
)

// Order represents an order in the system
type Order struct {
	ID          uint           `gorm:"primarykey" json:"id" example:"1"`
	OrderNumber string         `gorm:"unique;not null" json:"order_number" example:"EC20231215123456"`
	CreatedAt   time.Time      `json:"created_at" example:"2023-01-01T00:00:00Z"`
	UpdateAt    time.Time      `json:"update_at" example:"2023-01-01T00:00:00Z"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-" swaggerignore:"true"`
	UserID      uint           `gorm:"index;not null" json:"user_id" example:"1"`
	TotalPrice  float64        ` gorm:"type:decimal(10,2);not null" json:"total_price" example:"100.00"`
	Description string         `gorm:"size:256" json:"description" example:"Order for product A"`
	Status      int8           `gorm:"default:1;not null" json:"status" example:"1"`
}

func (Order) TableName() string {
	return "app_schema.orders" // 指定schema为app_schema；PostgreSQL格式: schema.table_name
}

func (o *Order) BeforeCreate(tx *gorm.DB) error {
	o.CreatedAt = time.Now()
	o.UpdateAt = time.Now()
	if o.Status == 0 {
		o.Status = 1
	}
	return nil
}

func (o *Order) BeforeUpdate(tx *gorm.DB) error {
	o.UpdateAt = time.Now()
	return nil
}
