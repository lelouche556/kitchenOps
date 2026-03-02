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
