package main

import (
	"context"
	"fmt"
	"os"

	"github.com/ai-task-manager/backend/internal/config"
	"github.com/ai-task-manager/backend/internal/database"
	"github.com/ai-task-manager/backend/internal/mcp"
	"go.uber.org/zap"
)

func main() {
	// 加载配置
	cfg, err := config.Load("")
	if err != nil {
		fmt.Printf("加载配置失败: %v\n", err)
		os.Exit(1)
	}

	// 初始化日志
	var logger *zap.Logger
	if cfg.Server.Mode == "release" {
		logger, _ = zap.NewProduction()
	} else {
		logger, _ = zap.NewDevelopment()
	}
	defer logger.Sync()

	// 初始化数据库
	if err := database.Init(&cfg.Database, logger); err != nil {
		logger.Fatal("初始化数据库失败", zap.Error(err))
	}
	defer database.Close()

	// 创建 MCP 服务器
	server := mcp.NewServer(cfg, logger)

	// 启动服务器
	if err := server.Start(context.Background()); err != nil {
		logger.Fatal("启动 MCP 服务器失败", zap.Error(err))
	}
}
