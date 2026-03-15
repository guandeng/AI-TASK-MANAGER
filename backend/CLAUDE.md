# AI Task Manager - 后端协作规范

## 技术栈
- Go 1.22+
- Gin + GORM
- MySQL 8.0

## Go 语法规范
- 使用现代 Go 语法 (Go 1.22+)
- 使用 `any` 代替 `interface{}`
- 使用 `:=` 短变量声明
- 使用泛型时优先类型推断
- **不要定义或使用枚举类型**，数据库字段使用 `string` 类型配合 GORM 的 `size` 标签

## 数据库字段规范
- 所有字段使用基本类型（string, int, bool, time.Time 等）
- 固定值的字段使用 `string` + `size` 标签，不用枚举类型
- 示例：`Category string `gorm:"size:20"` // 可选值：frontend, backend`

## API 设计规范
- 只使用 GET 和 POST 请求
- GET 用于查询操作
- POST 用于创建、更新、删除操作
- 删除接口使用 POST + `/delete` 路径，如：`POST /api/tasks/:id/delete`

## AI 助手行为准则
- 直接修改代码，无需详细解释
- 输出简洁，避免冗余
- 完成后简短回复"完成"或"已修改"
- 不要输出代码片段预览，直接编辑文件

## MCP 开发规范
- MCP Server 实现参考 `internal/mcp/server.go`
- MCP 配置示例参考 `docs/mcp-config-examples.json`

## 数据库文档
- SQL 脚本位于 `docs/sql/`
- UTF8 配置参考 `docs/UTF8_CONFIGURATION.md`
