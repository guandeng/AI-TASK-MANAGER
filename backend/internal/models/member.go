package models

import (
	"database/sql/driver"
	"encoding/json"
	"time"
)

// Member 团队成员模型
type Member struct {
	ID         uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	Name       string    `gorm:"size:100;not null" json:"name"`
	Email      *string   `gorm:"size:255;uniqueIndex:uk_email" json:"email,omitempty"`
	Avatar     *string   `gorm:"size:500" json:"avatar,omitempty"`
	Role       string    `gorm:"size:50;not null;default:member;index:idx_role" json:"role"` // admin, leader, member
	Department *string   `gorm:"size:100;index:idx_department" json:"department,omitempty"`
	Skills     StringArray `gorm:"size:500;type:json" json:"skills"`
	Status     string    `gorm:"size:20;not null;default:active;index:idx_status" json:"status"` // active, inactive
	CreatedAt  time.Time `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt  time.Time `gorm:"autoUpdateTime" json:"updatedAt"`

	// 关联
	Assignments       []Assignment       `gorm:"foreignKey:MemberID" json:"assignments,omitempty"`
	SubtaskAssignments []SubtaskAssignment `gorm:"foreignKey:MemberID" json:"subtaskAssignments,omitempty"`
}

// TableName 指定表名
func (Member) TableName() string {
	return "task_member"
}

// StringArray 自定义类型用于存储字符串数组
type StringArray []string

// Value 实现 driver.Valuer 接口
func (s StringArray) Value() (driver.Value, error) {
	if s == nil {
		return nil, nil
	}
	return json.Marshal(s)
}

// Scan 实现 sql.Scanner 接口
func (s *StringArray) Scan(value interface{}) error {
	if value == nil {
		*s = nil
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(bytes, s)
}
