# Implementation Status vs Plan

## Completed
- Postgres schema draft for core entities and event/outbox tables: `db/schema.sql`
- Atomic task generation from recipe steps with dependency gating.
- Staff scoring and capacity-aware assignment logic.
- Transactional task assignment path in Postgres repository using locking clauses (`FOR UPDATE` / `SKIP LOCKED` semantics via GORM).
- Event publishing hooks from domain transitions (Kafka publisher implementation available).
- Event-driven allocation orchestrator.
- Temporal workflow + activity for asynchronous allocation processing.
- Topic consumer bridge that starts Temporal workflow executions from incoming events.
- REST API with router/controller/service/repository layers.
- Redis sorted-set ready queue implementation (`ZADD` + `ZPOPMIN`) with in-memory fallback.

## Partially Implemented
- Redis queue uses `ZPOPMIN`; Lua scripting for richer atomic multi-step dequeue/lock is not yet added.
- Metrics endpoint/exporter (Prometheus) is not yet implemented.
- Device-authenticated staff actions and role/security middleware are not yet implemented.

## Pending
- Aging score refresh worker for long-waiting tasks currently queued in Redis.
- Outbox dispatcher worker for guaranteed event delivery/retry semantics.
- gRPC server using proto contracts (proto currently used as API contract artifact).
- Load tests for skip-locked duplicate prevention and throughput tuning.
