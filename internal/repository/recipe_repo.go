package repository

import (
	"context"
	"sort"

	"SwishAssignment/internal/models"
)

func (r *KitchenRepository) GetActiveRecipeStepsByItems(ctx context.Context, itemKeys []string) (map[string][]models.RecipeStepModel, error) {
	if len(itemKeys) == 0 {
		return map[string][]models.RecipeStepModel{}, nil
	}

	keys := make([]string, 0, len(itemKeys))
	seen := make(map[string]struct{}, len(itemKeys))
	for _, key := range itemKeys {
		if key == "" {
			continue
		}
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		keys = append(keys, key)
	}
	sort.Strings(keys)

	var rows []models.RecipeStepModel
	if err := r.db.WithContext(ctx).
		Where("item_key IN ? AND is_active = TRUE", keys).
		Order("item_key ASC, step_order ASC").
		Find(&rows).Error; err != nil {
		return nil, err
	}

	out := make(map[string][]models.RecipeStepModel, len(keys))
	for _, key := range keys {
		out[key] = []models.RecipeStepModel{}
	}
	for _, row := range rows {
		out[row.ItemKey] = append(out[row.ItemKey], row)
	}
	return out, nil
}

func (r *KitchenRepository) GetRecipeStepDependenciesByItems(ctx context.Context, itemKeys []string) (map[string]map[int][]int, error) {
	if len(itemKeys) == 0 {
		return map[string]map[int][]int{}, nil
	}

	keys := make([]string, 0, len(itemKeys))
	seen := make(map[string]struct{}, len(itemKeys))
	for _, key := range itemKeys {
		if key == "" {
			continue
		}
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		keys = append(keys, key)
	}
	sort.Strings(keys)

	var rows []models.RecipeStepDependencyModel
	if err := r.db.WithContext(ctx).
		Where("item_key IN ?", keys).
		Order("item_key ASC, step_order ASC, depends_on_step_order ASC").
		Find(&rows).Error; err != nil {
		return nil, err
	}

	out := make(map[string]map[int][]int, len(keys))
	for _, key := range keys {
		out[key] = map[int][]int{}
	}
	for _, row := range rows {
		if _, ok := out[row.ItemKey][row.StepOrder]; !ok {
			out[row.ItemKey][row.StepOrder] = []int{}
		}
		out[row.ItemKey][row.StepOrder] = append(out[row.ItemKey][row.StepOrder], row.DependsOnStepOrder)
	}
	return out, nil
}

func (r *KitchenRepository) GetMachineCapacitiesByIDs(ctx context.Context, machineIDs []uint64) (map[uint64]int, error) {
	if len(machineIDs) == 0 {
		return map[uint64]int{}, nil
	}

	ids := make([]uint64, 0, len(machineIDs))
	seen := make(map[uint64]struct{}, len(machineIDs))
	for _, id := range machineIDs {
		if id == 0 {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		ids = append(ids, id)
	}
	if len(ids) == 0 {
		return map[uint64]int{}, nil
	}

	var rows []models.MachineModel
	if err := r.db.WithContext(ctx).Where("id IN ?", ids).Find(&rows).Error; err != nil {
		return nil, err
	}

	out := make(map[uint64]int, len(rows))
	for _, row := range rows {
		out[row.ID] = row.Capacity
	}
	return out, nil
}
