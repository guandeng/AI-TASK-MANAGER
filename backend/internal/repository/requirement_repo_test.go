package repository

import (
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/ai-task-manager/backend/internal/database"
	"github.com/gin-gonic/gin"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func setupRequirementRepoTest(t *testing.T) (RequirementRepository, sqlmock.Sqlmock) {
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
	return NewRequirementRepository(), mock
}

func TestRequirementRepository_GetByID(t *testing.T) {
	repo, mock := setupRequirementRepoTest(t)

	t.Run("获取需求成功", func(t *testing.T) {
		now := time.Now()
		rows := sqlmock.NewRows([]string{
			"id", "title", "content", "status", "priority", "documents", "created_at", "updated_at",
		}).AddRow(1, "测试需求", "需求内容", "active", "high", "[]", now, now)

		// 需要导入 models，但这个简单测试实际上不需要

		mock.ExpectQuery("SELECT \\* FROM `task_requirement`").
			WillReturnRows(rows)

		req, err := repo.GetByID(1)
		if err != nil {
			t.Errorf("获取需求失败: %v", err)
		}
		if req.Title != "测试需求" {
			t.Errorf("期望标题 '测试需求', 实际 '%s'", req.Title)
		}
	})

	t.Run("需求不存在", func(t *testing.T) {
		mock.ExpectQuery("SELECT \\* FROM `task_requirement`").
			WillReturnRows(sqlmock.NewRows([]string{"id"}))

		_, err := repo.GetByID(999)
		if err == nil {
			t.Error("期望返回错误，但成功了")
		}
	})
}
