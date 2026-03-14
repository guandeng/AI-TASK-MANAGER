package handlers

import (
	"fmt"

	"github.com/ai-task-manager/backend/internal/models"
	"github.com/ai-task-manager/backend/internal/repository"
	"github.com/ai-task-manager/backend/pkg/response"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// MenuHandler 菜单处理器
type MenuHandler struct {
	logger *zap.Logger
	repo   repository.MenuRepository
}

// NewMenuHandler 创建菜单处理器
func NewMenuHandler(logger *zap.Logger) *MenuHandler {
	return &MenuHandler{
		logger: logger,
		repo:   repository.NewMenuRepository(),
	}
}

// List 获取菜单列表
func (h *MenuHandler) List(c *gin.Context) {
	page := c.DefaultQuery("page", "1")
	pageSize := c.DefaultQuery("pageSize", "100")

	var p int
	var ps int
	if _, err := fmt.Sscanf(page, "%d", &p); err != nil {
		p = 1
	}
	if _, err := fmt.Sscanf(pageSize, "%d", &ps); err != nil {
		ps = 100
	}

	menus, total, err := h.repo.List(p, ps)
	if err != nil {
		h.logger.Error("获取菜单列表失败", zap.Error(err))
		response.ServerError(c, "获取菜单列表失败")
		return
	}

	response.SuccessPage(c, menus, total, p, ps)
}

// Tree 获取菜单树
func (h *MenuHandler) Tree(c *gin.Context) {
	menus, err := h.repo.GetTree()
	if err != nil {
		h.logger.Error("获取菜单树失败", zap.Error(err))
		response.ServerError(c, "获取菜单树失败")
		return
	}

	response.Success(c, menus)
}

// Get 获取单个菜单
func (h *MenuHandler) Get(c *gin.Context) {
	key := c.Param("key")
	if key == "" {
		response.BadRequest(c, "菜单标识不能为空")
		return
	}

	menu, err := h.repo.GetByKey(key)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			response.NotFound(c, "菜单不存在")
			return
		}
		h.logger.Error("获取菜单详情失败", zap.Error(err))
		response.ServerError(c, "获取菜单详情失败")
		return
	}

	response.Success(c, menu)
}

// Create 创建菜单
func (h *MenuHandler) Create(c *gin.Context) {
	var menu models.Menu
	if err := c.ShouldBindJSON(&menu); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}

	// 验证必填字段
	if menu.Key == "" {
		response.BadRequest(c, "菜单标识不能为空")
		return
	}
	if menu.Title == "" {
		response.BadRequest(c, "菜单名称不能为空")
		return
	}

	// 检查 key 是否已存在
	existingMenu, _ := h.repo.GetByKey(menu.Key)
	if existingMenu != nil {
		response.BadRequest(c, "菜单标识已存在")
		return
	}

	if err := h.repo.Create(&menu); err != nil {
		h.logger.Error("创建菜单失败", zap.Error(err))
		response.ServerError(c, "创建菜单失败")
		return
	}

	response.SuccessWithMessage(c, "创建菜单成功", gin.H{"menu": menu})
}

// Update 更新菜单
func (h *MenuHandler) Update(c *gin.Context) {
	key := c.Param("key")
	if key == "" {
		response.BadRequest(c, "菜单标识不能为空")
		return
	}

	// 检查菜单是否存在
	menu, err := h.repo.GetByKey(key)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			response.NotFound(c, "菜单不存在")
			return
		}
		h.logger.Error("获取菜单失败", zap.Error(err))
		response.ServerError(c, "获取菜单失败")
		return
	}

	// 绑定更新数据
	var updateData models.Menu
	if err := c.ShouldBindJSON(&updateData); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}

	// 更新字段
	if updateData.Title != "" {
		menu.Title = updateData.Title
	}
	if updateData.ParentKey != nil {
		menu.ParentKey = updateData.ParentKey
	}
	if updateData.Path != nil {
		menu.Path = updateData.Path
	}
	if updateData.RouteName != nil {
		menu.RouteName = updateData.RouteName
	}
	if updateData.Icon != nil {
		menu.Icon = updateData.Icon
	}
	if updateData.HideInMenu != nil {
		menu.HideInMenu = updateData.HideInMenu
	}
	if updateData.Fixed != nil {
		menu.Fixed = updateData.Fixed
	}
	if updateData.I18nKey != nil {
		menu.I18nKey = updateData.I18nKey
	}
	if updateData.Href != nil {
		menu.Href = updateData.Href
	}
	if updateData.NewWindow != nil {
		menu.NewWindow = updateData.NewWindow
	}
	menu.Sort = updateData.Sort
	menu.Enabled = updateData.Enabled

	if err := h.repo.Update(menu); err != nil {
		h.logger.Error("更新菜单失败", zap.Error(err))
		response.ServerError(c, "更新菜单失败")
		return
	}

	response.SuccessWithMessage(c, "更新菜单成功", gin.H{"menu": menu})
}

// Delete 删除菜单
func (h *MenuHandler) Delete(c *gin.Context) {
	key := c.Param("key")
	if key == "" {
		response.BadRequest(c, "菜单标识不能为空")
		return
	}

	if err := h.repo.Delete(key); err != nil {
		h.logger.Error("删除菜单失败", zap.Error(err))
		response.ServerError(c, "删除菜单失败")
		return
	}

	response.SuccessWithMessage(c, "删除菜单成功", nil)
}

// BatchDelete 批量删除菜单
func (h *MenuHandler) BatchDelete(c *gin.Context) {
	var req struct {
		Keys []string `json:"keys"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}

	if len(req.Keys) == 0 {
		response.BadRequest(c, "菜单标识列表不能为空")
		return
	}

	if err := h.repo.BatchDelete(req.Keys); err != nil {
		h.logger.Error("批量删除菜单失败", zap.Error(err))
		response.ServerError(c, "批量删除菜单失败")
		return
	}

	response.SuccessWithMessage(c, "批量删除成功", gin.H{"deletedKeys": req.Keys})
}

// Reorder 菜单排序
func (h *MenuHandler) Reorder(c *gin.Context) {
	var orderData []map[string]interface{}
	if err := c.ShouldBindJSON(&orderData); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}

	if err := h.repo.Reorder(orderData); err != nil {
		h.logger.Error("菜单排序失败", zap.Error(err))
		response.ServerError(c, "菜单排序失败")
		return
	}

	response.SuccessWithMessage(c, "菜单排序成功", nil)
}

// Move 移动菜单
func (h *MenuHandler) Move(c *gin.Context) {
	key := c.Param("key")
	if key == "" {
		response.BadRequest(c, "菜单标识不能为空")
		return
	}

	var req struct {
		TargetParentKey *string `json:"targetParentKey"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}

	if err := h.repo.Move(key, req.TargetParentKey); err != nil {
		h.logger.Error("移动菜单失败", zap.Error(err))
		response.ServerError(c, "移动菜单失败")
		return
	}

	response.SuccessWithMessage(c, "移动菜单成功", nil)
}

// Toggle 切换菜单启用状态
func (h *MenuHandler) Toggle(c *gin.Context) {
	key := c.Param("key")
	if key == "" {
		response.BadRequest(c, "菜单标识不能为空")
		return
	}

	var req struct {
		Enabled bool `json:"enabled"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}

	if err := h.repo.ToggleEnabled(key, req.Enabled); err != nil {
		h.logger.Error("切换菜单状态失败", zap.Error(err))
		response.ServerError(c, "切换菜单状态失败")
		return
	}

	response.SuccessWithMessage(c, "菜单状态已更新", nil)
}
