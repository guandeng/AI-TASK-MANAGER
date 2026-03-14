# 安全检查

全面的安全扫描，检测潜在的安全漏洞。

## 用法

```bash
/security-check [选项]
```

## 选项

| 选项 | 说明 |
|-----|------|
| `all` | 所有检查（默认） |
| `backend` | 仅后端 |
| `frontend` | 仅前端 |
| `deps` | 仅依赖检查 |
| `code` | 仅代码检查 |

---

## 后端安全检查

### 1. 依赖安全扫描

```bash
cd backend

# Gosec 安全扫描
gosec ./... 2>&1 || echo "未安装 gosec: go install github.com/securego/gosec/v2/cmd/gosec@latest"

# 检查过时的依赖
go list -u -m all 2>&1 | grep -E "vulnerability|security" || echo "依赖安全检查通过"
```

### 2. SQL 注入检查

```bash
cd backend

echo "=== SQL 注入风险 ==="

# 检查直接拼接 SQL
grep -rn "fmt\.Sprintf.*\(SELECT\|INSERT\|UPDATE\|DELETE\)" --include="*.go" . | \
  grep -v "_test.go" || echo "未发现明显 SQL 拼接"

# 检查 Raw SQL 使用
grep -rn "\.Raw\|\.Exec" --include="*.go" internal/handlers/ | head -10

# 检查参数化查询使用情况
grep -rn "Where.*?" --include="*.go" internal/handlers/ | wc -l
```

### 3. 命令注入检查

```bash
cd backend

echo "=== 命令注入风险 ==="

# 检查 exec 使用
grep -rn "exec\.Command\|os\.Exec" --include="*.go" . | grep -v "_test.go" || echo "未发现 exec 使用"

# 检查用户输入用于命令
grep -rn "c\.Param\|c\.Query" --include="*.go" internal/handlers/ | grep -i "exec\|run\|sh\|bash" || echo "未发现风险"
```

### 4. 路径遍历检查

```bash
cd backend

echo "=== 路径遍历风险 ==="

# 检查文件操作
grep -rn "os\.Open\|os\.ReadFile\|ioutil\.ReadFile" --include="*.go" internal/handlers/ | head -10

# 检查用户输入用于路径（高风险）
grep -rn "c\.Param.*\(path\|file\|dir\)" --include="*.go" internal/handlers/ || echo "未发现明显风险"
```

### 5. 凭证泄露检查

```bash
cd backend

echo "=== 硬编码凭证检查 ==="

# 检查硬编码密钥/密码
grep -rniE "(secret|api_key|apikey|password|token)\s*=\s*[\"'][^\"']+[\"']" \
  --include="*.go" . | grep -v "_test.go" | grep -v "example\|sample\|mock" || \
  echo "未发现明显硬编码凭证"

# 检查日志中的敏感信息
grep -rn "zap\..*\(password\|secret\|token\|key\)" --include="*.go" . || echo "日志安全检查通过"
```

### 6. XSS 防护检查

```bash
cd backend

echo "=== XSS 风险 ==="

# 检查 HTML 输出
grep -rn "template\.HTML\|html-template" --include="*.go" internal/handlers/ || echo "未发现 HTML 模板输出"
```

---

## 前端安全检查

### 1. 依赖安全扫描

```bash
cd frontend

echo "=== 依赖安全检查 ==="

# npm audit
npm audit --audit-level=high 2>&1 | tail -20 || echo "npm audit 通过"

# 检查过时的依赖
npm outdated 2>&1 | head -20
```

### 2. XSS 风险检查

```bash
cd frontend

echo "=== XSS 风险 ==="

# 检查 v-html 使用
grep -rn "v-html" --include="*.vue" src || echo "未发现 v-html"

# 检查 dangerouslySetInnerHTML
grep -rn "dangerouslySetInnerHTML" --include="*.tsx" --include="*.jsx" src || echo "未发现 dangerouslySetInnerHTML"

# 检查 innerHTML 赋值
grep -rn "innerHTML\s*=" --include="*.ts" --include="*.js" src || echo "未发现 innerHTML 赋值"
```

### 3. 敏感信息检查

```bash
cd frontend

echo "=== 敏感信息检查 ==="

# 检查 localStorage 中的敏感数据
grep -rn "localStorage\.\(setItem\|getItem\)" --include="*.ts" --include="*.vue" src | \
  grep -iE "token|password|secret|key" || echo "localStorage 安全检查通过"

# 检查硬编码密钥
grep -rniE "(api.*key|secret|token)\s*[:=]\s*['\"][^'\"]+['\"]" \
  --include="*.ts" --include="*.vue" src | grep -v "node_modules" || \
  echo "未发现硬编码密钥"
```

### 4. 不安全 API 检查

```bash
cd frontend

echo "=== API 安全检查 ==="

# 检查 HTTP 请求（应为 HTTPS）
grep -rn "http://" --include="*.ts" --include="*.vue" src || echo "未发现 HTTP 请求"

# 检查 eval 使用
grep -rn "eval(\|new Function(" --include="*.ts" --include="*.vue" src || echo "未发现 eval/Function"
```

### 5. 输入验证检查

```bash
cd frontend

echo "=== 输入验证检查 ==="

# 检查表单验证规则
grep -rn "rules:" --include="*.vue" src/views | head -10

# 检查 API 错误处理
grep -rn "\.catch\|try.*catch" --include="*.ts" src/service/api | head -10
```

---

## 输出格式

```
== 安全检查报告 ==

🔴 高危问题:
  [编号] [类型] 文件：行号
  问题描述
  建议修复方案

🟡 中危问题:
  ...

🟢 低危问题:
  ...

✅ 通过项:
  - 依赖安全检查
  - SQL 注入防护
  - ...
```

---

## 修复建议优先级

1. **立即修复**（高危）:
   - SQL 注入漏洞
   - 命令注入漏洞
   - 硬编码凭证
   - 认证绕过

2. **尽快修复**（中危）:
   - XSS 风险
   - 路径遍历
   - 敏感信息泄露

3. **计划修复**（低危）:
   - 过时的依赖
   - 代码规范问题
   - 日志完善
