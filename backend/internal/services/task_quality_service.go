package services

import (
	"encoding/json"
	"fmt"

	"github.com/ai-task-manager/backend/internal/models"
	"github.com/ai-task-manager/backend/internal/repository"
)

// TaskQualityService 任务质量评分服务接口
type TaskQualityService interface {
	// 评分
	Score(taskID uint64) (*models.TaskQualityScore, error)
	// 获取评分历史
	GetScoreHistory(taskID uint64, page, pageSize int) ([]models.TaskQualityScore, int64, error)
	// 获取评分详情
	GetScoreDetail(scoreID uint64) (*models.TaskQualityScore, error)
	// 删除评分
	DeleteScore(scoreID uint64) error
	// 恢复评分版本
	RestoreScore(scoreID uint64) error
}

// taskQualityService 任务质量评分服务实现
type taskQualityService struct {
	repo       repository.TaskQualityRepository
	taskRepo   repository.TaskRepository
	aiSvc      AIService
}

// NewTaskQualityService 创建任务质量评分服务
func NewTaskQualityService(aiSvc AIService) TaskQualityService {
	return &taskQualityService{
		repo:     repository.NewTaskQualityRepository(),
		taskRepo: repository.NewTaskRepository(),
		aiSvc:    aiSvc,
	}
}

// Score 对任务进行 AI 评分
func (s *taskQualityService) Score(taskID uint64) (*models.TaskQualityScore, error) {
	// 获取任务详情（包含子任务）
	task, err := s.taskRepo.GetByID(taskID)
	if err != nil {
		return nil, fmt.Errorf("获取任务失败：%w", err)
	}
	if task == nil {
		return nil, fmt.Errorf("任务不存在")
	}

	// 构建评分 Prompt
	prompt := s.buildScorePrompt(task)

	// 调用 AI 进行评分
	response, err := s.aiSvc.Chat(prompt)
	if err != nil {
		return nil, fmt.Errorf("AI 评分失败：%w", err)
	}

	// 解析评分结果
	evaluation, scores, err := s.parseScoreResponse(response)
	if err != nil {
		return nil, fmt.Errorf("解析评分结果失败：%w", err)
	}

	// 计算总分
	totalScore := (float64(scores.Scores.Clarity+scores.Scores.Completeness+scores.Scores.Structure+scores.Scores.Actionability+scores.Scores.Consistency) / 5) * 10

	// 获取下一个版本号
	version, err := s.repo.GetNextVersion(taskID)
	if err != nil {
		return nil, fmt.Errorf("获取版本号失败：%w", err)
	}

	// 序列化评价内容
	evaluationJSON, err := json.Marshal(evaluation)
	if err != nil {
		return nil, fmt.Errorf("序列化评价内容失败：%w", err)
	}

	// 创建任务快照
	snapshot, err := s.createTaskSnapshot(task)
	if err != nil {
		return nil, fmt.Errorf("创建任务快照失败：%w", err)
	}

	// 获取 AI Provider
	aiProvider := "default"

	// 创建评分记录
	score := &models.TaskQualityScore{
		TaskID:             taskID,
		Version:            version,
		TotalScore:         totalScore,
		ClarityScore:       scores.Scores.Clarity,
		CompletenessScore:  scores.Scores.Completeness,
		StructureScore:     scores.Scores.Structure,
		ActionabilityScore: scores.Scores.Actionability,
		ConsistencyScore:   scores.Scores.Consistency,
		Evaluation:         string(evaluationJSON),
		TaskSnapshot:       snapshot,
		AIProvider:         aiProvider,
	}

	if err := s.repo.Create(score); err != nil {
		return nil, fmt.Errorf("保存评分记录失败：%w", err)
	}

	return score, nil
}

// buildScorePrompt 构建评分 Prompt
func (s *taskQualityService) buildScorePrompt(task *models.Task) string {
	subtasksJSON, _ := json.Marshal(task.Subtasks)

	return fmt.Sprintf(`你是一位资深软件工程专家和提示词工程师，请对以下任务进行质量评分。

【任务信息】
- 标题：%s
- 描述：%s
- 详情：%s
- 模块：%s
- 输入依赖：%s
- 输出交付物：%s
- 风险点：%s
- 验收标准：%s
- 测试策略：%s
- 子任务数量：%d
- 子任务列表：%s

【评分维度】
请从以下 5 个维度进行评分（每项 1-10 分）：

1. 清晰度 (Clarity)：任务描述是否清晰明确，无歧义
   - 标题是否能准确反映任务内容
   - 描述和详情是否容易理解
   - 开发者是否能快速理解任务目标

2. 完整性 (Completeness)：是否包含了必要的上下文和信息
   - 是否有完整的输入/输出定义
   - 是否有明确的风险点
   - 是否有验收标准

3. 结构化 (Structure)：任务结构是否清晰、层次分明
   - 信息组织是否有条理
   - 子任务划分是否合理
   - 依赖关系是否清晰

4. 可执行性 (Actionability)：开发者是否能基于任务描述直接执行
   - 是否有具体的实现细节
   - 是否有代码接口定义
   - 是否有代码实现提示

5. 一致性 (Consistency)：与项目规范和技术栈的兼容性
   - 是否符合项目技术栈（Go + Gin + GORM）
   - 是否符合 API 规范（GET/POST）
   - 命名和风格是否一致

【评分标准】
- 9-10 分：优秀，几乎完美
- 7-8 分：良好，有小幅改进空间
- 5-6 分：一般，需要明显改进
- 3-4 分：较差，存在严重问题
- 1-2 分：很差，几乎无法使用

请以 JSON 格式返回评分结果：
{
  "scores": {
    "clarity": 1-10,
    "completeness": 1-10,
    "structure": 1-10,
    "actionability": 1-10,
    "consistency": 1-10
  },
  "totalScore": 总分 (5 项平均分*10，保留 1 位小数),
  "strengths": ["优点 1", "优点 2", ...],
  "weaknesses": ["缺点 1", "缺点 2", ...],
  "suggestions": [
    {"issue": "问题描述", "suggestion": "改进建议"}
  ],
  "analysis": "详细分析报告（200-500 字）"
}

请只返回 JSON，不要包含其他内容。`,
		task.Title,
		task.Description,
		task.Details,
		strOrEmpty(task.Module),
		strOrEmpty(task.Input),
		strOrEmpty(task.Output),
		strOrEmpty(task.Risk),
		strOrEmpty(task.AcceptanceCriteria),
		task.TestStrategy,
		len(task.Subtasks),
		string(subtasksJSON),
	)
}

// strOrEmpty 返回字符串或空字符串
func strOrEmpty(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// ScoreResult AI 评分结果结构
type ScoreResult struct {
	Scores      struct {
		Clarity       int `json:"clarity"`
		Completeness  int `json:"completeness"`
		Structure     int `json:"structure"`
		Actionability int `json:"actionability"`
		Consistency   int `json:"consistency"`
	} `json:"scores"`
	TotalScore  float64            `json:"totalScore"`
	Strengths   []string           `json:"strengths"`
	Weaknesses  []string           `json:"weaknesses"`
	Suggestions []models.EvalSuggestion `json:"suggestions"`
	Analysis    string             `json:"analysis"`
}

// parseScoreResponse 解析评分响应
func (s *taskQualityService) parseScoreResponse(response string) (*models.EvaluationData, *ScoreResult, error) {
	// 清理响应，移除 Markdown 代码块标记
	cleanedResponse := response
	cleanedResponse = removeMarkdownBlocks(cleanedResponse)

	var result ScoreResult
	if err := json.Unmarshal([]byte(cleanedResponse), &result); err != nil {
		return nil, nil, fmt.Errorf("解析 JSON 失败：%w", err)
	}

	// 构建评价数据
	evaluation := &models.EvaluationData{
		Strengths:   result.Strengths,
		Weaknesses:  result.Weaknesses,
		Suggestions: result.Suggestions,
		Analysis:    result.Analysis,
	}

	return evaluation, &result, nil
}

// removeMarkdownBlocks 移除 Markdown 代码块标记
func removeMarkdownBlocks(s string) string {
	// 简单的实现，移除 ```json 和 ``` 标记
	result := s
	for _, marker := range []string{"```json", "```"} {
		result = replaceAll(result, marker, "")
	}
	return trimWhitespace(result)
}

func replaceAll(s, old, new string) string {
	result := ""
	for {
		idx := indexOf(s, old)
		if idx == -1 {
			break
		}
		result += s[:idx] + new
		s = s[idx+len(old):]
	}
	return result + s
}

func indexOf(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		match := true
		for j := 0; j < len(substr); j++ {
			if s[i+j] != substr[j] {
				match = false
				break
			}
		}
		if match {
			return i
		}
	}
	return -1
}

func trimWhitespace(s string) string {
	start := 0
	end := len(s)
	for start < end && (s[start] == ' ' || s[start] == '\n' || s[start] == '\r' || s[start] == '\t') {
		start++
	}
	for end > start && (s[end-1] == ' ' || s[end-1] == '\n' || s[end-1] == '\r' || s[end-1] == '\t') {
		end--
	}
	return s[start:end]
}

// createTaskSnapshot 创建任务快照
func (s *taskQualityService) createTaskSnapshot(task *models.Task) (string, error) {
	snapshot := map[string]interface{}{
		"id":                 task.ID,
		"title":              task.Title,
		"titleTrans":         task.TitleTrans,
		"description":        task.Description,
		"descriptionTrans":   task.DescriptionTrans,
		"status":             task.Status,
		"priority":           task.Priority,
		"category":           task.Category,
		"details":            task.Details,
		"detailsTrans":       task.DetailsTrans,
		"testStrategy":       task.TestStrategy,
		"testStrategyTrans":  task.TestStrategyTrans,
		"module":             task.Module,
		"input":              task.Input,
		"output":             task.Output,
		"risk":               task.Risk,
		"acceptanceCriteria": task.AcceptanceCriteria,
		"startDate":          task.StartDate,
		"dueDate":            task.DueDate,
		"estimatedHours":     task.EstimatedHours,
		"actualHours":        task.ActualHours,
		"subtasks":           task.Subtasks,
	}
	snapshotJSON, err := json.Marshal(snapshot)
	if err != nil {
		return "", err
	}
	return string(snapshotJSON), nil
}

// GetScoreHistory 获取评分历史
func (s *taskQualityService) GetScoreHistory(taskID uint64, page, pageSize int) ([]models.TaskQualityScore, int64, error) {
	return s.repo.GetByTaskID(taskID, page, pageSize)
}

// GetScoreDetail 获取评分详情
func (s *taskQualityService) GetScoreDetail(scoreID uint64) (*models.TaskQualityScore, error) {
	return s.repo.GetByID(scoreID)
}

// DeleteScore 删除评分记录
func (s *taskQualityService) DeleteScore(scoreID uint64) error {
	return s.repo.Delete(scoreID)
}

// RestoreScore 恢复到评分版本
func (s *taskQualityService) RestoreScore(scoreID uint64) error {
	// 获取评分记录
	score, err := s.repo.GetByID(scoreID)
	if err != nil {
		return fmt.Errorf("获取评分记录失败：%w", err)
	}

	// 解析任务快照
	var snapshot map[string]interface{}
	if err := json.Unmarshal([]byte(score.TaskSnapshot), &snapshot); err != nil {
		return fmt.Errorf("解析任务快照失败：%w", err)
	}

	// 从快照中移除不应该更新的字段
	delete(snapshot, "id")
	delete(snapshot, "createdAt")
	delete(snapshot, "updatedAt")
	delete(snapshot, "deletedAt")
	delete(snapshot, "isExpanding")
	delete(snapshot, "expandMessageId")
	delete(snapshot, "expandStartedAt")

	// 构建更新 map（使用驼峰命名，UpdateWithMap 会自动转换为蛇形）
	updates := map[string]interface{}{
		"title":              snapshot["title"],
		"titleTrans":         snapshot["titleTrans"],
		"description":        snapshot["description"],
		"descriptionTrans":   snapshot["descriptionTrans"],
		"details":            snapshot["details"],
		"detailsTrans":       snapshot["detailsTrans"],
		"testStrategy":       snapshot["testStrategy"],
		"testStrategyTrans":  snapshot["testStrategyTrans"],
		"module":             snapshot["module"],
		"input":              snapshot["input"],
		"output":             snapshot["output"],
		"risk":               snapshot["risk"],
		"acceptanceCriteria": snapshot["acceptanceCriteria"],
		"startDate":          snapshot["startDate"],
		"dueDate":            snapshot["dueDate"],
		"estimatedHours":     snapshot["estimatedHours"],
		"actualHours":        snapshot["actualHours"],
	}

	// 更新任务
	if err := s.taskRepo.UpdateWithMap(score.TaskID, updates); err != nil {
		return fmt.Errorf("更新任务失败：%w", err)
	}

	// TODO: 恢复子任务
	// 如果有子任务快照，也需要恢复

	return nil
}
