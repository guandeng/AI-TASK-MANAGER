package models

import (
	"time"
)

// Task 任务模型
type Task struct {
	ID               uint64     `gorm:"primaryKey;autoIncrement" json:"id"`
	RequirementID    *uint64    `gorm:"index:idx_requirement_id" json:"requirementId,omitempty"`
	Title            string     `gorm:"size:500;not null" json:"title"`
	TitleTrans       *string    `gorm:"size:500" json:"titleTrans,omitempty"`
	Description      string     `gorm:"type:text" json:"description"`
	DescriptionTrans *string    `gorm:"type:text" json:"descriptionTrans,omitempty"`
	Status           string     `gorm:"size:50;not null;default:pending;index:idx_status" json:"status"`
	Priority         string     `gorm:"size:20;not null;default:medium;index:idx_priority" json:"priority"`
	Details          string     `gorm:"type:text" json:"details"`
	DetailsTrans     *string    `gorm:"type:text" json:"detailsTrans,omitempty"`
	TestStrategy     string     `gorm:"type:text" json:"testStrategy"`
	TestStrategyTrans *string   `gorm:"type:text" json:"testStrategyTrans,omitempty"`
	Assignee         *string    `gorm:"size:100" json:"assignee,omitempty"`
	IsExpanding      bool       `gorm:"default:false" json:"isExpanding"`
	ExpandMessageID  *uint64    `json:"expandMessageId,omitempty"`
	StartDate        *time.Time `gorm:"type:date" json:"startDate,omitempty"`
	DueDate          *time.Time `gorm:"type:date" json:"dueDate,omitempty"`
	CompletedAt      *time.Time `json:"completedAt,omitempty"`
	EstimatedHours   *float64   `gorm:"type:decimal(10,2)" json:"estimatedHours,omitempty"`
	ActualHours      *float64   `gorm:"type:decimal(10,2)" json:"actualHours,omitempty"`
	CreatedAt        time.Time  `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt        time.Time  `gorm:"autoUpdateTime" json:"updatedAt"`

	// 关联
	Subtasks     []Subtask       `gorm:"foreignKey:TaskID;constraint:OnDelete:CASCADE" json:"subtasks,omitempty"`
	Dependencies []TaskDependency `gorm:"foreignKey:TaskID;constraint:OnDelete:CASCADE" json:"dependencies,omitempty"`
	Assignments  []Assignment    `gorm:"foreignKey:TaskID;constraint:OnDelete:CASCADE" json:"assignments,omitempty"`
	Activities   []ActivityLog   `gorm:"foreignKey:TaskID" json:"activities,omitempty"`
}

// TableName 指定表名
func (Task) TableName() string {
	return "task_task"
}

// Subtask 子任务模型
type Subtask struct {
	ID               uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	TaskID           uint64    `gorm:"not null;index:idx_task_id;uniqueIndex:uk_task_subtask" json:"taskId"`
	Title            string    `gorm:"size:500;not null" json:"title"`
	TitleTrans       *string   `gorm:"size:500" json:"titleTrans,omitempty"`
	Description      string    `gorm:"type:text" json:"description"`
	DescriptionTrans *string   `gorm:"type:text" json:"descriptionTrans,omitempty"`
	Details          string    `gorm:"type:text" json:"details"`
	DetailsTrans     *string   `gorm:"type:text" json:"detailsTrans,omitempty"`
	Status           string    `gorm:"size:50;not null;default:pending;index:idx_status" json:"status"`
	SortOrder        uint      `gorm:"default:0" json:"sortOrder"`
	CreatedAt        time.Time `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt        time.Time `gorm:"autoUpdateTime" json:"updatedAt"`

	// 关联
	Task          Task                  `gorm:"foreignKey:TaskID" json:"-"`
	Dependencies  []SubtaskDependency   `gorm:"foreignKey:SubtaskID;constraint:OnDelete:CASCADE" json:"dependencies,omitempty"`
	Assignments   []SubtaskAssignment   `gorm:"foreignKey:SubtaskID;constraint:OnDelete:CASCADE" json:"assignments,omitempty"`
}

// TableName 指定表名
func (Subtask) TableName() string {
	return "task_subtask"
}

// TaskDependency 任务依赖关系
type TaskDependency struct {
	ID             uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	TaskID         uint64    `gorm:"not null;uniqueIndex:uk_task_dependency;index:idx_depends_on" json:"taskId"`
	DependsOnTaskID uint64   `gorm:"not null;uniqueIndex:uk_task_dependency" json:"dependsOnTaskId"`
	CreatedAt      time.Time `gorm:"autoCreateTime" json:"createdAt"`

	// 关联
	Task         Task `gorm:"foreignKey:TaskID;constraint:OnDelete:CASCADE" json:"-"`
	DependsOnTask Task `gorm:"foreignKey:DependsOnTaskID;constraint:OnDelete:CASCADE" json:"-"`
}

// TableName 指定表名
func (TaskDependency) TableName() string {
	return "task_dependency"
}

// SubtaskDependency 子任务依赖关系
type SubtaskDependency struct {
	ID                  uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	SubtaskID           uint64    `gorm:"not null;uniqueIndex:uk_subtask_dependency;index:idx_depends_on" json:"subtaskId"`
	DependsOnSubtaskID  uint64    `gorm:"not null;uniqueIndex:uk_subtask_dependency" json:"dependsOnSubtaskId"`
	CreatedAt           time.Time `gorm:"autoCreateTime" json:"createdAt"`

	// 关联
	Subtask         Subtask `gorm:"foreignKey:SubtaskID;constraint:OnDelete:CASCADE" json:"-"`
	DependsOnSubtask Subtask `gorm:"foreignKey:DependsOnSubtaskID;constraint:OnDelete:CASCADE" json:"-"`
}

// TableName 指定表名
func (SubtaskDependency) TableName() string {
	return "task_subtask_dependency"
}
