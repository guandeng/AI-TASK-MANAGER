package handlers

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/ai-task-manager/backend/internal/config"
	"github.com/ai-task-manager/backend/internal/database"
	"github.com/ai-task-manager/backend/internal/models"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func setupComplexityHandlerTest(t *testing.T) (*ComplexityHandler, *gin.Engine, sqlmock.Sqlmock) {
	logger := zap.NewNop()

	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("创建 mock 数据库失败: %v", err)
	}

	gormDB, err := gorm.Open(mysql.New(mysql.Config{
		Conn:                      sqlDB,
		SkipInitializeWithVersion: true,
	}), &gorm.Config{})
	if err != nil {
		t.Fatalf("创建 gorm 连接失败: %v", err)
	}

	database.DB = gormDB

	cfg := &config.Config{
		Knowledge: config.KnowledgeConfig{
			Enabled:   true,
			Paths:     []string{},
			MaxSize:   500,
			MaxFiles:  50,
			FileTypes: []string{".md", ".txt"},
		},
	}

	handler := NewComplexityHandler(logger, cfg)
	router := gin.New()

	return handler, router, mock
}

func TestComplexityHandler_AnalyzeTask(t *testing.T) {
	t.Run("无效的任务ID", func(t *testing.T) {
		handler, router, _ := setupComplexityHandlerTest(t)
		router.POST("/tasks/:taskId/analyze", handler.AnalyzeTask)

		req := httptest.NewRequest(http.MethodPost, "/tasks/abc/analyze", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("期望状态码 %d, 实际 %d", http.StatusBadRequest, w.Code)
		}
	})

	t.Run("任务不存在", func(t *testing.T) {
		handler, router, mock := setupComplexityHandlerTest(t)
		router.POST("/tasks/:taskId/analyze", handler.AnalyzeTask)

		mock.ExpectQuery("SELECT \\* FROM `task_task`").
			WillReturnRows(sqlmock.NewRows([]string{"id"}))

		req := httptest.NewRequest(http.MethodPost, "/tasks/999/analyze", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("期望状态码 %d, 实际 %d", http.StatusNotFound, w.Code)
		}
	})
}

func TestComplexityHandler_AnalyzeRequirement(t *testing.T) {
	t.Run("无效的需求ID", func(t *testing.T) {
		handler, router, _ := setupComplexityHandlerTest(t)
		router.POST("/requirements/:id/analyze", handler.AnalyzeRequirement)

		req := httptest.NewRequest(http.MethodPost, "/requirements/abc/analyze", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("期望状态码 %d, 实际 %d", http.StatusBadRequest, w.Code)
		}
	})

	t.Run("需求不存在", func(t *testing.T) {
		handler, router, mock := setupComplexityHandlerTest(t)
		router.POST("/requirements/:id/analyze", handler.AnalyzeRequirement)

		mock.ExpectQuery("SELECT \\* FROM `task_requirement`").
			WillReturnRows(sqlmock.NewRows([]string{"id"}))

		req := httptest.NewRequest(http.MethodPost, "/requirements/999/analyze", bytes.NewBufferString("{}"))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("期望状态码 %d, 实际 %d", http.StatusNotFound, w.Code)
		}
	})
}

func TestComplexityHandler_GetComplexityReport(t *testing.T) {
	t.Run("无效的报告ID", func(t *testing.T) {
		handler, router, _ := setupComplexityHandlerTest(t)
		router.GET("/complexity/reports/:reportId", handler.GetComplexityReport)

		req := httptest.NewRequest(http.MethodGet, "/complexity/reports/abc", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("期望状态码 %d, 实际 %d", http.StatusBadRequest, w.Code)
		}
	})
}

func TestComplexityHandler_GetRequirementReports(t *testing.T) {
	t.Run("无效的需求ID", func(t *testing.T) {
		handler, router, _ := setupComplexityHandlerTest(t)
		router.GET("/requirements/:id/complexity-reports", handler.GetRequirementReports)

		req := httptest.NewRequest(http.MethodGet, "/requirements/abc/complexity-reports", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("期望状态码 %d, 实际 %d", http.StatusBadRequest, w.Code)
		}
	})
}

func TestComplexityHandler_AnalyzeTasksAsync(t *testing.T) {
	t.Run("无效的需求ID", func(t *testing.T) {
		handler, router, _ := setupComplexityHandlerTest(t)
		router.POST("/requirements/:id/analyze-async", handler.AnalyzeTasksAsync)

		req := httptest.NewRequest(http.MethodPost, "/requirements/abc/analyze-async", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("期望状态码 %d, 实际 %d", http.StatusBadRequest, w.Code)
		}
	})

	t.Run("需求不存在", func(t *testing.T) {
		handler, router, mock := setupComplexityHandlerTest(t)
		router.POST("/requirements/:id/analyze-async", handler.AnalyzeTasksAsync)

		mock.ExpectQuery("SELECT \\* FROM `task_requirement`").
			WillReturnRows(sqlmock.NewRows([]string{"id"}))

		req := httptest.NewRequest(http.MethodPost, "/requirements/999/analyze-async", bytes.NewBufferString("{}"))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("期望状态码 %d, 实际 %d", http.StatusNotFound, w.Code)
		}
	})
}

// DependencyHandler 测试

func setupDependencyHandlerTest(t *testing.T) (*DependencyHandler, *gin.Engine, sqlmock.Sqlmock) {
	logger := zap.NewNop()

	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("创建 mock 数据库失败: %v", err)
	}

	gormDB, err := gorm.Open(mysql.New(mysql.Config{
		Conn:                      sqlDB,
		SkipInitializeWithVersion: true,
	}), &gorm.Config{})
	if err != nil {
		t.Fatalf("创建 gorm 连接失败: %v", err)
	}

	database.DB = gormDB

	handler := NewDependencyHandler(logger)
	router := gin.New()

	return handler, router, mock
}

func TestDependencyHandler_FixDependencies(t *testing.T) {
	t.Run("无依赖关系", func(t *testing.T) {
		handler, router, mock := setupDependencyHandlerTest(t)
		router.POST("/tasks/dependencies/fix", handler.FixDependencies)

		mock.ExpectQuery("SELECT").
			WillReturnRows(sqlmock.NewRows([]string{"id", "task_id", "depends_on_task_id"}))

		req := httptest.NewRequest(http.MethodPost, "/tasks/dependencies/fix", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("期望状态码 %d, 实际 %d", http.StatusOK, w.Code)
		}
	})

	t.Run("修复无效依赖", func(t *testing.T) {
		handler, router, mock := setupDependencyHandlerTest(t)
		router.POST("/tasks/dependencies/fix", handler.FixDependencies)

		// 依赖关系数据
		rows := sqlmock.NewRows([]string{"id", "task_id", "depends_on_task_id", "dependency_type"}).
			AddRow(1, 1, 999, "finish_to_start"). // 依赖的任务不存在
			AddRow(2, 2, 2, "finish_to_start")    // 自引用

		mock.ExpectQuery("SELECT").
			WillReturnRows(rows)

		// 检查任务 1 是否存在 -> 存在
		mock.ExpectQuery("SELECT").
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
		// 检查依赖任务 999 是否存在 -> 不存在
		mock.ExpectQuery("SELECT").
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
		// 删除依赖
		mock.ExpectBegin()
		mock.ExpectExec("DELETE").
			WillReturnResult(sqlmock.NewResult(0, 1))
		mock.ExpectCommit()

		// 检查任务 2 是否存在 -> 存在
		mock.ExpectQuery("SELECT").
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
		// 检查依赖任务 2 是否存在 -> 存在（但自引用）
		mock.ExpectQuery("SELECT").
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
		// 删除自引用依赖
		mock.ExpectBegin()
		mock.ExpectExec("DELETE").
			WillReturnResult(sqlmock.NewResult(0, 1))
		mock.ExpectCommit()

		req := httptest.NewRequest(http.MethodPost, "/tasks/dependencies/fix", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("期望状态码 %d, 实际 %d", http.StatusOK, w.Code)
		}
	})
}

func TestDependencyHandler_GetDependencyGraph(t *testing.T) {
	t.Run("获取依赖图成功", func(t *testing.T) {
		handler, router, mock := setupDependencyHandlerTest(t)
		router.GET("/tasks/dependencies/graph", handler.GetDependencyGraph)

		// 任务数据
		taskRows := sqlmock.NewRows([]string{"id", "title", "status", "requirement_id"}).
			AddRow(1, "任务1", "pending", 1).
			AddRow(2, "任务2", "pending", 1)

		mock.ExpectQuery("SELECT").
			WillReturnRows(taskRows)

		// 依赖关系数据
		depRows := sqlmock.NewRows([]string{"id", "task_id", "depends_on_task_id"}).
			AddRow(1, 2, 1)

		mock.ExpectQuery("SELECT").
			WillReturnRows(depRows)

		req := httptest.NewRequest(http.MethodGet, "/tasks/dependencies/graph", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("期望状态码 %d, 实际 %d", http.StatusOK, w.Code)
		}
	})

	t.Run("按需求过滤", func(t *testing.T) {
		handler, router, mock := setupDependencyHandlerTest(t)
		router.GET("/tasks/dependencies/graph", handler.GetDependencyGraph)

		// 任务数据
		taskRows := sqlmock.NewRows([]string{"id", "title", "status", "requirement_id"}).
			AddRow(1, "任务1", "pending", 1)

		mock.ExpectQuery("SELECT").
			WillReturnRows(taskRows)

		// 依赖关系数据
		mock.ExpectQuery("SELECT").
			WillReturnRows(sqlmock.NewRows([]string{"id", "task_id", "depends_on_task_id"}))

		req := httptest.NewRequest(http.MethodGet, "/tasks/dependencies/graph?requirementId=1", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("期望状态码 %d, 实际 %d", http.StatusOK, w.Code)
		}
	})
}

func TestDependencyHandler_GetNextTasks(t *testing.T) {
	t.Run("获取可执行任务", func(t *testing.T) {
		handler, router, mock := setupDependencyHandlerTest(t)
		router.GET("/tasks/next", handler.GetNextTasks)

		// 待处理任务
		pendingRows := sqlmock.NewRows([]string{"id", "title", "status", "priority", "requirement_id"}).
			AddRow(1, "任务1", "pending", "high", 1).
			AddRow(2, "任务2", "pending", "medium", 1)

		mock.ExpectQuery("SELECT").
			WillReturnRows(pendingRows)

		// 依赖关系
		mock.ExpectQuery("SELECT").
			WillReturnRows(sqlmock.NewRows([]string{"id", "task_id", "depends_on_task_id"}))

		// 已完成任务
		mock.ExpectQuery("SELECT").
			WillReturnRows(sqlmock.NewRows([]string{"id", "title", "status"}))

		req := httptest.NewRequest(http.MethodGet, "/tasks/next", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("期望状态码 %d, 实际 %d", http.StatusOK, w.Code)
		}
	})

	t.Run("带依赖的任务", func(t *testing.T) {
		handler, router, mock := setupDependencyHandlerTest(t)
		router.GET("/tasks/next", handler.GetNextTasks)

		// 待处理任务
		pendingRows := sqlmock.NewRows([]string{"id", "title", "status", "priority", "requirement_id"}).
			AddRow(1, "任务1", "pending", "high", 1).
			AddRow(2, "任务2", "pending", "medium", 1)

		mock.ExpectQuery("SELECT").
			WillReturnRows(pendingRows)

		// 依赖关系：任务2 依赖任务1
		depRows := sqlmock.NewRows([]string{"id", "task_id", "depends_on_task_id"}).
			AddRow(1, 2, 1)

		mock.ExpectQuery("SELECT").
			WillReturnRows(depRows)

		// 已完成任务
		mock.ExpectQuery("SELECT").
			WillReturnRows(sqlmock.NewRows([]string{"id", "title", "status"}))

		req := httptest.NewRequest(http.MethodGet, "/tasks/next", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("期望状态码 %d, 实际 %d", http.StatusOK, w.Code)
		}
	})
}

// Mock ComplexityService for testing
type MockComplexityService struct{}

func (m *MockComplexityService) AnalyzeTask(task *models.Task, aiSvc interface{}) (*models.ComplexityAnalysis, error) {
	return &models.ComplexityAnalysis{
		TaskID:          int(task.ID),
		TaskTitle:       task.Title,
		ComplexityScore: 5,
		ComplexityLevel: "medium",
	}, nil
}

func init() {
	gin.SetMode(gin.TestMode)
}
