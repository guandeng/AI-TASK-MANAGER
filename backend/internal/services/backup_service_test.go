package services

import (
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

func setupBackupServiceTest(t *testing.T) (BackupService, sqlmock.Sqlmock) {
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
	logger := zap.NewNop()
	service := NewBackupService(logger)
	return service, mock
}

func TestBackupService_DeleteBackup(t *testing.T) {
	service, mock := setupBackupServiceTest(t)

	t.Run("删除备份成功", func(t *testing.T) {
		mock.ExpectBegin()
		mock.ExpectExec("DELETE").WillReturnResult(sqlmock.NewResult(0, 1))
		mock.ExpectCommit()

		err := service.DeleteBackup(1)
		if err != nil {
			t.Errorf("删除备份失败: %v", err)
		}
	})
}

func TestBackupService_DeleteSchedule(t *testing.T) {
	service, mock := setupBackupServiceTest(t)

	t.Run("删除计划成功", func(t *testing.T) {
		mock.ExpectBegin()
		mock.ExpectExec("DELETE").WillReturnResult(sqlmock.NewResult(0, 1))
		mock.ExpectCommit()

		err := service.DeleteSchedule(1)
		if err != nil {
			t.Errorf("删除计划失败: %v", err)
		}
	})
}
