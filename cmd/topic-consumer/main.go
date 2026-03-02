package main

import (
	"context"
	"log"

	"SwishAssignment/internal/config"
	"SwishAssignment/internal/messaging"
	"SwishAssignment/internal/temporalx"
	"go.temporal.io/sdk/client"
)

func main() {
	ctx := context.Background()

	temporalClient, err := client.Dial(client.Options{HostPort: config.EnvOrDefault("TEMPORAL_ADDRESS", "localhost:7233")})
	if err != nil {
		log.Fatalf("unable to create temporal client: %v", err)
	}
	defer temporalClient.Close()

	kafkaTopic, err := messaging.NewKafkaTopic(messaging.KafkaConfig{
		Brokers:       config.EnvCSV("KAFKA_BROKERS", "localhost:9092"),
		Topic:         config.EnvOrDefault("KAFKA_TOPIC", "kitchen.domain.events"),
		GroupID:       config.EnvOrDefault("KAFKA_GROUP_ID", "kitchen-topic-consumer"),
		Partition:     config.EnvInt("KAFKA_PARTITION", -1),
		FromBeginning: config.EnvBool("KAFKA_FROM_BEGINNING", false),
	})
	if err != nil {
		log.Fatalf("unable to create kafka topic client: %v", err)
	}
	defer func() {
		if err := kafkaTopic.Close(); err != nil {
			log.Printf("kafka close error: %v", err)
		}
	}()

	bridge := temporalx.NewTopicToTemporalBridge(temporalClient)
	if err := temporalx.RunTopicConsumer(ctx, kafkaTopic, bridge); err != nil {
		log.Fatalf("topic consumer failed: %v", err)
	}
}
