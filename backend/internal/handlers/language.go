package handlers

import (
	"strconv"

	"github.com/ai-task-manager/backend/internal/database"
	"github.com/ai-task-manager/backend/internal/models"
	"github.com/ai-task-manager/backend/pkg/response"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// LanguageHandler 语言处理器
type LanguageHandler struct {
	logger *zap.Logger
}

// NewLanguageHandler 创建语言处理器
func NewLanguageHandler(logger *zap.Logger) *LanguageHandler {
	return &LanguageHandler{logger: logger}
}

// List 获取语言列表
func (h *LanguageHandler) List(c *gin.Context) {
	db := database.GetDB()

	var languages []models.Language
	query := db.Model(&models.Language{})

	// 按分类筛选
	if category := c.Query("category"); category != "" {
		query = query.Where("category = ?", category)
	}

	// 只返回启用的��言（除非指定了 all 参数）
	if c.Query("all") != "true" {
		query = query.Where("is_active = ?", true)
	}

	if err := query.Order("category ASC, sort_order ASC, id ASC").Find(&languages).Error; err != nil {
		h.logger.Error("获取语言列表失败", zap.Error(err))
		response.Error(c, 500, "获取语言列表失败")
		return
	}

	// 调试日志
	h.logger.Info("获取语言列表", zap.Int("count", len(languages)))
	for i, lang := range languages {
		h.logger.Info("语言数据",
			zap.Int("index", i),
			zap.Uint64("id", lang.ID),
			zap.String("name", lang.Name),
			zap.String("category", string(lang.Category)),
		)
	}

	response.Success(c, languages)
}

// Get 获取语言详情
func (h *LanguageHandler) Get(c *gin.Context) {
	db := database.GetDB()

	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的语言 ID")
		return
	}

	var language models.Language
	if err := db.First(&language, id).Error; err != nil {
		response.NotFound(c, "语言不存在")
		return
	}

	response.Success(c, language)
}

// Create 创建语言
func (h *LanguageHandler) Create(c *gin.Context) {
	db := database.GetDB()

	var req models.Language
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数格式错误")
		return
	}

	if req.Name == "" {
		response.BadRequest(c, "语言名称不能为空")
		return
	}

	if err := db.Create(&req).Error; err != nil {
		h.logger.Error("创建语言失败", zap.Error(err))
		response.Error(c, 500, "创建语言失败")
		return
	}

	response.Success(c, req)
}

// Update 更新语言
func (h *LanguageHandler) Update(c *gin.Context) {
	db := database.GetDB()

	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的语言 ID")
		return
	}

	var language models.Language
	if err := db.First(&language, id).Error; err != nil {
		response.NotFound(c, "语言不存在")
		return
	}

	var updates map[string]interface{}
	if err := c.ShouldBindJSON(&updates); err != nil {
		response.BadRequest(c, "请求参数格式错误")
		return
	}

	if err := db.Model(&language).Updates(updates).Error; err != nil {
		h.logger.Error("更新语言失败", zap.Error(err))
		response.Error(c, 500, "更新语言失败")
		return
	}

	db.First(&language, id)
	response.Success(c, language)
}

// Delete 删除语言
func (h *LanguageHandler) Delete(c *gin.Context) {
	db := database.GetDB()

	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的语言 ID")
		return
	}

	if err := db.Delete(&models.Language{}, id).Error; err != nil {
		h.logger.Error("删除语言失败", zap.Error(err))
		response.Error(c, 500, "删除语言失败")
		return
	}

	response.Success(c, nil)
}
