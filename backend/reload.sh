#!/bin/bash

# AI Task Manager - 重启脚本
# 用于同时重启 Go 后端和 Vue 前端

set -e

# ���色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 项目根目录（backend 目录的父目录）
PROJECT_ROOT="$(cd "$(dirname "$0")/.." && pwd)"
BACKEND_ROOT="$(cd "$(dirname "$0")" && pwd)"

echo -e "${BLUE}======================================${NC}"
echo -e "${BLUE}   AI Task Manager - 服务重启${NC}"
echo -e "${BLUE}======================================${NC}"

# 停止已有服务
echo -e "${YELLOW}[1/4] 停止现有服务...${NC}"

# 停止 Go 后端 (端口 8080)
pkill -f "go run ./cmd/server" 2>/dev/null || true
pkill -f "$BACKEND_ROOT/bin/server" 2>/dev/null || true
lsof -ti:8080 | xargs kill -9 2>/dev/null || true
echo -e "  ${GREEN}✓${NC} Go 后端已停止 (端口 8080)"

# 停止 Vue 前端 (端口 9527)
pkill -f "vite.*9527" 2>/dev/null || true
lsof -ti:9527 | xargs kill -9 2>/dev/null || true
echo -e "  ${GREEN}✓${NC} Vue 前端已停止 (端口 9527)"

# 清理日志
BACKEND_LOG_DIR="$BACKEND_ROOT/logs"
FRONTEND_LOG_DIR="$PROJECT_ROOT/frontend/logs"
mkdir -p "$BACKEND_LOG_DIR" "$FRONTEND_LOG_DIR"
rm -f "$BACKEND_LOG_DIR/backend.log" "$FRONTEND_LOG_DIR/frontend.log"

sleep 1

# 启动 Go 后端
echo -e "${YELLOW}[2/4] 启动 Go 后端...${NC}"
cd "$BACKEND_ROOT"

# 重新编译 Go 后端
echo -e "  ${BLUE}编译 Go 后端...${NC}"
go build -o bin/server ./cmd/server

# 后台启动 Go 后端
nohup ./bin/server > "$BACKEND_LOG_DIR/backend.log" 2>&1 &
BACKEND_PID=$!
echo -e "  ${GREEN}✓${NC} Go 后端已启动 (PID: $BACKEND_PID, 端口: 8080)"

# 等待后端启动
sleep 2

# 检查后端是否成功启动
if curl -s http://localhost:8080/health > /dev/null 2>&1; then
    echo -e "  ${GREEN}✓${NC} 后端健康检查通过"
else
    echo -e "  ${RED}✗${NC} 后端健康检查失败，请查看日志: "$BACKEND_LOG_DIR"/backend.log"
fi

# 启动 Vue 前端
echo -e "${YELLOW}[3/4] 启动 Vue 前端...${NC}"
cd "$PROJECT_ROOT/frontend"

# 检查依赖
if [ ! -d "node_modules" ]; then
    echo -e "  ${BLUE}安装前端依赖...${NC}"
    pnpm install
fi

# 后台启动 Vue 前端
nohup pnpm dev > "$FRONTEND_LOG_DIR/frontend.log" 2>&1 &
FRONTEND_PID=$!
echo -e "  ${GREEN}✓${NC} Vue 前端正在启动 (PID: $FRONTEND_PID, 端口: 9527)"

# 等待前端启动
sleep 3

# 显示状态
echo -e "${YELLOW}[4/4] 服务状态${NC}"
echo -e "${BLUE}--------------------------------------${NC}"
echo -e "  后端地址: ${GREEN}http://localhost:8080${NC}"
echo -e "  前端地址: ${GREEN}http://localhost:9527${NC}"
echo -e "  健康检查: ${GREEN}http://localhost:8080/health${NC}"
echo -e "${BLUE}--------------------------------------${NC}"
echo -e "  后端日志: "$BACKEND_LOG_DIR"/backend.log"
echo -e "  前端日志: "$FRONTEND_LOG_DIR"/frontend.log"
echo -e "${BLUE}--------------------------------------${NC}"
echo -e "  查看后端日志: ${YELLOW}tail -f "$BACKEND_LOG_DIR"/backend.log${NC}"
echo -e "  查看前端日志: ${YELLOW}tail -f "$FRONTEND_LOG_DIR"/frontend.log${NC}"
echo -e "${BLUE}======================================${NC}"
echo -e "${GREEN}所有服务已重启完成！${NC}"

# 保存 PID 到文件
echo $BACKEND_PID > "$PROJECT_ROOT/.backend.pid"
echo $FRONTEND_PID > "$PROJECT_ROOT/.frontend.pid"
