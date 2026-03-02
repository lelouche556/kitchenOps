package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"SwishAssignment/internal/config"
	"SwishAssignment/internal/models"
	"github.com/segmentio/kafka-go"
)

func main() {
	brokers := config.EnvCSV("KAFKA_BROKERS", "localhost:9092")
	topic := config.EnvOrDefault("KAFKA_TOPIC", "kitchen.domain.events")
	fromBeginning := config.EnvBool("KAFKA_FROM_BEGINNING", true)

	startOffset := kafka.LastOffset
	if fromBeginning {
		startOffset = kafka.FirstOffset
	}

	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:     brokers,
		Topic:       topic,
		MinBytes:    1,
		MaxBytes:    10e6,
		StartOffset: startOffset,
		MaxWait:     500 * time.Millisecond,
	})
	defer func() {
		if err := reader.Close(); err != nil {
			log.Printf("reader close error: %v", err)
		}
	}()

	ctx := context.Background()
	for {
		msg, err := reader.ReadMessage(ctx)
		if err != nil {
			log.Fatalf("read error: %v", err)
		}

		var event models.DomainEvent
		if err := json.Unmarshal(msg.Value, &event); err != nil {
			fmt.Printf("offset=%d key=%s raw=%s\n", msg.Offset, string(msg.Key), string(msg.Value))
			continue
		}

		fmt.Printf("offset=%d partition=%d key=%s type=%s ts=%d payload=%s\n", msg.Offset, msg.Partition, string(msg.Key), event.Type, event.Timestamp, event.Payload)
	}
}
