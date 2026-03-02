package orchestrator

import (
	"context"

	"SwishAssignment/internal/application"
	"SwishAssignment/internal/models"
)

type AllocationOrchestrator struct {
	Kitchen *application.KitchenAppService
}

func NewAllocationOrchestrator(kitchen *application.KitchenAppService) *AllocationOrchestrator {
	return &AllocationOrchestrator{Kitchen: kitchen}
}

func (o *AllocationOrchestrator) HandleEvent(ctx context.Context, event models.DomainEvent) error {
	switch event.Type {
	case models.EventTaskCreated, models.EventTaskReady, models.EventTaskAssigned, models.EventTaskStarted, models.EventTaskCompleted, models.EventTaskRequeued:
		_, err := o.Kitchen.AllocateOnce(ctx)
		return err
	}
	return nil
}
