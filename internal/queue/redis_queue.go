package queue

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisConfig struct {
	Addr     string
	Password string
	DB       int
	Key      string
}

type RedisReadyQueue struct {
	client *redis.Client
	key    string
}

func NewRedisReadyQueue(cfg RedisConfig) (*RedisReadyQueue, error) {
	if strings.TrimSpace(cfg.Addr) == "" {
		return nil, fmt.Errorf("redis address is required")
	}
	if strings.TrimSpace(cfg.Key) == "" {
		cfg.Key = "tasks:pending"
	}

	client := redis.NewClient(&redis.Options{
		Addr:         cfg.Addr,
		Password:     cfg.Password,
		DB:           cfg.DB,
		DialTimeout:  2 * time.Second,
		ReadTimeout:  2 * time.Second,
		WriteTimeout: 2 * time.Second,
	})
	if err := client.Ping(context.Background()).Err(); err != nil {
		return nil, err
	}

	return &RedisReadyQueue{client: client, key: cfg.Key}, nil
}

func (q *RedisReadyQueue) Enqueue(ctx context.Context, taskID string, score float64) error {
	return q.client.ZAdd(ctx, q.key, redis.Z{Score: score, Member: taskID}).Err()
}

func (q *RedisReadyQueue) PopTopN(ctx context.Context, n int) ([]string, error) {
	if n <= 0 {
		return nil, nil
	}
	items, err := q.client.ZPopMin(ctx, q.key, int64(n)).Result()
	if err != nil {
		return nil, err
	}
	out := make([]string, 0, len(items))
	for _, item := range items {
		if taskID, ok := item.Member.(string); ok {
			out = append(out, taskID)
		}
	}
	return out, nil
}

func (q *RedisReadyQueue) Len(ctx context.Context) (int64, error) {
	return q.client.ZCard(ctx, q.key).Result()
}

func (q *RedisReadyQueue) Close() error {
	return q.client.Close()
}
