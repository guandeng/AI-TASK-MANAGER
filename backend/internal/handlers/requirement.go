package handlers

import (
	"strconv"

	"github.com/ai-task-manager/backend/internal/database"
	"github.com/ai-task-manager/backend/internal/models"
	"github.com/ai-task-manager/backend/pkg/response"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// RequirementHandler 需求处理器
type RequirementHandler struct {
	logger *zap.Logger
}

// NewRequirementHandler 创建需求处理器
func NewRequirementHandler(logger *zap.Logger) *RequirementHandler {
	return &RequirementHandler{logger: logger}
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

	// 构建查询
	query := db.Model(&models.Requirement{})

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

	db.Model(&models.Requirement{}).Count(&stats.Total)
	db.Model(&models.Requirement{}).Where("status = ?", "draft").Count(&stats.Draft)
	db.Model(&models.Requirement{}).Where("status = ?", "reviewing").Count(&stats.Reviewing)
	db.Model(&models.Requirement{}).Where("status = ?", "approved").Count(&stats.Approved)
	db.Model(&models.Requirement{}).Where("status = ?", "completed").Count(&stats.Completed)

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
	if err := db.First(&requirement, id).Error; err != nil {
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
	if err := db.First(&req, id).Error; err != nil {
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

	if err := db.Delete(&models.Requirement{}, id).Error; err != nil {
		h.logger.Error("删除需求失败", zap.Error(err))
		response.Error(c, 500, "删除需求失败")
		return
	}

	response.Success(c, nil)
}

// UploadDocument 上传文档
func (h *RequirementHandler) UploadDocument(c *gin.Context) {
	// TODO: 实现文件上传
	response.Success(c, nil)
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
	// TODO: 实现文件下载
	response.Success(c, nil)
}

// SplitTasks 需求拆分为任务（AI）
func (h *RequirementHandler) SplitTasks(c *gin.Context) {
	// TODO: 实现 AI 拆分
	response.Success(c, nil)
}
