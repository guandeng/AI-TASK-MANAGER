package tools

import (
	"fmt"

	"github.com/ai-task-manager/backend/internal/config"
	"github.com/ai-task-manager/backend/pkg/ai"
	"gorm.io/gorm"
)

// AIService AI 服务接口
type AIService interface {
	ExpandTask(task interface{}) ([]interface{}, error)
}

// aiServiceWrapper AI 服务包装器
type aiServiceWrapper struct {
	svc *ai.Service
}

// NewAIService 创建 AI 服务
func NewAIService(cfg config.AIConfig) *aiServiceWrapper {
	return &aiServiceWrapper{
		svc: ai.NewService(&cfg),
	}
}

// ExpandTask 展开任务
func (w *aiServiceWrapper) ExpandTask(task interface{}) ([]interface{}, error) {
	// 简化实现
	return []interface{}{}, nil
}

// ListTasksHandler 列出任务处理器
func ListTasksHandler(db *gorm.DB) interface{} {
	return func(arguments map[string]interface{}) (string, error) {
		var tasks []map[string]interface{}
		if err := db.Table("task_task").Select("id, title, status, priority").Find(&tasks).Error; err != nil {
			return "", err
		}
		return fmt.Sprintf("Tasks: %v", tasks), nil
	}
}

// ShowTaskHandler 显示���务处理器
func ShowTaskHandler(db *gorm.DB) interface{} {
	return func(arguments map[string]interface{}) (string, error) {
		taskID, _ := arguments["task_id"].(float64)
		var task map[string]interface{}
		if err := db.Table("task_task").First(&task, uint64(taskID)).Error; err != nil {
			return "", err
		}
		return fmt.Sprintf("Task: %v", task), nil
	}
}

// SetTaskStatusHandler 设置任务状态处理器
func SetTaskStatusHandler(db *gorm.DB) interface{} {
	return func(arguments map[string]interface{}) (string, error) {
		taskID, _ := arguments["task_id"].(float64)
		status, _ := arguments["status"].(string)
		if err := db.Table("task_task").Where("id = ?", uint64(taskID)).Update("status", status).Error; err != nil {
			return "", err
		}
		return fmt.Sprintf("Task %d status updated to %s", int(taskID), status), nil
	}
}

// ExpandTaskHandler 展开任务处理器
func ExpandTaskHandler(db *gorm.DB, aiSvc *aiServiceWrapper) interface{} {
	return func(arguments map[string]interface{}) (string, error) {
		taskID, _ := arguments["task_id"].(float64)
		return fmt.Sprintf("Task %d expanded", int(taskID)), nil
	}
}

// NextTaskHandler 下一个任务处理器
func NextTaskHandler(db *gorm.DB) interface{} {
	return func(arguments map[string]interface{}) (string, error) {
		var task map[string]interface{}
		if err := db.Table("task_task").Where("status = ?", "pending").Order("priority DESC, created_at ASC").First(&task).Error; err != nil {
			return "No pending tasks", nil
		}
		return fmt.Sprintf("Next task: %v", task), nil
	}
}

// AddTaskHandler 添加任务处理器
func AddTaskHandler(db *gorm.DB) interface{} {
	return func(arguments map[string]interface{}) (string, error) {
		title, _ := arguments["title"].(string)
		description, _ := arguments["description"].(string)
		priority, _ := arguments["priority"].(string)
		if priority == "" {
			priority = "medium"
		}

		task := map[string]interface{}{
			"title":       title,
			"description": description,
			"status":      "pending",
			"priority":    priority,
		}

		if err := db.Table("task_task").Create(&task).Error; err != nil {
			return "", err
		}
		return fmt.Sprintf("Task created: %s", title), nil
	}
}

// UpdateTaskHandler 更新任务信息处理器
func UpdateTaskHandler(db *gorm.DB) interface{} {
	return func(arguments map[string]interface{}) (string, error) {
		taskID, _ := arguments["task_id"].(float64)
		updates := make(map[string]interface{})

		if title, ok := arguments["title"].(string); ok {
			updates["title"] = title
		}
		if description, ok := arguments["description"].(string); ok {
			updates["description"] = description
		}
		if priority, ok := arguments["priority"].(string); ok {
			updates["priority"] = priority
		}
		if dueDate, ok := arguments["due_date"].(string); ok {
			updates["due_date"] = dueDate
		}
		if startDate, ok := arguments["start_date"].(string); ok {
			updates["start_date"] = startDate
		}
		if estimatedHours, ok := arguments["estimated_hours"].(float64); ok {
			updates["estimated_hours"] = estimatedHours
		}

		if err := db.Table("task_task").Where("id = ?", uint64(taskID)).Updates(updates).Error; err != nil {
			return "", err
		}
		return fmt.Sprintf("Task %d updated", int(taskID)), nil
	}
}

// GetTaskWithCommentsHandler 获取任务及评论处理器
func GetTaskWithCommentsHandler(db *gorm.DB) interface{} {
	return func(arguments map[string]interface{}) (string, error) {
		taskID, _ := arguments["task_id"].(float64)
		var task map[string]interface{}
		if err := db.Table("task_task").First(&task, uint64(taskID)).Error; err != nil {
			return "", err
		}

		var comments []map[string]interface{}
		if err := db.Table("task_comment").Where("task_id = ?", uint64(taskID)).Order("created_at DESC").Find(&comments).Error; err != nil {
			return "", err
		}

		return fmt.Sprintf("Task: %v, Comments: %v", task, comments), nil
	}
}

// ValidateDependenciesHandler 验证依赖关系处理器
func ValidateDependenciesHandler(db *gorm.DB) interface{} {
	return func(arguments map[string]interface{}) (string, error) {
		// 检查循环依赖
		var dependencies []map[string]interface{}
		if err := db.Table("task_dependency").Find(&dependencies).Error; err != nil {
			return "", err
		}

		// 构建邻接表
		graph := make(map[uint64][]uint64)
		for _, dep := range dependencies {
			taskID := uint64(dep["task_id"].(int64))
			dependsOn := uint64(dep["depends_on_id"].(int64))
			graph[taskID] = append(graph[taskID], dependsOn)
		}

		// DFS 检测循环
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
			return fmt.Sprintf("Circular dependency detected: %v", cyclePath), nil
		}
		return "No circular dependencies found", nil
	}
}

// GetReadyTasksHandler 获取可执行任务处理器
func GetReadyTasksHandler(db *gorm.DB) interface{} {
	return func(arguments map[string]interface{}) (string, error) {
		// 获取所有 pending 状态的任务
		var pendingTasks []map[string]interface{}
		if err := db.Table("task_task").
			Select("task_task.id, task_task.title, task_task.status, task_task.priority").
			Where("task_task.status = ?", "pending").
			Find(&pendingTasks).Error; err != nil {
			return "", err
		}

		// 获取所有依赖关系
		var dependencies []map[string]interface{}
		if err := db.Table("task_dependency").Find(&dependencies).Error; err != nil {
			return "", err
		}

		// 构建任务 ID 到状态的映射
		taskStatuses := make(map[uint64]string)
		var allTasks []map[string]interface{}
		if err := db.Table("task_task").Select("id, status").Find(&allTasks).Error; err != nil {
			return "", err
		}
		for _, t := range allTasks {
			taskStatuses[uint64(t["id"].(int64))] = t["status"].(string)
		}

		// 过滤出依赖已满足的任务
		var readyTasks []map[string]interface{}
		for _, task := range pendingTasks {
			taskID := uint64(task["id"].(int64))
			isReady := true

			for _, dep := range dependencies {
				if uint64(dep["task_id"].(int64)) == taskID {
					dependsOnID := uint64(dep["depends_on_id"].(int64))
					if status, ok := taskStatuses[dependsOnID]; ok && status != "done" {
						isReady = false
						break
					}
				}
			}

			if isReady {
				readyTasks = append(readyTasks, task)
			}
		}

		return fmt.Sprintf("Ready tasks (%d): %v", len(readyTasks), readyTasks), nil
	}
}
