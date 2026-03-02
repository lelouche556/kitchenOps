package models

import "time"

type TaskModel struct {
	ID                uint64     `gorm:"column:id;primaryKey"`
	OrderID           uint64     `gorm:"column:order_id"`
	Description       string     `gorm:"column:description"`
	CounterID         uint64     `gorm:"column:counter_id"`
	MachineID         *uint64    `gorm:"column:machine_id"`
	EstimateSecs      int        `gorm:"column:estimate_secs"`
	BasePriority      float64    `gorm:"column:base_priority"`
	PendingDeps       int        `gorm:"column:pending_deps"`
	AssignedStaffID   *uint64    `gorm:"column:assigned_staff_id"`
	AssignedMachineID *uint64    `gorm:"column:assigned_machine_id"`
	Status            string     `gorm:"column:status"`
	AssignmentVersion int        `gorm:"column:assignment_version"`
	ClaimedBy         *string    `gorm:"column:claimed_by"`
	ClaimedUntil      *time.Time `gorm:"column:claimed_until"`
	AssignedAt        *time.Time `gorm:"column:assigned_at"`
	StartedAt         *time.Time `gorm:"column:started_at"`
	CompletedAt       *time.Time `gorm:"column:completed_at"`
	CreatedAt         time.Time  `gorm:"column:created_at"`
}

func (TaskModel) TableName() string { return "tasks" }
