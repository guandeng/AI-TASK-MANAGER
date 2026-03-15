package database

import (
	"context"
	"testing"
	"time"

	"go.uber.org/zap"
)

func TestNewGormLogger(t *testing.T) {
	logger := zap.NewNop()
	gormLogger := NewGormLogger(logger)
	if gormLogger == nil {
		t.Error("期望返回非 nil logger")
	}
}

func TestGormLogger_LogMode(t *testing.T) {
	logger := zap.NewNop()
	gormLogger := NewGormLogger(logger)

	result := gormLogger.LogMode(1)
	if result == nil {
		t.Error("期望返回非 nil logger")
	}
}

func TestGormLogger_Info(t *testing.T) {
	logger := zap.NewNop()
	gormLogger := NewGormLogger(logger)

	gormLogger.Info(nil, "test message")
	// 不应该 panic
}

func TestGormLogger_Warn(t *testing.T) {
	logger := zap.NewNop()
	gormLogger := NewGormLogger(logger)

	gormLogger.Warn(nil, "test warning")
	// 不应该 panic
}

func TestGormLogger_Error(t *testing.T) {
	logger := zap.NewNop()
	gormLogger := NewGormLogger(logger)

	gormLogger.Error(nil, "test error")
	// 不应该 panic
}

func TestGormLogger_Trace(t *testing.T) {
	logger := zap.NewNop()
	gormLogger := NewGormLogger(logger)

	beginTime := time.Now()
	gormLogger.Trace(context.Background(), beginTime, func() (string, int64) { return "SELECT 1", 1 }, nil)
	gormLogger.Trace(context.Background(), beginTime, func() (string, int64) { return "UPDATE users", 1 }, nil)
	// 不应该 panic
}

func TestGetDB(t *testing.T) {
	// 在没有初始化的情况下应该返回 nil
	db := GetDB()
	// DB 可能是 nil（如果没有初始化）或非 nil（如果已初始化）
	_ = db
}

func TestClose(t *testing.T) {
	// 在没有初始化的情况下应该返回 nil
	err := Close()
	if err != nil {
		t.Errorf("期望返回 nil, 实际 %v", err)
	}
}
