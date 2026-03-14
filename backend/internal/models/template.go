package models

import (
	"time"
)

// ProjectTemplate 项目模板
type ProjectTemplate struct {
	ID          uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	Name        string    `gorm:"size:255;not null" json:"name"`
	Description string    `gorm:"type:text" json:"description,omitempty"`
	Category    string    `gorm:"size:50;default:'other'" json:"category,omitempty"`
	IsPublic    bool      `gorm:"default:true" json:"isPublic"`
	CreatedBy   *uint64   `json:"createdBy,omitempty"`
	UsageCount  int       `gorm:"default:0" json:"usageCount"`
	Tags        *string   `gorm:"size:500" json:"tags,omitempty"`
	CreatedAt   time.Time `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime" json:"updatedAt"`

	// 关联
	Tasks []ProjectTemplateTask `gorm:"foreignKey:TemplateID;constraint:OnDelete:CASCADE" json:"tasks,omitempty"`
}

// TableName 指定表名
func (ProjectTemplate) TableName() string {
	return "task_project_template"
}

// ProjectTemplateTask 项目模板任务
type ProjectTemplateTask struct {
	ID             uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	TemplateID     uint64    `gorm:"not null;index:idx_template_id" json:"templateId"`
	Title          string    `gorm:"size:500;not null" json:"title"`
	Description    string    `gorm:"type:text" json:"description,omitempty"`
	Priority       string    `gorm:"size:20;not null;default:'medium'" json:"priority"`
	SortOrder      uint      `gorm:"default:0" json:"order"`
	EstimatedHours *float64  `gorm:"type:decimal(10,2)" json:"estimatedHours,omitempty"`
	Dependencies   *string   `gorm:"type:text" json:"dependencies,omitempty"` // JSON array of task IDs
	CreatedAt      time.Time `gorm:"autoCreateTime" json:"createdAt"`

	// 关联
	Subtasks []ProjectTemplateSubtask `gorm:"foreignKey:TemplateTaskID;constraint:OnDelete:CASCADE" json:"subtasks,omitempty"`
}

// TableName 指定表名
func (ProjectTemplateTask) TableName() string {
	return "task_project_template_task"
}

// ProjectTemplateSubtask 项目模板子任务
type ProjectTemplateSubtask struct {
	ID             uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	TemplateTaskID uint64    `gorm:"not null;index:idx_template_task_id" json:"templateTaskId"`
	Title          string    `gorm:"size:500;not null" json:"title"`
	Description    string    `gorm:"type:text" json:"description,omitempty"`
	SortOrder      uint      `gorm:"default:0" json:"order"`
	EstimatedHours *float64  `gorm:"type:decimal(10,2)" json:"estimatedHours,omitempty"`
	CreatedAt      time.Time `gorm:"autoCreateTime" json:"createdAt"`
}

// TableName 指定表名
func (ProjectTemplateSubtask) TableName() string {
	return "task_project_template_subtask"
}

// TaskTemplate 独立任务模板
type TaskTemplate struct {
	ID             uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	Name           string    `gorm:"size:255;not null" json:"name"`
	Description    string    `gorm:"type:text" json:"description,omitempty"`
	Title          string    `gorm:"size:500;not null" json:"title"`
	TaskDescription string   `gorm:"type:text" json:"taskDescription,omitempty"`
	Priority       string    `gorm:"size:20;not null;default:'medium'" json:"priority"`
	EstimatedHours *float64  `gorm:"type:decimal(10,2)" json:"estimatedHours,omitempty"`
	Subtasks       *string   `gorm:"type:text" json:"subtasks,omitempty"` // JSON array of strings
	Tags           *string   `gorm:"size:500" json:"tags,omitempty"`
	IsPublic       bool      `gorm:"default:true" json:"isPublic"`
	CreatedBy      *uint64   `json:"createdBy,omitempty"`
	UsageCount     int       `gorm:"default:0" json:"usageCount"`
	CreatedAt      time.Time `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt      time.Time `gorm:"autoUpdateTime" json:"updatedAt"`
}

// TableName 指定表名
func (TaskTemplate) TableName() string {
	return "task_template"
}
