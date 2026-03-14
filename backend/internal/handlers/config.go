package handlers

import (
	"os"
	"path/filepath"

	"github.com/ai-task-manager/backend/internal/config"
	"github.com/ai-task-manager/backend/pkg/response"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
)

// ConfigHandler 配置处理器
type ConfigHandler struct {
	logger     *zap.Logger
	config     *config.Config
	configPath string
}

// NewConfigHandler 创建配置处理器
func NewConfigHandler(logger *zap.Logger, cfg *config.Config) *ConfigHandler {
	return &ConfigHandler{
		logger: logger,
		config: cfg,
	}
}

// SetConfigPath 设置配置文件路径
func (h *ConfigHandler) SetConfigPath(path string) {
	h.configPath = path
}

// Get 获取配置
func (h *ConfigHandler) Get(c *gin.Context) {
	// 返回安全配置（不包含敏感信息如 API Key）
	safeConfig := gin.H{
		"ai": gin.H{
			"provider": h.config.AI.Provider,
			"providers": gin.H{
				"qwen": gin.H{
					"enabled": h.config.AI.Providers["qwen"].Enabled,
					"model":   h.config.AI.Providers["qwen"].Model,
					"baseUrl": h.config.AI.Providers["qwen"].BaseURL,
				},
				"gemini": gin.H{
					"enabled": h.config.AI.Providers["gemini"].Enabled,
					"model":   h.config.AI.Providers["gemini"].Model,
					"baseUrl": h.config.AI.Providers["gemini"].BaseURL,
				},
				"perplexity": gin.H{
					"enabled": h.config.AI.Providers["perplexity"].Enabled,
					"model":   h.config.AI.Providers["perplexity"].Model,
					"baseUrl": h.config.AI.Providers["perplexity"].BaseURL,
				},
			},
			"parameters": h.config.AI.Parameters,
		},
		"general": h.config.General,
	}
	response.Success(c, safeConfig)
}

// UpdateRequest 更新配置请求
type UpdateRequest struct {
	AI      *AIConfigUpdate      `json:"ai"`
	General *GeneralConfigUpdate `json:"general"`
}

// AIConfigUpdate AI 配置更新
type AIConfigUpdate struct {
	Provider   string                        `json:"provider"`
	Providers  map[string]ProviderConfigUpdate `json:"providers"`
	Parameters *AIParametersUpdate           `json:"parameters"`
}

// ProviderConfigUpdate 提供商配置更新
type ProviderConfigUpdate struct {
	Enabled bool   `json:"enabled"`
	APIKey  string `json:"apiKey"`
	Model   string `json:"model"`
	BaseURL string `json:"baseUrl"`
}

// AIParametersUpdate AI 参数更新
type AIParametersUpdate struct {
	MaxTokens   int     `json:"maxTokens"`
	Temperature float64 `json:"temperature"`
}

// GeneralConfigUpdate 通用配置更新
type GeneralConfigUpdate struct {
	Debug           bool   `json:"debug"`
	LogLevel        string `json:"logLevel"`
	DefaultSubtasks int    `json:"defaultSubtasks"`
	DefaultPriority string `json:"defaultPriority"`
	ProjectName     string `json:"projectName"`
	UseChinese      bool   `json:"useChinese"`
}

// Update 更新配置
func (h *ConfigHandler) Update(c *gin.Context) {
	var req UpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "无效的请求参数")
		return
	}

	// 更新 AI 配置
	if req.AI != nil {
		if req.AI.Provider != "" {
			h.config.AI.Provider = req.AI.Provider
		}
		if req.AI.Parameters != nil {
			h.config.AI.Parameters.MaxTokens = req.AI.Parameters.MaxTokens
			h.config.AI.Parameters.Temperature = req.AI.Parameters.Temperature
		}
		for name, provider := range req.AI.Providers {
			if p, ok := h.config.AI.Providers[name]; ok {
				p.Enabled = provider.Enabled
				if provider.APIKey != "" {
					p.APIKey = provider.APIKey
				}
				if provider.Model != "" {
					p.Model = provider.Model
				}
				if provider.BaseURL != "" {
					p.BaseURL = provider.BaseURL
				}
				h.config.AI.Providers[name] = p
			} else {
				h.config.AI.Providers[name] = config.AIProvider{
					Enabled: provider.Enabled,
					APIKey:  provider.APIKey,
					Model:   provider.Model,
					BaseURL: provider.BaseURL,
				}
			}
		}
	}

	// 更新通用配置
	if req.General != nil {
		h.config.General.Debug = req.General.Debug
		if req.General.LogLevel != "" {
			h.config.General.LogLevel = req.General.LogLevel
		}
		if req.General.DefaultSubtasks > 0 {
			h.config.General.DefaultSubtasks = req.General.DefaultSubtasks
		}
		if req.General.DefaultPriority != "" {
			h.config.General.DefaultPriority = req.General.DefaultPriority
		}
		if req.General.ProjectName != "" {
			h.config.General.ProjectName = req.General.ProjectName
		}
	}

	// 保存配置到文件
	if err := h.saveConfig(); err != nil {
		h.logger.Error("保存配置失败", zap.Error(err))
		response.Error(c, 500, "保存配置失败")
		return
	}

	// 更新全局配置
	config.GlobalConfig = h.config

	response.Success(c, gin.H{"message": "配置更新成功"})
}

// UpdateAIProvider 更新 AI 提供商配置
func (h *ConfigHandler) UpdateAIProvider(c *gin.Context) {
	var req struct {
		Provider string `json:"provider"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "无效的请求参数")
		return
	}

	if req.Provider == "" {
		response.BadRequest(c, "提供商不能为空")
		return
	}

	// 检查提供商是否存在
	if _, ok := h.config.AI.Providers[req.Provider]; !ok {
		response.BadRequest(c, "不支持的提供商")
		return
	}

	h.config.AI.Provider = req.Provider

	// 保存配置到文件
	if err := h.saveConfig(); err != nil {
		h.logger.Error("保存配置失败", zap.Error(err))
		response.Error(c, 500, "保存配置失败")
		return
	}

	// 更新全局配置
	config.GlobalConfig = h.config

	response.Success(c, gin.H{"message": "AI 提供商切换成功"})
}

// UpdateSpecificProvider 更新指定提供商配置
func (h *ConfigHandler) UpdateSpecificProvider(c *gin.Context) {
	provider := c.Param("provider")
	if provider == "" {
		response.BadRequest(c, "提供商不能为空")
		return
	}

	var req ProviderConfigUpdate
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "无效的请求参数")
		return
	}

	// 更新或创建提供商配置
	if p, ok := h.config.AI.Providers[provider]; ok {
		p.Enabled = req.Enabled
		if req.APIKey != "" {
			p.APIKey = req.APIKey
		}
		if req.Model != "" {
			p.Model = req.Model
		}
		if req.BaseURL != "" {
			p.BaseURL = req.BaseURL
		}
		h.config.AI.Providers[provider] = p
	} else {
		h.config.AI.Providers[provider] = config.AIProvider{
			Enabled: req.Enabled,
			APIKey:  req.APIKey,
			Model:   req.Model,
			BaseURL: req.BaseURL,
		}
	}

	// 保存配置到文件
	if err := h.saveConfig(); err != nil {
		h.logger.Error("保存配置失败", zap.Error(err))
		response.Error(c, 500, "保存配置失败")
		return
	}

	// 更新全局配置
	config.GlobalConfig = h.config

	response.Success(c, gin.H{"message": "提供商配置更新成功"})
}

// Reset 重置配置
func (h *ConfigHandler) Reset(c *gin.Context) {
	// 重置为默认值
	h.config = &config.Config{
		Server: config.ServerConfig{
			Port: 8080,
			Mode: "debug",
		},
		Database: config.DatabaseConfig{
			Host:        "localhost",
			Port:        3306,
			Database:    "ai_task",
			Username:    "root",
			Password:    "123456",
			TablePrefix: "task_",
			Charset:     "utf8mb4",
			Collation:   "utf8mb4_unicode_ci",
			PoolSize:    10,
		},
		AI: config.AIConfig{
			Provider: "qwen",
			Providers: map[string]config.AIProvider{
				"qwen": {
					Enabled: true,
					Model:   "qwen-max",
				},
				"gemini": {
					Enabled: false,
					Model:   "gemini-pro",
				},
				"perplexity": {
					Enabled: false,
					Model:   "llama-3.1-sonar-large-128k-online",
				},
			},
			Parameters: config.AIParameters{
				MaxTokens:   8192,
				Temperature: 0.7,
			},
		},
		General: config.GeneralConfig{
			Debug:           false,
			LogLevel:        "info",
			DefaultSubtasks: 3,
			DefaultPriority: "medium",
			ProjectName:     "AI Task Manager",
		},
	}

	// 保存配置到文件
	if err := h.saveConfig(); err != nil {
		h.logger.Error("重置配置失败", zap.Error(err))
		response.Error(c, 500, "重置配置失败")
		return
	}

	// 更新全局配置
	config.GlobalConfig = h.config

	response.Success(c, gin.H{"message": "配置已重置为默认值"})
}

// saveConfig 保存配置到文件
func (h *ConfigHandler) saveConfig() error {
	// 如果没有配置文件路径，尝试查找
	if h.configPath == "" {
		h.configPath = "config.yaml"
		// 检查常见配置文件位置
		locations := []string{"config.yaml", "config.yml", "./config/config.yaml", "./backend/config.yaml"}
		for _, loc := range locations {
			if _, err := os.Stat(loc); err == nil {
				h.configPath = loc
				break
			}
		}
	}

	// 确保目录存在
	dir := filepath.Dir(h.configPath)
	if dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
	}

	// 将配置转换为 YAML 格式
	data, err := yaml.Marshal(h.config)
	if err != nil {
		return err
	}

	// 写入文件
	if err := os.WriteFile(h.configPath, data, 0644); err != nil {
		return err
	}

	// 重新加载 viper 配置
	viper.SetConfigFile(h.configPath)
	if err := viper.ReadInConfig(); err != nil {
		h.logger.Warn("重新加载 viper 配置失败", zap.Error(err))
	}

	return nil
}
