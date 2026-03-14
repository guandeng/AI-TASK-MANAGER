package config

import (
	"fmt"
	"os"

	"github.com/spf13/viper"
)

// Config 全局配置结构
type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Database DatabaseConfig `mapstructure:"database"`
	AI       AIConfig       `mapstructure:"ai"`
	General  GeneralConfig  `mapstructure:"general"`
}

// ServerConfig 服务器配置
type ServerConfig struct {
	Port int    `mapstructure:"port"`
	Mode string `mapstructure:"mode"` // debug/release
}

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	Host       string `mapstructure:"host"`
	Port       int    `mapstructure:"port"`
	Database   string `mapstructure:"database"`
	Username   string `mapstructure:"username"`
	Password   string `mapstructure:"password"`
	TablePrefix string `mapstructure:"table_prefix"`
	Charset    string `mapstructure:"charset"`
	Collation  string `mapstructure:"collation"`
	PoolSize   int    `mapstructure:"pool_size"`
}

// AIConfig AI 配置
type AIConfig struct {
	Provider   string                 `mapstructure:"provider"`
	Providers  map[string]AIProvider  `mapstructure:"providers"`
	Parameters AIParameters           `mapstructure:"parameters"`
}

// AIProvider AI 提供商配置
type AIProvider struct {
	Enabled bool   `mapstructure:"enabled"`
	APIKey  string `mapstructure:"api_key"`
	Model   string `mapstructure:"model"`
	BaseURL string `mapstructure:"base_url"`
}

// AIParameters AI 参数配置
type AIParameters struct {
	MaxTokens   int     `mapstructure:"max_tokens"`
	Temperature float64 `mapstructure:"temperature"`
}

// GeneralConfig 通用配置
type GeneralConfig struct {
	Debug           bool   `mapstructure:"debug"`
	LogLevel        string `mapstructure:"log_level"`
	DefaultSubtasks int    `mapstructure:"default_subtasks"`
	DefaultPriority string `mapstructure:"default_priority"`
	ProjectName     string `mapstructure:"project_name"`
}

var GlobalConfig *Config

// Load 加载配置文件
func Load(configPath string) (*Config, error) {
	v := viper.New()

	// 设置默认值
	setDefaults(v)

	// 设置配置文件
	if configPath != "" {
		v.SetConfigFile(configPath)
	} else {
		// 尝试从多个位置加载配置
		v.SetConfigName("config")
		v.SetConfigType("yaml")
		v.AddConfigPath(".")
		v.AddConfigPath("./backend")
		v.AddConfigPath("./config")
	}

	// 支持环境变量覆盖
	v.AutomaticEnv()
	v.SetEnvPrefix("ATM") // AI Task Manager 前缀

	// 读取配置文件
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("读取配置文件失败: %w", err)
		}
		// 配置文件不存在时使用默认值
		fmt.Println("未找到配置文件，使用默认值")
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("解析配置文件失败: %w", err)
	}

	// 从环境变量读取敏感信息
	if apiKey := os.Getenv("AI_API_KEY"); apiKey != "" {
		if provider, ok := cfg.AI.Providers[cfg.AI.Provider]; ok {
			provider.APIKey = apiKey
			cfg.AI.Providers[cfg.AI.Provider] = provider
		}
	}

	// 从环境变量读取数据库配置
	if dbPassword := os.Getenv("DB_PASSWORD"); dbPassword != "" {
		cfg.Database.Password = dbPassword
	}

	GlobalConfig = &cfg
	return &cfg, nil
}

// setDefaults 设置默认值
func setDefaults(v *viper.Viper) {
	// 服务器默认值
	v.SetDefault("server.port", 8080)
	v.SetDefault("server.mode", "debug")

	// 数据库默认值
	v.SetDefault("database.host", "localhost")
	v.SetDefault("database.port", 3306)
	v.SetDefault("database.database", "ai_task")
	v.SetDefault("database.username", "root")
	v.SetDefault("database.password", "123456")
	v.SetDefault("database.table_prefix", "task_")
	v.SetDefault("database.charset", "utf8mb4")
	v.SetDefault("database.collation", "utf8mb4_unicode_ci")
	v.SetDefault("database.pool_size", 10)

	// AI 默认值
	v.SetDefault("ai.provider", "qwen")
	v.SetDefault("ai.parameters.max_tokens", 8192)
	v.SetDefault("ai.parameters.temperature", 0.7)

	// 通用默认值
	v.SetDefault("general.debug", false)
	v.SetDefault("general.log_level", "info")
	v.SetDefault("general.default_subtasks", 3)
	v.SetDefault("general.default_priority", "medium")
	v.SetDefault("general.project_name", "AI Task Manager")
}

// GetDSN 获取数据库连接字符串
func (c *DatabaseConfig) GetDSN() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&collation=%s&parseTime=True&loc=Asia%%2FShanghai",
		c.Username,
		c.Password,
		c.Host,
		c.Port,
		c.Database,
		c.Charset,
		c.Collation,
	)
}
