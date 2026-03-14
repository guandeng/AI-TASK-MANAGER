# CRUD 代码生成

根据给定的模型/表名，自动生成完整的 CRUD 代码。

## 用法

```bash
/gen-crud <模型名> [选项]
```

## 选项

| 选项 | 说明 | 示例 |
|-----|------|------|
| `--module=<name>` | 模块名称（用于路由和目录） | `--module=user` |
| `--type=list\|detail\|all` | 生成类型 | `--type=all` |
| `--fields` | 字段列表（逗号分隔） | `--fields=title,content,status` |

## 示例

```bash
# 生成用户模块完整 CRUD
/gen-crud User --module=user --type=all

# 仅生成列表页
/gen-crud Message --module=message --type=list

# 生成项目模板 CRUD
/gen-crud ProjectTemplate --module=template --fields=name,description,type
```

---

## 生成内容

### 后端文件

1. **Model** - `backend/internal/models/<model>.go`
2. **Handler** - `backend/internal/handlers/<model>.go`
3. **Routes** - 更新 `backend/cmd/server/main.go` 路由注册

### 前端文件

1. **API 类型** - `frontend/src/typings/api/<module>.ts`
2. **API 请求** - `frontend/src/service/api/<module>/index.ts`
3. **列表页** - `frontend/src/views/<module>/list/index.vue`
4. **详情页** - `frontend/src/views/<module>/detail/[id].vue`
5. **Store** - `frontend/src/store/modules/<module>/index.ts`

---

## 代码规范

### Model 规范

```go
package models

// <Model> 模型说明
type <Model> struct {
    ID          uint64         `gorm:"primary_key;auto_increment" json:"id"`
    // 业务字段...
    CreatedAt   time.Time      `json:"createdAt"`
    UpdatedAt   time.Time      `json:"updatedAt"`
    DeletedAt   gorm.DeletedAt `gorm:"index" json:"deletedAt,omitempty"`
}

// TableName 表名
func (<Model>) TableName() string {
    return "task_<table_name>"
}
```

### Handler 规范

必须包含以下方法：
- `List` - 列表（支持分页、筛选、搜索）
- `Get` - 详情
- `Create` - 创建
- `Update` - 更新
- `Delete` - 删除
- `Statistics` - 统计（可选）

### 前端规范

- 使用 Pinia 管理状态
- 使用 TypeScript 类型定义
- 使用 Naive UI 组件
- 列表支持：分页、筛选、批量操作
- 所有用户交互使用中文

---

## 执行步骤

### 1. 分析需求

根据模型名确定：
- 表名（task_前缀 + 复数小写）
- 路由前缀
- 菜单名称

### 2. 生成后端代码

#### Step 2.1: 创建/更新 Model
检查 `backend/internal/models/` 目录，创建或更新模型文件。

#### Step 2.2: 创建 Handler
在 `backend/internal/handlers/` 创建处理器，包含完整 CRUD 方法。

#### Step 2.3: 注册路由
在 `backend/cmd/server/main.go` 中注册路由。

### 3. 生成前端代码

#### Step 3.1: 创建类型定义
在 `frontend/src/typings/api/` 创建类型文件。

#### Step 3.2: 创建 API 服务
在 `frontend/src/service/api/` 创建 API 请求模块。

#### Step 3.3: 创建 Store
在 `frontend/src/store/modules/` 创建状态管理。

#### Step 3.4: 创建页面组件
- 列表页：支持表格、筛选、分页
- 详情页：支持表单编辑、关联数据展示

### 4. 验证

- 后端编译检查：`cd backend && go build ./...`
- 前端类型检查：`cd frontend && npm run typecheck`

---

## 输出格式

生成完成后输出：

```
✅ 已生成 <模块名> CRUD 代码

后端:
  - models/<model>.go
  - handlers/<model>.go
  - 路由已注册

前端:
  - typings/api/<module>.ts
  - service/api/<module>/index.ts
  - store/modules/<module>/index.ts
  - views/<module>/list/index.vue
  - views/<module>/detail/[id].vue

下一步:
  1. 创建数据库迁移
  2. 配置菜单权限
  3. 重启服务测试
```
