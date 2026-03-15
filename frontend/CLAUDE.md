# AI Task Manager - 前端协作规范

## 技术栈
- Vue 3 + TypeScript
- Element Plus
- Pinia + Vue Router

## 代码规范
- 使用 TypeScript 严格模式
- 组件使用 Composition API + `<script setup>` 语法
- 使用驼峰命名（camelCase）：`codeInterface`、`acceptanceCriteria`、`relatedFiles`
- API 请求和响应都使用驼峰命名

## 数据库字段映射
后端返回的蛇形命名字段（如 `code_interface`）需要在响应拦截器中转换为驼峰命名（如 `codeInterface`）

## JSON 字段存储
以下字段在数据库中以 JSON 字符串存储，前端传对象/数组：
- `codeInterface` - 代码接口定义对象
- `acceptanceCriteria` - 验收标准数组
- `relatedFiles` - 关联文件数组

## Vue 响应式数据同步
Watch 监听对象数组变化时，需要比较所有需要响应式更新的字段，确保新增字段能触发 UI 刷新。

## AI 助手行为准则
- 直接修改代码，无需详细解释
- 输出简洁，避免冗余
- 完成后简短回复"完成"或"已修改"
- 不要输出代码片段预览，直接编辑文件
- 按钮不加 icon，保持简洁

## 语言规范
- 项目只支持中文，不支持多语言
- 所有用户可见文本使用中文
- 路由、菜单、按钮、提示等均使用中文
