package models

import "time"

type RecipeStepModel struct {
	ID           uint64    `gorm:"column:id;primaryKey"`
	ItemKey      string    `gorm:"column:item_key"`
	StepOrder    int       `gorm:"column:step_order"`
	Description  string    `gorm:"column:description"`
	CounterID    uint64    `gorm:"column:counter_id"`
	MachineID    *uint64   `gorm:"column:machine_id"`
	EstimateSecs int       `gorm:"column:estimate_secs"`
	BasePriority float64   `gorm:"column:base_priority"`
	IsActive     bool      `gorm:"column:is_active"`
	CreatedAt    time.Time `gorm:"column:created_at"`
}

func (RecipeStepModel) TableName() string { return "recipe_steps" }
