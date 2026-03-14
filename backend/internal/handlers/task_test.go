package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/ai-task-manager/backend/internal/config"
	"github.com/ai-task-manager/backend/internal/database"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func setupTaskTest(t *testing.T) (*TaskHandler, *gin.Engine) {
	logger := zap.NewNop()
	cfg := &config.Config{}
	handler := NewTaskHandler(logger, cfg)

	router := gin.New()
	return handler, router
}

func setupTaskTestWithDB(t *testing.T) (*TaskHandler, *gin.Engine, sqlmock.Sqlmock) {
	logger := zap.NewNop()
	cfg := &config.Config{}

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

	handler := NewTaskHandler(logger, cfg)
	router := gin.New()

	return handler, router, mock
}

func TestTaskHandler_List(t *testing.T) {
	handler, router := setupTaskTest(t)
	router.GET("/tasks", handler.List)

	tests := []struct {
		name   string
		query  string
		expect int
	}{
		{
			name:   "无筛选条件",
			query:  "",
			expect: http.StatusOK,
		},
		{
			name:   "按状态筛选",
			query:  "?status=pending",
			expect: http.StatusOK,
		},
		{
			name:   "按优先级筛选",
			query:  "?priority=high",
			expect: http.StatusOK,
		},
		{
			name:   "按负责人筛选",
			query:  "?assignee=john",
			expect: http.StatusOK,
		},
		{
			name:   "按关键词筛选",
			query:  "?keyword=test",
			expect: http.StatusOK,
		},
		{
			name:   "分页参数",
			query:  "?page=1&pageSize=10",
			expect: http.StatusOK,
		},
		{
			name:   "无效分页参数使用默认值",
			query:  "?page=-1&pageSize=200",
			expect: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/tasks"+tt.query, nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			if w.Code != tt.expect {
				t.Errorf("期望状态码 %d, 实际 %d", tt.expect, w.Code)
			}
		})
	}
}

func TestTaskHandler_Get(t *testing.T) {
	handler, router, mock := setupTaskTestWithDB(t)
	router.GET("/tasks/:taskId", handler.Get)

	tests := []struct {
		name      string
		taskID    string
		expect    int
		checkErr  bool
		setupMock func()
	}{
		{
			name:     "无效的任务ID",
			taskID:   "invalid",
			expect:   http.StatusBadRequest,
			checkErr: true,
			setupMock: func() {},
		},
		{
			name:   "任务不存在",
			taskID: "99999",
			expect: http.StatusNotFound,
			setupMock: func() {
				// Mock 查询任务不存在
				mock.ExpectQuery("SELECT \\* FROM `task_task` WHERE").
					WillReturnRows(sqlmock.NewRows([]string{"id"}))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()

			req := httptest.NewRequest(http.MethodGet, "/tasks/"+tt.taskID, nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			if w.Code != tt.expect {
				t.Errorf("期望状态码 %d, 实际 %d", tt.expect, w.Code)
			}

			if tt.checkErr {
				var resp map[string]interface{}
				if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
					t.Fatalf("解析响应失败: %v", err)
				}
				if resp["code"].(float64) == 0 {
					t.Error("期望返回错误码，实际返回成功")
				}
			}
		})
	}
}

func TestTaskHandler_Update(t *testing.T) {
	handler, router := setupTaskTest(t)
	router.PUT("/tasks/:taskId", handler.Update)

	tests := []struct {
		name   string
		taskID string
		body   interface{}
		expect int
	}{
		{
			name:   "无效的任务ID",
			taskID: "invalid",
			body:   map[string]interface{}{"title": "test"},
			expect: http.StatusBadRequest,
		},
		{
			name:   "无效的JSON",
			taskID: "1",
			body:   "invalid json",
			expect: http.StatusBadRequest,
		},
		{
			name:   "有效的更新请求",
			taskID: "1",
			body:   map[string]interface{}{"title": "新标题", "status": "done"},
			expect: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var bodyBytes []byte
			var err error
			if tt.body != nil {
				bodyBytes, err = json.Marshal(tt.body)
				if err != nil {
					t.Fatalf("序列化请求体失败: %v", err)
				}
			}

			req := httptest.NewRequest(http.MethodPut, "/tasks/"+tt.taskID, bytes.NewBuffer(bodyBytes))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			if w.Code != tt.expect {
				t.Errorf("期望状态码 %d, 实际 %d", tt.expect, w.Code)
			}
		})
	}
}

func TestTaskHandler_Delete(t *testing.T) {
	handler, router := setupTaskTest(t)
	router.DELETE("/tasks/:taskId", handler.Delete)

	tests := []struct {
		name   string
		taskID string
		expect int
	}{
		{
			name:   "无效的任务ID",
			taskID: "invalid",
			expect: http.StatusBadRequest,
		},
		{
			name:   "有效的任务ID",
			taskID: "1",
			expect: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodDelete, "/tasks/"+tt.taskID, nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			if w.Code != tt.expect {
				t.Errorf("期望状态码 %d, 实际 %d", tt.expect, w.Code)
			}
		})
	}
}

func TestTaskHandler_BatchDelete(t *testing.T) {
	handler, router := setupTaskTest(t)
	router.DELETE("/tasks/batch", handler.BatchDelete)

	tests := []struct {
		name   string
		body   interface{}
		expect int
	}{
		{
			name:   "无效的JSON",
			body:   "invalid",
			expect: http.StatusBadRequest,
		},
		{
			name:   "空的ID列表",
			body:   map[string]interface{}{"ids": []uint64{}},
			expect: http.StatusBadRequest,
		},
		{
			name:   "有效的批量删除请求",
			body:   map[string]interface{}{"ids": []uint64{1, 2, 3}},
			expect: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bodyBytes, err := json.Marshal(tt.body)
			if err != nil {
				t.Fatalf("序列化请求体失败: %v", err)
			}

			req := httptest.NewRequest(http.MethodDelete, "/tasks/batch", bytes.NewBuffer(bodyBytes))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			if w.Code != tt.expect {
				t.Errorf("期望状态码 %d, 实际 %d", tt.expect, w.Code)
			}
		})
	}
}

func TestTaskHandler_UpdateTime(t *testing.T) {
	handler, router := setupTaskTest(t)
	router.PUT("/tasks/:taskId/time", handler.UpdateTime)

	tests := []struct {
		name   string
		taskID string
		body   interface{}
		expect int
	}{
		{
			name:   "无效的任务ID",
			taskID: "invalid",
			body:   map[string]interface{}{},
			expect: http.StatusBadRequest,
		},
		{
			name:   "有效的更新时间请求",
			taskID: "1",
			body: map[string]interface{}{
				"startDate":      "2024-01-01",
				"dueDate":        "2024-01-31",
				"estimatedHours": 10.5,
				"actualHours":    8.0,
			},
			expect: http.StatusOK,
		},
		{
			name:   "部分更新时间",
			taskID: "1",
			body: map[string]interface{}{
				"estimatedHours": 5.0,
			},
			expect: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bodyBytes, err := json.Marshal(tt.body)
			if err != nil {
				t.Fatalf("序列化请求体失败: %v", err)
			}

			req := httptest.NewRequest(http.MethodPut, "/tasks/"+tt.taskID+"/time", bytes.NewBuffer(bodyBytes))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			if w.Code != tt.expect {
				t.Errorf("期望状态码 %d, 实际 %d", tt.expect, w.Code)
			}
		})
	}
}

func TestTaskHandler_UpdateSubtask(t *testing.T) {
	handler, router := setupTaskTest(t)
	router.PUT("/tasks/:taskId/subtasks/:subtaskId", handler.UpdateSubtask)

	tests := []struct {
		name      string
		taskID    string
		subtaskID string
		body      interface{}
		expect    int
	}{
		{
			name:      "无效的任务ID",
			taskID:    "invalid",
			subtaskID: "1",
			body:      map[string]interface{}{},
			expect:    http.StatusBadRequest,
		},
		{
			name:      "无效的子任务ID",
			taskID:    "1",
			subtaskID: "invalid",
			body:      map[string]interface{}{},
			expect:    http.StatusBadRequest,
		},
		{
			name:      "有效的更新请求",
			taskID:    "1",
			subtaskID: "1",
			body:      map[string]interface{}{"title": "更新子任务", "status": "done"},
			expect:    http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bodyBytes, err := json.Marshal(tt.body)
			if err != nil {
				t.Fatalf("序列化请求体失败: %v", err)
			}

			req := httptest.NewRequest(http.MethodPut, "/tasks/"+tt.taskID+"/subtasks/"+tt.subtaskID, bytes.NewBuffer(bodyBytes))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			if w.Code != tt.expect {
				t.Errorf("期望状态码 %d, 实际 %d", tt.expect, w.Code)
			}
		})
	}
}

func TestTaskHandler_DeleteSubtask(t *testing.T) {
	handler, router := setupTaskTest(t)
	router.DELETE("/tasks/:taskId/subtasks/:subtaskId", handler.DeleteSubtask)

	tests := []struct {
		name      string
		taskID    string
		subtaskID string
		expect    int
	}{
		{
			name:      "无效的任务ID",
			taskID:    "invalid",
			subtaskID: "1",
			expect:    http.StatusBadRequest,
		},
		{
			name:      "无效的子任务ID",
			taskID:    "1",
			subtaskID: "invalid",
			expect:    http.StatusBadRequest,
		},
		{
			name:      "有效的删除请求",
			taskID:    "1",
			subtaskID: "1",
			expect:    http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodDelete, "/tasks/"+tt.taskID+"/subtasks/"+tt.subtaskID, nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			if w.Code != tt.expect {
				t.Errorf("期望状态码 %d, 实际 %d", tt.expect, w.Code)
			}
		})
	}
}

func TestTaskHandler_DeleteAllSubtasks(t *testing.T) {
	handler, router := setupTaskTest(t)
	router.DELETE("/tasks/:taskId/subtasks", handler.DeleteAllSubtasks)

	tests := []struct {
		name   string
		taskID string
		expect int
	}{
		{
			name:   "无效的任务ID",
			taskID: "invalid",
			expect: http.StatusBadRequest,
		},
		{
			name:   "有效的删除请求",
			taskID: "1",
			expect: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodDelete, "/tasks/"+tt.taskID+"/subtasks", nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			if w.Code != tt.expect {
				t.Errorf("期望状态码 %d, 实际 %d", tt.expect, w.Code)
			}
		})
	}
}

func TestTaskHandler_ReorderSubtasks(t *testing.T) {
	handler, router := setupTaskTest(t)
	router.PUT("/tasks/:taskId/subtasks/reorder", handler.ReorderSubtasks)

	tests := []struct {
		name   string
		taskID string
		body   interface{}
		expect int
	}{
		{
			name:   "无效的任务ID",
			taskID: "invalid",
			body:   map[string]interface{}{"subtaskIds": []uint64{1, 2, 3}},
			expect: http.StatusBadRequest,
		},
		{
			name:   "有效的重排序请求",
			taskID: "1",
			body:   map[string]interface{}{"subtaskIds": []uint64{3, 2, 1}},
			expect: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bodyBytes, err := json.Marshal(tt.body)
			if err != nil {
				t.Fatalf("序列化请求体失败: %v", err)
			}

			req := httptest.NewRequest(http.MethodPut, "/tasks/"+tt.taskID+"/subtasks/reorder", bytes.NewBuffer(bodyBytes))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			if w.Code != tt.expect {
				t.Errorf("期望状态码 %d, 实际 %d", tt.expect, w.Code)
			}
		})
	}
}

func TestTaskHandler_RegenerateSubtask(t *testing.T) {
	handler, router := setupTaskTest(t)
	router.POST("/tasks/:taskId/subtasks/:subtaskId/regenerate", handler.RegenerateSubtask)

	tests := []struct {
		name      string
		taskID    string
		subtaskID string
		expect    int
	}{
		{
			name:      "无效的任务ID",
			taskID:    "invalid",
			subtaskID: "1",
			expect:    http.StatusBadRequest,
		},
		{
			name:      "无效的子任务ID",
			taskID:    "1",
			subtaskID: "invalid",
			expect:    http.StatusBadRequest,
		},
		{
			name:      "有效的重新生成请求",
			taskID:    "1",
			subtaskID: "1",
			expect:    http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/tasks/"+tt.taskID+"/subtasks/"+tt.subtaskID+"/regenerate", nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			if w.Code != tt.expect {
				t.Errorf("期望状态码 %d, 实际 %d", tt.expect, w.Code)
			}
		})
	}
}

func TestTaskHandler_ExpandTask(t *testing.T) {
	handler, router := setupTaskTest(t)
	router.POST("/tasks/:taskId/expand", handler.ExpandTask)

	tests := []struct {
		name   string
		taskID string
		expect int
	}{
		{
			name:   "无效的任务ID",
			taskID: "invalid",
			expect: http.StatusBadRequest,
		},
		{
			name:   "有效的展开请求",
			taskID: "1",
			expect: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/tasks/"+tt.taskID+"/expand", nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			if w.Code != tt.expect {
				t.Errorf("期望状态码 %d, 实际 %d", tt.expect, w.Code)
			}
		})
	}
}

func TestTaskHandler_ExpandTaskAsync(t *testing.T) {
	handler, router := setupTaskTest(t)
	router.POST("/tasks/:taskId/expand/async", handler.ExpandTaskAsync)

	tests := []struct {
		name   string
		taskID string
		expect int
	}{
		{
			name:   "无效的任务ID",
			taskID: "invalid",
			expect: http.StatusBadRequest,
		},
		{
			name:   "有效的异步展开请求",
			taskID: "1",
			expect: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/tasks/"+tt.taskID+"/expand/async", nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			if w.Code != tt.expect {
				t.Errorf("期望状态码 %d, 实际 %d", tt.expect, w.Code)
			}
		})
	}
}

func TestTaskHandler_GetAssignments(t *testing.T) {
	handler, router := setupTaskTest(t)
	router.GET("/tasks/assignments", handler.GetAssignments)

	req := httptest.NewRequest(http.MethodGet, "/tasks/assignments", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("期望状态码 %d, 实际 %d", http.StatusOK, w.Code)
	}
}

func TestTaskHandler_GetAssignmentOverview(t *testing.T) {
	handler, router := setupTaskTest(t)
	router.GET("/tasks/assignments/overview", handler.GetAssignmentOverview)

	req := httptest.NewRequest(http.MethodGet, "/tasks/assignments/overview", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("期望状态码 %d, 实际 %d", http.StatusOK, w.Code)
	}
}

func TestTaskHandler_CreateAssignment(t *testing.T) {
	handler, router := setupTaskTest(t)
	router.POST("/tasks/assignments", handler.CreateAssignment)

	req := httptest.NewRequest(http.MethodPost, "/tasks/assignments", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("期望状态码 %d, 实际 %d", http.StatusOK, w.Code)
	}
}

func TestTaskHandler_DeleteAssignment(t *testing.T) {
	handler, router := setupTaskTest(t)
	router.DELETE("/tasks/assignments/:id", handler.DeleteAssignment)

	req := httptest.NewRequest(http.MethodDelete, "/tasks/assignments/1", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("期望状态码 %d, 实际 %d", http.StatusOK, w.Code)
	}
}

func TestTaskHandler_GetSubtaskAssignments(t *testing.T) {
	handler, router := setupTaskTest(t)
	router.GET("/tasks/subtasks/assignments", handler.GetSubtaskAssignments)

	req := httptest.NewRequest(http.MethodGet, "/tasks/subtasks/assignments", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("期望状态码 %d, 实际 %d", http.StatusOK, w.Code)
	}
}

func TestTaskHandler_CreateSubtaskAssignment(t *testing.T) {
	handler, router := setupTaskTest(t)
	router.POST("/tasks/subtasks/assignments", handler.CreateSubtaskAssignment)

	req := httptest.NewRequest(http.MethodPost, "/tasks/subtasks/assignments", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("期望状态码 %d, 实际 %d", http.StatusOK, w.Code)
	}
}

func TestTaskHandler_DeleteSubtaskAssignment(t *testing.T) {
	handler, router := setupTaskTest(t)
	router.DELETE("/tasks/subtasks/assignments/:id", handler.DeleteSubtaskAssignment)

	req := httptest.NewRequest(http.MethodDelete, "/tasks/subtasks/assignments/1", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("期望状态码 %d, 实际 %d", http.StatusOK, w.Code)
	}
}
