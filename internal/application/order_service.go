package application

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"SwishAssignment/internal/models"
)

func (s *KitchenAppService) ConfirmOrder(ctx context.Context, req ConfirmOrderRequest) (uint64, []uint64, error) {
	if len(req.Items) == 0 {
		return 0, nil, fmt.Errorf("items are required")
	}
	if req.ExternalOrderID == "" {
		req.ExternalOrderID = fmt.Sprintf("order-%d", time.Now().UnixNano())
	}

	itemKeys := make([]string, 0, len(req.Items))
	for item := range req.Items {
		itemKeys = append(itemKeys, item)
	}
	stepsByItem, err := s.repo.GetActiveRecipeStepsByItems(ctx, itemKeys)
	if err != nil {
		return 0, nil, err
	}
	explicitDepsByItem, err := s.repo.GetRecipeStepDependenciesByItems(ctx, itemKeys)
	if err != nil {
		return 0, nil, err
	}

	machineIDs := make([]uint64, 0)
	for _, steps := range stepsByItem {
		for _, step := range steps {
			if step.MachineID != nil {
				machineIDs = append(machineIDs, *step.MachineID)
			}
		}
	}
	machineCapByID, err := s.repo.GetMachineCapacitiesByIDs(ctx, machineIDs)
	if err != nil {
		return 0, nil, err
	}

	var taskRows []models.TaskModel
	depLinks := make([]taskDepPair, 0)
	for item, qty := range req.Items {
		steps := stepsByItem[item]
		if len(steps) == 0 {
			return 0, nil, fmt.Errorf("recipe not found for item %s", item)
		}
		if qty <= 0 {
			continue
		}
		for i := 0; i < qty; i++ {
			stepToTaskIdx := make(map[int]int, len(steps))
			for _, step := range steps {
				if step.MachineID == nil {
					return 0, nil, fmt.Errorf("recipe step missing machine for item %s step %d", item, step.StepOrder)
				}
				taskRows = append(taskRows, models.TaskModel{
					Description:  item + " :: " + step.Description,
					CounterID:    step.CounterID,
					MachineID:    step.MachineID,
					EstimateSecs: step.EstimateSecs,
					BasePriority: float64(step.BasePriority),
					PendingDeps:  0,
				})
				stepToTaskIdx[step.StepOrder] = len(taskRows) - 1
			}
			depLinks = append(depLinks, inferRecipeTaskDeps(steps, stepToTaskIdx, explicitDepsByItem[item], machineCapByID)...)
		}
	}

	for _, d := range depLinks {
		if d.TaskIdx < 0 || d.TaskIdx >= len(taskRows) || d.DepIdx < 0 || d.DepIdx >= len(taskRows) {
			return 0, nil, fmt.Errorf("invalid derived dependency indices task=%d dep=%d", d.TaskIdx, d.DepIdx)
		}
		if d.TaskIdx == d.DepIdx {
			return 0, nil, fmt.Errorf("invalid self dependency derived for task index %d", d.TaskIdx)
		}
	}

	if len(taskRows) == 0 {
		return 0, nil, fmt.Errorf("no tasks generated for order")
	}

	order, tasks, err := s.repo.CreateOrderWithTasks(ctx, req.ExternalOrderID, taskRows, nil)
	if err != nil {
		return 0, nil, err
	}

	if len(depLinks) > 0 {
		deps := make([]models.TaskDependencyModel, 0, len(depLinks))
		seen := make(map[string]struct{}, len(depLinks))
		for _, link := range depLinks {
			dep := models.TaskDependencyModel{
				TaskID:          tasks[link.TaskIdx].ID,
				DependsOnTaskID: tasks[link.DepIdx].ID,
			}
			key := strconv.FormatUint(dep.TaskID, 10) + ":" + strconv.FormatUint(dep.DependsOnTaskID, 10)
			if _, ok := seen[key]; ok {
				continue
			}
			seen[key] = struct{}{}
			deps = append(deps, dep)
		}
		if err := s.repo.InsertTaskDependencies(ctx, deps); err != nil {
			return 0, nil, err
		}
		if err := s.repo.RecalculatePendingDepsForOrder(ctx, order.ID); err != nil {
			return 0, nil, err
		}
	}

	readyTasks, err := s.repo.ListOrderReadyTasks(ctx, order.ID)
	if err != nil {
		return 0, nil, err
	}

	readyIDs := make([]uint64, 0)
	for _, t := range readyTasks {
		score := s.queueScore(t.BasePriority, t.CreatedAt)
		if err := s.readyQueue.Enqueue(ctx, strconv.FormatUint(t.ID, 10), score); err != nil {
			return 0, nil, err
		}
		readyIDs = append(readyIDs, t.ID)
		if err := s.repo.InsertDomainEventNoTx(ctx, "task", strconv.FormatUint(t.ID, 10), models.DomainEvent{
			Type:      models.EventTaskReady,
			Timestamp: int(time.Now().Unix()),
			Payload:   strconv.FormatUint(t.ID, 10),
		}); err != nil {
			return 0, nil, err
		}
	}

	return order.ID, readyIDs, nil
}
