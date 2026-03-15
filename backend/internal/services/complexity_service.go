package services

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/ai-task-manager/backend/internal/database"
	"github.com/ai-task-manager/backend/internal/models"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// ComplexityService 复杂度分析服务
type ComplexityService struct {
	logger *zap.Logger
	db     *gorm.DB
}

// NewComplexityService 创建复杂度分析服务
func NewComplexityService(logger *zap.Logger) *ComplexityService {
	return &ComplexityService{
		logger: logger,
		db:     database.GetDB(),
	}
}

// AnalyzeTask 分析单个任务的复杂度
func (s *ComplexityService) AnalyzeTask(task *models.Task, aiService AIService) (*models.ComplexityAnalysis, error) {
	// 构建分析提示词
	prompt := s.buildAnalysisPrompt(task)

	// 调用 AI 分析
	response, err := aiService.Chat(prompt)
	if err != nil {
		return nil, fmt.Errorf("AI 分析失败: %w", err)
	}

	// 解析结果
	analysis, err := s.parseAnalysis(response, task.ID)
	if err != nil {
		return nil, err
	}

	return analysis, nil
}

// AnalyzeTasks 批量分析任务复杂度
func (s *ComplexityService) AnalyzeTasks(requirementID uint64, taskType string, aiService AIService, knowledgeContext string) (*models.ComplexityReportData, error) {
	// 获取需求的所有任务
	var tasks []models.Task
	if err := s.db.Where("requirement_id = ?", requirementID).Find(&tasks).Error; err != nil {
		return nil, fmt.Errorf("获取任务列表失败: %w", err)
	}

	if len(tasks) == 0 {
		return nil, fmt.Errorf("没有可分析的任务")
	}

	// 构建批量分析提示词
	prompt := s.buildBatchAnalysisPrompt(tasks, knowledgeContext)

	// 调用 AI 分析
	response, err := aiService.Chat(prompt)
	if err != nil {
		return nil, fmt.Errorf("AI 分析失败: %w", err)
	}

	// 解析结果
	report, err := s.parseBatchAnalysis(response, tasks)
	if err != nil {
		return nil, err
	}

	return report, nil
}

// AnalyzeRequirement 分析需求的复杂度并生成报告
func (s *ComplexityService) AnalyzeRequirement(requirementID uint64, requirement *models.Requirement, aiService AIService, knowledgeContext string) (*models.ComplexityReportData, error) {
	// 构建需求分析提示词
	prompt := s.buildRequirementAnalysisPrompt(requirement, knowledgeContext)

	// 调用 AI 分析
	response, err := aiService.Chat(prompt)
	if err != nil {
		return nil, fmt.Errorf("AI 分析失败: %w", err)
	}

	// 解析结果
	report, err := s.parseBatchAnalysis(response, nil)
	if err != nil {
		return nil, err
	}

	return report, nil
}

// SaveReport 保存复杂度报告
func (s *ComplexityService) SaveReport(requirementID *uint64, reportData *models.ComplexityReportData) (*models.TaskComplexityReport, error) {
	reportJSON, err := json.Marshal(reportData)
	if err != nil {
		return nil, fmt.Errorf("序列化报告失败: %w", err)
	}

	report := &models.TaskComplexityReport{
		RequirementID: requirementID,
		Status:        "completed",
		ReportData:    string(reportJSON),
	}

	if err := s.db.Create(report).Error; err != nil {
		return nil, fmt.Errorf("保存报告失败: %w", err)
	}

	return report, nil
}

// GetReport 获取复杂度报告
func (s *ComplexityService) GetReport(reportID uint64) (*models.TaskComplexityReport, error) {
	var report models.TaskComplexityReport
	if err := s.db.First(&report, reportID).Error; err != nil {
		return nil, err
	}
	return &report, nil
}

// GetReportsByRequirement 获取需求的所有复杂度报告
func (s *ComplexityService) GetReportsByRequirement(requirementID uint64) ([]models.TaskComplexityReport, error) {
	var reports []models.TaskComplexityReport
	if err := s.db.Where("requirement_id = ?", requirementID).Order("created_at DESC").Find(&reports).Error; err != nil {
		return nil, err
	}
	return reports, nil
}

// buildAnalysisPrompt 构建单个任务分析提示词
func (s *ComplexityService) buildAnalysisPrompt(task *models.Task) string {
	return fmt.Sprintf(`请分析以下任务的复杂度，并返回 JSON 格式的分析结果。

任务标题：%s
任务描述：%s
任务详情：%s
测试策略：%s

请返回以下 JSON 格式（只返回 JSON，不要其他内容）：
{
  "taskId": %d,
  "taskTitle": "%s",
  "complexityScore": 5,
  "complexityLevel": "medium",
  "reasoning": "分析理由...",
  "subtaskCount": 3,
  "timeEstimate": "2-4 hours",
  "dependencies": [],
  "riskFactors": ["风险1", "风险2"]
}

复杂度评分标准（1-10）：
1-3: 低复杂度 - 简单的增删改查，无复杂逻辑
4-6: 中等复杂度 - 需要一定的业务逻辑处理
7-10: 高复杂度 - 涉及多个系统集成、复杂算法或安全要求

请只返回 JSON 对象。`,
		task.Title,
		task.Description,
		task.Details,
		task.TestStrategy,
		task.ID,
		task.Title,
	)
}

// buildBatchAnalysisPrompt 构建批量分析提示词
func (s *ComplexityService) buildBatchAnalysisPrompt(tasks []models.Task, knowledgeContext string) string {
	var taskList strings.Builder
	for i, task := range tasks {
		taskList.WriteString(fmt.Sprintf("\n任务 %d (ID: %d):\n", i+1, task.ID))
		taskList.WriteString(fmt.Sprintf("  标题: %s\n", task.Title))
		taskList.WriteString(fmt.Sprintf("  描述: %s\n", task.Description))
		taskList.WriteString(fmt.Sprintf("  详情: %s\n", task.Details))
		taskList.WriteString(fmt.Sprintf("  优先级: %s\n", task.Priority))
	}

	knowledgeSection := ""
	if knowledgeContext != "" {
		knowledgeSection = fmt.Sprintf("\n\n业务知识背景：\n%s", knowledgeContext)
	}

	return fmt.Sprintf(`请分析以下任务的复杂度，并返回 JSON 格式的分析结果。%s

任务列表：%s

请返回以下 JSON 格式（只返回 JSON 数组，不要其他内容）：
{
  "analyses": [
    {
      "taskId": 1,
      "taskTitle": "任务标题",
      "complexityScore": 5,
      "complexityLevel": "medium",
      "reasoning": "分析理由",
      "subtaskCount": 3,
      "timeEstimate": "2-4 hours",
      "dependencies": [],
      "riskFactors": []
    }
  ],
  "summary": {
    "totalTasks": 10,
    "lowComplexity": 3,
    "mediumComplexity": 5,
    "highComplexity": 2,
    "averageScore": 5
  }
}

复杂度评分标准（1-10）：
1-3: 低复杂度
4-6: 中等复杂度
7-10: 高复杂度

请只返回 JSON 对象。`,
		knowledgeSection,
		taskList.String(),
	)
}

// buildRequirementAnalysisPrompt 构建需求分析提示词
func (s *ComplexityService) buildRequirementAnalysisPrompt(requirement *models.Requirement, knowledgeContext string) string {
	knowledgeSection := ""
	if knowledgeContext != "" {
		knowledgeSection = fmt.Sprintf("\n\n业务知识背景：\n%s", knowledgeContext)
	}

	return fmt.Sprintf(`请分析以下需求的复杂度，预估需要拆分为多少任务，并返回 JSON 格式的分析结果。%s

需求标题：%s
需求内容：
%s

请返回以下 JSON 格式（只返回 JSON，不要其他内容）：
{
  "analyses": [
    {
      "taskId": 0,
      "taskTitle": "预估任务标题",
      "complexityScore": 5,
      "complexityLevel": "medium",
      "reasoning": "分析理由",
      "subtaskCount": 3,
      "timeEstimate": "2-4 hours",
      "dependencies": [],
      "riskFactors": []
    }
  ],
  "summary": {
    "totalTasks": 10,
    "lowComplexity": 3,
    "mediumComplexity": 5,
    "highComplexity": 2,
    "averageScore": 5
  }
}

请预估需求可能拆分为哪些主要任务，并为每个任务进行复杂度评估。

请只返回 JSON 对象。`,
		knowledgeSection,
		requirement.Title,
		requirement.Content,
	)
}

// parseAnalysis 解析单个任务分析结果
func (s *ComplexityService) parseAnalysis(response string, taskID uint64) (*models.ComplexityAnalysis, error) {
	// 清理响应
	cleanedResponse := strings.TrimSpace(response)
	cleanedResponse = strings.TrimPrefix(cleanedResponse, "```json")
	cleanedResponse = strings.TrimPrefix(cleanedResponse, "```")
	cleanedResponse = strings.TrimSuffix(cleanedResponse, "```")
	cleanedResponse = strings.TrimSpace(cleanedResponse)

	// 提取 JSON
	jsonStart := strings.Index(cleanedResponse, "{")
	jsonEnd := strings.LastIndex(cleanedResponse, "}")
	if jsonStart == -1 || jsonEnd == -1 {
		return nil, fmt.Errorf("未找到有效的 JSON")
	}
	jsonStr := cleanedResponse[jsonStart : jsonEnd+1]

	var analysis models.ComplexityAnalysis
	if err := json.Unmarshal([]byte(jsonStr), &analysis); err != nil {
		return nil, fmt.Errorf("解析分析结果失败: %w", err)
	}

	// 确保 taskId 正确
	analysis.TaskID = int(taskID)

	return &analysis, nil
}

// parseBatchAnalysis 解析批量分析结果
func (s *ComplexityService) parseBatchAnalysis(response string, tasks []models.Task) (*models.ComplexityReportData, error) {
	// 清理响应
	cleanedResponse := strings.TrimSpace(response)
	cleanedResponse = strings.TrimPrefix(cleanedResponse, "```json")
	cleanedResponse = strings.TrimPrefix(cleanedResponse, "```")
	cleanedResponse = strings.TrimSuffix(cleanedResponse, "```")
	cleanedResponse = strings.TrimSpace(cleanedResponse)

	// 提取 JSON
	jsonStart := strings.Index(cleanedResponse, "{")
	jsonEnd := strings.LastIndex(cleanedResponse, "}")
	if jsonStart == -1 || jsonEnd == -1 {
		return nil, fmt.Errorf("未找到有效的 JSON")
	}
	jsonStr := cleanedResponse[jsonStart : jsonEnd+1]

	var report models.ComplexityReportData
	if err := json.Unmarshal([]byte(jsonStr), &report); err != nil {
		return nil, fmt.Errorf("解析分析结果失败: %w", err)
	}

	return &report, nil
}
