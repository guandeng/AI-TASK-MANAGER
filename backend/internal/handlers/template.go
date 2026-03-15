package handlers

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/ai-task-manager/backend/internal/config"
	"github.com/ai-task-manager/backend/internal/database"
	"github.com/ai-task-manager/backend/internal/models"
	"github.com/ai-task-manager/backend/internal/services"
	"github.com/ai-task-manager/backend/pkg/ai"
	"github.com/ai-task-manager/backend/pkg/response"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// TemplateHandler 模板处理器
type TemplateHandler struct {
	logger  *zap.Logger
	aiSvc   services.AIService
}

// NewTemplateHandler 创建模板处理器
func NewTemplateHandler(logger *zap.Logger, cfg *config.Config) *TemplateHandler {
	var aiSvc services.AIService
	if cfg != nil && cfg.AI.Provider != "" {
		aiSvc = ai.NewService(&cfg.AI)
	}
	return &TemplateHandler{
		logger: logger,
		aiSvc:  aiSvc,
	}
}

// ============ 项目模板 ============

// ListProjectTemplates 获取项目模板列表
func (h *TemplateHandler) ListProjectTemplates(c *gin.Context) {
	db := database.GetDB()

	var templates []models.ProjectTemplate
	query := db.Model(&models.ProjectTemplate{}).Where("deleted_at IS NULL")

	// 筛选条件
	if category := c.Query("category"); category != "" {
		query = query.Where("category = ?", category)
	}
	if keyword := c.Query("keyword"); keyword != "" {
		query = query.Where("name LIKE ?", "%"+keyword+"%")
	}

	if err := query.Order("created_at DESC").Find(&templates).Error; err != nil {
		h.logger.Error("获取项目模板列表失败", zap.Error(err))
		response.Error(c, 500, "获取项目模板列表失败")
		return
	}

	response.Success(c, templates)
}

// GetProjectTemplate 获取项目模板详情
func (h *TemplateHandler) GetProjectTemplate(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的模板 ID")
		return
	}

	db := database.GetDB()
	var template models.ProjectTemplate
	if err := db.Where("deleted_at IS NULL").Preload("Tasks.Subtasks").First(&template, id).Error; err != nil {
		response.NotFound(c, "模板不存在")
		return
	}

	response.Success(c, template)
}

// CreateProjectTemplateRequest 创建项目模板请求
type CreateProjectTemplateRequest struct {
	Name        string                              `json:"name" binding:"required"`
	Description string                              `json:"description"`
	Category    string                              `json:"category"`
	IsPublic    bool                                `json:"isPublic"`
	Tags        []string                            `json:"tags"`
	FieldSchema json.RawMessage                     `json:"fieldSchema"`
	Tasks       []CreateTemplateTaskRequest         `json:"tasks"`
}

// CreateTemplateTaskRequest 创建模板任务请求
type CreateTemplateTaskRequest struct {
	Title          string                            `json:"title" binding:"required"`
	Description    string                            `json:"description"`
	Priority       string                            `json:"priority"`
	Order          uint                              `json:"order"`
	EstimatedHours *float64                          `json:"estimatedHours"`
	Dependencies   []uint64                          `json:"dependencies"`
	Subtasks       []CreateTemplateSubtaskRequest    `json:"subtasks"`
}

// CreateTemplateSubtaskRequest 创建模板子任务请求
type CreateTemplateSubtaskRequest struct {
	Title          string   `json:"title" binding:"required"`
	Description    string   `json:"description"`
	Order          uint     `json:"order"`
	EstimatedHours *float64 `json:"estimatedHours"`
}

// CreateProjectTemplate 创建项目模板
func (h *TemplateHandler) CreateProjectTemplate(c *gin.Context) {
	var req CreateProjectTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数格式错误")
		return
	}

	db := database.GetDB()

	// 处理标签
	var tagsStr *string
	if len(req.Tags) > 0 {
		tagsJSON, _ := json.Marshal(req.Tags)
		s := string(tagsJSON)
		tagsStr = &s
	}

	// 处理字段定义
	var fieldSchemaStr *string
	if len(req.FieldSchema) > 0 {
		s := string(req.FieldSchema)
		fieldSchemaStr = &s
	}

	// 创建模板
	template := models.ProjectTemplate{
		Name:        req.Name,
		Description: req.Description,
		Category:    req.Category,
		IsPublic:    req.IsPublic,
		Tags:        tagsStr,
		FieldSchema: fieldSchemaStr,
		UsageCount:  0,
	}

	if err := db.Create(&template).Error; err != nil {
		h.logger.Error("创建项目模板失败", zap.Error(err))
		response.Error(c, 500, "创建项目模板失败")
		return
	}

	// 创建任务
	for _, taskReq := range req.Tasks {
		priority := taskReq.Priority
		if priority == "" {
			priority = "medium"
		}

		var depsStr *string
		if len(taskReq.Dependencies) > 0 {
			depsJSON, _ := json.Marshal(taskReq.Dependencies)
			s := string(depsJSON)
			depsStr = &s
		}

		task := models.ProjectTemplateTask{
			TemplateID:     template.ID,
			Title:          taskReq.Title,
			Description:    taskReq.Description,
			Priority:       priority,
			SortOrder:      taskReq.Order,
			EstimatedHours: taskReq.EstimatedHours,
			Dependencies:   depsStr,
		}

		if err := db.Create(&task).Error; err != nil {
			continue
		}

		// 创建子任务
		for _, subtaskReq := range taskReq.Subtasks {
			subtask := models.ProjectTemplateSubtask{
				TemplateTaskID: task.ID,
				Title:          subtaskReq.Title,
				Description:    subtaskReq.Description,
				SortOrder:      subtaskReq.Order,
				EstimatedHours: subtaskReq.EstimatedHours,
			}
			db.Create(&subtask)
		}
	}

	// 重新加载
	db.Preload("Tasks.Subtasks").First(&template, template.ID)

	response.Success(c, template)
}

// UpdateProjectTemplateRequest 更新项目模板请求
type UpdateProjectTemplateRequest struct {
	Name        string                              `json:"name" binding:"required"`
	Description string                              `json:"description"`
	Category    string                              `json:"category"`
	IsPublic    bool                                `json:"isPublic"`
	Tags        []string                            `json:"tags"`
	FieldSchema json.RawMessage                     `json:"fieldSchema"`
	Tasks       []CreateTemplateTaskRequest         `json:"tasks"`
}

// UpdateProjectTemplate 更新项目模板
func (h *TemplateHandler) UpdateProjectTemplate(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的模板 ID")
		return
	}

	var req UpdateProjectTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数格式错误")
		return
	}

	db := database.GetDB()

	var template models.ProjectTemplate
	if err := db.Where("deleted_at IS NULL").First(&template, id).Error; err != nil {
		response.NotFound(c, "模板不存在")
		return
	}

	// 处理标签
	var tagsStr *string
	if len(req.Tags) > 0 {
		tagsJSON, _ := json.Marshal(req.Tags)
		s := string(tagsJSON)
		tagsStr = &s
	}

	// 处理字段定义
	var fieldSchemaStr *string
	if len(req.FieldSchema) > 0 {
		s := string(req.FieldSchema)
		fieldSchemaStr = &s
	}

	// 更新基本信息
	updates := map[string]interface{}{
		"name":         req.Name,
		"description":  req.Description,
		"category":     req.Category,
		"is_public":    req.IsPublic,
		"tags":         tagsStr,
		"field_schema": fieldSchemaStr,
	}

	if err := db.Model(&template).Updates(updates).Error; err != nil {
		h.logger.Error("更新项目模板失败", zap.Error(err))
		response.Error(c, 500, "更新项目模板失败")
		return
	}

	// 删除旧的任务和子任务
	db.Where("template_id = ?", id).Delete(&models.ProjectTemplateTask{})

	// 创建新任务
	for _, taskReq := range req.Tasks {
		priority := taskReq.Priority
		if priority == "" {
			priority = "medium"
		}

		var depsStr *string
		if len(taskReq.Dependencies) > 0 {
			depsJSON, _ := json.Marshal(taskReq.Dependencies)
			s := string(depsJSON)
			depsStr = &s
		}

		task := models.ProjectTemplateTask{
			TemplateID:     template.ID,
			Title:          taskReq.Title,
			Description:    taskReq.Description,
			Priority:       priority,
			SortOrder:      taskReq.Order,
			EstimatedHours: taskReq.EstimatedHours,
			Dependencies:   depsStr,
		}

		if err := db.Create(&task).Error; err != nil {
			continue
		}

		for _, subtaskReq := range taskReq.Subtasks {
			subtask := models.ProjectTemplateSubtask{
				TemplateTaskID: task.ID,
				Title:          subtaskReq.Title,
				Description:    subtaskReq.Description,
				SortOrder:      subtaskReq.Order,
				EstimatedHours: subtaskReq.EstimatedHours,
			}
			db.Create(&subtask)
		}
	}

	// 重新加载
	db.Preload("Tasks.Subtasks").First(&template, template.ID)

	response.Success(c, template)
}

// DeleteProjectTemplate 删除项目模板（软删除）
func (h *TemplateHandler) DeleteProjectTemplate(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的模板 ID")
		return
	}

	db := database.GetDB()

	// 软删除：设置 deleted_at
	now := time.Now()
	if err := db.Model(&models.ProjectTemplate{}).Where("id = ? AND deleted_at IS NULL", id).Update("deleted_at", &now).Error; err != nil {
		h.logger.Error("删除项目模板失败", zap.Error(err))
		response.Error(c, 500, "删除项目模板失败")
		return
	}

	response.Success(c, gin.H{"success": true, "message": "删除成功"})
}

// InstantiateProjectTemplate 实例化项目模板
func (h *TemplateHandler) InstantiateProjectTemplate(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的模板 ID")
		return
	}

	var req struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数格式错误")
		return
	}

	db := database.GetDB()

	// 获取模板
	var template models.ProjectTemplate
	if err := db.Preload("Tasks.Subtasks").First(&template, id).Error; err != nil {
		response.NotFound(c, "模板不存在")
		return
	}

	// 创建需求
	requirement := models.Requirement{
		Title:    req.Name,
		Content:  req.Description,
		Status:   "pending",
		Priority: "medium",
	}

	if err := db.Create(&requirement).Error; err != nil {
		h.logger.Error("创建需求失败", zap.Error(err))
		response.Error(c, 500, "创建需求失败")
		return
	}

	// 创建任务
	taskIDs := make([]uint64, 0)
	taskMap := make(map[uint64]uint64) // template task id -> new task id

	for _, templateTask := range template.Tasks {
		task := models.Task{
			RequirementID: &requirement.ID,
			Title:         templateTask.Title,
			Description:   templateTask.Description,
			Status:        "pending",
			Priority:      templateTask.Priority,
			EstimatedHours: templateTask.EstimatedHours,
		}

		if err := db.Create(&task).Error; err != nil {
			continue
		}

		taskMap[templateTask.ID] = task.ID
		taskIDs = append(taskIDs, task.ID)

		// 创建子任务
		for _, templateSubtask := range templateTask.Subtasks {
			subtask := models.Subtask{
				TaskID:      task.ID,
				Title:       templateSubtask.Title,
				Description: templateSubtask.Description,
				Status:      "pending",
				SortOrder:   templateSubtask.SortOrder,
			}
			db.Create(&subtask)
		}
	}

	// 更新使用次数
	db.Model(&template).Update("usage_count", template.UsageCount+1)

	response.Success(c, gin.H{
		"requirementId": requirement.ID,
		"taskIds":       taskIDs,
	})
}

// ============ 任务模板 ============

// ListTaskTemplates 获取任务模板列表
func (h *TemplateHandler) ListTaskTemplates(c *gin.Context) {
	db := database.GetDB()

	var templates []models.TaskTemplate
	query := db.Model(&models.TaskTemplate{}).Where("deleted_at IS NULL")

	if keyword := c.Query("keyword"); keyword != "" {
		query = query.Where("name LIKE ? OR title LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}

	if err := query.Order("created_at DESC").Find(&templates).Error; err != nil {
		h.logger.Error("获取任务模板列表失败", zap.Error(err))
		response.Error(c, 500, "获取任务模板列表失败")
		return
	}

	response.Success(c, templates)
}

// GetTaskTemplate 获取任务模板详情
func (h *TemplateHandler) GetTaskTemplate(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的模板 ID")
		return
	}

	db := database.GetDB()
	var template models.TaskTemplate
	if err := db.Where("deleted_at IS NULL").First(&template, id).Error; err != nil {
		response.NotFound(c, "模板不存在")
		return
	}

	response.Success(c, template)
}

// CreateTaskTemplateRequest 创建任务模板请求
type CreateTaskTemplateRequest struct {
	Name            string    `json:"name" binding:"required"`
	Description     string    `json:"description"`
	Title           string    `json:"title" binding:"required"`
	TaskDescription string    `json:"taskDescription"`
	Priority        string    `json:"priority"`
	EstimatedHours  *float64  `json:"estimatedHours"`
	Subtasks        []string  `json:"subtasks"`
	Tags            []string  `json:"tags"`
	IsPublic        bool      `json:"isPublic"`
}

// CreateTaskTemplate 创建任务模板
func (h *TemplateHandler) CreateTaskTemplate(c *gin.Context) {
	var req CreateTaskTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数格式错误")
		return
	}

	db := database.GetDB()

	// 处理子任务
	var subtasksStr *string
	if len(req.Subtasks) > 0 {
		subtasksJSON, _ := json.Marshal(req.Subtasks)
		s := string(subtasksJSON)
		subtasksStr = &s
	}

	// 处理标签
	var tagsStr *string
	if len(req.Tags) > 0 {
		tagsJSON, _ := json.Marshal(req.Tags)
		s := string(tagsJSON)
		tagsStr = &s
	}

	priority := req.Priority
	if priority == "" {
		priority = "medium"
	}

	template := models.TaskTemplate{
		Name:            req.Name,
		Description:     req.Description,
		Title:           req.Title,
		TaskDescription: req.TaskDescription,
		Priority:        priority,
		EstimatedHours:  req.EstimatedHours,
		Subtasks:        subtasksStr,
		Tags:            tagsStr,
		IsPublic:        req.IsPublic,
		UsageCount:      0,
	}

	if err := db.Create(&template).Error; err != nil {
		h.logger.Error("创建任务模板失败", zap.Error(err))
		response.Error(c, 500, "创建任务模板失败")
		return
	}

	response.Success(c, template)
}

// UpdateTaskTemplate 更新任务模板
func (h *TemplateHandler) UpdateTaskTemplate(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的模板 ID")
		return
	}

	var req CreateTaskTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数格式错误")
		return
	}

	db := database.GetDB()

	var template models.TaskTemplate
	if err := db.Where("deleted_at IS NULL").First(&template, id).Error; err != nil {
		response.NotFound(c, "模板不存在")
		return
	}

	// 处理子任务
	var subtasksStr *string
	if len(req.Subtasks) > 0 {
		subtasksJSON, _ := json.Marshal(req.Subtasks)
		s := string(subtasksJSON)
		subtasksStr = &s
	}

	// 处理标签
	var tagsStr *string
	if len(req.Tags) > 0 {
		tagsJSON, _ := json.Marshal(req.Tags)
		s := string(tagsJSON)
		tagsStr = &s
	}

	priority := req.Priority
	if priority == "" {
		priority = "medium"
	}

	updates := map[string]interface{}{
		"name":             req.Name,
		"description":      req.Description,
		"title":            req.Title,
		"task_description": req.TaskDescription,
		"priority":         priority,
		"estimated_hours":  req.EstimatedHours,
		"subtasks":         subtasksStr,
		"tags":             tagsStr,
		"is_public":        req.IsPublic,
	}

	if err := db.Model(&template).Updates(updates).Error; err != nil {
		h.logger.Error("更新任务模板失败", zap.Error(err))
		response.Error(c, 500, "更新任务模板失败")
		return
	}

	db.First(&template, template.ID)

	response.Success(c, template)
}

// DeleteTaskTemplate 删除任务模板（软删除）
func (h *TemplateHandler) DeleteTaskTemplate(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的模板 ID")
		return
	}

	db := database.GetDB()

	// 软删除：设置 deleted_at
	now := time.Now()
	if err := db.Model(&models.TaskTemplate{}).Where("id = ? AND deleted_at IS NULL", id).Update("deleted_at", &now).Error; err != nil {
		h.logger.Error("删除任务模板失败", zap.Error(err))
		response.Error(c, 500, "删除任务模板失败")
		return
	}

	response.Success(c, gin.H{"success": true, "message": "删除成功"})
}

// InstantiateTaskTemplate 实例化任务模板
func (h *TemplateHandler) InstantiateTaskTemplate(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的模板 ID")
		return
	}

	var req struct {
		Title         string `json:"title"`
		RequirementID *uint64 `json:"requirementId"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数格式错误")
		return
	}

	db := database.GetDB()

	// 获取模板
	var template models.TaskTemplate
	if err := db.First(&template, id).Error; err != nil {
		response.NotFound(c, "模板不存在")
		return
	}

	title := req.Title
	if title == "" {
		title = template.Title
	}

	// 创建任务
	task := models.Task{
		RequirementID:   req.RequirementID,
		Title:           title,
		Description:     template.TaskDescription,
		Status:          "pending",
		Priority:        template.Priority,
		EstimatedHours:  template.EstimatedHours,
	}

	if err := db.Create(&task).Error; err != nil {
		h.logger.Error("创建任务失败", zap.Error(err))
		response.Error(c, 500, "创建任务失败")
		return
	}

	// 创建子任务
	if template.Subtasks != nil {
		var subtaskTitles []string
		if err := json.Unmarshal([]byte(*template.Subtasks), &subtaskTitles); err == nil {
			for i, subtaskTitle := range subtaskTitles {
				subtask := models.Subtask{
					TaskID:    task.ID,
					Title:     subtaskTitle,
					Status:    "pending",
					SortOrder: uint(i),
				}
				db.Create(&subtask)
			}
		}
	}

	// 更新使用次数
	db.Model(&template).Update("usage_count", template.UsageCount+1)

	response.Success(c, gin.H{
		"taskId": task.ID,
	})
}

// ============ 项目模板打分 ============

// ScoreTemplateRequest 评分请求
type ScoreTemplateRequest struct {
	ID uint64 `json:"id" binding:"required"`
}

// ScoreProjectTemplate 对项目模板进行打分（同步版本）
func (h *TemplateHandler) ScoreProjectTemplate(c *gin.Context) {
	var req ScoreTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误：需要 id")
		return
	}

	db := database.GetDB()
	var template models.ProjectTemplate
	if err := db.Where("deleted_at IS NULL").Preload("Tasks.Subtasks").First(&template, req.ID).Error; err != nil {
		response.NotFound(c, "模板不存在")
		return
	}

	// 构建评分 Prompt
	prompt := h.buildScorePrompt(&template)

	// 调用 AI 进行评分
	aiResponse, err := h.aiSvc.Chat(prompt)
	if err != nil {
		h.logger.Error("AI 评分失败", zap.Uint64("templateID", req.ID), zap.Error(err))
		response.Error(c, 500, "AI 评分失败："+err.Error())
		return
	}

	// 解析评分结果
	scoreResult, totalScore, err := h.parseScoreResponse(aiResponse)
	if err != nil {
		h.logger.Error("解析评分结果失败", zap.Uint64("templateID", req.ID), zap.Error(err))
		response.Error(c, 500, "解析评分结果失败："+err.Error())
		return
	}

	// 保存评分结果到模板
	scoreJSON, _ := json.Marshal(gin.H{
		"totalScore":  totalScore,
		"scores":      scoreResult.Scores,
		"strengths":   scoreResult.Strengths,
		"weaknesses":  scoreResult.Weaknesses,
		"suggestions": scoreResult.Suggestions,
		"analysis":    scoreResult.Analysis,
		"scoredAt":    time.Now().Format(time.RFC3339),
	})
	scoreStr := string(scoreJSON)
	db.Model(&template).Update("last_score", &scoreStr)

	response.Success(c, gin.H{
		"score":       scoreResult,
		"totalScore":  totalScore,
		"evaluation":  aiResponse,
	})
}

// ScoreProjectTemplateAsync 异步对项目模板进行打分
func (h *TemplateHandler) ScoreProjectTemplateAsync(c *gin.Context) {
	var req ScoreTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误：需要 id")
		return
	}

	db := database.GetDB()
	var template models.ProjectTemplate
	if err := db.Where("deleted_at IS NULL").Preload("Tasks.Subtasks").First(&template, req.ID).Error; err != nil {
		response.NotFound(c, "模板不存在")
		return
	}

	// 创建消息记录
	title := "项目模板评分"
	content := template.Name
	message := models.Message{
		TaskID:  nil,
		Type:    "score_template",
		Status:  "processing",
		Title:   title,
		Content: &content,
	}
	if err := db.Create(&message).Error; err != nil {
		h.logger.Error("创建消息记录失败", zap.Error(err))
		response.Error(c, 500, "创建消息记录失败")
		return
	}

	// 异步执行评分
	go func(templateID uint64) {
		defer func() {
			// 完成后清理
		}()

		// 重新获取模板（确保在 goroutine 中使用最新数据）
		var tmpl models.ProjectTemplate
		if err := db.Where("deleted_at IS NULL").Preload("Tasks.Subtasks").First(&tmpl, templateID).Error; err != nil {
			h.logger.Error("获取模板失败", zap.Uint64("templateID", templateID), zap.Error(err))
			return
		}

		// 构建评分提示词
		prompt := h.buildScorePrompt(&tmpl)

		// 调用 AI 进行评分
		aiResponse, err := h.aiSvc.Chat(prompt)
		if err != nil {
			h.logger.Error("异步评分失败", zap.Uint64("templateID", templateID), zap.Error(err))
			// 更新消息状态为失败
			errMsg := err.Error()
			db.Model(&message).Updates(map[string]interface{}{
				"status":        "failed",
				"error_message": &errMsg,
			})
		} else {
			// 解析评分结果
			scoreResult, totalScore, _ := h.parseScoreResponse(aiResponse)
			resultSummary := fmt.Sprintf("评分完成，总分：%.1f", totalScore)

			// 保存评分结果到模板
			scoreJSON, _ := json.Marshal(gin.H{
				"totalScore":  totalScore,
				"scores":      scoreResult.Scores,
				"strengths":   scoreResult.Strengths,
				"weaknesses":  scoreResult.Weaknesses,
				"suggestions": scoreResult.Suggestions,
				"analysis":    scoreResult.Analysis,
				"scoredAt":    time.Now().Format(time.RFC3339),
			})
			scoreStr := string(scoreJSON)
			db.Model(&tmpl).Update("last_score", &scoreStr)

			// 更新消息状态为成功
			db.Model(&message).Updates(map[string]interface{}{
				"status":         "success",
				"result_summary": &resultSummary,
			})
		}
	}(req.ID)

	response.Success(c, gin.H{
		"messageId": message.ID,
		"message":   "评分已开始，完成后会通知您",
	})
}

// ScoreResult 模板评分结果
type ScoreResult struct {
	Scores      struct {
		Clarity       int `json:"clarity"`
		Completeness  int `json:"completeness"`
		Structure     int `json:"structure"`
		Actionability int `json:"actionability"`
		Consistency   int `json:"consistency"`
	} `json:"scores"`
	TotalScore  float64 `json:"totalScore"`
	Strengths   []string `json:"strengths"`
	Weaknesses  []string `json:"weaknesses"`
	Suggestions []struct {
		Issue      string `json:"issue"`
		Suggestion string `json:"suggestion"`
	} `json:"suggestions"`
	Analysis    string   `json:"analysis"`
}

// buildScorePrompt 构建评分 Prompt
func (h *TemplateHandler) buildScorePrompt(template *models.ProjectTemplate) string {
	tasksInfo := ""
	for i, task := range template.Tasks {
		tasksInfo += strconv.Itoa(i+1) + ". " + task.Title + "\n"
		if task.Description != "" {
			tasksInfo += "   描述：" + task.Description + "\n"
		}
		tasksInfo += "   优先级：" + task.Priority + "\n"
		if task.EstimatedHours != nil {
			tasksInfo += "   预估工时：" + strconv.FormatFloat(*task.EstimatedHours, 'f', 1, 64) + "h\n"
		}
		if len(task.Subtasks) > 0 {
			tasksInfo += "   子任务:\n"
			for _, st := range task.Subtasks {
				tasksInfo += "     - " + st.Title + "\n"
			}
		}
	}

	return `你是一位资深提示词工程师和项目管理专家，请对以下项目模板进行质量评分。

【模板信息】
- 名称：` + template.Name + `
- 描述：` + template.Description + `
- 分类：` + template.Category + `
- 任务数量：` + strconv.Itoa(len(template.Tasks)) + `

【任务列表】
` + tasksInfo + `

【评分维度】
请从以下 5 个维度进行评分（每项 1-10 分）：

1. 清晰度 (Clarity)：模板描述是否清晰明确，无歧义
   - 模板名称是否能准确反映模板用途
   - 描述是否清晰
   - 任务划分是否容易理解

2. 完整性 (Completeness)：是否包含了必要的上下文和信息
   - 是否有完整的任务列表
   - 是否有明确的优先级
   - 是否有合理的工时预估

3. 结构化 (Structure)：模板结构是否清晰、层次分明
   - 任务划分是否有条理
   - 子任务划分是否合理
   - 依赖关系是否清晰

4. 可执行性 (Actionability)：用户是否能基于模板直接创建项目
   - 任务是否有具体的描述
   - 是否有明确的验收标准
   - 是否有可参考的实现细节

5. 一致性 (Consistency)：与项目规范的兼容性
   - 任务优先级设置是否合理
   - 工时预估是否准确
   - 整体风格是否一致

【评分标准】
- 9-10 分：优秀，几乎完美
- 7-8 分：良好，有小幅改进空间
- 5-6 分：一般，需要明显改进
- 3-4 分：较差，存在严重问题
- 1-2 分：很差，几乎无法使用

请以 JSON 格式返回评分结果：
{
  "scores": {
    "clarity": 1-10,
    "completeness": 1-10,
    "structure": 1-10,
    "actionability": 1-10,
    "consistency": 1-10
  },
  "totalScore": 总分 (5 项平均分*10，保留 1 位小数),
  "strengths": ["优点 1", "优点 2", ...],
  "weaknesses": ["缺点 1", "缺点 2", ...],
  "suggestions": [
    {"issue": "问题描述", "suggestion": "改进建议"}
  ],
  "analysis": "详细分析报告（200-500 字）"
}

请只返回 JSON，不要包含其他内容。`
}

// parseScoreResponse 解析评分响应
func (h *TemplateHandler) parseScoreResponse(responseText string) (*ScoreResult, float64, error) {
	cleanedResponse := responseText
	for _, marker := range []string{"```json", "```"} {
		cleanedResponse = replaceAllStr(cleanedResponse, marker, "")
	}
	cleanedResponse = trimWhitespaceStr(cleanedResponse)

	var result ScoreResult
	if err := json.Unmarshal([]byte(cleanedResponse), &result); err != nil {
		return nil, 0, err
	}

	return &result, result.TotalScore, nil
}

// replaceAllStr 替换所有匹配
func replaceAllStr(s, old, new string) string {
	return strings.ReplaceAll(s, old, new)
}

// trimWhitespaceStr 修剪空白
func trimWhitespaceStr(s string) string {
	return strings.TrimSpace(s)
}
