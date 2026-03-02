CREATE TABLE IF NOT EXISTS orders (
    id BIGSERIAL PRIMARY KEY,
    external_order_id TEXT UNIQUE,
    status TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS counters (
    id BIGSERIAL PRIMARY KEY,
    name TEXT UNIQUE NOT NULL,
    capacity INTEGER NOT NULL,
    in_use INTEGER NOT NULL DEFAULT 0,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS staff (
    id BIGSERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    shift_start TIMESTAMPTZ NOT NULL,
    shift_end TIMESTAMPTZ NOT NULL,
    max_parallel INTEGER NOT NULL,
    efficiency_multiplier NUMERIC(6,3) NOT NULL DEFAULT 1.0,
    active_tasks INTEGER NOT NULL DEFAULT 0,
    active_seconds BIGINT NOT NULL DEFAULT 0,
    on_break BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS staff_skills (
    staff_id BIGINT NOT NULL REFERENCES staff(id) ON DELETE CASCADE,
    counter_id BIGINT NOT NULL REFERENCES counters(id) ON DELETE CASCADE,
    PRIMARY KEY (staff_id, counter_id)
);

CREATE TABLE IF NOT EXISTS machines (
    id BIGSERIAL PRIMARY KEY,
    name TEXT UNIQUE NOT NULL,
    counter_id BIGINT NOT NULL REFERENCES counters(id) ON DELETE RESTRICT,
    machine_type TEXT NOT NULL,
    capacity INTEGER NOT NULL DEFAULT 1,
    in_use INTEGER NOT NULL DEFAULT 0,
    is_up BOOLEAN NOT NULL DEFAULT TRUE,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS recipe_steps (
    id BIGSERIAL PRIMARY KEY,
    item_key TEXT NOT NULL,
    step_order INTEGER NOT NULL,
    description TEXT NOT NULL,
    counter_id BIGINT NOT NULL REFERENCES counters(id) ON DELETE RESTRICT,
    machine_id BIGINT REFERENCES machines(id) ON DELETE RESTRICT,
    estimate_secs INTEGER NOT NULL,
    base_priority NUMERIC(10,3) NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (item_key, step_order)
);

CREATE INDEX IF NOT EXISTS idx_recipe_steps_item_active_order ON recipe_steps(item_key, is_active, step_order);

CREATE TABLE IF NOT EXISTS recipe_step_dependencies (
    item_key TEXT NOT NULL,
    step_order INTEGER NOT NULL,
    depends_on_step_order INTEGER NOT NULL,
    PRIMARY KEY (item_key, step_order, depends_on_step_order),
    FOREIGN KEY (item_key, step_order) REFERENCES recipe_steps(item_key, step_order) ON DELETE CASCADE,
    FOREIGN KEY (item_key, depends_on_step_order) REFERENCES recipe_steps(item_key, step_order) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_recipe_step_deps_item_step
    ON recipe_step_dependencies(item_key, step_order);

CREATE TABLE IF NOT EXISTS tasks (
    id BIGSERIAL PRIMARY KEY,
    order_id BIGINT NOT NULL REFERENCES orders(id),
    description TEXT NOT NULL,
    counter_id BIGINT NOT NULL REFERENCES counters(id) ON DELETE RESTRICT,
    machine_id BIGINT REFERENCES machines(id) ON DELETE RESTRICT,
    estimate_secs INTEGER NOT NULL,
    base_priority NUMERIC(10,3) NOT NULL,
    pending_deps INTEGER NOT NULL DEFAULT 0,
    assigned_staff_id BIGINT,
    assigned_machine_id BIGINT,
    status TEXT NOT NULL,
    assignment_version INTEGER NOT NULL DEFAULT 0,
    claimed_by TEXT,
    claimed_until TIMESTAMPTZ,
    assigned_at TIMESTAMPTZ,
    started_at TIMESTAMPTZ,
    completed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_tasks_status_priority ON tasks(status, base_priority, created_at);
CREATE INDEX IF NOT EXISTS idx_tasks_claimed_until ON tasks(claimed_until);

CREATE TABLE IF NOT EXISTS task_dependencies (
    task_id BIGINT NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
    depends_on_task_id BIGINT NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
    PRIMARY KEY (task_id, depends_on_task_id)
);

CREATE TABLE IF NOT EXISTS domain_events (
    id BIGSERIAL PRIMARY KEY,
    aggregate_type TEXT NOT NULL,
    aggregate_id TEXT NOT NULL,
    event_type TEXT NOT NULL,
    payload JSONB NOT NULL,
    attempts INTEGER NOT NULL DEFAULT 0,
    next_retry_at TIMESTAMPTZ,
    published_at TIMESTAMPTZ,
    last_error TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_domain_events_pending
    ON domain_events (published_at, next_retry_at, id);
