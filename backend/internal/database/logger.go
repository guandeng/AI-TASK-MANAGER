package database

import (
	"context"
	"errors"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

// GormLogger 自定义 GORM 日志适配器
type GormLogger struct {
	zapLogger *zap.Logger
}

// NewGormLogger 创建新的 GORM 日志适配器
func NewGormLogger(zapLogger *zap.Logger) *GormLogger {
	return &GormLogger{
		zapLogger: zapLogger,
	}
}

// LogMode 设置日志模式
func (l *GormLogger) LogMode(level gormlogger.LogLevel) gormlogger.Interface {
	return l
}

// Info 打印 Info 级别日志
func (l *GormLogger) Info(ctx context.Context, msg string, data ...interface{}) {
	l.zapLogger.Sugar().Infof(msg, data...)
}

// Warn 打印 Warn 级别日志
func (l *GormLogger) Warn(ctx context.Context, msg string, data ...interface{}) {
	l.zapLogger.Sugar().Warnf(msg, data...)
}

// Error 打印 Error 级别日志
func (l *GormLogger) Error(ctx context.Context, msg string, data ...interface{}) {
	l.zapLogger.Sugar().Errorf(msg, data...)
}

// Trace 打印 SQL 执行日志
func (l *GormLogger) Trace(ctx context.Context, begin time.Time, fc func() (sql string, rowsAffected int64), err error) {
	elapsed := time.Since(begin)
	sql, rows := fc()

	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		l.zapLogger.Error("SQL Error",
			zap.String("sql", sql),
			zap.Int64("rows", rows),
			zap.Duration("elapsed", elapsed),
			zap.Error(err),
		)
		return
	}

	l.zapLogger.Debug("SQL",
		zap.String("sql", sql),
		zap.Int64("rows", rows),
		zap.Duration("elapsed", elapsed),
	)
}
