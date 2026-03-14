# 开发环境启动

快速启动开发环境，包括后端、前端服务。

## 用法

```bash
/dev-setup [选项]
```

## 选项

| 选项 | 说明 |
|-----|------|
| `all` | 启动所有服务（默认） |
| `backend` | 仅启动后端 |
| `frontend` | 仅启动前端 |
| `check` | 检查环境依赖 |
| `stop` | 停止所有服务 |
| `restart` | 重启所有服务 |

---

## 启动步骤

### 1. 环境检查 (--check)

```bash
# 检查 Go 版本
go version

# 检查 Node.js 版本
node -v

# 检查 pnpm 版本
pnpm -v

# 检查数据库连接
# 检查 Redis 连接（如有）
```

### 2. 启动后端 (backend)

```bash
cd backend

# 加载环境变量
cp .env.example .env 2>/dev/null || true

# 运行数据库迁移
go run cmd/server/main.go migrate

# 启动服务
go run cmd/server/main.go
```

后端默认端口：`8000`

### 3. 启动前端 (frontend)

```bash
cd frontend

# 安装依赖（如有需要）
pnpm install

# 启动开发服务器
pnpm dev
```

前端默认端口：`5173`

---

## 健康检查

启动后自动检查：

```bash
# 检查后端
curl -s http://localhost:8000/api/health

# 检查前端
curl -s http://localhost:5173
```

---

## 快速访问

- 前端开发服务器：http://localhost:5173
- 后端 API 服务：http://localhost:8000
- API 文档：http://localhost:8000/swagger
