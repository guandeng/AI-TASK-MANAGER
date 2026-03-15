package main

import (
	"fmt"
	"os"

	"github.com/ai-task-manager/backend/internal/config"
	"github.com/ai-task-manager/backend/internal/database"
	"go.uber.org/zap"
)

func main() {
	// 加载配置
	cfg, err := config.Load("")
	if err != nil {
		fmt.Printf("加载配置失败：%v\n", err)
		os.Exit(1)
	}

	// 初始化日志
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	// 初始化数据库（执行自动迁移）
	if err := database.InitWithOptions(&cfg.Database, logger, true); err != nil {
		fmt.Printf("数据库迁移失败：%v\n", err)
		os.Exit(1)
	}
	defer database.Close()

	fmt.Println("数据库迁移完成！")
}
