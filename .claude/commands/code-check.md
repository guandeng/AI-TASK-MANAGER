# 代码检查

全面代码检测工具，支持多种检测类型和参数选择。

## 用法

```bash
/code-check [类型]
```

## 检测类型

| 类型 | 说明 |
|-----|------|
| `all` | 运行所有检查（默认） |
| `security` | 安全检查 |
| `quality` | 代码质量检查 |
| `style` | 代码风格检查 |
| `type` | 类型检查 |
| `test` | 测试检查 |
| `backend` | 仅后端检查 |
| `frontend` | 仅前端检查 |

---

## 步骤

### 1. 安全检查 (security)

#### 后端安全检查
```bash
cd backend

# Gosec 安全扫描
gosec ./... 2>/dev/null || echo "未安装 gosec: go install github.com/securego/gosec/v2/cmd/gosec@latest"

# 检查硬编码凭证
echo "=== 检查硬编码凭证 ==="
grep -rn "\"[a-zA-Z0-9_]*[sS]ecret[sS] *[sS]tored[^\"]*\"" --include="*.go" . || echo "未发现明显硬编码密钥"
grep -rn "password.*=.*\"[^\"]+\"" --include="*.go" . || echo "未发现硬编码密码"

# 检查 SQL 注入风险
echo "=== 检查 SQL 注入风险 ==="
grep -rn "db\.Exec\|db\.Raw\|db\.Where.*%" --include="*.go" ./internal/handlers | head -20

# 检查命令注入风险
echo "=== 检查命令注入风险 ==="
grep -rn "exec\.\|os\.Exec\|exec\.Command" --include="*.go" . | grep -v "_test.go"

# 检查路径遍历风险
echo "=== 检查路径遍历风险 ==="
grep -rn "os\.Open\|ioutil\.ReadFile" --include="*.go" ./internal/handlers | head -10
```

#### 前端安全检查
```bash
cd frontend

# 检查 dangerouslySetInnerHTML 或 v-html
echo "=== 检查 XSS 风险 (v-html) ==="
grep -rn "v-html\|dangerouslySetInnerHTML" --include="*.vue" --include="*.tsx" src || echo "未发现 v-html 使用"

# 检查 localStorage 中的敏感信息
echo "=== 检查 localStorage 使用 ==="
grep -rn "localStorage\|sessionStorage" --include="*.ts" --include="*.vue" src | grep -i "token\|password\|secret" || echo "未发现敏感存储"

# 检查 console.log (生产环境应移除)
echo "=== 检查 console.log ==="
grep -rn "console\.log" --include="*.ts" --include="*.vue" src/views src/components | head -20

# 检查 eval/Function 使用
echo "=== 检查 eval/Function 使用 ==="
grep -rn "eval(\|new Function(" --include="*.ts" --include="*.vue" src || echo "未发现 eval/Function 使用"

# npm audit 依赖安全检查
npm audit --audit-level=high 2>/dev/null || echo "npm audit 未执行或无需关注"
```

---

### 2. 代码质量检查 (quality)

#### 后端质量
```bash
cd backend

# go vet 静态分析
echo "=== go vet ==="
go vet ./...

# 未使用的代码检测
echo "=== 未使用代码检测 (unused) ==="
which unused >/dev/null 2>&1 && unused ./... || echo "unused 未安装：go install golang.org/x/tools/go/analysis/passes/unused/cmd/unused@latest"

# 复杂度检测 (gocyclo)
echo "=== 函数复杂度检测 ==="
which gocyclo >/dev/null 2>&1 && gocyclo -over 15 . || echo "gocyclo 未安装：go install github.com/fzipp/gocyclo/cmd/gocyclo@latest"

# 检查错误处理
echo "=== 检查未处理的错误 ==="
grep -rn ":= [a-zA-Z].*(" --include="*.go" ./internal/handlers | grep -v "err" | head -20
```

#### 前端质量
```bash
cd frontend

# ESLint 检查
echo "=== ESLint ==="
npm run lint -- --quiet 2>&1 | head -50

# 检查组件复杂度
echo "=== 大文件检测 (>500 行) ==="
find src -name "*.vue" -o -name "*.ts" -o -name "*.tsx" | while read f; do
  lines=$(wc -l < "$f")
  if [ "$lines" -gt 500 ]; then
    echo "$f: $lines 行"
  fi
done | head -20
```

---

### 3. 代码风格检查 (style)

#### 后端风格
```bash
cd backend

# gofmt 检查
echo "=== gofmt 检查 ==="
unformatted=$(gofmt -l .)
if [ -n "$unformatted" ]; then
  echo "以下文件格式不规范:"
  echo "$unformatted"
else
  echo "所有 Go 文件格式规范"
fi

# goimports 检查导入顺序
echo "=== 导入顺序检查 ==="
goimports -l . | head -10 || echo "goimports 未安装或全部规范"
```

#### 前端风格
```bash
cd frontend

# Prettier/ESLint 风格检查
echo "=== 代码风格检查 ==="
npm run lint 2>&1 | tail -30
```

---

### 4. 类型检查 (type)

#### 后端类型
```bash
cd backend

# 编译检查
echo "=== 编译检查 ==="
go build ./... && echo "编译通过" || echo "编译失败"
```

#### 前端类型
```bash
cd frontend

# TypeScript 类型检查
echo "=== TypeScript 类型检查 ==="
npm run typecheck
```

---

### 5. 测试检查 (test)

```bash
# 后端测试
echo "=== 后端测试 ==="
cd backend && go test ./... -short -timeout=30s

# 前端测试
echo "=== 前端测试 ==="
cd frontend && npm test 2>&1 | tail -20
```

---

### 6. 仅后端检查 (backend)

运行以下检查：
- 后端安全检查
- 后端质量检查
- 后端风格检查
- 后端类型检查

---

### 7. 仅前端检查 (frontend)

运行以下检查：
- 前端安全检查
- 前端质量检查
- 前端风格检查
- 前端类型检查

---

## 执行逻辑

根据传入参数执行对应的检查：

- `all` 或不传参数 → 运行所有检查
- `security` → 仅安全检查
- `quality` → 仅质量检查
- `style` → 仅风格检查
- `type` → 仅类型检查
- `test` → 仅测试检查
- `backend` → 仅后端相关检查
- `frontend` → 仅前端相关检查

## 输出格式

每个检查模块输出：
- ✅ 通过
- ⚠️ 警告（非阻塞）
- ❌ 失败（阻塞）

最后输出总结报告。
