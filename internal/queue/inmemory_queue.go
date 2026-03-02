package queue

import (
	"context"
	"sort"
	"sync"
)

type inMemoryItem struct {
	TaskID string
	Score  float64
}

type InMemoryReadyQueue struct {
	mu    sync.Mutex
	items map[string]float64
}

func NewInMemoryReadyQueue() *InMemoryReadyQueue {
	return &InMemoryReadyQueue{items: map[string]float64{}}
}

func (q *InMemoryReadyQueue) Enqueue(_ context.Context, taskID string, score float64) error {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.items[taskID] = score
	return nil
}

func (q *InMemoryReadyQueue) PopTopN(_ context.Context, n int) ([]string, error) {
	q.mu.Lock()
	defer q.mu.Unlock()
	if n <= 0 {
		return nil, nil
	}

	list := make([]inMemoryItem, 0, len(q.items))
	for id, score := range q.items {
		list = append(list, inMemoryItem{TaskID: id, Score: score})
	}
	sort.Slice(list, func(i, j int) bool {
		if list[i].Score == list[j].Score {
			return list[i].TaskID < list[j].TaskID
		}
		return list[i].Score < list[j].Score
	})

	limit := n
	if len(list) < limit {
		limit = len(list)
	}
	out := make([]string, 0, limit)
	for i := 0; i < limit; i++ {
		out = append(out, list[i].TaskID)
		delete(q.items, list[i].TaskID)
	}
	return out, nil
}

func (q *InMemoryReadyQueue) Len(_ context.Context) (int64, error) {
	q.mu.Lock()
	defer q.mu.Unlock()
	return int64(len(q.items)), nil
}

func (q *InMemoryReadyQueue) Close() error { return nil }
