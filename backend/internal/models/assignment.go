package models

import (
	"time"
)

// Assignment 任务分配模型
type Assignment struct {
	ID             uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	TaskID         uint64    `gorm:"not null;uniqueIndex:uk_task_member;index:idx_member_id" json:"taskId"`
	MemberID       uint64    `gorm:"not null;uniqueIndex:uk_task_member" json:"memberId"`
	Role           string    `gorm:"size:50;not null;default:assignee;index:idx_role" json:"role"` // assignee, reviewer, collaborator
	AssignedBy     *uint64   `json:"assignedBy,omitempty"`
	EstimatedHours *float64  `gorm:"type:decimal(10,2)" json:"estimatedHours,omitempty"`
	ActualHours    *float64  `gorm:"type:decimal(10,2)" json:"actualHours,omitempty"`
	CreatedAt      time.Time `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt      time.Time `gorm:"autoUpdateTime" json:"updatedAt"`

	// 关联
	Task   Task   `gorm:"foreignKey:TaskID" json:"-"`
	Member Member `gorm:"foreignKey:MemberID" json:"member,omitempty"`
}

// TableName 指定表名
func (Assignment) TableName() string {
	return "task_assignment"
}

// SubtaskAssignment 子任务分配模型
type SubtaskAssignment struct {
	ID             uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	SubtaskID      uint64    `gorm:"not null;uniqueIndex:uk_subtask_member;index:idx_member_id" json:"subtaskId"`
	MemberID       uint64    `gorm:"not null;uniqueIndex:uk_subtask_member" json:"memberId"`
	Role           string    `gorm:"size:50;not null;default:assignee" json:"role"`
	AssignedBy     *uint64   `json:"assignedBy,omitempty"`
	EstimatedHours *float64  `gorm:"type:decimal(10,2)" json:"estimatedHours,omitempty"`
	ActualHours    *float64  `gorm:"type:decimal(10,2)" json:"actualHours,omitempty"`
	CreatedAt      time.Time `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt      time.Time `gorm:"autoUpdateTime" json:"updatedAt"`

	// 关联
	Subtask Subtask `gorm:"foreignKey:SubtaskID" json:"-"`
	Member  Member  `gorm:"foreignKey:MemberID" json:"member,omitempty"`
}

// TableName 指定表名
func (SubtaskAssignment) TableName() string {
	return "task_subtask_assignment"
}
