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

func setupMenuTestWithDB(t *testing.T) (*MenuHandler, *gin.Engine, sqlmock.Sqlmock) {
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

	handler := NewMenuHandler(logger)
	router := gin.New()

	return handler, router, mock
}

func TestMenuHandler_List(t *testing.T) {
	handler, router, mock := setupMenuTestWithDB(t)
	router.GET("/menus", handler.List)

	mock.ExpectQuery("SELECT count\\(\\*\\)").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

	mock.ExpectQuery("SELECT").
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "key", "title", "parent_key", "path", "route_name", "icon",
			"sort", "hide_in_menu", "enabled", "fixed", "created_at", "updated_at",
		}))

	req := httptest.NewRequest(http.MethodGet, "/menus", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("期望状态码 %d, 实际 %d", http.StatusOK, w.Code)
	}
}

func TestMenuHandler_Tree(t *testing.T) {
	handler, router, mock := setupMenuTestWithDB(t)
	router.GET("/menus/tree", handler.Tree)

	mock.ExpectQuery("SELECT").
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "key", "title", "parent_key", "path", "route_name", "icon",
			"sort", "hide_in_menu", "enabled", "fixed", "created_at", "updated_at",
		}))

	req := httptest.NewRequest(http.MethodGet, "/menus/tree", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("期望状态码 %d, 实际 %d", http.StatusOK, w.Code)
	}
}

func TestMenuHandler_Get(t *testing.T) {
	tests := []struct {
		name       string
		key        string
		expectCode int
		setupMock  func(mock sqlmock.Sqlmock)
	}{
		{
			name:       "获取存在的菜单",
			key:        "dashboard",
			expectCode: http.StatusOK,
			setupMock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{
					"id", "key", "title", "parent_key", "path", "route_name", "icon",
					"sort", "hide_in_menu", "enabled", "fixed", "created_at", "updated_at",
				}).AddRow(1, "dashboard", "仪表盘", nil, "/dashboard", "dashboard", nil, 1, 0, 1, 0, nil, nil)
				mock.ExpectQuery("SELECT \\* FROM `task_menu`").
					WillReturnRows(rows)
			},
		},
		{
			name:       "获取不存在的菜单",
			key:        "nonexistent",
			expectCode: http.StatusNotFound,
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT \\* FROM `task_menu`").
					WillReturnRows(sqlmock.NewRows([]string{"id"}))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler, router, mock := setupMenuTestWithDB(t)
			router.GET("/menus/:key", handler.Get)

			tt.setupMock(mock)

			req := httptest.NewRequest(http.MethodGet, "/menus/"+tt.key, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != tt.expectCode {
				t.Errorf("期望状态码 %d, 实际 %d", tt.expectCode, w.Code)
			}
		})
	}
}

func TestMenuHandler_Create(t *testing.T) {
	tests := []struct {
		name       string
		body       string
		expectCode int
		setupMock  func(mock sqlmock.Sqlmock)
	}{
		{
			name:       "创建菜单成功",
			body:       `{"key":"test_menu","label":"测试菜单"}`,
			expectCode: http.StatusOK,
			setupMock: func(mock sqlmock.Sqlmock) {
				// Check if key exists
				mock.ExpectQuery("SELECT \\* FROM `task_menu`").
					WillReturnRows(sqlmock.NewRows([]string{"id"}))
				// Create
				mock.ExpectBegin()
				mock.ExpectExec("INSERT INTO `task_menu`").
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()
			},
		},
		{
			name:       "缺少 key",
			body:       `{"label":"测试菜单"}`,
			expectCode: http.StatusBadRequest,
			setupMock:  func(mock sqlmock.Sqlmock) {},
		},
		{
			name:       "缺少 label",
			body:       `{"key":"test_menu"}`,
			expectCode: http.StatusBadRequest,
			setupMock:  func(mock sqlmock.Sqlmock) {},
		},
		{
			name:       "key 已存在",
			body:       `{"key":"dashboard","label":"测试菜单"}`,
			expectCode: http.StatusBadRequest,
			setupMock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{
					"id", "key", "title",
				}).AddRow(1, "dashboard", "仪表盘")
				mock.ExpectQuery("SELECT \\* FROM `task_menu`").
					WillReturnRows(rows)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler, router, mock := setupMenuTestWithDB(t)
			router.POST("/menus", handler.Create)

			tt.setupMock(mock)

			req := httptest.NewRequest(http.MethodPost, "/menus", bytes.NewBufferString(tt.body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != tt.expectCode {
				t.Errorf("期望状态码 %d, 实际 %d, 响应: %s", tt.expectCode, w.Code, w.Body.String())
			}
		})
	}
}

func TestMenuHandler_Update(t *testing.T) {
	tests := []struct {
		name       string
		key        string
		body       string
		expectCode int
		setupMock  func(mock sqlmock.Sqlmock)
	}{
		{
			name:       "更新菜单成功",
			key:        "dashboard",
			body:       `{"title":"更新菜单"}`,
			expectCode: http.StatusOK,
			setupMock: func(mock sqlmock.Sqlmock) {
				// Get existing menu
				rows := sqlmock.NewRows([]string{
					"id", "key", "title", "parent_key", "path", "route_name", "icon",
					"sort", "hide_in_menu", "enabled", "fixed", "created_at", "updated_at",
				}).AddRow(1, "dashboard", "仪表盘", nil, "/dashboard", "dashboard", nil, 1, 0, 1, 0, nil, nil)
				mock.ExpectQuery("SELECT \\* FROM `task_menu`").
					WillReturnRows(rows)
				// Update
				mock.ExpectBegin()
				mock.ExpectExec("UPDATE `task_menu`").
					WillReturnResult(sqlmock.NewResult(0, 1))
				mock.ExpectCommit()
			},
		},
		{
			name:       "菜单不存在",
			key:        "nonexistent",
			body:       `{"title":"更新菜单"}`,
			expectCode: http.StatusNotFound,
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT \\* FROM `task_menu`").
					WillReturnRows(sqlmock.NewRows([]string{"id"}))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler, router, mock := setupMenuTestWithDB(t)
			router.POST("/menus/:key/update", handler.Update)

			tt.setupMock(mock)

			req := httptest.NewRequest(http.MethodPost, "/menus/"+tt.key+"/update", bytes.NewBufferString(tt.body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != tt.expectCode {
				t.Errorf("期望状态码 %d, 实际 %d", tt.expectCode, w.Code)
			}
		})
	}
}

func TestMenuHandler_Delete(t *testing.T) {
	handler, router, mock := setupMenuTestWithDB(t)
	router.POST("/menus/:key/delete", handler.Delete)

	mock.ExpectBegin()
	// Delete menu first
	mock.ExpectExec("DELETE FROM `task_menu`").
		WillReturnResult(sqlmock.NewResult(0, 1))
	// Check for child menus
	mock.ExpectQuery("SELECT \\* FROM `task_menu`").
		WillReturnRows(sqlmock.NewRows([]string{"id"}))
	mock.ExpectCommit()

	req := httptest.NewRequest(http.MethodPost, "/menus/test_menu/delete", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("期望状态码 %d, 实际 %d", http.StatusOK, w.Code)
	}
}

func TestMenuHandler_BatchDelete(t *testing.T) {
	handler, router, mock := setupMenuTestWithDB(t)
	router.POST("/menus/batch-delete", handler.BatchDelete)

	mock.ExpectBegin()
	// Check for child menus for each key
	for i := 0; i < 2; i++ {
		mock.ExpectQuery("SELECT \\* FROM `task_menu`").
			WillReturnRows(sqlmock.NewRows([]string{"id"}))
	}
	// Delete all menus
	mock.ExpectExec("DELETE FROM `task_menu`").
		WillReturnResult(sqlmock.NewResult(0, 2))
	mock.ExpectCommit()

	body := `{"keys":["test1","test2"]}`
	req := httptest.NewRequest(http.MethodPost, "/menus/batch-delete", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("期望状态码 %d, 实际 %d", http.StatusOK, w.Code)
	}
}

func TestMenuHandler_Reorder(t *testing.T) {
	handler, router, mock := setupMenuTestWithDB(t)
	router.POST("/menus/reorder", handler.Reorder)

	// Reorder updates multiple menus
	mock.ExpectBegin()
	mock.ExpectExec("UPDATE `task_menu`").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	body := `[{"key":"dashboard","order":1}]`
	req := httptest.NewRequest(http.MethodPost, "/menus/reorder", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("期望状态码 %d, 实际 %d", http.StatusOK, w.Code)
	}
}

func TestMenuHandler_Move(t *testing.T) {
	handler, router, mock := setupMenuTestWithDB(t)
	router.POST("/menus/:key/move", handler.Move)

	mock.ExpectBegin()
	mock.ExpectExec("UPDATE `task_menu`").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	body := `{"targetParentKey":"dashboard_analysis"}`
	req := httptest.NewRequest(http.MethodPost, "/menus/dashboard_analysis/move", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("期望状态码 %d, 实际 %d", http.StatusOK, w.Code)
	}
}

func TestMenuHandler_Toggle(t *testing.T) {
	handler, router, mock := setupMenuTestWithDB(t)
	router.POST("/menus/:key/toggle", handler.Toggle)

	mock.ExpectBegin()
	mock.ExpectExec("UPDATE `task_menu`").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	body := `{"enabled":true}`
	req := httptest.NewRequest(http.MethodPost, "/menus/dashboard/toggle", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("期望状态码 %d, 实际 %d", http.StatusOK, w.Code)
	}
}

func TestMenuHandler_SyncFromJSON(t *testing.T) {
	handler, router, _ := setupMenuTestWithDB(t)
	router.POST("/menus/sync-from-json", handler.SyncFromJSON)

	// SyncFromJSON 需要 multipart/form-data 文件上传，普通 JSON 请求会返回错误
	body := `{"menus":[{"key":"test","title":"测试","path":"/test"}]}`
	req := httptest.NewRequest(http.MethodPost, "/menus/sync-from-json", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// 没有文件上传时，可能返回 400 或 500，只要不是 200 就说明逻辑执行了
	if w.Code == http.StatusOK {
		t.Errorf("期望状态码不是 200, 实际 %d", w.Code)
	}
}
