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

// TaskWithDependencies 带依赖索引的任务
type TaskWithDependencies struct {
	Task         models.Task
	Dependencies []int // 依赖的任务索引
}

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
func (s *Service) SplitRequirement(requirement *models.Requirement, taskType string) ([]TaskWithDependencies, error) {
	return s.SplitRequirementWithLanguage(requirement, taskType, nil)
}

// SplitRequirementWithLanguage 将需求拆分为任务（带语言信息）
// taskType: frontend, backend, fullstack
func (s *Service) SplitRequirementWithLanguage(requirement *models.Requirement, taskType string, language *models.Language) ([]TaskWithDependencies, error) {
	return s.SplitRequirementWithTemplate(requirement, taskType, language, nil)
}

// SplitRequirementWithTemplate 将需求拆分为任务（带模板字段定义）
// taskType: frontend, backend, fullstack
// fieldSchema: 模板字段定义 JSON
func (s *Service) SplitRequirementWithTemplate(requirement *models.Requirement, taskType string, language *models.Language, fieldSchema *string) ([]TaskWithDependencies, error) {
	prompt := s.buildSplitRequirementPromptWithTemplate(requirement, taskType, language, fieldSchema)
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

	var subtasksData []map[string]interface{}

	if err := json.Unmarshal([]byte(jsonStr), &subtasksData); err != nil {
		return nil, fmt.Errorf("failed to parse subtasks: %w", err)
	}

	result := make([]models.Subtask, len(subtasksData))
	for i, st := range subtasksData {
		// 处理优先级
		priority := getString(st["priority"])
		if priority == "" {
			priority = "medium"
		}

		// 处理已知字段
		title := getString(st["title"])
		description := getString(st["description"])
		details := getString(st["details"])
		codeInterface := st["codeInterface"]
		acceptanceCriteria := st["acceptanceCriteria"]
		relatedFiles := st["relatedFiles"]
		codeHints := getStringPtr(st["codeHints"])

		// 处理 codeInterface -> JSON 字符串
		var codeInterfaceStr *string
		if codeInterface != nil {
			jsonBytes, err := json.Marshal(codeInterface)
			if err == nil {
				str := string(jsonBytes)
				codeInterfaceStr = &str
			}
		}

		// 处理 acceptanceCriteria -> JSON 字符串
		var acceptanceCriteriaStr *string
		if ac, ok := acceptanceCriteria.([]interface{}); ok && len(ac) > 0 {
			jsonBytes, err := json.Marshal(acceptanceCriteria)
			if err == nil {
				str := string(jsonBytes)
				acceptanceCriteriaStr = &str
			}
		}

		// 处理 relatedFiles -> JSON 字符串
		var relatedFilesStr *string
		if rf, ok := relatedFiles.([]interface{}); ok && len(rf) > 0 {
			jsonBytes, err := json.Marshal(relatedFiles)
			if err == nil {
				str := string(jsonBytes)
				relatedFilesStr = &str
			}
		}

		// 处理 codeHints
		var codeHintsStr *string
		if codeHints != nil && *codeHints != "" {
			codeHintsStr = codeHints
		}

		// 处理 CustomFields - 存储所有非已知字段
		customFields := make(map[string]interface{})
		knownFields := map[string]bool{
			"title": true, "description": true, "details": true,
			"priority": true, "codeInterface": true, "acceptanceCriteria": true,
			"relatedFiles": true, "codeHints": true,
		}

		for key, value := range st {
			if !knownFields[key] && value != nil {
				customFields[key] = value
			}
		}

		// 序列化 CustomFields 为 JSON 字符串
		var customFieldsStr *string
		if len(customFields) > 0 {
			jsonBytes, err := json.Marshal(customFields)
			if err == nil {
				str := string(jsonBytes)
				customFieldsStr = &str
			}
		}

		result[i] = models.Subtask{
			TaskID:             taskID,
			Title:              title,
			Description:        description,
			Details:            details,
			Status:             "pending",
			Priority:           priority,
			SortOrder:          uint(i),
			CodeInterface:      codeInterfaceStr,
			AcceptanceCriteria: acceptanceCriteriaStr,
			RelatedFiles:       relatedFilesStr,
			CodeHints:          codeHintsStr,
			CustomFields:       customFieldsStr,
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

请以 JSON 数组格式返回任务列表，每个任务必须包含以下 9 个字段：
- title: 任务名称（清晰、可搜索）
- module: 模块归属（MCP 接入/AI 能力/数据处理/接口封装）
- input: 输入（依赖什么：结构、接口、权限、环境）
- output: 输出（交付物：代码、文档、接口、Demo）
- acceptanceCriteria: 验收标准（可测、可看、可验证的数组）
- risk: 风险点（可能的技术风险、依赖风险等）
- priority: 优先级（high/medium/low）
- estimatedHours: 预估工时（小时数）
- details: 实现细节（具体的实现步骤、关键技术点、注意事项）

可选字段：
- description: 任务描述
- testStrategy: 测试策略
- dependencies: 依赖的任务索引数组

拆分原则：
1. 按照功能模块拆分，每个任务专注于一个功能点
2. 优先级设置：基础架构和高优先级功能设为 high，核心功能设为 medium，辅助功能设为 low
3. 合理设置依赖关系，确保任务可以按顺序执行
4. 每个任务应该足够独立，可以单独开发和测试
5. 为每个任务提供明确的测试策略，说明如何验证功能正确性
6. 预估工时应基于实现复杂度合理评估

示例格式：
[
  {
    "title": "用户登录接口开发",
    "module": "接口封装",
    "input": "用户表结构、JWT 配置、密码加密库",
    "output": "登录 API 接口、单元测试代码",
    "acceptanceCriteria": [
      "用户名密码验证正确返回 token",
      "密码错误返回 401 错误",
      "参数缺失返回 400 错误"
    ],
    "risk": "密码加密方式变更可能导致旧数据不兼容",
    "priority": "high",
    "estimatedHours": 4,
    "details": "1. 创建 User 模型，包含用户名、密码哈希等字段\n2. 实现密码加密函数，使用 bcrypt 算法\n3. 创建登录处理函数，验证用户名密码\n4. 密码正确时生成 JWT token 并返回\n5. 密码错误时返回 401 错误\n6. 添加输入参数验证中间件",
    "dependencies": []
  }
]

请只返回 JSON 数组，不要包含其他内容。`,
		typeGuidance,
		requirement.Title,
		requirement.Content,
	)
}

// buildSplitRequirementPromptWithTemplate 构建带模板字段定义的需求拆分提示词
func (s *Service) buildSplitRequirementPromptWithTemplate(requirement *models.Requirement, taskType string, language *models.Language, fieldSchema *string) string {
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

	// 构建字段定义说明
	var fieldSchemaGuidance string
	var fieldsSection string

	if fieldSchema != nil {
		var schema struct {
			Fields []struct {
				Name     string   `json:"name"`
				Label    string   `json:"label"`
				Type     string   `json:"type"`
				Options  []string `json:"options,omitempty"`
				Required bool     `json:"required"`
			} `json:"fields"`
		}
		if err := json.Unmarshal([]byte(*fieldSchema), &schema); err == nil && len(schema.Fields) > 0 {
			fieldSchemaGuidance = "根据项目模板定义，任务必须包含以下字段："

			var fieldsList []string
			var requiredFields []string
			var optionalFields []string

			for _, field := range schema.Fields {
				fieldDesc := fmt.Sprintf("- %s: %s（%s", field.Name, field.Label, field.Type)
				if len(field.Options) > 0 {
					fieldDesc += fmt.Sprintf("，选项：%s", strings.Join(field.Options, "/"))
				}
				if field.Required {
					fieldDesc += "，必填）"
					requiredFields = append(requiredFields, field.Name)
				} else {
					fieldDesc += "，可选）"
					optionalFields = append(optionalFields, field.Name)
				}
				fieldsList = append(fieldsList, fieldDesc)
			}

			fieldsSection = strings.Join(fieldsList, "\n")

			// 构建示例对象
			exampleObj := make(map[string]interface{})
			for _, field := range schema.Fields {
				switch field.Name {
				case "module":
					if len(field.Options) > 0 {
						exampleObj["module"] = field.Options[0]
					} else {
						exampleObj["module"] = "模块名称"
					}
				case "input":
					exampleObj["input"] = "依赖什么：结构、接口、权限、环境"
				case "output":
					exampleObj["output"] = "交付物：代码、文档、接口、Demo"
				case "acceptanceCriteria":
					exampleObj["acceptanceCriteria"] = []string{"验收条件 1", "验收条件 2"}
				case "risk":
					exampleObj["risk"] = "可能的技术风险、依赖风险等"
				case "estimatedHours":
					exampleObj["estimatedHours"] = 4
				default:
					if field.Type == "text" || field.Type == "textarea" {
						exampleObj[field.Name] = "字段值"
					} else if field.Type == "number" {
						exampleObj[field.Name] = 0
					} else if field.Type == "array" {
						exampleObj[field.Name] = []string{"项目 1", "项目 2"}
					} else if field.Type == "select" && len(field.Options) > 0 {
						exampleObj[field.Name] = field.Options[0]
					} else {
						exampleObj[field.Name] = "字段值"
					}
				}
			}
			exampleJSON, _ := json.MarshalIndent(exampleObj, "    ", "  ")
			fieldsSection += fmt.Sprintf("\n\n示例字段格式：\n%s", string(exampleJSON))
		}
	}

	// 如果没有定义字段模式，使用默认 8 个字段
	if fieldsSection == "" {
		fieldsSection = `- title: 任务名称（清晰、可搜索）
- module: 模块归属（MCP 接入/AI 能力/数据处理/接口封装）
- input: 输入（依赖什么：结构、接口、权限、环境）
- output: 输出（交付物：代码、文档、接口、Demo）
- acceptanceCriteria: 验收标准（可测、可看、可验证的数组）
- risk: 风险点（可能的技术风险、依赖风险等）
- priority: 优先级（high/medium/low）
- estimatedHours: 预估工时（小时数）`
	}

	return fmt.Sprintf(`你是一个 AI 助手，帮助将产品需求文档（PRD）拆分为开发任务。

请分析以下需求并拆分为合适的开发任务。每个任务应该是一个具体的、可执行的开发单元。

%s
%s
需求标题：%s
需求内容：
%s

%s
%s

可选字段：
- description: 任务描述
- details: 实现细节
- testStrategy: 测试策略
- dependencies: 依赖的任务索引数组

拆分原则：
1. 按照功能模块拆分，每个任务专注于一个功能点
2. 优先级设置：基础架构和高优先级功能设为 high，核心功能设为 medium，辅助功能设为 low
3. 合理设置依赖关系，确保任务可以按顺序执行
4. 每个任务应该足够独立，可以单独开发和测试
5. 为每个任务提供明确的测试策略，说明如何验证功能正确性
6. 预估工时应基于实现复杂度合理评估

示例格式：
[
  {
    "title": "用户登录接口开发",
    "module": "接口封装",
    "input": "用户表结构、JWT 配置、密码加密库",
    "output": "登录 API 接口、单元测试代码",
    "acceptanceCriteria": [
      "用户名密码验证正确返回 token",
      "密码错误返回 401 错误",
      "参数缺失返回 400 错误"
    ],
    "risk": "密码加密方式变更可能导致旧数据不兼容",
    "priority": "high",
    "estimatedHours": 4,
    "dependencies": []
  }
]

请只返回 JSON 数组，不要包含其他内容。`,
		typeGuidance,
		techStackGuidance,
		requirement.Title,
		requirement.Content,
		fieldSchemaGuidance,
		fieldsSection,
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

请以 JSON 数组格式返回任务列表，每个任务必须包含以下 9 个字段：
- title: 任务名称（清晰、可搜索）
- module: 模块归属（MCP 接入/AI 能力/数据处理/接口封装）
- input: 输入（依赖什么：结构、接口、权限、环境）
- output: 输出（交付物：代码、文档、接口、Demo）
- acceptanceCriteria: 验收标准（可测、可看、可验证的数组）
- risk: 风险点（可能的技术风险、依赖风险等）
- priority: 优先级（high/medium/low）
- estimatedHours: 预估工时（小时数）
- details: 实现细节（具体的实现步骤、关键技术点、注意事项）

可选字段：
- description: 任务描述
- testStrategy: 测试策略
- dependencies: 依赖的任务索引数组

拆分原则：
1. 按照功能模块拆分，每个任务专注于一个功能点
2. 优先级设置：基础架构和高优先级功能设为 high，核心功能设为 medium，辅助功能设为 low
3. 合理设置依赖关系，确保任务可以按顺序执行
4. 每个任务应该足够独立，可以单独开发和测试
5. 为每个任务提供明确的测试策略，说明如何验证功能正确性
6. 预估工时应基于实现复杂度合理评估

示例格式：
[
  {
    "title": "用户登录接口开发",
    "module": "接口封装",
    "input": "用户表结构、JWT 配置、密码加密库",
    "output": "登录 API 接口、单元测试代码",
    "acceptanceCriteria": [
      "用户名密码验证正确返回 token",
      "密码错误返回 401 错误",
      "参数缺失返回 400 错误"
    ],
    "risk": "密码加密方式变更可能导致旧数据不兼容",
    "priority": "high",
    "estimatedHours": 4,
    "details": "1. 创建 User 模型，包含用户名、密码哈希等字段\n2. 实现密码加密函数，使用 bcrypt 算法\n3. 创建登录处理函数，验证用户名密码\n4. 密码正确时生成 JWT token 并返回\n5. 密码错误时返回 401 错误\n6. 添加输入参数验证中间件",
    "dependencies": []
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
func (s *Service) parseTasks(response string, requirementID uint64) ([]TaskWithDependencies, error) {
	// 尝试从响应中提取 JSON 数组
	jsonStart := strings.Index(response, "[")
	jsonEnd := strings.LastIndex(response, "]")

	if jsonStart == -1 || jsonEnd == -1 || jsonEnd < jsonStart {
		return nil, fmt.Errorf("could not find valid JSON array in response")
	}

	jsonStr := response[jsonStart : jsonEnd+1]

	// 使用 map 解析以支持动态字段
	var tasksData []map[string]interface{}

	if err := json.Unmarshal([]byte(jsonStr), &tasksData); err != nil {
		return nil, fmt.Errorf("failed to parse tasks: %w", err)
	}

	reqID := requirementID
	result := make([]TaskWithDependencies, len(tasksData))
	for i, td := range tasksData {
		// 提取已知字段
		title := getString(td["title"])
		description := getString(td["description"])
		details := getString(td["details"])
		testStrategy := getString(td["testStrategy"])
		priority := getString(td["priority"])
		if priority == "" {
			priority = "medium"
		}
		dependencies := getIntArray(td["dependencies"])

		// 提取独立字段（数据库已有字段）
		module := getStringPtr(td["module"])
		input := getStringPtr(td["input"])
		output := getStringPtr(td["output"])
		risk := getStringPtr(td["risk"])
		estimatedHours := getFloat64Ptr(td["estimatedHours"])

		// 处理 acceptanceCriteria -> JSON 字符串
		var acceptanceCriteriaStr *string
		if ac, ok := td["acceptanceCriteria"].([]interface{}); ok && len(ac) > 0 {
			jsonBytes, err := json.Marshal(ac)
			if err == nil {
				str := string(jsonBytes)
				acceptanceCriteriaStr = &str
			}
		}

		// 构建 CustomFields - 存储除已知字段外的所有字段
		customFields := make(map[string]interface{})
		knownFields := map[string]bool{
			"title": true, "description": true, "details": true,
			"testStrategy": true, "priority": true, "dependencies": true,
			"module": true, "input": true, "output": true, "risk": true,
			"estimatedHours": true, "acceptanceCriteria": true,
		}

		for key, value := range td {
			if !knownFields[key] && value != nil {
				customFields[key] = value
			}
		}

		// 序列化 CustomFields
		var customFieldsStr *string
		if len(customFields) > 0 {
			jsonBytes, err := json.Marshal(customFields)
			if err == nil {
				str := string(jsonBytes)
				customFieldsStr = &str
			}
		}

		result[i] = TaskWithDependencies{
			Task: models.Task{
				RequirementID:    &reqID,
				Title:            title,
				Description:      description,
				Details:          details,
				TestStrategy:     testStrategy,
				Status:           "pending",
				Priority:         priority,
				Module:           module,
				Input:            input,
				Output:           output,
				Risk:             risk,
				AcceptanceCriteria: acceptanceCriteriaStr,
				EstimatedHours:   estimatedHours,
				CustomFields:     customFieldsStr,
			},
			Dependencies: dependencies,
		}
	}

	return result, nil
}

// getString 从 interface{} 获取 string
func getString(v interface{}) string {
	if v == nil {
		return ""
	}
	if s, ok := v.(string); ok {
		return s
	}
	// 处理其他类型
	jsonBytes, _ := json.Marshal(v)
	return string(jsonBytes)
}

// getStringPtr 从 interface{} 获取 *string
func getStringPtr(v interface{}) *string {
	if v == nil {
		return nil
	}
	if s, ok := v.(string); ok {
		if s == "" {
			return nil
		}
		return &s
	}
	// 处理其他类型
	jsonBytes, _ := json.Marshal(v)
	str := string(jsonBytes)
	if str == "" || str == "null" {
		return nil
	}
	return &str
}

// getIntArray 从 interface{} 获取 []int
func getIntArray(v interface{}) []int {
	if v == nil {
		return nil
	}
	if arr, ok := v.([]interface{}); ok {
		result := make([]int, len(arr))
		for i, item := range arr {
			if num, ok := item.(float64); ok {
				result[i] = int(num)
			}
		}
		return result
	}
	if arr, ok := v.([]int); ok {
		return arr
	}
	return nil
}

// getFloat64Ptr 从 interface{} 获取 *float64
func getFloat64Ptr(v interface{}) *float64 {
	if v == nil {
		return nil
	}
	if num, ok := v.(float64); ok {
		if num == 0 {
			return nil
		}
		return &num
	}
	return nil
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
