package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ai-task-manager/backend/internal/config"
	"github.com/ai-task-manager/backend/internal/database"
	"github.com/ai-task-manager/backend/internal/handlers"
	"github.com/ai-task-manager/backend/internal/middleware"
	"github.com/ai-task-manager/backend/internal/services"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func main() {
	// 加载配置
	cfg, err := config.Load("")
	if err != nil {
		fmt.Printf("加载配置失败: %v\n", err)
		os.Exit(1)
	}

	// 初始化日志
	var logger *zap.Logger
	if cfg.Server.Mode == "release" {
		logger, _ = zap.NewProduction()
		gin.SetMode(gin.ReleaseMode)
	} else {
		logger, _ = zap.NewDevelopment()
		gin.SetMode(gin.DebugMode)
	}
	defer logger.Sync()

	// 初始化数据库
	if err := database.Init(&cfg.Database, logger); err != nil {
		logger.Fatal("初始化数据库失败", zap.Error(err))
	}
	defer database.Close()

	// 创建 Gin 引擎
	r := gin.New()

	// 使用中间件
	r.Use(middleware.Recovery(logger))
	r.Use(middleware.Logger(logger))
	r.Use(middleware.CORS())

	// 注册路由
	registerRoutes(r, logger, cfg)

	// 启动服务器
	addr := fmt.Sprintf(":%d", cfg.Server.Port)
	srv := &http.Server{
		Addr:    addr,
		Handler: r,
	}

	// 优雅关闭
	go func() {
		logger.Info("服务器启动", zap.String("addr", addr))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("服务器启动失败", zap.Error(err))
		}
	}()

	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("正在关闭服务器...")

	// 给 5 秒时间处理未完成的请求
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Fatal("服务器强制关闭", zap.Error(err))
	}

	logger.Info("服务器已关闭")
}

// registerRoutes 注册所有路由
func registerRoutes(r *gin.Engine, logger *zap.Logger, cfg *config.Config) {
	// 健康检查
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// API 路由组
	api := r.Group("/api")
	{
		// 任务管理
		taskHandler := handlers.NewTaskHandler(logger, cfg)
		tasks := api.Group("/tasks")
		{
			tasks.GET("", taskHandler.List)
			tasks.GET("/:taskId", taskHandler.Get)
			tasks.POST("/:taskId/update", taskHandler.Update)
			tasks.POST("/:taskId/delete", taskHandler.Delete)
			tasks.POST("/batch-delete", taskHandler.BatchDelete)
			tasks.POST("/:taskId/time", taskHandler.UpdateTime)

			// 子任务
			tasks.POST("/:taskId/subtasks/:subtaskId/update", taskHandler.UpdateSubtask)
			tasks.POST("/:taskId/subtasks/:subtaskId/delete", taskHandler.DeleteSubtask)
			tasks.POST("/:taskId/subtasks/delete", taskHandler.DeleteAllSubtasks)
			tasks.POST("/:taskId/subtasks/reorder", taskHandler.ReorderSubtasks)
			tasks.POST("/:taskId/subtasks/:subtaskId/regenerate", taskHandler.RegenerateSubtask)

			// AI 功能
			tasks.POST("/:taskId/expand", taskHandler.ExpandTask)
			tasks.POST("/:taskId/expand-async", taskHandler.ExpandTaskAsync)

			// 任务分配
			tasks.GET("/:taskId/assignments", taskHandler.GetAssignments)
			tasks.GET("/:taskId/assignments/overview", taskHandler.GetAssignmentOverview)
			tasks.POST("/:taskId/assignments", taskHandler.CreateAssignment)
			tasks.POST("/:taskId/assignments/:assignmentId/delete", taskHandler.DeleteAssignment)

			// 子任务分配
			tasks.GET("/:taskId/subtasks/:subtaskId/assignments", taskHandler.GetSubtaskAssignments)
			tasks.POST("/:taskId/subtasks/:subtaskId/assignments", taskHandler.CreateSubtaskAssignment)
			tasks.POST("/:taskId/subtasks/:subtaskId/assignments/:assignmentId/delete", taskHandler.DeleteSubtaskAssignment)

			// 评论管理
			commentHandler := handlers.NewCommentHandler(logger)
			tasks.GET("/:taskId/comments", commentHandler.List)
			tasks.GET("/:taskId/comments/tree", commentHandler.GetTree)
			tasks.GET("/:taskId/comments/statistics", commentHandler.GetStatistics)
			tasks.GET("/:taskId/comments/:commentId", commentHandler.Get)
			tasks.POST("/:taskId/comments", commentHandler.Create)
			tasks.POST("/:taskId/comments/:commentId/update", commentHandler.Update)
			tasks.POST("/:taskId/comments/:commentId/delete", commentHandler.Delete)
			tasks.GET("/:taskId/comments/:commentId/replies", commentHandler.GetReplies)
		}

		// 需求管理
		requirementHandler := handlers.NewRequirementHandler(logger, cfg)
		requirements := api.Group("/requirements")
		{
			requirements.GET("", requirementHandler.List)
			requirements.GET("/statistics", requirementHandler.Statistics)
			requirements.GET("/:id", requirementHandler.Get)
			requirements.POST("", requirementHandler.Create)
			requirements.POST("/:id/update", requirementHandler.Update)
			requirements.POST("/:id/delete", requirementHandler.Delete)
			requirements.POST("/:id/documents", requirementHandler.UploadDocument)
			requirements.POST("/:id/documents/:docId/delete", requirementHandler.DeleteDocument)
			requirements.GET("/:id/documents/:docId/download", requirementHandler.DownloadDocument)
			requirements.POST("/:id/split-tasks", requirementHandler.SplitTasks)
			requirements.POST("/:id/split-tasks-async", requirementHandler.SplitTasksAsync)
		}

		// 成员管理
		memberHandler := handlers.NewMemberHandler(logger)
		members := api.Group("/members")
		{
			members.GET("", memberHandler.List)
			members.GET("/:id", memberHandler.Get)
			members.POST("", memberHandler.Create)
			members.POST("/:id/update", memberHandler.Update)
			members.POST("/:id/delete", memberHandler.Delete)
			members.GET("/:id/assignments", memberHandler.GetAssignments)
			members.GET("/:id/workload", memberHandler.GetWorkload)
		}

		// 活动日志
		activityHandler := handlers.NewActivityHandler(logger)
		activities := api.Group("/activities")
		{
			activities.GET("", activityHandler.List)
			activities.GET("/statistics", activityHandler.Statistics)
		}
		// 任务活动日志
		api.GET("/tasks/:taskId/activities", activityHandler.ListByTask)

		// 消息管理
		messageHandler := handlers.NewMessageHandler(logger)
		messages := api.Group("/messages")
		{
			messages.GET("", messageHandler.List)
			messages.GET("/unread-count", messageHandler.UnreadCount)
			messages.POST("/:id/read", messageHandler.MarkRead)
			messages.POST("/read-all", messageHandler.MarkAllRead)
			messages.POST("/:id/delete", messageHandler.Delete)
		}

		// 菜单管理
		menuHandler := handlers.NewMenuHandler(logger)
		menus := api.Group("/menus")
		{
			menus.GET("", menuHandler.List)
			menus.GET("/tree", menuHandler.Tree)
			menus.GET("/:key", menuHandler.Get)
			menus.POST("", menuHandler.Create)
			menus.POST("/:key/update", menuHandler.Update)
			menus.POST("/:key/delete", menuHandler.Delete)
			menus.POST("/batch-delete", menuHandler.BatchDelete)
			menus.POST("/reorder", menuHandler.Reorder)
			menus.POST("/:key/move", menuHandler.Move)
			menus.POST("/:key/toggle", menuHandler.Toggle)
		}

		// 配置管理
		configHandler := handlers.NewConfigHandler(logger, cfg)
		configGroup := api.Group("/config")
		{
			configGroup.GET("", configHandler.Get)
			configGroup.POST("/update", configHandler.Update)
			configGroup.POST("/ai-provider", configHandler.UpdateAIProvider)
			configGroup.POST("/ai-provider/:provider", configHandler.UpdateSpecificProvider)
			configGroup.POST("/reset", configHandler.Reset)
		}

		// 模板管理
		templateHandler := handlers.NewTemplateHandler(logger)

		// 项目模板
		projectTemplates := api.Group("/templates/projects")
		{
			projectTemplates.GET("", templateHandler.ListProjectTemplates)
			projectTemplates.GET("/:id", templateHandler.GetProjectTemplate)
			projectTemplates.POST("", templateHandler.CreateProjectTemplate)
			projectTemplates.POST("/:id/update", templateHandler.UpdateProjectTemplate)
			projectTemplates.POST("/:id/delete", templateHandler.DeleteProjectTemplate)
			projectTemplates.POST("/:id/instantiate", templateHandler.InstantiateProjectTemplate)
		}

		// 任务模板
		taskTemplates := api.Group("/templates/tasks")
		{
			taskTemplates.GET("", templateHandler.ListTaskTemplates)
			taskTemplates.GET("/:id", templateHandler.GetTaskTemplate)
			taskTemplates.POST("", templateHandler.CreateTaskTemplate)
			taskTemplates.POST("/:id/update", templateHandler.UpdateTaskTemplate)
			taskTemplates.POST("/:id/delete", templateHandler.DeleteTaskTemplate)
			taskTemplates.POST("/:id/instantiate", templateHandler.InstantiateTaskTemplate)
		}

		// 备份管理
		backupService := services.NewBackupService(logger)
		backupHandler := handlers.NewBackupHandler(logger, backupService)
		backupGroup := api.Group("/requirements/:id/backups")
		{
			backupGroup.GET("", backupHandler.List)
			backupGroup.POST("/create", backupHandler.Create)
			backupGroup.POST("/:backupId/restore", backupHandler.Restore)
			backupGroup.POST("/:backupId/delete", backupHandler.Delete)
			backupGroup.GET("/schedule", backupHandler.GetSchedule)
			backupGroup.POST("/schedule/update", backupHandler.UpdateSchedule)
			backupGroup.POST("/schedule/disable", backupHandler.DisableSchedule)
		}

		// 启动备份调度器
		backupService.StartScheduler()
	}
}
