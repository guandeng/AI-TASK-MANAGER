package handlers

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ai-task-manager/backend/internal/config"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func setupKnowledgeTest() (*KnowledgeHandler, *gin.Engine) {
	logger := zap.NewNop()
	cfg := &config.Config{
		Knowledge: config.KnowledgeConfig{
			Enabled:   true,
			Paths:     []string{"/data/knowledge"},
			MaxSize:   500,
			MaxFiles:  50,
			FileTypes: []string{".md", ".txt"},
		},
	}
	handler := NewKnowledgeHandler(logger, cfg)
	router := gin.New()
	return handler, router
}

func TestKnowledgeHandler_GetSummary(t *testing.T) {
	handler, router := setupKnowledgeTest()
	router.GET("/knowledge/summary", handler.GetSummary)

	req := httptest.NewRequest(http.MethodGet, "/knowledge/summary", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("期望状态码 %d, 实际 %d", http.StatusOK, w.Code)
	}
}

func TestKnowledgeHandler_GetSummary_WithPaths(t *testing.T) {
	handler, router := setupKnowledgeTest()
	router.GET("/knowledge/summary", handler.GetSummary)

	req := httptest.NewRequest(http.MethodGet, "/knowledge/summary?paths=/custom/path1&paths=/custom/path2", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("期望状态码 %d, 实际 %d", http.StatusOK, w.Code)
	}
}

func TestKnowledgeHandler_Load(t *testing.T) {
	handler, router := setupKnowledgeTest()
	router.POST("/knowledge/load", handler.Load)

	body := `{"paths":["/path1","/path2"],"additionalContext":"额外上下文"}`
	req := httptest.NewRequest(http.MethodPost, "/knowledge/load", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("期望状态码 %d, 实际 %d", http.StatusOK, w.Code)
	}
}

func TestKnowledgeHandler_Load_EmptyPaths(t *testing.T) {
	handler, router := setupKnowledgeTest()
	router.POST("/knowledge/load", handler.Load)

	body := `{"paths":[]}`
	req := httptest.NewRequest(http.MethodPost, "/knowledge/load", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("期望状态码 %d, 实际 %d", http.StatusOK, w.Code)
	}
}

func TestKnowledgeHandler_Load_InvalidJSON(t *testing.T) {
	handler, router := setupKnowledgeTest()
	router.POST("/knowledge/load", handler.Load)

	req := httptest.NewRequest(http.MethodPost, "/knowledge/load", bytes.NewBufferString(`invalid json`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("期望状态码 %d, 实际 %d", http.StatusBadRequest, w.Code)
	}
}
