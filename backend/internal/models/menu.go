package models

import (
	"time"
)

// Menu 菜单模型
type Menu struct {
	ID          uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	Key         string    `gorm:"size:100;not null;uniqueIndex:uk_key" json:"key"`
	ParentKey   *string   `gorm:"size:100;index:idx_parent_key" json:"parentKey,omitempty"`
	Title       string    `gorm:"size:100;not null" json:"title"`
	TitleEn     *string   `gorm:"size:100" json:"titleEn,omitempty"`
	Path        *string   `gorm:"size:200" json:"path,omitempty"`
	Icon        *string   `gorm:"size:50" json:"icon,omitempty"`
	Sort        uint      `gorm:"default:0" json:"sort"`
	Enabled     bool      `gorm:"default:true" json:"enabled"`
	CreatedAt   time.Time `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime" json:"updatedAt"`

	// 关联（非数据库字段，用于构建树形结构）
	Children []Menu `gorm:"-" json:"children,omitempty"`
}

// TableName 指定表名
func (Menu) TableName() string {
	return "task_menu"
}
