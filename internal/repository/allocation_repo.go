package repository

import (
	"context"
	"errors"
	"time"

	"SwishAssignment/internal/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func (r *KitchenRepository) LockTaskForAssignment(ctx context.Context, tx *gorm.DB, taskID uint64) (models.TaskModel, error) {
	var task models.TaskModel
	err := tx.WithContext(ctx).
		Clauses(clause.Locking{Strength: "UPDATE", Options: "SKIP LOCKED"}).
		Where("id = ? AND status = ? AND pending_deps = 0", taskID, string(models.TaskUnassigned)).
		First(&task).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return models.TaskModel{}, ErrNotFound
	}
	return task, err
}

func (r *KitchenRepository) FindEligibleStaff(ctx context.Context, tx *gorm.DB, counterID uint64, now time.Time) ([]models.StaffModel, error) {
	var staff []models.StaffModel
	err := tx.WithContext(ctx).
		Table("staff s").
		Select("s.*").
		Joins("JOIN staff_skills sk ON sk.staff_id = s.id").
		Where("sk.counter_id = ? AND s.on_break = FALSE AND s.shift_start <= ? AND s.shift_end >= ? AND s.active_tasks < s.max_parallel", counterID, now, now).
		Scan(&staff).Error
	return staff, err
}

func (r *KitchenRepository) LockCounter(ctx context.Context, tx *gorm.DB, counterID uint64) (models.CounterModel, error) {
	var c models.CounterModel
	err := tx.WithContext(ctx).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("id = ?", counterID).
		First(&c).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return models.CounterModel{}, ErrNotFound
	}
	return c, err
}

func (r *KitchenRepository) LockMachineForTask(ctx context.Context, tx *gorm.DB, machineID, counterID uint64) (models.MachineModel, error) {
	var m models.MachineModel
	err := tx.WithContext(ctx).
		Clauses(clause.Locking{Strength: "UPDATE", Options: "SKIP LOCKED"}).
		Where("machines.id = ? AND machines.counter_id = ? AND machines.is_up = TRUE AND machines.in_use < machines.capacity", machineID, counterID).
		Order("machines.id ASC").
		First(&m).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return models.MachineModel{}, ErrNotFound
	}
	return m, err
}

func (r *KitchenRepository) LockStaff(ctx context.Context, tx *gorm.DB, staffID uint64) (models.StaffModel, error) {
	var s models.StaffModel
	err := tx.WithContext(ctx).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("id = ?", staffID).
		First(&s).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return models.StaffModel{}, ErrNotFound
	}
	return s, err
}

func (r *KitchenRepository) UpdateTaskAssignment(ctx context.Context, tx *gorm.DB, taskID, staffID, machineID uint64, assignedAt time.Time) error {
	return tx.WithContext(ctx).Model(&models.TaskModel{}).
		Where("id = ?", taskID).
		Updates(map[string]any{
			"assigned_staff_id":   staffID,
			"assigned_machine_id": machineID,
			"status":              string(models.TaskAssigned),
			"assigned_at":         assignedAt,
			"assignment_version":  gorm.Expr("assignment_version + 1"),
		}).Error
}

func (r *KitchenRepository) IncrementStaffActiveTasks(ctx context.Context, tx *gorm.DB, staffID uint64, delta int) error {
	return tx.WithContext(ctx).Model(&models.StaffModel{}).Where("id = ?", staffID).
		Update("active_tasks", gorm.Expr("GREATEST(active_tasks + ?, 0)", delta)).Error
}
