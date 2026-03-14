package models

import (
	"time"
)

// Message 消息模型
type Message struct {
	ID            uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	TaskID        uint64    `gorm:"not null;index:idx_task_id" json:"taskId"`
	Type          string    `gorm:"size:50;not null;index:idx_type" json:"type"` // expand_task, regenerate_subtask
	Status        string    `gorm:"size:20;not null;default:pending;index:idx_status" json:"status"` // pending, processing, success, failed
	Title         string    `gorm:"size:255;not null" json:"title"`
	Content       *string   `gorm:"type:text" json:"content,omitempty"`
	ErrorMessage  *string   `gorm:"type:text" json:"errorMessage,omitempty"`
	ResultSummary *string   `gorm:"size:500" json:"resultSummary,omitempty"`
	IsRead        bool      `gorm:"default:false;index:idx_is_read" json:"isRead"`
	CreatedAt     time.Time `gorm:"autoCreateTime;index:idx_created_at" json:"createdAt"`
	UpdatedAt     time.Time `gorm:"autoUpdateTime" json:"updatedAt"`

	// 关联
	Task Task `gorm:"foreignKey:TaskID" json:"-"`
}

// TableName 指定表名
func (Message) TableName() string {
	return "task_message"
}
