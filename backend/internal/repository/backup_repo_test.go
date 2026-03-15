package repository

import (
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/ai-task-manager/backend/internal/database"
	"github.com/ai-task-manager/backend/internal/models"
	"github.com/gin-gonic/gin"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func setupBackupRepoTest(t *testing.T) (BackupRepository, sqlmock.Sqlmock) {
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
	return NewBackupRepository(), mock
}

func TestBackupRepository_CreateBackup(t *testing.T) {
	repo, mock := setupBackupRepoTest(t)

	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO `task_backup`").
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	backup := &models.Backup{
		RequirementID: 1,
		BackupType:    "full",
		DataSnapshot:  `{"data": "test"}`,
		TaskCount:     5,
		Status:        "success",
	}

	err := repo.CreateBackup(backup)
	if err != nil {
		t.Errorf("创建备份失败: %v", err)
	}
}

func TestBackupRepository_GetBackupsByRequirementID(t *testing.T) {
	repo, mock := setupBackupRepoTest(t)

	// COUNT 查询
	mock.ExpectQuery("SELECT").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(2))

	// SELECT 查询
	rows := sqlmock.NewRows([]string{
		"id", "requirement_id", "backup_type", "data_snapshot", "task_count", "status", "created_at",
	}).AddRow(1, 1, "full", "{}", 5, "success", time.Now()).
		AddRow(2, 1, "full", "{}", 3, "success", time.Now())

	mock.ExpectQuery("SELECT").
		WillReturnRows(rows)

	backups, total, err := repo.GetBackupsByRequirementID(1, 1, 10)
	if err != nil {
		t.Errorf("获取备份列表失败: %v", err)
	}
	if total != 2 {
		t.Errorf("期望总数 2, 实际 %d", total)
	}
	if len(backups) != 2 {
		t.Errorf("期望 2 条记录, 实际 %d", len(backups))
	}
}

func TestBackupRepository_GetBackupByID(t *testing.T) {
	repo, mock := setupBackupRepoTest(t)

	t.Run("获取备份成功", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{
			"id", "requirement_id", "backup_type", "data_snapshot", "task_count", "status", "created_at",
		}).AddRow(1, 1, "full", `{"data": "test"}`, 5, "success", time.Now())

		mock.ExpectQuery("SELECT \\* FROM `task_backup`").
			WillReturnRows(rows)

		backup, err := repo.GetBackupByID(1)
		if err != nil {
			t.Errorf("获取备份失败: %v", err)
		}
		if backup.ID != 1 {
			t.Errorf("期望 ID 1, 实际 %d", backup.ID)
		}
	})

	t.Run("备份不存在", func(t *testing.T) {
		mock.ExpectQuery("SELECT \\* FROM `task_backup`").
			WillReturnRows(sqlmock.NewRows([]string{"id"}))

		_, err := repo.GetBackupByID(999)
		if err == nil {
			t.Error("期望返回错误，但成功了")
		}
	})
}

func TestBackupRepository_DeleteBackup(t *testing.T) {
	repo, mock := setupBackupRepoTest(t)

	mock.ExpectBegin()
	mock.ExpectExec("DELETE FROM `task_backup`").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	err := repo.DeleteBackup(1)
	if err != nil {
		t.Errorf("删除备份失败: %v", err)
	}
}

func TestBackupRepository_GetScheduleByRequirementID(t *testing.T) {
	repo, mock := setupBackupRepoTest(t)

	t.Run("获取计划成功", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{
			"id", "requirement_id", "enabled", "interval_type", "interval_value", "retention_count", "last_backup_at",
		}).AddRow(1, 1, true, "hour", 1, 10, time.Now())

		mock.ExpectQuery("SELECT \\* FROM `task_backup_schedule`").
			WillReturnRows(rows)

		schedule, err := repo.GetScheduleByRequirementID(1)
		if err != nil {
			t.Errorf("获取计划失败: %v", err)
		}
		if !schedule.Enabled {
			t.Error("期望启用状态")
		}
	})

	t.Run("计划不存在", func(t *testing.T) {
		mock.ExpectQuery("SELECT \\* FROM `task_backup_schedule`").
			WillReturnRows(sqlmock.NewRows([]string{"id"}))

		_, err := repo.GetScheduleByRequirementID(999)
		if err == nil {
			t.Error("期望返回错误，但成功了")
		}
	})
}

func TestBackupRepository_UpsertSchedule(t *testing.T) {
	repo, mock := setupBackupRepoTest(t)

	t.Run("插入新计划", func(t *testing.T) {
		// 检查是否存在
		mock.ExpectQuery("SELECT \\* FROM `task_backup_schedule`").
			WillReturnRows(sqlmock.NewRows([]string{"id"}))

		// 插入
		mock.ExpectBegin()
		mock.ExpectExec("INSERT INTO `task_backup_schedule`").
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		schedule := &models.BackupSchedule{
			RequirementID:  1,
			Enabled:        true,
			IntervalType:   "hour",
			IntervalValue:  1,
			RetentionCount: 10,
		}
		err := repo.UpsertSchedule(schedule)
		if err != nil {
			t.Errorf("插入计划失败: %v", err)
		}
	})

	t.Run("更新已存在的计划", func(t *testing.T) {
		// 检查是否存在
		rows := sqlmock.NewRows([]string{
			"id", "requirement_id", "enabled", "interval_type", "interval_value", "retention_count",
		}).AddRow(1, 1, false, "minute", 30, 5)

		mock.ExpectQuery("SELECT \\* FROM `task_backup_schedule`").
			WillReturnRows(rows)

		// 更新
		mock.ExpectBegin()
		mock.ExpectExec("UPDATE `task_backup_schedule`").
			WillReturnResult(sqlmock.NewResult(0, 1))
		mock.ExpectCommit()

		schedule := &models.BackupSchedule{
			RequirementID:  1,
			Enabled:        true,
			IntervalType:   "hour",
			IntervalValue:  1,
			RetentionCount: 10,
		}
		err := repo.UpsertSchedule(schedule)
		if err != nil {
			t.Errorf("更新计划失败: %v", err)
		}
	})
}

func TestBackupRepository_DeleteSchedule(t *testing.T) {
	repo, mock := setupBackupRepoTest(t)

	mock.ExpectBegin()
	mock.ExpectExec("DELETE FROM `task_backup_schedule`").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	err := repo.DeleteSchedule(1)
	if err != nil {
		t.Errorf("删除计划失败: %v", err)
	}
}

func TestBackupRepository_UpdateLastBackupAt(t *testing.T) {
	repo, mock := setupBackupRepoTest(t)

	mock.ExpectBegin()
	mock.ExpectExec("UPDATE `task_backup_schedule`").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	err := repo.UpdateLastBackupAt(1)
	if err != nil {
		t.Errorf("更新备份时间失败: %v", err)
	}
}

func TestBackupRepository_GetEnabledSchedules(t *testing.T) {
	repo, mock := setupBackupRepoTest(t)

	rows := sqlmock.NewRows([]string{
		"id", "requirement_id", "enabled", "interval_type", "interval_value", "retention_count",
	}).AddRow(1, 1, true, "hour", 1, 10).
		AddRow(2, 2, true, "day", 1, 5)

	mock.ExpectQuery("SELECT \\* FROM `task_backup_schedule`").
		WillReturnRows(rows)

	schedules, err := repo.GetEnabledSchedules()
	if err != nil {
		t.Errorf("获取启用计划失败: %v", err)
	}
	if len(schedules) != 2 {
		t.Errorf("期望 2 条记录, 实际 %d", len(schedules))
	}
}
