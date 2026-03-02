package queue

import "context"

type ReadyQueue interface {
	Enqueue(ctx context.Context, taskID string, score float64) error
	PopTopN(ctx context.Context, n int) ([]string, error)
	Len(ctx context.Context) (int64, error)
	Close() error
}
