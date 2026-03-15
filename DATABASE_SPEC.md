# 数据库设计规范

> 基于《阿里巴巴 Java 开发手册》数据库规范，结合本项目实际情况制定

## 一、基础规范

| 规范 | 说明 |
|------|------|
| **存储引擎** | 必须使用 InnoDB（支持事务、行级锁、并发性能更好） |
| **字符集** | 必须使用 `utf8mb4` + `utf8mb4_unicode_ci`（支持 emoji，无乱码风险） |
| **表注释** | 表必须有中文注释，说明表的用途 |
| **字段注释** | 字段必须有中文注释，说明字段含义 |
| **禁用项** | 禁止使用存储过程、视图、触发器、Event、外键级联 |
| **大对象** | 禁止存储大文件或大照片，只存 URI |

## 二、命名规范

| 规范 | 示例 |
|------|------|
| **表名** | 小写+下划线，`task_` 前缀，见名知意 | `task_member`, `task_subtask` |
| **字段名** | 小写+下划线，不超过 32 字符 | `created_at`, `is_read` |
| **主键** | `id`，bigint unsigned，自增 | `id bigint unsigned NOT NULL AUTO_INCREMENT` |
| **索引** | `idx_` 前缀 | `idx_task_id`, `idx_status` |
| **唯一索引** | `uniq_` 前缀 | `uniq_email` |

## 三、表设计规范

### 3.1 强制规范

1. **单实例表数目 < 500**
2. **单表列数目 < 30**
3. **表必须有主键**，推荐自增主键
4. **禁止使用外键**，外键约束由应用层控制

### 3.2 主键设计原则

```sql
-- 正确：自增主键
id bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '主键ID'

-- 原因：
-- 1. 主键递增，提高插入性能，避免 page 分裂
-- 2. 减少表碎片，提升空间和内存使用
-- 3. 较短的数据类型减少索引磁盘空间
```

## 四、字段设计规范

### 4.1 NULL 规范

```sql
-- 正确：NOT NULL + 默认值
status varchar(20) NOT NULL DEFAULT 'pending' COMMENT '状态'
is_read tinyint(1) NOT NULL DEFAULT 0 COMMENT '是否已读'

-- 错误：允许 NULL
status varchar(20) COMMENT '状态'
```

**原因**：
- NULL 使索引/统计/比较更复杂
- NULL 需要额外存储空间
- NULL 只能用 `IS NULL` / `IS NOT NULL` 判断

### 4.2 类型选择

| 场景 | 推荐类型 | 示例 |
|------|----------|------|
| 主键/ID | `bigint unsigned` | `id bigint unsigned` |
| 外键关联 | `bigint unsigned` | `task_id bigint unsigned` |
| 状态/类型 | `varchar(20)` | `status varchar(20)` |
| 短文本 | `varchar(n)` | `title varchar(500)` |
| 长文本 | `text` | `content text` |
| 布尔值 | `tinyint(1)` | `is_read tinyint(1)` |
| 金额 | `decimal(10,2)` | `price decimal(10,2)` |
| 手机号 | `varchar(20)` | `phone varchar(20)` |
| 时间 | `datetime(3)` | `created_at datetime(3)` |

### 4.3 禁用类型

| 禁用类型 | 替代方案 |
|----------|----------|
| `ENUM` | `varchar(20)` 或 `tinyint` |
| `BLOB` | 文件系统存储 + URI |
| `decimal` 存货币 | `int` 存分，或 `decimal(10,2)` |

## 五、索引设计规范

### 5.1 索引数量

- 单表索引建议 ≤ 5 个
- 单索引字段数 ≤ 5 个

### 5.2 索引原则

```sql
-- 1. 区分度高的字段放前面
CREATE INDEX idx_status_created ON task_task (status, created_at);

-- 2. 外键字段必须建索引
CREATE INDEX idx_task_id ON task_subtask (task_id);

-- 3. 查询条件字段建索引
CREATE INDEX idx_is_read ON task_message (is_read);
```

### 5.3 禁止建索引的场景

- 更新频繁的字段
- 区分度不高的字段（如性别、状态只有 2-3 个值）

## 六、SQL 使用规范

### 6.1 禁止事项

```sql
-- 禁止 SELECT *
SELECT * FROM task_task;  -- ❌

-- 正确：只查需要的字段
SELECT id, title, status FROM task_task;  -- ✅

-- 禁止隐式转换
SELECT * FROM task_member WHERE phone = 13800000000;  -- ❌
SELECT * FROM task_member WHERE phone = '13800000000';  -- ✅

-- 禁止 WHERE 条件用函数
SELECT * FROM task_task WHERE DATE(created_at) = '2024-01-01';  -- ❌
SELECT * FROM task_task WHERE created_at >= '2024-01-01' AND created_at < '2024-01-02';  -- ✅

-- 禁止 % 开头的模糊查询
SELECT * FROM task_task WHERE title LIKE '%关键词%';  -- ❌
SELECT * FROM task_task WHERE title LIKE '关键词%';  -- ✅

-- 禁止大表 JOIN
SELECT * FROM task_task t1 JOIN task_subtask t2 ON t1.id = t2.task_id;  -- ❌ 大表禁用

-- 禁止 OR 改用 IN
SELECT * FROM task_task WHERE status = 'pending' OR status = 'in-progress';  -- ❌
SELECT * FROM task_task WHERE status IN ('pending', 'in-progress');  -- ✅
```

### 6.2 INSERT 规范

```sql
-- 禁止不指定列
INSERT INTO task_task VALUES (...);  -- ❌

-- 正确：明确指定列
INSERT INTO task_task (title, status, priority) VALUES ('任务1', 'pending', 'high');  -- ✅
```

## 七、本项目问题分析

### 7.1 已发现问题

| 问题 | 表 | 修复建议 |
|------|-----|----------|
| 缺少表注释 | 大部分表 | 添加 `COMMENT '表用途'` |
| 缺少字段注释 | 大部分字段 | 添加 `COMMENT '字段含义'` |
| 允许 NULL | `created_at`, `updated_at` 等字段 | 改为 `NOT NULL DEFAULT CURRENT_TIMESTAMP` |
| 外键约束 | 已删除 | ✅ 已修复 |
| TEXT 类型 | 部分表 | 评估是否需要改 varchar |

### 7.2 建议修复的表

```sql
-- 示例：修复 task_task 表
ALTER TABLE task_task COMMENT '任务表';

ALTER TABLE task_task
  MODIFY COLUMN status varchar(20) NOT NULL DEFAULT 'pending' COMMENT '状态：pending/in-progress/done',
  MODIFY COLUMN priority varchar(20) NOT NULL DEFAULT 'medium' COMMENT '优先级：high/medium/low',
  MODIFY COLUMN created_at datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) COMMENT '创建时间',
  MODIFY COLUMN updated_at datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3) COMMENT '更新时间';
```

## 八、GORM 模型规范

### 8.1 标准 Model 定义

```go
type Task struct {
    ID        uint64    `gorm:"primaryKey;autoIncrement;comment:主键ID" json:"id"`
    Title     string    `gorm:"column:title;type:varchar(500);not null;comment:任务标题" json:"title"`
    Status    string    `gorm:"column:status;type:varchar(20);not null;default:pending;comment:状态" json:"status"`
    Priority  string    `gorm:"column:priority;type:varchar(20);not null;default:medium;comment:优先级" json:"priority"`
    CreatedAt time.Time `gorm:"column:created_at;type:datetime(3);not null;default:CURRENT_TIMESTAMP(3);comment:创建时间" json:"createdAt"`
    UpdatedAt time.Time `gorm:"column:updated_at;type:datetime(3);not null;default:CURRENT_TIMESTAMP(3);autoUpdateTime;comment:更新时间" json:"updatedAt"`

    // 关联（不使用外键约束）
    Subtasks []Subtask `gorm:"foreignKey:TaskID;constraint:OnDelete:CASCADE" json:"-"`
}

func (Task) TableName() string {
    return "task_task"
}
```

### 8.2 GORM 注意事项

1. **不使用外键约束**：`constraint:OnDelete:CASCADE` 只是逻辑关系，数据库层面不建外键
2. **所有字段加 comment**：方便后续维护
3. **时间字段用 datetime(3)**：保留毫秒精度
4. **使用指针表示可空**：`*uint64`、`*string`

---

## 参考资料

- [阿里巴巴 MySQL 数据库规范](https://developer.aliyun.com/article/834372)
- [阿里 MySQL 索引优化与 SQL 开发规范](https://developer.aliyun.com/article/1573143)
