package handlers

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ai-task-manager/backend/internal/models"
	"github.com/ai-task-manager/backend/internal/services"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

func init() {
	gin.SetMode(gin.TestMode)
}

// MockBackupService 模拟备份服务
type MockBackupService struct {
	backups      []models.Backup
	total        int64
	createErr    error
	restoreErr   error
	deleteErr    error
	schedule     *models.BackupSchedule
	scheduleErr  error
	updateSchErr error
}

func (m *MockBackupService) GetBackups(requirementID uint64, page, pageSize int) ([]models.Backup, int64, error) {
	return m.backups, m.total, nil
}

func (m *MockBackupService) CreateBackup(requirementID uint64) (*models.Backup, error) {
	if m.createErr != nil {
		return nil, m.createErr
	}
	return &models.Backup{ID: 1, RequirementID: requirementID}, nil
}

func (m *MockBackupService) RestoreBackup(backupID uint64) error {
	return m.restoreErr
}

func (m *MockBackupService) DeleteBackup(backupID uint64) error {
	return m.deleteErr
}

func (m *MockBackupService) GetBackupByID(id uint64) (*models.Backup, error) {
	return nil, nil
}

func (m *MockBackupService) GetSchedule(requirementID uint64) (*models.BackupSchedule, error) {
	if m.scheduleErr != nil {
		return nil, m.scheduleErr
	}
	return m.schedule, nil
}

func (m *MockBackupService) UpdateSchedule(requirementID uint64, enabled bool, intervalType string, intervalValue, retentionCount int) error {
	return m.updateSchErr
}

func (m *MockBackupService) DeleteSchedule(requirementID uint64) error {
	return nil
}

func (m *MockBackupService) StartScheduler() {}

func (m *MockBackupService) StopScheduler() {}

var _ services.BackupService = (*MockBackupService)(nil)

func setupBackupTest(mockService *MockBackupService) (*BackupHandler, *gin.Engine) {
	logger := zap.NewNop()
	handler := NewBackupHandler(logger, mockService)
	router := gin.New()
	return handler, router
}

func TestBackupHandler_List(t *testing.T) {
	tests := []struct {
		name       string
		reqID      string
		expectCode int
		mock       *MockBackupService
	}{
		{
			name:       "获取备份列表成功",
			reqID:      "1",
			expectCode: http.StatusOK,
			mock: &MockBackupService{
				backups: []models.Backup{{ID: 1, RequirementID: 1}},
				total:   1,
			},
		},
		{
			name:       "无效的需求ID",
			reqID:      "abc",
			expectCode: http.StatusBadRequest,
			mock:       &MockBackupService{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler, router := setupBackupTest(tt.mock)
			router.GET("/requirements/:id/backups", handler.List)

			req := httptest.NewRequest(http.MethodGet, "/requirements/"+tt.reqID+"/backups", nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != tt.expectCode {
				t.Errorf("期望状态码 %d, 实际 %d", tt.expectCode, w.Code)
			}
		})
	}
}

func TestBackupHandler_Create(t *testing.T) {
	tests := []struct {
		name       string
		reqID      string
		expectCode int
		mock       *MockBackupService
	}{
		{
			name:       "创建备份成功",
			reqID:      "1",
			expectCode: http.StatusOK,
			mock:       &MockBackupService{},
		},
		{
			name:       "无效的需求ID",
			reqID:      "abc",
			expectCode: http.StatusBadRequest,
			mock:       &MockBackupService{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler, router := setupBackupTest(tt.mock)
			router.POST("/requirements/:id/backups/create", handler.Create)

			req := httptest.NewRequest(http.MethodPost, "/requirements/"+tt.reqID+"/backups/create", nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != tt.expectCode {
				t.Errorf("期望状态码 %d, 实际 %d", tt.expectCode, w.Code)
			}
		})
	}
}

func TestBackupHandler_Restore(t *testing.T) {
	tests := []struct {
		name       string
		backupID   string
		expectCode int
		mock       *MockBackupService
	}{
		{
			name:       "恢复备份成功",
			backupID:   "1",
			expectCode: http.StatusOK,
			mock:       &MockBackupService{},
		},
		{
			name:       "无效的备份ID",
			backupID:   "abc",
			expectCode: http.StatusBadRequest,
			mock:       &MockBackupService{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler, router := setupBackupTest(tt.mock)
			router.POST("/requirements/:id/backups/:backupId/restore", handler.Restore)

			req := httptest.NewRequest(http.MethodPost, "/requirements/1/backups/"+tt.backupID+"/restore", nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != tt.expectCode {
				t.Errorf("期望状态码 %d, 实际 %d", tt.expectCode, w.Code)
			}
		})
	}
}

func TestBackupHandler_Delete(t *testing.T) {
	tests := []struct {
		name       string
		backupID   string
		expectCode int
		mock       *MockBackupService
	}{
		{
			name:       "删除备份成功",
			backupID:   "1",
			expectCode: http.StatusOK,
			mock:       &MockBackupService{},
		},
		{
			name:       "无效的备份ID",
			backupID:   "abc",
			expectCode: http.StatusBadRequest,
			mock:       &MockBackupService{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler, router := setupBackupTest(tt.mock)
			router.POST("/requirements/:id/backups/:backupId/delete", handler.Delete)

			req := httptest.NewRequest(http.MethodPost, "/requirements/1/backups/"+tt.backupID+"/delete", nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != tt.expectCode {
				t.Errorf("期望状态码 %d, 实际 %d", tt.expectCode, w.Code)
			}
		})
	}
}

func TestBackupHandler_GetSchedule(t *testing.T) {
	tests := []struct {
		name       string
		reqID      string
		expectCode int
		mock       *MockBackupService
	}{
		{
			name:       "获取计划成功",
			reqID:      "1",
			expectCode: http.StatusOK,
			mock: &MockBackupService{
				schedule: &models.BackupSchedule{ID: 1, RequirementID: 1, Enabled: true},
			},
		},
		{
			name:       "无计划返回null",
			reqID:      "1",
			expectCode: http.StatusOK,
			mock: &MockBackupService{
				scheduleErr: gorm.ErrRecordNotFound,
			},
		},
		{
			name:       "无效的需求ID",
			reqID:      "abc",
			expectCode: http.StatusBadRequest,
			mock:       &MockBackupService{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler, router := setupBackupTest(tt.mock)
			router.GET("/requirements/:id/backups/schedule", handler.GetSchedule)

			req := httptest.NewRequest(http.MethodGet, "/requirements/"+tt.reqID+"/backups/schedule", nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != tt.expectCode {
				t.Errorf("期望状态码 %d, 实际 %d", tt.expectCode, w.Code)
			}
		})
	}
}

func TestBackupHandler_UpdateSchedule(t *testing.T) {
	tests := []struct {
		name       string
		reqID      string
		body       string
		expectCode int
		mock       *MockBackupService
	}{
		{
			name:       "更新计划成功",
			reqID:      "1",
			body:       `{"enabled":true,"intervalType":"hour","intervalValue":1,"retentionCount":10}`,
			expectCode: http.StatusOK,
			mock:       &MockBackupService{},
		},
		{
			name:       "无效的间隔类型",
			reqID:      "1",
			body:       `{"enabled":true,"intervalType":"day","intervalValue":1,"retentionCount":10}`,
			expectCode: http.StatusBadRequest,
			mock:       &MockBackupService{},
		},
		{
			name:       "间隔值必须大于0",
			reqID:      "1",
			body:       `{"enabled":true,"intervalType":"hour","intervalValue":0,"retentionCount":10}`,
			expectCode: http.StatusBadRequest,
			mock:       &MockBackupService{},
		},
		{
			name:       "无效的需求ID",
			reqID:      "abc",
			body:       `{"enabled":true,"intervalType":"hour","intervalValue":1,"retentionCount":10}`,
			expectCode: http.StatusBadRequest,
			mock:       &MockBackupService{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler, router := setupBackupTest(tt.mock)
			router.POST("/requirements/:id/backups/schedule/update", handler.UpdateSchedule)

			req := httptest.NewRequest(http.MethodPost, "/requirements/"+tt.reqID+"/backups/schedule/update", bytes.NewBufferString(tt.body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != tt.expectCode {
				t.Errorf("期望状态码 %d, 实际 %d, 响应: %s", tt.expectCode, w.Code, w.Body.String())
			}
		})
	}
}

func TestBackupHandler_DisableSchedule(t *testing.T) {
	tests := []struct {
		name       string
		reqID      string
		expectCode int
		mock       *MockBackupService
	}{
		{
			name:       "禁用计划成功",
			reqID:      "1",
			expectCode: http.StatusOK,
			mock: &MockBackupService{
				schedule: &models.BackupSchedule{ID: 1, RequirementID: 1, Enabled: true, IntervalType: "hour", IntervalValue: 1},
			},
		},
		{
			name:       "无计划时也返回成功",
			reqID:      "1",
			expectCode: http.StatusOK,
			mock:       &MockBackupService{schedule: nil, scheduleErr: gorm.ErrRecordNotFound},
		},
		{
			name:       "无效的需求ID",
			reqID:      "abc",
			expectCode: http.StatusBadRequest,
			mock:       &MockBackupService{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler, router := setupBackupTest(tt.mock)
			router.POST("/requirements/:id/backups/schedule/disable", handler.DisableSchedule)

			req := httptest.NewRequest(http.MethodPost, "/requirements/"+tt.reqID+"/backups/schedule/disable", nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != tt.expectCode {
				t.Errorf("期望状态码 %d, 实际 %d", tt.expectCode, w.Code)
			}
		})
	}
}
