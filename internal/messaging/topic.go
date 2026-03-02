package messaging

import (
	"context"

	"SwishAssignment/internal/models"
)

type Topic interface {
	Publish(ctx context.Context, event models.DomainEvent) error
	Consume(ctx context.Context, handler func(context.Context, models.DomainEvent) error) error
}

type InMemoryTopic struct {
	events chan models.DomainEvent
}

func NewInMemoryTopic(buffer int) *InMemoryTopic {
	if buffer <= 0 {
		buffer = 256
	}
	return &InMemoryTopic{events: make(chan models.DomainEvent, buffer)}
}

func (t *InMemoryTopic) Publish(ctx context.Context, event models.DomainEvent) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case t.events <- event:
		return nil
	}
}

func (t *InMemoryTopic) Consume(ctx context.Context, handler func(context.Context, models.DomainEvent) error) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case evt := <-t.events:
			if err := handler(ctx, evt); err != nil {
				return err
			}
		}
	}
}
