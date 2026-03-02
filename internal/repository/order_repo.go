package repository

import (
	"context"

	"SwishAssignment/internal/models"
	"gorm.io/gorm"
)

func (r *KitchenRepository) CreateOrderWithTasks(ctx context.Context, externalOrderID string, tasks []models.TaskModel, deps []models.TaskDependencyModel) (models.OrderModel, []models.TaskModel, error) {
	var order models.OrderModel
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		order = models.OrderModel{ExternalOrderID: externalOrderID, Status: string(models.OrderConfirmed)}
		if err := tx.Create(&order).Error; err != nil {
			return err
		}

		for i := range tasks {
			tasks[i].OrderID = order.ID
			tasks[i].Status = string(models.TaskUnassigned)
			if err := tx.Create(&tasks[i]).Error; err != nil {
				return err
			}
		}

		for _, d := range deps {
			if err := tx.Create(&d).Error; err != nil {
				return err
			}
		}
		return nil
	})
	return order, tasks, err
}

func (r *KitchenRepository) InsertTaskDependencies(ctx context.Context, deps []models.TaskDependencyModel) error {
	if len(deps) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).Create(&deps).Error
}

func (r *KitchenRepository) RecalculatePendingDepsForOrder(ctx context.Context, orderID uint64) error {
	return r.db.WithContext(ctx).Exec(`
		UPDATE tasks t
		SET pending_deps = sub.dep_count
		FROM (
			SELECT td.task_id, COUNT(*)::int AS dep_count
			FROM task_dependencies td
			JOIN tasks tx ON tx.id = td.task_id
			WHERE tx.order_id = ?
			GROUP BY td.task_id
		) sub
		WHERE t.id = sub.task_id
	`, orderID).Error
}

func (r *KitchenRepository) ListOrderReadyTasks(ctx context.Context, orderID uint64) ([]models.TaskModel, error) {
	var tasks []models.TaskModel
	err := r.db.WithContext(ctx).Where("order_id = ? AND status = ? AND pending_deps = 0", orderID, string(models.TaskUnassigned)).
		Order("id ASC").
		Find(&tasks).Error
	return tasks, err
}

func (r *KitchenRepository) CompleteOrderIfAllTasksDone(ctx context.Context, tx *gorm.DB, orderID uint64) (bool, error) {
	var remaining int64
	if err := tx.WithContext(ctx).Model(&models.TaskModel{}).
		Where("order_id = ? AND status <> ?", orderID, string(models.TaskCompleted)).
		Count(&remaining).Error; err != nil {
		return false, err
	}
	if remaining > 0 {
		return false, nil
	}

	if err := tx.WithContext(ctx).Model(&models.OrderModel{}).
		Where("id = ? AND status <> ?", orderID, string(models.OrderCompleted)).
		Update("status", string(models.OrderCompleted)).Error; err != nil {
		return false, err
	}
	return true, nil
}

func (r *KitchenRepository) MarkOrderInProgressIfConfirmed(ctx context.Context, tx *gorm.DB, orderID uint64) error {
	return tx.WithContext(ctx).Model(&models.OrderModel{}).
		Where("id = ? AND status = ?", orderID, string(models.OrderConfirmed)).
		Update("status", string(models.OrderInProgress)).Error
}

func (r *KitchenRepository) MarkOrderPartCompletedIfNotDone(ctx context.Context, tx *gorm.DB, orderID uint64) error {
	return tx.WithContext(ctx).Model(&models.OrderModel{}).
		Where("id = ? AND status IN ?", orderID, []string{string(models.OrderConfirmed), string(models.OrderInProgress)}).
		Update("status", string(models.OrderPartCompleted)).Error
}
