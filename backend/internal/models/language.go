package models

import "time"

// Language 编程语言/技术栈
type Language struct {
	ID          uint64 `gorm:"primaryKey;autoIncrement" json:"id"`
	Name        string `gorm:"column:name;size:50;not null;uniqueIndex:idx_name_category" json:"name"`        // 语言名称，如 Go、Java、Python
	DisplayName string `gorm:"column:display_name;size:100;not null" json:"displayName"`                      // 显示名称
	Category    string `gorm:"column:category;size:20;not null;uniqueIndex:idx_name_category" json:"category"` // 分类：frontend/backend
	Framework   string           `gorm:"column:framework;size:100" json:"framework"`                                    // 框架说明
	Description string           `gorm:"column:description;size:500" json:"description"`                                // 详细描述
	CodeHints   string           `gorm:"column:code_hints;type:text" json:"codeHints"`                                  // 代码提示模板
	Remark      string           `gorm:"column:remark;size:500" json:"remark"`                                          // Claude开发备注说明
	IsActive    bool             `gorm:"column:is_active;default:true" json:"isActive"`                                 // 是否启用
	SortOrder   uint             `gorm:"column:sort_order;default:0" json:"sortOrder"`                                  // 排序
	CreatedAt   time.Time        `gorm:"column:created_at;autoCreateTime" json:"createdAt"`
		UpdatedAt   time.Time        `gorm:"column:updated_at;autoUpdateTime" json:"updatedAt"`
}

// TableName 指定表名
func (Language) TableName() string {
	return "task_languages"
}
