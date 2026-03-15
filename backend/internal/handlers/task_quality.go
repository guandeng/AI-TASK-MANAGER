package handlers

import (
	"strconv"

	"github.com/ai-task-manager/backend/internal/config"
	"github.com/ai-task-manager/backend/internal/services"
	"github.com/ai-task-manager/backend/pkg/ai"
	"github.com/ai-task-manager/backend/pkg/response"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// TaskQualityHandler 任务质量评分处理器
type TaskQualityHandler struct {
	service services.TaskQualityService
	logger  *zap.Logger
}

// NewTaskQualityHandler 创建任务质量评分处理器
func NewTaskQualityHandler(logger *zap.Logger, cfg *config.Config) *TaskQualityHandler {
	var aiSvc services.AIService
	if cfg != nil && cfg.AI.Provider != "" {
		aiSvc = ai.NewService(&cfg.AI)
	}
	return &TaskQualityHandler{
		service: services.NewTaskQualityService(aiSvc),
		logger:  logger,
	}
}

// Score 对任务进行评分
// POST /api/tasks/:taskId/score
func (h *TaskQualityHandler) Score(c *gin.Context) {
	taskID, err := strconv.ParseUint(c.Param("taskId"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的任务 ID")
		return
	}

	h.logger.Info("开始对任务进行评分", zap.Uint64("taskID", taskID))

	score, err := h.service.Score(taskID)
	if err != nil {
		h.logger.Error("评分失败", zap.Uint64("taskID", taskID), zap.Error(err))
		response.Error(c, 500, "评分失败："+err.Error())
		return
	}

	h.logger.Info("评分完成",
		zap.Uint64("taskID", taskID),
		zap.Int("version", score.Version),
		zap.Float64("totalScore", score.TotalScore),
	)

	response.Success(c, score)
}

// ListScores 获取评分历史列表
// GET /api/tasks/:taskId/scores
func (h *TaskQualityHandler) ListScores(c *gin.Context) {
	taskID, err := strconv.ParseUint(c.Param("taskId"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的任务 ID")
		return
	}

	// 解析分页参数
	page := 1
	pageSize := 20
	if p := c.Query("page"); p != "" {
		if parsed, err := strconv.Atoi(p); err == nil && parsed > 0 {
			page = parsed
		}
	}
	if ps := c.Query("pageSize"); ps != "" {
		if parsed, err := strconv.Atoi(ps); err == nil && parsed > 0 && parsed <= 100 {
			pageSize = parsed
		}
	}

	scores, total, err := h.service.GetScoreHistory(taskID, page, pageSize)
	if err != nil {
		h.logger.Error("获取评分历史失败", zap.Uint64("taskID", taskID), zap.Error(err))
		response.Error(c, 500, "获取评分历史失败："+err.Error())
		return
	}

	response.Success(c, gin.H{
		"list":     scores,
		"total":    total,
		"page":     page,
		"pageSize": pageSize,
	})
}

// GetScore 获取评分详情
// GET /api/tasks/:taskId/scores/:scoreId
func (h *TaskQualityHandler) GetScore(c *gin.Context) {
	scoreID, err := strconv.ParseUint(c.Param("scoreId"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的评分 ID")
		return
	}

	score, err := h.service.GetScoreDetail(scoreID)
	if err != nil {
		h.logger.Error("获取评分详情失败", zap.Uint64("scoreID", scoreID), zap.Error(err))
		response.Error(c, 500, "获取评分详情失败："+err.Error())
		return
	}

	response.Success(c, score)
}

// DeleteScore 删除评分记录
// DELETE /api/tasks/:taskId/scores/:scoreId
func (h *TaskQualityHandler) DeleteScore(c *gin.Context) {
	scoreID, err := strconv.ParseUint(c.Param("scoreId"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的评分 ID")
		return
	}

	if err := h.service.DeleteScore(scoreID); err != nil {
		h.logger.Error("删除评分记录失败", zap.Uint64("scoreID", scoreID), zap.Error(err))
		response.Error(c, 500, "删除评分记录失败："+err.Error())
		return
	}

	response.Success(c, nil)
}

// RestoreScore 恢复到评分版本
// POST /api/tasks/:taskId/scores/:scoreId/restore
func (h *TaskQualityHandler) RestoreScore(c *gin.Context) {
	scoreID, err := strconv.ParseUint(c.Param("scoreId"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的评分 ID")
		return
	}

	if err := h.service.RestoreScore(scoreID); err != nil {
		h.logger.Error("恢复评分版本失败", zap.Uint64("scoreID", scoreID), zap.Error(err))
		response.Error(c, 500, "恢复评分版本失败："+err.Error())
		return
	}

	response.Success(c, nil)
}
