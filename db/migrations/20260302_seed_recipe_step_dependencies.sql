INSERT INTO recipe_step_dependencies (item_key, step_order, depends_on_step_order)
VALUES
  ('burger_combo', 4, 1),
  ('burger_combo', 4, 2),
  ('burger_combo', 4, 3),
  ('wrap', 3, 1),
  ('wrap', 3, 2),
  ('cold_coffee', 2, 1)
ON CONFLICT (item_key, step_order, depends_on_step_order) DO NOTHING;
