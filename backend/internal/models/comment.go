package models

import (
	"time"
)

// Comment 评论模型
type Comment struct {
	ID        uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	TaskID    uint64    `gorm:"not null;index:idx_task_id" json:"taskId"`
	SubtaskID *uint64   `gorm:"index:idx_subtask_id" json:"subtaskId,omitempty"`
	MemberID  uint64    `gorm:"not null;index:idx_member_id" json:"memberId"`
	ParentID  *uint64   `gorm:"index:idx_parent_id" json:"parentId,omitempty"`
	Content   string    `gorm:"type:text;not null" json:"content"`
	Mentions  string    `gorm:"type:text" json:"mentions,omitempty"` // JSON array of member IDs
	CreatedAt time.Time `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updatedAt"`

	// 关联
	Member   Member    `gorm:"foreignKey:MemberID" json:"member,omitempty"`
	Replies  []Comment `gorm:"foreignKey:ParentID" json:"replies,omitempty"`
}

// TableName 指定表名
func (Comment) TableName() string {
	return "task_comment"
}

// CommentStatistics 评论统计
type CommentStatistics struct {
	Total         int64 `json:"total"`
	UniqueAuthors int64 `json:"uniqueAuthors"`
}
