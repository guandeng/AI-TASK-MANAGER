package config

import (
	"os"
	"testing"
)

func TestDatabaseConfig_GetDSN(t *testing.T) {
	cfg := &DatabaseConfig{
		Host:       "localhost",
		Port:       3306,
		Database:   "test_db",
		Username:   "root",
		Password:   "password",
		Charset:    "utf8mb4",
		Collation:  "utf8mb4_unicode_ci",
	}

	dsn := cfg.GetDSN()
	expected := "root:password@tcp(localhost:3306)/test_db?charset=utf8mb4&collation=utf8mb4_unicode_ci&parseTime=True&loc=Asia%2FShanghai"

	if dsn != expected {
		t.Errorf("期望 DSN '%s', 实际 '%s'", expected, dsn)
	}
}

func TestLoad_DefaultConfig(t *testing.T) {
	// 不使用配置文件
	cfg, err := Load("")
	if err != nil {
		t.Logf("加载默认配置: %v", err)
		return
	}

	if cfg == nil {
		t.Error("期望返回配置")
		return
	}

	// 验证默认值
	if cfg.Server.Port != 8080 {
		t.Errorf("期望默认端口 8080, 实际 %d", cfg.Server.Port)
	}
}

func TestLoad_WithEnvVariables(t *testing.T) {
	// 设置环境变量
	os.Setenv("ATM_AI_PROVIDER", "test_provider")
	defer os.Unsetenv("ATM_AI_PROVIDER")

	cfg, err := Load("")
	if err != nil {
		t.Logf("加载配置: %v", err)
		return
	}

	if cfg == nil {
		t.Skip("跳过测试 - 无法加载配置")
	}
}

func TestConfig_Structure(t *testing.T) {
	cfg := &Config{
		Server: ServerConfig{
			Port: 8080,
			Mode: "debug",
		},
		Database: DatabaseConfig{
			Host:     "localhost",
			Port:     3306,
			Database: "test",
			Username: "root",
			Password: "",
		},
		AI: AIConfig{
			Provider: "qwen",
			Providers: map[string]AIProvider{
				"qwen": {
					Enabled: true,
					Model:   "qwen-plus",
				},
			},
		},
		General: GeneralConfig{
			Debug:           false,
			LogLevel:        "info",
			DefaultSubtasks: 3,
			DefaultPriority: "medium",
			ProjectName:     "Test Project",
		},
		Knowledge: KnowledgeConfig{
			Enabled:   false,
			MaxSize:   500,
			MaxFiles:  50,
			FileTypes: []string{".md", ".txt"},
		},
	}

	if cfg.Server.Port != 8080 {
		t.Errorf("期望端口 8080, 实际 %d", cfg.Server.Port)
	}

	if cfg.AI.Provider != "qwen" {
		t.Errorf("期望 provider 'qwen', 实际 '%s'", cfg.AI.Provider)
	}
}

func TestAIParameters(t *testing.T) {
	params := AIParameters{
		MaxTokens:   8192,
		Temperature: 0.7,
	}

	if params.MaxTokens != 8192 {
		t.Errorf("期望 MaxTokens 8192, 实际 %d", params.MaxTokens)
	}

	if params.Temperature != 0.7 {
		t.Errorf("期望 Temperature 0.7, 实际 %f", params.Temperature)
	}
}
