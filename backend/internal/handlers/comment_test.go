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

func setupCommentTestWithDB(t *testing.T) (*CommentHandler, *gin.Engine, sqlmock.Sqlmock) {
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

	handler := NewCommentHandler(logger)
	router := gin.New()

	return handler, router, mock
}

func TestCommentHandler_List(t *testing.T) {
	tests := []struct {
		name       string
		taskID     string
		expectCode int
		setupMock  func(mock sqlmock.Sqlmock)
	}{
		{
			name:       "获取评论列表成功",
			taskID:     "1",
			expectCode: http.StatusOK,
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT").
					WillReturnRows(sqlmock.NewRows([]string{
						"id", "task_id", "subtask_id", "member_id", "parent_id", "content", "mentions", "created_at", "updated_at",
					}))
			},
		},
		{
			name:       "无效的任务ID",
			taskID:     "abc",
			expectCode: http.StatusBadRequest,
			setupMock:  func(mock sqlmock.Sqlmock) {},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler, router, mock := setupCommentTestWithDB(t)
			router.GET("/tasks/:taskId/comments", handler.List)

			tt.setupMock(mock)

			req := httptest.NewRequest(http.MethodGet, "/tasks/"+tt.taskID+"/comments", nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != tt.expectCode {
				t.Errorf("期望状态码 %d, 实际 %d", tt.expectCode, w.Code)
			}
		})
	}
}

func TestCommentHandler_GetTree(t *testing.T) {
	tests := []struct {
		name       string
		taskID     string
		expectCode int
		setupMock  func(mock sqlmock.Sqlmock)
	}{
		{
			name:       "获取评论树成功",
			taskID:     "1",
			expectCode: http.StatusOK,
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT").
					WillReturnRows(sqlmock.NewRows([]string{
						"id", "task_id", "subtask_id", "member_id", "parent_id", "content", "mentions", "created_at", "updated_at",
					}))
			},
		},
		{
			name:       "无效的任务ID",
			taskID:     "abc",
			expectCode: http.StatusBadRequest,
			setupMock:  func(mock sqlmock.Sqlmock) {},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler, router, mock := setupCommentTestWithDB(t)
			router.GET("/tasks/:taskId/comments/tree", handler.GetTree)

			tt.setupMock(mock)

			req := httptest.NewRequest(http.MethodGet, "/tasks/"+tt.taskID+"/comments/tree", nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != tt.expectCode {
				t.Errorf("期望状态码 %d, 实际 %d", tt.expectCode, w.Code)
			}
		})
	}
}

func TestCommentHandler_GetStatistics(t *testing.T) {
	tests := []struct {
		name       string
		taskID     string
		expectCode int
		setupMock  func(mock sqlmock.Sqlmock)
	}{
		{
			name:       "获取统计成功",
			taskID:     "1",
			expectCode: http.StatusOK,
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT count\\(\\*\\)").
					WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(10))
				mock.ExpectQuery("SELECT count\\(DISTINCT `member_id`\\)").
					WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(3))
			},
		},
		{
			name:       "无效的任务ID",
			taskID:     "abc",
			expectCode: http.StatusBadRequest,
			setupMock:  func(mock sqlmock.Sqlmock) {},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler, router, mock := setupCommentTestWithDB(t)
			router.GET("/tasks/:taskId/comments/statistics", handler.GetStatistics)

			tt.setupMock(mock)

			req := httptest.NewRequest(http.MethodGet, "/tasks/"+tt.taskID+"/comments/statistics", nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != tt.expectCode {
				t.Errorf("期望状态码 %d, 实际 %d", tt.expectCode, w.Code)
			}
		})
	}
}

func TestCommentHandler_Get(t *testing.T) {
	tests := []struct {
		name       string
		taskID     string
		commentID  string
		expectCode int
		setupMock  func(mock sqlmock.Sqlmock)
	}{
		{
			name:       "获取评论成功",
			taskID:     "1",
			commentID:  "1",
			expectCode: http.StatusOK,
			setupMock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{
					"id", "task_id", "subtask_id", "member_id", "parent_id", "content", "mentions", "created_at", "updated_at",
				}).AddRow(1, 1, nil, 1, nil, "测试评论", "[]", nil, nil)
				mock.ExpectQuery("SELECT").
					WillReturnRows(rows)
				// Mock member preload
				memberRows := sqlmock.NewRows([]string{
					"id", "name", "email", "avatar", "role", "department", "skills", "status", "created_at", "updated_at",
				}).AddRow(1, "测试用户", "test@example.com", nil, "member", nil, nil, "active", nil, nil)
				mock.ExpectQuery("SELECT \\* FROM `task_member`").
					WillReturnRows(memberRows)
			},
		},
		{
			name:       "评论不存在",
			taskID:     "1",
			commentID:  "999",
			expectCode: http.StatusNotFound,
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT").
					WillReturnRows(sqlmock.NewRows([]string{"id"}))
			},
		},
		{
			name:       "无效的任务ID",
			taskID:     "abc",
			commentID:  "1",
			expectCode: http.StatusBadRequest,
			setupMock:  func(mock sqlmock.Sqlmock) {},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler, router, mock := setupCommentTestWithDB(t)
			router.GET("/tasks/:taskId/comments/:commentId", handler.Get)

			tt.setupMock(mock)

			req := httptest.NewRequest(http.MethodGet, "/tasks/"+tt.taskID+"/comments/"+tt.commentID, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != tt.expectCode {
				t.Errorf("期望状态码 %d, 实际 %d", tt.expectCode, w.Code)
			}
		})
	}
}

func TestCommentHandler_Create(t *testing.T) {
	tests := []struct {
		name       string
		taskID     string
		body       string
		expectCode int
		setupMock  func(mock sqlmock.Sqlmock)
	}{
		{
			name:       "创建评论成功",
			taskID:     "1",
			body:       `{"memberId":1,"content":"测试评论"}`,
			expectCode: http.StatusOK,
			setupMock: func(mock sqlmock.Sqlmock) {
				// Check task exists
				taskRows := sqlmock.NewRows([]string{
					"id", "requirement_id", "title", "description", "status", "priority", "dependencies",
				}).AddRow(1, nil, "测试任务", "描述", "pending", "medium", "[]")
				mock.ExpectQuery("SELECT \\* FROM `task_task`").
					WillReturnRows(taskRows)
				// Create comment
				mock.ExpectBegin()
				mock.ExpectExec("INSERT INTO `task_comment`").
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()
				// Reload comment
				commentRows := sqlmock.NewRows([]string{
					"id", "task_id", "subtask_id", "member_id", "parent_id", "content", "mentions", "created_at", "updated_at",
				}).AddRow(1, 1, nil, 1, nil, "测试评论", "[]", nil, nil)
				mock.ExpectQuery("SELECT").
					WillReturnRows(commentRows)
			},
		},
		{
			name:       "任务不存在",
			taskID:     "999",
			body:       `{"memberId":1,"content":"测试评论"}`,
			expectCode: http.StatusNotFound,
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT \\* FROM `task_task`").
					WillReturnRows(sqlmock.NewRows([]string{"id"}))
			},
		},
		{
			name:       "缺少必填字段",
			taskID:     "1",
			body:       `{"content":"测试评论"}`,
			expectCode: http.StatusBadRequest,
			setupMock:  func(mock sqlmock.Sqlmock) {},
		},
		{
			name:       "无效的任务ID",
			taskID:     "abc",
			body:       `{"memberId":1,"content":"测试评论"}`,
			expectCode: http.StatusBadRequest,
			setupMock:  func(mock sqlmock.Sqlmock) {},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler, router, mock := setupCommentTestWithDB(t)
			router.POST("/tasks/:taskId/comments", handler.Create)

			tt.setupMock(mock)

			req := httptest.NewRequest(http.MethodPost, "/tasks/"+tt.taskID+"/comments", bytes.NewBufferString(tt.body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != tt.expectCode {
				t.Errorf("期望状态码 %d, 实际 %d, 响应: %s", tt.expectCode, w.Code, w.Body.String())
			}
		})
	}
}

func TestCommentHandler_Update(t *testing.T) {
	tests := []struct {
		name       string
		taskID     string
		commentID  string
		body       string
		expectCode int
		setupMock  func(mock sqlmock.Sqlmock)
	}{
		{
			name:       "更新评论成功",
			taskID:     "1",
			commentID:  "1",
			body:       `{"content":"更新后的评论"}`,
			expectCode: http.StatusOK,
			setupMock: func(mock sqlmock.Sqlmock) {
				// Find comment
				rows := sqlmock.NewRows([]string{
					"id", "task_id", "subtask_id", "member_id", "parent_id", "content", "mentions", "created_at", "updated_at",
				}).AddRow(1, 1, nil, 1, nil, "测试评论", "[]", nil, nil)
				mock.ExpectQuery("SELECT").
					WillReturnRows(rows)
				// Update
				mock.ExpectBegin()
				mock.ExpectExec("UPDATE `task_comment`").
					WillReturnResult(sqlmock.NewResult(0, 1))
				mock.ExpectCommit()
				// Reload
				reloadRows := sqlmock.NewRows([]string{
					"id", "task_id", "subtask_id", "member_id", "parent_id", "content", "mentions", "created_at", "updated_at",
				}).AddRow(1, 1, nil, 1, nil, "更新后的评论", "[]", nil, nil)
				mock.ExpectQuery("SELECT").
					WillReturnRows(reloadRows)
			},
		},
		{
			name:       "评论不存在",
			taskID:     "1",
			commentID:  "999",
			body:       `{"content":"更新后的评论"}`,
			expectCode: http.StatusNotFound,
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT").
					WillReturnRows(sqlmock.NewRows([]string{"id"}))
			},
		},
		{
			name:       "无效的任务ID",
			taskID:     "abc",
			commentID:  "1",
			body:       `{"content":"更新后的评论"}`,
			expectCode: http.StatusBadRequest,
			setupMock:  func(mock sqlmock.Sqlmock) {},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler, router, mock := setupCommentTestWithDB(t)
			router.PUT("/tasks/:taskId/comments/:commentId", handler.Update)

			tt.setupMock(mock)

			req := httptest.NewRequest(http.MethodPut, "/tasks/"+tt.taskID+"/comments/"+tt.commentID, bytes.NewBufferString(tt.body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != tt.expectCode {
				t.Errorf("期望状态码 %d, 实际 %d", tt.expectCode, w.Code)
			}
		})
	}
}

func TestCommentHandler_Delete(t *testing.T) {
	tests := []struct {
		name       string
		taskID     string
		commentID  string
		body       string
		expectCode int
		setupMock  func(mock sqlmock.Sqlmock)
	}{
		{
			name:       "删除评论成功",
			taskID:     "1",
			commentID:  "1",
			body:       `{"memberId":1}`,
			expectCode: http.StatusOK,
			setupMock: func(mock sqlmock.Sqlmock) {
				// Find comment
				rows := sqlmock.NewRows([]string{
					"id", "task_id", "subtask_id", "member_id", "parent_id", "content", "mentions", "created_at", "updated_at",
				}).AddRow(1, 1, nil, 1, nil, "测试评论", "[]", nil, nil)
				mock.ExpectQuery("SELECT").
					WillReturnRows(rows)
				// Delete
				mock.ExpectBegin()
				mock.ExpectExec("DELETE FROM `task_comment`").
					WillReturnResult(sqlmock.NewResult(0, 1))
				mock.ExpectCommit()
			},
		},
		{
			name:       "评论不存在",
			taskID:     "1",
			commentID:  "999",
			body:       `{"memberId":1}`,
			expectCode: http.StatusNotFound,
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT").
					WillReturnRows(sqlmock.NewRows([]string{"id"}))
			},
		},
		{
			name:       "无效的任务ID",
			taskID:     "abc",
			commentID:  "1",
			body:       `{"memberId":1}`,
			expectCode: http.StatusBadRequest,
			setupMock:  func(mock sqlmock.Sqlmock) {},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler, router, mock := setupCommentTestWithDB(t)
			router.DELETE("/tasks/:taskId/comments/:commentId", handler.Delete)

			tt.setupMock(mock)

			req := httptest.NewRequest(http.MethodDelete, "/tasks/"+tt.taskID+"/comments/"+tt.commentID, bytes.NewBufferString(tt.body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != tt.expectCode {
				t.Errorf("期望状态码 %d, 实际 %d", tt.expectCode, w.Code)
			}
		})
	}
}

func TestCommentHandler_GetReplies(t *testing.T) {
	tests := []struct {
		name       string
		taskID     string
		commentID  string
		expectCode int
		setupMock  func(mock sqlmock.Sqlmock)
	}{
		{
			name:       "获取回复成功",
			taskID:     "1",
			commentID:  "1",
			expectCode: http.StatusOK,
			setupMock: func(mock sqlmock.Sqlmock) {
				// Check parent comment exists
				parentRows := sqlmock.NewRows([]string{
					"id", "task_id", "subtask_id", "member_id", "parent_id", "content", "mentions", "created_at", "updated_at",
				}).AddRow(1, 1, nil, 1, nil, "父评论", "[]", nil, nil)
				mock.ExpectQuery("SELECT").
					WillReturnRows(parentRows)
				// Get replies
				mock.ExpectQuery("SELECT").
					WillReturnRows(sqlmock.NewRows([]string{
						"id", "task_id", "subtask_id", "member_id", "parent_id", "content", "mentions", "created_at", "updated_at",
					}))
			},
		},
		{
			name:       "父评论不存在",
			taskID:     "1",
			commentID:  "999",
			expectCode: http.StatusNotFound,
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT").
					WillReturnRows(sqlmock.NewRows([]string{"id"}))
			},
		},
		{
			name:       "无效的任务ID",
			taskID:     "abc",
			commentID:  "1",
			expectCode: http.StatusBadRequest,
			setupMock:  func(mock sqlmock.Sqlmock) {},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler, router, mock := setupCommentTestWithDB(t)
			router.GET("/tasks/:taskId/comments/:commentId/replies", handler.GetReplies)

			tt.setupMock(mock)

			req := httptest.NewRequest(http.MethodGet, "/tasks/"+tt.taskID+"/comments/"+tt.commentID+"/replies", nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != tt.expectCode {
				t.Errorf("期望状态码 %d, 实际 %d", tt.expectCode, w.Code)
			}
		})
	}
}
