package repository

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"

	"github.com/ai-task-manager/backend/internal/database"
	"github.com/ai-task-manager/backend/internal/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// ErrDatabaseNotInitialized 数据库未初始化错误
var ErrDatabaseNotInitialized = errors.New("数据库未初始化")

// MenuRepository 菜单仓储接口
type MenuRepository interface {
	// 基础 CRUD
	List(page, pageSize int) ([]models.Menu, int64, error)
	GetByKey(key string) (*models.Menu, error)
	Create(menu *models.Menu) error
	Update(menu *models.Menu) error
	Delete(key string) error
	BatchDelete(keys []string) error

	// 菜单树
	GetTree() ([]models.Menu, error)

	// 菜单操作
	Reorder(orderData []map[string]interface{}) error
	Move(key string, parentKey *string) error
	ToggleEnabled(key string, enabled bool) error

	// 同步
	SyncFromJSONFile() error
}

// menuRepository 菜单仓储实现
type menuRepository struct {
	db *gorm.DB
}

// NewMenuRepository 创建菜单仓储
func NewMenuRepository() MenuRepository {
	return &menuRepository{db: database.GetDB()}
}

// List 获取菜单列表
func (r *menuRepository) List(page, pageSize int) ([]models.Menu, int64, error) {
	if r.db == nil {
		return nil, 0, ErrDatabaseNotInitialized
	}

	var menus []models.Menu
	var total int64

	// 计算总数
	if err := r.db.Model(&models.Menu{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * pageSize
	if err := r.db.Order("sort, created_at").Offset(offset).Limit(pageSize).Find(&menus).Error; err != nil {
		return nil, 0, err
	}

	return menus, total, nil
}

// GetByKey 根据 key 获取菜单
func (r *menuRepository) GetByKey(key string) (*models.Menu, error) {
	if r.db == nil {
		return nil, ErrDatabaseNotInitialized
	}

	var menu models.Menu
	if err := r.db.First(&menu, "`key` = ?", key).Error; err != nil {
		return nil, err
	}
	return &menu, nil
}

// Create 创建菜单
func (r *menuRepository) Create(menu *models.Menu) error {
	if r.db == nil {
		return ErrDatabaseNotInitialized
	}
	return r.db.Create(menu).Error
}

// Update 更新菜单
func (r *menuRepository) Update(menu *models.Menu) error {
	if r.db == nil {
		return ErrDatabaseNotInitialized
	}
	return r.db.Save(menu).Error
}

// Delete 删除菜单
func (r *menuRepository) Delete(key string) error {
	if r.db == nil {
		return ErrDatabaseNotInitialized
	}
	return r.db.Transaction(func(tx *gorm.DB) error {
		// 删除菜单本身
		if err := tx.Where("`key` = ?", key).Delete(&models.Menu{}).Error; err != nil {
			return err
		}
		// 递归删除子菜单
		return r.deleteChildren(tx, key)
	})
}

// deleteChildren 递归删除子菜单
func (r *menuRepository) deleteChildren(tx *gorm.DB, parentKey string) error {
	var children []models.Menu
	if err := tx.Where("parent_key = ?", parentKey).Find(&children).Error; err != nil {
		return err
	}

	for _, child := range children {
		if err := tx.Where("`key` = ?", child.Key).Delete(&models.Menu{}).Error; err != nil {
			return err
		}
		// 递归删除子菜单的子菜单
		if err := r.deleteChildren(tx, child.Key); err != nil {
			return err
		}
	}
	return nil
}

// BatchDelete 批量删除菜单
func (r *menuRepository) BatchDelete(keys []string) error {
	if r.db == nil {
		return ErrDatabaseNotInitialized
	}
	return r.db.Transaction(func(tx *gorm.DB) error {
		for _, key := range keys {
			if err := r.deleteChildren(tx, key); err != nil {
				return err
			}
		}
		return tx.Where("key IN ?", keys).Delete(&models.Menu{}).Error
	})
}

// GetTree 获取菜单树
func (r *menuRepository) GetTree() ([]models.Menu, error) {
	if r.db == nil {
		return nil, ErrDatabaseNotInitialized
	}

	var allMenus []models.Menu
	if err := r.db.Order("sort, created_at").Find(&allMenus).Error; err != nil {
		return nil, err
	}

	// 构建菜单树
	menuMap := make(map[string]*models.Menu)
	for i := range allMenus {
		menuMap[allMenus[i].Key] = &allMenus[i]
	}

	var roots []models.Menu
	for _, menu := range allMenus {
		if menu.ParentKey == nil || *menu.ParentKey == "" {
			roots = append(roots, menu)
		} else {
			if parent, ok := menuMap[*menu.ParentKey]; ok {
				parent.Children = append(parent.Children, menu)
			} else {
				// 父节点不存在，作为根节点
				roots = append(roots, menu)
			}
		}
	}

	return roots, nil
}

// Reorder 菜单排序
func (r *menuRepository) Reorder(orderData []map[string]interface{}) error {
	if r.db == nil {
		return ErrDatabaseNotInitialized
	}
	return r.db.Transaction(func(tx *gorm.DB) error {
		for _, item := range orderData {
			key, ok := item["key"].(string)
			if !ok {
				continue
			}
			order, ok := item["order"].(float64)
			if !ok {
				continue
			}
			if err := tx.Model(&models.Menu{}).
				Where("`key` = ?", key).
				Update("sort", uint(order)).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// Move 移动菜单
func (r *menuRepository) Move(key string, parentKey *string) error {
	if r.db == nil {
		return ErrDatabaseNotInitialized
	}
	return r.db.Model(&models.Menu{}).
		Where("`key` = ?", key).
		Update("parent_key", parentKey).Error
}

// ToggleEnabled 切换菜单启用状态
func (r *menuRepository) ToggleEnabled(key string, enabled bool) error {
	if r.db == nil {
		return ErrDatabaseNotInitialized
	}
	return r.db.Model(&models.Menu{}).
		Where("`key` = ?", key).
		Update("enabled", enabled).Error
}

// JSONMenu JSON 菜单结构
type JSONMenu struct {
	Key         string  `json:"key"`
	Label       string  `json:"label"`
	Icon        *string `json:"icon"`
	Path        *string `json:"path"`
	RouteName   *string `json:"routeName"`
	Order       int     `json:"order"`
	Enabled     bool    `json:"enabled"`
	HideInMenu  bool    `json:"hideInMenu"`
	ParentKey   *string `json:"parentKey"`
	I18nKey     *string `json:"i18nKey"`
	Href        *string `json:"href"`
	NewWindow   bool    `json:"newWindow"`
}

// SyncFromJSONFile 从 menu.json 同步菜单到数据库
func (r *menuRepository) SyncFromJSONFile() error {
	if r.db == nil {
		return ErrDatabaseNotInitialized
	}

	// 读取 menu.json 文件（尝试多个路径）
	paths := []string{
		filepath.Join("..", "menu.json"), // 从 backend 目录向上找
		"menu.json",                       // 当前目录
		filepath.Join(".", "menu.json"),   // 当前目录
	}

	var data []byte
	var err error
	for _, jsonPath := range paths {
		data, err = os.ReadFile(jsonPath)
		if err == nil {
			break
		}
	}
	if err != nil {
		return err
	}

	var jsonMenus []JSONMenu
	if err := json.Unmarshal(data, &jsonMenus); err != nil {
		return err
	}

	return r.db.Transaction(func(tx *gorm.DB) error {
		for _, jm := range jsonMenus {
			hideInMenu := jm.HideInMenu
			newWindow := jm.NewWindow

			menu := models.Menu{
				Key:        jm.Key,
				Title:      jm.Label,
				Icon:       jm.Icon,
				Path:       jm.Path,
				RouteName:  jm.RouteName,
				Sort:       uint(jm.Order),
				Enabled:    jm.Enabled,
				HideInMenu: &hideInMenu,
				ParentKey:  jm.ParentKey,
				I18nKey:    jm.I18nKey,
				Href:       jm.Href,
				NewWindow:  &newWindow,
			}

			// 使用 GORM 的 upsert（OnConflict）
			if err := tx.Clauses(clause.OnConflict{
				Columns:   []clause.Column{{Name: "key"}},
				DoUpdates: clause.AssignmentColumns([]string{"title", "icon", "path", "route_name", "sort", "enabled", "hide_in_menu", "parent_key", "i18n_key", "href", "new_window", "updated_at"}),
			}).Create(&menu).Error; err != nil {
				return err
			}
		}
		return nil
	})
}
