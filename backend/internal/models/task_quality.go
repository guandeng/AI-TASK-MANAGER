package models

import (
	"time"
)

// TaskQualityScore 任务质量评分记录
type TaskQualityScore struct {
	ID                uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	TaskID            uint64    `gorm:"not null;index:idx_task_id" json:"taskId"`
	Version           int       `gorm:"not null;index:idx_version" json:"version"` // 版本号，从 1 开始递增
	TotalScore        float64   `gorm:"type:decimal(5,2);not null" json:"totalScore"`
	ClarityScore      int       `gorm:"not null" json:"clarityScore"`        // 清晰度 1-10
	CompletenessScore int       `gorm:"not null" json:"completenessScore"`   // 完整性 1-10
	StructureScore    int       `gorm:"not null" json:"structureScore"`      // 结构化 1-10
	ActionabilityScore int      `gorm:"not null" json:"actionabilityScore"`  // 可执行性 1-10
	ConsistencyScore  int       `gorm:"not null" json:"consistencyScore"`    // 一致性 1-10
	Evaluation        string    `gorm:"type:text" json:"evaluation"`         // 评价内容（JSON 格式）
	TaskSnapshot      string    `gorm:"type:longtext" json:"taskSnapshot"`   // 任务快照（JSON，评分时的任务状态）
	AIProvider        string    `gorm:"size:50" json:"aiProvider"`           // AI 提供商
	CreatedAt         time.Time `gorm:"autoCreateTime" json:"createdAt"`

	// 关联
	Task Task `gorm:"foreignKey:TaskID;constraint:OnDelete:CASCADE" json:"-"`
}

// TableName 指定表名
func (TaskQualityScore) TableName() string {
	return "task_quality_score"
}

// EvaluationData 评价数据结构
type EvaluationData struct {
	Strengths   []string          `json:"strengths"`   // 优点
	Weaknesses  []string          `json:"weaknesses"`  // 缺点
	Suggestions []EvalSuggestion  `json:"suggestions"` // 改进建议
	Analysis    string            `json:"analysis"`    // 详细分析
}

// EvalSuggestion 改进建议
type EvalSuggestion struct {
	Issue      string `json:"issue"`      // 问题描述
	Suggestion string `json:"suggestion"` // 改进建议
}
