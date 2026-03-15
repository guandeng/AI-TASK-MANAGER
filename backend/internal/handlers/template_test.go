package handlers

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/ai-task-manager/backend/internal/database"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func setupTemplateTestWithDB(t *testing.T) (*TemplateHandler, *gin.Engine, sqlmock.Sqlmock) {
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

	handler := NewTemplateHandler(logger)
	router := gin.New()

	return handler, router, mock
}

func TestTemplateHandler_ListProjectTemplates(t *testing.T) {
	handler, router, mock := setupTemplateTestWithDB(t)
	router.GET("/project-templates", handler.ListProjectTemplates)

	mock.ExpectQuery("SELECT").
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "name", "description", "category", "is_public", "usage_count", "created_at", "updated_at",
		}).AddRow(1, "任务模板", "描述", "task", true, 0, nil, nil))

	req := httptest.NewRequest(http.MethodGet, "/project-templates", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("期望状态码 %d, 实际 %d", http.StatusOK, w.Code)
	}
}

func TestTemplateHandler_GetProjectTemplate(t *testing.T) {
	tests := []struct {
		name       string
		id         string
		expectCode int
		setupMock  func(mock sqlmock.Sqlmock)
	}{
		{
			name:       "获取模板成功",
			id:         "1",
			expectCode: http.StatusOK,
			setupMock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{
					"id", "name", "description", "category", "is_public", "usage_count", "created_at", "updated_at",
				}).AddRow(1, "任务模板", "描述", "task", true, 0, nil, nil)
				mock.ExpectQuery("SELECT \\* FROM `task_project_template`").
					WillReturnRows(rows)
				mock.ExpectQuery("SELECT").WillReturnRows(sqlmock.NewRows([]string{"id"}))
			},
		},
		{
			name:       "模板不存在",
			id:         "999",
			expectCode: http.StatusNotFound,
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT \\* FROM `task_project_template`").
					WillReturnRows(sqlmock.NewRows([]string{"id"}))
			},
		},
		{
			name:       "无效的ID",
			id:         "abc",
			expectCode: http.StatusBadRequest,
			setupMock:  func(mock sqlmock.Sqlmock) {},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler, router, mock := setupTemplateTestWithDB(t)
			router.GET("/project-templates/:id", handler.GetProjectTemplate)

			tt.setupMock(mock)

			req := httptest.NewRequest(http.MethodGet, "/project-templates/"+tt.id, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != tt.expectCode {
				t.Errorf("期望状态码 %d, 实际 %d", tt.expectCode, w.Code)
			}
		})
	}
}

func TestTemplateHandler_CreateProjectTemplate(t *testing.T) {
	tests := []struct {
		name       string
		body       string
		expectCode int
		setupMock  func(mock sqlmock.Sqlmock)
	}{
		{
			name:       "创建模板成功",
			body:       `{"name":"新模板","description":"描述","category":"task","isPublic":true,"tasks":[]}`,
			expectCode: http.StatusOK,
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec("INSERT INTO `task_project_template`").
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()
				mock.ExpectQuery("SELECT").WillReturnRows(sqlmock.NewRows([]string{"id"}))
			},
		},
		{
			name:       "无效JSON",
			body:       `invalid json`,
			expectCode: http.StatusBadRequest,
			setupMock:  func(mock sqlmock.Sqlmock) {},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler, router, mock := setupTemplateTestWithDB(t)
			router.POST("/project-templates", handler.CreateProjectTemplate)

			tt.setupMock(mock)

			req := httptest.NewRequest(http.MethodPost, "/project-templates", bytes.NewBufferString(tt.body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != tt.expectCode {
				t.Errorf("期望状态码 %d, 实际 %d", tt.expectCode, w.Code)
			}
		})
	}
}

func TestTemplateHandler_DeleteProjectTemplate(t *testing.T) {
	tests := []struct {
		name       string
		id         string
		expectCode int
		setupMock  func(mock sqlmock.Sqlmock)
	}{
		{
			name:       "删除模板成功",
			id:         "1",
			expectCode: http.StatusOK,
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec("DELETE FROM `task_project_template`").
					WillReturnResult(sqlmock.NewResult(0, 1))
				mock.ExpectCommit()
			},
		},
		{
			name:       "无效的ID",
			id:         "abc",
			expectCode: http.StatusBadRequest,
			setupMock:  func(mock sqlmock.Sqlmock) {},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler, router, mock := setupTemplateTestWithDB(t)
			router.DELETE("/project-templates/:id", handler.DeleteProjectTemplate)

			tt.setupMock(mock)

			req := httptest.NewRequest(http.MethodDelete, "/project-templates/"+tt.id, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != tt.expectCode {
				t.Errorf("期望状态码 %d, 实际 %d", tt.expectCode, w.Code)
			}
		})
	}
}

func TestTemplateHandler_ListTaskTemplates(t *testing.T) {
	handler, router, mock := setupTemplateTestWithDB(t)
	router.GET("/task-templates", handler.ListTaskTemplates)

	mock.ExpectQuery("SELECT").
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "name", "description", "title", "task_description", "priority", "usage_count", "created_at", "updated_at",
		}).AddRow(1, "任务模板", "描述", "标题", "任务描述", "medium", 0, nil, nil))

	req := httptest.NewRequest(http.MethodGet, "/task-templates", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("期望状态码 %d, 实际 %d", http.StatusOK, w.Code)
	}
}

func TestTemplateHandler_GetTaskTemplate(t *testing.T) {
	tests := []struct {
		name       string
		id         string
		expectCode int
		setupMock  func(mock sqlmock.Sqlmock)
	}{
		{
			name:       "获取任务模板成功",
			id:         "1",
			expectCode: http.StatusOK,
			setupMock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{
					"id", "name", "description", "title", "task_description", "priority", "usage_count", "created_at", "updated_at",
				}).AddRow(1, "任务模板", "描述", "标题", "任务描述", "medium", 0, nil, nil)
				mock.ExpectQuery("SELECT \\* FROM `task_template`").
					WillReturnRows(rows)
			},
		},
		{
			name:       "模板不存在",
			id:         "999",
			expectCode: http.StatusNotFound,
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT \\* FROM `task_template`").
					WillReturnRows(sqlmock.NewRows([]string{"id"}))
			},
		},
		{
			name:       "无效的ID",
			id:         "abc",
			expectCode: http.StatusBadRequest,
			setupMock:  func(mock sqlmock.Sqlmock) {},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler, router, mock := setupTemplateTestWithDB(t)
			router.GET("/task-templates/:id", handler.GetTaskTemplate)

			tt.setupMock(mock)

			req := httptest.NewRequest(http.MethodGet, "/task-templates/"+tt.id, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != tt.expectCode {
				t.Errorf("期望状态码 %d, 实际 %d", tt.expectCode, w.Code)
			}
		})
	}
}

func TestTemplateHandler_CreateTaskTemplate(t *testing.T) {
	tests := []struct {
		name       string
		body       string
		expectCode int
		setupMock  func(mock sqlmock.Sqlmock)
	}{
		{
			name:       "创建任务模板成功",
			body:       `{"name":"新模板","title":"任务标题"}`,
			expectCode: http.StatusOK,
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec("INSERT INTO `task_template`").
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()
			},
		},
		{
			name:       "无效JSON",
			body:       `invalid json`,
			expectCode: http.StatusBadRequest,
			setupMock:  func(mock sqlmock.Sqlmock) {},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler, router, mock := setupTemplateTestWithDB(t)
			router.POST("/task-templates", handler.CreateTaskTemplate)

			tt.setupMock(mock)

			req := httptest.NewRequest(http.MethodPost, "/task-templates", bytes.NewBufferString(tt.body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != tt.expectCode {
				t.Errorf("期望状态码 %d, 实际 %d", tt.expectCode, w.Code)
			}
		})
	}
}

func TestTemplateHandler_DeleteTaskTemplate(t *testing.T) {
	tests := []struct {
		name       string
		id         string
		expectCode int
		setupMock  func(mock sqlmock.Sqlmock)
	}{
		{
			name:       "删除任务模板成功",
			id:         "1",
			expectCode: http.StatusOK,
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec("DELETE FROM `task_template`").
					WillReturnResult(sqlmock.NewResult(0, 1))
				mock.ExpectCommit()
			},
		},
		{
			name:       "无效的ID",
			id:         "abc",
			expectCode: http.StatusBadRequest,
			setupMock:  func(mock sqlmock.Sqlmock) {},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler, router, mock := setupTemplateTestWithDB(t)
			router.DELETE("/task-templates/:id", handler.DeleteTaskTemplate)

			tt.setupMock(mock)

			req := httptest.NewRequest(http.MethodDelete, "/task-templates/"+tt.id, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != tt.expectCode {
				t.Errorf("期望状态码 %d, 实际 %d", tt.expectCode, w.Code)
			}
		})
	}
}
