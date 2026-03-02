-- Normalize counter references to foreign keys by id.

-- 1) staff_skills: counter(text) -> counter_id(bigint FK)
ALTER TABLE staff_skills ADD COLUMN IF NOT EXISTS counter_id BIGINT;

DO $$
BEGIN
  IF EXISTS (
    SELECT 1 FROM information_schema.columns
    WHERE table_schema = 'public' AND table_name = 'staff_skills' AND column_name = 'counter'
  ) THEN
    EXECUTE '
      UPDATE staff_skills sk
      SET counter_id = c.id
      FROM counters c
      WHERE sk.counter_id IS NULL
        AND sk.counter = c.name';
  END IF;
END $$;

DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1 FROM pg_constraint WHERE conname = 'staff_skills_counter_id_fkey'
  ) THEN
    ALTER TABLE staff_skills
      ADD CONSTRAINT staff_skills_counter_id_fkey
      FOREIGN KEY (counter_id) REFERENCES counters(id) ON DELETE CASCADE;
  END IF;
END $$;

ALTER TABLE staff_skills DROP CONSTRAINT IF EXISTS staff_skills_pkey;
ALTER TABLE staff_skills ALTER COLUMN counter_id SET NOT NULL;
ALTER TABLE staff_skills ADD CONSTRAINT staff_skills_pkey PRIMARY KEY (staff_id, counter_id);
ALTER TABLE staff_skills DROP COLUMN IF EXISTS counter;

-- 2) machines: counter_name(text) -> counter_id(bigint FK)
ALTER TABLE machines ADD COLUMN IF NOT EXISTS counter_id BIGINT;

DO $$
BEGIN
  IF EXISTS (
    SELECT 1 FROM information_schema.columns
    WHERE table_schema = 'public' AND table_name = 'machines' AND column_name = 'counter_name'
  ) THEN
    EXECUTE '
      UPDATE machines m
      SET counter_id = c.id
      FROM counters c
      WHERE m.counter_id IS NULL
        AND m.counter_name = c.name';
  END IF;
END $$;

DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1 FROM pg_constraint WHERE conname = 'machines_counter_id_fkey'
  ) THEN
    ALTER TABLE machines
      ADD CONSTRAINT machines_counter_id_fkey
      FOREIGN KEY (counter_id) REFERENCES counters(id) ON DELETE RESTRICT;
  END IF;
END $$;

ALTER TABLE machines ALTER COLUMN counter_id SET NOT NULL;
ALTER TABLE machines DROP COLUMN IF EXISTS counter_name;

-- 3) recipe_steps: counter(text) -> counter_id(bigint FK)
ALTER TABLE recipe_steps ADD COLUMN IF NOT EXISTS counter_id BIGINT;

DO $$
BEGIN
  IF EXISTS (
    SELECT 1 FROM information_schema.columns
    WHERE table_schema = 'public' AND table_name = 'recipe_steps' AND column_name = 'counter'
  ) THEN
    EXECUTE '
      UPDATE recipe_steps rs
      SET counter_id = c.id
      FROM counters c
      WHERE rs.counter_id IS NULL
        AND rs.counter = c.name';
  END IF;
END $$;

DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1 FROM pg_constraint WHERE conname = 'recipe_steps_counter_id_fkey'
  ) THEN
    ALTER TABLE recipe_steps
      ADD CONSTRAINT recipe_steps_counter_id_fkey
      FOREIGN KEY (counter_id) REFERENCES counters(id) ON DELETE RESTRICT;
  END IF;
END $$;

ALTER TABLE recipe_steps ALTER COLUMN counter_id SET NOT NULL;
ALTER TABLE recipe_steps DROP COLUMN IF EXISTS counter;

-- 4) tasks: counter(text) -> counter_id(bigint FK)
ALTER TABLE tasks ADD COLUMN IF NOT EXISTS counter_id BIGINT;

DO $$
BEGIN
  IF EXISTS (
    SELECT 1 FROM information_schema.columns
    WHERE table_schema = 'public' AND table_name = 'tasks' AND column_name = 'counter'
  ) THEN
    EXECUTE '
      UPDATE tasks t
      SET counter_id = c.id
      FROM counters c
      WHERE t.counter_id IS NULL
        AND t.counter = c.name';
  END IF;
END $$;

DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1 FROM pg_constraint WHERE conname = 'tasks_counter_id_fkey'
  ) THEN
    ALTER TABLE tasks
      ADD CONSTRAINT tasks_counter_id_fkey
      FOREIGN KEY (counter_id) REFERENCES counters(id) ON DELETE RESTRICT;
  END IF;
END $$;

-- If your tasks table has rows, this requires all rows to be mappable to counters.
ALTER TABLE tasks ALTER COLUMN counter_id SET NOT NULL;
ALTER TABLE tasks DROP COLUMN IF EXISTS counter;
