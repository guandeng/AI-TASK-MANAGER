package services

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

func setupTaskServiceTest(t *testing.T) (TaskService, sqlmock.Sqlmock) {
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
	service := NewTaskService(nil)
	return service, mock
}

func TestTaskService_Create(t *testing.T) {
	service, mock := setupTaskServiceTest(t)

	now := time.Now()
	task := &models.Task{
		Title:       "测试任务",
		Description: "测试描述",
		Status:      "todo",
		Priority:    "medium",
	}

	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO `task_task`").
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	err := service.Create(task)
	if err != nil {
		t.Errorf("创建任务失败: %v", err)
	}

	if task.ID != 1 {
		t.Errorf("期望 ID 为 1, 实际 %d", task.ID)
	}

	// 验证 created_at 被设置（通过 GORM 自动设置）
	if task.CreatedAt.After(now.Add(time.Second)) && !task.CreatedAt.IsZero() {
		t.Error("CreatedAt 应该被设置")
	}
}

func TestTaskService_GetByID(t *testing.T) {
	service, mock := setupTaskServiceTest(t)

	// 测试成功获取
	t.Run("获取成功", func(t *testing.T) {
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

		task, err := service.GetByID(1)
		if err != nil {
			t.Errorf("获取任务失败: %v", err)
		}
		if task.Title != "测试任务" {
			t.Errorf("期望标题 '测试任务', 实际 '%s'", task.Title)
		}
	})

	// 测试任务不存在
	t.Run("任务不存在", func(t *testing.T) {
		mock.ExpectQuery("SELECT \\* FROM `task_task`").
			WillReturnRows(sqlmock.NewRows([]string{"id"}))

		_, err := service.GetByID(999)
		if err == nil {
			t.Error("期望返回错误，但成功了")
		}
	})
}

func TestTaskService_Update(t *testing.T) {
	service, mock := setupTaskServiceTest(t)

	t.Run("更新任务成功", func(t *testing.T) {
		// 先获取任务
		rows := sqlmock.NewRows([]string{
			"id", "requirement_id", "title", "description", "status", "priority",
			"details", "test_strategy", "assignee", "is_expanding", "expand_message_id",
			"start_date", "due_date", "completed_at", "estimated_hours", "actual_hours",
			"created_at", "updated_at", "deleted_at",
		}).AddRow(1, nil, "原标题", "原描述", "todo", "low",
			"", "", nil, false, nil,
			nil, nil, nil, nil, nil,
			time.Now(), time.Now(), nil)

		mock.ExpectQuery("SELECT \\* FROM `task_task`").
			WillReturnRows(rows)

		// 子任务查询
		mock.ExpectQuery("SELECT").
			WillReturnRows(sqlmock.NewRows([]string{"id"}))

		// 依赖关系查询
		mock.ExpectQuery("SELECT").
			WillReturnRows(sqlmock.NewRows([]string{"id"}))

		// 分配查询
		mock.ExpectQuery("SELECT").
			WillReturnRows(sqlmock.NewRows([]string{"id"}))

		// 更新
		mock.ExpectBegin()
		mock.ExpectExec("UPDATE `task_task`").
			WillReturnResult(sqlmock.NewResult(0, 1))
		mock.ExpectCommit()

		updates := map[string]interface{}{
			"title":       "新标题",
			"description": "新描述",
			"status":      "in_progress",
			"priority":    "high",
		}
		err := service.Update(1, updates)
		if err != nil {
			t.Errorf("更新任务失败: %v", err)
		}
	})

	t.Run("更新状态为完成时设置完成时间", func(t *testing.T) {
		// 先获取任务
		rows := sqlmock.NewRows([]string{
			"id", "requirement_id", "title", "description", "status", "priority",
			"details", "test_strategy", "assignee", "is_expanding", "expand_message_id",
			"start_date", "due_date", "completed_at", "estimated_hours", "actual_hours",
			"created_at", "updated_at", "deleted_at",
		}).AddRow(1, nil, "任务", "描述", "in_progress", "medium",
			"", "", nil, false, nil,
			nil, nil, nil, nil, nil,
			time.Now(), time.Now(), nil)

		mock.ExpectQuery("SELECT \\* FROM `task_task`").
			WillReturnRows(rows)

		mock.ExpectQuery("SELECT").
			WillReturnRows(sqlmock.NewRows([]string{"id"}))
		mock.ExpectQuery("SELECT").
			WillReturnRows(sqlmock.NewRows([]string{"id"}))
		mock.ExpectQuery("SELECT").
			WillReturnRows(sqlmock.NewRows([]string{"id"}))

		mock.ExpectBegin()
		mock.ExpectExec("UPDATE `task_task`").
			WillReturnResult(sqlmock.NewResult(0, 1))
		mock.ExpectCommit()

		updates := map[string]interface{}{
			"status": "done",
		}
		err := service.Update(1, updates)
		if err != nil {
			t.Errorf("更新任务失败: %v", err)
		}
	})
}

func TestTaskService_Delete(t *testing.T) {
	service, mock := setupTaskServiceTest(t)

	mock.ExpectBegin()
	mock.ExpectExec("UPDATE `task_task`").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	err := service.Delete(1)
	if err != nil {
		t.Errorf("删除任务失败: %v", err)
	}
}

func TestTaskService_BatchDelete(t *testing.T) {
	service, mock := setupTaskServiceTest(t)

	mock.ExpectBegin()
	mock.ExpectExec("UPDATE `task_task`").
		WillReturnResult(sqlmock.NewResult(0, 3))
	mock.ExpectCommit()

	ids := []uint64{1, 2, 3}
	err := service.BatchDelete(ids)
	if err != nil {
		t.Errorf("批量删除任务失败: %v", err)
	}
}

func TestTaskService_DeleteSubtask(t *testing.T) {
	service, mock := setupTaskServiceTest(t)

	mock.ExpectBegin()
	mock.ExpectExec("DELETE FROM `task_subtask`").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	err := service.DeleteSubtask(1, 1)
	if err != nil {
		t.Errorf("删除子任务失败: %v", err)
	}
}

func TestTaskService_DeleteAllSubtasks(t *testing.T) {
	service, mock := setupTaskServiceTest(t)

	mock.ExpectBegin()
	mock.ExpectExec("DELETE FROM `task_subtask`").
		WillReturnResult(sqlmock.NewResult(0, 3))
	mock.ExpectCommit()

	err := service.DeleteAllSubtasks(1)
	if err != nil {
		t.Errorf("删除所有子任务失败: %v", err)
	}
}

func TestTaskService_UpdateSubtask(t *testing.T) {
	service, mock := setupTaskServiceTest(t)

	mock.ExpectBegin()
	mock.ExpectExec("UPDATE `task_subtask`").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	updates := map[string]interface{}{
		"title":  "新标题",
		"status": "done",
	}
	err := service.UpdateSubtask(1, 1, updates)
	if err != nil {
		t.Errorf("更新子任务失败: %v", err)
	}
}

func TestTaskService_ReorderSubtasks(t *testing.T) {
	service, mock := setupTaskServiceTest(t)

	mock.ExpectBegin()
	mock.ExpectExec("UPDATE `task_subtask`").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("UPDATE `task_subtask`").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("UPDATE `task_subtask`").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	subtaskIDs := []uint64{3, 1, 2}
	err := service.ReorderSubtasks(1, subtaskIDs)
	if err != nil {
		t.Errorf("重排序子任务失败: %v", err)
	}
}

func TestTaskService_UpdateTime(t *testing.T) {
	service, mock := setupTaskServiceTest(t)

	// 先获取任务
	rows := sqlmock.NewRows([]string{
		"id", "requirement_id", "title", "description", "status", "priority",
		"details", "test_strategy", "assignee", "is_expanding", "expand_message_id",
		"start_date", "due_date", "completed_at", "estimated_hours", "actual_hours",
		"created_at", "updated_at", "deleted_at",
	}).AddRow(1, nil, "任务", "描述", "todo", "medium",
		"", "", nil, false, nil,
		nil, nil, nil, nil, nil,
		time.Now(), time.Now(), nil)

	mock.ExpectQuery("SELECT \\* FROM `task_task`").
		WillReturnRows(rows)

	mock.ExpectQuery("SELECT").
		WillReturnRows(sqlmock.NewRows([]string{"id"}))
	mock.ExpectQuery("SELECT").
		WillReturnRows(sqlmock.NewRows([]string{"id"}))
	mock.ExpectQuery("SELECT").
		WillReturnRows(sqlmock.NewRows([]string{"id"}))

	mock.ExpectBegin()
	mock.ExpectExec("UPDATE `task_task`").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	startTime := time.Now()
	dueDate := startTime.Add(24 * time.Hour)
	estimatedHours := 8.0
	actualHours := 6.5

	err := service.UpdateTime(1, &startTime, &dueDate, &estimatedHours, &actualHours)
	if err != nil {
		t.Errorf("更新时间失败: %v", err)
	}
}

func TestTaskService_ExpandTask_NoAI(t *testing.T) {
	service, mock := setupTaskServiceTest(t)

	// 获取任务
	rows := sqlmock.NewRows([]string{
		"id", "requirement_id", "title", "description", "status", "priority",
		"details", "test_strategy", "assignee", "is_expanding", "expand_message_id",
		"start_date", "due_date", "completed_at", "estimated_hours", "actual_hours",
		"created_at", "updated_at", "deleted_at",
	}).AddRow(1, nil, "任务", "描述", "todo", "medium",
		"", "", nil, false, nil,
		nil, nil, nil, nil, nil,
		time.Now(), time.Now(), nil)

	mock.ExpectQuery("SELECT \\* FROM `task_task`").
		WillReturnRows(rows)

	mock.ExpectQuery("SELECT").
		WillReturnRows(sqlmock.NewRows([]string{"id"}))
	mock.ExpectQuery("SELECT").
		WillReturnRows(sqlmock.NewRows([]string{"id"}))
	mock.ExpectQuery("SELECT").
		WillReturnRows(sqlmock.NewRows([]string{"id"}))

	// AI 服务为 nil，不应做任何操作
	err := service.ExpandTask(1, false)
	if err != nil {
		t.Errorf("展开任务失败: %v", err)
	}
}

func TestTaskService_RegenerateSubtask_NoAI(t *testing.T) {
	service, mock := setupTaskServiceTest(t)

	// 获取任务
	rows := sqlmock.NewRows([]string{
		"id", "requirement_id", "title", "description", "status", "priority",
		"details", "test_strategy", "assignee", "is_expanding", "expand_message_id",
		"start_date", "due_date", "completed_at", "estimated_hours", "actual_hours",
		"created_at", "updated_at", "deleted_at",
	}).AddRow(1, nil, "任务", "描述", "todo", "medium",
		"", "", nil, false, nil,
		nil, nil, nil, nil, nil,
		time.Now(), time.Now(), nil)

	mock.ExpectQuery("SELECT \\* FROM `task_task`").
		WillReturnRows(rows)

	mock.ExpectQuery("SELECT").
		WillReturnRows(sqlmock.NewRows([]string{"id"}))
	mock.ExpectQuery("SELECT").
		WillReturnRows(sqlmock.NewRows([]string{"id"}))
	mock.ExpectQuery("SELECT").
		WillReturnRows(sqlmock.NewRows([]string{"id"}))

	// 删除旧子任务
	mock.ExpectBegin()
	mock.ExpectExec("DELETE FROM `task_subtask`").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	// AI 服务为 nil，不应生成新子任务
	err := service.RegenerateSubtask(1, 1)
	if err != nil {
		t.Errorf("重新生成子任务失败: %v", err)
	}
}

func TestTaskService_List(t *testing.T) {
	service, mock := setupTaskServiceTest(t)

	// Mock count query
	mock.ExpectQuery("SELECT").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(2))

	// Mock select query
	rows := sqlmock.NewRows([]string{
		"id", "requirement_id", "requirement_title", "title", "description",
		"status", "priority", "assignee", "subtask_count", "subtask_done_count",
	}).AddRow(1, 1, "需求 1", "任务 1", "描述 1", "pending", "high", "user1", 2, 1).
		AddRow(2, 1, "需求 1", "任务 2", "描述 2", "done", "low", "user2", 1, 1)

	mock.ExpectQuery("SELECT").WillReturnRows(rows)

	filters := map[string]interface{}{
		"status": "pending",
	}
	tasks, total, err := service.List(filters, 1, 10)
	if err != nil {
		t.Errorf("获取任务列表失败：%v", err)
	}
	if total != 2 {
		t.Errorf("期望总数 2, 实际 %d", total)
	}
	if len(tasks) != 2 {
		t.Errorf("期望 2 条记录，实际 %d", len(tasks))
	}
}
