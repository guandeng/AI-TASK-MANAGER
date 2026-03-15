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

func setupLanguageTestWithDB(t *testing.T) (*LanguageHandler, *gin.Engine, sqlmock.Sqlmock) {
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

	handler := NewLanguageHandler(logger)
	router := gin.New()

	return handler, router, mock
}

func TestLanguageHandler_List(t *testing.T) {
	handler, router, mock := setupLanguageTestWithDB(t)
	router.GET("/languages", handler.List)

	mock.ExpectQuery("SELECT").
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "name", "display_name", "framework", "description", "code_hints", "remark", "is_active", "sort_order", "created_at", "updated_at",
		}).AddRow(1, "Go", "Go (Gin + GORM)", "Gin + GORM", "Go语言", "", "", true, 1, nil, nil).
			AddRow(2, "Java", "Java (Spring)", "Spring Boot", "Java语言", "", "", true, 2, nil, nil))

	req := httptest.NewRequest(http.MethodGet, "/languages", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("期望状态码 %d, 实际 %d", http.StatusOK, w.Code)
	}
}

func TestLanguageHandler_List_All(t *testing.T) {
	handler, router, mock := setupLanguageTestWithDB(t)
	router.GET("/languages", handler.List)

	mock.ExpectQuery("SELECT").
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "name", "display_name", "framework", "description", "code_hints", "remark", "is_active", "sort_order", "created_at", "updated_at",
		}).AddRow(1, "Go", "Go (Gin + GORM)", "Gin + GORM", "Go语言", "", "", true, 1, nil, nil).
			AddRow(2, "Java", "Java (Spring)", "Spring Boot", "Java语言", "", "", false, 2, nil, nil))

	req := httptest.NewRequest(http.MethodGet, "/languages?all=true", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("期望状态码 %d, 实际 %d", http.StatusOK, w.Code)
	}
}

func TestLanguageHandler_Get(t *testing.T) {
	tests := []struct {
		name       string
		id         string
		expectCode int
		setupMock  func(mock sqlmock.Sqlmock)
	}{
		{
			name:       "获取语言成功",
			id:         "1",
			expectCode: http.StatusOK,
			setupMock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{
					"id", "name", "display_name", "framework", "description", "code_hints", "remark", "is_active", "sort_order", "created_at", "updated_at",
				}).AddRow(1, "Go", "Go (Gin + GORM)", "Gin + GORM", "Go语言", "", "", true, 1, nil, nil)
				mock.ExpectQuery("SELECT \\* FROM `languages`").
					WillReturnRows(rows)
			},
		},
		{
			name:       "语言不存在",
			id:         "999",
			expectCode: http.StatusNotFound,
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT \\* FROM `languages`").
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
			handler, router, mock := setupLanguageTestWithDB(t)
			router.GET("/languages/:id", handler.Get)

			tt.setupMock(mock)

			req := httptest.NewRequest(http.MethodGet, "/languages/"+tt.id, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != tt.expectCode {
				t.Errorf("期望状态码 %d, 实际 %d", tt.expectCode, w.Code)
			}
		})
	}
}

func TestLanguageHandler_Create(t *testing.T) {
	tests := []struct {
		name       string
		body       string
		expectCode int
		setupMock  func(mock sqlmock.Sqlmock)
	}{
		{
			name:       "创建语言成功",
			body:       `{"name":"Rust","displayName":"Rust","isActive":true,"sortOrder":3}`,
			expectCode: http.StatusOK,
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec("INSERT INTO `languages`").
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()
			},
		},
		{
			name:       "名称不能为空",
			body:       `{"displayName":"Rust"}`,
			expectCode: http.StatusBadRequest,
			setupMock:  func(mock sqlmock.Sqlmock) {},
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
			handler, router, mock := setupLanguageTestWithDB(t)
			router.POST("/languages", handler.Create)

			tt.setupMock(mock)

			req := httptest.NewRequest(http.MethodPost, "/languages", bytes.NewBufferString(tt.body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != tt.expectCode {
				t.Errorf("期望状态码 %d, 实际 %d", tt.expectCode, w.Code)
			}
		})
	}
}

func TestLanguageHandler_Update(t *testing.T) {
	tests := []struct {
		name       string
		id         string
		body       string
		expectCode int
		setupMock  func(mock sqlmock.Sqlmock)
	}{
		{
			name:       "更新语言成功",
			id:         "1",
			body:       `{"name":"Go更新"}`,
			expectCode: http.StatusOK,
			setupMock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{
					"id", "name", "display_name", "framework", "description", "code_hints", "remark", "is_active", "sort_order", "created_at", "updated_at",
				}).AddRow(1, "Go", "Go (Gin + GORM)", "Gin + GORM", "Go语言", "", "", true, 1, nil, nil)
				mock.ExpectQuery("SELECT \\* FROM `languages`").
					WillReturnRows(rows)
				mock.ExpectBegin()
				mock.ExpectExec("UPDATE `languages`").
					WillReturnResult(sqlmock.NewResult(0, 1))
				mock.ExpectCommit()
				reloadRows := sqlmock.NewRows([]string{
					"id", "name", "display_name", "framework", "description", "code_hints", "remark", "is_active", "sort_order", "created_at", "updated_at",
				}).AddRow(1, "Go更新", "Go (Gin + GORM)", "Gin + GORM", "Go语言", "", "", true, 1, nil, nil)
				mock.ExpectQuery("SELECT \\* FROM `languages`").
					WillReturnRows(reloadRows)
			},
		},
		{
			name:       "语言不存在",
			id:         "999",
			body:       `{"name":"测试"}`,
			expectCode: http.StatusNotFound,
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT \\* FROM `languages`").
					WillReturnRows(sqlmock.NewRows([]string{"id"}))
			},
		},
		{
			name:       "无效的ID",
			id:         "abc",
			body:       `{"name":"测试"}`,
			expectCode: http.StatusBadRequest,
			setupMock:  func(mock sqlmock.Sqlmock) {},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler, router, mock := setupLanguageTestWithDB(t)
			router.PUT("/languages/:id", handler.Update)

			tt.setupMock(mock)

			req := httptest.NewRequest(http.MethodPut, "/languages/"+tt.id, bytes.NewBufferString(tt.body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != tt.expectCode {
				t.Errorf("期望状态码 %d, 实际 %d", tt.expectCode, w.Code)
			}
		})
	}
}

func TestLanguageHandler_Delete(t *testing.T) {
	tests := []struct {
		name       string
		id         string
		expectCode int
		setupMock  func(mock sqlmock.Sqlmock)
	}{
		{
			name:       "删除语言成功",
			id:         "1",
			expectCode: http.StatusOK,
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec("DELETE FROM `languages`").
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
			handler, router, mock := setupLanguageTestWithDB(t)
			router.DELETE("/languages/:id", handler.Delete)

			tt.setupMock(mock)

			req := httptest.NewRequest(http.MethodDelete, "/languages/"+tt.id, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != tt.expectCode {
				t.Errorf("期望状态码 %d, 实际 %d", tt.expectCode, w.Code)
			}
		})
	}
}
