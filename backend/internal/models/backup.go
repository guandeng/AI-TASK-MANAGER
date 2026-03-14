package models

import (
	"time"
)

// Backup 备份记录
type Backup struct {
	ID            uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	RequirementID uint64    `gorm:"not null;index:idx_requirement_id" json:"requirementId"`
	BackupType    string    `gorm:"size:50;not null" json:"type"`
	DataSnapshot  string    `gorm:"type:longtext" json:"dataSnapshot"`
	TaskCount     int       `gorm:"default:0" json:"taskCount"`
	Status        string    `gorm:"size:20;default:success" json:"status"`
	ErrorMessage  *string   `gorm:"type:text" json:"errorMessage,omitempty"`
	CreatedAt     time.Time `gorm:"autoCreateTime;index:idx_created_at" json:"createdAt"`
}

// TableName 指定表名
func (Backup) TableName() string {
	return "task_backup"
}

// BackupSchedule 备份计划配置
type BackupSchedule struct {
	ID             uint64     `gorm:"primaryKey;autoIncrement" json:"id"`
	RequirementID  uint64     `gorm:"uniqueIndex;not null" json:"requirementId"`
	Enabled        bool       `gorm:"default:false" json:"enabled"`
	IntervalType   string     `gorm:"size:20;not null" json:"intervalType"`
	IntervalValue  int        `gorm:"not null" json:"intervalValue"`
	RetentionCount int        `gorm:"default:10" json:"retentionCount"`
	LastBackupAt   *time.Time `json:"lastBackupAt"`
	CreatedAt      time.Time  `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt      time.Time  `gorm:"autoUpdateTime" json:"updatedAt"`
}

// TableName 指定表名
func (BackupSchedule) TableName() string {
	return "task_backup_schedule"
}
