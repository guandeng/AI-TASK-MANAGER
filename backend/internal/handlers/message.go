package handlers

import (
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
	response.Success(c, []interface{}{})
}

// UnreadCount 获取未读消息数
func (h *MessageHandler) UnreadCount(c *gin.Context) {
	response.Success(c, gin.H{"count": 0})
}

// MarkRead 标记消息已读
func (h *MessageHandler) MarkRead(c *gin.Context) {
	response.Success(c, nil)
}

// MarkAllRead 标记全部已读
func (h *MessageHandler) MarkAllRead(c *gin.Context) {
	response.Success(c, nil)
}

// Delete 删除消息
func (h *MessageHandler) Delete(c *gin.Context) {
	response.Success(c, nil)
}
