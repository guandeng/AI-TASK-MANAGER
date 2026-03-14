package handlers

import (
	"github.com/ai-task-manager/backend/pkg/response"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// ActivityHandler 活动日志处理器
type ActivityHandler struct {
	logger *zap.Logger
}

// NewActivityHandler 创建活动日志处理器
func NewActivityHandler(logger *zap.Logger) *ActivityHandler {
	return &ActivityHandler{logger: logger}
}

// List 获取活动日志列表
func (h *ActivityHandler) List(c *gin.Context) {
	response.Success(c, []interface{}{})
}

// ListByTask 获取任务的活动日志
func (h *ActivityHandler) ListByTask(c *gin.Context) {
	response.Success(c, []interface{}{})
}

// Statistics 获取活动统计
func (h *ActivityHandler) Statistics(c *gin.Context) {
	response.Success(c, gin.H{
		"total":    0,
		"today":    0,
		"thisWeek": 0,
	})
}
