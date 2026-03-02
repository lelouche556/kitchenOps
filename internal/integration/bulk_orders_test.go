package integration_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"SwishAssignment/internal/db"
	"SwishAssignment/internal/models"
	"gorm.io/gorm"
)

type confirmOrderResponse struct {
	OrderID      uint64   `json:"order_id"`
	ReadyTaskIDs []uint64 `json:"ready_task_ids"`
}

func TestBulkOrdersEndToEnd(t *testing.T) {
	if os.Getenv("INTEGRATION_TEST") != "1" {
		t.Skip("set INTEGRATION_TEST=1 to run integration tests")
	}

	apiBase := envOrDefault("INTEGRATION_API_BASE", "http://localhost:8080")
	pgDSN := envOrDefault("INTEGRATION_PG_DSN", "host=localhost port=5432 user=postgres password=postgres dbname=kangaroo_paw sslmode=disable")

	gdb, err := db.OpenPostgres(pgDSN)
	if err != nil {
		t.Fatalf("postgres connect failed: %v", err)
	}
	if err := ensureBurgerRecipeFastAndDependent(gdb); err != nil {
		t.Fatalf("recipe test setup failed: %v", err)
	}

	const orderCount = 12
	orderIDs, err := createBulkOrders(apiBase, orderCount)
	if err != nil {
		t.Fatalf("bulk order creation failed: %v", err)
	}
	if len(orderIDs) != orderCount {
		t.Fatalf("expected %d order IDs, got %d", orderCount, len(orderIDs))
	}

	if err := waitForTasksCreated(gdb, orderIDs, 4, 20*time.Second); err != nil {
		t.Fatalf("tasks not created as expected: %v", err)
	}

	if err := driveManualStartsUntilComplete(apiBase, gdb, orderIDs, 90*time.Second); err != nil {
		t.Fatalf("orders did not complete: %v", err)
	}

	if err := assertOrdersCompleted(gdb, orderIDs); err != nil {
		t.Fatalf("order completion assertion failed: %v", err)
	}
	if err := assertAssembleStepDependsOnPrep(gdb, orderIDs); err != nil {
		t.Fatalf("dependency timing assertion failed: %v", err)
	}
}

func createBulkOrders(apiBase string, count int) ([]uint64, error) {
	client := &http.Client{Timeout: 5 * time.Second}
	orderIDs := make([]uint64, 0, count)
	var mu sync.Mutex
	var wg sync.WaitGroup
	errCh := make(chan error, count)
	prefix := fmt.Sprintf("it-bulk-%d", time.Now().UnixNano())

	for i := 0; i < count; i++ {
		i := i
		wg.Add(1)
		go func() {
			defer wg.Done()
			reqBody := map[string]any{
				"external_order_id": prefix + "-" + strconv.Itoa(i+1),
				"items": map[string]int{
					"burger_combo": 1,
				},
			}
			body, _ := json.Marshal(reqBody)
			resp, err := client.Post(apiBase+"/api/v1/orders/confirm", "application/json", bytes.NewReader(body))
			if err != nil {
				errCh <- err
				return
			}
			defer resp.Body.Close()
			if resp.StatusCode != http.StatusCreated {
				errCh <- fmt.Errorf("confirm order status=%d", resp.StatusCode)
				return
			}
			var parsed confirmOrderResponse
			if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
				errCh <- err
				return
			}
			if parsed.OrderID == 0 {
				errCh <- fmt.Errorf("empty order id in response")
				return
			}
			mu.Lock()
			orderIDs = append(orderIDs, parsed.OrderID)
			mu.Unlock()
		}()
	}
	wg.Wait()
	close(errCh)
	for err := range errCh {
		if err != nil {
			return nil, err
		}
	}
	return orderIDs, nil
}

func waitForTasksCreated(gdb *gorm.DB, orderIDs []uint64, expectedPerOrder int, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		var rows []struct {
			OrderID uint64
			Cnt     int
		}
		if err := gdb.Table("tasks").
			Select("order_id, COUNT(*) AS cnt").
			Where("order_id IN ?", orderIDs).
			Group("order_id").
			Scan(&rows).Error; err != nil {
			return err
		}
		if len(rows) == len(orderIDs) {
			ok := true
			for _, r := range rows {
				if r.Cnt != expectedPerOrder {
					ok = false
					break
				}
			}
			if ok {
				return nil
			}
		}
		time.Sleep(200 * time.Millisecond)
	}
	return fmt.Errorf("timeout waiting for task creation")
}

func driveManualStartsUntilComplete(apiBase string, gdb *gorm.DB, orderIDs []uint64, timeout time.Duration) error {
	client := &http.Client{Timeout: 5 * time.Second}
	started := map[uint64]struct{}{}
	deadline := time.Now().Add(timeout)

	for time.Now().Before(deadline) {
		tasks, err := listAssignedNotStartedTasks(gdb, orderIDs)
		if err != nil {
			return err
		}
		for _, task := range tasks {
			if _, ok := started[task.ID]; ok {
				continue
			}
			req, _ := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/api/v1/tasks/%d/start", apiBase, task.ID), nil)
			resp, err := client.Do(req)
			if err != nil {
				return err
			}
			_ = resp.Body.Close()
			// 200 is expected; 400 can happen if concurrent loop already advanced task.
			if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusBadRequest {
				return fmt.Errorf("start task %d unexpected status %d", task.ID, resp.StatusCode)
			}
			started[task.ID] = struct{}{}
		}

		allDone, err := areAllOrdersCompleted(gdb, orderIDs)
		if err != nil {
			return err
		}
		if allDone {
			return nil
		}
		time.Sleep(250 * time.Millisecond)
	}
	return fmt.Errorf("timeout waiting for orders to complete")
}

func listAssignedNotStartedTasks(gdb *gorm.DB, orderIDs []uint64) ([]models.TaskModel, error) {
	var tasks []models.TaskModel
	err := gdb.Where("order_id IN ? AND status = ? AND started_at IS NULL", orderIDs, string(models.TaskAssigned)).
		Order("id ASC").
		Find(&tasks).Error
	return tasks, err
}

func areAllOrdersCompleted(gdb *gorm.DB, orderIDs []uint64) (bool, error) {
	var remaining int64
	err := gdb.Model(&models.OrderModel{}).
		Where("id IN ? AND status <> ?", orderIDs, string(models.OrderCompleted)).
		Count(&remaining).Error
	return remaining == 0, err
}

func assertOrdersCompleted(gdb *gorm.DB, orderIDs []uint64) error {
	var remaining int64
	if err := gdb.Model(&models.TaskModel{}).
		Where("order_id IN ? AND status <> ?", orderIDs, string(models.TaskCompleted)).
		Count(&remaining).Error; err != nil {
		return err
	}
	if remaining != 0 {
		return fmt.Errorf("found %d non-completed tasks", remaining)
	}
	allDone, err := areAllOrdersCompleted(gdb, orderIDs)
	if err != nil {
		return err
	}
	if !allDone {
		return fmt.Errorf("some orders are not in COMPLETED state")
	}
	return nil
}

func assertAssembleStepDependsOnPrep(gdb *gorm.DB, orderIDs []uint64) error {
	for _, orderID := range orderIDs {
		var assemble models.TaskModel
		if err := gdb.Where("order_id = ? AND description ILIKE ?", orderID, "%assemble burger%").
			First(&assemble).Error; err != nil {
			return fmt.Errorf("order %d assemble task missing: %w", orderID, err)
		}
		if assemble.StartedAt == nil {
			return fmt.Errorf("order %d assemble task has nil started_at", orderID)
		}

		var prep []models.TaskModel
		if err := gdb.Where("order_id = ? AND description NOT ILIKE ?", orderID, "%assemble burger%").
			Find(&prep).Error; err != nil {
			return fmt.Errorf("order %d prep query failed: %w", orderID, err)
		}
		if len(prep) != 3 {
			return fmt.Errorf("order %d expected 3 prep tasks, got %d", orderID, len(prep))
		}
		latestPrepComplete := time.Time{}
		for _, p := range prep {
			if p.CompletedAt == nil {
				return fmt.Errorf("order %d prep task %d not completed", orderID, p.ID)
			}
			if p.CompletedAt.After(latestPrepComplete) {
				latestPrepComplete = *p.CompletedAt
			}
		}
		if assemble.StartedAt.Before(latestPrepComplete) {
			return fmt.Errorf("order %d assemble started before prep completed", orderID)
		}
	}
	return nil
}

func ensureBurgerRecipeFastAndDependent(gdb *gorm.DB) error {
	if err := gdb.Exec("UPDATE recipe_steps SET estimate_secs = 1 WHERE item_key = ?", "burger_combo").Error; err != nil {
		return err
	}
	deps := []struct {
		ItemKey   string
		StepOrder int
		DependsOn int
	}{
		{"burger_combo", 4, 1},
		{"burger_combo", 4, 2},
		{"burger_combo", 4, 3},
	}
	for _, d := range deps {
		if err := gdb.Exec(`
			INSERT INTO recipe_step_dependencies (item_key, step_order, depends_on_step_order)
			VALUES (?, ?, ?)
			ON CONFLICT (item_key, step_order, depends_on_step_order) DO NOTHING
		`, d.ItemKey, d.StepOrder, d.DependsOn).Error; err != nil {
			return err
		}
	}
	return nil
}

func envOrDefault(key, fallback string) string {
	val := strings.TrimSpace(os.Getenv(key))
	if val == "" {
		return fallback
	}
	return val
}
