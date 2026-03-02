package models

import "time"

type StaffModel struct {
	ID                   uint64    `gorm:"column:id;primaryKey"`
	Name                 string    `gorm:"column:name"`
	ShiftStart           time.Time `gorm:"column:shift_start"`
	ShiftEnd             time.Time `gorm:"column:shift_end"`
	MaxParallel          int       `gorm:"column:max_parallel"`
	EfficiencyMultiplier float64   `gorm:"column:efficiency_multiplier"`
	ActiveTasks          int       `gorm:"column:active_tasks"`
	ActiveSeconds        int64     `gorm:"column:active_seconds"`
	OnBreak              bool      `gorm:"column:on_break"`
	CreatedAt            time.Time `gorm:"column:created_at"`
}

func (StaffModel) TableName() string { return "staff" }
