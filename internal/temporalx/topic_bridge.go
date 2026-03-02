package temporalx

import (
	"context"
	"fmt"
	"time"

	"SwishAssignment/internal/messaging"
	"SwishAssignment/internal/models"
	"go.temporal.io/sdk/client"
)

type TopicToTemporalBridge struct {
	Client                 client.Client
	TaskQueue              string
	AllocationWorkflowID   string
	AutoCompleteWorkflowID string
}

func NewTopicToTemporalBridge(c client.Client) *TopicToTemporalBridge {
	return &TopicToTemporalBridge{
		Client:                 c,
		TaskQueue:              AllocationTaskQueue,
		AllocationWorkflowID:   "allocation-event",
		AutoCompleteWorkflowID: "auto-complete-event",
	}
}

func (b *TopicToTemporalBridge) HandleEvent(ctx context.Context, event models.DomainEvent) error {
	workflowFunc := AllocationWorkflow
	baseID := b.AllocationWorkflowID
	if event.Type == models.EventTaskStarted {
		workflowFunc = AutoCompleteTaskWorkflow
		baseID = b.AutoCompleteWorkflowID
	}

	options := client.StartWorkflowOptions{
		ID:        fmt.Sprintf("%s-%s-%d", baseID, event.Type, time.Now().UnixNano()),
		TaskQueue: b.TaskQueue,
	}

	_, err := b.Client.ExecuteWorkflow(ctx, options, workflowFunc, event)
	return err
}

func RunTopicConsumer(ctx context.Context, topic messaging.Topic, bridge *TopicToTemporalBridge) error {
	return topic.Consume(ctx, bridge.HandleEvent)
}
