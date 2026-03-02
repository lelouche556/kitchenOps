package models

type TaskDependencyModel struct {
	TaskID          uint64 `gorm:"column:task_id;primaryKey"`
	DependsOnTaskID uint64 `gorm:"column:depends_on_task_id;primaryKey"`
}

func (TaskDependencyModel) TableName() string { return "task_dependencies" }
