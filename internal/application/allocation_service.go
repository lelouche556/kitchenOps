package application

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"SwishAssignment/internal/models"
	"gorm.io/gorm"
)

func (s *KitchenAppService) AllocateOnce(ctx context.Context) (uint64, error) {
	now := time.Now()
	if err := s.requeueExpiredAssignedTasks(ctx, now); err != nil {
		return 0, err
	}

	ids, err := s.readyQueue.PopTopN(ctx, 20)
	if err != nil {
		return 0, err
	}
	if len(ids) == 0 {
		return 0, nil
	}

	var firstAssigned uint64
	for _, raw := range ids {
		taskID, convErr := strconv.ParseUint(raw, 10, 64)
		if convErr != nil {
			continue
		}

		var assignedTaskID uint64
		err := s.repo.DB().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
			task, err := s.repo.LockTaskForAssignment(ctx, tx, taskID)
			if err != nil {
				return err
			}

			counter, err := s.repo.LockCounter(ctx, tx, task.CounterID)
			if err != nil {
				return err
			}
			if counter.InUse >= counter.Capacity {
				return fmt.Errorf("counter busy")
			}

			staffList, err := s.repo.FindEligibleStaff(ctx, tx, task.CounterID, now)
			if err != nil {
				return err
			}
			if len(staffList) == 0 {
				return fmt.Errorf("no eligible staff")
			}

			if task.MachineID == nil {
				return fmt.Errorf("task %d missing machine_id", task.ID)
			}

			machine, err := s.repo.LockMachineForTask(ctx, tx, *task.MachineID, task.CounterID)
			if err != nil {
				return err
			}

			pick := s.pickStaff(staffList, now)
			if _, err := s.repo.LockStaff(ctx, tx, pick.ID); err != nil {
				return err
			}

			if err := s.repo.UpdateTaskAssignment(ctx, tx, task.ID, pick.ID, machine.ID, now); err != nil {
				return err
			}
			if err := s.repo.IncrementCounterUsage(ctx, tx, task.CounterID, 1); err != nil {
				return err
			}
			if err := s.repo.IncrementMachineUsage(ctx, tx, machine.ID, 1); err != nil {
				return err
			}
			if err := s.repo.MarkOrderInProgressIfConfirmed(ctx, tx, task.OrderID); err != nil {
				return err
			}
			if err := s.repo.IncrementStaffActiveTasks(ctx, tx, pick.ID, 1); err != nil {
				return err
			}
			if err := s.repo.InsertDomainEvent(ctx, tx, "task", strconv.FormatUint(task.ID, 10), models.DomainEvent{
				Type:      models.EventTaskAssigned,
				Timestamp: int(now.Unix()),
				Payload:   fmt.Sprintf("task=%d,staff=%d,machine=%d", task.ID, pick.ID, machine.ID),
			}); err != nil {
				return err
			}

			assignedTaskID = task.ID
			return nil
		})
		if err != nil {
			if task, getErr := s.repo.GetTaskByID(ctx, taskID); getErr == nil && task.Status == string(models.TaskUnassigned) {
				_ = s.readyQueue.Enqueue(ctx, strconv.FormatUint(task.ID, 10), s.queueScore(task.BasePriority, task.CreatedAt))
			}
			continue
		}

		if firstAssigned == 0 {
			firstAssigned = assignedTaskID
		}
	}

	return firstAssigned, nil
}

func (s *KitchenAppService) requeueExpiredAssignedTasks(ctx context.Context, now time.Time) error {
	cutoff := now.Add(-s.assignTTL)
	requeued, err := s.repo.RequeueExpiredAssignedTasks(ctx, cutoff, 50)
	if err != nil {
		return err
	}
	for _, task := range requeued {
		boostedPriority := task.BasePriority + 10
		if err := s.readyQueue.Enqueue(ctx, strconv.FormatUint(task.ID, 10), s.queueScore(boostedPriority, task.CreatedAt)); err != nil {
			return err
		}
	}
	return nil
}
