package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/ai-task-manager/backend/internal/config"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func setupConfigTest(t *testing.T) (*ConfigHandler, *gin.Engine) {
	logger := zap.NewNop()

	cfg := &config.Config{
		AI: config.AIConfig{
			Provider: "qwen",
			Providers: map[string]config.AIProvider{
				"qwen": {
					Enabled: true,
					Model:   "qwen-turbo",
					APIKey:  "test-api-key",
				},
				"gemini": {
					Enabled: false,
					Model:   "gemini-pro",
				},
				"perplexity": {
					Enabled: false,
					Model:   "llama-3",
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

	handler := NewConfigHandler(logger, cfg)

	router := gin.New()
	return handler, router
}

func TestConfigHandler_Get(t *testing.T) {
	handler, router := setupConfigTest(t)
	router.GET("/config", handler.Get)

	req := httptest.NewRequest(http.MethodGet, "/config", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("期望状态码 %d, 实际 %d", http.StatusOK, w.Code)
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("解析响应失败: %v", err)
	}

	if resp["code"].(float64) != 0 {
		t.Errorf("期望 code 为 0, 实际 %v", resp["code"])
	}

	data := resp["data"].(map[string]interface{})

	// 验证 AI 配置存在
	ai, ok := data["ai"].(map[string]interface{})
	if !ok {
		t.Fatal("期望 data 包含 ai 配置")
	}

	if ai["provider"] != "qwen" {
		t.Errorf("期望 provider 为 qwen, 实际 %v", ai["provider"])
	}

	// 验证提供商配置
	providers, ok := ai["providers"].(map[string]interface{})
	if !ok {
		t.Fatal("期望 ai 包含 providers 配置")
	}

	// 验证 qwen 提供商
	qwen, ok := providers["qwen"].(map[string]interface{})
	if !ok {
		t.Fatal("期望 providers 包含 qwen 配置")
	}
	if qwen["enabled"] != true {
		t.Error("期望 qwen enabled 为 true")
	}

	// 验证参数配置存在
	if _, ok := ai["parameters"]; !ok {
		t.Fatal("期望 ai 包含 parameters 配置")
	}

	// 验证通用配置存在
	if _, ok := data["general"]; !ok {
		t.Fatal("期望 data 包含 general 配置")
	}
}

func TestConfigHandler_Get_HidesSensitiveData(t *testing.T) {
	handler, router := setupConfigTest(t)
	router.GET("/config", handler.Get)

	req := httptest.NewRequest(http.MethodGet, "/config", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("解析响应失败: %v", err)
	}

	data := resp["data"].(map[string]interface{})
	ai := data["ai"].(map[string]interface{})
	providers := ai["providers"].(map[string]interface{})
	qwen := providers["qwen"].(map[string]interface{})

	// 验证敏感信息（API Key）不被返回
	if _, ok := qwen["apiKey"]; ok {
		t.Error("期望不返回 apiKey，但实际返回了")
	}
}

func TestConfigHandler_Update(t *testing.T) {
	handler, router := setupConfigTest(t)
	router.PUT("/config", handler.Update)

	// 需要发送有效的 JSON body
	body := `{}`
	req := httptest.NewRequest(http.MethodPut, "/config", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("期望状态码 %d, 实际 %d", http.StatusOK, w.Code)
	}
}

func TestConfigHandler_UpdateAIProvider(t *testing.T) {
	handler, router := setupConfigTest(t)
	router.PUT("/config/ai/provider", handler.UpdateAIProvider)

	// 需要发送有效的 JSON body
	body := `{"provider": "qwen"}`
	req := httptest.NewRequest(http.MethodPut, "/config/ai/provider", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("期望状态码 %d, 实际 %d", http.StatusOK, w.Code)
	}
}

func TestConfigHandler_UpdateSpecificProvider(t *testing.T) {
	handler, router := setupConfigTest(t)
	router.PUT("/config/ai/providers/:provider", handler.UpdateSpecificProvider)

	// 需要发送有效的 JSON body
	body := `{"enabled": true}`
	req := httptest.NewRequest(http.MethodPut, "/config/ai/providers/qwen", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("期望状态码 %d, 实际 %d", http.StatusOK, w.Code)
	}
}

func TestConfigHandler_Reset(t *testing.T) {
	handler, router := setupConfigTest(t)
	router.POST("/config/reset", handler.Reset)

	req := httptest.NewRequest(http.MethodPost, "/config/reset", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("期望状态码 %d, 实际 %d", http.StatusOK, w.Code)
	}
}

func TestConfigHandler_UpdateAIProvider_InvalidProvider(t *testing.T) {
	handler, router := setupConfigTest(t)
	router.PUT("/config/ai/provider", handler.UpdateAIProvider)

	tests := []struct {
		name       string
		body       string
		expectCode int
	}{
		{
			name:       "空提供商",
			body:       `{"provider": ""}`,
			expectCode: http.StatusBadRequest,
		},
		{
			name:       "不支持的提供商",
			body:       `{"provider": "unsupported"}`,
			expectCode: http.StatusBadRequest,
		},
		{
			name:       "无效JSON",
			body:       `invalid json`,
			expectCode: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPut, "/config/ai/provider", strings.NewReader(tt.body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			if w.Code != tt.expectCode {
				t.Errorf("期望状态码 %d, 实际 %d", tt.expectCode, w.Code)
			}
		})
	}
}

func TestConfigHandler_UpdateSpecificProvider_InvalidJSON(t *testing.T) {
	handler, router := setupConfigTest(t)
	router.PUT("/config/ai/providers/:provider", handler.UpdateSpecificProvider)

	// 无效 JSON
	req := httptest.NewRequest(http.MethodPut, "/config/ai/providers/qwen", strings.NewReader(`invalid json`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("期望状态码 %d, 实际 %d", http.StatusBadRequest, w.Code)
	}
}

func TestConfigHandler_SetConfigPath(t *testing.T) {
	handler, _ := setupConfigTest(t)

	// 测试设置配置路径
	handler.SetConfigPath("/custom/config.yaml")

	if handler.configPath != "/custom/config.yaml" {
		t.Errorf("期望 configPath 为 /custom/config.yaml, 实际 %s", handler.configPath)
	}
}

func TestConfigHandler_Update_WithAIParameters(t *testing.T) {
	handler, router := setupConfigTest(t)
	router.PUT("/config", handler.Update)

	body := `{
		"ai": {
			"provider": "gemini",
			"parameters": {
				"maxTokens": 4096,
				"temperature": 0.5
			},
			"providers": {
				"gemini": {
					"enabled": true,
					"model": "gemini-pro-vision"
				}
			}
		},
		"general": {
			"debug": true,
			"logLevel": "debug",
			"defaultSubtasks": 5,
			"defaultPriority": "high",
			"projectName": "Test Project"
		}
	}`

	req := httptest.NewRequest(http.MethodPut, "/config", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("期望状态码 %d, 实际 %d", http.StatusOK, w.Code)
	}

	// 验证配置已更新
	if handler.config.AI.Provider != "gemini" {
		t.Errorf("期望 provider 为 gemini, 实际 %s", handler.config.AI.Provider)
	}

	if handler.config.AI.Parameters.MaxTokens != 4096 {
		t.Errorf("期望 MaxTokens 为 4096, 实际 %d", handler.config.AI.Parameters.MaxTokens)
	}

	if handler.config.General.Debug != true {
		t.Error("期望 Debug 为 true")
	}
}

func TestConfigHandler_Update_InvalidJSON(t *testing.T) {
	handler, router := setupConfigTest(t)
	router.PUT("/config", handler.Update)

	req := httptest.NewRequest(http.MethodPut, "/config", strings.NewReader(`invalid json`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("期望状态码 %d, 实际 %d", http.StatusBadRequest, w.Code)
	}
}

func TestConfigHandler_UpdateSpecificProvider_WithAPIKey(t *testing.T) {
	handler, router := setupConfigTest(t)
	router.PUT("/config/ai/providers/:provider", handler.UpdateSpecificProvider)

	body := `{
		"enabled": true,
		"apiKey": "new-api-key",
		"model": "qwen-max",
		"baseUrl": "https://new.api.url"
	}`

	req := httptest.NewRequest(http.MethodPut, "/config/ai/providers/qwen", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("期望状态码 %d, 实际 %d", http.StatusOK, w.Code)
	}

	// 验证提供商配置已更新
	if handler.config.AI.Providers["qwen"].APIKey != "new-api-key" {
		t.Error("期望 API Key 已更新")
	}
}
