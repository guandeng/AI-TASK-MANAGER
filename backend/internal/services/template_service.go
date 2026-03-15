package services

import (
	"encoding/json"

	"github.com/ai-task-manager/backend/internal/database"
	"github.com/ai-task-manager/backend/internal/models"
	"go.uber.org/zap"
)

// TemplateService 模板服务
type TemplateService struct {
	logger *zap.Logger
}

// NewTemplateService 创建模板服务
func NewTemplateService(logger *zap.Logger) *TemplateService {
	return &TemplateService{
		logger: logger,
	}
}

// FieldDefinition 字段定义
type FieldDefinition struct {
	Name     string   `json:"name"`
	Label    string   `json:"label"`
	Type     string   `json:"type"`
	Options  []string `json:"options,omitempty"`
	Required bool     `json:"required"`
}

// FieldSchema 字段模式
type FieldSchema struct {
	Fields []FieldDefinition `json:"fields"`
}

// InitDefaultTemplates 初始化默认项目模板
func (s *TemplateService) InitDefaultTemplates() error {
	db := database.GetDB()

	// 检查是否已存在默认模板
	var count int64
	if err := db.Model(&models.ProjectTemplate{}).Where("category IN ?", []string{"frontend", "backend"}).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		s.logger.Info("默认项目模板已存在，跳过初始化")
		return nil
	}

	// 前端项目模板字段定义
	frontendFieldSchema := FieldSchema{
		Fields: []FieldDefinition{
			{Name: "module", Label: "模块归属", Type: "select", Options: []string{"组件开发", "页面开发", "路由配置", "状态管理", "样式开发", "工具函数"}, Required: true},
			{Name: "input", Label: "输入依赖", Type: "text", Required: true},
			{Name: "output", Label: "输出交付物", Type: "text", Required: true},
			{Name: "acceptanceCriteria", Label: "验收标准", Type: "array", Required: true},
			{Name: "risk", Label: "风险点", Type: "text", Required: false},
			{Name: "estimatedHours", Label: "预估工时", Type: "number", Required: true},
		},
	}

	// 后端项目模板字段定义
	backendFieldSchema := FieldSchema{
		Fields: []FieldDefinition{
			{Name: "module", Label: "模块归属", Type: "select", Options: []string{"API 接口", "数据库设计", "业务逻辑", "中间件", "数据处理", "工具函数"}, Required: true},
			{Name: "input", Label: "输入依赖", Type: "text", Required: true},
			{Name: "output", Label: "输出交付物", Type: "text", Required: true},
			{Name: "acceptanceCriteria", Label: "验收标准", Type: "array", Required: true},
			{Name: "risk", Label: "风险点", Type: "text", Required: false},
			{Name: "estimatedHours", Label: "预估工时", Type: "number", Required: true},
		},
	}

	// 序列化字段定义
	frontendSchemaJSON, _ := json.Marshal(frontendFieldSchema)
	backendSchemaJSON, _ := json.Marshal(backendFieldSchema)

	// 创建前端项目模板
	frontendTemplate := models.ProjectTemplate{
		Name:        "前端项目开发模板",
		Description: "适用于前端项目开发的任务模板，包含组件开发、页面开发、状态管理等模块",
		Category:    "frontend",
		IsPublic:    true,
		FieldSchema: ptr(string(frontendSchemaJSON)),
		UsageCount:  0,
	}

	// 创建后端项目模板
	backendTemplate := models.ProjectTemplate{
		Name:        "后端项目开发模板",
		Description: "适用于后端项目开发的任务模板，包含 API 接口、数据库设计、业务逻辑等模块",
		Category:    "backend",
		IsPublic:    true,
		FieldSchema: ptr(string(backendSchemaJSON)),
		UsageCount:  0,
	}

	if err := db.Create(&frontendTemplate).Error; err != nil {
		s.logger.Error("创建前端项目模板失败", zap.Error(err))
		return err
	}

	if err := db.Create(&backendTemplate).Error; err != nil {
		s.logger.Error("创建后端项目模板失败", zap.Error(err))
		return err
	}

	s.logger.Info("默认项目模板初始化成功",
		zap.Uint64("frontend_template_id", frontendTemplate.ID),
		zap.Uint64("backend_template_id", backendTemplate.ID))

	return nil
}

func ptr[T any](v T) *T {
	return &v
}
