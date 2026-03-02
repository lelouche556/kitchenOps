package main

import (
	"context"
	"log"
	"net/http"
	"strings"

	"SwishAssignment/internal/api/controllers"
	apirouter "SwishAssignment/internal/api/router"
	"SwishAssignment/internal/application"
	"SwishAssignment/internal/config"
	"SwishAssignment/internal/db"
	"SwishAssignment/internal/messaging"
	"SwishAssignment/internal/outbox"
	"SwishAssignment/internal/queue"
	"SwishAssignment/internal/repository"
)

func main() {
	pgDSN := config.EnvOrDefault("PG_DSN", "host=localhost port=5432 user=postgres password=postgres dbname=kangaroo_paw sslmode=disable")
	gormDB, err := db.OpenPostgres(pgDSN)
	if err != nil {
		log.Fatalf("postgres connection failed: %v", err)
	}

	repo := repository.NewKitchenRepository(gormDB)

	var readyQueue queue.ReadyQueue = queue.NewInMemoryReadyQueue()
	if redisAddr := strings.TrimSpace(config.EnvOrDefault("REDIS_ADDR", "")); redisAddr != "" {
		redisQueue, err := queue.NewRedisReadyQueue(queue.RedisConfig{
			Addr:     redisAddr,
			Password: config.EnvOrDefault("REDIS_PASSWORD", ""),
			DB:       config.EnvInt("REDIS_DB", 0),
			Key:      config.EnvOrDefault("REDIS_PENDING_KEY", "tasks:pending"),
		})
		if err != nil {
			log.Fatalf("redis queue init failed: %v", err)
		}
		defer func() {
			_ = redisQueue.Close()
		}()
		readyQueue = redisQueue
	}

	var publisher application.EventPublisher
	if brokers := strings.TrimSpace(config.EnvOrDefault("KAFKA_BROKERS", "")); brokers != "" {
		topic, err := messaging.NewKafkaTopic(messaging.KafkaConfig{
			Brokers:       strings.Split(brokers, ","),
			Topic:         config.EnvOrDefault("KAFKA_TOPIC", "kitchen.domain.events"),
			GroupID:       "",
			Partition:     -1,
			FromBeginning: false,
		})
		if err != nil {
			log.Fatalf("kafka topic init failed: %v", err)
		}
		defer func() {
			_ = topic.Close()
		}()
		publisher = topic
	}

	service := application.NewKitchenAppService(repo, readyQueue, publisher)
	if publisher != nil {
		dispatcher := outbox.NewDispatcher(repo, publisher)
		go dispatcher.Run(context.Background())
	}
	controller := controllers.NewKitchenController(service)
	h := apirouter.New(controller)

	addr := config.EnvOrDefault("HTTP_ADDR", ":8080")
	log.Printf("api listening on %s", addr)
	if err := http.ListenAndServe(addr, h); err != nil {
		log.Fatalf("http server failed: %v", err)
	}
}
