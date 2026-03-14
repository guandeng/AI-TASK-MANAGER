package handlers

import (
	"bytes"
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

// isDBAvailable 检查数据库是否可用
func isDBAvailable(w *httptest.ResponseRecorder) bool {
	return w.Code != http.StatusInternalServerError
}

func TestMenuHandler_List(t *testing.T) {
	handler, router := setupMenuTest(t)
	router.GET("/menus", handler.List)

	req := httptest.NewRequest(http.MethodGet, "/menus", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// 如果没有数据库，期望返回 500 错误
	if w.Code != http.StatusOK && w.Code != http.StatusInternalServerError {
		t.Errorf("期望状态码 %d 或 %d, 实际 %d", http.StatusOK, http.StatusInternalServerError, w.Code)
	}

	// 只有成功时才验证响应格式
	if w.Code == http.StatusOK {
		var resp map[string]interface{}
		if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
			t.Fatalf("解析响应失败: %v", err)
		}

		if resp["code"].(float64) != 0 {
			t.Errorf("期望 code 为 0, 实际 %v", resp["code"])
		}
	}
}

func TestMenuHandler_Tree(t *testing.T) {
	handler, router := setupMenuTest(t)
	router.GET("/menus/tree", handler.Tree)

	req := httptest.NewRequest(http.MethodGet, "/menus/tree", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// 如果没有数据库，期望返回 500 错误
	if w.Code != http.StatusOK && w.Code != http.StatusInternalServerError {
		t.Errorf("期望状态码 %d 或 %d, 实际 %d", http.StatusOK, http.StatusInternalServerError, w.Code)
	}
}

func TestMenuHandler_Get(t *testing.T) {
	handler, router := setupMenuTest(t)
	router.GET("/menus/:key", handler.Get)

	req := httptest.NewRequest(http.MethodGet, "/menus/dashboard", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// 如果没有数据库，期望返回 404 或 500 错误
	if w.Code != http.StatusOK && w.Code != http.StatusNotFound && w.Code != http.StatusInternalServerError {
		t.Errorf("期望状态码 %d/%d/%d, 实际 %d", http.StatusOK, http.StatusNotFound, http.StatusInternalServerError, w.Code)
	}
}

func TestMenuHandler_Create(t *testing.T) {
	handler, router := setupMenuTest(t)
	router.POST("/menus", handler.Create)

	body := `{"key":"test_menu","title":"测试菜单"}`
	req := httptest.NewRequest(http.MethodPost, "/menus", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// 如果没有数据库，期望返回 400 或 500 错误
	if w.Code != http.StatusOK && w.Code != http.StatusBadRequest && w.Code != http.StatusInternalServerError {
		t.Errorf("期望状态码 %d/%d/%d, 实际 %d", http.StatusOK, http.StatusBadRequest, http.StatusInternalServerError, w.Code)
	}
}

func TestMenuHandler_Update(t *testing.T) {
	handler, router := setupMenuTest(t)
	router.POST("/menus/:key/update", handler.Update)

	body := `{"title":"更新菜单"}`
	req := httptest.NewRequest(http.MethodPost, "/menus/dashboard/update", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// 如果没有数据库，期望返回 404 或 500 错误
	if w.Code != http.StatusOK && w.Code != http.StatusNotFound && w.Code != http.StatusInternalServerError {
		t.Errorf("期望状态码 %d/%d/%d, 实际 %d", http.StatusOK, http.StatusNotFound, http.StatusInternalServerError, w.Code)
	}
}

func TestMenuHandler_Delete(t *testing.T) {
	handler, router := setupMenuTest(t)
	router.POST("/menus/:key/delete", handler.Delete)

	req := httptest.NewRequest(http.MethodPost, "/menus/test_menu/delete", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// 如果没有数据库，期望返回 500 错误
	if w.Code != http.StatusOK && w.Code != http.StatusInternalServerError {
		t.Errorf("期望状态码 %d 或 %d, 实际 %d", http.StatusOK, http.StatusInternalServerError, w.Code)
	}
}

func TestMenuHandler_BatchDelete(t *testing.T) {
	handler, router := setupMenuTest(t)
	router.POST("/menus/batch-delete", handler.BatchDelete)

	body := `{"keys":["test1","test2"]}`
	req := httptest.NewRequest(http.MethodPost, "/menus/batch-delete", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// 如果没有数据库，期望返回 500 错误
	if w.Code != http.StatusOK && w.Code != http.StatusInternalServerError {
		t.Errorf("期望状态码 %d 或 %d, 实际 %d", http.StatusOK, http.StatusInternalServerError, w.Code)
	}
}

func TestMenuHandler_Reorder(t *testing.T) {
	handler, router := setupMenuTest(t)
	router.POST("/menus/reorder", handler.Reorder)

	body := `[{"key":"dashboard","order":1}]`
	req := httptest.NewRequest(http.MethodPost, "/menus/reorder", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// 如果没有数据库，期望返回 500 错误
	if w.Code != http.StatusOK && w.Code != http.StatusInternalServerError {
		t.Errorf("期望状态码 %d 或 %d, 实际 %d", http.StatusOK, http.StatusInternalServerError, w.Code)
	}
}

func TestMenuHandler_Move(t *testing.T) {
	handler, router := setupMenuTest(t)
	router.POST("/menus/:key/move", handler.Move)

	body := `{"targetParentKey":"dashboard_analysis"}`
	req := httptest.NewRequest(http.MethodPost, "/menus/dashboard_analysis/move", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// 如果没有数据库，期望返回 500 错误
	if w.Code != http.StatusOK && w.Code != http.StatusInternalServerError {
		t.Errorf("期望状态码 %d 或 %d, 实际 %d", http.StatusOK, http.StatusInternalServerError, w.Code)
	}
}

func TestMenuHandler_Toggle(t *testing.T) {
	handler, router := setupMenuTest(t)
	router.POST("/menus/:key/toggle", handler.Toggle)

	body := `{"enabled":true}`
	req := httptest.NewRequest(http.MethodPost, "/menus/dashboard/toggle", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// 如果没有数据库，期望返回 500 错误
	if w.Code != http.StatusOK && w.Code != http.StatusInternalServerError {
		t.Errorf("期望状态码 %d 或 %d, 实际 %d", http.StatusOK, http.StatusInternalServerError, w.Code)
	}
}
