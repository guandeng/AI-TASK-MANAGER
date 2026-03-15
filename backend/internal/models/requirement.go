package models

import (
	"time"
)

// Requirement 需求模型
type Requirement struct {
	ID          uint64     `gorm:"primaryKey;autoIncrement" json:"id"`
	Title       string     `gorm:"size:500;not null" json:"title"`
	Content     string     `gorm:"type:longtext" json:"content"`
	Status      string     `gorm:"size:50;not null;default:draft;index:idx_status" json:"status"`
	Priority    string     `gorm:"size:20;not null;default:medium;index:idx_priority" json:"priority"`
	Tags        *string    `gorm:"size:500" json:"tags,omitempty"`
	Assignee    *string    `gorm:"size:100;index:idx_assignee" json:"assignee,omitempty"`
	CreatedAt   time.Time  `gorm:"autoCreateTime;index:idx_created_at" json:"createdAt"`
	UpdatedAt   time.Time  `gorm:"autoUpdateTime" json:"updatedAt"`
	DeletedAt   *time.Time `gorm:"index:idx_deleted_at" json:"deletedAt,omitempty"`

	// 关联
	Documents  []RequirementDocument `gorm:"foreignKey:RequirementID" json:"documents,omitempty"`
	Tasks      []Task                `gorm:"foreignKey:RequirementID" json:"tasks,omitempty"`
}

// TableName 指定表名
func (Requirement) TableName() string {
	return "task_requirement"
}

// RequirementDocument 需求文档模型
type RequirementDocument struct {
	ID             uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	RequirementID  uint64    `gorm:"not null;index:idx_requirement_id" json:"requirementId"`
	Name           string    `gorm:"size:255;not null" json:"name"`
	Path           string    `gorm:"size:500;not null" json:"path"`
	Size           uint64    `gorm:"default:0" json:"size"`
	MimeType       *string   `gorm:"size:100" json:"mimeType,omitempty"`
	Description    *string   `gorm:"size:500" json:"description,omitempty"`
	UploadedBy     *string   `gorm:"size:100" json:"uploadedBy,omitempty"`
	CreatedAt      time.Time `gorm:"autoCreateTime" json:"createdAt"`

	// 关联
	Requirement Requirement `gorm:"foreignKey:RequirementID" json:"-"`
}

// TableName 指定表名
func (RequirementDocument) TableName() string {
	return "task_requirement_document"
}
