# 死代码检测

检测并报告未使用的代码。

## 用法

```bash
/dead-code [选项]
```

## 选项

| 选项 | 说明 |
|-----|------|
| `all` | 所有检查（默认） |
| `backend` | 仅后端 |
| `frontend` | 仅前端 |

---

## 后端死代码检测

```bash
cd backend

echo "=== Go 未使用检测 ==="

# 使用 staticcheck
staticcheck ./... 2>&1 | grep -E "U1000|unused" || echo "staticcheck 未发现未使用代码"

# 使用 vet 检查
go vet ./... 2>&1 | grep -E "unused" || echo "go vet 未发现未使用代码"

# 检查未导出的私有函数
echo "=== 未导出的私有函数 ==="
grep -rn "^func [a-z][a-zA-Z0-9_]*(" --include="*.go" internal/ | \
  grep -v "_test.go" | head -20
```

---

## 前端死代码检测

```bash
cd frontend

echo "=== TypeScript 未使用检测 ==="

# 使用 ESLint 检测未使用变量
npm run lint 2>&1 | grep -E "@typescript-eslint/no-unused-vars|no-unused-vars" | head -20

# 检查未使用的导入
echo "=== 未使用的导入 ==="
grep -rn "import.*from" --include="*.ts" --include="*.vue" src | head -30
```

---

## 输出格式

```
== 死代码检测报告 ==

后端:
  未使用的函数 (5):
    - internal/handlers/old.go:15 - func deprecatedHandler()

  未使用的常量 (3):
    - internal/config/constants.go:8 - const DeprecatedStatus

前端:
  未使用的导入 (8):
    - src/views/task/list/index.vue:15 - import { unusedFunc }

建议删除的文件:
  - backend/internal/handlers/old_handler.go
```
