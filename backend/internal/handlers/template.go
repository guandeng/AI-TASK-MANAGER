package handlers

import (
	"encoding/json"
	"strconv"

	"github.com/ai-task-manager/backend/internal/database"
	"github.com/ai-task-manager/backend/internal/models"
	"github.com/ai-task-manager/backend/pkg/response"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// TemplateHandler 模板处理器
type TemplateHandler struct {
	logger *zap.Logger
}

// NewTemplateHandler 创建模板处理器
func NewTemplateHandler(logger *zap.Logger) *TemplateHandler {
	return &TemplateHandler{
		logger: logger,
	}
}

// ============ 项目模板 ============

// ListProjectTemplates 获取项目模板列表
func (h *TemplateHandler) ListProjectTemplates(c *gin.Context) {
	db := database.GetDB()

	var templates []models.ProjectTemplate
	query := db.Model(&models.ProjectTemplate{})

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
	if err := db.Preload("Tasks.Subtasks").First(&template, id).Error; err != nil {
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

	// 创建模板
	template := models.ProjectTemplate{
		Name:        req.Name,
		Description: req.Description,
		Category:    req.Category,
		IsPublic:    req.IsPublic,
		Tags:        tagsStr,
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

// UpdateProjectTemplate 更新项目模板
func (h *TemplateHandler) UpdateProjectTemplate(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的模板 ID")
		return
	}

	var req CreateProjectTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数格式错误")
		return
	}

	db := database.GetDB()

	var template models.ProjectTemplate
	if err := db.First(&template, id).Error; err != nil {
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

	// 更新基本信息
	updates := map[string]interface{}{
		"name":        req.Name,
		"description": req.Description,
		"category":    req.Category,
		"is_public":   req.IsPublic,
		"tags":        tagsStr,
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

// DeleteProjectTemplate 删除项目模板
func (h *TemplateHandler) DeleteProjectTemplate(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的模板 ID")
		return
	}

	db := database.GetDB()

	if err := db.Delete(&models.ProjectTemplate{}, id).Error; err != nil {
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
	query := db.Model(&models.TaskTemplate{})

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
	if err := db.First(&template, id).Error; err != nil {
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
	if err := db.First(&template, id).Error; err != nil {
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

// DeleteTaskTemplate 删除任务模板
func (h *TemplateHandler) DeleteTaskTemplate(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的模板 ID")
		return
	}

	db := database.GetDB()

	if err := db.Delete(&models.TaskTemplate{}, id).Error; err != nil {
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
