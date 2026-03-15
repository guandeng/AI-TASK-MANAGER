package services

import (
	"encoding/json"
	"testing"

	"github.com/ai-task-manager/backend/internal/models"
	"go.uber.org/zap"
)

// MockAIService 模拟 AI 服务
type MockAIService struct {
	ChatResponse    string
	ResearchResponse string
	ChatError       error
}

func (m *MockAIService) Chat(prompt string) (string, error) {
	if m.ChatError != nil {
		return "", m.ChatError
	}
	return m.ChatResponse, nil
}

func (m *MockAIService) Research(prompt string) (string, error) {
	if m.ChatError != nil {
		return "", m.ChatError
	}
	if m.ResearchResponse != "" {
		return m.ResearchResponse, nil
	}
	return m.ChatResponse, nil
}

func (m *MockAIService) ExpandTask(task *models.Task) ([]models.Subtask, error) {
	return nil, nil
}

func (m *MockAIService) GenerateSubtask(task *models.Task) (*models.Subtask, error) {
	return nil, nil
}

func setupComplexityTest() *ComplexityService {
	return NewComplexityService(zap.NewNop())
}

func TestComplexityService_BuildAnalysisPrompt(t *testing.T) {
	service := setupComplexityTest()

	task := &models.Task{
		ID:          1,
		Title:       "实现用户登录功能",
		Description: "实现基于 JWT 的用户登录认证",
		Details:     "使用 bcrypt 加密密码，JWT 生成 token",
		TestStrategy: "测试登录成功和失败场景",
	}

	prompt := service.buildAnalysisPrompt(task)

	// 验证提示词包含关键信息
	if !contains(prompt, task.Title) {
		t.Error("提示词应包含任务标题")
	}
	if !contains(prompt, task.Description) {
		t.Error("提示词应包含任务描述")
	}
	if !contains(prompt, "复杂度") {
		t.Error("提示词应包含复杂度分析要求")
	}
	if !contains(prompt, "JSON") {
		t.Error("提示词应要求返回 JSON 格式")
	}
}

func TestComplexityService_ParseAnalysis(t *testing.T) {
	service := setupComplexityTest()

	tests := []struct {
		name        string
		response    string
		taskID      uint64
		expectScore int
		expectLevel string
		expectError bool
	}{
		{
			name: "有效响应",
			response: `{
				"taskId": 1,
				"taskTitle": "测试任务",
				"complexityScore": 5,
				"complexityLevel": "medium",
				"reasoning": "需要中等复杂度的逻辑处理",
				"subtaskCount": 3,
				"timeEstimate": "2-4 hours",
				"dependencies": [],
				"riskFactors": ["需要处理边界情况"]
			}`,
			taskID:      1,
			expectScore: 5,
			expectLevel: "medium",
			expectError: false,
		},
		{
			name: "带 markdown 代码块",
			response: "```json\n" + `{
				"taskId": 2,
				"taskTitle": "测试任务2",
				"complexityScore": 8,
				"complexityLevel": "high",
				"reasoning": "涉及复杂算法",
				"subtaskCount": 5,
				"timeEstimate": "1-2 days",
				"dependencies": [1],
				"riskFactors": ["性能优化"]
			}` + "\n```",
			taskID:      2,
			expectScore: 8,
			expectLevel: "high",
			expectError: false,
		},
		{
			name:        "无效 JSON",
			response:    "not a valid json",
			taskID:      1,
			expectError: true,
		},
		{
			name:        "空响应",
			response:    "",
			taskID:      1,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			analysis, err := service.parseAnalysis(tt.response, tt.taskID)

			if tt.expectError {
				if err == nil {
					t.Error("期望返回错误，但成功了")
				}
				return
			}

			if err != nil {
				t.Errorf("不期望错误: %v", err)
				return
			}

			if analysis.ComplexityScore != tt.expectScore {
				t.Errorf("期望分数 %d, 实际 %d", tt.expectScore, analysis.ComplexityScore)
			}
			if analysis.ComplexityLevel != tt.expectLevel {
				t.Errorf("期望级别 %s, 实际 %s", tt.expectLevel, analysis.ComplexityLevel)
			}
			if analysis.TaskID != int(tt.taskID) {
				t.Errorf("期望任务 ID %d, 实际 %d", tt.taskID, analysis.TaskID)
			}
		})
	}
}

func TestComplexityService_ParseBatchAnalysis(t *testing.T) {
	service := setupComplexityTest()

	tests := []struct {
		name          string
		response      string
		expectCount   int
		expectTotal   int
		expectError   bool
	}{
		{
			name: "有效批量响应",
			response: `{
				"analyses": [
					{
						"taskId": 1,
						"taskTitle": "任务1",
						"complexityScore": 3,
						"complexityLevel": "low",
						"reasoning": "简单任务",
						"subtaskCount": 2,
						"timeEstimate": "1 hour",
						"dependencies": [],
						"riskFactors": []
					},
					{
						"taskId": 2,
						"taskTitle": "任务2",
						"complexityScore": 7,
						"complexityLevel": "high",
						"reasoning": "复杂任务",
						"subtaskCount": 5,
						"timeEstimate": "1 day",
						"dependencies": [1],
						"riskFactors": ["需要仔细测试"]
					}
				],
				"summary": {
					"totalTasks": 2,
					"lowComplexity": 1,
					"mediumComplexity": 0,
					"highComplexity": 1,
					"averageScore": 5
				}
			}`,
			expectCount: 2,
			expectTotal: 2,
			expectError: false,
		},
		{
			name:        "无效 JSON",
			response:    "invalid",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			report, err := service.parseBatchAnalysis(tt.response, nil)

			if tt.expectError {
				if err == nil {
					t.Error("期望返回错误")
				}
				return
			}

			if err != nil {
				t.Errorf("不期望错误: %v", err)
				return
			}

			if len(report.Analyses) != tt.expectCount {
				t.Errorf("期望 %d 个分析, 实际 %d", tt.expectCount, len(report.Analyses))
			}
			if report.Summary.TotalTasks != tt.expectTotal {
				t.Errorf("期望总任务数 %d, 实际 %d", tt.expectTotal, report.Summary.TotalTasks)
			}
		})
	}
}

func TestComplexityService_BuildBatchAnalysisPrompt(t *testing.T) {
	service := setupComplexityTest()

	tasks := []models.Task{
		{ID: 1, Title: "任务1", Description: "描述1", Details: "详情1", Priority: "high"},
		{ID: 2, Title: "任务2", Description: "描述2", Details: "详情2", Priority: "medium"},
	}

	// 无知识库
	prompt1 := service.buildBatchAnalysisPrompt(tasks, "")
	if !contains(prompt1, "任务1") || !contains(prompt1, "任务2") {
		t.Error("提示词应包含所有任务")
	}
	if contains(prompt1, "知识库") {
		t.Error("不应包含知识库引用")
	}

	// 带知识库
	prompt2 := service.buildBatchAnalysisPrompt(tasks, "这是业务知识背景...")
	if !contains(prompt2, "业务知识") {
		t.Error("应包含知识库引用")
	}
}

func TestComplexityService_BuildRequirementAnalysisPrompt(t *testing.T) {
	service := setupComplexityTest()

	requirement := &models.Requirement{
		ID:      1,
		Title:   "用户管理系统",
		Content: "实现一个完整的用户管理系统，包括注册、登录、权限管理等功能",
	}

	prompt := service.buildRequirementAnalysisPrompt(requirement, "")

	if !contains(prompt, requirement.Title) {
		t.Error("提示词应包含需求标题")
	}
	if !contains(prompt, requirement.Content) {
		t.Error("提示词应包含需求内容")
	}
	if !contains(prompt, "预估") {
		t.Error("提示词应要求预估任务")
	}
}

func TestComplexityService_AnalyzeTask_Integration(t *testing.T) {
	service := setupComplexityTest()

	task := &models.Task{
		ID:          1,
		Title:       "测试任务",
		Description: "测试描述",
	}

	mockAI := &MockAIService{
		ChatResponse: `{
			"taskId": 1,
			"taskTitle": "测试任务",
			"complexityScore": 5,
			"complexityLevel": "medium",
			"reasoning": "测试用",
			"subtaskCount": 3,
			"timeEstimate": "2 hours",
			"dependencies": [],
			"riskFactors": []
		}`,
	}

	analysis, err := service.AnalyzeTask(task, mockAI)
	if err != nil {
		t.Errorf("分析失败: %v", err)
		return
	}

	if analysis.ComplexityScore != 5 {
		t.Errorf("期望分数 5, 实际 %d", analysis.ComplexityScore)
	}
	if analysis.ComplexityLevel != "medium" {
		t.Errorf("期望级别 medium, 实际 %s", analysis.ComplexityLevel)
	}
}

// 辅助函数
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// TestComplexityAnalysis_JSONMarshal 测试复杂度分析结构序列化
func TestComplexityAnalysis_JSONMarshal(t *testing.T) {
	analysis := models.ComplexityAnalysis{
		TaskID:          1,
		TaskTitle:       "测试任务",
		ComplexityScore: 5,
		ComplexityLevel: "medium",
		Reasoning:       "测试用",
		SubtaskCount:    3,
		TimeEstimate:    "2 hours",
		Dependencies:    []int{1, 2},
		RiskFactors:     []string{"风险1", "风险2"},
	}

	data, err := json.Marshal(analysis)
	if err != nil {
		t.Errorf("序列化失败: %v", err)
		return
	}

	var decoded models.ComplexityAnalysis
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Errorf("反序列化失败: %v", err)
		return
	}

	if decoded.TaskID != analysis.TaskID {
		t.Errorf("TaskID 不匹配")
	}
	if decoded.ComplexityScore != analysis.ComplexityScore {
		t.Errorf("ComplexityScore 不匹配")
	}
}

// TestComplexityReportData_JSONMarshal 测试报告数据序列化
func TestComplexityReportData_JSONMarshal(t *testing.T) {
	report := models.ComplexityReportData{
		Analyses: []models.ComplexityAnalysis{
			{TaskID: 1, TaskTitle: "任务1", ComplexityScore: 3, ComplexityLevel: "low"},
			{TaskID: 2, TaskTitle: "任务2", ComplexityScore: 7, ComplexityLevel: "high"},
		},
		Summary: models.ComplexitySummary{
			TotalTasks:       2,
			LowComplexity:    1,
			MediumComplexity: 0,
			HighComplexity:   1,
			AverageScore:     5,
		},
		GeneratedAt: "2024-01-01T00:00:00Z",
	}

	data, err := json.Marshal(report)
	if err != nil {
		t.Errorf("序列化失败: %v", err)
		return
	}

	var decoded models.ComplexityReportData
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Errorf("反序列化失败: %v", err)
		return
	}

	if len(decoded.Analyses) != 2 {
		t.Errorf("期望 2 个分析, 实际 %d", len(decoded.Analyses))
	}
	if decoded.Summary.TotalTasks != 2 {
		t.Errorf("期望总任务数 2, 实际 %d", decoded.Summary.TotalTasks)
	}
}
