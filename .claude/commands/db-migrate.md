# 数据库迁移检查

扫描 models 目录下的模型定义，对比当前数据库 schema，生成迁移 SQL 脚本。

## 用法

```bash
/db-migrate [选项]
```

## 选项

| 选项 | 说明 |
|-----|------|
| `--dry-run` | 仅预览，不执行 |
| `--output=<file>` | 输出到指定文件 |
| `--rollback` | 生成回滚 SQL |

---

## 步骤

### 1. 扫描 Models

```bash
cd backend

# 扫描所有模型文件
grep -rn "type.*struct" --include="*.go" internal/models/
```

### 2. 分析模型结构

对每个模型分析：
- 表名
- 字段定义
- 索引定义
- 约束条件

### 3. 对比数据库

```bash
# 连接数据库并获取当前 schema
mysql -u root -p -e "DESCRIBE task_xxx;"

# 或使用 GORM 自动迁移预览
go run -c "gorm.AutoMigrate()"
```

### 4. 生成迁移 SQL

```sql
-- 新增字段
ALTER TABLE `task_xxx` ADD COLUMN `xxx` VARCHAR(255) COMMENT '字段说明';

-- 修改字段
ALTER TABLE `task_xxx` MODIFY COLUMN `xxx` VARCHAR(500) NOT NULL;

-- 新增索引
CREATE INDEX `idx_xxx` ON `task_xxx` (`xxx`);
```

### 5. 执行迁移

```bash
cd backend
go run cmd/server/main.go migrate
```

---

## 输出

```
== 数据库迁移检查 ==

模型变更:
  + task_message: 新增字段 result_summary
  ~ task_task: 修改字段 priority 长度

索引变更:
  + idx_status on task_message

生成的迁移 SQL:
----------------------------------------
ALTER TABLE `task_message` ADD COLUMN `result_summary` TEXT DEFAULT NULL COMMENT '结果摘要';
CREATE INDEX `idx_status` ON `task_message` (`status`);
----------------------------------------

执行：/db-migrate --apply
```
