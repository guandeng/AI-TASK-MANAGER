package handlers

import (
	"github.com/ai-task-manager/backend/internal/config"
	"github.com/ai-task-manager/backend/pkg/response"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// KnowledgeHandler 知识库处理器
type KnowledgeHandler struct {
	logger   *zap.Logger
	service  *KnowledgeServiceWrapper
}

// KnowledgeServiceWrapper 知识库服务包装
type KnowledgeServiceWrapper struct {
	cfg    *config.KnowledgeConfig
	logger *zap.Logger
}

// NewKnowledgeHandler 创建知识库处理器
func NewKnowledgeHandler(logger *zap.Logger, cfg *config.Config) *KnowledgeHandler {
	return &KnowledgeHandler{
		logger: logger,
		service: &KnowledgeServiceWrapper{
			cfg:    &cfg.Knowledge,
			logger: logger,
		},
	}
}

// GetSummary 获取知识库摘要
// GET /api/knowledge/summary
func (h *KnowledgeHandler) GetSummary(c *gin.Context) {
	paths := c.QueryArray("paths")

	summary := map[string]interface{}{
		"enabled":    h.service.cfg.Enabled,
		"paths":      h.service.cfg.Paths,
		"maxSize":    h.service.cfg.MaxSize,
		"maxFiles":   h.service.cfg.MaxFiles,
		"fileTypes":  h.service.cfg.FileTypes,
		"customPaths": paths,
	}

	response.Success(c, summary)
}

// LoadRequest 加载知识库请求
type LoadRequest struct {
	Paths            []string `json:"paths"`
	AdditionalContext string   `json:"additionalContext"`
}

// Load 加载知识库内容
// POST /api/knowledge/load
func (h *KnowledgeHandler) Load(c *gin.Context) {
	var req LoadRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数格式错误")
		return
	}

	// 返回加载状态（实际内容由服务层处理）
	response.Success(c, gin.H{
		"status":   "loaded",
		"paths":    req.Paths,
		"hasContext": req.AdditionalContext != "",
	})
}
