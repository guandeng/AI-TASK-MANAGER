# 数据库初始化

## 方式一：使用 SQL 脚本（推荐生产环境）

```bash
# 执行 init.sql 初始化数据库
mysql -u root -p ai_task < backend/docs/sql/init.sql
```

## 方式二：使用 GORM 自动迁移（推荐开发环境）

### 默认（执行自动迁移）
```bash
cd backend
go run cmd/server/main.go
```

### 跳过自动迁移
```bash
cd backend
DB_SKIP_MIGRATE=true go run cmd/server/main.go
```

## 方式三：使用专用迁移命令

### 仅执行数据库迁移
```bash
cd backend
go run cmd/migrate/main.go
```

### 仅连接数据库（不迁移）
```bash
cd backend
go run cmd/db-migrate/main.go
```

## 环境变量

| 变量名 | 说明 | 默认值 |
|--------|------|--------|
| `DB_SKIP_MIGRATE` | 是否跳过自动迁移 | `false` |

## 数据库配置

配置文件 `config.yaml` 或使用环境变量：

```yaml
database:
  host: localhost
  port: 3306
  database: ai_task
  username: root
  password: 123456
  table_prefix: task_
  charset: utf8mb4
  collation: utf8mb4_unicode_ci
  pool_size: 10
```

或使用环境变量：

```bash
export ATM_DATABASE_HOST=localhost
export ATM_DATABASE_PORT=3306
export ATM_DATABASE_DATABASE=ai_task
export ATM_DATABASE_USERNAME=root
export ATM_DATABASE_PASSWORD=123456
export ATM_DATABASE_TABLE_PREFIX=task_
```
