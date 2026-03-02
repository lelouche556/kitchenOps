package repository

import (
	"context"
	"encoding/json"
	"time"

	"SwishAssignment/internal/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func (r *KitchenRepository) InsertDomainEvent(ctx context.Context, tx *gorm.DB, aggregateType string, aggregateID string, event models.DomainEvent) error {
	payload, err := json.Marshal(event)
	if err != nil {
		return err
	}
	rec := models.DomainEventModel{
		AggregateType: aggregateType,
		AggregateID:   aggregateID,
		EventType:     string(event.Type),
		Payload:       payload,
	}
	return tx.WithContext(ctx).Create(&rec).Error
}

func (r *KitchenRepository) InsertDomainEventNoTx(ctx context.Context, aggregateType string, aggregateID string, event models.DomainEvent) error {
	payload, err := json.Marshal(event)
	if err != nil {
		return err
	}
	rec := models.DomainEventModel{
		AggregateType: aggregateType,
		AggregateID:   aggregateID,
		EventType:     string(event.Type),
		Payload:       payload,
	}
	return r.db.WithContext(ctx).Create(&rec).Error
}

func (r *KitchenRepository) LockPendingOutboxEvents(ctx context.Context, tx *gorm.DB, now time.Time, limit int) ([]models.DomainEventModel, error) {
	if limit <= 0 {
		limit = 100
	}
	var rows []models.DomainEventModel
	err := tx.WithContext(ctx).
		Clauses(clause.Locking{Strength: "UPDATE", Options: "SKIP LOCKED"}).
		Where("published_at IS NULL AND (next_retry_at IS NULL OR next_retry_at <= ?)", now).
		Order("id ASC").
		Limit(limit).
		Find(&rows).Error
	return rows, err
}

func (r *KitchenRepository) MarkOutboxPublished(ctx context.Context, tx *gorm.DB, eventID uint64, publishedAt time.Time) error {
	return tx.WithContext(ctx).Model(&models.DomainEventModel{}).
		Where("id = ?", eventID).
		Updates(map[string]any{
			"published_at":  publishedAt,
			"last_error":    nil,
			"next_retry_at": nil,
		}).Error
}

func (r *KitchenRepository) MarkOutboxRetry(ctx context.Context, tx *gorm.DB, eventID uint64, errText string, nextRetryAt time.Time) error {
	return tx.WithContext(ctx).Model(&models.DomainEventModel{}).
		Where("id = ?", eventID).
		Updates(map[string]any{
			"attempts":      gorm.Expr("attempts + 1"),
			"last_error":    errText,
			"next_retry_at": nextRetryAt,
		}).Error
}
