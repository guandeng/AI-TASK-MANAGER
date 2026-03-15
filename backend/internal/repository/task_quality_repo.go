package repository

import (
	"github.com/ai-task-manager/backend/internal/database"
	"github.com/ai-task-manager/backend/internal/models"
	"gorm.io/gorm"
)

// TaskQualityRepository 任务质量评分仓储接口
type TaskQualityRepository interface {
	// 评分 CRUD
	Create(score *models.TaskQualityScore) error
	GetByID(id uint64) (*models.TaskQualityScore, error)
	GetByTaskID(taskID uint64, page, pageSize int) ([]models.TaskQualityScore, int64, error)
	GetLatestByTaskID(taskID uint64) (*models.TaskQualityScore, error)
	Delete(id uint64) error

	// 版本管理
	GetNextVersion(taskID uint64) (int, error)
}

// taskQualityRepository 任务质量评分仓储实现
type taskQualityRepository struct {
	db *gorm.DB
}

// NewTaskQualityRepository 创建任务质量评分仓储
func NewTaskQualityRepository() TaskQualityRepository {
	return &taskQualityRepository{db: database.GetDB()}
}

// Create 创建评分记录
func (r *taskQualityRepository) Create(score *models.TaskQualityScore) error {
	return r.db.Create(score).Error
}

// GetByID 根据 ID 获取评分
func (r *taskQualityRepository) GetByID(id uint64) (*models.TaskQualityScore, error) {
	var score models.TaskQualityScore
	err := r.db.Where("id = ?", id).First(&score).Error
	if err != nil {
		return nil, err
	}
	return &score, nil
}

// GetByTaskID 获取任务的评分历史
func (r *taskQualityRepository) GetByTaskID(taskID uint64, page, pageSize int) ([]models.TaskQualityScore, int64, error) {
	var scores []models.TaskQualityScore
	var total int64

	query := r.db.Model(&models.TaskQualityScore{}).Where("task_id = ?", taskID)

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询，按版本号倒序排列
	offset := (page - 1) * pageSize
	if err := query.Order("version DESC").Offset(offset).Limit(pageSize).Find(&scores).Error; err != nil {
		return nil, 0, err
	}

	return scores, total, nil
}

// GetLatestByTaskID 获取任务的最新评分
func (r *taskQualityRepository) GetLatestByTaskID(taskID uint64) (*models.TaskQualityScore, error) {
	var score models.TaskQualityScore
	err := r.db.Where("task_id = ?", taskID).Order("version DESC").First(&score).Error
	if err != nil {
		return nil, err
	}
	return &score, nil
}

// Delete 删除评分记录
func (r *taskQualityRepository) Delete(id uint64) error {
	return r.db.Delete(&models.TaskQualityScore{}, id).Error
}

// GetNextVersion 获取下一个版本号
func (r *taskQualityRepository) GetNextVersion(taskID uint64) (int, error) {
	var maxVersion int
	err := r.db.Model(&models.TaskQualityScore{}).
		Where("task_id = ?", taskID).
		Select("COALESCE(MAX(version), 0)").
		Scan(&maxVersion).Error
	if err != nil {
		return 1, err
	}
	return maxVersion + 1, nil
}
