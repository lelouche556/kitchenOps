package models

import "time"

type MachineModel struct {
	ID          uint64    `gorm:"column:id;primaryKey"`
	Name        string    `gorm:"column:name"`
	CounterID   uint64    `gorm:"column:counter_id"`
	MachineType string    `gorm:"column:machine_type"`
	Capacity    int       `gorm:"column:capacity"`
	InUse       int       `gorm:"column:in_use"`
	IsUp        bool      `gorm:"column:is_up"`
	UpdatedAt   time.Time `gorm:"column:updated_at"`
}

func (MachineModel) TableName() string { return "machines" }
