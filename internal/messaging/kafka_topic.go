package messaging

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"SwishAssignment/internal/models"
	"github.com/segmentio/kafka-go"
)

type KafkaConfig struct {
	Brokers       []string
	Topic         string
	GroupID       string
	Partition     int
	FromBeginning bool
}

type KafkaTopic struct {
	writer *kafka.Writer
	reader *kafka.Reader
}

func NewKafkaTopic(cfg KafkaConfig) (*KafkaTopic, error) {
	if len(cfg.Brokers) == 0 {
		return nil, fmt.Errorf("at least one kafka broker is required")
	}
	if strings.TrimSpace(cfg.Topic) == "" {
		return nil, fmt.Errorf("kafka topic is required")
	}
	writer := &kafka.Writer{
		Addr:         kafka.TCP(cfg.Brokers...),
		Topic:        cfg.Topic,
		Balancer:     &kafka.LeastBytes{},
		RequiredAcks: kafka.RequireAll,
		Async:        false,
	}

	startOffset := kafka.LastOffset
	if cfg.FromBeginning {
		startOffset = kafka.FirstOffset
	}

	var reader *kafka.Reader
	if strings.TrimSpace(cfg.GroupID) != "" || cfg.Partition >= 0 {
		readerCfg := kafka.ReaderConfig{
			Brokers:     cfg.Brokers,
			Topic:       cfg.Topic,
			MinBytes:    1,
			MaxBytes:    10e6,
			StartOffset: startOffset,
			MaxWait:     500 * time.Millisecond,
		}
		if strings.TrimSpace(cfg.GroupID) != "" {
			readerCfg.GroupID = cfg.GroupID
		} else {
			readerCfg.Partition = cfg.Partition
		}
		reader = kafka.NewReader(readerCfg)
	}

	return &KafkaTopic{writer: writer, reader: reader}, nil
}

func (t *KafkaTopic) Publish(ctx context.Context, event models.DomainEvent) error {
	body, err := json.Marshal(event)
	if err != nil {
		return err
	}
	msg := kafka.Message{
		Key:   []byte(string(event.Type)),
		Value: body,
		Time:  time.Now(),
	}
	var lastErr error
	for attempt := 1; attempt <= 3; attempt++ {
		if err := t.writer.WriteMessages(ctx, msg); err == nil {
			return nil
		} else {
			lastErr = err
			if ctx.Err() != nil {
				return ctx.Err()
			}
			time.Sleep(time.Duration(attempt*150) * time.Millisecond)
		}
	}
	return fmt.Errorf("publish failed after retries: %w", lastErr)
}

func (t *KafkaTopic) Consume(ctx context.Context, handler func(context.Context, models.DomainEvent) error) error {
	if t.reader == nil {
		return fmt.Errorf("kafka reader is not configured")
	}
	for {
		msg, err := t.reader.ReadMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return nil
			}
			return err
		}

		var evt models.DomainEvent
		if err := json.Unmarshal(msg.Value, &evt); err != nil {
			log.Printf("kafka consume: invalid payload skipped: %v", err)
			continue
		}

		var handled bool
		for attempt := 1; attempt <= 3; attempt++ {
			if err := handler(ctx, evt); err != nil {
				log.Printf("kafka consume: handler failed (attempt=%d type=%s payload=%s): %v", attempt, evt.Type, evt.Payload, err)
				if ctx.Err() != nil {
					return nil
				}
				time.Sleep(time.Duration(attempt*200) * time.Millisecond)
				continue
			}
			handled = true
			break
		}
		if !handled {
			log.Printf("kafka consume: dropping event after retries type=%s payload=%s", evt.Type, evt.Payload)
		}
	}
}

func (t *KafkaTopic) Close() error {
	var firstErr error
	if t.reader != nil {
		if err := t.reader.Close(); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	if t.writer != nil {
		if err := t.writer.Close(); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	return firstErr
}
