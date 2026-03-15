package services

import (
	"testing"
	"time"

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

func TestBackupService_GetBackupByID(t *testing.T) {
	service, mock := setupBackupServiceTest(t)

	t.Run("获取备份成功", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{
			"id", "requirement_id", "backup_type", "data_snapshot", "task_count", "status", "created_at",
		}).AddRow(1, 1, "full", "{}", 5, "success", time.Now())

		mock.ExpectQuery("SELECT \\* FROM `task_backup`").
			WillReturnRows(rows)

		backup, err := service.GetBackupByID(1)
		if err != nil {
			t.Errorf("获取备份失败：%v", err)
		}
		if backup.ID != 1 {
			t.Errorf("期望 ID 为 1, 实际 %d", backup.ID)
		}
	})

	t.Run("备份不存在", func(t *testing.T) {
		mock.ExpectQuery("SELECT \\* FROM `task_backup`").
			WillReturnRows(sqlmock.NewRows([]string{"id"}))

		_, err := service.GetBackupByID(999)
		if err == nil {
			t.Error("期望返回错误，但成功了")
		}
	})
}

func TestBackupService_GetBackups(t *testing.T) {
	service, mock := setupBackupServiceTest(t)

	t.Run("获取备份列表成功", func(t *testing.T) {
		// COUNT 查询
		mock.ExpectQuery("SELECT count\\(\\*\\) FROM `task_backup`").
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(2))

		// 数据查询
		rows := sqlmock.NewRows([]string{
			"id", "requirement_id", "backup_type", "status", "created_at",
		}).AddRow(1, 1, "full", "success", time.Now()).
			AddRow(2, 1, "full", "success", time.Now().Add(-time.Hour))

		mock.ExpectQuery("SELECT \\* FROM `task_backup`").
			WillReturnRows(rows)

		backups, total, err := service.GetBackups(1, 1, 10)
		if err != nil {
			t.Errorf("获取备份列表失败：%v", err)
		}
		if total != 2 {
			t.Errorf("期望总数 2, 实际 %d", total)
		}
		if len(backups) != 2 {
			t.Errorf("期望 2 条记录，实际 %d", len(backups))
		}
	})
}

func TestBackupService_GetSchedule(t *testing.T) {
	service, mock := setupBackupServiceTest(t)

	t.Run("获取计划成功", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{
			"id", "requirement_id", "interval_type", "interval_value", "last_backup_at", "created_at",
		}).AddRow(1, 1, "daily", 1, time.Now(), time.Now())

		mock.ExpectQuery("SELECT \\* FROM `task_backup_schedule`").
			WillReturnRows(rows)

		schedule, err := service.GetSchedule(1)
		if err != nil {
			t.Errorf("获取计划失败：%v", err)
		}
		if schedule.RequirementID != 1 {
			t.Errorf("期望需求 ID 为 1, 实际 %d", schedule.RequirementID)
		}
	})

	t.Run("计划不存在", func(t *testing.T) {
		mock.ExpectQuery("SELECT \\* FROM `task_backup_schedule`").
			WillReturnRows(sqlmock.NewRows([]string{"id"}))

		schedule, err := service.GetSchedule(999)
		if err == nil {
			t.Error("期望返回错误，但成功了")
		}
		if schedule != nil {
			t.Error("期望返回 nil")
		}
	})
}
