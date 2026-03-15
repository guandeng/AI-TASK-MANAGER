package mcp

import (
	"context"
	"encoding/json"
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
		mcp.WithDescription("列出任务，支持按需求过滤"),
		mcp.WithNumber("requirement_id", mcp.Description("需求ID（可选）")),
		mcp.WithString("requirement_name", mcp.Description("需求名称/关键字（可选）")),
		mcp.WithString("status", mcp.Description("任务状态过滤 (pending/in_progress/done/cancelled)")),
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

	// update_task 工具
	s.server.AddTool(mcp.NewTool("update_task",
		mcp.WithDescription("更新任务信息"),
		mcp.WithNumber("task_id", mcp.Required(), mcp.Description("任务 ID")),
		mcp.WithString("title", mcp.Description("任务标题")),
		mcp.WithString("description", mcp.Description("任务描述")),
		mcp.WithString("priority", mcp.Description("优先级")),
		mcp.WithString("due_date", mcp.Description("截止日期")),
		mcp.WithString("start_date", mcp.Description("开始日期")),
		mcp.WithNumber("estimated_hours", mcp.Description("预估工时（小时）")),
	), s.handleUpdateTask)

	// get_task_with_comments 工具
	s.server.AddTool(mcp.NewTool("get_task_with_comments",
		mcp.WithDescription("获取任务及评论"),
		mcp.WithNumber("task_id", mcp.Required(), mcp.Description("任务 ID")),
	), s.handleGetTaskWithComments)

	// validate_dependencies 工具
	s.server.AddTool(mcp.NewTool("validate_dependencies",
		mcp.WithDescription("验证任务依赖关系（检查循环依赖）"),
	), s.handleValidateDependencies)

	// get_ready_tasks 工具
	s.server.AddTool(mcp.NewTool("get_ready_tasks",
		mcp.WithDescription("获取所有依赖已满足的可执行任务"),
	), s.handleGetReadyTasks)

	// search_requirements 工具
	s.server.AddTool(mcp.NewTool("search_requirements",
		mcp.WithDescription("按关键字搜索需求，返回匹配的需求列表供用户选择"),
		mcp.WithString("keyword", mcp.Required(), mcp.Description("搜索关键字（需求标题）")),
	), s.handleSearchRequirements)

	// get_requirement_tasks 工具
	s.server.AddTool(mcp.NewTool("get_requirement_tasks",
		mcp.WithDescription("获取指定需求下的所有任务，如果提供ID则直接查询，如果提供名称则搜索匹配的需求"),
		mcp.WithNumber("requirement_id", mcp.Description("需求ID（可选）")),
		mcp.WithString("requirement_name", mcp.Description("需求名称/关键字（可选）")),
	), s.handleGetRequirementTasks)
}

// 处理函数
func (s *Server) handleListTasks(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	requirementID := request.GetInt("requirement_id", 0)
	requirementName := request.GetString("requirement_name", "")
	statusFilter := request.GetString("status", "")

	// 如果提供了需求名称但没有ID，先搜索需求
	if requirementID == 0 && requirementName != "" {
		var requirements []models.Requirement
		if err := s.db.Where("title LIKE ? AND deleted_at IS NULL", "%"+requirementName+"%").
			Order("created_at DESC").
			Find(&requirements).Error; err != nil {
			return nil, err
		}

		if len(requirements) == 0 {
			return mcp.NewToolResultText(fmt.Sprintf("未找到包含 \"%s\" 的需求", requirementName)), nil
		}

		if len(requirements) > 1 {
			// 多个匹配，返回列表让用户确认
			result := fmt.Sprintf("找到 %d 个匹配的需求，请确认选择哪个：\n", len(requirements))
			for _, req := range requirements {
				result += fmt.Sprintf("  - [ID: %d] %s (状态: %s)\n", req.ID, req.Title, req.Status)
			}
			return mcp.NewToolResultText(result), nil
		}

		// 只有一个匹配，直接使用
		requirementID = int(requirements[0].ID)
	}

	// 构建查询
	query := s.db.Model(&models.Task{}).Where("status != ?", "paused")

	if requirementID > 0 {
		query = query.Where("requirement_id = ?", requirementID)
	}
	if statusFilter != "" {
		query = query.Where("status = ?", statusFilter)
	}

	var tasks []models.Task
	if err := query.Order("created_at DESC").Find(&tasks).Error; err != nil {
		return nil, err
	}

	// 构建结构化结果
	result := make([]map[string]any, len(tasks))
	for i, task := range tasks {
		var reqID any = nil
		if task.RequirementID != nil {
			reqID = *task.RequirementID
		}
		result[i] = map[string]any{
			"id":          task.ID,
			"title":       task.Title,
			"status":      task.Status,
			"priority":    task.Priority,
			"description": task.Description,
			"requirementId": reqID,
		}
	}

	// 返回 JSON 格式
	jsonData, err := json.Marshal(result)
	if err != nil {
		return nil, err
	}
	return mcp.NewToolResultText(string(jsonData)), nil
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
	// 排除暂停状态的任务
	if err := s.db.Where("status = ?", "pending").Where("status != ?", "paused").Order("priority DESC, created_at ASC").First(&task).Error; err != nil {
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

func (s *Server) handleUpdateTask(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	taskID := request.GetInt("task_id", 0)
	if taskID == 0 {
		return nil, fmt.Errorf("invalid task_id")
	}

	updates := make(map[string]interface{})

	if title := request.GetString("title", ""); title != "" {
		updates["title"] = title
	}
	if description := request.GetString("description", ""); description != "" {
		updates["description"] = description
	}
	if priority := request.GetString("priority", ""); priority != "" {
		updates["priority"] = priority
	}
	if dueDate := request.GetString("due_date", ""); dueDate != "" {
		updates["due_date"] = dueDate
	}
	if startDate := request.GetString("start_date", ""); startDate != "" {
		updates["start_date"] = startDate
	}
	if estimatedHours := request.GetFloat("estimated_hours", 0); estimatedHours > 0 {
		updates["estimated_hours"] = estimatedHours
	}

	if err := s.db.Model(&models.Task{}).Where("id = ?", uint64(taskID)).Updates(updates).Error; err != nil {
		return nil, err
	}

	return mcp.NewToolResultText(fmt.Sprintf("Task %d updated", taskID)), nil
}

func (s *Server) handleGetTaskWithComments(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	taskID := request.GetInt("task_id", 0)
	if taskID == 0 {
		return nil, fmt.Errorf("invalid task_id")
	}

	var task models.Task
	if err := s.db.First(&task, uint64(taskID)).Error; err != nil {
		return nil, err
	}

	var comments []models.Comment
	if err := s.db.Where("task_id = ?", uint64(taskID)).Order("created_at DESC").Find(&comments).Error; err != nil {
		return nil, err
	}

	result := fmt.Sprintf("Task: %s (ID: %d, Status: %s)\nComments (%d):\n", task.Title, task.ID, task.Status, len(comments))
	for _, comment := range comments {
		result += fmt.Sprintf("  - [%d] Member%d: %s\n", comment.ID, comment.MemberID, comment.Content)
	}

	return mcp.NewToolResultText(result), nil
}

func (s *Server) handleValidateDependencies(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var dependencies []models.TaskDependency
	if err := s.db.Find(&dependencies).Error; err != nil {
		return nil, err
	}

	// 构建邻接表
	graph := make(map[uint64][]uint64)
	for _, dep := range dependencies {
		graph[dep.TaskID] = append(graph[dep.TaskID], dep.DependsOnTaskID)
	}

	// DFS 检测循环依赖
	visited := make(map[uint64]int) // 0: unvisited, 1: visiting, 2: visited
	var hasCycle bool
	var cyclePath []uint64

	var dfs func(node uint64, path []uint64) bool
	dfs = func(node uint64, path []uint64) bool {
		if visited[node] == 1 {
			hasCycle = true
			cyclePath = append(path, node)
			return true
		}
		if visited[node] == 2 {
			return false
		}

		visited[node] = 1
		path = append(path, node)

		for _, neighbor := range graph[node] {
			if dfs(neighbor, path) {
				return true
			}
		}

		visited[node] = 2
		return false
	}

	for node := range graph {
		if visited[node] == 0 {
			if dfs(node, []uint64{}) {
				break
			}
		}
	}

	if hasCycle {
		return mcp.NewToolResultText(fmt.Sprintf("Circular dependency detected: %v", cyclePath)), nil
	}
	return mcp.NewToolResultText("No circular dependencies found. All dependencies are valid."), nil
}

func (s *Server) handleGetReadyTasks(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// 获取所有 pending 状态的任务，排除暂停状态
	var pendingTasks []models.Task
	if err := s.db.Where("status = ?", "pending").Where("status != ?", "paused").Find(&pendingTasks).Error; err != nil {
		return nil, err
	}

	// 获取所有依赖关系
	var dependencies []models.TaskDependency
	if err := s.db.Find(&dependencies).Error; err != nil {
		return nil, err
	}

	// 构建任务 ID 到状态的映射
	taskStatuses := make(map[uint64]string)
	var allTasks []models.Task
	if err := s.db.Find(&allTasks).Error; err != nil {
		return nil, err
	}
	for _, t := range allTasks {
		taskStatuses[t.ID] = t.Status
	}

	// 过滤出依赖已满足的任务
	var readyTasks []models.Task
	for _, task := range pendingTasks {
		isReady := true

		for _, dep := range dependencies {
			if dep.TaskID == task.ID {
				if status, ok := taskStatuses[dep.DependsOnTaskID]; ok && status != "done" {
					isReady = false
					break
				}
			}
		}

		if isReady {
			readyTasks = append(readyTasks, task)
		}
	}

	result := fmt.Sprintf("Ready tasks (%d):\n", len(readyTasks))
	for _, task := range readyTasks {
		result += fmt.Sprintf("  - [%d] %s (Priority: %s)\n", task.ID, task.Title, task.Priority)
	}

	return mcp.NewToolResultText(result), nil
}

func (s *Server) handleSearchRequirements(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	keyword, err := request.RequireString("keyword")
	if err != nil {
		return nil, err
	}

	var requirements []models.Requirement
	// 搜索标题包含关键字的未删除需求
	if err := s.db.Where("title LIKE ? AND deleted_at IS NULL", "%"+keyword+"%").
		Order("created_at DESC").
		Find(&requirements).Error; err != nil {
		return nil, err
	}

	if len(requirements) == 0 {
		return mcp.NewToolResultText(fmt.Sprintf("未找到包含 \"%s\" 的需求", keyword)), nil
	}

	result := fmt.Sprintf("找到 %d 个匹配的需求，请选择一个：\n", len(requirements))
	for _, req := range requirements {
		status := req.Status
		if status == "" {
			status = "draft"
		}
		result += fmt.Sprintf("  - [ID: %d] %s (状态: %s, 优先级: %s)\n", req.ID, req.Title, status, req.Priority)
	}

	return mcp.NewToolResultText(result), nil
}

func (s *Server) handleGetRequirementTasks(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	requirementID := request.GetInt("requirement_id", 0)
	requirementName := request.GetString("requirement_name", "")

	// 如果没有提供任何参数
	if requirementID == 0 && requirementName == "" {
		return nil, fmt.Errorf("请提供 requirement_id 或 requirement_name")
	}

	// 如果提供的是名称，先搜索需求
	if requirementID == 0 && requirementName != "" {
		var requirements []models.Requirement
		if err := s.db.Where("title LIKE ? AND deleted_at IS NULL", "%"+requirementName+"%").
			Order("created_at DESC").
			Find(&requirements).Error; err != nil {
			return nil, err
		}

		if len(requirements) == 0 {
			return mcp.NewToolResultText(fmt.Sprintf("未找到包含 \"%s\" 的需求", requirementName)), nil
		}

		if len(requirements) > 1 {
			result := fmt.Sprintf("找到 %d 个匹配的需求，请确认选择哪个：\n", len(requirements))
			for _, req := range requirements {
				result += fmt.Sprintf("  - [ID: %d] %s (状态: %s)\n", req.ID, req.Title, req.Status)
			}
			return mcp.NewToolResultText(result), nil
		}

		// 只有一个匹配，直接使用
		requirementID = int(requirements[0].ID)
	}

	// 获取需求信息
	var requirement models.Requirement
	if err := s.db.Where("id = ? AND deleted_at IS NULL", requirementID).First(&requirement).Error; err != nil {
		return mcp.NewToolResultText(fmt.Sprintf("需求 [ID: %d] 不存在或已删除", requirementID)), nil
	}

	// 获取该需求下的所有任务，预加载子任务
	var tasks []models.Task
	if err := s.db.Where("requirement_id = ?", requirementID).
		Preload("Subtasks").
		Order("created_at DESC").
		Find(&tasks).Error; err != nil {
		return nil, err
	}

	// 构建结构化结果
	result := make([]map[string]any, 0, len(tasks))
	for _, task := range tasks {
		taskData := map[string]any{
			"id":          task.ID,
			"title":       task.Title,
			"status":      task.Status,
			"priority":    task.Priority,
			"category":    task.Category,
			"description": task.Description,
			"details":     task.Details,
			"acceptanceCriteria": task.AcceptanceCriteria,
			"input":       task.Input,
			"output":      task.Output,
			"risk":        task.Risk,
			"module":      task.Module,
			"testStrategy": task.TestStrategy,
		}

		// 添加子任务列表
		if len(task.Subtasks) > 0 {
			subtasks := make([]map[string]any, 0, len(task.Subtasks))
			for _, subtask := range task.Subtasks {
				subtaskData := map[string]any{
					"id":               subtask.ID,
					"title":            subtask.Title,
					"status":           subtask.Status,
					"priority":         subtask.Priority,
					"description":      subtask.Description,
					"details":          subtask.Details,
					"codeInterface":    subtask.CodeInterface,
					"acceptanceCriteria": subtask.AcceptanceCriteria,
					"relatedFiles":     subtask.RelatedFiles,
					"codeHints":        subtask.CodeHints,
				}
				subtasks = append(subtasks, subtaskData)
			}
			taskData["subtasks"] = subtasks
		}

		result = append(result, taskData)
	}

	// 返回 JSON 格式
	jsonData, err := json.Marshal(result)
	if err != nil {
		return nil, err
	}
	return mcp.NewToolResultText(string(jsonData)), nil
}
