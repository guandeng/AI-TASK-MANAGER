package ai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/ai-task-manager/backend/internal/config"
	"github.com/ai-task-manager/backend/internal/models"
)

// Provider AI 提供商接口
type Provider interface {
	ExpandTask(task *models.Task) ([]models.Subtask, error)
	Chat(prompt string) (string, error)
}

// Service AI 服务
type Service struct {
	cfg      *config.AIConfig
	provider Provider
}

// NewService 创建 AI 服务
func NewService(cfg *config.AIConfig) *Service {
	return &Service{
		cfg: cfg,
	}
}

// ExpandTask 展开任务（生成子任务）
func (s *Service) ExpandTask(task *models.Task) ([]models.Subtask, error) {
	prompt := s.buildExpandPrompt(task)
	response, err := s.Chat(prompt)
	if err != nil {
		return nil, err
	}
	return s.parseSubtasks(response, task.ID)
}

// Chat 发送聊天请求
func (s *Service) Chat(prompt string) (string, error) {
	provider := s.cfg.Provider
	switch provider {
	case "qwen", "perplexity":
		return s.chatOpenAICompat(provider, prompt)
	case "gemini":
		return s.chatGemini(prompt)
	default:
		return "", fmt.Errorf("unsupported AI provider: %s", provider)
	}
}

// chatOpenAICompat 使用 OpenAI 兼容 API 发送请求
func (s *Service) chatOpenAICompat(provider string, prompt string) (string, error) {
	cfg, ok := s.cfg.Providers[provider]
	if !ok || !cfg.Enabled {
		return "", fmt.Errorf("provider %s is not enabled", provider)
	}

	reqBody := map[string]interface{}{
		"model": cfg.Model,
		"messages": []map[string]string{
			{"role": "user", "content": prompt},
		},
		"max_tokens":   s.cfg.Parameters.MaxTokens,
		"temperature":  s.cfg.Parameters.Temperature,
	}

	jsonBody, _ := json.Marshal(reqBody)
	req, err := http.NewRequest("POST", cfg.BaseURL+"/chat/completions", bytes.NewBuffer(jsonBody))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+cfg.APIKey)

	client := &http.Client{Timeout: 120 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	var result struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
		Error struct {
			Message string `json:"message"`
		} `json:"error"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return "", err
	}

	if result.Error.Message != "" {
		return "", fmt.Errorf("AI API error: %s", result.Error.Message)
	}

	if len(result.Choices) == 0 {
		return "", fmt.Errorf("no response from AI")
	}

	return result.Choices[0].Message.Content, nil
}

// chatGemini 使用 Gemini API 发送请求
func (s *Service) chatGemini(prompt string) (string, error) {
	cfg, ok := s.cfg.Providers["gemini"]
	if !ok || !cfg.Enabled {
		return "", fmt.Errorf("gemini provider is not enabled")
	}

	baseURL := cfg.BaseURL
	if baseURL == "" {
		baseURL = "https://generativelanguage.googleapis.com/v1beta"
	}

	reqBody := map[string]interface{}{
		"contents": []map[string]interface{}{
			{
				"parts": []map[string]string{
					{"text": prompt},
				},
			},
		},
		"generationConfig": map[string]interface{}{
			"maxOutputTokens": s.cfg.Parameters.MaxTokens,
			"temperature":     s.cfg.Parameters.Temperature,
		},
	}

	jsonBody, _ := json.Marshal(reqBody)
	url := fmt.Sprintf("%s/models/%s:generateContent?key=%s", baseURL, cfg.Model, cfg.APIKey)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 120 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	var result struct {
		Candidates []struct {
			Content struct {
				Parts []struct {
					Text string `json:"text"`
				} `json:"parts"`
			} `json:"content"`
		} `json:"candidates"`
		Error struct {
			Message string `json:"message"`
		} `json:"error"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return "", err
	}

	if result.Error.Message != "" {
		return "", fmt.Errorf("Gemini API error: %s", result.Error.Message)
	}

	if len(result.Candidates) == 0 || len(result.Candidates[0].Content.Parts) == 0 {
		return "", fmt.Errorf("no response from Gemini")
	}

	return result.Candidates[0].Content.Parts[0].Text, nil
}

// buildExpandPrompt 构建任务展开提示词
func (s *Service) buildExpandPrompt(task *models.Task) string {
	return fmt.Sprintf(`请将以下任务拆分为 %d 个子任务。每个子任务应该是一个具体的、可执行的步骤。

任务标题：%s
任务描述：%s
任务详情：%s

请以 JSON 数组格式返回子任务列表，每个子任务包含以下字段：
- title: 子任务标题（必填）
- description: 子任务描述（可选）
- details: 详细说明（可选）

示例格式：
[
  {"title": "子任务1标题", "description": "子任务1描述", "details": "详细说明"},
  {"title": "子任务2标题", "description": "子任务2描述", "details": "详细说明"}
]

请只返回 JSON 数组，不要包含其他内容。`,
		3, // 默认子任务数量
		task.Title,
		task.Description,
		task.Details,
	)
}

// parseSubtasks 解析子任务
func (s *Service) parseSubtasks(response string, taskID uint64) ([]models.Subtask, error) {
	var subtasks []struct {
		Title       string `json:"title"`
		Description string `json:"description"`
		Details     string `json:"details"`
	}

	if err := json.Unmarshal([]byte(response), &subtasks); err != nil {
		return nil, fmt.Errorf("failed to parse subtasks: %w", err)
	}

	result := make([]models.Subtask, len(subtasks))
	for i, st := range subtasks {
		result[i] = models.Subtask{
			TaskID:      taskID,
			Title:       st.Title,
			Description: st.Description,
			Details:     st.Details,
			Status:      "pending",
			SortOrder:   uint(i),
		}
	}

	return result, nil
}

// GenerateSubtask 生成单个子任务
func (s *Service) GenerateSubtask(task *models.Task) (*models.Subtask, error) {
	subtasks, err := s.ExpandTask(task)
	if err != nil {
		return nil, err
	}
	if len(subtasks) == 0 {
		return nil, fmt.Errorf("no subtask generated")
	}
	return &subtasks[0], nil
}
