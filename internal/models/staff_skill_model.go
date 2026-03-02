package models

type StaffSkillModel struct {
	StaffID   uint64 `gorm:"column:staff_id;primaryKey"`
	CounterID uint64 `gorm:"column:counter_id;primaryKey"`
}

func (StaffSkillModel) TableName() string { return "staff_skills" }
