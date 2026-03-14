package services

import (
	"time"

	"github.com/ai-task-manager/backend/internal/models"
	"github.com/ai-task-manager/backend/internal/repository"
)

// TaskService 任务服务接口
type TaskService interface {
	// 任务 CRUD
	List(filters map[string]interface{}, page, pageSize int) ([]models.Task, int64, error)
	GetByID(id uint64) (*models.Task, error)
	Create(task *models.Task) error
	Update(id uint64, updates map[string]interface{}) error
	Delete(id uint64) error
	BatchDelete(ids []uint64) error

	// 子任务
	UpdateSubtask(taskID, subtaskID uint64, updates map[string]interface{}) error
	DeleteSubtask(taskID, subtaskID uint64) error
	DeleteAllSubtasks(taskID uint64) error
	ReorderSubtasks(taskID uint64, subtaskIDs []uint64) error
	RegenerateSubtask(taskID, subtaskID uint64) error

	// AI 功能
	ExpandTask(taskID uint64, sync bool) error

	// 时间更新
	UpdateTime(id uint64, startTime, dueDate *time.Time, estimatedHours, actualHours *float64) error
}

// taskService 任务服务实现
type taskService struct {
	repo   repository.TaskRepository
	aiSvc  AIService
}

// NewTaskService 创建任务服务
func NewTaskService(aiSvc AIService) TaskService {
	return &taskService{
		repo:  repository.NewTaskRepository(),
		aiSvc: aiSvc,
	}
}

// List 获取任务列表
func (s *taskService) List(filters map[string]interface{}, page, pageSize int) ([]models.Task, int64, error) {
	return s.repo.List(filters, page, pageSize)
}

// GetByID 根据 ID 获取任务
func (s *taskService) GetByID(id uint64) (*models.Task, error) {
	return s.repo.GetByID(id)
}

// Create 创建任务
func (s *taskService) Create(task *models.Task) error {
	return s.repo.Create(task)
}

// Update 更新任务
func (s *taskService) Update(id uint64, updates map[string]interface{}) error {
	task, err := s.repo.GetByID(id)
	if err != nil {
		return err
	}

	// 应用更新
	if title, ok := updates["title"].(string); ok {
		task.Title = title
	}
	if description, ok := updates["description"].(string); ok {
		task.Description = description
	}
	if status, ok := updates["status"].(string); ok {
		task.Status = status
		if status == "done" || status == "completed" {
			now := time.Now()
			task.CompletedAt = &now
		}
	}
	if priority, ok := updates["priority"].(string); ok {
		task.Priority = priority
	}
	if details, ok := updates["details"].(string); ok {
		task.Details = details
	}
	if testStrategy, ok := updates["testStrategy"].(string); ok {
		task.TestStrategy = testStrategy
	}
	if assignee, ok := updates["assignee"].(string); ok {
		task.Assignee = &assignee
	}

	return s.repo.Update(task)
}

// Delete 删除任务
func (s *taskService) Delete(id uint64) error {
	return s.repo.Delete(id)
}

// BatchDelete 批量删除任务
func (s *taskService) BatchDelete(ids []uint64) error {
	return s.repo.BatchDelete(ids)
}

// UpdateSubtask 更新子任务
func (s *taskService) UpdateSubtask(taskID, subtaskID uint64, updates map[string]interface{}) error {
	subtask := &models.Subtask{
		ID:     subtaskID,
		TaskID: taskID,
	}

	if title, ok := updates["title"].(string); ok {
		subtask.Title = title
	}
	if description, ok := updates["description"].(string); ok {
		subtask.Description = description
	}
	if status, ok := updates["status"].(string); ok {
		subtask.Status = status
	}
	if details, ok := updates["details"].(string); ok {
		subtask.Details = details
	}

	return s.repo.UpdateSubtask(subtask)
}

// DeleteSubtask 删除子任务
func (s *taskService) DeleteSubtask(taskID, subtaskID uint64) error {
	return s.repo.DeleteSubtask(taskID, subtaskID)
}

// DeleteAllSubtasks 删除任务的所有子任务
func (s *taskService) DeleteAllSubtasks(taskID uint64) error {
	return s.repo.DeleteAllSubtasks(taskID)
}

// ReorderSubtasks 重新排序子任务
func (s *taskService) ReorderSubtasks(taskID uint64, subtaskIDs []uint64) error {
	return s.repo.ReorderSubtasks(taskID, subtaskIDs)
}

// RegenerateSubtask 重新生成子任务（AI）
func (s *taskService) RegenerateSubtask(taskID, subtaskID uint64) error {
	// 获取任务
	task, err := s.repo.GetByID(taskID)
	if err != nil {
		return err
	}

	// 删除旧子任务
	if err := s.repo.DeleteSubtask(taskID, subtaskID); err != nil {
		return err
	}

	// 使用 AI 生成新子任务
	if s.aiSvc != nil {
		newSubtask, err := s.aiSvc.GenerateSubtask(task)
		if err != nil {
			return err
		}
		return s.repo.CreateSubtask(newSubtask)
	}

	return nil
}

// ExpandTask 展开任务（AI 生成子任务）
func (s *taskService) ExpandTask(taskID uint64, sync bool) error {
	task, err := s.repo.GetByID(taskID)
	if err != nil {
		return err
	}

	if s.aiSvc == nil {
		return nil
	}

	// 使用 AI 生成子任务
	subtasks, err := s.aiSvc.ExpandTask(task)
	if err != nil {
		return err
	}

	// 保存子任务
	for _, subtask := range subtasks {
		subtask.TaskID = taskID
		if err := s.repo.CreateSubtask(&subtask); err != nil {
			return err
		}
	}

	return nil
}

// UpdateTime 更新任务时间
func (s *taskService) UpdateTime(id uint64, startTime, dueDate *time.Time, estimatedHours, actualHours *float64) error {
	task, err := s.repo.GetByID(id)
	if err != nil {
		return err
	}

	if startTime != nil {
		task.StartDate = startTime
	}
	if dueDate != nil {
		task.DueDate = dueDate
	}
	if estimatedHours != nil {
		task.EstimatedHours = estimatedHours
	}
	if actualHours != nil {
		task.ActualHours = actualHours
	}

	return s.repo.Update(task)
}

// AIService AI 服务接口
type AIService interface {
	ExpandTask(task *models.Task) ([]models.Subtask, error)
	GenerateSubtask(task *models.Task) (*models.Subtask, error)
}
