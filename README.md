# Kangaroo Paw - Real-Time Kitchen Scheduling Engine

Kangaroo Paw is a backend scheduling system for kitchen execution. It explodes orders into atomic tasks, applies dependency-aware readiness, assigns tasks under staff/counter/machine constraints, and drives execution using Kafka + Temporal.

## Current Capabilities

- REST API with router/controller/service/repository layering
- PostgreSQL as system-of-record (orders, tasks, dependencies, counters, machines, staff, domain events)
- Redis sorted set ready queue (`tasks:pending`) for fast priority lookup
- Kafka topic for domain-event transport (`kitchen.domain.events`)
- Temporal workflows for asynchronous allocation and auto-completion timing
- Outbox dispatcher for reliable event publishing with retry/backoff
- Assignment timeout requeue for stale `ASSIGNED` tasks
- Recipe and dependency model fully DB-driven (no hardcoded recipe map)

## End-to-End Flow

1. `POST /api/v1/orders/confirm`
2. `recipe_steps` are fetched from Postgres and one task per step is created.
3. Task dependencies are built from:
   - explicit `recipe_step_dependencies` (preferred), or
   - fallback inference (assembly fan-in + single-capacity machine serialization).
4. Tasks with `pending_deps = 0` are pushed to Redis sorted set and corresponding `TASK_READY` is inserted into outbox (`domain_events`).
5. Outbox dispatcher publishes events to Kafka with retries and marks events as published.
6. Topic consumer reads Kafka and starts Temporal workflows.
   - `TASK_READY`, `TASK_COMPLETED`, `TASK_REQUEUED`, etc. -> Allocation workflow
   - `TASK_STARTED` -> Auto-complete workflow (sleep `estimate_secs`, then complete)
7. Allocation transaction (with locks) selects feasible task/staff/machine and assigns.
8. Manual start API marks task `STARTED`.
9. Temporal auto-complete marks task `COMPLETED`, releases capacities, updates staff metrics, unlocks child tasks, emits `TASK_READY`.

## Lifecycle States

### Order status

- `CONFIRMED`
- `IN_PROGRESS` (when first task is assigned)
- `PART_COMPLETED` (when at least one task is completed)
- `COMPLETED` (when all tasks are completed)

### Task status

- `UNASSIGNED`
- `ASSIGNED`
- `STARTED`
- `COMPLETED`

## Key Scheduling Rules

- A task is assignable only when:
  - `pending_deps = 0`
  - counter has available capacity
  - machine is up and has available capacity
  - eligible staff exists (skill match, in shift, not on break, below max parallel)
- Staff pick score:
  - `weightLoad*active_tasks + weightUtil*utilization - weightEff*efficiency_multiplier`
- Aging priority score in ready queue:
  - `basePriority + agingFactor * waitTime` (stored as negative for min-pop)
- Stale assignment protection:
  - tasks left `ASSIGNED` without `STARTED` beyond TTL are atomically requeued with boosted priority and `TASK_REQUEUED` event.

## Repository Layout

```text
cmd/
  api/
  topic-consumer/
  temporal-worker/
  kafka-tail/

internal/
  api/
    router/
    controllers/
  application/
  repository/
  models/
  messaging/
  temporalx/
  orchestrator/
  queue/
  outbox/
  db/
  config/

db/
  schema.sql
  migrations/

docker/
  postgres/init/
```

## Services and Ports

- API: `http://localhost:8080`
- PostgreSQL: `localhost:5432`
- Redis: `localhost:6379`
- Kafka broker (Redpanda): `localhost:9092`
- Kafka Console: `http://localhost:8081`
- Temporal frontend: `localhost:7233`
- Temporal UI: `http://localhost:8088`

## Database Defaults

- Host: `localhost`
- Port: `5432`
- DB: `kangaroo_paw`
- User: `postgres`
- Password: `postgres`

JDBC:

```text
jdbc:postgresql://localhost:5432/kangaroo_paw
```

## Run Locally (Docker)

```bash
docker compose up --build -d
```

Check status:

```bash
docker compose ps
curl -s http://localhost:8080/health
```

Expected health response:

```json
{"status":"ok"}
```

## Migrations for Existing DB Volumes

If you already had data volume before recent refactors, apply these once:

```bash
docker exec -i kangaroo-postgres psql -U postgres -d kangaroo_paw < db/migrations/20260302_counter_fk_normalization.sql
docker exec -i kangaroo-postgres psql -U postgres -d kangaroo_paw < db/migrations/20260302_machine_fk_normalization.sql
docker exec -i kangaroo-postgres psql -U postgres -d kangaroo_paw < db/migrations/20260302_outbox_reliability.sql
docker exec -i kangaroo-postgres psql -U postgres -d kangaroo_paw < db/migrations/20260302_recipe_step_dependencies.sql
docker exec -i kangaroo-postgres psql -U postgres -d kangaroo_paw < db/migrations/20260302_seed_recipe_step_dependencies.sql
```

## API Reference

- `GET /health`
- `POST /api/v1/orders/confirm`
- `POST /api/v1/allocator/run-once`
- `POST /api/v1/tasks/{taskID}/start`
- `POST /api/v1/tasks/{taskID}/complete`
- `GET /api/v1/tasks/{taskID}`

Example order confirm:

```bash
curl -X POST http://localhost:8080/api/v1/orders/confirm \
  -H 'content-type: application/json' \
  -d '{"external_order_id":"ord-1001","items":{"burger_combo":1,"cold_coffee":2,"wrap":1}}'
```

Manual start (completion is automated by Temporal timer):

```bash
curl -X POST http://localhost:8080/api/v1/tasks/5/start
```

## Kafka / Redis / DB Inspection

Kafka messages:

- Web UI: `http://localhost:8081` -> Topics -> `kitchen.domain.events`
- CLI:

```bash
docker exec -it kangaroo-kafka rpk topic consume kitchen.domain.events -f '%o %k %v\n'
```

Redis queue:

```bash
docker exec -it kangaroo-redis redis-cli ZCARD tasks:pending
docker exec -it kangaroo-redis redis-cli ZRANGE tasks:pending 0 50 WITHSCORES
```

Postgres quick checks:

```bash
docker exec -it kangaroo-postgres psql -U postgres -d kangaroo_paw -c '\dt'
docker exec -it kangaroo-postgres psql -U postgres -d kangaroo_paw -c 'SELECT id,status,created_at FROM orders ORDER BY id DESC LIMIT 10;'
docker exec -it kangaroo-postgres psql -U postgres -d kangaroo_paw -c 'SELECT id,order_id,status,pending_deps,assigned_staff_id,assigned_machine_id,assigned_at,started_at,completed_at FROM tasks ORDER BY id DESC LIMIT 30;'
docker exec -it kangaroo-postgres psql -U postgres -d kangaroo_paw -c 'SELECT id,event_type,attempts,published_at,next_retry_at,last_error FROM domain_events ORDER BY id DESC LIMIT 30;'
```

## Testing

### Unit + package tests

```bash
go test ./...
```

### Integration test: bulk concurrent order flow

Added test:

- `internal/integration/bulk_orders_test.go`

What it validates:

- concurrent creation of multiple `burger_combo` orders
- task creation and dependency gating
- repeated manual task starts via API
- auto-completion via Temporal timer
- final `COMPLETED` order/task states
- assemble-step starts only after all prep steps complete

Run it against running local stack:

```bash
INTEGRATION_TEST=1 go test ./internal/integration -run TestBulkOrdersEndToEnd -v
```

Optional overrides:

- `INTEGRATION_API_BASE` (default `http://localhost:8080`)
- `INTEGRATION_PG_DSN` (default local postgres dsn)

## Notes

- Recipe behavior is data-driven from DB; add/edit steps using `recipe_steps` and dependencies using `recipe_step_dependencies`.
- For best observability, keep Kafka Console and Temporal UI open while running load/integration tests.
