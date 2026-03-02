package models

import "time"

type CounterModel struct {
	ID        uint64    `gorm:"column:id;primaryKey"`
	Name      string    `gorm:"column:name"`
	Capacity  int       `gorm:"column:capacity"`
	InUse     int       `gorm:"column:in_use"`
	UpdatedAt time.Time `gorm:"column:updated_at"`
}

func (CounterModel) TableName() string { return "counters" }
