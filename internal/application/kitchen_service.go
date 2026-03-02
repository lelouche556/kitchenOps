package application

import (
	"context"
	"sync"
	"time"

	"SwishAssignment/internal/models"
	"SwishAssignment/internal/queue"
	"SwishAssignment/internal/repository"
)

type EventPublisher interface {
	Publish(ctx context.Context, event models.DomainEvent) error
}

type ConfirmOrderRequest struct {
	ExternalOrderID string         `json:"external_order_id"`
	Items           map[string]int `json:"items"`
}

type KitchenAppService struct {
	repo        *repository.KitchenRepository
	readyQueue  queue.ReadyQueue
	publisher   EventPublisher
	agingFactor float64
	weightLoad  float64
	weightUtil  float64
	weightEff   float64
	assignTTL   time.Duration
}

var (
	kitchenAppOnce sync.Once
	kitchenAppInst *KitchenAppService
)

func NewKitchenAppService(repo *repository.KitchenRepository, readyQueue queue.ReadyQueue, publisher EventPublisher) *KitchenAppService {
	kitchenAppOnce.Do(func() {
		kitchenAppInst = &KitchenAppService{
			repo:        repo,
			readyQueue:  readyQueue,
			publisher:   publisher,
			agingFactor: 0.14,
			weightLoad:  1.2,
			weightUtil:  0.9,
			weightEff:   1.1,
			assignTTL:   45 * time.Second,
		}
	})
	return kitchenAppInst
}

func (s *KitchenAppService) publish(ctx context.Context, event models.DomainEvent) error {
	if s.publisher == nil {
		return nil
	}
	return s.publisher.Publish(ctx, event)
}
