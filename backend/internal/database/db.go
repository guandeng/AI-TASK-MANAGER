package database

import (
	"fmt"
	"os"
	"time"

	"github.com/ai-task-manager/backend/internal/config"
	"github.com/ai-task-manager/backend/internal/models"
	"go.uber.org/zap"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

// Init 初始化数据库连接（默认启用自动迁移）
func Init(cfg *config.DatabaseConfig, log *zap.Logger) error {
	return InitWithOptions(cfg, log, isAutoMigrateEnabled())
}

// InitWithOptions 初始化数据库连接（可配置选项）
func InitWithOptions(cfg *config.DatabaseConfig, log *zap.Logger, withAutoMigrate bool) error {
	var err error

	// 配置 GORM 日志
	var gormLogger logger.Interface
	if log != nil {
		gormLogger = NewGormLogger(log)
	} else {
		gormLogger = logger.Default
	}

	// 连接数据库
	dsn := cfg.GetDSN()
	DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: gormLogger,
		DisableForeignKeyConstraintWhenMigrating: true, // 禁用外键约束
	})
	if err != nil {
		return fmt.Errorf("连接数据库失败：%w", err)
	}

	// 获取底层 sql.DB 并配置连接池
	sqlDB, err := DB.DB()
	if err != nil {
		return fmt.Errorf("获取数据库连接失败：%w", err)
	}

	sqlDB.SetMaxIdleConns(cfg.PoolSize / 2)
	sqlDB.SetMaxOpenConns(cfg.PoolSize)
	sqlDB.SetConnMaxLifetime(time.Hour)

	// 测试连接
	if err := sqlDB.Ping(); err != nil {
		return fmt.Errorf("数据库连接测试失败：%w", err)
	}

	// 根据选项决定是否自动迁移
	if withAutoMigrate {
		if err := autoMigrate(); err != nil {
			return fmt.Errorf("数据库迁移失败：%w", err)
		}
	} else {
		fmt.Println("已跳过自动迁移，请手动执行 init.sql 初始化数据库")
	}

	return nil
}

// isAutoMigrateEnabled 检查是否启用自动迁移（通过环境变量控制）
func isAutoMigrateEnabled() bool {
	return os.Getenv("DB_SKIP_MIGRATE") != "true"
}

// autoMigrate 自动迁移数据库表结构
func autoMigrate() error {
	return DB.AutoMigrate(
		&models.Task{},
		&models.Subtask{},
		&models.TaskDependency{},
		&models.SubtaskDependency{},
		&models.Requirement{},
		&models.RequirementDocument{},
		&models.Member{},
		&models.Assignment{},
		&models.SubtaskAssignment{},
		&models.Comment{},
		&models.ActivityLog{},
		&models.Message{},
		&models.Menu{},
		&models.Config{},
		&models.ProjectTemplate{},
		&models.ProjectTemplateTask{},
		&models.ProjectTemplateSubtask{},
		&models.TaskTemplate{},
		&models.Backup{},
		&models.BackupSchedule{},
		&models.TaskComplexityReport{},
		&models.Language{},
		&models.TaskQualityScore{},
	)
}

// Close 关闭数据库连接
func Close() error {
	if DB != nil {
		sqlDB, err := DB.DB()
		if err != nil {
			return err
		}
		return sqlDB.Close()
	}
	return nil
}

// GetDB 获取数据库实例
func GetDB() *gorm.DB {
	return DB
}

// SetTestDB 设置测试数据库（用于单元测试）
func SetTestDB(db *gorm.DB) {
	DB = db
}
