package repository

import (
	"encoding/json"
	"time"

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
	UpdateSubtaskWithMap(taskID, subtaskID uint64, updates map[string]interface{}) error
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

	// 构建基础查询 - 排除已删除的任务
	baseQuery := r.db.Model(&models.Task{}).Where("deleted_at IS NULL")

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

	// 分页查询，使用 LEFT JOIN 获取需求标题和子任务统计
	offset := (page - 1) * pageSize

	// 构建查询 - 添加子任务统计子查询，排除已删除的任务
	query := r.db.Table("task_task").
		Select(`task_task.*,
			task_requirement.title as requirement_title,
			(SELECT COUNT(*) FROM task_subtask WHERE task_subtask.task_id = task_task.id) as subtask_count,
			(SELECT COUNT(*) FROM task_subtask WHERE task_subtask.task_id = task_task.id AND task_subtask.status = 'done') as subtask_done_count`).
		Joins("LEFT JOIN task_requirement ON task_task.requirement_id = task_requirement.id").
		Where("task_task.deleted_at IS NULL")

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

	// 使用 Scan 扫描结果，确保 requirement_title 被正确映射
	// 注意：gorm tag 用于数据库列映射，json tag 用于 JSON 序列化
	type TaskScan struct {
		ID               uint64     `gorm:"column:id" json:"id"`
		RequirementID    *uint64    `gorm:"column:requirement_id" json:"requirementId"`
		RequirementTitle string     `gorm:"column:requirement_title" json:"requirementTitle"`
		Title            string     `gorm:"column:title" json:"title"`
		TitleTrans       *string    `gorm:"column:title_trans" json:"titleTrans"`
		Description      string     `gorm:"column:description" json:"description"`
		DescriptionTrans *string    `gorm:"column:description_trans" json:"descriptionTrans"`
		Status           string     `gorm:"column:status" json:"status"`
		Priority         string     `gorm:"column:priority" json:"priority"`
		Details          string     `gorm:"column:details" json:"details"`
		DetailsTrans     *string    `gorm:"column:details_trans" json:"detailsTrans"`
		TestStrategy     string     `gorm:"column:test_strategy" json:"testStrategy"`
		TestStrategyTrans *string   `gorm:"column:test_strategy_trans" json:"testStrategyTrans"`
		Assignee         *string    `gorm:"column:assignee" json:"assignee"`
		IsExpanding      bool       `gorm:"column:is_expanding" json:"isExpanding"`
		ExpandMessageID  *uint64    `gorm:"column:expand_message_id" json:"expandMessageId"`
		StartDate        *time.Time `gorm:"column:start_date" json:"startDate"`
		DueDate          *time.Time `gorm:"column:due_date" json:"dueDate"`
		CompletedAt      *time.Time `gorm:"column:completed_at" json:"completedAt"`
		EstimatedHours   *float64   `gorm:"column:estimated_hours" json:"estimatedHours"`
		ActualHours      *float64   `gorm:"column:actual_hours" json:"actualHours"`
		CreatedAt        time.Time  `gorm:"column:created_at" json:"createdAt"`
		UpdatedAt        time.Time  `gorm:"column:updated_at" json:"updatedAt"`
		SubtaskCount     int        `gorm:"column:subtask_count" json:"subtaskCount"`
		SubtaskDoneCount int        `gorm:"column:subtask_done_count" json:"subtaskDoneCount"`
	}

	var scannedTasks []TaskScan
	if err := query.Order("task_task.created_at DESC").
		Offset(offset).
		Limit(pageSize).
		Scan(&scannedTasks).Error; err != nil {
		return nil, 0, err
	}

	// 转换为 []models.Task
	for _, st := range scannedTasks {
		tasks = append(tasks, models.Task{
			ID:               st.ID,
			RequirementID:    st.RequirementID,
			RequirementTitle: st.RequirementTitle,
			Title:            st.Title,
			TitleTrans:       st.TitleTrans,
			Description:      st.Description,
			DescriptionTrans: st.DescriptionTrans,
			Status:           st.Status,
			Priority:         st.Priority,
			Details:          st.Details,
			DetailsTrans:     st.DetailsTrans,
			TestStrategy:     st.TestStrategy,
			TestStrategyTrans: st.TestStrategyTrans,
			Assignee:         st.Assignee,
			IsExpanding:      st.IsExpanding,
			ExpandMessageID:  st.ExpandMessageID,
			StartDate:        st.StartDate,
			DueDate:          st.DueDate,
			CompletedAt:      st.CompletedAt,
			EstimatedHours:   st.EstimatedHours,
			ActualHours:      st.ActualHours,
			CreatedAt:        st.CreatedAt,
			UpdatedAt:        st.UpdatedAt,
			SubtaskCount:     st.SubtaskCount,
			SubtaskDoneCount: st.SubtaskDoneCount,
		})
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

// Delete 软删除任务
func (r *taskRepository) Delete(id uint64) error {
	now := time.Now()
	return r.db.Model(&models.Task{}).Where("id = ?", id).Update("deleted_at", &now).Error
}

// BatchDelete 批量软删除任务
func (r *taskRepository) BatchDelete(ids []uint64) error {
	now := time.Now()
	return r.db.Model(&models.Task{}).Where("id IN ?", ids).Update("deleted_at", &now).Error
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
	return r.db.Model(&models.Subtask{}).Where("id = ? AND task_id = ?", subtask.ID, subtask.TaskID).Updates(subtask).Error
}

// UpdateSubtaskWithMap 使用 map 更新子任务（只更新传入的字段）
func (r *taskRepository) UpdateSubtaskWithMap(taskID, subtaskID uint64, updates map[string]interface{}) error {
	if len(updates) == 0 {
		return nil
	}

	// 将前端驼峰命名转换为数据库蛇形命名，并处理 JSON 字段
	jsonFields := map[string]bool{
		"codeInterface":      true,
		"acceptanceCriteria": true,
		"relatedFiles":       true,
	}

	dbUpdates := make(map[string]interface{})
	for key, value := range updates {
		dbKey := camelToSnake(key)

		// 对于 JSON 字段，需要序列化为字符串
		if jsonFields[key] && value != nil {
			jsonBytes, err := json.Marshal(value)
			if err != nil {
				return err
			}
			dbUpdates[dbKey] = string(jsonBytes)
		} else {
			dbUpdates[dbKey] = value
		}
	}

	return r.db.Model(&models.Subtask{}).Where("id = ? AND task_id = ?", subtaskID, taskID).Updates(dbUpdates).Error
}

// camelToSnake 将驼峰命名转换为蛇形命名
func camelToSnake(s string) string {
	var result []rune
	for i, r := range s {
		if i > 0 && r >= 'A' && r <= 'Z' {
			result = append(result, '_', r+32) // 转小写
		} else if r >= 'A' && r <= 'Z' {
			result = append(result, r+32) // 转小写
		} else {
			result = append(result, r)
		}
	}
	return string(result)
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
