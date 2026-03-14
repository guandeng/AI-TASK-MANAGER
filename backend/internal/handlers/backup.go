package handlers

import (
	"strconv"

	"github.com/ai-task-manager/backend/internal/services"
	"github.com/ai-task-manager/backend/pkg/response"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// BackupHandler 备份处理器
type BackupHandler struct {
	logger        *zap.Logger
	backupService services.BackupService
}

// NewBackupHandler 创建备份处理器
func NewBackupHandler(logger *zap.Logger, backupService services.BackupService) *BackupHandler {
	return &BackupHandler{
		logger:        logger,
		backupService: backupService,
	}
}

// List 获取备份列表
// GET /api/requirements/:id/backups
func (h *BackupHandler) List(c *gin.Context) {
	requirementID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的需求 ID")
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	backups, total, err := h.backupService.GetBackups(requirementID, page, pageSize)
	if err != nil {
		h.logger.Error("获取备份列表失败", zap.Error(err))
		response.Error(c, 500, "获取备份列表失败")
		return
	}

	response.SuccessPage(c, backups, total, page, pageSize)
}

// Create 立即创建备份
// POST /api/requirements/:id/backups/create
func (h *BackupHandler) Create(c *gin.Context) {
	requirementID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的需求 ID")
		return
	}

	backup, err := h.backupService.CreateBackup(requirementID)
	if err != nil {
		h.logger.Error("创建备份失败", zap.Error(err))
		response.Error(c, 500, "创建备份失败："+err.Error())
		return
	}

	response.Success(c, backup)
}

// Restore 恢复备份
// POST /api/requirements/:id/backups/:backupId/restore
func (h *BackupHandler) Restore(c *gin.Context) {
	backupID, err := strconv.ParseUint(c.Param("backupId"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的备份 ID")
		return
	}

	if err := h.backupService.RestoreBackup(backupID); err != nil {
		h.logger.Error("恢复备份失败", zap.Uint64("backupID", backupID), zap.Error(err))
		response.Error(c, 500, "恢复备份失败："+err.Error())
		return
	}

	response.Success(c, gin.H{"message": "恢复成功"})
}

// Delete 删除备份
// POST /api/requirements/:id/backups/:backupId/delete
func (h *BackupHandler) Delete(c *gin.Context) {
	backupID, err := strconv.ParseUint(c.Param("backupId"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的备份 ID")
		return
	}

	if err := h.backupService.DeleteBackup(backupID); err != nil {
		h.logger.Error("删除备份失败", zap.Uint64("backupID", backupID), zap.Error(err))
		response.Error(c, 500, "删除备份失败")
		return
	}

	response.Success(c, nil)
}

// GetSchedule 获取备份计划
// GET /api/requirements/:id/backups/schedule
func (h *BackupHandler) GetSchedule(c *gin.Context) {
	requirementID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的需求 ID")
		return
	}

	schedule, err := h.backupService.GetSchedule(requirementID)
	if err != nil {
		h.logger.Error("获取备份计划失败", zap.Error(err))
		response.Error(c, 500, "获取备份计划失败")
		return
	}

	response.Success(c, schedule)
}

// UpdateSchedule 更新备份计划
// POST /api/requirements/:id/backups/schedule/update
func (h *BackupHandler) UpdateSchedule(c *gin.Context) {
	requirementID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的需求 ID")
		return
	}

	var req struct {
		Enabled        bool   `json:"enabled"`
		IntervalType   string `json:"intervalType"`
		IntervalValue  int    `json:"intervalValue"`
		RetentionCount int    `json:"retentionCount"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数格式错误")
		return
	}

	// 验证参数
	if req.IntervalType != "minute" && req.IntervalType != "hour" {
		response.BadRequest(c, "无效的间隔类型，只能是 minute 或 hour")
		return
	}
	if req.IntervalValue < 1 {
		response.BadRequest(c, "间隔值必须大于 0")
		return
	}
	if req.RetentionCount < 1 {
		req.RetentionCount = 10
	}

	if err := h.backupService.UpdateSchedule(requirementID, req.Enabled, req.IntervalType, req.IntervalValue, req.RetentionCount); err != nil {
		h.logger.Error("更新备份计划失败", zap.Error(err))
		response.Error(c, 500, "更新备份计划失败")
		return
	}

	response.Success(c, nil)
}

// DisableSchedule 禁用备份计划
// POST /api/requirements/:id/backups/schedule/disable
func (h *BackupHandler) DisableSchedule(c *gin.Context) {
	requirementID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的需求 ID")
		return
	}

	// 获取现有计划
	schedule, err := h.backupService.GetSchedule(requirementID)
	if err != nil {
		response.Error(c, 500, "获取备份计划失败")
		return
	}

	if schedule != nil {
		// 禁用计划
		if err := h.backupService.UpdateSchedule(requirementID, false, schedule.IntervalType, schedule.IntervalValue, schedule.RetentionCount); err != nil {
			h.logger.Error("禁用备份计划失败", zap.Error(err))
			response.Error(c, 500, "禁用备份计划失败")
			return
		}
	}

	response.Success(c, nil)
}
