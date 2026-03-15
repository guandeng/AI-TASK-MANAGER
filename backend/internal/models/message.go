package models

import (
	"time"
)

// Message 消息模型
type Message struct {
	ID            uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	TaskID        *uint64   `gorm:"column:task_id;index:idx_task_id" json:"taskId"` // 可空，需求拆分时没有关联任务
	RequirementID *uint64   `gorm:"column:requirement_id;index:idx_requirement_id" json:"requirementId"` // 需求ID，用于需求拆分场景
	Type          string    `gorm:"column:type;size:50;not null;index:idx_type" json:"type"` // expand_task, regenerate_subtask, split_requirement
	Status        string    `gorm:"column:status;size:20;not null;default:pending;index:idx_status" json:"status"` // pending, processing, success, failed
	Title         string    `gorm:"column:title;size:255;not null" json:"title"`
	Content       *string   `gorm:"column:content;type:text" json:"content,omitempty"`
	ErrorMessage  *string   `gorm:"column:error_message;type:text" json:"errorMessage,omitempty"`
	ResultSummary *string   `gorm:"column:result_summary;size:500" json:"resultSummary,omitempty"`
	IsRead        bool      `gorm:"column:is_read;default:false;index:idx_is_read" json:"isRead"`
	CreatedAt     time.Time `gorm:"column:created_at;autoCreateTime;index:idx_created_at" json:"createdAt"`
	UpdatedAt     time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updatedAt"`

	// 关联（不使用外键约束，避免迁移问题）
	Task        *Task        `gorm:"foreignKey:TaskID;constraint:OnDelete:SET NULL" json:"-"`
	Requirement *Requirement `gorm:"foreignKey:RequirementID;constraint:OnDelete:SET NULL" json:"-"`
}

// TableName 指定表名
func (Message) TableName() string {
	return "task_message"
}
