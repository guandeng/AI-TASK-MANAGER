package handlers

import (
	"strconv"

	"github.com/ai-task-manager/backend/internal/database"
	"github.com/ai-task-manager/backend/internal/models"
	"github.com/ai-task-manager/backend/pkg/response"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// MemberHandler 成员处理器
type MemberHandler struct {
	logger *zap.Logger
}

// NewMemberHandler 创建成员处理器
func NewMemberHandler(logger *zap.Logger) *MemberHandler {
	return &MemberHandler{logger: logger}
}

// List 获取成员列表
func (h *MemberHandler) List(c *gin.Context) {
	db := database.GetDB()

	// 解析筛选条件
	query := db.Model(&models.Member{})

	if status := c.Query("status"); status != "" {
		query = query.Where("status = ?", status)
	}
	if role := c.Query("role"); role != "" {
		query = query.Where("role = ?", role)
	}
	if department := c.Query("department"); department != "" {
		query = query.Where("department = ?", department)
	}
	if keyword := c.Query("keyword"); keyword != "" {
		query = query.Where("name LIKE ? OR email LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
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

	// 计算总数
	var total int64
	if err := query.Count(&total).Error; err != nil {
		h.logger.Error("获取成员总数失败", zap.Error(err))
		response.Error(c, 500, "获取成员列表失败")
		return
	}

	// 分页查询
	var members []models.Member
	offset := (page - 1) * pageSize
	if err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&members).Error; err != nil {
		h.logger.Error("获取成员列表失败", zap.Error(err))
		response.Error(c, 500, "获取成员列表失败")
		return
	}

	response.SuccessPage(c, members, total, page, pageSize)
}

// Get 获取成员详情
func (h *MemberHandler) Get(c *gin.Context) {
	db := database.GetDB()

	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的成员 ID")
		return
	}

	var member models.Member
	if err := db.First(&member, id).Error; err != nil {
		response.NotFound(c, "成员不存在")
		return
	}

	response.Success(c, member)
}

// CreateMemberRequest 创建成员请求
type CreateMemberRequest struct {
	Name       string   `json:"name" binding:"required"`
	Email      *string  `json:"email"`
	Avatar     *string  `json:"avatar"`
	Role       string   `json:"role"`
	Department *string  `json:"department"`
	Skills     []string `json:"skills"`
	Status     string   `json:"status"`
}

// Create 创建成员
func (h *MemberHandler) Create(c *gin.Context) {
	db := database.GetDB()

	var req CreateMemberRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数格式错误")
		return
	}

	// 设置默认值
	if req.Role == "" {
		req.Role = "member"
	}
	if req.Status == "" {
		req.Status = "active"
	}

	member := models.Member{
		Name:       req.Name,
		Email:      req.Email,
		Avatar:     req.Avatar,
		Role:       req.Role,
		Department: req.Department,
		Skills:     req.Skills,
		Status:     req.Status,
	}

	if err := db.Create(&member).Error; err != nil {
		h.logger.Error("创建成员失败", zap.Error(err))
		response.Error(c, 500, "创建成员失败")
		return
	}

	response.Success(c, member)
}

// UpdateMemberRequest 更新成员请求
type UpdateMemberRequest struct {
	Name       string   `json:"name"`
	Email      *string  `json:"email"`
	Avatar     *string  `json:"avatar"`
	Role       string   `json:"role"`
	Department *string  `json:"department"`
	Skills     []string `json:"skills"`
	Status     string   `json:"status"`
}

// Update 更新成员
func (h *MemberHandler) Update(c *gin.Context) {
	db := database.GetDB()

	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的成员 ID")
		return
	}

	var member models.Member
	if err := db.First(&member, id).Error; err != nil {
		response.NotFound(c, "成员不存在")
		return
	}

	var req UpdateMemberRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数格式错误")
		return
	}

	// 更新字段
	updates := make(map[string]interface{})
	if req.Name != "" {
		updates["name"] = req.Name
	}
	if req.Email != nil {
		updates["email"] = req.Email
	}
	if req.Avatar != nil {
		updates["avatar"] = req.Avatar
	}
	if req.Role != "" {
		updates["role"] = req.Role
	}
	if req.Department != nil {
		updates["department"] = req.Department
	}
	if req.Skills != nil {
		updates["skills"] = req.Skills
	}
	if req.Status != "" {
		updates["status"] = req.Status
	}

	if len(updates) > 0 {
		if err := db.Model(&member).Updates(updates).Error; err != nil {
			h.logger.Error("更新成员失败", zap.Error(err))
			response.Error(c, 500, "更新成员失败")
			return
		}
	}

	// 重新获取更新后的数据
	db.First(&member, id)
	response.Success(c, member)
}

// Delete 删除成员
func (h *MemberHandler) Delete(c *gin.Context) {
	db := database.GetDB()

	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的成员 ID")
		return
	}

	// 检查成员是否存在
	var member models.Member
	if err := db.First(&member, id).Error; err != nil {
		response.NotFound(c, "成员不存在")
		return
	}

	// 删除成员（会级联删除相关的分配记录）
	if err := db.Delete(&member).Error; err != nil {
		h.logger.Error("删除成员失败", zap.Error(err))
		response.Error(c, 500, "删除成员失败")
		return
	}

	response.Success(c, gin.H{"success": true, "message": "删除成功"})
}

// MemberAssignment 成员任务分配响应
type MemberAssignment struct {
	TaskID         uint64   `json:"taskId"`
	TaskTitle      string   `json:"taskTitle"`
	TaskStatus     string   `json:"taskStatus"`
	Role           string   `json:"role"`
	EstimatedHours *float64 `json:"estimatedHours"`
	ActualHours    *float64 `json:"actualHours"`
	CreatedAt      string   `json:"createdAt"`
}

// GetAssignments 获取成员的任务分配
func (h *MemberHandler) GetAssignments(c *gin.Context) {
	db := database.GetDB()

	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的成员 ID")
		return
	}

	// 获取任务分配
	var assignments []models.Assignment
	if err := db.Where("member_id = ?", id).Preload("Task").Find(&assignments).Error; err != nil {
		h.logger.Error("获取成员任务分配失败", zap.Error(err))
		response.Error(c, 500, "获取成员任务分配失败")
		return
	}

	// 获取子任务分配
	var subtaskAssignments []models.SubtaskAssignment
	if err := db.Where("member_id = ?", id).Preload("Subtask").Preload("Subtask.Task").Find(&subtaskAssignments).Error; err != nil {
		h.logger.Error("获取成员子任务分配失败", zap.Error(err))
		response.Error(c, 500, "获取成员子任务分配失败")
		return
	}

	// 组装响应
	result := make([]MemberAssignment, 0)

	// 添加任务分配
	for _, a := range assignments {
		if a.Task.ID != 0 {
			result = append(result, MemberAssignment{
				TaskID:         a.TaskID,
				TaskTitle:      a.Task.Title,
				TaskStatus:     a.Task.Status,
				Role:           a.Role,
				EstimatedHours: a.EstimatedHours,
				ActualHours:    a.ActualHours,
				CreatedAt:      a.CreatedAt.Format("2006-01-02 15:04:05"),
			})
		}
	}

	// 添加子任务分配
	for _, sa := range subtaskAssignments {
		if sa.Subtask.ID != 0 && sa.Subtask.Task.ID != 0 {
			result = append(result, MemberAssignment{
				TaskID:         sa.Subtask.TaskID,
				TaskTitle:      sa.Subtask.Task.Title + " / " + sa.Subtask.Title,
				TaskStatus:     sa.Subtask.Task.Status,
				Role:           sa.Role,
				EstimatedHours: sa.EstimatedHours,
				ActualHours:    sa.ActualHours,
				CreatedAt:      sa.CreatedAt.Format("2006-01-02 15:04:05"),
			})
		}
	}

	// 筛选条件
	if role := c.Query("role"); role != "" {
		filtered := make([]MemberAssignment, 0)
		for _, ma := range result {
			if ma.Role == role {
				filtered = append(filtered, ma)
			}
		}
		result = filtered
	}

	if status := c.Query("status"); status != "" {
		filtered := make([]MemberAssignment, 0)
		for _, ma := range result {
			if ma.TaskStatus == status {
				filtered = append(filtered, ma)
			}
		}
		result = filtered
	}

	response.Success(c, result)
}

// MemberWorkload 成员工作量响应
type MemberWorkload struct {
	MemberID       uint64  `json:"memberId"`
	TaskCount      int64   `json:"taskCount"`
	ActiveTasks    int64   `json:"activeTasks"`
	CompletedTasks int64   `json:"completedTasks"`
	EstimatedHours float64 `json:"estimatedHours"`
	ActualHours    float64 `json:"actualHours"`
}

// GetWorkload 获取成员工作量
func (h *MemberHandler) GetWorkload(c *gin.Context) {
	db := database.GetDB()

	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的成员 ID")
		return
	}

	workload := MemberWorkload{
		MemberID: id,
	}

	// 统计任务分配
	var taskAssignments []models.Assignment
	db.Where("member_id = ?", id).Preload("Task").Find(&taskAssignments)

	taskSet := make(map[uint64]bool)
	for _, a := range taskAssignments {
		if a.Task.ID != 0 {
			taskSet[a.TaskID] = true
			if a.EstimatedHours != nil {
				workload.EstimatedHours += *a.EstimatedHours
			}
			if a.ActualHours != nil {
				workload.ActualHours += *a.ActualHours
			}
			if a.Task.Status == "done" || a.Task.Status == "completed" {
				workload.CompletedTasks++
			} else if a.Task.Status == "in-progress" {
				workload.ActiveTasks++
			}
		}
	}

	// 统计子任务分配
	var subtaskAssignments []models.SubtaskAssignment
	db.Where("member_id = ?", id).Preload("Subtask").Find(&subtaskAssignments)

	subtaskSet := make(map[uint64]bool)
	for _, sa := range subtaskAssignments {
		if sa.Subtask.ID != 0 {
			subtaskSet[sa.SubtaskID] = true
			// 关联的任务也计入
			taskSet[sa.Subtask.TaskID] = true
			if sa.EstimatedHours != nil {
				workload.EstimatedHours += *sa.EstimatedHours
			}
			if sa.ActualHours != nil {
				workload.ActualHours += *sa.ActualHours
			}
			if sa.Subtask.Status == "done" || sa.Subtask.Status == "completed" {
				workload.CompletedTasks++
			} else if sa.Subtask.Status == "in-progress" {
				workload.ActiveTasks++
			}
		}
	}

	workload.TaskCount = int64(len(taskSet) + len(subtaskSet))

	response.Success(c, workload)
}
