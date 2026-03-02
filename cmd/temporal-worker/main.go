package main

import (
	"log"
	"strings"

	"SwishAssignment/internal/application"
	"SwishAssignment/internal/config"
	"SwishAssignment/internal/db"
	"SwishAssignment/internal/orchestrator"
	"SwishAssignment/internal/queue"
	"SwishAssignment/internal/repository"
	"SwishAssignment/internal/temporalx"
	"go.temporal.io/sdk/client"
)

func main() {
	c, err := client.Dial(client.Options{HostPort: config.EnvOrDefault("TEMPORAL_ADDRESS", "localhost:7233")})
	if err != nil {
		log.Fatalf("unable to create temporal client: %v", err)
	}
	defer c.Close()

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
		defer func() { _ = redisQueue.Close() }()
		readyQueue = redisQueue
	}

	kitchen := application.NewKitchenAppService(repo, readyQueue, nil)
	allocationOrchestrator := orchestrator.NewAllocationOrchestrator(kitchen)

	if err := temporalx.StartAllocationWorker(c, allocationOrchestrator); err != nil {
		log.Fatalf("unable to start temporal worker: %v", err)
	}
}
