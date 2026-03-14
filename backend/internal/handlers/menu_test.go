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

func setupMenuTest(t *testing.T) (*MenuHandler, *gin.Engine) {
	logger := zap.NewNop()
	handler := NewMenuHandler(logger)

	router := gin.New()
	return handler, router
}

func TestMenuHandler_List(t *testing.T) {
	handler, router := setupMenuTest(t)
	router.GET("/menus", handler.List)

	req := httptest.NewRequest(http.MethodGet, "/menus", nil)
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

func TestMenuHandler_Tree(t *testing.T) {
	handler, router := setupMenuTest(t)
	router.GET("/menus/tree", handler.Tree)

	req := httptest.NewRequest(http.MethodGet, "/menus/tree", nil)
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

func TestMenuHandler_Get(t *testing.T) {
	handler, router := setupMenuTest(t)
	router.GET("/menus/:id", handler.Get)

	req := httptest.NewRequest(http.MethodGet, "/menus/1", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("期望状态码 %d, 实际 %d", http.StatusOK, w.Code)
	}
}

func TestMenuHandler_Create(t *testing.T) {
	handler, router := setupMenuTest(t)
	router.POST("/menus", handler.Create)

	req := httptest.NewRequest(http.MethodPost, "/menus", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("期望状态码 %d, 实际 %d", http.StatusOK, w.Code)
	}
}

func TestMenuHandler_Update(t *testing.T) {
	handler, router := setupMenuTest(t)
	router.PUT("/menus/:id", handler.Update)

	req := httptest.NewRequest(http.MethodPut, "/menus/1", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("期望状态码 %d, 实际 %d", http.StatusOK, w.Code)
	}
}

func TestMenuHandler_Delete(t *testing.T) {
	handler, router := setupMenuTest(t)
	router.DELETE("/menus/:id", handler.Delete)

	req := httptest.NewRequest(http.MethodDelete, "/menus/1", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("期望状态码 %d, 实际 %d", http.StatusOK, w.Code)
	}
}

func TestMenuHandler_BatchDelete(t *testing.T) {
	handler, router := setupMenuTest(t)
	router.DELETE("/menus/batch", handler.BatchDelete)

	req := httptest.NewRequest(http.MethodDelete, "/menus/batch", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("期望状态码 %d, 实际 %d", http.StatusOK, w.Code)
	}
}

func TestMenuHandler_Reorder(t *testing.T) {
	handler, router := setupMenuTest(t)
	router.PUT("/menus/reorder", handler.Reorder)

	req := httptest.NewRequest(http.MethodPut, "/menus/reorder", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("期望状态码 %d, 实际 %d", http.StatusOK, w.Code)
	}
}

func TestMenuHandler_Move(t *testing.T) {
	handler, router := setupMenuTest(t)
	router.PUT("/menus/:id/move", handler.Move)

	req := httptest.NewRequest(http.MethodPut, "/menus/1/move", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("期望状态码 %d, 实际 %d", http.StatusOK, w.Code)
	}
}

func TestMenuHandler_Toggle(t *testing.T) {
	handler, router := setupMenuTest(t)
	router.PUT("/menus/:id/toggle", handler.Toggle)

	req := httptest.NewRequest(http.MethodPut, "/menus/1/toggle", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("期望状态码 %d, 实际 %d", http.StatusOK, w.Code)
	}
}
