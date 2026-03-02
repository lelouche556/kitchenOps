package repository

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"SwishAssignment/internal/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func (r *KitchenRepository) UpdateTaskStatusStarted(ctx context.Context, tx *gorm.DB, taskID uint64, startedAt time.Time) (models.TaskModel, error) {
	var task models.TaskModel
	if err := tx.WithContext(ctx).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("id = ?", taskID).
		First(&task).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return models.TaskModel{}, ErrNotFound
		}
		return models.TaskModel{}, err
	}
	if task.Status != string(models.TaskAssigned) {
		return models.TaskModel{}, fmt.Errorf("task %d is not ASSIGNED", task.ID)
	}
	if err := tx.WithContext(ctx).Model(&models.TaskModel{}).Where("id = ?", task.ID).
		Updates(map[string]any{"status": string(models.TaskStarted), "started_at": startedAt}).Error; err != nil {
		return models.TaskModel{}, err
	}
	task.Status = string(models.TaskStarted)
	task.StartedAt = &startedAt
	return task, nil
}

func (r *KitchenRepository) UpdateTaskStatusCompleted(ctx context.Context, tx *gorm.DB, taskID uint64, completedAt time.Time) (models.TaskModel, error) {
	var task models.TaskModel
	if err := tx.WithContext(ctx).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("id = ?", taskID).
		First(&task).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return models.TaskModel{}, ErrNotFound
		}
		return models.TaskModel{}, err
	}
	if task.Status != string(models.TaskStarted) {
		return models.TaskModel{}, fmt.Errorf("task %d is not STARTED", task.ID)
	}
	if err := tx.WithContext(ctx).Model(&models.TaskModel{}).Where("id = ?", task.ID).
		Updates(map[string]any{"status": string(models.TaskCompleted), "completed_at": completedAt}).Error; err != nil {
		return models.TaskModel{}, err
	}
	task.Status = string(models.TaskCompleted)
	task.CompletedAt = &completedAt
	return task, nil
}

func (r *KitchenRepository) IncrementCounterUsage(ctx context.Context, tx *gorm.DB, counterID uint64, delta int) error {
	return tx.WithContext(ctx).Model(&models.CounterModel{}).Where("id = ?", counterID).
		Updates(map[string]any{"in_use": gorm.Expr("GREATEST(in_use + ?, 0)", delta), "updated_at": time.Now()}).Error
}

func (r *KitchenRepository) IncrementMachineUsage(ctx context.Context, tx *gorm.DB, machineID uint64, delta int) error {
	return tx.WithContext(ctx).Model(&models.MachineModel{}).Where("id = ?", machineID).
		Updates(map[string]any{"in_use": gorm.Expr("GREATEST(in_use + ?, 0)", delta), "updated_at": time.Now()}).Error
}

func (r *KitchenRepository) DecrementDependentPendingDeps(ctx context.Context, tx *gorm.DB, completedTaskID uint64) ([]models.TaskModel, error) {
	if err := tx.WithContext(ctx).Exec(`
		UPDATE tasks t
		SET pending_deps = GREATEST(t.pending_deps - 1, 0)
		FROM task_dependencies d
		WHERE d.depends_on_task_id = ? AND d.task_id = t.id AND t.status = ?`, completedTaskID, string(models.TaskUnassigned)).Error; err != nil {
		return nil, err
	}

	var ready []models.TaskModel
	err := tx.WithContext(ctx).Raw(`
		SELECT t.*
		FROM tasks t
		JOIN task_dependencies d ON d.task_id = t.id
		WHERE d.depends_on_task_id = ? AND t.status = ? AND t.pending_deps = 0`, completedTaskID, string(models.TaskUnassigned)).Scan(&ready).Error
	return ready, err
}

func (r *KitchenRepository) UpdateStaffEfficiencyAndActive(ctx context.Context, tx *gorm.DB, staffID uint64, activeDelta int, estimateSecs int, startedAt, completedAt time.Time) error {
	actualSecs := int(completedAt.Sub(startedAt).Seconds())
	if actualSecs <= 0 {
		actualSecs = 1
	}
	perf := float64(estimateSecs) / float64(actualSecs)
	return tx.WithContext(ctx).Model(&models.StaffModel{}).Where("id = ?", staffID).
		Updates(map[string]any{
			"active_tasks":          gorm.Expr("GREATEST(active_tasks + ?, 0)", activeDelta),
			"active_seconds":        gorm.Expr("active_seconds + ?", actualSecs),
			"efficiency_multiplier": gorm.Expr("(efficiency_multiplier * 0.85) + (? * 0.15)", perf),
		}).Error
}

func (r *KitchenRepository) GetTaskByID(ctx context.Context, taskID uint64) (models.TaskModel, error) {
	var task models.TaskModel
	err := r.db.WithContext(ctx).Where("id = ?", taskID).First(&task).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return models.TaskModel{}, ErrNotFound
	}
	return task, err
}

func (r *KitchenRepository) RequeueExpiredAssignedTasks(ctx context.Context, cutoff time.Time, limit int) ([]models.TaskModel, error) {
	if limit <= 0 {
		limit = 50
	}
	requeued := make([]models.TaskModel, 0)
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var tasks []models.TaskModel
		if err := tx.WithContext(ctx).
			Clauses(clause.Locking{Strength: "UPDATE", Options: "SKIP LOCKED"}).
			Where("status = ? AND started_at IS NULL AND assigned_at IS NOT NULL AND assigned_at < ?", string(models.TaskAssigned), cutoff).
			Order("assigned_at ASC").
			Limit(limit).
			Find(&tasks).Error; err != nil {
			return err
		}

		for _, task := range tasks {
			if err := tx.WithContext(ctx).Model(&models.TaskModel{}).Where("id = ?", task.ID).
				Updates(map[string]any{
					"status":              string(models.TaskUnassigned),
					"assigned_staff_id":   nil,
					"assigned_machine_id": nil,
					"assigned_at":         nil,
					"claimed_by":          nil,
					"claimed_until":       nil,
				}).Error; err != nil {
				return err
			}
			if task.AssignedStaffID != nil {
				if err := tx.WithContext(ctx).Model(&models.StaffModel{}).Where("id = ?", *task.AssignedStaffID).
					Update("active_tasks", gorm.Expr("GREATEST(active_tasks - 1, 0)")).Error; err != nil {
					return err
				}
			}
			if err := tx.WithContext(ctx).Model(&models.CounterModel{}).Where("id = ?", task.CounterID).
				Updates(map[string]any{
					"in_use":     gorm.Expr("GREATEST(in_use - 1, 0)"),
					"updated_at": time.Now(),
				}).Error; err != nil {
				return err
			}
			if task.AssignedMachineID != nil {
				if err := tx.WithContext(ctx).Model(&models.MachineModel{}).Where("id = ?", *task.AssignedMachineID).
					Updates(map[string]any{
						"in_use":     gorm.Expr("GREATEST(in_use - 1, 0)"),
						"updated_at": time.Now(),
					}).Error; err != nil {
					return err
				}
			}
			if err := r.InsertDomainEvent(ctx, tx, "task", strconv.FormatUint(task.ID, 10), models.DomainEvent{
				Type:      models.EventTaskRequeued,
				Timestamp: int(time.Now().Unix()),
				Payload:   strconv.FormatUint(task.ID, 10),
			}); err != nil {
				return err
			}
			task.Status = string(models.TaskUnassigned)
			task.AssignedStaffID = nil
			task.AssignedMachineID = nil
			task.AssignedAt = nil
			requeued = append(requeued, task)
		}
		return nil
	})
	return requeued, err
}
