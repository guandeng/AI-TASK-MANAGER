package mcp

import (
	"context"
	"fmt"

	"github.com/ai-task-manager/backend/internal/config"
	"github.com/ai-task-manager/backend/internal/database"
	"github.com/ai-task-manager/backend/internal/models"
	"github.com/ai-task-manager/backend/pkg/ai"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// Server MCP 服务器
type Server struct {
	server *server.MCPServer
	cfg    *config.Config
	logger *zap.Logger
	db     *gorm.DB
	aiSvc  *ai.Service
}

// NewServer 创建 MCP 服务器
func NewServer(cfg *config.Config, logger *zap.Logger) *Server {
	mcpServer := server.NewMCPServer(
		"AI Task Manager MCP Server",
		"1.0.0",
	)

	return &Server{
		server: mcpServer,
		cfg:    cfg,
		logger: logger,
		db:     database.GetDB(),
		aiSvc:  ai.NewService(&cfg.AI),
	}
}

// Start 启动 MCP 服务器
func (s *Server) Start(ctx context.Context) error {
	// 注册工具
	s.registerTools()
	// 启动 stdio 服务器
	return server.ServeStdio(s.server)
}

// registerTools 注册所有工具
func (s *Server) registerTools() {
	// list_tasks 工具
	s.server.AddTool(mcp.NewTool("list_tasks",
		mcp.WithDescription("列出所有任务"),
	), s.handleListTasks)

	// show_task 工具
	s.server.AddTool(mcp.NewTool("show_task",
		mcp.WithDescription("显示任务详情"),
		mcp.WithNumber("task_id", mcp.Required(), mcp.Description("任务ID")),
	), s.handleShowTask)

	// set_task_status 工具
	s.server.AddTool(mcp.NewTool("set_task_status",
		mcp.WithDescription("设置任务状态"),
		mcp.WithNumber("task_id", mcp.Required(), mcp.Description("任务ID")),
		mcp.WithString("status", mcp.Required(), mcp.Description("任务状态 (pending/in_progress/done/cancelled)")),
	), s.handleSetTaskStatus)

	// expand_task 工具
	s.server.AddTool(mcp.NewTool("expand_task",
		mcp.WithDescription("展开任务（AI 生成子任务）"),
		mcp.WithNumber("task_id", mcp.Required(), mcp.Description("任务ID")),
	), s.handleExpandTask)

	// next_task 工具
	s.server.AddTool(mcp.NewTool("next_task",
		mcp.WithDescription("获取下一个待办任务"),
	), s.handleNextTask)

	// add_task 工具
	s.server.AddTool(mcp.NewTool("add_task",
		mcp.WithDescription("添加新任务"),
		mcp.WithString("title", mcp.Required(), mcp.Description("任务标题")),
		mcp.WithString("description", mcp.Description("任务描述")),
		mcp.WithString("priority", mcp.Description("优先级 (high/medium/low)")),
	), s.handleAddTask)
}

// 处理函数
func (s *Server) handleListTasks(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var tasks []models.Task
	if err := s.db.Order("created_at DESC").Find(&tasks).Error; err != nil {
		return nil, err
	}

	result := make([]map[string]interface{}, len(tasks))
	for i, task := range tasks {
		result[i] = map[string]interface{}{
			"id":       task.ID,
			"title":    task.Title,
			"status":   task.Status,
			"priority": task.Priority,
		}
	}

	return mcp.NewToolResultText(fmt.Sprintf("%v", result)), nil
}

func (s *Server) handleShowTask(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	taskID := request.GetInt("task_id", 0)
	if taskID == 0 {
		return nil, fmt.Errorf("invalid task_id")
	}

	var task models.Task
	if err := s.db.First(&task, uint64(taskID)).Error; err != nil {
		return nil, err
	}

	return mcp.NewToolResultText(fmt.Sprintf("Task: %s\nStatus: %s\nPriority: %s\nDescription: %s",
		task.Title, task.Status, task.Priority, task.Description)), nil
}

func (s *Server) handleSetTaskStatus(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	taskID := request.GetInt("task_id", 0)
	if taskID == 0 {
		return nil, fmt.Errorf("invalid task_id")
	}

	status, err := request.RequireString("status")
	if err != nil {
		return nil, err
	}

	if err := s.db.Model(&models.Task{}).Where("id = ?", uint64(taskID)).Update("status", status).Error; err != nil {
		return nil, err
	}

	return mcp.NewToolResultText(fmt.Sprintf("Task %d status updated to %s", taskID, status)), nil
}

func (s *Server) handleExpandTask(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	taskID := request.GetInt("task_id", 0)
	if taskID == 0 {
		return nil, fmt.Errorf("invalid task_id")
	}

	var task models.Task
	if err := s.db.First(&task, uint64(taskID)).Error; err != nil {
		return nil, err
	}

	subtasks, err := s.aiSvc.ExpandTask(&task)
	if err != nil {
		return nil, err
	}

	for i := range subtasks {
		subtasks[i].TaskID = task.ID
		if err := s.db.Create(&subtasks[i]).Error; err != nil {
			return nil, err
		}
	}

	return mcp.NewToolResultText(fmt.Sprintf("Task expanded into %d subtasks", len(subtasks))), nil
}

func (s *Server) handleNextTask(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var task models.Task
	if err := s.db.Where("status = ?", "pending").Order("priority DESC, created_at ASC").First(&task).Error; err != nil {
		return mcp.NewToolResultText("No pending tasks found"), nil
	}

	return mcp.NewToolResultText(fmt.Sprintf("Next task: %s (ID: %d, Priority: %s)",
		task.Title, task.ID, task.Priority)), nil
}

func (s *Server) handleAddTask(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	title, err := request.RequireString("title")
	if err != nil {
		return nil, err
	}

	description := request.GetString("description", "")
	priority := request.GetString("priority", "medium")

	task := models.Task{
		Title:       title,
		Description: description,
		Status:      "pending",
		Priority:    priority,
	}

	if err := s.db.Create(&task).Error; err != nil {
		return nil, err
	}

	return mcp.NewToolResultText(fmt.Sprintf("Task created: %s (ID: %d)", task.Title, task.ID)), nil
}
