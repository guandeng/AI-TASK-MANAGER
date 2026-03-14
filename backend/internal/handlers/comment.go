package handlers

import (
	"encoding/json"
	"strconv"

	"github.com/ai-task-manager/backend/internal/database"
	"github.com/ai-task-manager/backend/internal/models"
	"github.com/ai-task-manager/backend/pkg/response"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// CommentHandler 评论处理器
type CommentHandler struct {
	logger *zap.Logger
}

// NewCommentHandler 创建评论处理器
func NewCommentHandler(logger *zap.Logger) *CommentHandler {
	return &CommentHandler{
		logger: logger,
	}
}

// List 获取任务评论列表
func (h *CommentHandler) List(c *gin.Context) {
	taskID, err := strconv.ParseUint(c.Param("taskId"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的任务 ID")
		return
	}

	db := database.GetDB()

	// 获取查询参数
	subtaskID := c.Query("subtaskId")
	limit := c.Query("limit")

	query := db.Where("task_id = ?", taskID)

	// 如果指定了子任务ID
	if subtaskID != "" {
		if sid, err := strconv.ParseUint(subtaskID, 10, 64); err == nil {
			query = query.Where("subtask_id = ?", sid)
		}
	}

	// 只获取顶级评论（没有父评论的）
	query = query.Where("parent_id IS NULL")

	// 限制数量
	if limit != "" {
		if l, err := strconv.Atoi(limit); err == nil && l > 0 {
			query = query.Limit(l)
		}
	}

	var comments []models.Comment
	if err := query.Preload("Member").Order("created_at DESC").Find(&comments).Error; err != nil {
		h.logger.Error("获取评论列表失败", zap.Error(err))
		response.Error(c, 500, "获取评论列表失败")
		return
	}

	// 获取每个评论的回复数量
	for i := range comments {
		var replyCount int64
		db.Model(&models.Comment{}).Where("parent_id = ?", comments[i].ID).Count(&replyCount)
		comments[i].Replies = []models.Comment{} // 初始化空数组
	}

	response.Success(c, comments)
}

// GetTree 获取评论树形结构
func (h *CommentHandler) GetTree(c *gin.Context) {
	taskID, err := strconv.ParseUint(c.Param("taskId"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的任务 ID")
		return
	}

	db := database.GetDB()

	// 获取所有顶级评论
	var comments []models.Comment
	if err := db.Where("task_id = ? AND parent_id IS NULL", taskID).
		Preload("Member").
		Order("created_at DESC").
		Find(&comments).Error; err != nil {
		h.logger.Error("获取评论列表失败", zap.Error(err))
		response.Error(c, 500, "获取评论列表失败")
		return
	}

	// 为每个评论加载回复
	for i := range comments {
		h.loadReplies(db, &comments[i])
	}

	response.Success(c, comments)
}

// loadReplies 递归加载评论回复
func (h *CommentHandler) loadReplies(db *gorm.DB, comment *models.Comment) {
	var replies []models.Comment
	db.Where("parent_id = ?", comment.ID).
		Preload("Member").
		Order("created_at ASC").
		Find(&replies)

	comment.Replies = replies

	// 递归加载子回复
	for i := range replies {
		h.loadReplies(db, &replies[i])
	}
}

// GetStatistics 获取评论统计
func (h *CommentHandler) GetStatistics(c *gin.Context) {
	taskID, err := strconv.ParseUint(c.Param("taskId"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的任务 ID")
		return
	}

	db := database.GetDB()

	var stats models.CommentStatistics

	// 总评论数
	db.Model(&models.Comment{}).Where("task_id = ?", taskID).Count(&stats.Total)

	// 唯一评论者数量
	db.Model(&models.Comment{}).Where("task_id = ?", taskID).
		Distinct("member_id").Count(&stats.UniqueAuthors)

	response.Success(c, stats)
}

// Get 获取单个评论
func (h *CommentHandler) Get(c *gin.Context) {
	taskID, err := strconv.ParseUint(c.Param("taskId"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的任务 ID")
		return
	}

	commentID, err := strconv.ParseUint(c.Param("commentId"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的评论 ID")
		return
	}

	db := database.GetDB()

	var comment models.Comment
	if err := db.Where("id = ? AND task_id = ?", commentID, taskID).
		Preload("Member").
		First(&comment).Error; err != nil {
		response.NotFound(c, "评论不存在")
		return
	}

	response.Success(c, comment)
}

// Create 创建评论
func (h *CommentHandler) Create(c *gin.Context) {
	taskID, err := strconv.ParseUint(c.Param("taskId"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的任务 ID")
		return
	}

	var req struct {
		MemberID  uint64  `json:"memberId" binding:"required"`
		SubtaskID *uint64 `json:"subtaskId"`
		ParentID  *uint64 `json:"parentId"`
		Content   string  `json:"content" binding:"required"`
		Mentions  []uint64 `json:"mentions"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数格式错误")
		return
	}

	db := database.GetDB()

	// 验证任务是否存在
	var task models.Task
	if err := db.First(&task, taskID).Error; err != nil {
		response.NotFound(c, "任务不存在")
		return
	}

	// 如果有父评论，验证父评论是否存在
	if req.ParentID != nil {
		var parentComment models.Comment
		if err := db.Where("id = ? AND task_id = ?", *req.ParentID, taskID).First(&parentComment).Error; err != nil {
			response.BadRequest(c, "父评论不存在")
			return
		}
	}

	// 序列化 mentions
	var mentionsJSON string
	if len(req.Mentions) > 0 {
		mentionsBytes, _ := json.Marshal(req.Mentions)
		mentionsJSON = string(mentionsBytes)
	}

	comment := models.Comment{
		TaskID:    taskID,
		SubtaskID: req.SubtaskID,
		MemberID:  req.MemberID,
		ParentID:  req.ParentID,
		Content:   req.Content,
		Mentions:  mentionsJSON,
	}

	if err := db.Create(&comment).Error; err != nil {
		h.logger.Error("创建评论失败", zap.Error(err))
		response.Error(c, 500, "创建评论失败")
		return
	}

	// 加载成员信息
	db.Preload("Member").First(&comment, comment.ID)

	response.Success(c, comment)
}

// Update 更新评论
func (h *CommentHandler) Update(c *gin.Context) {
	taskID, err := strconv.ParseUint(c.Param("taskId"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的任务 ID")
		return
	}

	commentID, err := strconv.ParseUint(c.Param("commentId"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的评论 ID")
		return
	}

	var req struct {
		Content string   `json:"content"`
		Mentions []uint64 `json:"mentions"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数格式错误")
		return
	}

	db := database.GetDB()

	// 查找评论
	var comment models.Comment
	if err := db.Where("id = ? AND task_id = ?", commentID, taskID).First(&comment).Error; err != nil {
		response.NotFound(c, "评论不存在")
		return
	}

	// 更新内容
	updates := map[string]interface{}{}
	if req.Content != "" {
		updates["content"] = req.Content
	}
	if req.Mentions != nil {
		mentionsBytes, _ := json.Marshal(req.Mentions)
		updates["mentions"] = string(mentionsBytes)
	}

	if err := db.Model(&comment).Updates(updates).Error; err != nil {
		h.logger.Error("更新评论失败", zap.Error(err))
		response.Error(c, 500, "更新评论失败")
		return
	}

	// 重新加载
	db.Preload("Member").First(&comment, comment.ID)

	response.Success(c, comment)
}

// Delete 删除评论
func (h *CommentHandler) Delete(c *gin.Context) {
	taskID, err := strconv.ParseUint(c.Param("taskId"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的任务 ID")
		return
	}

	commentID, err := strconv.ParseUint(c.Param("commentId"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的评论 ID")
		return
	}

	var req struct {
		MemberID uint64 `json:"memberId"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数格式错误")
		return
	}

	db := database.GetDB()

	// 查找评论
	var comment models.Comment
	if err := db.Where("id = ? AND task_id = ?", commentID, taskID).First(&comment).Error; err != nil {
		response.NotFound(c, "评论不存在")
		return
	}

	// 验证是否是评论作者（可选的安全检查）
	if req.MemberID != 0 && comment.MemberID != req.MemberID {
		response.BadRequest(c, "只能删除自己的评论")
		return
	}

	// 删除评论（级联删除回复）
	if err := db.Delete(&comment).Error; err != nil {
		h.logger.Error("删除评论失败", zap.Error(err))
		response.Error(c, 500, "删除评论失败")
		return
	}

	response.Success(c, gin.H{"success": true, "message": "删除成功"})
}

// GetReplies 获取评论回复列表
func (h *CommentHandler) GetReplies(c *gin.Context) {
	taskID, err := strconv.ParseUint(c.Param("taskId"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的任务 ID")
		return
	}

	commentID, err := strconv.ParseUint(c.Param("commentId"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的评论 ID")
		return
	}

	db := database.GetDB()

	// 验证父评论是否存在
	var parentComment models.Comment
	if err := db.Where("id = ? AND task_id = ?", commentID, taskID).First(&parentComment).Error; err != nil {
		response.NotFound(c, "评论不存在")
		return
	}

	// 获取回复列表
	var replies []models.Comment
	if err := db.Where("parent_id = ?", commentID).
		Preload("Member").
		Order("created_at ASC").
		Find(&replies).Error; err != nil {
		h.logger.Error("获取回复列表失败", zap.Error(err))
		response.Error(c, 500, "获取回复列表失败")
		return
	}

	response.Success(c, replies)
}
