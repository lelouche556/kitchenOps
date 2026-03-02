-- Seed baseline kitchen resources for first-time local runs.
-- This script is idempotent and safe to run multiple times.

INSERT INTO counters (name, capacity, in_use)
SELECT v.name, v.capacity, 0
FROM (VALUES
  ('fryer', 4),
  ('microwave', 2),
  ('griller', 6),
  ('beverage', 2),
  ('adhoc', 99)
) AS v(name, capacity)
WHERE NOT EXISTS (
  SELECT 1 FROM counters c WHERE c.name = v.name
);

INSERT INTO machines (name, counter_id, machine_type, capacity, in_use, is_up)
SELECT v.name, c.id, v.machine_type, v.capacity, 0, TRUE
FROM (VALUES
  ('grill-1', 'griller', 'grill', 6),
  ('fryer-1', 'fryer', 'fryer', 4),
  ('blender-1', 'beverage', 'blender', 2),
  ('micro-1', 'microwave', 'micro', 2),
  ('packing-1', 'adhoc', 'packing', 99)
) AS v(name, counter_name, machine_type, capacity)
JOIN counters c ON c.name = v.counter_name
WHERE NOT EXISTS (
  SELECT 1 FROM machines m WHERE m.name = v.name
);

INSERT INTO staff (name, shift_start, shift_end, max_parallel, efficiency_multiplier, active_tasks, active_seconds, on_break)
SELECT v.name, now() - interval '1 hour', now() + interval '8 hour', v.max_parallel, v.efficiency, 0, 0, FALSE
FROM (VALUES
  ('Asha', 3, 1.10),
  ('Ravi', 2, 0.95),
  ('Noah', 2, 1.00)
) AS v(name, max_parallel, efficiency)
WHERE NOT EXISTS (
  SELECT 1 FROM staff s WHERE s.name = v.name
);

INSERT INTO staff_skills (staff_id, counter_id)
SELECT s.id, c.id
FROM (VALUES
  ('Asha', 'griller'),
  ('Asha', 'fryer'),
  ('Ravi', 'beverage'),
  ('Ravi', 'adhoc'),
  ('Noah', 'microwave'),
  ('Noah', 'adhoc'),
  ('Noah', 'griller')
) AS v(name, counter)
JOIN staff s ON s.name = v.name
JOIN counters c ON c.name = v.counter
WHERE NOT EXISTS (
  SELECT 1
  FROM staff_skills sk
  WHERE sk.staff_id = s.id AND sk.counter_id = c.id
);

INSERT INTO recipe_steps (item_key, step_order, description, counter_id, machine_id, estimate_secs, base_priority, is_active)
SELECT v.item_key, v.step_order, v.description, c.id, m.id, v.estimate_secs, v.base_priority, TRUE
FROM (VALUES
  ('burger_combo', 1, 'grill patty', 'griller', 'grill-1', 80, 20.0),
  ('burger_combo', 2, 'toast bun', 'griller', 'grill-1', 25, 16.0),
  ('burger_combo', 3, 'fry fries', 'fryer', 'fryer-1', 90, 18.0),
  ('burger_combo', 4, 'assemble burger', 'adhoc', 'packing-1', 35, 14.0),
  ('wrap', 1, 'warm tortilla', 'microwave', 'micro-1', 20, 16.0),
  ('wrap', 2, 'grill filling', 'griller', 'grill-1', 60, 17.0),
  ('wrap', 3, 'assemble wrap', 'adhoc', 'packing-1', 45, 15.0),
  ('cold_coffee', 1, 'blend coffee', 'beverage', 'blender-1', 55, 16.0),
  ('cold_coffee', 2, 'cup assembly', 'beverage', 'blender-1', 30, 12.0),
  ('dip_cup', 1, 'fill dip cup', 'adhoc', 'packing-1', 12, 6.0)
) AS v(item_key, step_order, description, counter, machine_name, estimate_secs, base_priority)
JOIN counters c ON c.name = v.counter
JOIN machines m ON m.name = v.machine_name
ON CONFLICT (item_key, step_order) DO UPDATE
SET
  description = EXCLUDED.description,
  counter_id = EXCLUDED.counter_id,
  machine_id = EXCLUDED.machine_id,
  estimate_secs = EXCLUDED.estimate_secs,
  base_priority = EXCLUDED.base_priority,
  is_active = EXCLUDED.is_active;

INSERT INTO recipe_step_dependencies (item_key, step_order, depends_on_step_order)
VALUES
  ('burger_combo', 4, 1),
  ('burger_combo', 4, 2),
  ('burger_combo', 4, 3),
  ('wrap', 3, 1),
  ('wrap', 3, 2),
  ('cold_coffee', 2, 1)
ON CONFLICT (item_key, step_order, depends_on_step_order) DO NOTHING;
