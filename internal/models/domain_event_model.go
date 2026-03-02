package models

import "time"

type DomainEventModel struct {
	ID            uint64     `gorm:"column:id;primaryKey"`
	AggregateType string     `gorm:"column:aggregate_type"`
	AggregateID   string     `gorm:"column:aggregate_id"`
	EventType     string     `gorm:"column:event_type"`
	Payload       []byte     `gorm:"column:payload;type:jsonb"`
	Attempts      int        `gorm:"column:attempts"`
	NextRetryAt   *time.Time `gorm:"column:next_retry_at"`
	PublishedAt   *time.Time `gorm:"column:published_at"`
	LastError     *string    `gorm:"column:last_error"`
	CreatedAt     time.Time  `gorm:"column:created_at"`
}

func (DomainEventModel) TableName() string { return "domain_events" }
