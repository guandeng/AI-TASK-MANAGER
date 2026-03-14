package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func setupMessageTest(t *testing.T) (*MessageHandler, *gin.Engine) {
	logger := zap.NewNop()
	handler := NewMessageHandler(logger)

	router := gin.New()
	return handler, router
}

func TestMessageHandler_List(t *testing.T) {
	handler, router := setupMessageTest(t)
	router.GET("/messages", handler.List)

	req := httptest.NewRequest(http.MethodGet, "/messages", nil)
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

func TestMessageHandler_UnreadCount(t *testing.T) {
	handler, router := setupMessageTest(t)
	router.GET("/messages/unread-count", handler.UnreadCount)

	req := httptest.NewRequest(http.MethodGet, "/messages/unread-count", nil)
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
	if _, ok := data["count"]; !ok {
		t.Error("期望 data 包含 count")
	}
}

func TestMessageHandler_MarkRead(t *testing.T) {
	handler, router := setupMessageTest(t)
	router.PUT("/messages/:id/read", handler.MarkRead)

	req := httptest.NewRequest(http.MethodPut, "/messages/1/read", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("期望状态码 %d, 实际 %d", http.StatusOK, w.Code)
	}
}

func TestMessageHandler_MarkAllRead(t *testing.T) {
	handler, router := setupMessageTest(t)
	router.PUT("/messages/read-all", handler.MarkAllRead)

	req := httptest.NewRequest(http.MethodPut, "/messages/read-all", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("期望状态码 %d, 实际 %d", http.StatusOK, w.Code)
	}
}

func TestMessageHandler_Delete(t *testing.T) {
	handler, router := setupMessageTest(t)
	router.DELETE("/messages/:id", handler.Delete)

	req := httptest.NewRequest(http.MethodDelete, "/messages/1", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("期望状态码 %d, 实际 %d", http.StatusOK, w.Code)
	}
}
