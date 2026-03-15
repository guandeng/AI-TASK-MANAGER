package handlers

import (
	"github.com/ai-task-manager/backend/internal/config"
	"github.com/ai-task-manager/backend/pkg/ai"
	"github.com/ai-task-manager/backend/pkg/response"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// PromptScoreHandler 提示词质量评分处理器
type PromptScoreHandler struct {
	aiService *ai.Service
	logger    *zap.Logger
}

// NewPromptScoreHandler 创建提示词评分处理器
func NewPromptScoreHandler(logger *zap.Logger, cfg *config.Config) *PromptScoreHandler {
	var aiSvc *ai.Service
	if cfg != nil && cfg.AI.Provider != "" {
		aiSvc = ai.NewService(&cfg.AI)
	}
	return &PromptScoreHandler{
		aiService: aiSvc,
		logger:    logger,
	}
}

// ScoreRequirementSplitPrompt 对需求拆分任务的提示词进行评分
func (h *PromptScoreHandler) ScoreRequirementSplitPrompt(c *gin.Context) {
	var req struct {
		RequirementTitle   string `json:"requirementTitle"`
		RequirementContent string `json:"requirementContent"`
		TaskType           string `json:"taskType"` // frontend, backend, fullstack
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数格式错误")
		return
	}

	if req.RequirementTitle == "" || req.RequirementContent == "" {
		response.BadRequest(c, "需求标题和内容不能为空")
		return
	}

	// 构建评分提示词
	prompt := h.buildRequirementScorePrompt(req.RequirementTitle, req.RequirementContent, req.TaskType)

	// 调用 AI 进行评分
	result, err := h.aiService.Chat(prompt)
	if err != nil {
		h.logger.Error("AI 评分失败", zap.Error(err))
		response.Error(c, 500, "AI 评分失败："+err.Error())
		return
	}

	response.Success(c, gin.H{
		"prompt":         prompt,
		"scoreResult":    result,
		"requirementTitle": req.RequirementTitle,
		"taskType":       req.TaskType,
	})
}

// ScoreTaskExpandPrompt 对任务拆分子任务的提示词进行评分
func (h *PromptScoreHandler) ScoreTaskExpandPrompt(c *gin.Context) {
	var req struct {
		TaskTitle       string `json:"taskTitle"`
		TaskDescription string `json:"taskDescription"`
		TaskDetails     string `json:"taskDetails"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数格式错误")
		return
	}

	if req.TaskTitle == "" {
		response.BadRequest(c, "任务标题不能为空")
		return
	}

	// 构建评分提示词
	prompt := h.buildTaskExpandScorePrompt(req.TaskTitle, req.TaskDescription, req.TaskDetails)

	// 调用 AI 进行评分
	result, err := h.aiService.Chat(prompt)
	if err != nil {
		h.logger.Error("AI 评分失败", zap.Error(err))
		response.Error(c, 500, "AI 评分失败："+err.Error())
		return
	}

	response.Success(c, gin.H{
		"prompt":        prompt,
		"scoreResult":   result,
		"taskTitle":     req.TaskTitle,
		"taskDescription": req.TaskDescription,
	})
}

// buildRequirementScorePrompt 构建需求拆分提示词的评分提示
func (h *PromptScoreHandler) buildRequirementScorePrompt(title, content, taskType string) string {
	return `你是一个专业的提示词工程师，专门评估用于将产品需求拆分为开发任务的提示词质量。

请评估以下提示词的质量，并从以下维度进行打分（每项 1-10 分）：

1. 清晰度 (Clarity)：提示词是否清晰明确，无歧义
2. 完整性 (Completeness)：是否包含了必要的上下文和信息
3. 结构化 (Structure)：输出格式要求是否结构化、易于解析
4. 可执行性 (Actionability)：AI 是否能基于提示词生成可执行的任务
5. 一致性 (Consistency)：与项目技术栈和规范的兼容性

评分标准：
- 9-10 分：优秀，几乎完美
- 7-8 分：良好，有小幅改进空间
- 5-6 分：一般，需要明显改进
- 3-4 分：较差，存在严重问题
- 1-2 分：很差，几乎无法使用

需求信息：
- 标题：` + title + `
- 类型：` + taskType + `
- 内容：` + content + `

当前使用的提示词模板：
"""
你是一个 AI 助手，帮助将产品需求文档（PRD）拆分为开发任务。

请分析以下需求并拆分为合适的开发任务。每个任务应该是一个具体的、可执行的开发单元。

需要生成前端和后端的完整任务，包括：
- 前端页面和组件开发
- 前端路由和状态管理
- 前端 API 调用和数据展示
- 前端样式和交互效果
- 前端表单验证和处理
- 后端 API 接口开发
- 数据库设计和操作
- 前后端联调
- 完整的功能实现

需求标题：[需求标题]
需求内容：
[需求内容]

请以 JSON 数组格式返回任务列表，每个任务必须包含以下 8 个字段：
- title: 任务名称（清晰、可搜索）
- module: 模块归属（MCP 接入/AI 能力/数据处理/接口封装）
- input: 输入（依赖什么：结构、接口、权限、环境）
- output: 输出（交付物：代码、文档、接口、Demo）
- acceptanceCriteria: 验收标准（可测、可看、可验证的数组）
- risk: 风险点（可能的技术风险、依赖风险等）
- priority: 优先级（high/medium/low）
- estimatedHours: 预估工时（小时数）

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

请只返回 JSON 数组，不要包含其他内容。
"""

请以 JSON 格式返回评分结果：
{
  "scores": {
    "clarity": 分数，
    "completeness": 分数，
    "structure": 分数，
    "actionability": 分数，
    "consistency": 分数
  },
  "totalScore": 总分（5 项平均分，保留 1 位小数）,
  "strengths": ["优点 1", "优点 2", ...],
  "weaknesses": ["缺点 1", "缺点 2", ...],
  "suggestions": [
    {
      "issue": "问题描述",
      "suggestion": "改进建议"
    }
  ],
  "analysis": "详细分析报告（200-500 字）"
}

请只返回 JSON，不要包含其他内容。`
}

// buildTaskExpandScorePrompt 构建任务展开提示词的评分提示
func (h *PromptScoreHandler) buildTaskExpandScorePrompt(title, description, details string) string {
	return `你是一个专业的提示词工程师，专门评估用于将任务拆分为子任务的提示词质量。

请评估以下提示词的质量，并从以下维度进行打分（每项 1-10 分）：

1. 清晰度 (Clarity)：提示词是否清晰明确，无歧义
2. 完整性 (Completeness)：是否包含了必要的上下文和信息
3. 结构化 (Structure)：输出格式要求是否结构化、易于解析
4. 可执行性 (Actionability)：AI 是否能基于提示词生成可执行的子任务
5. 一致性 (Consistency)：与项目技术栈和规范的兼容性

评分标准：
- 9-10 分：优秀，几乎完美
- 7-8 分：良好，有小幅改进空间
- 5-6 分：一般，需要明显改进
- 3-4 分：较差，存在严重问题
- 1-2 分：很差，几乎无法使用

任务信息：
- 标题：` + title + `
- 描述：` + description + `
- 详情：` + details + `

当前使用的提示词模板：
"""
请将以下任务拆分为 3 个子任务。每个子任务应该是一个具体的、可执行的步骤。

技术栈说明：
- 后端：Go 语言，使用 Gin 框架和 GORM
- API 规范：使用 GET 和 POST 请求，删除操作使用 POST + /delete 路径

任务标题：[任务标题]
任务描述：[任务描述]
任务详情：[任务详情]

请以 JSON 数组格式返回子任务列表，每个子任务包含以下字段：
- title: 子任务标题（必填）
- description: 子任务描述（可选）
- details: 详细说明（可选）
- priority: 优先级（high/medium/low，默认 medium）
- codeInterface: 代码接口定义，JSON 对象包含：
  - name: 函数/方法名称
  - inputs: 输入参数类型定义（TypeScript 格式）
  - outputs: 输出类型定义（TypeScript 格式）
  - example: 使用示例代码
- acceptanceCriteria: 验收标准数组，每项包含：
  - description: 验收条件描述
  - completed: 是否完成（默认 false）
- relatedFiles: 关联的源文件路径数组（如 ["src/auth/service.ts", "src/auth/types.ts"]）
- codeHints: 代码实现提示和建议

示例格式：
[
  {
    "title": "验证用户登录",
    "description": "实现用户登录验证逻辑",
    "details": "校验用户名密码，生成 JWT token",
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

请只返回 JSON 数组，不要包含其他内容。
"""

请以 JSON 格式返回评分结果：
{
  "scores": {
    "clarity": 分数，
    "completeness": 分数，
    "structure": 分数，
    "actionability": 分数，
    "consistency": 分数
  },
  "totalScore": 总分（5 项平均分，保留 1 位小数）,
  "strengths": ["优点 1", "优点 2", ...],
  "weaknesses": ["缺点 1", "缺点 2", ...],
  "suggestions": [
    {
      "issue": "问题描述",
      "suggestion": "改进建议"
    }
  ],
  "analysis": "详细分析报告（200-500 字）"
}

请只返回 JSON，不要包含其他内容。`
}
