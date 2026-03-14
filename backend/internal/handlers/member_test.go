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

func setupMemberTest(t *testing.T) (*MemberHandler, *gin.Engine) {
	logger := zap.NewNop()
	handler := NewMemberHandler(logger)

	router := gin.New()
	return handler, router
}

func TestMemberHandler_List(t *testing.T) {
	handler, router := setupMemberTest(t)
	router.GET("/members", handler.List)

	req := httptest.NewRequest(http.MethodGet, "/members", nil)
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

func TestMemberHandler_Get(t *testing.T) {
	handler, router := setupMemberTest(t)
	router.GET("/members/:id", handler.Get)

	req := httptest.NewRequest(http.MethodGet, "/members/1", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("期望状态码 %d, 实际 %d", http.StatusOK, w.Code)
	}
}

func TestMemberHandler_Create(t *testing.T) {
	handler, router := setupMemberTest(t)
	router.POST("/members", handler.Create)

	req := httptest.NewRequest(http.MethodPost, "/members", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("期望状态码 %d, 实际 %d", http.StatusOK, w.Code)
	}
}

func TestMemberHandler_Update(t *testing.T) {
	handler, router := setupMemberTest(t)
	router.PUT("/members/:id", handler.Update)

	req := httptest.NewRequest(http.MethodPut, "/members/1", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("期望状态码 %d, 实际 %d", http.StatusOK, w.Code)
	}
}

func TestMemberHandler_Delete(t *testing.T) {
	handler, router := setupMemberTest(t)
	router.DELETE("/members/:id", handler.Delete)

	req := httptest.NewRequest(http.MethodDelete, "/members/1", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("期望状态码 %d, 实际 %d", http.StatusOK, w.Code)
	}
}

func TestMemberHandler_GetAssignments(t *testing.T) {
	handler, router := setupMemberTest(t)
	router.GET("/members/:id/assignments", handler.GetAssignments)

	req := httptest.NewRequest(http.MethodGet, "/members/1/assignments", nil)
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

func TestMemberHandler_GetWorkload(t *testing.T) {
	handler, router := setupMemberTest(t)
	router.GET("/members/:id/workload", handler.GetWorkload)

	req := httptest.NewRequest(http.MethodGet, "/members/1/workload", nil)
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
	expectedKeys := []string{"totalTasks", "activeTasks", "completedTasks", "estimatedHours", "actualHours"}
	for _, key := range expectedKeys {
		if _, ok := data[key]; !ok {
			t.Errorf("期望 data 包含 %s", key)
		}
	}
}
