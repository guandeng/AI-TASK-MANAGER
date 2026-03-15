package handlers

import (
	"strconv"

	"github.com/ai-task-manager/backend/internal/config"
	"github.com/ai-task-manager/backend/internal/database"
	"github.com/ai-task-manager/backend/internal/models"
	"github.com/ai-task-manager/backend/internal/services"
	"github.com/ai-task-manager/backend/pkg/ai"
	"github.com/ai-task-manager/backend/pkg/response"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// ComplexityHandler 复杂度分析处理器
type ComplexityHandler struct {
	logger             *zap.Logger
	complexityService  *services.ComplexityService
	knowledgeService   *services.KnowledgeService
	aiService          services.AIService
}

// NewComplexityHandler 创建复杂度分析处理器
func NewComplexityHandler(logger *zap.Logger, cfg *config.Config) *ComplexityHandler {
	var aiSvc services.AIService
	if cfg != nil && cfg.AI.Provider != "" {
		aiSvc = ai.NewService(&cfg.AI)
	}
	return &ComplexityHandler{
		logger:            logger,
		complexityService: services.NewComplexityService(logger),
		knowledgeService:  services.NewKnowledgeService(&cfg.Knowledge, logger),
		aiService:         aiSvc,
	}
}

// AnalyzeTask 分析单个任务复杂度
// POST /api/tasks/:taskId/analyze
func (h *ComplexityHandler) AnalyzeTask(c *gin.Context) {
	taskID, err := strconv.ParseUint(c.Param("taskId"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的任务 ID")
		return
	}

	// 获取任务
	db := database.GetDB()
	var task models.Task
	if err := db.First(&task, taskID).Error; err != nil {
		response.NotFound(c, "任务不存在")
		return
	}

	// 检查 AI 服务
	if h.aiService == nil {
		response.Error(c, 500, "AI 服务未配置")
		return
	}

	// 分析任务
	analysis, err := h.complexityService.AnalyzeTask(&task, h.aiService)
	if err != nil {
		h.logger.Error("分析任务复杂度失败", zap.Error(err))
		response.Error(c, 500, "分析任务复杂度失败: "+err.Error())
		return
	}

	response.Success(c, analysis)
}

// AnalyzeRequirement 分析需求复杂度
// POST /api/requirements/:id/analyze
func (h *ComplexityHandler) AnalyzeRequirement(c *gin.Context) {
	requirementID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的需求 ID")
		return
	}

	var req struct {
		KnowledgePaths []string `json:"knowledgePaths"`
		UseKnowledge   bool     `json:"useKnowledge"`
	}
	c.ShouldBindJSON(&req)

	// 获取需求
	db := database.GetDB()
	var requirement models.Requirement
	if err := db.First(&requirement, requirementID).Error; err != nil {
		response.NotFound(c, "需求不存在")
		return
	}

	// 检查 AI 服务
	if h.aiService == nil {
		response.Error(c, 500, "AI 服务未配置")
		return
	}

	// 加载知识库
	var knowledgeContext string
	if req.UseKnowledge {
		ctx, err := h.knowledgeService.LoadKnowledge(req.KnowledgePaths)
		if err != nil {
			h.logger.Warn("加载知识库失败", zap.Error(err))
		}
		knowledgeContext = ctx
	}

	// 分析需求
	report, err := h.complexityService.AnalyzeRequirement(requirementID, &requirement, h.aiService, knowledgeContext)
	if err != nil {
		h.logger.Error("分析需求复杂度失败", zap.Error(err))
		response.Error(c, 500, "分析需求复杂度失败: "+err.Error())
		return
	}

	// 保存报告
	savedReport, err := h.complexityService.SaveReport(&requirementID, report)
	if err != nil {
		h.logger.Warn("保存复杂度报告失败", zap.Error(err))
	}

	response.Success(c, gin.H{
		"reportId": savedReport.ID,
		"analysis": report,
	})
}

// GetComplexityReport 获取复杂度报告
// GET /api/complexity/reports/:reportId
func (h *ComplexityHandler) GetComplexityReport(c *gin.Context) {
	reportID, err := strconv.ParseUint(c.Param("reportId"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的报告 ID")
		return
	}

	report, err := h.complexityService.GetReport(reportID)
	if err != nil {
		response.NotFound(c, "报告不存在")
		return
	}

	response.Success(c, report)
}

// GetRequirementReports 获取需求的所有复杂度报告
// GET /api/requirements/:id/complexity-reports
func (h *ComplexityHandler) GetRequirementReports(c *gin.Context) {
	requirementID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的需求 ID")
		return
	}

	reports, err := h.complexityService.GetReportsByRequirement(requirementID)
	if err != nil {
		h.logger.Error("获取复杂度报告失败", zap.Error(err))
		response.Error(c, 500, "获取复杂度报告失败")
		return
	}

	response.Success(c, reports)
}

// AnalyzeTasksAsync 异步分析任务复杂度
// POST /api/requirements/:id/analyze-async
func (h *ComplexityHandler) AnalyzeTasksAsync(c *gin.Context) {
	requirementID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的需求 ID")
		return
	}

	var req struct {
		KnowledgePaths []string `json:"knowledgePaths"`
		UseKnowledge   bool     `json:"useKnowledge"`
	}
	c.ShouldBindJSON(&req)

	// 检查需求是否存在
	db := database.GetDB()
	var requirement models.Requirement
	if err := db.First(&requirement, requirementID).Error; err != nil {
		response.NotFound(c, "需求不存在")
		return
	}

	// 检查 AI 服务
	if h.aiService == nil {
		response.Error(c, 500, "AI 服务未配置")
		return
	}

	// 创建消息记录
	title := "复杂度分析"
	content := requirement.Title
	message := models.Message{
		RequirementID: &requirementID,
		Type:          "complexity_analysis",
		Status:        "processing",
		Title:         title,
		Content:       &content,
	}
	if err := db.Create(&message).Error; err != nil {
		response.Error(c, 500, "创建消息记录失败")
		return
	}

	// 异步执行分析
	go func() {
		// 加载知识库
		var knowledgeContext string
		if req.UseKnowledge {
			ctx, err := h.knowledgeService.LoadKnowledge(req.KnowledgePaths)
			if err != nil {
				h.logger.Warn("加载知识库失败", zap.Error(err))
			}
			knowledgeContext = ctx
		}

		// 分析
		report, err := h.complexityService.AnalyzeRequirement(requirementID, &requirement, h.aiService, knowledgeContext)
		if err != nil {
			h.logger.Error("异步分析复杂度失败", zap.Error(err))
			errMsg := err.Error()
			db.Model(&message).Updates(map[string]interface{}{
				"status":        "failed",
				"error_message": &errMsg,
			})
			return
		}

		// 保存报告
		savedReport, err := h.complexityService.SaveReport(&requirementID, report)
		if err != nil {
			h.logger.Warn("保存复杂度报告失败", zap.Error(err))
		}

		// 更新消息状态
		resultSummary := "复杂度分析完成"
		reportID := savedReport.ID
		db.Model(&message).Updates(map[string]interface{}{
			"status":         "success",
			"result_summary": &resultSummary,
			"related_id":     &reportID,
		})
	}()

	response.Success(c, gin.H{
		"messageId": message.ID,
		"message":   "复杂度分析已开始，完成后会通知您",
	})
}

// DependencyHandler 依赖管理处理器
type DependencyHandler struct {
	logger *zap.Logger
	db     *gorm.DB
}

// NewDependencyHandler 创建依赖管理处理器
func NewDependencyHandler(logger *zap.Logger) *DependencyHandler {
	return &DependencyHandler{
		logger: logger,
		db:     database.GetDB(),
	}
}

// FixDependencies 自动修复无效的依赖关系
// POST /api/tasks/dependencies/fix
func (h *DependencyHandler) FixDependencies(c *gin.Context) {
	// 获取所有依赖关系
	var dependencies []models.TaskDependency
	if err := h.db.Find(&dependencies).Error; err != nil {
		h.logger.Error("获取依赖关系失败", zap.Error(err))
		response.Error(c, 500, "获取依赖关系失败")
		return
	}

	var fixed []uint64
	var removed []uint64

	for _, dep := range dependencies {
		// 检查任务是否存在
		var taskExists int64
		h.db.Model(&models.Task{}).Where("id = ?", dep.TaskID).Count(&taskExists)
		if taskExists == 0 {
			// 任务不存在，删除依赖
			h.db.Delete(&dep)
			removed = append(removed, dep.ID)
			continue
		}

		// 检查依赖的任务是否存在
		var depTaskExists int64
		h.db.Model(&models.Task{}).Where("id = ?", dep.DependsOnTaskID).Count(&depTaskExists)
		if depTaskExists == 0 {
			// 依赖的任务不存在，删除依赖
			h.db.Delete(&dep)
			removed = append(removed, dep.ID)
			continue
		}

		// 检查自引用
		if dep.TaskID == dep.DependsOnTaskID {
			h.db.Delete(&dep)
			removed = append(removed, dep.ID)
			continue
		}

		fixed = append(fixed, dep.ID)
	}

	response.Success(c, gin.H{
		"fixed":  len(fixed),
		"removed": len(removed),
		"removedIds": removed,
	})
}

// GetDependencyGraph 获取依赖关系图
// GET /api/tasks/dependencies/graph
func (h *DependencyHandler) GetDependencyGraph(c *gin.Context) {
	requirementID := c.Query("requirementId")

	// 获取任务
	var tasks []models.Task
	query := h.db.Model(&models.Task{})
	if requirementID != "" {
		query = query.Where("requirement_id = ?", requirementID)
	}
	if err := query.Find(&tasks).Error; err != nil {
		h.logger.Error("获取任务失败", zap.Error(err))
		response.Error(c, 500, "获取任务失败")
		return
	}

	// 获取依赖关系
	var dependencies []models.TaskDependency
	depQuery := h.db.Model(&models.TaskDependency{})
	if requirementID != "" {
		depQuery = depQuery.Where("task_id IN (SELECT id FROM task_task WHERE requirement_id = ?)", requirementID)
	}
	if err := depQuery.Find(&dependencies).Error; err != nil {
		h.logger.Error("获取依赖关系失败", zap.Error(err))
		response.Error(c, 500, "获取依赖关系失败")
		return
	}

	// 构建图数据
	nodes := make([]map[string]interface{}, len(tasks))
	for i, task := range tasks {
		nodes[i] = map[string]interface{}{
			"id":     task.ID,
			"label":  task.Title,
			"status": task.Status,
		}
	}

	edges := make([]map[string]interface{}, len(dependencies))
	for i, dep := range dependencies {
		edges[i] = map[string]interface{}{
			"source": dep.DependsOnTaskID,
			"target": dep.TaskID,
		}
	}

	response.Success(c, gin.H{
		"nodes": nodes,
		"edges": edges,
	})
}

// GetNextTasks 获取接下来可执行的任务（依赖已满足）
// GET /api/tasks/next
func (h *DependencyHandler) GetNextTasks(c *gin.Context) {
	requirementID := c.Query("requirementId")
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "5"))

	// 获取所有待处理任务
	var pendingTasks []models.Task
	query := h.db.Where("status = ?", "pending")
	if requirementID != "" {
		query = query.Where("requirement_id = ?", requirementID)
	}
	if err := query.Find(&pendingTasks).Error; err != nil {
		h.logger.Error("获取待处理任务失败", zap.Error(err))
		response.Error(c, 500, "获取待处理任务失败")
		return
	}

	// 获取所有依赖关系
	var dependencies []models.TaskDependency
	depQuery := h.db.Model(&models.TaskDependency{})
	if requirementID != "" {
		depQuery = depQuery.Where("task_id IN (SELECT id FROM task_task WHERE requirement_id = ?)", requirementID)
	}
	if err := depQuery.Find(&dependencies).Error; err != nil {
		h.logger.Error("获取依赖关系失败", zap.Error(err))
		response.Error(c, 500, "获取依赖关系失败")
		return
	}

	// 构建依赖映射
	depMap := make(map[uint64][]uint64)
	for _, dep := range dependencies {
		depMap[dep.TaskID] = append(depMap[dep.TaskID], dep.DependsOnTaskID)
	}

	// 获取已完成任务
	var completedTasks []models.Task
	h.db.Where("status = ?", "done").Find(&completedTasks)
	completedIDs := make(map[uint64]bool)
	for _, t := range completedTasks {
		completedIDs[t.ID] = true
	}

	// 找出可执行的任务
	var readyTasks []models.Task
	for _, task := range pendingTasks {
		deps := depMap[task.ID]
		allDepsCompleted := true
		for _, depID := range deps {
			if !completedIDs[depID] {
				allDepsCompleted = false
				break
			}
		}
		if allDepsCompleted {
			readyTasks = append(readyTasks, task)
		}
		if len(readyTasks) >= limit {
			break
		}
	}

	// 按优先级排序
	priorityOrder := map[string]int{"high": 0, "medium": 1, "low": 2}
	for i := 0; i < len(readyTasks)-1; i++ {
		for j := i + 1; j < len(readyTasks); j++ {
			if priorityOrder[readyTasks[i].Priority] > priorityOrder[readyTasks[j].Priority] {
				readyTasks[i], readyTasks[j] = readyTasks[j], readyTasks[i]
			}
		}
	}

	response.Success(c, readyTasks)
}
