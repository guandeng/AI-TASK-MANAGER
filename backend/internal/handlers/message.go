package handlers

import (
	"strconv"

	"github.com/ai-task-manager/backend/internal/database"
	"github.com/ai-task-manager/backend/internal/models"
	"github.com/ai-task-manager/backend/pkg/response"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// MessageHandler 消息处理器
type MessageHandler struct {
	logger *zap.Logger
}

// NewMessageHandler 创建消息处理器
func NewMessageHandler(logger *zap.Logger) *MessageHandler {
	return &MessageHandler{logger: logger}
}

// List 获取消息列表
func (h *MessageHandler) List(c *gin.Context) {
	db := database.GetDB()
	if db == nil {
		response.ServerError(c, "数据库未初始化")
		return
	}

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
	query := db.Model(&models.Message{})

	// 筛选条件：按任务ID筛选
	if taskID := c.Query("taskId"); taskID != "" {
		query = query.Where("task_id = ?", taskID)
	}

	// 筛选条件：按状态筛选
	if status := c.Query("status"); status != "" {
		query = query.Where("status = ?", status)
	}

	// 筛选条件：只看未读
	if unreadOnly := c.Query("unread"); unreadOnly == "true" {
		query = query.Where("is_read = ?", false)
	}

	// 计算总数
	var total int64
	if err := query.Count(&total).Error; err != nil {
		h.logger.Error("获取消息总数失败", zap.Error(err))
		response.Error(c, 500, "获取消息列表失败")
		return
	}

	// 分页查询，按创建时间倒序
	var messages []models.Message
	offset := (page - 1) * pageSize
	if err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&messages).Error; err != nil {
		h.logger.Error("获取消息列表失败", zap.Error(err))
		response.Error(c, 500, "获取消息列表失败")
		return
	}

	response.SuccessPage(c, messages, total, page, pageSize)
}

// UnreadCount 获取未读消息数
func (h *MessageHandler) UnreadCount(c *gin.Context) {
	db := database.GetDB()
	if db == nil {
		response.ServerError(c, "数据库未初始化")
		return
	}

	var count int64
	if err := db.Model(&models.Message{}).Where("is_read = ?", false).Count(&count).Error; err != nil {
		h.logger.Error("获取未读消息数失败", zap.Error(err))
		response.Error(c, 500, "获取未读消息数失败")
		return
	}

	response.Success(c, gin.H{"count": count})
}

// MarkRead 标记消息已读
func (h *MessageHandler) MarkRead(c *gin.Context) {
	db := database.GetDB()
	if db == nil {
		response.ServerError(c, "数据库未初始化")
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的消息 ID")
		return
	}

	if err := db.Model(&models.Message{}).Where("id = ?", id).Update("is_read", true).Error; err != nil {
		h.logger.Error("标记消息已读失败", zap.Error(err))
		response.Error(c, 500, "标记消息已读失败")
		return
	}

	response.Success(c, nil)
}

// MarkAllRead 标记全部已读
func (h *MessageHandler) MarkAllRead(c *gin.Context) {
	db := database.GetDB()
	if db == nil {
		response.ServerError(c, "数据库未初始化")
		return
	}

	if err := db.Model(&models.Message{}).Where("is_read = ?", false).Update("is_read", true).Error; err != nil {
		h.logger.Error("标记全部已读失败", zap.Error(err))
		response.Error(c, 500, "��记全部已读失败")
		return
	}

	response.Success(c, nil)
}

// Delete 删除消息
func (h *MessageHandler) Delete(c *gin.Context) {
	db := database.GetDB()
	if db == nil {
		response.ServerError(c, "数据库未初始化")
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的消息 ID")
		return
	}

	if err := db.Delete(&models.Message{}, id).Error; err != nil {
		h.logger.Error("删除消息失败", zap.Error(err))
		response.Error(c, 500, "删除消息失败")
		return
	}

	response.Success(c, nil)
}
