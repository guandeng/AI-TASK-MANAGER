package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/ai-task-manager/backend/internal/database"
	"github.com/ai-task-manager/backend/internal/models"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func strPtr(s string) *string {
	return &s
}

func setupMemberTestWithDB(t *testing.T) (*MemberHandler, *gin.Engine, sqlmock.Sqlmock) {
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

	handler := NewMemberHandler(logger)
	router := gin.New()

	return handler, router, mock
}

func TestMemberHandler_List(t *testing.T) {
	tests := []struct {
		name        string
		query       string
		mockCount   int64
		mockMembers []models.Member
		expectCode  int
	}{
		{
			name:      "获取所有成员",
			query:     "",
			mockCount: 2,
			mockMembers: []models.Member{
				{ID: 1, Name: "张三", Email: strPtr("zhangsan@example.com"), Role: "admin", Status: "active"},
				{ID: 2, Name: "李四", Email: strPtr("lisi@example.com"), Role: "member", Status: "active"},
			},
			expectCode: http.StatusOK,
		},
		{
			name:        "空列表",
			query:       "",
			mockCount:   0,
			mockMembers: []models.Member{},
			expectCode:  http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler, router, mock := setupMemberTestWithDB(t)
			router.GET("/members", handler.List)

			mock.ExpectQuery("SELECT count\\(\\*\\)").
				WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(tt.mockCount))

			rows := sqlmock.NewRows([]string{
				"id", "name", "email", "avatar", "role", "department", "skills", "status", "created_at", "updated_at",
			})
			for _, m := range tt.mockMembers {
				rows.AddRow(m.ID, m.Name, m.Email, m.Avatar, m.Role, m.Department, m.Skills, m.Status, m.CreatedAt, m.UpdatedAt)
			}
			mock.ExpectQuery("SELECT").
				WillReturnRows(rows)

			req := httptest.NewRequest(http.MethodGet, "/members"+tt.query, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != tt.expectCode {
				t.Errorf("期望状态码 %d, 实际 %d", tt.expectCode, w.Code)
			}
		})
	}
}

func TestMemberHandler_Get(t *testing.T) {
	tests := []struct {
		name       string
		memberID   string
		mockMember *models.Member
		expectCode int
	}{
		{
			name:     "获取存在的成员",
			memberID: "1",
			mockMember: &models.Member{
				ID:     1,
				Name:   "张三",
				Email:  strPtr("zhangsan@example.com"),
				Role:   "admin",
				Status: "active",
			},
			expectCode: http.StatusOK,
		},
		{
			name:       "获取不存在的成员",
			memberID:   "999",
			mockMember: nil,
			expectCode: http.StatusNotFound,
		},
		{
			name:       "无效的成员ID",
			memberID:   "abc",
			expectCode: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler, router, mock := setupMemberTestWithDB(t)
			router.GET("/members/:id", handler.Get)

			if tt.mockMember != nil {
				rows := sqlmock.NewRows([]string{
					"id", "name", "email", "avatar", "role", "department", "skills", "status", "created_at", "updated_at",
				}).AddRow(
					tt.mockMember.ID, tt.mockMember.Name, tt.mockMember.Email, tt.mockMember.Avatar,
					tt.mockMember.Role, tt.mockMember.Department, tt.mockMember.Skills, tt.mockMember.Status,
					tt.mockMember.CreatedAt, tt.mockMember.UpdatedAt,
				)
				mock.ExpectQuery("SELECT \\* FROM `task_member`").
					WillReturnRows(rows)
			} else if tt.memberID != "abc" {
				mock.ExpectQuery("SELECT \\* FROM `task_member`").
					WillReturnRows(sqlmock.NewRows([]string{"id"}))
			}

			req := httptest.NewRequest(http.MethodGet, "/members/"+tt.memberID, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != tt.expectCode {
				t.Errorf("期望状态码 %d, 实际 %d", tt.expectCode, w.Code)
			}
		})
	}
}

func TestMemberHandler_Create(t *testing.T) {
	tests := []struct {
		name       string
		body       string
		expectCode int
		setupMock  func(mock sqlmock.Sqlmock)
	}{
		{
			name:       "创建成员成功",
			body:       `{"name":"王五","email":"wangwu@example.com","role":"member"}`,
			expectCode: http.StatusOK,
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec("INSERT INTO `task_member`").
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()
			},
		},
		{
			name:       "只有必填字段也能创建",
			body:       `{"name":"王五"}`,
			expectCode: http.StatusOK,
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec("INSERT INTO `task_member`").
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
			handler, router, mock := setupMemberTestWithDB(t)
			router.POST("/members", handler.Create)

			tt.setupMock(mock)

			req := httptest.NewRequest(http.MethodPost, "/members", strings.NewReader(tt.body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != tt.expectCode {
				t.Errorf("期望状态码 %d, 实际 %d, 响应: %s", tt.expectCode, w.Code, w.Body.String())
			}
		})
	}
}

func TestMemberHandler_Update(t *testing.T) {
	tests := []struct {
		name       string
		memberID   string
		body       string
		expectCode int
		setupMock  func(mock sqlmock.Sqlmock)
	}{
		{
			name:       "更新成员成功",
			memberID:   "1",
			body:       `{"name":"张三更新","email":"zhangsan_new@example.com"}`,
			expectCode: http.StatusOK,
			setupMock: func(mock sqlmock.Sqlmock) {
				// First query to get member
				rows := sqlmock.NewRows([]string{
					"id", "name", "email", "avatar", "role", "department", "skills", "status", "created_at", "updated_at",
				}).AddRow(1, "张三", "zhangsan@example.com", nil, "admin", nil, nil, "active", nil, nil)
				mock.ExpectQuery("SELECT \\* FROM `task_member`").
					WillReturnRows(rows)

				// Update
				mock.ExpectBegin()
				mock.ExpectExec("UPDATE `task_member`").
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()

				// Second query to get updated member
				rows2 := sqlmock.NewRows([]string{
					"id", "name", "email", "avatar", "role", "department", "skills", "status", "created_at", "updated_at",
				}).AddRow(1, "张三更新", "zhangsan_new@example.com", nil, "admin", nil, nil, "active", nil, nil)
				mock.ExpectQuery("SELECT \\* FROM `task_member`").
					WillReturnRows(rows2)
			},
		},
		{
			name:       "无效的成员ID",
			memberID:   "abc",
			body:       `{"name":"测试"}`,
			expectCode: http.StatusBadRequest,
			setupMock:  func(mock sqlmock.Sqlmock) {},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler, router, mock := setupMemberTestWithDB(t)
			router.PUT("/members/:id", handler.Update)

			tt.setupMock(mock)

			req := httptest.NewRequest(http.MethodPut, "/members/"+tt.memberID, strings.NewReader(tt.body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != tt.expectCode {
				t.Errorf("期望状态码 %d, 实际 %d, 响应: %s", tt.expectCode, w.Code, w.Body.String())
			}
		})
	}
}

func TestMemberHandler_Delete(t *testing.T) {
	tests := []struct {
		name       string
		memberID   string
		expectCode int
		setupMock  func(mock sqlmock.Sqlmock)
	}{
		{
			name:       "删除成员成功",
			memberID:   "1",
			expectCode: http.StatusOK,
			setupMock: func(mock sqlmock.Sqlmock) {
				// First query to check member exists
				rows := sqlmock.NewRows([]string{
					"id", "name", "email", "avatar", "role", "department", "skills", "status", "created_at", "updated_at",
				}).AddRow(1, "张三", "zhangsan@example.com", nil, "admin", nil, nil, "active", nil, nil)
				mock.ExpectQuery("SELECT \\* FROM `task_member`").
					WillReturnRows(rows)

				// Delete
				mock.ExpectBegin()
				mock.ExpectExec("DELETE FROM `task_member`").
					WillReturnResult(sqlmock.NewResult(0, 1))
				mock.ExpectCommit()
			},
		},
		{
			name:       "无效的成员ID",
			memberID:   "abc",
			expectCode: http.StatusBadRequest,
			setupMock:  func(mock sqlmock.Sqlmock) {},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler, router, mock := setupMemberTestWithDB(t)
			router.DELETE("/members/:id", handler.Delete)

			tt.setupMock(mock)

			req := httptest.NewRequest(http.MethodDelete, "/members/"+tt.memberID, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != tt.expectCode {
				t.Errorf("期望状态码 %d, 实际 %d, 响应: %s", tt.expectCode, w.Code, w.Body.String())
			}
		})
	}
}

func TestMemberHandler_GetAssignments(t *testing.T) {
	handler, router, mock := setupMemberTestWithDB(t)
	router.GET("/members/:id/assignments", handler.GetAssignments)

	// Mock task assignments query
	taskAssignmentRows := sqlmock.NewRows([]string{
		"id", "task_id", "member_id", "role", "estimated_hours", "actual_hours", "created_at",
	}).AddRow(1, 1, 1, "assignee", 5.0, 4.0, nil)
	mock.ExpectQuery("SELECT \\* FROM `task_assignment`").
		WillReturnRows(taskAssignmentRows)

	// Mock preloaded task
	taskRows := sqlmock.NewRows([]string{
		"id", "requirement_id", "title", "description", "status", "priority",
	}).AddRow(1, nil, "测试任务", "描述", "in-progress", "high")
	mock.ExpectQuery("SELECT \\* FROM `task`").
		WillReturnRows(taskRows)

	// Mock subtask assignments query (empty)
	subtaskAssignmentRows := sqlmock.NewRows([]string{
		"id", "subtask_id", "member_id", "role", "estimated_hours", "actual_hours", "created_at",
	})
	mock.ExpectQuery("SELECT \\* FROM `task_subtask_assignment`").
		WillReturnRows(subtaskAssignmentRows)

	// Mock preloaded subtask (empty result)
	mock.ExpectQuery("SELECT \\* FROM `task_subtask`").
		WillReturnRows(sqlmock.NewRows([]string{"id"}))

	req := httptest.NewRequest(http.MethodGet, "/members/1/assignments", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("期望状态码 %d, 实际 %d, 响应: %s", http.StatusOK, w.Code, w.Body.String())
	}
}

func TestMemberHandler_GetWorkload(t *testing.T) {
	handler, router, mock := setupMemberTestWithDB(t)
	router.GET("/members/:id/workload", handler.GetWorkload)

	// Mock task assignments query (empty)
	taskAssignmentRows := sqlmock.NewRows([]string{
		"id", "task_id", "member_id", "role", "estimated_hours", "actual_hours", "created_at",
	})
	mock.ExpectQuery("SELECT \\* FROM `task_assignment`").
		WillReturnRows(taskAssignmentRows)

	// Mock preloaded task (empty)
	mock.ExpectQuery("SELECT \\* FROM `task`").
		WillReturnRows(sqlmock.NewRows([]string{"id"}))

	// Mock subtask assignments query (empty)
	subtaskAssignmentRows := sqlmock.NewRows([]string{
		"id", "subtask_id", "member_id", "role", "estimated_hours", "actual_hours", "created_at",
	})
	mock.ExpectQuery("SELECT \\* FROM `task_subtask_assignment`").
		WillReturnRows(subtaskAssignmentRows)

	// Mock preloaded subtask (empty)
	mock.ExpectQuery("SELECT \\* FROM `task_subtask`").
		WillReturnRows(sqlmock.NewRows([]string{"id"}))

	req := httptest.NewRequest(http.MethodGet, "/members/1/workload", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("期望状态码 %d, 实际 %d, 响应: %s", http.StatusOK, w.Code, w.Body.String())
	}

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	data, ok := response["data"].(map[string]interface{})
	if !ok {
		t.Error("响应应包含 data 字段")
		return
	}

	if _, ok := data["taskCount"]; !ok {
		t.Error("响应应包含 taskCount 字段")
	}
}
