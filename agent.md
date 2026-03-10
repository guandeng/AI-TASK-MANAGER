# Agent 开发规范

## 目录

- [数据库设计规范](#数据库设计规范)
- [前端代码规范](#前端代码规范)
- [API 接口规范](#api-接口规范)
- [Git 提交规范](#git-提交规范)

---

## 数据库设计规范

### 禁用 ENUM 类型

**原因：**
- 新增/修改枚举值需要 DDL 操作（ALTER TABLE），线上环境 DDL 风险高（锁表、性能影响）
- 跨数据库迁移时 ENUM 兼容性差（如 MySQL → Oracle、PostgreSQL）
- ENUM 值分散在多处代码中，维护困难

**正确做法：**
```sql
-- ❌ 错误：使用 ENUM
`status` ENUM('pending', 'in-progress', 'done', 'deferred') NOT NULL DEFAULT 'pending'

-- ✅ 正确：使用 VARCHAR + 应用层校验
`status` VARCHAR(50) NOT NULL DEFAULT 'pending' COMMENT '任务状态: pending-待处理, in-progress-进行中, done-已完成, deferred-已延期'
```

### 表命名规范

- 使用 `task_` 前缀
- 小写 + 下划线命名
- 表名要有业务含义

```sql
-- ✅ 正确
CREATE TABLE `task_task` (        -- 任务表
CREATE TABLE `task_subtask` (     -- 子任务表
CREATE TABLE `task_dependency` (  -- 依赖关系表
CREATE TABLE `task_history` (     -- 历史记录表
```

### 字段命名规范

- 小写 + 下划线
- 使用有意义的名称
- 多语言字段使用 `_trans` 后缀表示翻译

```sql
`title` VARCHAR(500) NOT NULL COMMENT '任务标题(原文)',
`title_trans` VARCHAR(500) DEFAULT NULL COMMENT '任务标题(翻译)',
`created_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
`updated_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
```

### 必备字段

每个表**必须**包含以下字段：

```sql
`id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键',
`created_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
`updated_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
`deleted_at` TIMESTAMP NULL DEFAULT NULL COMMENT '删���时间(软删除)',
PRIMARY KEY (`id`)
```

**字段说明：**
- `id` - 主键，BIGINT UNSIGNED AUTO_INCREMENT
- `created_at` - 创建时间，记录首次插入时间
- `updated_at` - 更新时间，记录最后修改时间
- `deleted_at` - 删除时间，软删除标记字段，NULL 表示未删除

### 索引规范

- 主键自动创建索引
- 外键字段创建索引
- 常用查询字段创建索引
- 状态字段创建索引

```sql
INDEX `idx_status` (`status`),
INDEX `idx_assignee` (`assignee`),
INDEX `idx_created_at` (`created_at`)
```

---

## 前端代码规范

### 项目结构

```
frontend/
├── src/
│   ├── views/              # 页面组件
│   │   └── task/
│   │       └── list/       # 任务列表页
│   ├── components/         # 公共组件
│   │   ├── common/         # 通用组件
│   │   └── custom/         # 自定义组件
│   ├── store/              # Pinia 状态管理
│   │   └── modules/
│   │       └── task/       # 任务模块 store
│   ├── service/            # API 请求
│   │   └── api/
│   │       └── task/       # 任务 API
│   ├── typings/            # TypeScript 类型定义
│   │   └── api/
│   │       └── task.ts     # 任务类型
│   ├── locales/            # 多语言文件
│   └── router/             # 路由配置
└── docs/
    └── sql/                # 数据库 SQL 文件
```

### 命名规范

| 类型 | 规范 | 示例 |
|------|------|------|
| 数据库表名 | 小写 + 下划线 + 前缀 | `task_task`, `task_subtask` |
| 数据库字段 | 小写 + 下划线 | `created_at`, `title_trans` |
| TypeScript 接口 | PascalCase | `Task`, `Subtask`, `TaskListResponse` |
| TypeScript 类型 | PascalCase | `TaskStatus`, `TaskPriority` |
| Vue 组件 | PascalCase | `TaskList`, `TaskDetail` |
| 常量 | SCREAMING_SNAKE_CASE | `TASK_STATUS_OPTIONS` |
| 变量/函数 | camelCase | `taskStore`, `handleStatusChange` |

### 列表排序规范

- 前端列表页默认按 `id desc` 展示
- 需求列表按 `id desc` 排序
- 任务列��按 `id desc` 排序

### 菜单命名规范

- 菜单名称必须使用中文
- 在 `locales/langs/zh-cn.ts` 的 `route` 字段中定义中文菜单名
- 示例：
  ```typescript
  route: {
    requirement: '需求任务管理',
    requirement_list: '需求列表',
    requirement_task_list: '任务列表'
  }
  ```

### AI 生成规范

- 需求拆分任务时，生成内容必须使用中文
- `title`、`description`、`details`、`testStrategy` 必须输出中文
- 除技术标识符、库名、API 名、代码符号外，不输出英文句子

### 类型定义规范

```typescript
// ✅ 正确：使用 type 定义联合类型
export type TaskStatus = 'pending' | 'in-progress' | 'done' | 'deferred';
export type TaskPriority = 'high' | 'medium' | 'low';

// ✅ 正确：使用 interface 定义对象结构
export interface Task {
  id: number;
  title: string;
  titleTrans?: string;
  status: TaskStatus;
  priority: TaskPriority;
  assignee?: string;
  subtasks?: Subtask[];
}

// ✅ 正确：定义常量选项
export const TASK_STATUS_OPTIONS = [
  { label: '待处理', value: 'pending' },
  { label: '进行中', value: 'in-progress' },
  { label: '已完成', value: 'done' },
  { label: '已延期', value: 'deferred' }
] as const;
```

### Store 规范

```typescript
// store/modules/task/index.ts
export const useTaskStore = defineStore('task-store', () => {
  // 状态
  const tasks = ref<Task[]>([]);
  const loading = ref(false);

  // 计算属性
  const statistics = computed<TaskStatistics>(() => {
    // ...
  });

  // Actions
  async function loadTasks() {
    // ...
  }

  async function setTaskStatus(id: number, status: TaskStatus) {
    // ...
  }

  return {
    tasks,
    loading,
    statistics,
    loadTasks,
    setTaskStatus
  };
});
```

---

## API 接口规范

### RESTful 设计

| 操作 | 方法 | 路径 | 说明 |
|------|------|------|------|
| 获取任务列表 | GET | `/api/tasks` | 返回任务列表 |
| 获取任务详情 | GET | `/api/tasks/:id` | 返回单个任务 |
| 更新任务状态 | PATCH | `/api/tasks/:id` | 部分更新任务 |
| 更新子任务 | PATCH | `/api/tasks/:taskId/subtasks/:subtaskId` | 更新子任务 |

### 响应格式

```typescript
// 成功响应
{
  "code": "0000",
  "message": "success",
  "data": { ... }
}

// 错误响应
{
  "code": "1001",
  "message": "任务不存在",
  "data": null
}
```

---

## Git 提交规范

使用 Conventional Commits：

| 类型 | 说明 | 示例 |
|------|------|------|
| `feat` | 新功能 | `feat: 添加任务状态切换功能` |
| `fix` | 修复 bug | `fix: 修复任务列表分页问题` |
| `docs` | 文档更新 | `docs: 更新数据库设计文档` |
| `style` | 代码格式 | `style: 格式化代码` |
| `refactor` | 代码重构 | `refactor: 重构任务 store` |
| `test` | 测试相关 | `test: 添加任务单元测试` |
| `chore` | 构建/工具 | `chore: 更新依赖版本` |

### 分支命名

- `main` - 主分支
- `feature/xxx` - 功能分支
- `fix/xxx` - 修复分支
- `refactor/xxx` - 重构分支

---

## 开发注意事项

1. **状态管理**：使用 Pinia 进行状态管理，避免组件间直接传递复杂数据
2. **类型安全**：所有 API 响应和 Store 数据都要有 TypeScript 类型定义
3. **国际化**：用户可见的文本都要支持多语言（中文/英文）
4. **错误处理**：API 调用要有 try-catch，并显示友好的错误提示
5. **性能优化**：列表数据使用分页，避免一次性加载过多数据
6. **交互确认**：所有删除操作、状态更新、负责人变更、清空操作、批量修改等会变更数据的按钮或控件，必须先弹出确认提示，再执行实际请求
