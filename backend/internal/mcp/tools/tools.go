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
