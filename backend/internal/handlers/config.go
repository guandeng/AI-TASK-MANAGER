package handlers

import (
	"github.com/ai-task-manager/backend/internal/config"
	"github.com/ai-task-manager/backend/pkg/response"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// ConfigHandler 配置处理器
type ConfigHandler struct {
	logger *zap.Logger
	config *config.Config
}

// NewConfigHandler 创建配置处理��
func NewConfigHandler(logger *zap.Logger, cfg *config.Config) *ConfigHandler {
	return &ConfigHandler{
		logger: logger,
		config: cfg,
	}
}

// Get 获取配置
func (h *ConfigHandler) Get(c *gin.Context) {
	// 返回安全配置（不包含敏感信息）
	safeConfig := gin.H{
		"ai": gin.H{
			"provider": h.config.AI.Provider,
			"providers": gin.H{
				"qwen": gin.H{
					"enabled": h.config.AI.Providers["qwen"].Enabled,
					"model":   h.config.AI.Providers["qwen"].Model,
				},
				"gemini": gin.H{
					"enabled": h.config.AI.Providers["gemini"].Enabled,
					"model":   h.config.AI.Providers["gemini"].Model,
				},
				"perplexity": gin.H{
					"enabled": h.config.AI.Providers["perplexity"].Enabled,
					"model":   h.config.AI.Providers["perplexity"].Model,
				},
			},
			"parameters": h.config.AI.Parameters,
		},
		"general": h.config.General,
	}
	response.Success(c, safeConfig)
}

// Update 更新配置
func (h *ConfigHandler) Update(c *gin.Context) {
	// TODO: 实现
	response.Success(c, nil)
}

// UpdateAIProvider 更新 AI 提供商配置
func (h *ConfigHandler) UpdateAIProvider(c *gin.Context) {
	// TODO: 实现
	response.Success(c, nil)
}

// UpdateSpecificProvider 更新指定提供商配置
func (h *ConfigHandler) UpdateSpecificProvider(c *gin.Context) {
	// TODO: 实现
	response.Success(c, nil)
}

// Reset 重置配置
func (h *ConfigHandler) Reset(c *gin.Context) {
	// TODO: 实现
	response.Success(c, nil)
}
