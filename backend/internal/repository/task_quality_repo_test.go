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

func setupTaskQualityRepoTest(t *testing.T) (TaskQualityRepository, sqlmock.Sqlmock) {
	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("创建 mock 数据库失败：%v", err)
	}

	gormDB, err := gorm.Open(mysql.New(mysql.Config{
		Conn:                      sqlDB,
		SkipInitializeWithVersion: true,
	}), &gorm.Config{})
	if err != nil {
		t.Fatalf("创建 gorm 连接失败：%v", err)
	}

	database.DB = gormDB
	return NewTaskQualityRepository(), mock
}

func TestTaskQualityRepository_Create(t *testing.T) {
	repo, mock := setupTaskQualityRepoTest(t)

	t.Run("创建评分成功", func(t *testing.T) {
		now := time.Now()
		score := &models.TaskQualityScore{
			TaskID:           1,
			Version:          1,
			TotalScore:       85.0,
			ClarityScore:     8,
			CompletenessScore: 9,
			StructureScore:   8,
			ActionabilityScore: 8,
			ConsistencyScore: 9,
			Evaluation:       `{"strengths":["测试"]}`,
			TaskSnapshot:     `{"id":1}`,
			AIProvider:       "test",
			CreatedAt:        now,
		}

		mock.ExpectBegin()
		mock.ExpectExec("INSERT INTO `task_quality_score`").
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		err := repo.Create(score)
		if err != nil {
			t.Errorf("创建评分失败：%v", err)
		}
	})
}

func TestTaskQualityRepository_GetByID(t *testing.T) {
	repo, mock := setupTaskQualityRepoTest(t)

	t.Run("获取评分成功", func(t *testing.T) {
		now := time.Now()
		rows := sqlmock.NewRows([]string{
			"id", "task_id", "version", "total_score", "clarity_score", "completeness_score", "structure_score", "actionability_score", "consistency_score", "evaluation", "task_snapshot", "ai_provider", "created_at",
		}).AddRow(1, 1, 1, 85.0, 8, 9, 8, 8, 9, `{"strengths":["测试"]}`, `{"id":1}`, "test", now)

		mock.ExpectQuery("SELECT \\* FROM `task_quality_score` WHERE id = \\?").
			WillReturnRows(rows)

		score, err := repo.GetByID(1)
		if err != nil {
			t.Errorf("获取评分失败：%v", err)
		}
		if score.TotalScore != 85 {
			t.Errorf("期望分数 85, 实际 %.2f", score.TotalScore)
		}
	})

	t.Run("评分不存在", func(t *testing.T) {
		mock.ExpectQuery("SELECT \\* FROM `task_quality_scores` WHERE id = \\?").
			WillReturnRows(sqlmock.NewRows([]string{"id"}))

		_, err := repo.GetByID(999)
		if err == nil {
			t.Error("期望返回错误，但成功了")
		}
	})
}

func TestTaskQualityRepository_GetByTaskID(t *testing.T) {
	repo, mock := setupTaskQualityRepoTest(t)

	t.Run("获取评分列表成功", func(t *testing.T) {
		now := time.Now()
		rows := sqlmock.NewRows([]string{
			"id", "task_id", "version", "total_score", "clarity_score", "completeness_score", "structure_score", "actionability_score", "consistency_score", "evaluation", "task_snapshot", "ai_provider", "created_at",
		}).AddRow(1, 1, 2, 90.0, 9, 9, 9, 9, 9, `{"strengths":["测试"]}`, `{"id":1}`, "test", now).
			AddRow(2, 1, 1, 85.0, 8, 9, 8, 8, 9, `{"strengths":["测试"]}`, `{"id":1}`, "test", now.Add(-time.Hour))

		mock.ExpectQuery("SELECT count\\(\\*\\) FROM `task_quality_score`").
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(2))

		mock.ExpectQuery("SELECT \\* FROM `task_quality_score`").
			WillReturnRows(rows)

		scores, total, err := repo.GetByTaskID(1, 1, 10)
		if err != nil {
			t.Errorf("获取评分列表失败：%v", err)
		}
		if total != 2 {
			t.Errorf("期望总数 2, 实际 %d", total)
		}
		if len(scores) != 2 {
			t.Errorf("期望 2 条记录，实际 %d 条", len(scores))
		}
	})
}

func TestTaskQualityRepository_GetLatestByTaskID(t *testing.T) {
	repo, mock := setupTaskQualityRepoTest(t)

	t.Run("获取最新评分成功", func(t *testing.T) {
		now := time.Now()
		rows := sqlmock.NewRows([]string{
			"id", "task_id", "version", "total_score", "clarity_score", "completeness_score", "structure_score", "actionability_score", "consistency_score", "evaluation", "task_snapshot", "ai_provider", "created_at",
		}).AddRow(1, 1, 2, 90.0, 9, 9, 9, 9, 9, `{"strengths":["测试"]}`, `{"id":1}`, "test", now)

		mock.ExpectQuery("SELECT \\* FROM `task_quality_score` WHERE task_id = \\?").
			WillReturnRows(rows)

		score, err := repo.GetLatestByTaskID(1)
		if err != nil {
			t.Errorf("获取最新评分失败：%v", err)
		}
		if score.Version != 2 {
			t.Errorf("期望版本 2, 实际 %d", score.Version)
		}
	})

	t.Run("评分不存在", func(t *testing.T) {
		mock.ExpectQuery("SELECT \\* FROM `task_quality_score` WHERE task_id = \\?").
			WillReturnRows(sqlmock.NewRows([]string{"id"}))

		_, err := repo.GetLatestByTaskID(999)
		if err == nil {
			t.Error("期望返回错误，但成功了")
		}
	})
}

func TestTaskQualityRepository_Delete(t *testing.T) {
	repo, mock := setupTaskQualityRepoTest(t)

	t.Run("删除评分成功", func(t *testing.T) {
		mock.ExpectBegin()
		mock.ExpectExec("DELETE FROM `task_quality_score`").
			WillReturnResult(sqlmock.NewResult(0, 1))
		mock.ExpectCommit()

		err := repo.Delete(1)
		if err != nil {
			t.Errorf("删除评分失败：%v", err)
		}
	})
}

func TestTaskQualityRepository_GetNextVersion(t *testing.T) {
	repo, mock := setupTaskQualityRepoTest(t)

	t.Run("获取下一个版本号 - 已有记录", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"max_version"}).AddRow(3)

		mock.ExpectQuery("SELECT COALESCE\\(MAX\\(version\\), 0\\)").
			WillReturnRows(rows)

		version, err := repo.GetNextVersion(1)
		if err != nil {
			t.Errorf("获取版本号失败：%v", err)
		}
		if version != 4 {
			t.Errorf("期望版本 4, 实际 %d", version)
		}
	})

	t.Run("获取下一个版本号 - 无记录", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"max_version"}).AddRow(0)

		mock.ExpectQuery("SELECT COALESCE\\(MAX\\(version\\), 0\\)").
			WillReturnRows(rows)

		version, err := repo.GetNextVersion(999)
		if err != nil {
			t.Errorf("获取版本号失败：%v", err)
		}
		if version != 1 {
			t.Errorf("期望版本 1, 实际 %d", version)
		}
	})
}
