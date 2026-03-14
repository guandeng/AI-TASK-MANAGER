package handlers

import (
	"strconv"
	"time"

	"github.com/ai-task-manager/backend/internal/database"
	"github.com/ai-task-manager/backend/internal/models"
	"github.com/ai-task-manager/backend/internal/services"
	"github.com/ai-task-manager/backend/pkg/response"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// TaskHandler 任务处理器
type TaskHandler struct {
	service services.TaskService
	logger  *zap.Logger
}

// NewTaskHandler 创建任务处理器
func NewTaskHandler(logger *zap.Logger) *TaskHandler {
	return &TaskHandler{
		service: services.NewTaskService(nil), // 暂时不注入 AI 服务
		logger:  logger,
	}
}

// List 获取任务列表
func (h *TaskHandler) List(c *gin.Context) {
	// 解析筛选条件
	filters := make(map[string]interface{})
	if status := c.Query("status"); status != "" {
		filters["status"] = status
	}
	if priority := c.Query("priority"); priority != "" {
		filters["priority"] = priority
	}
	if requirementID := c.Query("requirementId"); requirementID != "" {
		if id, err := strconv.ParseUint(requirementID, 10, 64); err == nil {
			filters["requirementId"] = id
		}
	}
	if assignee := c.Query("assignee"); assignee != "" {
		filters["assignee"] = assignee
	}
	if keyword := c.Query("keyword"); keyword != "" {
		filters["keyword"] = keyword
	}

	// 分页参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	tasks, total, err := h.service.List(filters, page, pageSize)
	if err != nil {
		h.logger.Error("获取任务列表失败", zap.Error(err))
		response.Error(c, 500, "获取任务列表失败")
		return
	}

	response.SuccessPage(c, tasks, total, page, pageSize)
}

// Get 获取任务详情
func (h *TaskHandler) Get(c *gin.Context) {
	taskID, err := strconv.ParseUint(c.Param("taskId"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的任务 ID")
		return
	}

	task, err := h.service.GetByID(taskID)
	if err != nil {
		h.logger.Error("获取任务详情失败", zap.Error(err))
		response.NotFound(c, "任务不存在")
		return
	}

	response.Success(c, task)
}

// Update 更新任务
func (h *TaskHandler) Update(c *gin.Context) {
	taskID, err := strconv.ParseUint(c.Param("taskId"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的任务 ID")
		return
	}

	var updates map[string]interface{}
	if err := c.ShouldBindJSON(&updates); err != nil {
		response.BadRequest(c, "请求参数格式错误")
		return
	}

	if err := h.service.Update(taskID, updates); err != nil {
		h.logger.Error("更新任务失败", zap.Error(err))
		response.Error(c, 500, "更新任务失败")
		return
	}

	// 返回更新后的任务
	task, _ := h.service.GetByID(taskID)
	response.Success(c, task)
}

// Delete 删除任务
func (h *TaskHandler) Delete(c *gin.Context) {
	taskID, err := strconv.ParseUint(c.Param("taskId"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的任务 ID")
		return
	}

	if err := h.service.Delete(taskID); err != nil {
		h.logger.Error("删除任务失败", zap.Error(err))
		response.Error(c, 500, "删除任务失败")
		return
	}

	response.Success(c, nil)
}

// BatchDelete 批量删除任务
func (h *TaskHandler) BatchDelete(c *gin.Context) {
	var req struct {
		IDs []uint64 `json:"ids" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数格式错误")
		return
	}

	if len(req.IDs) == 0 {
		response.BadRequest(c, "请选择要删除的任务")
		return
	}

	if err := h.service.BatchDelete(req.IDs); err != nil {
		h.logger.Error("批量删除任务失败", zap.Error(err))
		response.Error(c, 500, "批量删除任务失败")
		return
	}

	response.Success(c, gin.H{"deleted": len(req.IDs)})
}

// UpdateTime 更新任务时间
func (h *TaskHandler) UpdateTime(c *gin.Context) {
	taskID, err := strconv.ParseUint(c.Param("taskId"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的任务 ID")
		return
	}

	var req struct {
		StartDate      *string  `json:"startDate"`
		DueDate        *string  `json:"dueDate"`
		EstimatedHours *float64 `json:"estimatedHours"`
		ActualHours    *float64 `json:"actualHours"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数格式错误")
		return
	}

	var startTime, dueDate *time.Time
	if req.StartDate != nil {
		t, err := time.Parse("2006-01-02", *req.StartDate)
		if err == nil {
			startTime = &t
		}
	}
	if req.DueDate != nil {
		t, err := time.Parse("2006-01-02", *req.DueDate)
		if err == nil {
			dueDate = &t
		}
	}

	if err := h.service.UpdateTime(taskID, startTime, dueDate, req.EstimatedHours, req.ActualHours); err != nil {
		h.logger.Error("更新任务时间失败", zap.Error(err))
		response.Error(c, 500, "更新任务时间失败")
		return
	}

	task, _ := h.service.GetByID(taskID)
	response.Success(c, task)
}

// UpdateSubtask 更新子任务
func (h *TaskHandler) UpdateSubtask(c *gin.Context) {
	taskID, err := strconv.ParseUint(c.Param("taskId"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的任务 ID")
		return
	}
	subtaskID, err := strconv.ParseUint(c.Param("subtaskId"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的子任务 ID")
		return
	}

	var updates map[string]interface{}
	if err := c.ShouldBindJSON(&updates); err != nil {
		response.BadRequest(c, "请求参数格式错误")
		return
	}

	if err := h.service.UpdateSubtask(taskID, subtaskID, updates); err != nil {
		h.logger.Error("更新子任务失败", zap.Error(err))
		response.Error(c, 500, "更新子任务失败")
		return
	}

	response.Success(c, nil)
}

// DeleteSubtask 删除子任务
func (h *TaskHandler) DeleteSubtask(c *gin.Context) {
	taskID, err := strconv.ParseUint(c.Param("taskId"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的任务 ID")
		return
	}
	subtaskID, err := strconv.ParseUint(c.Param("subtaskId"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的子任务 ID")
		return
	}

	if err := h.service.DeleteSubtask(taskID, subtaskID); err != nil {
		h.logger.Error("删除子任务失败", zap.Error(err))
		response.Error(c, 500, "删除子任务失败")
		return
	}

	response.Success(c, nil)
}

// DeleteAllSubtasks 删除所有子任务
func (h *TaskHandler) DeleteAllSubtasks(c *gin.Context) {
	taskID, err := strconv.ParseUint(c.Param("taskId"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的任务 ID")
		return
	}

	if err := h.service.DeleteAllSubtasks(taskID); err != nil {
		h.logger.Error("删除所有子任务失败", zap.Error(err))
		response.Error(c, 500, "删除所有子任务失败")
		return
	}

	response.Success(c, nil)
}

// ReorderSubtasks 重新排序子任务
func (h *TaskHandler) ReorderSubtasks(c *gin.Context) {
	taskID, err := strconv.ParseUint(c.Param("taskId"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的任务 ID")
		return
	}

	var req struct {
		SubtaskIDs []uint64 `json:"subtaskIds" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数格式错误")
		return
	}

	if err := h.service.ReorderSubtasks(taskID, req.SubtaskIDs); err != nil {
		h.logger.Error("重新排序子任务失败", zap.Error(err))
		response.Error(c, 500, "重新排序子任务失败")
		return
	}

	response.Success(c, nil)
}

// RegenerateSubtask 重新生成子任务
func (h *TaskHandler) RegenerateSubtask(c *gin.Context) {
	taskID, err := strconv.ParseUint(c.Param("taskId"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的任务 ID")
		return
	}
	subtaskID, err := strconv.ParseUint(c.Param("subtaskId"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的子任务 ID")
		return
	}

	if err := h.service.RegenerateSubtask(taskID, subtaskID); err != nil {
		h.logger.Error("重新生成子任务失败", zap.Error(err))
		response.Error(c, 500, "重新生成子任务失败")
		return
	}

	response.Success(c, nil)
}

// ExpandTask 同步展开任务（AI 生成子任务）
func (h *TaskHandler) ExpandTask(c *gin.Context) {
	taskID, err := strconv.ParseUint(c.Param("taskId"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的任务 ID")
		return
	}

	if err := h.service.ExpandTask(taskID, true); err != nil {
		h.logger.Error("展开任务失败", zap.Error(err))
		response.Error(c, 500, "展开任务失败")
		return
	}

	task, _ := h.service.GetByID(taskID)
	response.Success(c, task)
}

// ExpandTaskAsync 异步展开任务
func (h *TaskHandler) ExpandTaskAsync(c *gin.Context) {
	taskID, err := strconv.ParseUint(c.Param("taskId"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的任务 ID")
		return
	}

	// TODO: 实现异步展开（使用 goroutine + 消息队列）
	go func() {
		if err := h.service.ExpandTask(taskID, false); err != nil {
			h.logger.Error("异步展开任务失败", zap.Error(err))
		}
	}()

	response.Success(c, gin.H{"message": "任务展开已开始"})
}

// GetAssignments 获取任务分配列表
func (h *TaskHandler) GetAssignments(c *gin.Context) {
	taskID, err := strconv.ParseUint(c.Param("taskId"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的任务 ID")
		return
	}

	db := database.GetDB()
	var assignments []models.Assignment
	if err := db.Where("task_id = ?", taskID).Preload("Member").Find(&assignments).Error; err != nil {
		h.logger.Error("获取任务分配列表失败", zap.Error(err))
		response.Error(c, 500, "获取任务分配列表失败")
		return
	}

	response.Success(c, assignments)
}

// GetAssignmentOverview 获取分配概览
func (h *TaskHandler) GetAssignmentOverview(c *gin.Context) {
	taskID, err := strconv.ParseUint(c.Param("taskId"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的任务 ID")
		return
	}

	db := database.GetDB()
	var assignments []models.Assignment
	if err := db.Where("task_id = ?", taskID).Preload("Member").Find(&assignments).Error; err != nil {
		h.logger.Error("获取任务分配概览失败", zap.Error(err))
		response.Error(c, 500, "获取任务分配概览失败")
		return
	}

	// 统计各角色数量
	overview := gin.H{
		"total":        len(assignments),
		"assignees":    0,
		"reviewers":    0,
		"collaborators": 0,
		"members":      []gin.H{},
	}

	membersMap := make(map[uint64]gin.H)
	for _, a := range assignments {
		switch a.Role {
		case "assignee":
			overview["assignees"] = overview["assignees"].(int) + 1
		case "reviewer":
			overview["reviewers"] = overview["reviewers"].(int) + 1
		case "collaborator":
			overview["collaborators"] = overview["collaborators"].(int) + 1
		}

		if _, ok := membersMap[a.MemberID]; !ok {
			memberInfo := gin.H{
				"id":     a.MemberID,
				"name":   "",
				"avatar": nil,
				"roles":  []string{a.Role},
			}
			if a.Member.ID != 0 {
				memberInfo["name"] = a.Member.Name
				memberInfo["avatar"] = a.Member.Avatar
			}
			membersMap[a.MemberID] = memberInfo
		} else {
			memberInfo := membersMap[a.MemberID]
			roles := memberInfo["roles"].([]string)
			roles = append(roles, a.Role)
			memberInfo["roles"] = roles
			membersMap[a.MemberID] = memberInfo
		}
	}

	// 转换为数组
	members := make([]gin.H, 0, len(membersMap))
	for _, m := range membersMap {
		members = append(members, m)
	}
	overview["members"] = members

	response.Success(c, overview)
}

// CreateAssignment 创建任务分配
func (h *TaskHandler) CreateAssignment(c *gin.Context) {
	taskID, err := strconv.ParseUint(c.Param("taskId"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的任务 ID")
		return
	}

	var req struct {
		MemberID       uint64   `json:"memberId" binding:"required"`
		Role           string   `json:"role" binding:"required"`
		EstimatedHours *float64 `json:"estimatedHours"`
		ActualHours    *float64 `json:"actualHours"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数格式错误")
		return
	}

	// 验证角色
	validRoles := map[string]bool{"assignee": true, "reviewer": true, "collaborator": true}
	if !validRoles[req.Role] {
		response.BadRequest(c, "无效的角色类型")
		return
	}

	db := database.GetDB()

	// 检查是否已存在相同的分配
	var existingCount int64
	db.Model(&models.Assignment{}).Where("task_id = ? AND member_id = ? AND role = ?", taskID, req.MemberID, req.Role).Count(&existingCount)
	if existingCount > 0 {
		response.BadRequest(c, "该成员已被分配此角色")
		return
	}

	assignment := models.Assignment{
		TaskID:         taskID,
		MemberID:       req.MemberID,
		Role:           req.Role,
		EstimatedHours: req.EstimatedHours,
		ActualHours:    req.ActualHours,
	}

	if err := db.Create(&assignment).Error; err != nil {
		h.logger.Error("创建任务分配失败", zap.Error(err))
		response.Error(c, 500, "创建任务分配失败")
		return
	}

	// 加载成员信息
	db.Preload("Member").First(&assignment, assignment.ID)

	response.Success(c, assignment)
}

// DeleteAssignment 删除任务分配
func (h *TaskHandler) DeleteAssignment(c *gin.Context) {
	taskID, err := strconv.ParseUint(c.Param("taskId"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的任务 ID")
		return
	}
	assignmentID, err := strconv.ParseUint(c.Param("assignmentId"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的分配 ID")
		return
	}

	db := database.GetDB()
	if err := db.Where("id = ? AND task_id = ?", assignmentID, taskID).Delete(&models.Assignment{}).Error; err != nil {
		h.logger.Error("删除任务分配失败", zap.Error(err))
		response.Error(c, 500, "删除任务分配失败")
		return
	}

	response.Success(c, gin.H{"success": true, "message": "删除成功"})
}

// GetSubtaskAssignments 获取子任务分配
func (h *TaskHandler) GetSubtaskAssignments(c *gin.Context) {
	taskID, err := strconv.ParseUint(c.Param("taskId"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的任务 ID")
		return
	}
	subtaskID, err := strconv.ParseUint(c.Param("subtaskId"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的子任务 ID")
		return
	}

	db := database.GetDB()

	// 验证子任务属于该任务
	var subtask models.Subtask
	if err := db.Where("id = ? AND task_id = ?", subtaskID, taskID).First(&subtask).Error; err != nil {
		response.NotFound(c, "子任务不存在")
		return
	}

	var assignments []models.SubtaskAssignment
	if err := db.Where("subtask_id = ?", subtaskID).Preload("Member").Find(&assignments).Error; err != nil {
		h.logger.Error("获取子任务分配列表失败", zap.Error(err))
		response.Error(c, 500, "获取子任务分配列表失败")
		return
	}

	response.Success(c, assignments)
}

// CreateSubtaskAssignment 创建子任务分配
func (h *TaskHandler) CreateSubtaskAssignment(c *gin.Context) {
	taskID, err := strconv.ParseUint(c.Param("taskId"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的任务 ID")
		return
	}
	subtaskID, err := strconv.ParseUint(c.Param("subtaskId"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的子任务 ID")
		return
	}

	var req struct {
		MemberID       uint64   `json:"memberId" binding:"required"`
		Role           string   `json:"role" binding:"required"`
		EstimatedHours *float64 `json:"estimatedHours"`
		ActualHours    *float64 `json:"actualHours"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数格式错误")
		return
	}

	// 验证角色
	validRoles := map[string]bool{"assignee": true, "reviewer": true, "collaborator": true}
	if !validRoles[req.Role] {
		response.BadRequest(c, "无效的角色类型")
		return
	}

	db := database.GetDB()

	// 验证子任务属于该任务
	var subtask models.Subtask
	if err := db.Where("id = ? AND task_id = ?", subtaskID, taskID).First(&subtask).Error; err != nil {
		response.NotFound(c, "子任务不存在")
		return
	}

	// 检查是否已存在相同的分配
	var existingCount int64
	db.Model(&models.SubtaskAssignment{}).Where("subtask_id = ? AND member_id = ? AND role = ?", subtaskID, req.MemberID, req.Role).Count(&existingCount)
	if existingCount > 0 {
		response.BadRequest(c, "该成员已被分配此角色")
		return
	}

	assignment := models.SubtaskAssignment{
		SubtaskID:      subtaskID,
		MemberID:       req.MemberID,
		Role:           req.Role,
		EstimatedHours: req.EstimatedHours,
		ActualHours:    req.ActualHours,
	}

	if err := db.Create(&assignment).Error; err != nil {
		h.logger.Error("创建子任务分配失败", zap.Error(err))
		response.Error(c, 500, "创建子任务分配失败")
		return
	}

	// 加载成员信息
	db.Preload("Member").First(&assignment, assignment.ID)

	response.Success(c, assignment)
}

// DeleteSubtaskAssignment 删除子任务分配
func (h *TaskHandler) DeleteSubtaskAssignment(c *gin.Context) {
	taskID, err := strconv.ParseUint(c.Param("taskId"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的任务 ID")
		return
	}
	subtaskID, err := strconv.ParseUint(c.Param("subtaskId"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的子任务 ID")
		return
	}
	assignmentID, err := strconv.ParseUint(c.Param("assignmentId"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的分配 ID")
		return
	}

	db := database.GetDB()

	// 验证子任务属于该任务
	var subtask models.Subtask
	if err := db.Where("id = ? AND task_id = ?", subtaskID, taskID).First(&subtask).Error; err != nil {
		response.NotFound(c, "子任务不存在")
		return
	}

	if err := db.Where("id = ? AND subtask_id = ?", assignmentID, subtaskID).Delete(&models.SubtaskAssignment{}).Error; err != nil {
		h.logger.Error("删除子任务分配失败", zap.Error(err))
		response.Error(c, 500, "删除子任务分配失败")
		return
	}

	response.Success(c, gin.H{"success": true, "message": "删除成功"})
}
