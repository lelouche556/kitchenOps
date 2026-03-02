package application

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"SwishAssignment/internal/models"
	"gorm.io/gorm"
)

func (s *KitchenAppService) StartTask(ctx context.Context, taskID uint64) error {
	now := time.Now()

	if err := s.repo.DB().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		task, err := s.repo.UpdateTaskStatusStarted(ctx, tx, taskID, now)
		if err != nil {
			return err
		}
		if task.AssignedMachineID == nil {
			return fmt.Errorf("task %d missing assigned machine", task.ID)
		}
		if err := s.repo.InsertDomainEvent(ctx, tx, "task", strconv.FormatUint(task.ID, 10), models.DomainEvent{
			Type:      models.EventTaskStarted,
			Timestamp: int(now.Unix()),
			Payload:   strconv.FormatUint(task.ID, 10),
		}); err != nil {
			return err
		}
		return nil
	}); err != nil {
		return err
	}
	return nil
}

func (s *KitchenAppService) CompleteTask(ctx context.Context, taskID uint64) error {
	now := time.Now()

	if err := s.repo.DB().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		task, err := s.repo.UpdateTaskStatusCompleted(ctx, tx, taskID, now)
		if err != nil {
			return err
		}
		if task.AssignedMachineID == nil || task.AssignedStaffID == nil || task.StartedAt == nil {
			return fmt.Errorf("task %d missing assignment/start metadata", task.ID)
		}
		if err := s.repo.IncrementCounterUsage(ctx, tx, task.CounterID, -1); err != nil {
			return err
		}
		if err := s.repo.IncrementMachineUsage(ctx, tx, *task.AssignedMachineID, -1); err != nil {
			return err
		}
		if err := s.repo.UpdateStaffEfficiencyAndActive(ctx, tx, *task.AssignedStaffID, -1, task.EstimateSecs, *task.StartedAt, now); err != nil {
			return err
		}

		readyChildren, err := s.repo.DecrementDependentPendingDeps(ctx, tx, task.ID)
		if err != nil {
			return err
		}
		for _, child := range readyChildren {
			if child.Status == string(models.TaskUnassigned) && child.PendingDeps == 0 {
				if err := s.readyQueue.Enqueue(ctx, strconv.FormatUint(child.ID, 10), s.queueScore(child.BasePriority, child.CreatedAt)); err != nil {
					return err
				}
				if err := s.repo.InsertDomainEvent(ctx, tx, "task", strconv.FormatUint(child.ID, 10), models.DomainEvent{
					Type:      models.EventTaskReady,
					Timestamp: int(now.Unix()),
					Payload:   strconv.FormatUint(child.ID, 10),
				}); err != nil {
					return err
				}
			}
		}

		if err := s.repo.InsertDomainEvent(ctx, tx, "task", strconv.FormatUint(task.ID, 10), models.DomainEvent{
			Type:      models.EventTaskCompleted,
			Timestamp: int(now.Unix()),
			Payload:   strconv.FormatUint(task.ID, 10),
		}); err != nil {
			return err
		}
		if err := s.repo.MarkOrderPartCompletedIfNotDone(ctx, tx, task.OrderID); err != nil {
			return err
		}

		if _, err := s.repo.CompleteOrderIfAllTasksDone(ctx, tx, task.OrderID); err != nil {
			return err
		}
		return nil
	}); err != nil {
		return err
	}

	return nil
}

func (s *KitchenAppService) GetTask(ctx context.Context, taskID uint64) (models.TaskModel, error) {
	return s.repo.GetTaskByID(ctx, taskID)
}
