package database

import (
	"fmt"
	"time"

	"github.com/ai-task-manager/backend/internal/config"
	"github.com/ai-task-manager/backend/internal/models"
	"go.uber.org/zap"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

// Init 初始化数据库连接
func Init(cfg *config.DatabaseConfig, log *zap.Logger) error {
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
	})
	if err != nil {
		return fmt.Errorf("连接数据库失败: %w", err)
	}

	// 获取底层 sql.DB 并配置连接池
	sqlDB, err := DB.DB()
	if err != nil {
		return fmt.Errorf("获取数据库连接失败: %w", err)
	}

	sqlDB.SetMaxIdleConns(cfg.PoolSize / 2)
	sqlDB.SetMaxOpenConns(cfg.PoolSize)
	sqlDB.SetConnMaxLifetime(time.Hour)

	// 测试连接
	if err := sqlDB.Ping(); err != nil {
		return fmt.Errorf("数据库连接测试失败: %w", err)
	}

	// 自动迁移
	if err := autoMigrate(); err != nil {
		return fmt.Errorf("数据库迁移失败: %w", err)
	}

	return nil
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
