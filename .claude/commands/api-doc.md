# API 文档生成

扫描所有路由定义，生成/更新 API 文档。

## 用法

```bash
/api-doc [选项]
```

## 选项

| 选项 | 说明 | 默认值 |
|-----|------|-------|
| `--output=markdown\|json` | 输出格式 | `markdown` |
| `--module=<name>` | 仅生成指定模块 | 全部 |
| `--update` | 更新现有文档 | false |

---

## 步骤

### 1. 扫描后端路由

```bash
cd backend

# 扫描所有 Handler 文件
grep -rn "func.*\(.*\*gin.Context\)" --include="*.go" internal/handlers/

# 扫描路由注册
grep -rn "GET\|POST\|PUT\|DELETE\|PATCH" --include="*.go" cmd/server/
```

### 2. 分析 API 结构

对每个 API 端点分析：
- 路径
- HTTP 方法
- Handler 方法
- 请求参数（Query/Body/Path）
- 响应结构

### 3. 生成文档

#### Markdown 格式

```markdown
# API 文档

## 模块名

### 接口名称

- **路径**: `/api/xxx`
- **方法**: `GET/POST/PUT/DELETE`
- **说明**: 接口说明
- **请求参数**:
  | 参数 | 类型 | 必填 | 说明 |
  |-----|------|-----|------|
- **响应示例**:
  ```json
  {}
  ```
```

### 4. 更新文档

如果 `--update` 参数存在，对比现有文档更新。

---

## 输出

文档生成到 `docs/api/` 目录：
- `README.md` - 总览
- `<module>.md` - 各模块文档
