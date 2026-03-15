package mcp

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/ai-task-manager/backend/internal/config"
	"github.com/ai-task-manager/backend/internal/database"
	"github.com/ai-task-manager/backend/internal/models"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func setupTestServer(t *testing.T) (*Server, sqlmock.Sqlmock, func()) {
	t.Helper()

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create sqlmock: %v", err)
	}

	gormDB, err := gorm.Open(mysql.New(mysql.Config{
		Conn:                      db,
		SkipInitializeWithVersion: true,
	}), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to open gorm: %v", err)
	}

	// 设置全局 DB
	database.SetTestDB(gormDB)

	logger, _ := zap.NewDevelopment()
	cfg := &config.Config{}

	cleanup := func() {
		db.Close()
	}

	return NewServer(cfg, logger), mock, cleanup
}

func TestHandleGetRequirementTasks_JSON(t *testing.T) {
	s, mock, cleanup := setupTestServer(t)
	defer cleanup()

	// 模拟需求查询
	mock.ExpectQuery("SELECT \\* FROM `task_requirement` WHERE id = \\? AND deleted_at IS NULL ORDER BY `task_requirement`.`id` LIMIT \\?").
		WithArgs(1, 1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "title", "status"}).
			AddRow(1, "测试需求", "active"))

	// 模拟任务查询
	mock.ExpectQuery("SELECT \\* FROM `task_task` WHERE requirement_id = \\? AND deleted_at IS NULL ORDER BY created_at DESC").
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "requirement_id", "title", "status", "priority", "category", "description", "details", "acceptance_criteria", "input", "output", "risk", "module", "test_strategy"}).
			AddRow(1, 1, "任务 1", "pending", "high", "backend", "描述 1", "详情 1", nil, nil, nil, nil, nil, ""))

	// 模拟子任务查询
	mock.ExpectQuery("SELECT \\* FROM `task_subtask` WHERE `task_subtask`.`task_id` = \\?").
		WillReturnRows(sqlmock.NewRows([]string{"id", "task_id", "title", "status", "priority", "description", "details", "code_interface", "acceptance_criteria", "related_files", "code_hints"}).
			AddRow(1, 1, "子任务 1", "pending", "medium", "子描述", "子详情", nil, nil, nil, nil))

	req := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Arguments: map[string]interface{}{
				"requirement_id": 1,
				"format":         "json",
			},
		},
	}

	result, err := s.handleGetRequirementTasks(context.Background(), req)
	assert.NoError(t, err)
	assert.NotNil(t, result)

	// 解析 JSON 结果
	textContent, ok := mcp.AsTextContent(result.Content[0])
	assert.True(t, ok)
	var tasks []map[string]interface{}
	err = json.Unmarshal([]byte(textContent.Text), &tasks)
	assert.NoError(t, err)
	assert.Len(t, tasks, 1)

	// 验证任务字段
	task := tasks[0]
	assert.Equal(t, float64(1), task["id"])
	assert.Equal(t, "任务 1", task["title"])
	assert.Equal(t, "pending", task["status"])
	assert.Equal(t, "high", task["priority"])
	assert.Equal(t, "backend", task["category"])
	assert.Equal(t, "描述 1", task["description"])
	assert.Equal(t, "详情 1", task["details"])
}

func TestHandleGetRequirementTasks_Markdown(t *testing.T) {
	s, mock, cleanup := setupTestServer(t)
	defer cleanup()

	// 模拟需求查询
	mock.ExpectQuery("SELECT \\* FROM `task_requirement` WHERE id = \\? AND deleted_at IS NULL ORDER BY `task_requirement`.`id` LIMIT \\?").
		WithArgs(1, 1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "title", "status"}).
			AddRow(1, "测试需求", "active"))

	// 模拟任务查询
	mock.ExpectQuery("SELECT \\* FROM `task_task` WHERE requirement_id = \\? AND deleted_at IS NULL ORDER BY created_at DESC").
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "requirement_id", "title", "status", "priority", "category", "description", "details", "acceptance_criteria", "input", "output", "risk", "module", "test_strategy"}).
			AddRow(1, 1, "任务 1", "pending", "high", "backend", "描述 1", "详情 1", nil, nil, nil, nil, nil, ""))

	// 模拟子任务查询
	mock.ExpectQuery("SELECT \\* FROM `task_subtask` WHERE `task_subtask`.`task_id` = \\?").
		WillReturnRows(sqlmock.NewRows([]string{"id", "task_id", "title", "status", "priority", "description", "details", "code_interface", "acceptance_criteria", "related_files", "code_hints"}).
			AddRow(1, 1, "子任务 1", "pending", "medium", "子描述", "子详情", nil, nil, nil, nil))

	req := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Arguments: map[string]interface{}{
				"requirement_id": 1,
				"format":         "markdown",
			},
		},
	}

	result, err := s.handleGetRequirementTasks(context.Background(), req)
	assert.NoError(t, err)
	assert.NotNil(t, result)

	textContent, ok := mcp.AsTextContent(result.Content[0])
	assert.True(t, ok)
	output := textContent.Text

	// 验证 Markdown 格式
	assert.Contains(t, output, "**任务总数**: 1")
	assert.Contains(t, output, "## 任务列表")
	assert.Contains(t, output, "| 1 | 任务 1 | pending | 高 | backend |")
	assert.Contains(t, output, "| 状态 | pending |")
	assert.Contains(t, output, "| 优先级 | 高 |")
	assert.Contains(t, output, "**描述**:")
	assert.Contains(t, output, "**子任务**:")
	assert.Contains(t, output, "**1.1 子任务 1** [ID: 1]")
}

func TestHandleGetRequirementTasks_RequireParams(t *testing.T) {
	s, _, cleanup := setupTestServer(t)
	defer cleanup()

	req := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Arguments: map[string]interface{}{},
		},
	}

	_, err := s.handleGetRequirementTasks(context.Background(), req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "请提供 requirement_id 或 requirement_name")
}

func TestHandleGetRequirementTasks_RequirementNotFound(t *testing.T) {
	s, mock, cleanup := setupTestServer(t)
	defer cleanup()

	// 模拟需求不存在
	mock.ExpectQuery("SELECT \\* FROM `task_requirement` WHERE id = \\? AND deleted_at IS NULL ORDER BY `task_requirement`.`id` LIMIT \\?").
		WithArgs(999, 1).
		WillReturnError(gorm.ErrRecordNotFound)

	req := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Arguments: map[string]interface{}{
				"requirement_id": 999,
			},
		},
	}

	result, err := s.handleGetRequirementTasks(context.Background(), req)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	textContent, ok := mcp.AsTextContent(result.Content[0])
	assert.True(t, ok)
	assert.Contains(t, textContent.Text, "需求 [ID: 999] 不存在或已删除")
}

func TestBuildRequirementTasksMarkdown(t *testing.T) {
	s, _, cleanup := setupTestServer(t)
	defer cleanup()

	requirement := &models.Requirement{
		ID:     1,
		Title:  "测试需求",
		Status: "active",
	}

	tasks := []models.Task{
		{
			ID:          1,
			Title:       "任务 1",
			Status:      "pending",
			Priority:    "high",
			Category:    "backend",
			Description: "任务描述",
			Details:     "任务详情",
			Subtasks: []models.Subtask{
				{
					ID:          1,
					Title:       "子任务 1",
					Status:      "done",
					Priority:    "medium",
					Description: "子任务描述",
					Details:     "子任务详情",
				},
			},
		},
	}

	result := s.buildRequirementTasksMarkdown(requirement, tasks)
	assert.NotNil(t, result)

	textContent, ok := mcp.AsTextContent(result.Content[0])
	assert.True(t, ok)
	output := textContent.Text

	// 验证层级结构
	assert.Contains(t, output, "# 测试需求 [ID: 1]")
	assert.Contains(t, output, "**状态**: active")
	assert.Contains(t, output, "**任务总数**: 1")
	assert.Contains(t, output, "## 任务列表")
	assert.Contains(t, output, "| 1 | 任务 1 | pending | 高 | backend |")
	assert.Contains(t, output, "| 状态 | pending |")
	assert.Contains(t, output, "| 优先级 | 高 |")
	assert.Contains(t, output, "**描述**:")
	assert.Contains(t, output, "任务描述")
	assert.Contains(t, output, "**详情**:")
	assert.Contains(t, output, "任务详情")
	assert.Contains(t, output, "**子任务**:")
	assert.Contains(t, output, "**1.1 子任务 1** [ID: 1]")
	assert.Contains(t, output, "done")
}

func TestBuildRequirementTasksMarkdown_WithSubtask(t *testing.T) {
	s, _, cleanup := setupTestServer(t)
	defer cleanup()

	requirement := &models.Requirement{
		ID:       1,
		Title:    "API 开发",
		Status:   "active",
		Priority: "high",
	}

	tasks := []models.Task{
		{
			ID:       1,
			Title:    "用户接口",
			Status:   "pending",
			Priority: "high",
			Subtasks: []models.Subtask{
				{
					ID:          1,
					Title:       "获取用户信息",
					Status:      "pending",
					Priority:    "medium",
					Description: "实现获取用户信息的 API 接口",
				},
			},
		},
	}

	result := s.buildRequirementTasksMarkdown(requirement, tasks)
	textContent, ok := mcp.AsTextContent(result.Content[0])
	assert.True(t, ok)
	output := textContent.Text

	// 验证需求头部
	assert.Contains(t, output, "# API 开发 [ID: 1]")
	assert.Contains(t, output, "**状态**: active")
	assert.Contains(t, output, "**优先级**: 高")

	// 验证任务列表
	assert.Contains(t, output, "## 任务列表")
	assert.Contains(t, output, "| 用户接口 | pending | 高 |")

	// 验证任务详情
	assert.Contains(t, output, "## 任务详情")
	assert.Contains(t, output, "### 1. 用户接口 [ID: 1]")

	// 验证子任务
	assert.Contains(t, output, "**子任务**:")
	assert.Contains(t, output, "**1.1 获取用户信息** [ID: 1]")
	assert.Contains(t, output, "实现获取用户信息的 API 接口")
}

func TestBuildRequirementTasksMarkdown_MultipleTasks(t *testing.T) {
	s, _, cleanup := setupTestServer(t)
	defer cleanup()

	requirement := &models.Requirement{
		ID:     1,
		Title:  "多任务需求",
		Status: "active",
	}

	tasks := []models.Task{
		{
			ID:       1,
			Title:    "任务一",
			Status:   "pending",
			Priority: "high",
		},
		{
			ID:       2,
			Title:    "任务二",
			Status:   "in_progress",
			Priority: "medium",
		},
		{
			ID:       3,
			Title:    "任务三",
			Status:   "done",
			Priority: "low",
		},
	}

	result := s.buildRequirementTasksMarkdown(requirement, tasks)
	textContent, ok := mcp.AsTextContent(result.Content[0])
	assert.True(t, ok)
	output := textContent.Text

	assert.Contains(t, output, "# 多任务需求 [ID: 1]")
	assert.Contains(t, output, "**任务总数**: 3")
	assert.Contains(t, output, "## 1. 任务一 [ID: 1]")
	assert.Contains(t, output, "## 2. 任务二 [ID: 2]")
	assert.Contains(t, output, "## 3. 任务三 [ID: 3]")
	assert.Contains(t, output, "---")
}
