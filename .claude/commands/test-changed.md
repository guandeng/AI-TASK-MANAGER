# 变更测试

只运行与最近代码变更相关的测试，快速验证改动。

## 用法

```bash
/test-changed [选项]
```

## 选项

| 选项 | 说明 |
|-----|------|
| `all` | 所有相关测试（默认） |
| `backend` | 仅后端测试 |
| `frontend` | 仅前端测试 |
| `last` | 仅检查最后一次提交 |

---

## 步骤

### 1. 获取变更文件

```bash
# 最近 5 次提交的变更
git diff --name-only HEAD~5

# 暂存区的变更
git diff --staged --name-only
```

### 2. 运行后端测试

```bash
cd backend

# 如果 handlers 变更，运行 handler 测试
if git diff --name-only HEAD~5 | grep -q "handlers/"; then
  echo "=== 运行 Handler 测试 ==="
  go test ./internal/handlers -v
fi

# 如果 models 变更，运行 model 测试
if git diff --name-only HEAD~5 | grep -q "models/"; then
  echo "=== 运行 Model 测试 ==="
  go test ./internal/models -v
fi
```

### 3. 运行前端测试

```bash
cd frontend

# 运行相关组件测试
npm test -- --testPathPattern="<变更的组件名>"
```

---

## 输出格式

```
== 变更测试报告 ==

📝 变更文件 (最近 5 次提交):
  M backend/internal/handlers/task.go
  M frontend/src/views/task/list/index.vue

后端测试:
  ✅ Task Handler 测试 (5 passed)

前端测试:
  ✅ TaskList 组件测试 (8 passed)

✅ 所有相关测试通过
```

---

## 与完整测试的区别

| 特性 | test-changed | test-full |
|-----|-------------|-----------|
| 测试范围 | 仅变更相关 | 全部测试 |
| 执行时间 | 快 (1-3 分钟) | 慢 (5-10 分钟) |
| 使用场景 | 开发中验证 | 提交前/CI |
