package temporalx

import (
	"SwishAssignment/internal/orchestrator"
	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
)

func StartAllocationWorker(c client.Client, allocator *orchestrator.AllocationOrchestrator) error {
	w := worker.New(c, AllocationTaskQueue, worker.Options{})
	w.RegisterWorkflow(AllocationWorkflow)
	w.RegisterWorkflow(AutoCompleteTaskWorkflow)
	activities := &AllocationActivities{
		Orchestrator: allocator,
		Kitchen:      allocator.Kitchen,
	}
	w.RegisterActivityWithOptions(activities.RunAllocationActivity, activity.RegisterOptions{Name: "RunAllocationActivity"})
	w.RegisterActivityWithOptions(activities.PrepareAutoCompleteTaskActivity, activity.RegisterOptions{Name: "PrepareAutoCompleteTaskActivity"})
	w.RegisterActivityWithOptions(activities.CompleteTaskByIDActivity, activity.RegisterOptions{Name: "CompleteTaskByIDActivity"})
	return w.Run(worker.InterruptCh())
}
