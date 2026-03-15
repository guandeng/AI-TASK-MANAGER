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

func setupMenuRepoTest(t *testing.T) (MenuRepository, sqlmock.Sqlmock) {
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
	return NewMenuRepository(), mock
}

func strPtr(s string) *string {
	return &s
}

func TestMenuRepository_Create(t *testing.T) {
	repo, mock := setupMenuRepoTest(t)

	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO `task_menu`").
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	menu := &models.Menu{
		Key:     "test_menu",
		Title:   "测试菜单",
		Icon:    strPtr("icon-test"),
		Path:    strPtr("/test"),
		Enabled: true,
		Sort:    1,
	}

	err := repo.Create(menu)
	if err != nil {
		t.Errorf("创建菜单失败: %v", err)
	}
}

func TestMenuRepository_GetByKey(t *testing.T) {
	repo, mock := setupMenuRepoTest(t)

	t.Run("获取菜单成功", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{
			"id", "key", "title", "icon", "path", "parent_key", "enabled", "sort", "created_at", "updated_at",
		}).AddRow(1, "test_menu", "测试菜单", "icon-test", "/test", nil, true, 1, time.Now(), time.Now())

		mock.ExpectQuery("SELECT \\* FROM `task_menu`").
			WillReturnRows(rows)

		menu, err := repo.GetByKey("test_menu")
		if err != nil {
			t.Errorf("获取菜单失败: %v", err)
		}
		if menu.Key != "test_menu" {
			t.Errorf("期望 key 'test_menu', 实际 '%s'", menu.Key)
		}
	})

	t.Run("菜单不存在", func(t *testing.T) {
		mock.ExpectQuery("SELECT \\* FROM `task_menu`").
			WillReturnRows(sqlmock.NewRows([]string{"id"}))

		_, err := repo.GetByKey("nonexistent")
		if err == nil {
			t.Error("期望返回错误，但成功了")
		}
	})
}

func TestMenuRepository_List(t *testing.T) {
	repo, mock := setupMenuRepoTest(t)

	// COUNT 查询
	mock.ExpectQuery("SELECT").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(2))

	rows := sqlmock.NewRows([]string{
		"id", "key", "title", "icon", "path", "parent_key", "enabled", "sort", "created_at", "updated_at",
	}).AddRow(1, "menu1", "菜单1", "icon1", "/path1", nil, true, 1, time.Now(), time.Now()).
		AddRow(2, "menu2", "菜单2", "icon2", "/path2", nil, true, 2, time.Now(), time.Now())

	mock.ExpectQuery("SELECT").
		WillReturnRows(rows)

	menus, total, err := repo.List(1, 10)
	if err != nil {
		t.Errorf("获取菜单列表失败: %v", err)
	}
	if total != 2 {
		t.Errorf("期望总数 2, 实际 %d", total)
	}
	if len(menus) != 2 {
		t.Errorf("期望 2 个菜单, 实际 %d", len(menus))
	}
}

func TestMenuRepository_Update(t *testing.T) {
	repo, mock := setupMenuRepoTest(t)

	mock.ExpectBegin()
	mock.ExpectExec("UPDATE `task_menu`").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	menu := &models.Menu{
		ID:      1,
		Key:     "updated_menu",
		Title:   "更新后的菜单",
		Enabled: true,
	}

	err := repo.Update(menu)
	if err != nil {
		t.Errorf("更新菜单失败: %v", err)
	}
}

func TestMenuRepository_Delete(t *testing.T) {
	repo, mock := setupMenuRepoTest(t)

	mock.ExpectBegin()
	// 先删除菜单本身
	mock.ExpectExec("DELETE FROM `task_menu`").
		WillReturnResult(sqlmock.NewResult(0, 1))
	// 查询子菜单
	mock.ExpectQuery("SELECT").
		WillReturnRows(sqlmock.NewRows([]string{"id", "key", "parent_key"}))
	mock.ExpectCommit()

	err := repo.Delete("test_menu")
	if err != nil {
		t.Errorf("删除菜单失败: %v", err)
	}
}

func TestMenuRepository_BatchDelete(t *testing.T) {
	repo, mock := setupMenuRepoTest(t)

	mock.ExpectBegin()
	// 查询子菜单（每个 key）
	mock.ExpectQuery("SELECT").
		WillReturnRows(sqlmock.NewRows([]string{"id", "key", "parent_key"}))
	mock.ExpectQuery("SELECT").
		WillReturnRows(sqlmock.NewRows([]string{"id", "key", "parent_key"}))
	mock.ExpectQuery("SELECT").
		WillReturnRows(sqlmock.NewRows([]string{"id", "key", "parent_key"}))
	mock.ExpectExec("DELETE FROM `task_menu`").
		WillReturnResult(sqlmock.NewResult(0, 3))
	mock.ExpectCommit()

	keys := []string{"menu1", "menu2", "menu3"}
	err := repo.BatchDelete(keys)
	if err != nil {
		t.Errorf("批量删除菜单失败: %v", err)
	}
}

func TestMenuRepository_ToggleEnabled(t *testing.T) {
	repo, mock := setupMenuRepoTest(t)

	mock.ExpectBegin()
	mock.ExpectExec("UPDATE `task_menu`").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	err := repo.ToggleEnabled("test_menu", false)
	if err != nil {
		t.Errorf("切换菜单状态失败: %v", err)
	}
}

func TestMenuRepository_Reorder(t *testing.T) {
	repo, mock := setupMenuRepoTest(t)

	mock.ExpectBegin()
	mock.ExpectExec("UPDATE `task_menu`").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("UPDATE `task_menu`").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	orderData := []map[string]interface{}{
		{"key": "menu1", "order": float64(2)},
		{"key": "menu2", "order": float64(1)},
	}
	err := repo.Reorder(orderData)
	if err != nil {
		t.Errorf("重排序菜单失败: %v", err)
	}
}

func TestMenuRepository_Move(t *testing.T) {
	repo, mock := setupMenuRepoTest(t)

	mock.ExpectBegin()
	mock.ExpectExec("UPDATE `task_menu`").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	parentKey := "parent_menu"
	err := repo.Move("test_menu", &parentKey)
	if err != nil {
		t.Errorf("移动菜单失败: %v", err)
	}
}

func TestMenuRepository_GetTree(t *testing.T) {
	repo, mock := setupMenuRepoTest(t)

	rows := sqlmock.NewRows([]string{
		"id", "key", "title", "icon", "path", "parent_key", "enabled", "sort", "created_at", "updated_at",
	}).AddRow(1, "menu1", "菜单1", "icon1", "/path1", nil, true, 1, time.Now(), time.Now()).
		AddRow(2, "menu2", "菜单2", "icon2", "/path2", strPtr("menu1"), true, 1, time.Now(), time.Now())

	mock.ExpectQuery("SELECT").
		WillReturnRows(rows)

	menus, err := repo.GetTree()
	if err != nil {
		t.Errorf("获取菜单树失败: %v", err)
	}
	if len(menus) == 0 {
		t.Error("期望返回菜单树")
	}
}
