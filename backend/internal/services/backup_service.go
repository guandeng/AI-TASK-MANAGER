package services

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/ai-task-manager/backend/internal/database"
	"github.com/ai-task-manager/backend/internal/models"
	"github.com/ai-task-manager/backend/internal/repository"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// BackupService 备份服务接口
type BackupService interface {
	CreateBackup(requirementID uint64) (*models.Backup, error)
	RestoreBackup(backupID uint64) error
	GetBackups(requirementID uint64, page, pageSize int) ([]models.Backup, int64, error)
	GetBackupByID(id uint64) (*models.Backup, error)
	DeleteBackup(backupID uint64) error
	GetSchedule(requirementID uint64) (*models.BackupSchedule, error)
	UpdateSchedule(requirementID uint64, enabled bool, intervalType string, intervalValue int, retentionCount int) error
	DeleteSchedule(requirementID uint64) error
	StartScheduler()
	StopScheduler()
}

type backupService struct {
	repo        repository.BackupRepository
	taskRepo    repository.TaskRepository
	requirementRepo repository.RequirementRepository
	logger      *zap.Logger

	// 定时任务相关
	schedulerCtx    context.Context
	schedulerCancel context.CancelFunc
	tickers         map[uint64]*time.Ticker
	mu              sync.Mutex
}

// NewBackupService 创建备份服务
func NewBackupService(logger *zap.Logger) BackupService {
	return &backupService{
		repo:        repository.NewBackupRepository(),
		taskRepo:    repository.NewTaskRepository(),
		requirementRepo: repository.NewRequirementRepository(),
		logger:      logger,
		tickers:     make(map[uint64]*time.Ticker),
	}
}

// CreateBackup 创建备份
func (s *backupService) CreateBackup(requirementID uint64) (*models.Backup, error) {
	// 获取需求下的所有任务
	tasks, _, err := s.taskRepo.List(map[string]interface{}{"requirement_id": requirementID}, 1, 1000)
	if err != nil {
		return nil, fmt.Errorf("获取任务列表失败：%w", err)
	}

	// 获取需求详情
	requirement, err := s.requirementRepo.GetByID(requirementID)
	if err != nil {
		return nil, fmt.Errorf("获取需求失败：%w", err)
	}

	// 构建备份数据快照
	snapshot := map[string]interface{}{
		"requirement": requirement,
		"tasks":       tasks,
	}

	dataJSON, err := json.Marshal(snapshot)
	if err != nil {
		return nil, fmt.Errorf("序列化备份数据失败：%w", err)
	}

	backup := &models.Backup{
		RequirementID: requirementID,
		BackupType:    "full",
		DataSnapshot:  string(dataJSON),
		TaskCount:     len(tasks),
		Status:        "success",
	}

	if err := s.repo.CreateBackup(backup); err != nil {
		return nil, fmt.Errorf("创建备份记录失败：%w", err)
	}

	// 更新备份计划的上次备份时间
	_ = s.repo.UpdateLastBackupAt(requirementID)

	// 清理过期备份
	s.cleanupOldBackups(requirementID)

	s.logger.Info("备份创建成功", zap.Uint64("requirementID", requirementID), zap.Int("taskCount", len(tasks)))

	return backup, nil
}

// RestoreBackup 恢复备份
func (s *backupService) RestoreBackup(backupID uint64) error {
	backup, err := s.repo.GetBackupByID(backupID)
	if err != nil {
		return fmt.Errorf("获取备份失败：%w", err)
	}

	if backup.Status != "success" {
		return fmt.Errorf("备份状态异常，无法恢复")
	}

	// 解析备份数据
	var snapshot map[string]interface{}
	if err := json.Unmarshal([]byte(backup.DataSnapshot), &snapshot); err != nil {
		return fmt.Errorf("解析备份数据失败：%w", err)
	}

	db := database.GetDB()

	// 在事务中执行恢复
	return db.Transaction(func(tx *gorm.DB) error {
		// 恢复需求
		requirementData, ok := snapshot["requirement"].(map[string]interface{})
		if !ok {
			return fmt.Errorf("备份数据中需求格式错误")
		}

		// 更新需求状态
		reqID := backup.RequirementID
		updateData := map[string]interface{}{
			"title":   requirementData["title"],
			"content": requirementData["content"],
			"status":  requirementData["status"],
			"priority": requirementData["priority"],
		}
		if err := tx.Model(&models.Requirement{}).Where("id = ?", reqID).Updates(updateData).Error; err != nil {
			return fmt.Errorf("恢复需求失败：%w", err)
		}

		// 恢复任务
		tasksData, ok := snapshot["tasks"].([]interface{})
		if ok && len(tasksData) > 0 {
			// 先删除现有任务
			if err := tx.Where("requirement_id = ?", reqID).Delete(&models.Task{}).Error; err != nil {
				return fmt.Errorf("删除现有任务失败：%w", err)
			}

			// 重新插入任务
			for _, taskData := range tasksData {
				taskMap, ok := taskData.(map[string]interface{})
				if !ok {
					continue
				}

				task := models.Task{
					RequirementID:    &reqID,
					Title:            getString(taskMap, "title"),
					Description:      getString(taskMap, "description"),
					Status:           getString(taskMap, "status", "pending"),
					Priority:         getString(taskMap, "priority", "medium"),
					Details:          getString(taskMap, "details"),
					TestStrategy:     getString(taskMap, "testStrategy"),
					RequirementTitle: getString(taskMap, "requirement_title"),
				}

				if err := tx.Create(&task).Error; err != nil {
					return fmt.Errorf("恢复任务失败：%w", err)
				}
			}
		}

		return nil
	})
}

// GetBackups 获取备份列表
func (s *backupService) GetBackups(requirementID uint64, page, pageSize int) ([]models.Backup, int64, error) {
	return s.repo.GetBackupsByRequirementID(requirementID, page, pageSize)
}

// GetBackupByID 根据 ID 获取备份
func (s *backupService) GetBackupByID(id uint64) (*models.Backup, error) {
	return s.repo.GetBackupByID(id)
}

// DeleteBackup 删除备份
func (s *backupService) DeleteBackup(backupID uint64) error {
	return s.repo.DeleteBackup(backupID)
}

// GetSchedule 获取备份计划
func (s *backupService) GetSchedule(requirementID uint64) (*models.BackupSchedule, error) {
	return s.repo.GetScheduleByRequirementID(requirementID)
}

// UpdateSchedule 更新备份计划
func (s *backupService) UpdateSchedule(requirementID uint64, enabled bool, intervalType string, intervalValue int, retentionCount int) error {
	schedule := &models.BackupSchedule{
		RequirementID:  requirementID,
		Enabled:        enabled,
		IntervalType:   intervalType,
		IntervalValue:  intervalValue,
		RetentionCount: retentionCount,
	}

	if err := s.repo.UpsertSchedule(schedule); err != nil {
		return fmt.Errorf("更新备份计划失败：%w", err)
	}

	// 如果启用了备份，启动定时任务
	if enabled {
		s.startBackupTicker(requirementID, intervalType, intervalValue)
	} else {
		s.stopBackupTicker(requirementID)
	}

	return nil
}

// DeleteSchedule 删除备份计划
func (s *backupService) DeleteSchedule(requirementID uint64) error {
	s.stopBackupTicker(requirementID)
	return s.repo.DeleteSchedule(requirementID)
}

// StartScheduler 启动调度器
func (s *backupService) StartScheduler() {
	s.schedulerCtx, s.schedulerCancel = context.WithCancel(context.Background())

	// 加载所有启用的备份计划
	schedules, err := s.repo.GetEnabledSchedules()
	if err != nil {
		s.logger.Error("加载备份计划失败", zap.Error(err))
		return
	}

	for _, schedule := range schedules {
		if schedule.Enabled {
			s.startBackupTicker(schedule.RequirementID, schedule.IntervalType, schedule.IntervalValue)
		}
	}

	s.logger.Info("备份调度器已启动", zap.Int("activeSchedules", len(s.tickers)))
}

// StopScheduler 停止调度器
func (s *backupService) StopScheduler() {
	s.mu.Lock()
	defer s.mu.Unlock()

	for requirementID, ticker := range s.tickers {
		ticker.Stop()
		delete(s.tickers, requirementID)
	}

	if s.schedulerCancel != nil {
		s.schedulerCancel()
	}

	s.logger.Info("备份调度器已停止")
}

// startBackupTicker 启动备份定时器
func (s *backupService) startBackupTicker(requirementID uint64, intervalType string, intervalValue int) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 停止旧的定时器
	if ticker, ok := s.tickers[requirementID]; ok {
		ticker.Stop()
	}

	// 计算间隔
	var interval time.Duration
	switch intervalType {
	case "minute":
		interval = time.Duration(intervalValue) * time.Minute
	case "hour":
		interval = time.Duration(intervalValue) * time.Hour
	default:
		interval = 5 * time.Minute // 默认 5 分钟
	}

	s.logger.Info("启动备份定时任务", zap.Uint64("requirementID", requirementID), zap.Duration("interval", interval))

	ticker := time.NewTicker(interval)
	s.tickers[requirementID] = ticker

	go func() {
		for {
			select {
			case <-ticker.C:
				s.logger.Info("定时备份触发", zap.Uint64("requirementID", requirementID))
				_, err := s.CreateBackup(requirementID)
				if err != nil {
					s.logger.Error("定时备份失败", zap.Uint64("requirementID", requirementID), zap.Error(err))
				}
			case <-s.schedulerCtx.Done():
				return
			}
		}
	}()
}

// stopBackupTicker 停止备份定时器
func (s *backupService) stopBackupTicker(requirementID uint64) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if ticker, ok := s.tickers[requirementID]; ok {
		ticker.Stop()
		delete(s.tickers, requirementID)
		s.logger.Info("停止备份定时任务", zap.Uint64("requirementID", requirementID))
	}
}

// cleanupOldBackups 清理过期备份
func (s *backupService) cleanupOldBackups(requirementID uint64) {
	schedule, err := s.repo.GetScheduleByRequirementID(requirementID)
	if err != nil {
		return
	}

	retentionCount := schedule.RetentionCount
	if retentionCount <= 0 {
		retentionCount = 10 // 默认保留 10 个
	}

	// 获取所有备份
	backups, _, err := s.repo.GetBackupsByRequirementID(requirementID, 1, 1000)
	if err != nil {
		return
	}

	// 删除超出保留数量的备份
	if len(backups) > retentionCount {
		for i := retentionCount; i < len(backups); i++ {
			_ = s.repo.DeleteBackup(backups[i].ID)
		}
	}
}

// 辅助函数：从 map 中获取字符串值
func getString(m map[string]interface{}, key string, defaultVal ...string) string {
	if v, ok := m[key].(string); ok {
		return v
	}
	if len(defaultVal) > 0 {
		return defaultVal[0]
	}
	return ""
}

// Ensure interface implementation
var _ BackupService = (*backupService)(nil)
