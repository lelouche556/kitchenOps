package models

type RecipeStepDependencyModel struct {
	ItemKey            string `gorm:"column:item_key;primaryKey"`
	StepOrder          int    `gorm:"column:step_order;primaryKey"`
	DependsOnStepOrder int    `gorm:"column:depends_on_step_order;primaryKey"`
}

func (RecipeStepDependencyModel) TableName() string { return "recipe_step_dependencies" }
