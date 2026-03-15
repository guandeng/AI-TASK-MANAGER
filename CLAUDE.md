# AI Task Manager - 项目协作规范

## 技术栈
- 后端: Go + Gin + GORM
- 前端: Vue 3 + TypeScript + Element Plus

## API 设计规范
- 只使用 GET 和 POST 请求
- GET 用于查询操作
- POST 用于创建、更新、删除操作
- 删除接口使用 POST + `/delete` 路径，如：`POST /api/tasks/:id/delete`

## AI 助手行为准则
- 直接修改代码，无需详细解释
- 输出简洁，避免冗余
- 完成后简短回复"完成"或"已修改"
- 不要输出代��片段预览，直接编辑文件
