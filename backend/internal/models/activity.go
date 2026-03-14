package models

import (
	"database/sql/driver"
	"encoding/json"
	"time"
)

// ActivityLog 活动日志模型
type ActivityLog struct {
	ID          uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	TaskID      uint64    `gorm:"not null;index:idx_task_id" json:"taskId"`
	SubtaskID   *uint64   `gorm:"index:idx_subtask_id" json:"subtaskId,omitempty"`
	MemberID    *uint64   `gorm:"index:idx_member_id" json:"memberId,omitempty"`
	Action      string    `gorm:"size:50;not null;index:idx_action" json:"action"`
	FieldName   *string   `gorm:"size:50" json:"fieldName,omitempty"`
	OldValue    *string   `gorm:"type:text" json:"oldValue,omitempty"`
	NewValue    *string   `gorm:"type:text" json:"newValue,omitempty"`
	Description *string   `gorm:"size:500" json:"description,omitempty"`
	Metadata    JSONMap   `gorm:"type:text" json:"metadata,omitempty"`
	CreatedAt   time.Time `gorm:"autoCreateTime;index:idx_created_at" json:"createdAt"`

	// 关联
	Task   Task    `gorm:"foreignKey:TaskID" json:"-"`
	Member *Member `gorm:"foreignKey:MemberID" json:"member,omitempty"`
}

// TableName 指定表名
func (ActivityLog) TableName() string {
	return "task_activity_log"
}

// JSONMap 自定义类型用于存储 JSON 对象
type JSONMap map[string]interface{}

// Value 实现 driver.Valuer 接口
func (j JSONMap) Value() (driver.Value, error) {
	if j == nil {
		return nil, nil
	}
	return json.Marshal(j)
}

// Scan 实现 sql.Scanner 接口
func (j *JSONMap) Scan(value interface{}) error {
	if value == nil {
		*j = nil
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(bytes, j)
}
