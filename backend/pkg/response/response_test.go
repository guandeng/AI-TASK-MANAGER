package response

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func setupResponseTest() (*gin.Context, *httptest.ResponseRecorder) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	return c, w
}

func TestSuccess(t *testing.T) {
	c, w := setupResponseTest()
	data := map[string]string{"name": "test"}
	Success(c, data)

	if w.Code != http.StatusOK {
		t.Errorf("期望状态码 %d, 实际 %d", http.StatusOK, w.Code)
	}

	var resp Response
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Errorf("解析响应失败: %v", err)
	}
	if resp.Code != 0 {
		t.Errorf("期望 code 0, 实际 %d", resp.Code)
	}
	if resp.Message != "success" {
		t.Errorf("期望 message 'success', 实际 '%s'", resp.Message)
	}
}

func TestSuccessWithMessage(t *testing.T) {
	c, w := setupResponseTest()
	data := map[string]string{"name": "test"}
	SuccessWithMessage(c, "操作成功", data)

	if w.Code != http.StatusOK {
		t.Errorf("期望状态码 %d, 实际 %d", http.StatusOK, w.Code)
	}

	var resp Response
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Errorf("解析响应失败: %v", err)
	}
	if resp.Message != "操作成功" {
		t.Errorf("期望 message '操作成功', 实际 '%s'", resp.Message)
	}
}

func TestError(t *testing.T) {
	c, w := setupResponseTest()
	Error(c, 1001, "参数错误")

	if w.Code != http.StatusOK {
		t.Errorf("期望状态码 %d, 实际 %d", http.StatusOK, w.Code)
	}

	var resp Response
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Errorf("解析响应失败: %v", err)
	}
	if resp.Code != 1001 {
		t.Errorf("期望 code 1001, 实际 %d", resp.Code)
	}
	if resp.Message != "参数错误" {
		t.Errorf("期望 message '参数错误', 实际 '%s'", resp.Message)
	}
}

func TestErrorWithData(t *testing.T) {
	c, w := setupResponseTest()
	ErrorWithData(c, 1002, "验证失败", map[string]string{"field": "name"})

	if w.Code != http.StatusOK {
		t.Errorf("期望状态码 %d, 实际 %d", http.StatusOK, w.Code)
	}

	var resp Response
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Errorf("解析响应失败: %v", err)
	}
	if resp.Code != 1002 {
		t.Errorf("期望 code 1002, 实际 %d", resp.Code)
	}
	if resp.Data == nil {
		t.Error("期望 data 不为 nil")
	}
}

func TestBadRequest(t *testing.T) {
	c, w := setupResponseTest()
	BadRequest(c, "无效的请求参数")

	if w.Code != http.StatusBadRequest {
		t.Errorf("期望状态码 %d, 实际 %d", http.StatusBadRequest, w.Code)
	}

	var resp Response
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Errorf("解析响应失败: %v", err)
	}
	if resp.Code != http.StatusBadRequest {
		t.Errorf("期望 code %d, 实际 %d", http.StatusBadRequest, resp.Code)
	}
}

func TestNotFound(t *testing.T) {
	c, w := setupResponseTest()
	NotFound(c, "资源不存在")

	if w.Code != http.StatusNotFound {
		t.Errorf("期望状态码 %d, 实际 %d", http.StatusNotFound, w.Code)
	}

	var resp Response
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Errorf("解析响应失败: %v", err)
	}
	if resp.Code != http.StatusNotFound {
		t.Errorf("期望 code %d, 实际 %d", http.StatusNotFound, resp.Code)
	}
}

func TestServerError(t *testing.T) {
	c, w := setupResponseTest()
	ServerError(c, "服务器内部错误")

	if w.Code != http.StatusInternalServerError {
		t.Errorf("期望状态码 %d, 实际 %d", http.StatusInternalServerError, w.Code)
	}

	var resp Response
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Errorf("解析响应失败: %v", err)
	}
	if resp.Code != http.StatusInternalServerError {
		t.Errorf("期望 code %d, 实际 %d", http.StatusInternalServerError, resp.Code)
	}
}

func TestSuccessPage(t *testing.T) {
	c, w := setupResponseTest()
	list := []map[string]string{{"id": "1"}, {"id": "2"}}
	SuccessPage(c, list, 100, 1, 10)

	if w.Code != http.StatusOK {
		t.Errorf("期望状态码 %d, 实际 %d", http.StatusOK, w.Code)
	}

	var resp Response
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Errorf("解析响应失败: %v", err)
	}
	if resp.Code != 0 {
		t.Errorf("期望 code 0, 实际 %d", resp.Code)
	}

	data, ok := resp.Data.(map[string]interface{})
	if !ok {
		t.Error("期望 data 为 map")
		return
	}
	if data["total"].(float64) != 100 {
		t.Errorf("期望 total 100, 实际 %v", data["total"])
	}
	if data["page"].(float64) != 1 {
		t.Errorf("期望 page 1, 实际 %v", data["page"])
	}
	if data["pageSize"].(float64) != 10 {
		t.Errorf("期望 pageSize 10, 实际 %v", data["pageSize"])
	}
}
