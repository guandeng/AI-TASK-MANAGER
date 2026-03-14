package models

import (
	"time"
)

// Config 配置模型（元数据表）
type Config struct {
	MetaKey   string    `gorm:"size:100;primaryKey" json:"metaKey"`
	MetaValue *string   `gorm:"type:longtext" json:"metaValue,omitempty"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updatedAt"`
}

// TableName 指定表名
func (Config) TableName() string {
	return "task_meta"
}
