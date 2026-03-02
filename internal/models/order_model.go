package models

import "time"

type OrderModel struct {
	ID              uint64    `gorm:"column:id;primaryKey"`
	ExternalOrderID string    `gorm:"column:external_order_id"`
	Status          string    `gorm:"column:status"`
	CreatedAt       time.Time `gorm:"column:created_at"`
}

func (OrderModel) TableName() string { return "orders" }
