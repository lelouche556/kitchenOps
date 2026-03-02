package temporalx

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"SwishAssignment/internal/models"
	"SwishAssignment/internal/orchestrator"
	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/workflow"
)

const AllocationTaskQueue = "kitchen-allocation-task-queue"

func AllocationWorkflow(ctx workflow.Context, event models.DomainEvent) error {
	options := workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Second,
	}
	ctx = workflow.WithActivityOptions(ctx, options)
	return workflow.ExecuteActivity(ctx, "RunAllocationActivity", event).Get(ctx, nil)
}

func AutoCompleteTaskWorkflow(ctx workflow.Context, event models.DomainEvent) error {
	options := workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Second,
	}
	ctx = workflow.WithActivityOptions(ctx, options)

	var info AutoCompleteTaskInfo
	if err := workflow.ExecuteActivity(ctx, "PrepareAutoCompleteTaskActivity", event).Get(ctx, &info); err != nil {
		return err
	}
	if info.EstimateSecs < 1 {
		info.EstimateSecs = 1
	}

	if err := workflow.Sleep(ctx, time.Duration(info.EstimateSecs)*time.Second); err != nil {
		return err
	}
	return workflow.ExecuteActivity(ctx, "CompleteTaskByIDActivity", info.TaskID).Get(ctx, nil)
}

type AllocationActivities struct {
	Orchestrator *orchestrator.AllocationOrchestrator
	Kitchen      TaskLifecycleService
}

type TaskLifecycleService interface {
	GetTask(ctx context.Context, taskID uint64) (models.TaskModel, error)
	CompleteTask(ctx context.Context, taskID uint64) error
}

type AutoCompleteTaskInfo struct {
	TaskID       uint64
	EstimateSecs int
}

func (a *AllocationActivities) RunAllocationActivity(ctx context.Context, event models.DomainEvent) error {
	logger := activity.GetLogger(ctx)
	logger.Info("processing domain event", "type", string(event.Type), "payload", event.Payload)
	return a.Orchestrator.HandleEvent(ctx, event)
}

func (a *AllocationActivities) PrepareAutoCompleteTaskActivity(ctx context.Context, event models.DomainEvent) (AutoCompleteTaskInfo, error) {
	taskID, err := parseTaskIDPayload(event.Payload)
	if err != nil {
		return AutoCompleteTaskInfo{}, err
	}
	task, err := a.Kitchen.GetTask(ctx, taskID)
	if err != nil {
		return AutoCompleteTaskInfo{}, err
	}
	return AutoCompleteTaskInfo{
		TaskID:       task.ID,
		EstimateSecs: task.EstimateSecs,
	}, nil
}

func (a *AllocationActivities) CompleteTaskByIDActivity(ctx context.Context, taskID uint64) error {
	if err := a.Kitchen.CompleteTask(ctx, taskID); err != nil {
		msg := err.Error()
		if strings.Contains(msg, "is not STARTED") || strings.Contains(msg, "not found") {
			return nil
		}
		return err
	}
	return nil
}

func parseTaskIDPayload(payload string) (uint64, error) {
	id, err := strconv.ParseUint(strings.TrimSpace(payload), 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid task payload %q: %w", payload, err)
	}
	return id, nil
}
