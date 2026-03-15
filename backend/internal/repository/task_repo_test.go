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

func setupTaskRepoTest(t *testing.T) (TaskRepository, sqlmock.Sqlmock) {
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
	return NewTaskRepository(), mock
}

func TestTaskRepository_Create(t *testing.T) {
	repo, mock := setupTaskRepoTest(t)

	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO `task_task`").
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	task := &models.Task{
		Title:       "测试任务",
		Description: "测试描述",
		Status:      "todo",
		Priority:    "medium",
	}

	err := repo.Create(task)
	if err != nil {
		t.Errorf("创建任务失败: %v", err)
	}
}

func TestTaskRepository_GetByID(t *testing.T) {
	repo, mock := setupTaskRepoTest(t)

	t.Run("获取任务成功", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{
			"id", "requirement_id", "title", "description", "status", "priority",
			"details", "test_strategy", "assignee", "is_expanding", "expand_message_id",
			"start_date", "due_date", "completed_at", "estimated_hours", "actual_hours",
			"created_at", "updated_at", "deleted_at",
		}).AddRow(1, nil, "测试任务", "描述", "todo", "medium",
			"", "", nil, false, nil,
			nil, nil, nil, nil, nil,
			time.Now(), time.Now(), nil)

		mock.ExpectQuery("SELECT \\* FROM `task_task`").
			WillReturnRows(rows)

		// 子任务查询
		mock.ExpectQuery("SELECT").
			WillReturnRows(sqlmock.NewRows([]string{"id", "task_id", "title", "status"}))

		// 依赖关系查询
		mock.ExpectQuery("SELECT").
			WillReturnRows(sqlmock.NewRows([]string{"id", "task_id", "depends_on_task_id"}))

		// 分配查询
		mock.ExpectQuery("SELECT").
			WillReturnRows(sqlmock.NewRows([]string{"id", "task_id", "member_id"}))

		task, err := repo.GetByID(1)
		if err != nil {
			t.Errorf("获取任务失败: %v", err)
		}
		if task.Title != "测试任务" {
			t.Errorf("期望标题 '测试任务', 实际 '%s'", task.Title)
		}
	})

	t.Run("任务不存在", func(t *testing.T) {
		mock.ExpectQuery("SELECT \\* FROM `task_task`").
			WillReturnRows(sqlmock.NewRows([]string{"id"}))

		_, err := repo.GetByID(999)
		if err == nil {
			t.Error("期望返回错误，但成功了")
		}
	})
}

func TestTaskRepository_Update(t *testing.T) {
	repo, mock := setupTaskRepoTest(t)

	mock.ExpectBegin()
	mock.ExpectExec("UPDATE `task_task`").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	task := &models.Task{
		ID:          1,
		Title:       "更新后的标题",
		Description: "更新后的描述",
		Status:      "in_progress",
		Priority:    "high",
	}

	err := repo.Update(task)
	if err != nil {
		t.Errorf("更新任务失败: %v", err)
	}
}

func TestTaskRepository_Delete(t *testing.T) {
	repo, mock := setupTaskRepoTest(t)

	mock.ExpectBegin()
	mock.ExpectExec("UPDATE `task_task`").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	err := repo.Delete(1)
	if err != nil {
		t.Errorf("删除任务失败: %v", err)
	}
}

func TestTaskRepository_BatchDelete(t *testing.T) {
	repo, mock := setupTaskRepoTest(t)

	mock.ExpectBegin()
	mock.ExpectExec("UPDATE `task_task`").
		WillReturnResult(sqlmock.NewResult(0, 3))
	mock.ExpectCommit()

	ids := []uint64{1, 2, 3}
	err := repo.BatchDelete(ids)
	if err != nil {
		t.Errorf("批量删除任务失败: %v", err)
	}
}

func TestTaskRepository_GetSubtasks(t *testing.T) {
	repo, mock := setupTaskRepoTest(t)

	rows := sqlmock.NewRows([]string{
		"id", "task_id", "title", "description", "status", "priority", "sort_order",
	}).AddRow(1, 1, "子任务1", "描述1", "todo", "medium", 0).
		AddRow(2, 1, "子任务2", "描述2", "done", "high", 1)

	mock.ExpectQuery("SELECT").
		WillReturnRows(rows)

	subtasks, err := repo.GetSubtasks(1)
	if err != nil {
		t.Errorf("获取子任务失败: %v", err)
	}
	if len(subtasks) != 2 {
		t.Errorf("期望 2 个子任务, 实际 %d", len(subtasks))
	}
}

func TestTaskRepository_CreateSubtask(t *testing.T) {
	repo, mock := setupTaskRepoTest(t)

	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO `task_subtask`").
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	subtask := &models.Subtask{
		TaskID:  1,
		Title:   "新子任务",
		Status:  "todo",
		SortOrder: 0,
	}

	err := repo.CreateSubtask(subtask)
	if err != nil {
		t.Errorf("创建子任务失败: %v", err)
	}
}

func TestTaskRepository_UpdateSubtask(t *testing.T) {
	repo, mock := setupTaskRepoTest(t)

	mock.ExpectBegin()
	mock.ExpectExec("UPDATE `task_subtask`").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	subtask := &models.Subtask{
		ID:      1,
		TaskID:  1,
		Title:   "更新后的子任务",
		Status:  "done",
	}

	err := repo.UpdateSubtask(subtask)
	if err != nil {
		t.Errorf("更新子任务失败: %v", err)
	}
}

func TestTaskRepository_UpdateSubtaskWithMap(t *testing.T) {
	repo, mock := setupTaskRepoTest(t)

	mock.ExpectBegin()
	mock.ExpectExec("UPDATE `task_subtask`").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	updates := map[string]interface{}{
		"title":  "新标题",
		"status": "done",
	}
	err := repo.UpdateSubtaskWithMap(1, 1, updates)
	if err != nil {
		t.Errorf("更新子任务失败: %v", err)
	}
}

func TestTaskRepository_DeleteSubtask(t *testing.T) {
	repo, mock := setupTaskRepoTest(t)

	mock.ExpectBegin()
	mock.ExpectExec("DELETE FROM `task_subtask`").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	err := repo.DeleteSubtask(1, 1)
	if err != nil {
		t.Errorf("删除子任务失败: %v", err)
	}
}

func TestTaskRepository_DeleteAllSubtasks(t *testing.T) {
	repo, mock := setupTaskRepoTest(t)

	mock.ExpectBegin()
	mock.ExpectExec("DELETE FROM `task_subtask`").
		WillReturnResult(sqlmock.NewResult(0, 3))
	mock.ExpectCommit()

	err := repo.DeleteAllSubtasks(1)
	if err != nil {
		t.Errorf("删除所有子任务失败: %v", err)
	}
}

func TestTaskRepository_ReorderSubtasks(t *testing.T) {
	repo, mock := setupTaskRepoTest(t)

	mock.ExpectBegin()
	mock.ExpectExec("UPDATE `task_subtask`").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("UPDATE `task_subtask`").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("UPDATE `task_subtask`").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	subtaskIDs := []uint64{3, 1, 2}
	err := repo.ReorderSubtasks(1, subtaskIDs)
	if err != nil {
		t.Errorf("重排序子任务失败: %v", err)
	}
}

func TestTaskRepository_GetDependencies(t *testing.T) {
	repo, mock := setupTaskRepoTest(t)

	rows := sqlmock.NewRows([]string{
		"id", "task_id", "depends_on_task_id", "dependency_type",
	}).AddRow(1, 1, 2, "finish_to_start").
		AddRow(2, 1, 3, "finish_to_start")

	mock.ExpectQuery("SELECT").
		WillReturnRows(rows)

	deps, err := repo.GetDependencies(1)
	if err != nil {
		t.Errorf("获取依赖关系失败: %v", err)
	}
	if len(deps) != 2 {
		t.Errorf("期望 2 个依赖关系, 实际 %d", len(deps))
	}
}
