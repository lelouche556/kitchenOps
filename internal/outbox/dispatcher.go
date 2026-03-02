package outbox

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"SwishAssignment/internal/models"
	"SwishAssignment/internal/repository"
	"gorm.io/gorm"
)

type Publisher interface {
	Publish(ctx context.Context, event models.DomainEvent) error
}

type Dispatcher struct {
	repo      *repository.KitchenRepository
	publisher Publisher
	batchSize int
	interval  time.Duration
}

func NewDispatcher(repo *repository.KitchenRepository, publisher Publisher) *Dispatcher {
	return &Dispatcher{
		repo:      repo,
		publisher: publisher,
		batchSize: 100,
		interval:  300 * time.Millisecond,
	}
}

func (d *Dispatcher) Run(ctx context.Context) {
	ticker := time.NewTicker(d.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := d.DispatchOnce(ctx); err != nil && ctx.Err() == nil {
				log.Printf("outbox dispatch error: %v", err)
			}
		}
	}
}

func (d *Dispatcher) DispatchOnce(ctx context.Context) error {
	now := time.Now()
	return d.repo.DB().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		rows, err := d.repo.LockPendingOutboxEvents(ctx, tx, now, d.batchSize)
		if err != nil {
			return err
		}
		for _, row := range rows {
			var evt models.DomainEvent
			if err := json.Unmarshal(row.Payload, &evt); err != nil {
				retryAt := time.Now().Add(backoff(row.Attempts + 1))
				_ = d.repo.MarkOutboxRetry(ctx, tx, row.ID, "invalid event payload", retryAt)
				continue
			}

			if err := d.publisher.Publish(ctx, evt); err != nil {
				retryAt := time.Now().Add(backoff(row.Attempts + 1))
				_ = d.repo.MarkOutboxRetry(ctx, tx, row.ID, err.Error(), retryAt)
				continue
			}

			if err := d.repo.MarkOutboxPublished(ctx, tx, row.ID, time.Now()); err != nil {
				return err
			}
		}
		return nil
	})
}

func backoff(attempt int) time.Duration {
	if attempt < 1 {
		attempt = 1
	}
	d := time.Duration(attempt*attempt) * time.Second
	if d > 5*time.Minute {
		return 5 * time.Minute
	}
	return d
}
