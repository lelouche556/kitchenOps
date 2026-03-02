-- Normalize machine references to foreign keys by id.

-- 1) recipe_steps: machine_required(text) -> machine_id(bigint FK)
ALTER TABLE recipe_steps ADD COLUMN IF NOT EXISTS machine_id BIGINT;

DO $$
BEGIN
  IF EXISTS (
    SELECT 1 FROM information_schema.columns
    WHERE table_schema = 'public' AND table_name = 'recipe_steps' AND column_name = 'machine_required'
  ) THEN
    EXECUTE '
      UPDATE recipe_steps rs
      SET machine_id = m.id
      FROM machines m
      WHERE rs.machine_id IS NULL
        AND m.counter_id = rs.counter_id
        AND m.machine_type = rs.machine_required';
  END IF;
END $$;

DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1 FROM pg_constraint WHERE conname = 'recipe_steps_machine_id_fkey'
  ) THEN
    ALTER TABLE recipe_steps
      ADD CONSTRAINT recipe_steps_machine_id_fkey
      FOREIGN KEY (machine_id) REFERENCES machines(id) ON DELETE RESTRICT;
  END IF;
END $$;

ALTER TABLE recipe_steps DROP COLUMN IF EXISTS machine_required;

-- 2) tasks: machine_required(text) -> machine_id(bigint FK)
ALTER TABLE tasks ADD COLUMN IF NOT EXISTS machine_id BIGINT;

DO $$
BEGIN
  IF EXISTS (
    SELECT 1 FROM information_schema.columns
    WHERE table_schema = 'public' AND table_name = 'tasks' AND column_name = 'machine_required'
  ) THEN
    EXECUTE '
      UPDATE tasks t
      SET machine_id = m.id
      FROM machines m
      WHERE t.machine_id IS NULL
        AND m.counter_id = t.counter_id
        AND m.machine_type = t.machine_required';
  END IF;
END $$;

DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1 FROM pg_constraint WHERE conname = 'tasks_machine_id_fkey'
  ) THEN
    ALTER TABLE tasks
      ADD CONSTRAINT tasks_machine_id_fkey
      FOREIGN KEY (machine_id) REFERENCES machines(id) ON DELETE RESTRICT;
  END IF;
END $$;

ALTER TABLE tasks DROP COLUMN IF EXISTS machine_required;
