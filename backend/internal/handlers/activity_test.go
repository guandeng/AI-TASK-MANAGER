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

func setupActivityTest(t *testing.T) (*ActivityHandler, *gin.Engine) {
	logger := zap.NewNop()
	handler := NewActivityHandler(logger)

	router := gin.New()
	return handler, router
}

func TestActivityHandler_List(t *testing.T) {
	handler, router := setupActivityTest(t)
	router.GET("/activities", handler.List)

	req := httptest.NewRequest(http.MethodGet, "/activities", nil)
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

func TestActivityHandler_ListByTask(t *testing.T) {
	handler, router := setupActivityTest(t)
	router.GET("/tasks/:taskId/activities", handler.ListByTask)

	req := httptest.NewRequest(http.MethodGet, "/tasks/1/activities", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("期望状态码 %d, 实际 %d", http.StatusOK, w.Code)
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("解析响应失败: %v", err)
	}

	data := resp["data"].([]interface{})
	if data == nil {
		t.Error("期望 data 不为 nil")
	}
}

func TestActivityHandler_Statistics(t *testing.T) {
	handler, router := setupActivityTest(t)
	router.GET("/activities/statistics", handler.Statistics)

	req := httptest.NewRequest(http.MethodGet, "/activities/statistics", nil)
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
	expectedKeys := []string{"total", "today", "thisWeek"}
	for _, key := range expectedKeys {
		if _, ok := data[key]; !ok {
			t.Errorf("期望 data 包含 %s", key)
		}
	}
}
