package handlers

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/ai-task-manager/backend/internal/config"
	"github.com/ai-task-manager/backend/internal/database"
	"github.com/ai-task-manager/backend/internal/models"
	"github.com/ai-task-manager/backend/internal/repository"
	"github.com/ai-task-manager/backend/pkg/ai"
	"github.com/ai-task-manager/backend/pkg/response"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// RequirementHandler 需求处理器
type RequirementHandler struct {
	logger    *zap.Logger
	aiService *ai.Service
}

// NewRequirementHandler 创建需求处理器
func NewRequirementHandler(logger *zap.Logger, cfg *config.Config) *RequirementHandler {
	var aiSvc *ai.Service
	if cfg != nil {
		aiSvc = ai.NewService(&cfg.AI)
	}
	return &RequirementHandler{
		logger:    logger,
		aiService: aiSvc,
	}
}

// List 获取需求列表
func (h *RequirementHandler) List(c *gin.Context) {
	db := database.GetDB()

	// 解析分页参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	// 构建查询 - 排除已删除的记录
	query := db.Model(&models.Requirement{}).Where("deleted_at IS NULL")

	// 筛选条件
	if status := c.Query("status"); status != "" {
		query = query.Where("status = ?", status)
	}
	if priority := c.Query("priority"); priority != "" {
		query = query.Where("priority = ?", priority)
	}
	if assignee := c.Query("assignee"); assignee != "" {
		query = query.Where("assignee = ?", assignee)
	}
	if keyword := c.Query("keyword"); keyword != "" {
		query = query.Where("title LIKE ? OR content LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}

	// 计算总数
	var total int64
	if err := query.Count(&total).Error; err != nil {
		h.logger.Error("获取需求总数失败", zap.Error(err))
		response.Error(c, 500, "获取需求列表失败")
		return
	}

	// 分页查询
	var requirements []models.Requirement
	offset := (page - 1) * pageSize
	if err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&requirements).Error; err != nil {
		h.logger.Error("获取需求列表失败", zap.Error(err))
		response.Error(c, 500, "获取需求列表失败")
		return
	}

	response.SuccessPage(c, requirements, total, page, pageSize)
}

// Statistics 获取需求统计
func (h *RequirementHandler) Statistics(c *gin.Context) {
	db := database.GetDB()

	stats := struct {
		Total     int64 `json:"total"`
		Draft     int64 `json:"draft"`
		Reviewing int64 `json:"reviewing"`
		Approved  int64 `json:"approved"`
		Completed int64 `json:"completed"`
	}{}

	db.Model(&models.Requirement{}).Where("deleted_at IS NULL").Count(&stats.Total)
	db.Model(&models.Requirement{}).Where("status = ? AND deleted_at IS NULL", "draft").Count(&stats.Draft)
	db.Model(&models.Requirement{}).Where("status = ? AND deleted_at IS NULL", "reviewing").Count(&stats.Reviewing)
	db.Model(&models.Requirement{}).Where("status = ? AND deleted_at IS NULL", "approved").Count(&stats.Approved)
	db.Model(&models.Requirement{}).Where("status = ? AND deleted_at IS NULL", "completed").Count(&stats.Completed)

	response.Success(c, stats)
}

// Get 获取需求详情
func (h *RequirementHandler) Get(c *gin.Context) {
	db := database.GetDB()

	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的需求 ID")
		return
	}

	var requirement models.Requirement
	if err := db.Where("deleted_at IS NULL").First(&requirement, id).Error; err != nil {
		response.NotFound(c, "需求不存在")
		return
	}

	// 加载关联的文档
	var documents []models.RequirementDocument
	db.Where("requirement_id = ?", id).Find(&documents)
	requirement.Documents = documents

	response.Success(c, requirement)
}

// Create 创建需求
func (h *RequirementHandler) Create(c *gin.Context) {
	db := database.GetDB()

	var req models.Requirement
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数格式错误")
		return
	}

	// 设置默认值
	if req.Status == "" {
		req.Status = "draft"
	}
	if req.Priority == "" {
		req.Priority = "medium"
	}

	if err := db.Create(&req).Error; err != nil {
		h.logger.Error("创建需求失败", zap.Error(err))
		response.Error(c, 500, "创建需求失败")
		return
	}

	response.Success(c, req)
}

// Update 更新需求
func (h *RequirementHandler) Update(c *gin.Context) {
	db := database.GetDB()

	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的需求 ID")
		return
	}

	var req models.Requirement
	if err := db.Where("deleted_at IS NULL").First(&req, id).Error; err != nil {
		response.NotFound(c, "需求不存在")
		return
	}

	var updates map[string]interface{}
	if err := c.ShouldBindJSON(&updates); err != nil {
		response.BadRequest(c, "请求参数格式错误")
		return
	}

	if err := db.Model(&req).Updates(updates).Error; err != nil {
		h.logger.Error("更新需求失败", zap.Error(err))
		response.Error(c, 500, "更新需求失败")
		return
	}

	// 重新获取更新后的数据
	db.First(&req, id)
	response.Success(c, req)
}

// Delete 删除需求
func (h *RequirementHandler) Delete(c *gin.Context) {
	db := database.GetDB()

	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的需求 ID")
		return
	}

	// 软删除：设置 deleted_at
	now := time.Now()
	if err := db.Model(&models.Requirement{}).Where("id = ? AND deleted_at IS NULL", id).Update("deleted_at", &now).Error; err != nil {
		h.logger.Error("删除需求失败", zap.Error(err))
		response.Error(c, 500, "删除需求失败")
		return
	}

	response.Success(c, nil)
}

// UploadDocument 上传文档
func (h *RequirementHandler) UploadDocument(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的需求 ID")
		return
	}

	db := database.GetDB()

	// 验证需求是否存在
	var requirement models.Requirement
	if err := db.First(&requirement, id).Error; err != nil {
		response.NotFound(c, "需求不存在")
		return
	}

	// 获取上传的文件
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		response.BadRequest(c, "请选择要上传的文件")
		return
	}
	defer file.Close()

	// 验证文件大小（最大 50MB）
	const maxSize = 50 * 1024 * 1024
	if header.Size > maxSize {
		response.BadRequest(c, "文件大小不能超过 50MB")
		return
	}

	// 获取文件扩展名并验证
	ext := strings.ToLower(filepath.Ext(header.Filename))
	allowedExts := map[string]bool{
		".pdf": true, ".doc": true, ".docx": true, ".xls": true, ".xlsx": true,
		".ppt": true, ".pptx": true, ".txt": true, ".md": true,
		".png": true, ".jpg": true, ".jpeg": true, ".gif": true, ".bmp": true,
	}
	if !allowedExts[ext] {
		response.BadRequest(c, "不支持的文件类型")
		return
	}

	// 创建上传目录
	uploadDir := filepath.Join("uploads", "documents", strconv.FormatUint(id, 10))
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		h.logger.Error("创建上传目录失败", zap.Error(err))
		response.Error(c, 500, "上传失败")
		return
	}

	// 生成唯一文件名
	filename := fmt.Sprintf("%d_%d%s", time.Now().UnixNano(), id, ext)
	filePath := filepath.Join(uploadDir, filename)

	// 创建目标文件
	dst, err := os.Create(filePath)
	if err != nil {
		h.logger.Error("创建文件失败", zap.Error(err))
		response.Error(c, 500, "上传失败")
		return
	}
	defer dst.Close()

	// 复制文件内容
	if _, err := io.Copy(dst, file); err != nil {
		h.logger.Error("保存文件失败", zap.Error(err))
		response.Error(c, 500, "上传失败")
		return
	}

	// 获取 MIME 类型
	mimeType := getMimeType(ext)

	// 保存文档记录到数据库
	doc := models.RequirementDocument{
		RequirementID: id,
		Name:          header.Filename,
		Path:          filePath,
		Size:          uint64(header.Size),
		MimeType:      &mimeType,
	}

	if err := db.Create(&doc).Error; err != nil {
		// 删除已上传的文件
		os.Remove(filePath)
		h.logger.Error("保存文档记录失败", zap.Error(err))
		response.Error(c, 500, "上传失败")
		return
	}

	response.Success(c, doc)
}

// DeleteDocument 删除文档
func (h *RequirementHandler) DeleteDocument(c *gin.Context) {
	db := database.GetDB()

	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的需求 ID")
		return
	}
	docId, err := strconv.ParseUint(c.Param("docId"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的文档 ID")
		return
	}

	if err := db.Where("id = ? AND requirement_id = ?", docId, id).Delete(&models.RequirementDocument{}).Error; err != nil {
		h.logger.Error("删除文档失败", zap.Error(err))
		response.Error(c, 500, "删除文档失败")
		return
	}

	response.Success(c, nil)
}

// DownloadDocument 下载文档
func (h *RequirementHandler) DownloadDocument(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的需求 ID")
		return
	}
	docId, err := strconv.ParseUint(c.Param("docId"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的文档 ID")
		return
	}

	db := database.GetDB()

	// 查询文档记录
	var doc models.RequirementDocument
	if err := db.Where("id = ? AND requirement_id = ?", docId, id).First(&doc).Error; err != nil {
		response.NotFound(c, "文档不存在")
		return
	}

	// 检查文件是否存在
	if _, err := os.Stat(doc.Path); os.IsNotExist(err) {
		response.NotFound(c, "文件不存在")
		return
	}

	// 设置响应头
	c.Header("Content-Description", "File Transfer")
	c.Header("Content-Transfer-Encoding", "binary")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", doc.Name))
	c.Header("Content-Type", "application/octet-stream")
	c.Header("Content-Length", strconv.FormatUint(doc.Size, 10))

	// 发送文件
	c.File(doc.Path)
}

// getMimeType 根据文件扩展名获取 MIME 类型
func getMimeType(ext string) string {
	mimeTypes := map[string]string{
		".pdf":  "application/pdf",
		".doc":  "application/msword",
		".docx": "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
		".xls":  "application/vnd.ms-excel",
		".xlsx": "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
		".ppt":  "application/vnd.ms-powerpoint",
		".pptx": "application/vnd.openxmlformats-officedocument.presentationml.presentation",
		".txt":  "text/plain",
		".md":   "text/markdown",
		".png":  "image/png",
		".jpg":  "image/jpeg",
		".jpeg": "image/jpeg",
		".gif":  "image/gif",
		".bmp":  "image/bmp",
	}
	if mt, ok := mimeTypes[ext]; ok {
		return mt
	}
	return "application/octet-stream"
}

// SplitTasksRequest 拆分任务请求
type SplitTasksRequest struct {
	TaskType         string  `json:"taskType"`            // frontend, backend, fullstack
	LanguageID       uint64  `json:"languageId"`          // 语言 ID
	ProjectTemplateID *uint64 `json:"projectTemplateId"`  // 项目模板 ID（可选）
}

// SplitTasks 需求拆分为任务（AI）- 同步版本
func (h *RequirementHandler) SplitTasks(c *gin.Context) {
	db := database.GetDB()

	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的需求 ID")
		return
	}

	// 解析请求体
	var req SplitTasksRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		// 如果没有请求体，使用默认值
		req.TaskType = "backend"
	}

	// 验证任务类型
	if req.TaskType != "frontend" && req.TaskType != "backend" && req.TaskType != "fullstack" {
		req.TaskType = "backend"
	}

	// 获取需求详情
	var requirement models.Requirement
	if err := db.First(&requirement, id).Error; err != nil {
		response.NotFound(c, "需求不存在")
		return
	}

	// 检查 AI 服务是否可用
	if h.aiService == nil {
		response.Error(c, 500, "AI 服务未配置")
		return
	}

	// 调用 AI 服务拆分需求
	tasksWithDeps, err := h.aiService.SplitRequirement(&requirement, req.TaskType)
	if err != nil {
		h.logger.Error("AI 拆分需求失败", zap.Error(err))
		response.Error(c, 500, "AI 拆分需求失败: "+err.Error())
		return
	}

	// 提取任务并保存到数据库
	var tasks []models.Task
	if len(tasksWithDeps) > 0 {
		tasks = make([]models.Task, len(tasksWithDeps))
		for i, twd := range tasksWithDeps {
			tasks[i] = twd.Task
		}

		if err := db.Create(&tasks).Error; err != nil {
			h.logger.Error("保存任务失败", zap.Error(err))
			response.Error(c, 500, "保存任务失败")
			return
		}

		// 保存任务依赖关系
		for i, twd := range tasksWithDeps {
			if len(twd.Dependencies) > 0 {
				for _, depIndex := range twd.Dependencies {
					if depIndex >= 0 && depIndex < len(tasks) && depIndex != i {
						dependency := models.TaskDependency{
							TaskID:         tasks[i].ID,
							DependsOnTaskID: tasks[depIndex].ID,
						}
						if err := db.Create(&dependency).Error; err != nil {
							h.logger.Warn("保存任务依赖失败", zap.Error(err))
						}
					}
				}
			}
		}
	}

	// 更新需求状态为已拆分
	db.Model(&requirement).Update("status", "active")

	response.Success(c, gin.H{
		"success": true,
		"message": "成功拆分为任务",
		"tasks":   tasks,
	})
}

// SplitTasksAsync 需求拆分为任务（AI）- 异步版本
func (h *RequirementHandler) SplitTasksAsync(c *gin.Context) {
	db := database.GetDB()

	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的需求 ID")
		return
	}

	// 解析请求体
	var req SplitTasksRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		// 如果没有请求体，使用默认值
		req.TaskType = "backend"
	}

	// 验证任务类型
	if req.TaskType != "frontend" && req.TaskType != "backend" && req.TaskType != "fullstack" {
		req.TaskType = "backend"
	}

	// 获取语言信息
	var language *models.Language
	if req.LanguageID > 0 {
		language = &models.Language{}
		if err := db.First(language, req.LanguageID).Error; err != nil {
			language = nil
		}
	}

	// 获取项目模板的字段定义（如果指定了模板）
	var fieldSchema *string
	if req.ProjectTemplateID != nil && *req.ProjectTemplateID > 0 {
		var template models.ProjectTemplate
		if err := db.Select("field_schema").First(&template, *req.ProjectTemplateID).Error; err == nil {
			fieldSchema = template.FieldSchema
		}
	}

	// 获取需求详情
	var requirement models.Requirement
	if err := db.First(&requirement, id).Error; err != nil {
		response.NotFound(c, "需求不存在")
		return
	}

	// 检查 AI 服务是否可用
	if h.aiService == nil {
		response.Error(c, 500, "AI 服务未配置")
		return
	}

	// 创建消息记录
	title := "需求拆分任务"
	content := requirement.Title
	reqID := requirement.ID
	message := models.Message{
		RequirementID: &reqID, // 关联需求ID
		Type:          "split_requirement",
		Status:        "processing",
		Title:         title,
		Content:       &content,
	}
	if err := db.Create(&message).Error; err != nil {
		h.logger.Error("创建消息记录失败", zap.Error(err))
		response.Error(c, 500, "创建消息记录失败")
		return
	}

	// 异步执行拆分
	go func() {
		// 调用 AI 服务拆分需求（传入语言信息和模板字段定义）
		tasksWithDeps, err := h.aiService.SplitRequirementWithTemplate(&requirement, req.TaskType, language, fieldSchema)
		if err != nil {
			h.logger.Error("AI 拆分需求失败", zap.Error(err))
			// 更新消息状态为失败
			errMsg := err.Error()
			db.Model(&message).Updates(map[string]interface{}{
				"status":        "failed",
				"error_message": &errMsg,
			})
			return
		}

		// 提取任务并保存到数据库
		if len(tasksWithDeps) > 0 {
			tasks := make([]models.Task, len(tasksWithDeps))
			for i, twd := range tasksWithDeps {
				tasks[i] = twd.Task
			}

			if err := db.Create(&tasks).Error; err != nil {
				h.logger.Error("保存任务失败", zap.Error(err))
				// 更新消息状态为失败
				errMsg := "保存任务失败"
				db.Model(&message).Updates(map[string]interface{}{
					"status":        "failed",
					"error_message": &errMsg,
				})
				return
			}

			// 保存任务依赖关系
			for i, twd := range tasksWithDeps {
				if len(twd.Dependencies) > 0 {
					for _, depIndex := range twd.Dependencies {
						if depIndex >= 0 && depIndex < len(tasks) && depIndex != i {
							dependency := models.TaskDependency{
								TaskID:         tasks[i].ID,
								DependsOnTaskID: tasks[depIndex].ID,
							}
							if err := db.Create(&dependency).Error; err != nil {
								h.logger.Warn("保存任务依赖失败", zap.Error(err), zap.Uint64("taskId", tasks[i].ID), zap.Int("depIndex", depIndex))
							}
						}
					}
				}
			}
		}

		// 更新需求状态为已拆分
		db.Model(&requirement).Update("status", "active")

		// 更新消息状态为成功
		resultSummary := fmt.Sprintf("成功拆分为 %d 个任务", len(tasksWithDeps))
		db.Model(&message).Updates(map[string]interface{}{
			"status":         "success",
			"result_summary": &resultSummary,
		})
	}()

	response.Success(c, gin.H{
		"messageId": message.ID,
		"message":   "需求拆分已开始，完成后会通知您",
	})
}

// GetRequirementStructure 获取需求及其任务和子任务的结构化数据
// GET /api/requirements/:id/structure
func (h *RequirementHandler) GetRequirementStructure(c *gin.Context) {
	requirementID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的需求 ID")
		return
	}

	// 使用 taskService 获取结构化数据
	taskRepo := repository.NewTaskRepository()
	tree, err := taskRepo.GetRequirementWithTasksAndSubtasks(requirementID)
	if err != nil {
		h.logger.Error("获取需求结构化数据失败", zap.Uint64("requirementID", requirementID), zap.Error(err))
		response.NotFound(c, "需求不存在")
		return
	}

	response.Success(c, tree)
}
