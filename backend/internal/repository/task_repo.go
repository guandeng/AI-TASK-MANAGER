package repository

import (
	"github.com/ai-task-manager/backend/internal/database"
	"github.com/ai-task-manager/backend/internal/models"
	"gorm.io/gorm"
)

// TaskRepository 任务仓储接口
type TaskRepository interface {
	// 任务 CRUD
	List(filters map[string]interface{}, page, pageSize int) ([]models.Task, int64, error)
	GetByID(id uint64) (*models.Task, error)
	Create(task *models.Task) error
	Update(task *models.Task) error
	Delete(id uint64) error
	BatchDelete(ids []uint64) error

	// 子任务
	GetSubtasks(taskID uint64) ([]models.Subtask, error)
	CreateSubtask(subtask *models.Subtask) error
	UpdateSubtask(subtask *models.Subtask) error
	DeleteSubtask(taskID, subtaskID uint64) error
	DeleteAllSubtasks(taskID uint64) error
	ReorderSubtasks(taskID uint64, subtaskIDs []uint64) error

	// 依赖关系
	GetDependencies(taskID uint64) ([]models.TaskDependency, error)
}

// taskRepository 任务仓储实现
type taskRepository struct {
	db *gorm.DB
}

// NewTaskRepository 创建任务仓储
func NewTaskRepository() TaskRepository {
	return &taskRepository{db: database.GetDB()}
}

// List 获取任务列表
func (r *taskRepository) List(filters map[string]interface{}, page, pageSize int) ([]models.Task, int64, error) {
	var tasks []models.Task
	var total int64

	// 构建基础查询
	baseQuery := r.db.Model(&models.Task{})

	// 应用筛选条件到基础查询
	if status, ok := filters["status"]; ok {
		baseQuery = baseQuery.Where("status = ?", status)
	}
	if priority, ok := filters["priority"]; ok {
		baseQuery = baseQuery.Where("priority = ?", priority)
	}
	if requirementID, ok := filters["requirementId"]; ok {
		baseQuery = baseQuery.Where("requirement_id = ?", requirementID)
	}
	if assignee, ok := filters["assignee"]; ok {
		baseQuery = baseQuery.Where("assignee = ?", assignee)
	}
	if keyword, ok := filters["keyword"]; ok {
		baseQuery = baseQuery.Where("title LIKE ? OR description LIKE ?", "%"+keyword.(string)+"%", "%"+keyword.(string)+"%")
	}

	// 计算总数
	if err := baseQuery.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询，使用 LEFT JOIN 获取需求标题
	offset := (page - 1) * pageSize
	query := r.db.Table("task_task").
		Select("task_task.*, task_requirement.title as requirement_title").
		Joins("LEFT JOIN task_requirement ON task_task.requirement_id = task_requirement.id")

	// 应用相同的筛选条件
	if status, ok := filters["status"]; ok {
		query = query.Where("task_task.status = ?", status)
	}
	if priority, ok := filters["priority"]; ok {
		query = query.Where("task_task.priority = ?", priority)
	}
	if requirementID, ok := filters["requirementId"]; ok {
		query = query.Where("task_task.requirement_id = ?", requirementID)
	}
	if assignee, ok := filters["assignee"]; ok {
		query = query.Where("task_task.assignee = ?", assignee)
	}
	if keyword, ok := filters["keyword"]; ok {
		query = query.Where("task_task.title LIKE ? OR task_task.description LIKE ?", "%"+keyword.(string)+"%", "%"+keyword.(string)+"%")
	}

	if err := query.Order("task_task.created_at DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&tasks).Error; err != nil {
		return nil, 0, err
	}

	return tasks, total, nil
}

// GetByID 根据 ID 获取任务
func (r *taskRepository) GetByID(id uint64) (*models.Task, error) {
	var task models.Task
	if err := r.db.First(&task, id).Error; err != nil {
		return nil, err
	}

	// 加载子任务
	var subtasks []models.Subtask
	if err := r.db.Where("task_id = ?", id).Order("sort_order, id").Find(&subtasks).Error; err != nil {
		return nil, err
	}
	task.Subtasks = subtasks

	// 加载依赖关系
	var dependencies []models.TaskDependency
	if err := r.db.Where("task_id = ?", id).Find(&dependencies).Error; err != nil {
		return nil, err
	}
	task.Dependencies = dependencies

	// 加载分配
	var assignments []models.Assignment
	if err := r.db.Where("task_id = ?", id).Find(&assignments).Error; err != nil {
		return nil, err
	}
	task.Assignments = assignments

	return &task, nil
}

// Create 创建任务
func (r *taskRepository) Create(task *models.Task) error {
	return r.db.Create(task).Error
}

// Update 更新任务
func (r *taskRepository) Update(task *models.Task) error {
	return r.db.Save(task).Error
}

// Delete 删除任务
func (r *taskRepository) Delete(id uint64) error {
	return r.db.Delete(&models.Task{}, id).Error
}

// BatchDelete 批量删除任务
func (r *taskRepository) BatchDelete(ids []uint64) error {
	return r.db.Delete(&models.Task{}, ids).Error
}

// GetSubtasks 获取任务的子任务
func (r *taskRepository) GetSubtasks(taskID uint64) ([]models.Subtask, error) {
	var subtasks []models.Subtask
	if err := r.db.Where("task_id = ?", taskID).Order("sort_order, id").Find(&subtasks).Error; err != nil {
		return nil, err
	}
	return subtasks, nil
}

// CreateSubtask 创建子任务
func (r *taskRepository) CreateSubtask(subtask *models.Subtask) error {
	return r.db.Create(subtask).Error
}

// UpdateSubtask 更新子任务
func (r *taskRepository) UpdateSubtask(subtask *models.Subtask) error {
	return r.db.Save(subtask).Error
}

// DeleteSubtask 删除子任务
func (r *taskRepository) DeleteSubtask(taskID, subtaskID uint64) error {
	return r.db.Where("task_id = ? AND id = ?", taskID, subtaskID).Delete(&models.Subtask{}).Error
}

// DeleteAllSubtasks 删除任务的所有子任务
func (r *taskRepository) DeleteAllSubtasks(taskID uint64) error {
	return r.db.Where("task_id = ?", taskID).Delete(&models.Subtask{}).Error
}

// ReorderSubtasks 重新排序子任务
func (r *taskRepository) ReorderSubtasks(taskID uint64, subtaskIDs []uint64) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		for i, subtaskID := range subtaskIDs {
			if err := tx.Model(&models.Subtask{}).
				Where("id = ? AND task_id = ?", subtaskID, taskID).
				Update("sort_order", i).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// GetDependencies 获取任务依赖
func (r *taskRepository) GetDependencies(taskID uint64) ([]models.TaskDependency, error) {
	var dependencies []models.TaskDependency
	if err := r.db.Where("task_id = ?", taskID).Find(&dependencies).Error; err != nil {
		return nil, err
	}
	return dependencies, nil
}
