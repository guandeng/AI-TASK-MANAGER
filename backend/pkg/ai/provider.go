package ai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/ai-task-manager/backend/internal/config"
	"github.com/ai-task-manager/backend/internal/models"
)

// Provider AI 提供商接口
type Provider interface {
	ExpandTask(task *models.Task) ([]models.Subtask, error)
	Chat(prompt string) (string, error)
	Research(prompt string) (string, error) // Perplexity 研究功能
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

// SplitRequirement 将需求拆分为任务
// taskType: frontend, backend, fullstack
func (s *Service) SplitRequirement(requirement *models.Requirement, taskType string) ([]models.Task, error) {
	return s.SplitRequirementWithLanguage(requirement, taskType, nil)
}

// SplitRequirementWithLanguage 将需求拆分为任务（带语言信息）
// taskType: frontend, backend, fullstack
func (s *Service) SplitRequirementWithLanguage(requirement *models.Requirement, taskType string, language *models.Language) ([]models.Task, error) {
	prompt := s.buildSplitRequirementPromptWithLanguage(requirement, taskType, language)
	response, err := s.Chat(prompt)
	if err != nil {
		return nil, err
	}
	return s.parseTasks(response, requirement.ID)
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

// Research 使用 Perplexity 进行研究
func (s *Service) Research(prompt string) (string, error) {
	// 优先使用 Perplexity，如果没有配置则使用默认 provider
	if cfg, ok := s.cfg.Providers["perplexity"]; ok && cfg.Enabled {
		return s.chatOpenAICompat("perplexity", prompt)
	}
	// 回退到默认 provider
	return s.Chat(prompt)
}

// ExpandTaskWithKnowledge 使用知识库展开任务
func (s *Service) ExpandTaskWithKnowledge(task *models.Task, knowledgeContext string) ([]models.Subtask, error) {
	prompt := s.buildExpandPromptWithKnowledge(task, knowledgeContext)
	response, err := s.Chat(prompt)
	if err != nil {
		return nil, err
	}
	return s.parseSubtasks(response, task.ID)
}

// ExpandTaskWithResearch 使用研究功能展开任务
func (s *Service) ExpandTaskWithResearch(task *models.Task) ([]models.Subtask, error) {
	// 先进行研究
	researchPrompt := fmt.Sprintf(`请研究以下任务的实现方案，包括：
1. 最佳实践
2. 常见问题和解决方案
3. 推荐的技术栈和工具
4. 代码示例（如果有）

任务标题：%s
任务描述：%s
任务详情：%s

请提供详细的研究结果。`, task.Title, task.Description, task.Details)

	researchResult, err := s.Research(researchPrompt)
	if err != nil {
		return nil, fmt.Errorf("研究失败: %w", err)
	}

	// 基于研究结果展开任务
	prompt := s.buildExpandPromptWithResearch(task, researchResult)
	response, err := s.Chat(prompt)
	if err != nil {
		return nil, err
	}
	return s.parseSubtasks(response, task.ID)
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

技术栈说明：
- 后端：Go 语言，使用 Gin 框架和 GORM
- API 规范：使用 GET 和 POST 请求，删除操作使用 POST + /delete 路径

任务标题：%s
任务描述：%s
任务详情：%s

请以 JSON 数组格式返回子任务列表，每个子任务包含以下字段：
- title: 子任务标题（必填）
- description: 子任务描述（可选）
- details: 详细说明（可选）
- priority: 优先级（high/medium/low，默认medium）
- codeInterface: 代码接口定义，JSON对象包含：
  - name: 函数/方法名称
  - inputs: 输入参数类型定义（TypeScript格式）
  - outputs: 输出类型定义（TypeScript格式）
  - example: 使用示例代码
- acceptanceCriteria: 验收标准数组，每项包含：
  - description: 验收条件描述
  - completed: 是否完成（默认false）
- relatedFiles: 关联的源文件路径数组（如 ["src/auth/service.ts", "src/auth/types.ts"]）
- codeHints: 代码实现提示和建议

示例格式：
[
  {
    "title": "验证用户登录",
    "description": "实现用户登录验证逻辑",
    "details": "校验用户名密码，生成JWT token",
    "priority": "high",
    "codeInterface": {
      "name": "validateLogin",
      "inputs": "{ username: string; password: string }",
      "outputs": "{ success: boolean; token?: string; error?: string }",
      "example": "const result = await validateLogin({ username: 'admin', password: '123' });"
    },
    "acceptanceCriteria": [
      {"description": "用户名或密码为空返回 success: false", "completed": false},
      {"description": "密码错误返回 success: false 且 error 包含错误信息", "completed": false},
      {"description": "验证成功返回 token", "completed": false}
    ],
    "relatedFiles": ["src/auth/service.ts", "src/auth/types.ts"],
    "codeHints": "使用 bcrypt.compare 验证密码，使用 jsonwebtoken 生成 token"
  }
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
	// 清理响应，移除 Markdown 代码块标记
	cleanedResponse := response

	// 移除 ```json 和 ``` 标记
	re := regexp.MustCompile("```(?:json)?\\s*")
	cleanedResponse = re.ReplaceAllString(cleanedResponse, "")
	cleanedResponse = strings.TrimSpace(cleanedResponse)

	// 尝试从响应中提取 JSON 数组
	jsonStart := strings.Index(cleanedResponse, "[")
	jsonEnd := strings.LastIndex(cleanedResponse, "]")

	if jsonStart == -1 || jsonEnd == -1 || jsonEnd < jsonStart {
		return nil, fmt.Errorf("could not find valid JSON array in response")
	}

	jsonStr := cleanedResponse[jsonStart : jsonEnd+1]

	// 定义解析结构
	type CodeInterface struct {
		Name    string `json:"name"`
		Inputs  string `json:"inputs"`
		Outputs string `json:"outputs"`
		Example string `json:"example"`
	}

	type AcceptanceCriteria struct {
		ID          uint   `json:"id"`
		Description string `json:"description"`
		Completed   bool   `json:"completed"`
	}

	var subtasksData []struct {
		Title              string               `json:"title"`
		Description        string               `json:"description"`
		Details            string               `json:"details"`
		Priority           string               `json:"priority"`
		CodeInterface      *CodeInterface       `json:"codeInterface"`
		AcceptanceCriteria []AcceptanceCriteria `json:"acceptanceCriteria"`
		RelatedFiles       []string             `json:"relatedFiles"`
		CodeHints          string               `json:"codeHints"`
	}

	if err := json.Unmarshal([]byte(jsonStr), &subtasksData); err != nil {
		return nil, fmt.Errorf("failed to parse subtasks: %w", err)
	}

	result := make([]models.Subtask, len(subtasksData))
	for i, st := range subtasksData {
		// 处理优先级
		priority := st.Priority
		if priority == "" {
			priority = "medium"
		}

		// 处理 codeInterface -> JSON 字符串
		var codeInterfaceStr *string
		if st.CodeInterface != nil {
			jsonBytes, err := json.Marshal(st.CodeInterface)
			if err == nil {
				str := string(jsonBytes)
				codeInterfaceStr = &str
			}
		}

		// 处理 acceptanceCriteria -> JSON 字符串
		var acceptanceCriteriaStr *string
		if len(st.AcceptanceCriteria) > 0 {
			// 为每个验收标准分配 ID
			for j := range st.AcceptanceCriteria {
				if st.AcceptanceCriteria[j].ID == 0 {
					st.AcceptanceCriteria[j].ID = uint(j + 1)
				}
			}
			jsonBytes, err := json.Marshal(st.AcceptanceCriteria)
			if err == nil {
				str := string(jsonBytes)
				acceptanceCriteriaStr = &str
			}
		}

		// 处理 relatedFiles -> JSON 字符串
		var relatedFilesStr *string
		if len(st.RelatedFiles) > 0 {
			jsonBytes, err := json.Marshal(st.RelatedFiles)
			if err == nil {
				str := string(jsonBytes)
				relatedFilesStr = &str
			}
		}

		// 处理 codeHints
		var codeHintsStr *string
		if st.CodeHints != "" {
			codeHintsStr = &st.CodeHints
		}

		result[i] = models.Subtask{
			TaskID:             taskID,
			Title:              st.Title,
			Description:        st.Description,
			Details:            st.Details,
			Status:             "pending",
			Priority:           priority,
			SortOrder:          uint(i),
			CodeInterface:      codeInterfaceStr,
			AcceptanceCriteria: acceptanceCriteriaStr,
			RelatedFiles:       relatedFilesStr,
			CodeHints:          codeHintsStr,
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

// buildSplitRequirementPrompt 构建需求拆分提示词
func (s *Service) buildSplitRequirementPrompt(requirement *models.Requirement, taskType string) string {
	// 根据任务类型生成不同的提示
	var typeGuidance string
	switch taskType {
	case "frontend":
		typeGuidance = `IMPORTANT: 只需要生成前端相关的任务，包括：
- 前端页面和组件开发
- 前端路由和状态管理
- 前端 API 调用和数据展示
- 前端样式和交互效果
- 前端表单验证和处理

不要包含后端 API 开发、数据库设计、后端逻辑等后端任务。`
	case "backend":
		typeGuidance = `IMPORTANT: 只需要生成后端相关的任务，包括：
- 后端 API 接口开发
- 数据库设计和操作
- 后端业务逻辑实现
- 后端中间件和服务层
- 后端数据验证和处理

不要包含前端页面、组件、样式等前端任务。`
	case "fullstack":
		typeGuidance = `需要生成前端和后端的完整任务，包括：
- 前端页面和组件开发
- 后端 API 接口开发
- 数据库设计和操作
- 前后端联调
- 完整的功能实现`
	default:
		typeGuidance = `需要生成后端相关的任务。`
	}

	return fmt.Sprintf(`你是一个AI助手，帮助将产品需求文档（PRD）拆分为开发任务。

请分析以下需求并拆分为合适的开发任务。每个任务应该是一个具体的、可执行的开发单元。

%s

需求标题：%s
需求内容：
%s

请以 JSON 数组格式返回任务列表，每个任务包含以下字段：
- title: 任务标题（简洁明了）
- description: 任务描述（详细说明要做什么）
- details: 实现细节（技术方案、注意事项等）
- testStrategy: 测试策略（如何测试这个任务，包括单元测试、集成测试、边界条件等）
- priority: 优先级（high/medium/low）
- dependencies: 依赖的任务索引数组（如 [0, 1] 表示依赖第1和第2个任务）

拆分原则：
1. 按照功能模块拆分，每个任务专注于一个功能点
2. 优先级设置：基础架构和高优先级功能设为 high，核心功能设为 medium，辅助功能设为 low
3. 合理设置依赖关系，确保任务可以按顺序执行
4. 每个任务应该足够独立，可以单独开发和测试
5. 为每个任务提供明确的测试策略，说明如何验证功能正确性

示例格式：
[
  {
    "title": "任务1标题",
    "description": "任务1描述",
    "details": "实现细节",
    "testStrategy": "1. 单元测试：测试核心逻辑\n2. 边界测试：检查空值、边界值\n3. 集成测试：验证与其他模块的交互",
    "priority": "high",
    "dependencies": []
  },
  {
    "title": "任务2标题",
    "description": "任务2描述",
    "details": "实现细节",
    "testStrategy": "1. API测试：验证请求响应格式\n2. 异常测试：模拟错误场景",
    "priority": "medium",
    "dependencies": [0]
  }
]

请只返回 JSON 数组，不要包含其他内容。`,
		typeGuidance,
		requirement.Title,
		requirement.Content,
	)
}

// buildSplitRequirementPromptWithLanguage 构建带语言信息的需求拆分提示词
func (s *Service) buildSplitRequirementPromptWithLanguage(requirement *models.Requirement, taskType string, language *models.Language) string {
	// 根据任务类型生成不同的提示
	var typeGuidance string
	switch taskType {
	case "frontend":
		typeGuidance = `IMPORTANT: 只需要生成前端相关的任务，包括：
- 前端页面和组件开发
- 前端路由和状态管理
- 前端 API 调用和数据展示
- 前端样式和交互效果
- 前端表单验证和处理

不要包含后端 API 开发、数据库设计、后端逻辑等后端任务。`
	case "backend":
		typeGuidance = `IMPORTANT: 只需要生成后端相关的任务，包括：
- 后端 API 接口开发
- 数据库设计和操作
- 后端业务逻辑实现
- 后端中间件和服务层
- 后端数据验证和处理

不要包含前端页面、组件、样式等前端任务。`
	case "fullstack":
		typeGuidance = `需要生成前端和后端的完整任务，包括：
- 前端页面和组件开发
- 后端 API 接口开发
- 数据库设计和操作
- 前后端联调
- 完整的功能实现`
	default:
		typeGuidance = `需要生成后端相关的任务。`
	}

	// 构建语言/技术栈说明
	var techStackGuidance string
	if language != nil {
		techStackGuidance = fmt.Sprintf(`
技术栈说明：
- 语言：%s
- 框架：%s
- 描述：%s
`, language.DisplayName, language.Framework, language.Description)
		if language.CodeHints != "" {
			techStackGuidance += fmt.Sprintf(`
代码提示：%s
`, language.CodeHints)
		}
	} else {
		techStackGuidance = `
技术栈说明：
- 后端：Go 语言，使用 Gin 框架和 GORM
- API 规范：使用 GET 和 POST 请求，删除操作使用 POST + /delete 路径
`
	}

	return fmt.Sprintf(`你是一个AI助手，帮助将产品需求文档（PRD）拆分为开发任务。

请分析以下需求并拆分为合适的开发任务。每个任务应该是一个具体的、可执行的开发单元。

%s
%s
需求标题：%s
需求内容：
%s

请以 JSON 数组格式返回任务列表，每个任务包含以下字段：
- title: 任务标题（简洁明了）
- description: 任务描述（详细说明要做什么）
- details: 实现细节（技术方案、注意事项等）
- testStrategy: 测试策略（如何测试这个任务，包括单元测试、集成测试、边界条件等）
- priority: 优先级（high/medium/low）
- dependencies: 依赖的任务索引数组（如 [0, 1] 表示依赖第1和第2个任务）

拆分原则：
1. 按照功能模块拆分，每个任务专注于一个功能点
2. 优先级设置：基础架构和高优先级功能设为 high，核心功能设为 medium，辅助功能设为 low
3. 合理设置依赖关系，确保任务可以按顺序执行
4. 每个任务应该足够独立，可以单独开发和测试
5. 为每个任务提供明确的测试策略，说明如何验证功能正确性

示例格式：
[
  {
    "title": "任务1标题",
    "description": "任务1描述",
    "details": "实现细节",
    "testStrategy": "1. 单元测试：测试核心逻辑\n2. 边界测试：检查空值、边界值\n3. 集成测试：验证与其他模块的交互",
    "priority": "high",
    "dependencies": []
  },
  {
    "title": "任务2标题",
    "description": "任务2描述",
    "details": "实现细节",
    "testStrategy": "1. API测试：验证请求响应格式\n2. 异常测试：模拟错误场景",
    "priority": "medium",
    "dependencies": [0]
  }
]

请只返回 JSON 数组，不要包含其他内容。`,
		typeGuidance,
		techStackGuidance,
		requirement.Title,
		requirement.Content,
	)
}

// parseTasks 解析任务列表
func (s *Service) parseTasks(response string, requirementID uint64) ([]models.Task, error) {
	// 尝试从响应中提取 JSON 数组
	jsonStart := strings.Index(response, "[")
	jsonEnd := strings.LastIndex(response, "]")

	if jsonStart == -1 || jsonEnd == -1 || jsonEnd < jsonStart {
		return nil, fmt.Errorf("could not find valid JSON array in response")
	}

	jsonStr := response[jsonStart : jsonEnd+1]

	var tasksData []struct {
		Title        string `json:"title"`
		Description  string `json:"description"`
		Details      string `json:"details"`
		TestStrategy string `json:"testStrategy"`
		Priority     string `json:"priority"`
	}

	if err := json.Unmarshal([]byte(jsonStr), &tasksData); err != nil {
		return nil, fmt.Errorf("failed to parse tasks: %w", err)
	}

	reqID := requirementID
	result := make([]models.Task, len(tasksData))
	for i, td := range tasksData {
		priority := td.Priority
		if priority == "" {
			priority = "medium"
		}

		result[i] = models.Task{
			RequirementID: &reqID,
			Title:         td.Title,
			Description:   td.Description,
			Details:       td.Details,
			TestStrategy:  td.TestStrategy,
			Status:        "pending",
			Priority:      priority,
		}
	}

	return result, nil
}

// extractJSONFromString 从字符串中提取 JSON
func extractJSONFromString(text string) string {
	// 尝试匹配 JSON 数组
	reArray := regexp.MustCompile(`\[[\s\S]*\]`)
	if match := reArray.FindString(text); match != "" {
		return match
	}

	// 尝试匹配 JSON 对象
	reObject := regexp.MustCompile(`\{[\s\S]*\}`)
	if match := reObject.FindString(text); match != "" {
		return match
	}

	return text
}

// parseIntOrDefault 解析整数或返回默认值
func parseIntOrDefault(s string, defaultVal int) int {
	if s == "" {
		return defaultVal
	}
	val, err := strconv.Atoi(s)
	if err != nil {
		return defaultVal
	}
	return val
}

// buildExpandPromptWithKnowledge 构建带知识库的任务展开提示词
func (s *Service) buildExpandPromptWithKnowledge(task *models.Task, knowledgeContext string) string {
	return fmt.Sprintf(`请将以下任务拆分为 %d 个子任务。每个子任务应该是一个具体的、可执行的步骤。

技术栈说明：
- 后端：Go 语言，使用 Gin 框架和 GORM
- API 规范：使用 GET 和 POST 请求，删除操作使用 POST + /delete 路径

%s

任务标题：%s
任务描述：%s
任务详情：%s

请以 JSON 数组格式返回子任务列表，每个子任务包含以下字段：
- title: 子任务标题（必填）
- description: 子任务描述（可选）
- details: 详细说明（可选）
- priority: 优先级（high/medium/low，默认medium）
- codeInterface: 代码接口定义，JSON对象包含：
  - name: 函数/方法名称
  - inputs: 输入参数类型定义（TypeScript格式）
  - outputs: 输出类型定义（TypeScript格式）
  - example: 使用示例代码
- acceptanceCriteria: 验收标准数组，每项包含：
  - description: 验收条件描述
  - completed: 是否完成（默认false）
- relatedFiles: 关联的源文件路径数组（如 ["src/auth/service.ts", "src/auth/types.ts"]）
- codeHints: 代码实现提示和建议

示例格式：
[
  {
    "title": "验证用户登录",
    "description": "实现用户登录验证逻辑",
    "details": "校验用户名密码，生成JWT token",
    "priority": "high",
    "codeInterface": {
      "name": "validateLogin",
      "inputs": "{ username: string; password: string }",
      "outputs": "{ success: boolean; token?: string; error?: string }",
      "example": "const result = await validateLogin({ username: 'admin', password: '123' });"
    },
    "acceptanceCriteria": [
      {"description": "用户名或密码为空返回 success: false", "completed": false},
      {"description": "密码错误返回 success: false 且 error 包含错误信息", "completed": false},
      {"description": "验证成功返回 token", "completed": false}
    ],
    "relatedFiles": ["src/auth/service.ts", "src/auth/types.ts"],
    "codeHints": "使用 bcrypt.compare 验证密码，使用 jsonwebtoken 生成 token"
  }
]

请只返回 JSON 数组，不要包含其他内容。`,
		3, // 默认子任务数量
		knowledgeContext,
		task.Title,
		task.Description,
		task.Details,
	)
}

// buildExpandPromptWithResearch 构建带研究结果的任务展开提示词
func (s *Service) buildExpandPromptWithResearch(task *models.Task, researchResult string) string {
	return fmt.Sprintf(`基于研究结果，请将以下任务拆分为 %d 个子任务。每个子任务应该是一个具体的、可执行的步骤。

技术栈说明：
- 后端：Go 语言，使用 Gin 框架和 GORM
- API 规范：使用 GET 和 POST 请求，删除操作使用 POST + /delete 路径

研究结果：
%s

任务标题：%s
任务描述：%s
任务详情：%s

请根据研究结果中的最佳实践和建议来拆分任务。以 JSON 数组格式返回子任务列表，每个子任务包含以下字段：
- title: 子任务标题（必填）
- description: 子任务描述（可选）
- details: 详细说明（可选）
- priority: 优先级（high/medium/low，默认medium）
- codeInterface: 代码接口定义
- acceptanceCriteria: 验收标准数组
- relatedFiles: 关联的源文件路径数组
- codeHints: 代码实现提示和建议

请只返回 JSON 数组，不要包含其他内容。`,
		3, // 默认子任务数量
		researchResult,
		task.Title,
		task.Description,
		task.Details,
	)
}
