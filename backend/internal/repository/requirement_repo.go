package repository

import (
	"github.com/ai-task-manager/backend/internal/database"
	"github.com/ai-task-manager/backend/internal/models"
	"gorm.io/gorm"
)

// RequirementRepository 需求仓储接口
type RequirementRepository interface {
	GetByID(id uint64) (*models.Requirement, error)
}

type requirementRepository struct {
	db *gorm.DB
}

// NewRequirementRepository 创建需求仓储
func NewRequirementRepository() RequirementRepository {
	return &requirementRepository{db: database.GetDB()}
}

// GetByID 根据 ID 获取需求
func (r *requirementRepository) GetByID(id uint64) (*models.Requirement, error) {
	var requirement models.Requirement
	err := r.db.First(&requirement, id).Error
	if err != nil {
		return nil, err
	}
	return &requirement, nil
}
