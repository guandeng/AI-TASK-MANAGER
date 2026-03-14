package handlers

import (
	"bytes"
	"encoding/json"
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

func setupRequirementTestWithDB(t *testing.T) (*RequirementHandler, *gin.Engine, sqlmock.Sqlmock) {
	logger := zap.NewNop()
	handler := NewRequirementHandler(logger)

	// 创建 mock 数据库
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

	// 设置全局 DB
	database.DB = gormDB

	router := gin.New()
	return handler, router, mock
}

func TestRequirementHandler_List(t *testing.T) {
	handler, router, mock := setupRequirementTestWithDB(t)
	router.GET("/requirements", handler.List)

	// Mock count query
	rows := sqlmock.NewRows([]string{"count"}).AddRow(0)
	mock.ExpectQuery("SELECT count\\(\\*\\) FROM `task_requirement`").
		WillReturnRows(rows)

	// Mock select query
	emptyRows := sqlmock.NewRows([]string{"id", "created_at", "updated_at", "deleted_at", "title", "content", "status", "priority", "assignee"})
	mock.ExpectQuery("SELECT \\* FROM `task_requirement`").
		WillReturnRows(emptyRows)

	req := httptest.NewRequest(http.MethodGet, "/requirements", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("期望状态码 %d, 实际 %d", http.StatusOK, w.Code)
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("解析响应失败: %v", err)
	}

	if resp["code"].(float64) != 0 {
		t.Errorf("期望 code 为 0, 实际 %v", resp["code"])
	}
}

func TestRequirementHandler_Statistics(t *testing.T) {
	handler, router, mock := setupRequirementTestWithDB(t)
	router.GET("/requirements/statistics", handler.Statistics)

	// Mock count queries (total + 4 status counts)
	for i := 0; i < 5; i++ {
		rows := sqlmock.NewRows([]string{"count"}).AddRow(0)
		mock.ExpectQuery("SELECT count\\(\\*\\) FROM `task_requirement`").
			WillReturnRows(rows)
	}

	req := httptest.NewRequest(http.MethodGet, "/requirements/statistics", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("期望状态码 %d, 实际 %d", http.StatusOK, w.Code)
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("解析响应失败: %v", err)
	}

	data := resp["data"].(map[string]interface{})
	expectedKeys := []string{"total", "draft", "reviewing", "approved", "completed"}
	for _, key := range expectedKeys {
		if _, ok := data[key]; !ok {
			t.Errorf("期望 data 包含 %s", key)
		}
	}
}

func TestRequirementHandler_Get(t *testing.T) {
	handler, router, mock := setupRequirementTestWithDB(t)
	router.GET("/requirements/:id", handler.Get)

	// Mock 查询需求不存在
	mock.ExpectQuery("SELECT \\* FROM `task_requirement` WHERE").
		WillReturnRows(sqlmock.NewRows([]string{"id"}))

	req := httptest.NewRequest(http.MethodGet, "/requirements/1", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// 需求不存在应返回 404
	if w.Code != http.StatusNotFound {
		t.Errorf("期望状态码 %d, 实际 %d", http.StatusNotFound, w.Code)
	}
}

func TestRequirementHandler_Get_InvalidID(t *testing.T) {
	handler, router, _ := setupRequirementTestWithDB(t)
	router.GET("/requirements/:id", handler.Get)

	req := httptest.NewRequest(http.MethodGet, "/requirements/invalid", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("期望状态码 %d, 实际 %d", http.StatusBadRequest, w.Code)
	}
}

func TestRequirementHandler_Create(t *testing.T) {
	handler, router, mock := setupRequirementTestWithDB(t)
	router.POST("/requirements", handler.Create)

	// Mock 创建需求
	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO `task_requirement`").
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	body := map[string]interface{}{
		"title":   "测试需求",
		"content": "需求内容",
	}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/requirements", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("期望状态码 %d, 实际 %d", http.StatusOK, w.Code)
	}
}

func TestRequirementHandler_Create_InvalidJSON(t *testing.T) {
	handler, router, _ := setupRequirementTestWithDB(t)
	router.POST("/requirements", handler.Create)

	req := httptest.NewRequest(http.MethodPost, "/requirements", bytes.NewBuffer([]byte("invalid")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("期望状态码 %d, 实际 %d", http.StatusBadRequest, w.Code)
	}
}

func TestRequirementHandler_Update(t *testing.T) {
	handler, router, mock := setupRequirementTestWithDB(t)
	router.PUT("/requirements/:id", handler.Update)

	// Mock 查询需求不存在
	mock.ExpectQuery("SELECT \\* FROM `task_requirement` WHERE").
		WillReturnRows(sqlmock.NewRows([]string{"id"}))

	body := map[string]interface{}{
		"title": "更新标题",
	}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPut, "/requirements/1", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// 需求不存在应返回 404
	if w.Code != http.StatusNotFound {
		t.Errorf("期望状态码 %d, 实际 %d", http.StatusNotFound, w.Code)
	}
}

func TestRequirementHandler_Update_InvalidID(t *testing.T) {
	handler, router, _ := setupRequirementTestWithDB(t)
	router.PUT("/requirements/:id", handler.Update)

	req := httptest.NewRequest(http.MethodPut, "/requirements/invalid", nil)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("期望状态码 %d, 实际 %d", http.StatusBadRequest, w.Code)
	}
}

func TestRequirementHandler_Delete(t *testing.T) {
	handler, router, mock := setupRequirementTestWithDB(t)
	router.DELETE("/requirements/:id", handler.Delete)

	// Mock 删除
	mock.ExpectBegin()
	mock.ExpectExec("DELETE FROM `task_requirement`").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	req := httptest.NewRequest(http.MethodDelete, "/requirements/1", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("期望状态码 %d, 实际 %d", http.StatusOK, w.Code)
	}
}

func TestRequirementHandler_Delete_InvalidID(t *testing.T) {
	handler, router, _ := setupRequirementTestWithDB(t)
	router.DELETE("/requirements/:id", handler.Delete)

	req := httptest.NewRequest(http.MethodDelete, "/requirements/invalid", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("期望状态码 %d, 实际 %d", http.StatusBadRequest, w.Code)
	}
}

func TestRequirementHandler_UploadDocument(t *testing.T) {
	handler, router, _ := setupRequirementTestWithDB(t)
	router.POST("/requirements/:id/documents", handler.UploadDocument)

	req := httptest.NewRequest(http.MethodPost, "/requirements/1/documents", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("期望状态码 %d, 实际 %d", http.StatusOK, w.Code)
	}
}

func TestRequirementHandler_DeleteDocument(t *testing.T) {
	handler, router, mock := setupRequirementTestWithDB(t)
	router.DELETE("/requirements/:id/documents/:docId", handler.DeleteDocument)

	// Mock 删除
	mock.ExpectBegin()
	mock.ExpectExec("DELETE FROM `task_requirement_document`").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	req := httptest.NewRequest(http.MethodDelete, "/requirements/1/documents/1", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("期望状态码 %d, 实际 %d", http.StatusOK, w.Code)
	}
}

func TestRequirementHandler_DeleteDocument_InvalidIDs(t *testing.T) {
	handler, router, _ := setupRequirementTestWithDB(t)
	router.DELETE("/requirements/:id/documents/:docId", handler.DeleteDocument)

	tests := []struct {
		name   string
		id     string
		docId  string
		expect int
	}{
		{
			name:   "无效的需求ID",
			id:     "invalid",
			docId:  "1",
			expect: http.StatusBadRequest,
		},
		{
			name:   "无效的文档ID",
			id:     "1",
			docId:  "invalid",
			expect: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodDelete, "/requirements/"+tt.id+"/documents/"+tt.docId, nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			if w.Code != tt.expect {
				t.Errorf("期望状态码 %d, 实际 %d", tt.expect, w.Code)
			}
		})
	}
}

func TestRequirementHandler_SplitTasks(t *testing.T) {
	handler, router, _ := setupRequirementTestWithDB(t)
	router.POST("/requirements/:id/split", handler.SplitTasks)

	req := httptest.NewRequest(http.MethodPost, "/requirements/1/split", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("期望状态码 %d, 实际 %d", http.StatusOK, w.Code)
	}
}
