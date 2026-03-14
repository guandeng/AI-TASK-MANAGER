package repository

import (
	"github.com/ai-task-manager/backend/internal/database"
	"github.com/ai-task-manager/backend/internal/models"
	"gorm.io/gorm"
)

// BackupRepository 备份数据访问层
type BackupRepository interface {
	CreateBackup(backup *models.Backup) error
	GetBackupsByRequirementID(requirementID uint64, page, pageSize int) ([]models.Backup, int64, error)
	GetBackupByID(id uint64) (*models.Backup, error)
	DeleteBackup(id uint64) error
	GetScheduleByRequirementID(requirementID uint64) (*models.BackupSchedule, error)
	UpsertSchedule(schedule *models.BackupSchedule) error
	DeleteSchedule(requirementID uint64) error
	UpdateLastBackupAt(requirementID uint64) error
	GetEnabledSchedules() ([]models.BackupSchedule, error)
}

type backupRepository struct{}

// NewBackupRepository 创建备份数据访问层实例
func NewBackupRepository() BackupRepository {
	return &backupRepository{}
}

// CreateBackup 创建备份记录
func (r *backupRepository) CreateBackup(backup *models.Backup) error {
	db := database.GetDB()
	return db.Create(backup).Error
}

// GetBackupsByRequirementID 获取指定需求的备份列表
func (r *backupRepository) GetBackupsByRequirementID(requirementID uint64, page, pageSize int) ([]models.Backup, int64, error) {
	db := database.GetDB()
	var backups []models.Backup
	var total int64

	offset := (page - 1) * pageSize

	if err := db.Model(&models.Backup{}).Where("requirement_id = ?", requirementID).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := db.Where("requirement_id = ?", requirementID).
		Order("created_at DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&backups).Error

	return backups, total, err
}

// GetBackupByID 根据 ID 获取备份
func (r *backupRepository) GetBackupByID(id uint64) (*models.Backup, error) {
	db := database.GetDB()
	var backup models.Backup
	err := db.First(&backup, id).Error
	if err != nil {
		return nil, err
	}
	return &backup, nil
}

// DeleteBackup 删除备份
func (r *backupRepository) DeleteBackup(id uint64) error {
	db := database.GetDB()
	return db.Delete(&models.Backup{}, id).Error
}

// GetScheduleByRequirementID 获取指定需求的备份计划
func (r *backupRepository) GetScheduleByRequirementID(requirementID uint64) (*models.BackupSchedule, error) {
	db := database.GetDB()
	var schedule models.BackupSchedule
	err := db.Where("requirement_id = ?", requirementID).First(&schedule).Error
	if err != nil {
		return nil, err
	}
	return &schedule, nil
}

// UpsertSchedule 更新或插入备份计划
func (r *backupRepository) UpsertSchedule(schedule *models.BackupSchedule) error {
	db := database.GetDB()
	var existing models.BackupSchedule

	// 检查是否已存在
	result := db.Where("requirement_id = ?", schedule.RequirementID).First(&existing)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			// 不存在则插入
			return db.Create(schedule).Error
		}
		return result.Error
	}

	// 存在则更新
	return db.Model(&existing).Updates(map[string]interface{}{
		"enabled":         schedule.Enabled,
		"interval_type":   schedule.IntervalType,
		"interval_value":  schedule.IntervalValue,
		"retention_count": schedule.RetentionCount,
	}).Error
}

// DeleteSchedule 删除备份计划
func (r *backupRepository) DeleteSchedule(requirementID uint64) error {
	db := database.GetDB()
	return db.Where("requirement_id = ?", requirementID).Delete(&models.BackupSchedule{}).Error
}

// UpdateLastBackupAt 更新最后备份时间
func (r *backupRepository) UpdateLastBackupAt(requirementID uint64) error {
	db := database.GetDB()
	now := database.GetDB().NowFunc()
	return db.Model(&models.BackupSchedule{}).
		Where("requirement_id = ?", requirementID).
		Update("last_backup_at", now).Error
}

// GetEnabledSchedules 获取所有启用的备份计划
func (r *backupRepository) GetEnabledSchedules() ([]models.BackupSchedule, error) {
	db := database.GetDB()
	var schedules []models.BackupSchedule
	err := db.Where("enabled = ?", true).Find(&schedules).Error
	return schedules, err
}
