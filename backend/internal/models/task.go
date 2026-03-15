package models

import (
	"time"
)

// Task 任务模型
type Task struct {
	ID               uint64     `gorm:"primaryKey;autoIncrement" json:"id"`
	RequirementID    *uint64    `gorm:"index:idx_requirement_id" json:"requirementId,omitempty"`
	RequirementTitle string     `json:"requirementTitle,omitempty"` // JOIN 查询时从 requirement 表获取
	Title            string     `gorm:"size:500;not null" json:"title"`
	TitleTrans       *string    `gorm:"size:500" json:"titleTrans,omitempty"`
	Description      string     `gorm:"type:text" json:"description"`
	DescriptionTrans *string    `gorm:"type:text" json:"descriptionTrans,omitempty"`
	Status           string     `gorm:"size:50;not null;default:pending;index:idx_status" json:"status"`
	Priority         string     `gorm:"size:20;not null;default:medium;index:idx_priority" json:"priority"`
	Category         string     `gorm:"size:20;default:'';index:idx_category" json:"category"` // 任务分类：frontend/backend
	LanguageID       *uint64    `gorm:"index:idx_language_id" json:"languageId,omitempty"`
	LanguageName     string     `json:"languageName,omitempty"` // JOIN 查询时获取
	Details          string     `gorm:"type:text" json:"details"`
	DetailsTrans     *string    `gorm:"type:text" json:"detailsTrans,omitempty"`
	TestStrategy     string     `gorm:"type:text" json:"testStrategy"`
	TestStrategyTrans *string   `gorm:"type:text" json:"testStrategyTrans,omitempty"`
	Module           *string    `gorm:"size:100" json:"module,omitempty"` // 模块归属
	Input            *string    `gorm:"type:text" json:"input,omitempty"` // 输入依赖
	Output           *string    `gorm:"type:text" json:"output,omitempty"` // 输出交付物
	Risk             *string    `gorm:"type:text" json:"risk,omitempty"` // 风险点
	AcceptanceCriteria *string  `gorm:"type:text" json:"acceptanceCriteria,omitempty"` // 验收标准
	Assignee         *string    `gorm:"size:100" json:"assignee,omitempty"`
	CustomFields     *string    `gorm:"type:longtext" json:"customFields,omitempty"` // 自定义字段值 JSON
	IsExpanding      bool       `gorm:"default:false" json:"isExpanding"`
	ExpandMessageID  *uint64    `json:"expandMessageId,omitempty"`
	ExpandStartedAt  *time.Time `json:"expandStartedAt,omitempty"` // 拆分开始时间
	StartDate        *time.Time `gorm:"type:date" json:"startDate,omitempty"`
	DueDate          *time.Time `gorm:"type:date" json:"dueDate,omitempty"`
	CompletedAt      *time.Time `json:"completedAt,omitempty"`
	EstimatedHours   *float64   `gorm:"type:decimal(10,2)" json:"estimatedHours,omitempty"`
	ActualHours      *float64   `gorm:"type:decimal(10,2)" json:"actualHours,omitempty"`
	CreatedAt        time.Time  `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt        time.Time  `gorm:"autoUpdateTime" json:"updatedAt"`
	DeletedAt        *time.Time `gorm:"index:idx_deleted_at" json:"deletedAt,omitempty"`

	// 子任务统计（通过子查询获取）
	SubtaskCount     int `gorm:"-" json:"subtaskCount"`
	SubtaskDoneCount int `gorm:"-" json:"subtaskDoneCount"`

	// 关联
	Subtasks     []Subtask       `gorm:"foreignKey:TaskID;constraint:OnDelete:CASCADE" json:"subtasks,omitempty"`
	Dependencies []TaskDependency `gorm:"foreignKey:TaskID;constraint:OnDelete:CASCADE" json:"dependencies,omitempty"`
	Assignments  []Assignment    `gorm:"foreignKey:TaskID;constraint:OnDelete:CASCADE" json:"assignments,omitempty"`
	Activities   []ActivityLog   `gorm:"foreignKey:TaskID" json:"activities,omitempty"`
}

// TableName 指定表名
func (Task) TableName() string {
	return "task_task"
}

// Subtask 子任务模型
type Subtask struct {
	ID               uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	TaskID           uint64    `gorm:"not null;index:idx_task_id" json:"taskId"`
	Title            string    `gorm:"size:500;not null" json:"title"`
	TitleTrans       *string   `gorm:"size:500" json:"titleTrans,omitempty"`
	Description      string    `gorm:"type:text" json:"description"`
	DescriptionTrans *string   `gorm:"type:text" json:"descriptionTrans,omitempty"`
	Details          string    `gorm:"type:text" json:"details"`
	DetailsTrans     *string   `gorm:"type:text" json:"detailsTrans,omitempty"`
	Status           string    `gorm:"size:50;not null;default:pending;index:idx_status" json:"status"`
	Priority         string    `gorm:"size:20;not null;default:medium" json:"priority"`
	SortOrder        uint      `gorm:"default:0" json:"sortOrder"`
	EstimatedHours   *float64  `gorm:"type:decimal(10,2)" json:"estimatedHours,omitempty"`
	ActualHours      *float64  `gorm:"type:decimal(10,2)" json:"actualHours,omitempty"`
	CodeInterface    *string   `gorm:"type:text" json:"codeInterface,omitempty"`
	AcceptanceCriteria *string `gorm:"type:text" json:"acceptanceCriteria,omitempty"`
	RelatedFiles     *string   `gorm:"type:text" json:"relatedFiles,omitempty"`
	CodeHints        *string   `gorm:"type:text" json:"codeHints,omitempty"`
	CustomFields     *string   `gorm:"type:longtext" json:"customFields,omitempty"` // 自定义字段值 JSON
	CreatedAt        time.Time `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt        time.Time `gorm:"autoUpdateTime" json:"updatedAt"`

	// 关联
	Task          Task                  `gorm:"foreignKey:TaskID" json:"-"`
	Dependencies  []SubtaskDependency   `gorm:"foreignKey:SubtaskID;constraint:OnDelete:CASCADE" json:"dependencies,omitempty"`
	Assignments   []SubtaskAssignment   `gorm:"foreignKey:SubtaskID;constraint:OnDelete:CASCADE" json:"assignments,omitempty"`
}

// TableName 指定表名
func (Subtask) TableName() string {
	return "task_subtask"
}

// TaskDependency 任务依赖关系
type TaskDependency struct {
	ID             uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	TaskID         uint64    `gorm:"not null;uniqueIndex:uk_task_dependency;index:idx_depends_on" json:"taskId"`
	DependsOnTaskID uint64   `gorm:"not null;uniqueIndex:uk_task_dependency" json:"dependsOnTaskId"`
	CreatedAt      time.Time `gorm:"autoCreateTime" json:"createdAt"`

	// 关联
	Task         Task `gorm:"foreignKey:TaskID;constraint:OnDelete:CASCADE" json:"-"`
	DependsOnTask Task `gorm:"foreignKey:DependsOnTaskID;constraint:OnDelete:CASCADE" json:"-"`
}

// TableName 指定表名
func (TaskDependency) TableName() string {
	return "task_dependency"
}

// SubtaskDependency 子任务依赖关系
type SubtaskDependency struct {
	ID                  uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	SubtaskID           uint64    `gorm:"not null;uniqueIndex:uk_subtask_dependency;index:idx_depends_on" json:"subtaskId"`
	DependsOnSubtaskID  uint64    `gorm:"not null;uniqueIndex:uk_subtask_dependency" json:"dependsOnSubtaskId"`
	CreatedAt           time.Time `gorm:"autoCreateTime" json:"createdAt"`

	// 关联
	Subtask         Subtask `gorm:"foreignKey:SubtaskID;constraint:OnDelete:CASCADE" json:"-"`
	DependsOnSubtask Subtask `gorm:"foreignKey:DependsOnSubtaskID;constraint:OnDelete:CASCADE" json:"-"`
}

// TableName 指定表名
func (SubtaskDependency) TableName() string {
	return "task_subtask_dependency"
}

// TaskComplexityReport 任务复杂度分析报告
type TaskComplexityReport struct {
	ID           uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	RequirementID *uint64  `gorm:"index" json:"requirementId,omitempty"`
	Status       string    `gorm:"size:50;not null;default:pending" json:"status"` // pending, processing, completed, failed
	ReportData   string    `gorm:"type:longtext" json:"reportData"` // JSON 格式的报告数据
	ErrorMessage *string   `gorm:"type:text" json:"errorMessage,omitempty"`
	CreatedAt    time.Time `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt    time.Time `gorm:"autoUpdateTime" json:"updatedAt"`
}

// TableName 指定表名
func (TaskComplexityReport) TableName() string {
	return "task_complexity_report"
}

// ComplexityAnalysis 单个任务的复杂度分析
type ComplexityAnalysis struct {
	TaskID          int      `json:"taskId"`
	TaskTitle       string   `json:"taskTitle"`
	ComplexityScore int      `json:"complexityScore"` // 1-10
	ComplexityLevel string   `json:"complexityLevel"` // low, medium, high
	Reasoning       string   `json:"reasoning"`
	SubtaskCount    int      `json:"subtaskCount"`    // 建议的子任务数量
	TimeEstimate    string   `json:"timeEstimate"`    // 预估时间
	Dependencies    []int    `json:"dependencies"`    // 依赖的任务ID
	RiskFactors     []string `json:"riskFactors"`     // 风险因素
}

// ComplexityReportData 复杂度报告数据结构
type ComplexityReportData struct {
	Analyses    []ComplexityAnalysis `json:"analyses"`
	Summary     ComplexitySummary    `json:"summary"`
	GeneratedAt string               `json:"generatedAt"`
}

// ComplexitySummary 复杂度汇总
type ComplexitySummary struct {
	TotalTasks       int `json:"totalTasks"`
	LowComplexity    int `json:"lowComplexity"`
	MediumComplexity int `json:"mediumComplexity"`
	HighComplexity   int `json:"highComplexity"`
	AverageScore     int `json:"averageScore"`
}
