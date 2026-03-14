#!/bin/bash

# AI Task Manager - 停止脚本

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

echo -e "${BLUE}======================================${NC}"
echo -e "${BLUE}   AI Task Manager - 停止服务${NC}"
echo -e "${BLUE}======================================${NC}"

# 停止 Go 后端
echo -e "${YELLOW}停止 Go 后端...${NC}"
pkill -f "go run ./cmd/server" 2>/dev/null || true
pkill -f "backend/bin/server" 2>/dev/null || true
lsof -ti:8080 | xargs kill -9 2>/dev/null || true
echo -e "  ${GREEN}✓${NC} Go 后端已停止"

# 停止 Vue 前端
echo -e "${YELLOW}停止 Vue 前端...${NC}"
pkill -f "vite.*9527" 2>/dev/null || true
lsof -ti:9527 | xargs kill -9 2>/dev/null || true
echo -e "  ${GREEN}✓${NC} Vue 前端已停止"

# 清理 PID 文件
rm -f .backend.pid .frontend.pid 2>/dev/null || true

echo -e "${BLUE}======================================${NC}"
echo -e "${GREEN}所有服务已停止${NC}"
