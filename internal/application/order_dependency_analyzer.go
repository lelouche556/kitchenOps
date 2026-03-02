package application

import (
	"sort"
	"strconv"
	"strings"

	"SwishAssignment/internal/models"
)

type taskDepPair struct {
	TaskIdx int
	DepIdx  int
}

func inferRecipeTaskDeps(steps []models.RecipeStepModel, stepToTaskIdx map[int]int, explicitDeps map[int][]int, machineCapByID map[uint64]int) []taskDepPair {
	edges := make([]taskDepPair, 0)
	seen := make(map[string]struct{})

	addEdge := func(taskStep, depStep int) {
		taskIdx, okTask := stepToTaskIdx[taskStep]
		depIdx, okDep := stepToTaskIdx[depStep]
		if !okTask || !okDep || taskIdx == depIdx {
			return
		}
		key := strings.Join([]string{strconv.Itoa(taskIdx), strconv.Itoa(depIdx)}, ":")
		if _, ok := seen[key]; ok {
			return
		}
		seen[key] = struct{}{}
		edges = append(edges, taskDepPair{TaskIdx: taskIdx, DepIdx: depIdx})
	}

	// Explicit DB dependency definition takes precedence.
	if len(explicitDeps) > 0 {
		for taskStep, deps := range explicitDeps {
			for _, depStep := range deps {
				addEdge(taskStep, depStep)
			}
		}
		return edges
	}

	// Fallback analyzer:
	// 1) Assembly/finalization steps depend on all prior non-assembly steps.
	// 2) Steps sharing a single-capacity machine are serialized.
	stepOrders := make([]int, 0, len(steps))
	stepByOrder := make(map[int]models.RecipeStepModel, len(steps))
	for _, st := range steps {
		stepOrders = append(stepOrders, st.StepOrder)
		stepByOrder[st.StepOrder] = st
	}
	sort.Ints(stepOrders)

	assemblySteps := make([]int, 0)
	for _, so := range stepOrders {
		if isAssemblyLike(stepByOrder[so].Description) {
			assemblySteps = append(assemblySteps, so)
		}
	}
	for _, asm := range assemblySteps {
		for _, so := range stepOrders {
			if so >= asm {
				continue
			}
			if isAssemblyLike(stepByOrder[so].Description) {
				continue
			}
			addEdge(asm, so)
		}
	}

	previousByMachine := make(map[uint64]int)
	for _, so := range stepOrders {
		step := stepByOrder[so]
		if step.MachineID == nil {
			continue
		}
		mid := *step.MachineID
		if machineCapByID[mid] <= 1 {
			if prev, ok := previousByMachine[mid]; ok && prev != so {
				addEdge(so, prev)
			}
		}
		previousByMachine[mid] = so
	}

	return edges
}

func isAssemblyLike(desc string) bool {
	d := strings.ToLower(strings.TrimSpace(desc))
	if d == "" {
		return false
	}
	keywords := []string{
		"assemble",
		"assembly",
		"final",
		"finish",
		"pack",
		"plating",
		"plate",
		"serve",
	}
	for _, k := range keywords {
		if strings.Contains(d, k) {
			return true
		}
	}
	return false
}
