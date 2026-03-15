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
	"github.com/ai-task-manager/backend/internal/models"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// setupIntegrationTest 创建集成测试环境
func setupIntegrationTest(t *testing.T) (*TaskHandler, *gin.Engine, sqlmock.Sqlmock) {
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

// TestTaskHandler_List_Integration 列表接口集成测试
func TestTaskHandler_List_Integration(t *testing.T) {
	tests := []struct {
		name           string
		query          string
		mockTasks      []models.Task
		mockTotal      int64
		expectStatus   int
		expectCount    int
		expectTotal    int64
	}{
		{
			name:  "获取所有任务-默认分页",
			query: "",
			mockTasks: []models.Task{
				{ID: 1, Title: "任务1", Description: "描述1", Status: "pending", Priority: "high"},
				{ID: 2, Title: "任务2", Description: "描述2", Status: "done", Priority: "medium"},
			},
			mockTotal:    2,
			expectStatus: http.StatusOK,
			expectCount:  2,
			expectTotal:  2,
		},
		{
			name:  "按状态筛选-pending",
			query: "?status=pending",
			mockTasks: []models.Task{
				{ID: 1, Title: "任务1", Description: "描述1", Status: "pending", Priority: "high"},
			},
			mockTotal:    1,
			expectStatus: http.StatusOK,
			expectCount:  1,
			expectTotal:  1,
		},
		{
			name:  "按优先级筛选-high",
			query: "?priority=high",
			mockTasks: []models.Task{
				{ID: 1, Title: "任务1", Description: "描述1", Status: "pending", Priority: "high"},
			},
			mockTotal:    1,
			expectStatus: http.StatusOK,
			expectCount:  1,
			expectTotal:  1,
		},
		{
			name:  "按关键词筛选",
			query: "?keyword=测试",
			mockTasks: []models.Task{
				{ID: 1, Title: "测试任务", Description: "描述", Status: "pending", Priority: "medium"},
			},
			mockTotal:    1,
			expectStatus: http.StatusOK,
			expectCount:  1,
			expectTotal:  1,
		},
		{
			name:  "自定义分页",
			query: "?page=2&pageSize=5",
			mockTasks: []models.Task{
				{ID: 6, Title: "任务6", Description: "描述", Status: "pending", Priority: "medium"},
			},
			mockTotal:    10,
			expectStatus: http.StatusOK,
			expectCount:  1,
			expectTotal:  10,
		},
		{
			name:         "空列表",
			query:        "",
			mockTasks:    []models.Task{},
			mockTotal:    0,
			expectStatus: http.StatusOK,
			expectCount:  0,
			expectTotal:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler, router, mock := setupIntegrationTest(t)
			router.GET("/tasks", handler.List)

			// Mock COUNT 查询
			mock.ExpectQuery("SELECT count\\(\\*\\)").
				WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(tt.mockTotal))

			// Mock SELECT 查询
			rows := sqlmock.NewRows([]string{
				"id", "requirement_id", "title", "title_trans", "description", "description_trans",
				"status", "priority", "details", "details_trans", "test_strategy", "test_strategy_trans",
				"assignee", "is_expanding", "expand_message_id", "start_date", "due_date", "completed_at",
				"estimated_hours", "actual_hours", "created_at", "updated_at",
				"requirement_title", "subtask_count", "subtask_done_count",
			})

			for _, task := range tt.mockTasks {
				rows.AddRow(
					task.ID, task.RequirementID, task.Title, task.TitleTrans, task.Description, task.DescriptionTrans,
					task.Status, task.Priority, task.Details, task.DetailsTrans, task.TestStrategy, task.TestStrategyTrans,
					task.Assignee, task.IsExpanding, task.ExpandMessageID, task.StartDate, task.DueDate, task.CompletedAt,
					task.EstimatedHours, task.ActualHours, task.CreatedAt, task.UpdatedAt,
					task.RequirementTitle, task.SubtaskCount, task.SubtaskDoneCount,
				)
			}

			mock.ExpectQuery("SELECT").
				WillReturnRows(rows)

			// 执行请求
			req := httptest.NewRequest("GET", "/tasks"+tt.query, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			// 验证响应
			if w.Code != tt.expectStatus {
				t.Errorf("期望状态码 %d, 实际 %d, 响应: %s", tt.expectStatus, w.Code, w.Body.String())
			}

			if tt.expectStatus == http.StatusOK {
				var response map[string]interface{}
				if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
					t.Errorf("解析响应失败: %v", err)
					return
				}

				data, ok := response["data"].(map[string]interface{})
				if !ok {
					t.Errorf("响应格式错误: 缺少 data 字段")
					return
				}

				// list 可能是 nil 或空数组
				list, ok := data["list"].([]interface{})
				if !ok && tt.expectCount > 0 {
					t.Errorf("响应格式错误: 缺少 list 字段")
					return
				}

				actualCount := 0
				if list != nil {
					actualCount = len(list)
				}

				if actualCount != tt.expectCount {
					t.Errorf("期望 %d 条记录, 实际 %d 条", tt.expectCount, actualCount)
				}

				total, _ := data["total"].(float64)
				if int64(total) != tt.expectTotal {
					t.Errorf("期望总数 %d, 实际 %d", tt.expectTotal, int64(total))
				}
			}
		})
	}
}

// TestTaskHandler_Get_Integration 获取详情集成测试
func TestTaskHandler_Get_Integration(t *testing.T) {
	tests := []struct {
		name         string
		taskID       string
		mockTask     *models.Task
		mockSubtasks []models.Subtask
		expectStatus int
	}{
		{
			name:   "获取存在的任务",
			taskID: "1",
			mockTask: &models.Task{
				ID:          1,
				Title:       "测试任务",
				Description: "测试描述",
				Status:      "pending",
				Priority:    "high",
			},
			mockSubtasks: []models.Subtask{
				{ID: 1, TaskID: 1, Title: "子任务1", Status: "pending"},
			},
			expectStatus: http.StatusOK,
		},
		{
			name:         "获取不存在的任务",
			taskID:       "999",
			mockTask:     nil,
			expectStatus: http.StatusNotFound,
		},
		{
			name:         "无效的任务ID",
			taskID:       "abc",
			expectStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler, router, mock := setupIntegrationTest(t)
			router.GET("/tasks/:taskId", handler.Get)

			if tt.mockTask != nil {
				// Mock 主任务查询
				rows := sqlmock.NewRows([]string{
					"id", "requirement_id", "title", "title_trans", "description", "description_trans",
					"status", "priority", "details", "details_trans", "test_strategy", "test_strategy_trans",
					"assignee", "is_expanding", "expand_message_id", "start_date", "due_date", "completed_at",
					"estimated_hours", "actual_hours", "created_at", "updated_at",
				}).AddRow(
					tt.mockTask.ID, tt.mockTask.RequirementID, tt.mockTask.Title, tt.mockTask.TitleTrans,
					tt.mockTask.Description, tt.mockTask.DescriptionTrans, tt.mockTask.Status, tt.mockTask.Priority,
					tt.mockTask.Details, tt.mockTask.DetailsTrans, tt.mockTask.TestStrategy, tt.mockTask.TestStrategyTrans,
					tt.mockTask.Assignee, tt.mockTask.IsExpanding, tt.mockTask.ExpandMessageID, tt.mockTask.StartDate,
					tt.mockTask.DueDate, tt.mockTask.CompletedAt, tt.mockTask.EstimatedHours, tt.mockTask.ActualHours,
					tt.mockTask.CreatedAt, tt.mockTask.UpdatedAt,
				)
				mock.ExpectQuery("SELECT \\* FROM `task_task`").
					WillReturnRows(rows)

				// Mock 子任务查询
				subtaskRows := sqlmock.NewRows([]string{
					"id", "task_id", "title", "title_trans", "description", "description_trans",
					"details", "details_trans", "status", "priority", "sort_order",
					"estimated_hours", "actual_hours", "code_interface", "acceptance_criteria",
					"related_files", "code_hints", "created_at", "updated_at",
				})
				for _, st := range tt.mockSubtasks {
					subtaskRows.AddRow(
						st.ID, st.TaskID, st.Title, st.TitleTrans, st.Description, st.DescriptionTrans,
						st.Details, st.DetailsTrans, st.Status, st.Priority, st.SortOrder,
						st.EstimatedHours, st.ActualHours, st.CodeInterface, st.AcceptanceCriteria,
						st.RelatedFiles, st.CodeHints, st.CreatedAt, st.UpdatedAt,
					)
				}
				mock.ExpectQuery("SELECT \\* FROM `task_subtask`").
					WillReturnRows(subtaskRows)

				// Mock 依赖查询
				mock.ExpectQuery("SELECT \\* FROM `task_dependency`").
					WillReturnRows(sqlmock.NewRows([]string{"id", "task_id", "depends_on_task_id", "created_at"}))

				// Mock 分配查询
				mock.ExpectQuery("SELECT \\* FROM `task_assignment`").
					WillReturnRows(sqlmock.NewRows([]string{"id", "task_id", "member_id", "role", "estimated_hours", "actual_hours", "created_at"}))
			} else if tt.taskID != "abc" {
				// 任务不存在
				mock.ExpectQuery("SELECT \\* FROM `task_task`").
					WillReturnRows(sqlmock.NewRows([]string{"id"}))
			}

			// 执行请求
			req := httptest.NewRequest("GET", "/tasks/"+tt.taskID, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			// 验证状态码
			if w.Code != tt.expectStatus {
				t.Errorf("期望状态码 %d, 实际 %d, 响应: %s", tt.expectStatus, w.Code, w.Body.String())
			}
		})
	}
}

// TestTaskHandler_Update_Integration 更新任务集成测试
func TestTaskHandler_Update_Integration(t *testing.T) {
	tests := []struct {
		name         string
		taskID       string
		requestBody  map[string]interface{}
		mockTask     *models.Task
		expectStatus int
	}{
		{
			name:   "更新任务状态",
			taskID: "1",
			requestBody: map[string]interface{}{
				"status": "done",
			},
			mockTask: &models.Task{
				ID:          1,
				Title:       "测试任务",
				Status:      "pending",
				Priority:    "high",
			},
			expectStatus: http.StatusOK,
		},
		{
			name:   "更新任务标题和描述",
			taskID: "1",
			requestBody: map[string]interface{}{
				"title":       "新标题",
				"description": "新描述",
			},
			mockTask: &models.Task{
				ID:          1,
				Title:       "旧标题",
				Status:      "pending",
				Priority:    "high",
			},
			expectStatus: http.StatusOK,
		},
		{
			name:   "无效的任务ID",
			taskID: "abc",
			requestBody: map[string]interface{}{
				"status": "done",
			},
			expectStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler, router, mock := setupIntegrationTest(t)
			router.POST("/tasks/:taskId/update", handler.Update)

			if tt.mockTask != nil {
				// Mock 查询任务
				rows := sqlmock.NewRows([]string{
					"id", "requirement_id", "title", "description", "status", "priority",
					"details", "test_strategy", "assignee", "is_expanding", "created_at", "updated_at",
				}).AddRow(
					tt.mockTask.ID, tt.mockTask.RequirementID, tt.mockTask.Title, tt.mockTask.Description,
					tt.mockTask.Status, tt.mockTask.Priority, tt.mockTask.Details, tt.mockTask.TestStrategy,
					tt.mockTask.Assignee, tt.mockTask.IsExpanding, tt.mockTask.CreatedAt, tt.mockTask.UpdatedAt,
				)
				mock.ExpectQuery("SELECT \\* FROM `task_task`").
					WillReturnRows(rows)

				// Mock 更新
				mock.ExpectBegin()
				mock.ExpectExec("UPDATE `task_task`").
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()

				// Mock 重新获取
				mock.ExpectQuery("SELECT \\* FROM `task_task`").
					WillReturnRows(rows)
			}

			// 构建请求
			body, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest("POST", "/tasks/"+tt.taskID+"/update", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			// 验证
			if w.Code != tt.expectStatus {
				t.Errorf("期望状态码 %d, 实际 %d, 响应: %s", tt.expectStatus, w.Code, w.Body.String())
			}
		})
	}
}

// TestTaskHandler_BatchDelete_Integration 批量删除集成测试
func TestTaskHandler_BatchDelete_Integration(t *testing.T) {
	tests := []struct {
		name         string
		requestBody  map[string]interface{}
		expectStatus int
	}{
		{
			name: "批量删除多个任务",
			requestBody: map[string]interface{}{
				"ids": []uint64{1, 2, 3},
			},
			expectStatus: http.StatusOK,
		},
		{
			name: "空ID列表",
			requestBody: map[string]interface{}{
				"ids": []uint64{},
			},
			expectStatus: http.StatusBadRequest,
		},
		{
			name:         "缺少ID列表",
			requestBody:  map[string]interface{}{},
			expectStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler, router, mock := setupIntegrationTest(t)
			router.POST("/tasks/batch-delete", handler.BatchDelete)

			if tt.expectStatus == http.StatusOK {
				mock.ExpectBegin()
				mock.ExpectExec("DELETE FROM `task_task`").
					WillReturnResult(sqlmock.NewResult(0, 3))
				mock.ExpectCommit()
			}

			body, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest("POST", "/tasks/batch-delete", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != tt.expectStatus {
				t.Errorf("期望状态码 %d, 实际 %d, 响应: %s", tt.expectStatus, w.Code, w.Body.String())
			}
		})
	}
}

// TestTaskHandler_Dependencies_Integration 依赖管理集成测试
func TestTaskHandler_Dependencies_Integration(t *testing.T) {
	t.Run("添加依赖", func(t *testing.T) {
		handler, router, mock := setupIntegrationTest(t)
		router.POST("/tasks/:taskId/dependencies", handler.AddDependency)

		// Mock 检查任务存在
		mock.ExpectQuery("SELECT \\* FROM `task_task`").
			WillReturnRows(sqlmock.NewRows([]string{"id", "title", "status", "priority", "created_at", "updated_at"}).
				AddRow(1, "任务1", "pending", "high", nil, nil))
		mock.ExpectQuery("SELECT \\* FROM `task_task`").
			WillReturnRows(sqlmock.NewRows([]string{"id", "title", "status", "priority", "created_at", "updated_at"}).
				AddRow(2, "任务2", "pending", "high", nil, nil))

		// Mock 检查重复
		mock.ExpectQuery("SELECT count\\(\\*\\) FROM `task_dependency`").
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

		// Mock 插入
		mock.ExpectBegin()
		mock.ExpectExec("INSERT INTO `task_dependency`").
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		body, _ := json.Marshal(map[string]interface{}{"dependsOnTaskId": 2})
		req := httptest.NewRequest("POST", "/tasks/1/dependencies", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("期望状态码 %d, 实际 %d, 响应: %s", http.StatusOK, w.Code, w.Body.String())
		}
	})

	t.Run("不能依赖自己", func(t *testing.T) {
		handler, router, _ := setupIntegrationTest(t)
		router.POST("/tasks/:taskId/dependencies", handler.AddDependency)

		body, _ := json.Marshal(map[string]interface{}{"dependsOnTaskId": 1})
		req := httptest.NewRequest("POST", "/tasks/1/dependencies", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("期望状态码 %d, 实际 %d", http.StatusBadRequest, w.Code)
		}
	})

	t.Run("删除依赖", func(t *testing.T) {
		handler, router, mock := setupIntegrationTest(t)
		router.DELETE("/tasks/:taskId/dependencies/:dependsOnTaskId", handler.RemoveDependency)

		mock.ExpectBegin()
		mock.ExpectExec("DELETE FROM `task_dependency`").
			WillReturnResult(sqlmock.NewResult(0, 1))
		mock.ExpectCommit()

		req := httptest.NewRequest("DELETE", "/tasks/1/dependencies/2", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("期望状态码 %d, 实际 %d", http.StatusOK, w.Code)
		}
	})

	t.Run("验证依赖-无循环", func(t *testing.T) {
		handler, router, mock := setupIntegrationTest(t)
		router.GET("/tasks/dependencies/validate", handler.ValidateDependencies)

		// Mock 获取所有依赖
		mock.ExpectQuery("SELECT \\* FROM `task_dependency`").
			WillReturnRows(sqlmock.NewRows([]string{"id", "task_id", "depends_on_task_id", "created_at"}).
				AddRow(1, 2, 1, nil).
				AddRow(2, 3, 2, nil))

		req := httptest.NewRequest("GET", "/tasks/dependencies/validate", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("期望状态码 %d, 实际 %d", http.StatusOK, w.Code)
		}

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		data, _ := response["data"].(map[string]interface{})
		if data["valid"] != true {
			t.Error("期望依赖有效")
		}
	})
}

// TestTaskHandler_GetReadyTasks_Integration 获取可开始任务集成测试
func TestTaskHandler_GetReadyTasks_Integration(t *testing.T) {
	handler, router, mock := setupIntegrationTest(t)
	router.GET("/tasks/ready", handler.GetReadyTasks)

	// Mock pending 任务
	mock.ExpectQuery("SELECT \\* FROM `task_task`").
		WillReturnRows(sqlmock.NewRows([]string{"id", "title", "status", "priority", "created_at", "updated_at"}).
			AddRow(1, "任务1", "pending", "high", nil, nil).
			AddRow(2, "任务2", "pending", "medium", nil, nil).
			AddRow(3, "任务3", "pending", "low", nil, nil))

	// Mock 依赖关系
	mock.ExpectQuery("SELECT \\* FROM `task_dependency`").
		WillReturnRows(sqlmock.NewRows([]string{"id", "task_id", "depends_on_task_id", "created_at"}).
			AddRow(1, 2, 1, nil). // 任务2 依赖任务1
			AddRow(2, 3, 1, nil)) // 任务3 依赖任务1

	// Mock done 任务
	mock.ExpectQuery("SELECT \\* FROM `task_task`").
		WillReturnRows(sqlmock.NewRows([]string{"id", "title", "status", "priority", "created_at", "updated_at"})) // 没有已完成任务

	req := httptest.NewRequest("GET", "/tasks/ready", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("期望状态码 %d, 实际 %d, 响应: %s", http.StatusOK, w.Code, w.Body.String())
	}
}
