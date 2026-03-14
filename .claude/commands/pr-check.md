# PR 检查清单

提交前的完整检查清单，确保代码质量。

## 用法

```bash
/pr-check [选项]
```

## 选项

| 选项 | 说明 |
|-----|------|
| `all` | 完整检查（默认） |
| `quick` | 快速检查（仅必需项） |
| `backend` | 仅后端 |
| `frontend` | 仅前端 |

---

## 检查清单

### 必需检查项

- [ ] 代码格式化
- [ ] 编译/构建通过
- [ ] 类型检查通过
- [ ] 无 lint 错误
- [ ] 测试通过

---

## 步骤

### 1. 获取变更文件

```bash
# 暂存的文件
git diff --staged --name-only

# 未暂存的修改
git status --short
```

### 2. 后端检查

```bash
cd backend

echo "=== 后端格式化 ==="
gofmt -l . | head -10

echo "=== 后端编译 ==="
go build ./... && echo "✅ 编译通过" || echo "❌ 编译失败"

echo "=== go vet ==="
go vet ./... && echo "✅ go vet 通过" || echo "❌ go vet 失败"

echo "=== 后端测试 ==="
go test ./... -short -timeout=30s
```

### 3. 前端检查

```bash
cd frontend

echo "=== 前端类型检查 ==="
npm run typecheck

echo "=== 前端 Lint ==="
npm run lint -- --quiet

echo "=== 前端构建 ==="
npm run build
```

---

## 输出格式

```
== PR 检查报告 ==

📝 变更文件 (5):
  M backend/internal/handlers/task.go
  M frontend/src/views/task/list/index.vue

后端:
  ✅ go fmt 通过
  ✅ go build 通过
  ✅ go vet 通过
  ✅ 测试通过

前端:
  ✅ typecheck 通过
  ✅ lint 通过
  ✅ build 通过

✅ PR 检查通过，可以提交
```
