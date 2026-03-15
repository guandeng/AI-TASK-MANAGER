package handlers

import (
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

func setupMessageTestWithDB(t *testing.T) (*MessageHandler, *gin.Engine, sqlmock.Sqlmock) {
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

	handler := NewMessageHandler(logger)
	router := gin.New()

	return handler, router, mock
}

func TestMessageHandler_List(t *testing.T) {
	handler, router, mock := setupMessageTestWithDB(t)
	router.GET("/messages", handler.List)

	mock.ExpectQuery("SELECT count\\(\\*\\)").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

	mock.ExpectQuery("SELECT").
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "task_id", "type", "content", "status", "is_read", "created_at",
		}))

	req := httptest.NewRequest(http.MethodGet, "/messages", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("期望状态码 %d, 实际 %d", http.StatusOK, w.Code)
	}
}

func TestMessageHandler_UnreadCount(t *testing.T) {
	handler, router, mock := setupMessageTestWithDB(t)
	router.GET("/messages/unread-count", handler.UnreadCount)

	mock.ExpectQuery("SELECT count\\(\\*\\)").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(5))

	req := httptest.NewRequest(http.MethodGet, "/messages/unread-count", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("期望状态码 %d, 实际 %d", http.StatusOK, w.Code)
	}
}

func TestMessageHandler_MarkRead(t *testing.T) {
	tests := []struct {
		name       string
		messageID  string
		expectCode int
		setupMock  func(mock sqlmock.Sqlmock)
	}{
		{
			name:       "标记已读成功",
			messageID:  "1",
			expectCode: http.StatusOK,
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec("UPDATE `task_message`").
					WillReturnResult(sqlmock.NewResult(0, 1))
				mock.ExpectCommit()
			},
		},
		{
			name:       "无效的消息ID",
			messageID:  "abc",
			expectCode: http.StatusBadRequest,
			setupMock:  func(mock sqlmock.Sqlmock) {},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler, router, mock := setupMessageTestWithDB(t)
			router.PUT("/messages/:id/read", handler.MarkRead)

			tt.setupMock(mock)

			req := httptest.NewRequest(http.MethodPut, "/messages/"+tt.messageID+"/read", nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != tt.expectCode {
				t.Errorf("期望状态码 %d, 实际 %d", tt.expectCode, w.Code)
			}
		})
	}
}

func TestMessageHandler_MarkAllRead(t *testing.T) {
	handler, router, mock := setupMessageTestWithDB(t)
	router.PUT("/messages/read-all", handler.MarkAllRead)

	mock.ExpectBegin()
	mock.ExpectExec("UPDATE `task_message`").
		WillReturnResult(sqlmock.NewResult(0, 3))
	mock.ExpectCommit()

	req := httptest.NewRequest(http.MethodPut, "/messages/read-all", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("期望状态码 %d, 实际 %d", http.StatusOK, w.Code)
	}
}

func TestMessageHandler_Delete(t *testing.T) {
	tests := []struct {
		name       string
		messageID  string
		expectCode int
		setupMock  func(mock sqlmock.Sqlmock)
	}{
		{
			name:       "删除消息成功",
			messageID:  "1",
			expectCode: http.StatusOK,
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec("DELETE FROM `task_message`").
					WillReturnResult(sqlmock.NewResult(0, 1))
				mock.ExpectCommit()
			},
		},
		{
			name:       "无效的消息ID",
			messageID:  "abc",
			expectCode: http.StatusBadRequest,
			setupMock:  func(mock sqlmock.Sqlmock) {},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler, router, mock := setupMessageTestWithDB(t)
			router.DELETE("/messages/:id", handler.Delete)

			tt.setupMock(mock)

			req := httptest.NewRequest(http.MethodDelete, "/messages/"+tt.messageID, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != tt.expectCode {
				t.Errorf("期望状态码 %d, 实际 %d", tt.expectCode, w.Code)
			}
		})
	}
}
