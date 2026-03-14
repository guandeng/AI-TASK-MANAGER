# MCP 跨项目需求读取功能说明

## 功能概述

AI Task Manager 现在支持通过 MCP (Model Context Protocol) 跨项目读取需求文档并生成任务。这意味着你可以在一个项目中管理多个项目的需求,或者让 AI 助手从不同的项目目录读取需求。

## 使用场景

### 场景 1: 从其他项目的需求文档生成任务

你正在项目 A 中工作,但想要读取项目 B 的需求文档"UGC需求1.0版本.md"并生成任务:

```
AI 助手,请读取 /path/to/project-b/docs/UGC需求1.0版本.md 这个需求文档,
并在项目 B 中生成对应的任务。
```

### 场景 2: 用业务知识库增强任务生成

```
请读取需求文档 /path/to/project/prd.txt,
同时参考 /path/to/project/docs/业务知识库 目录,
生成更符合业务背景的任务。
```

### 场景 3: 查看已生成的任务和状态

```
请列出项目 /path/to/project 中的所有任务及其状态
```

## MCP 工具列表

### 1. parsePrd - 从需求文档生成任务

**参数:**
- `prdFile` (必需): 需求文档的路径 (可以是绝对路径或相对于项目根目录的相对路径)
- `projectRoot` (必需): 项目的根目录路径
- `numTasks` (可选): 要生成的任务数量 (如果不指定,AI 会根据复杂度自动决定)
- `knowledgeBase` (可选): 业务知识库路径 (文件或目录)
- `outputFile` (可选): 输出文件路径 (默认: tasks/tasks.json)

**示例调用:**
```json
{
  "name": "parsePrd",
  "arguments": {
    "prdFile": "/Users/xxx/projects/my-app/docs/UGC需求1.0版本.md",
    "projectRoot": "/Users/xxx/projects/my-app",
    "knowledgeBase": "/Users/xxx/projects/my-app/docs",
    "numTasks": 10
  }
}
```

### 2. listTasks - 列出所有任务

**参数:**
- `projectRoot` (必需): 项目的根目录路径
- `status` (可选): 按状态筛选任务 (pending, done, deferred)
- `withSubtasks` (可选): 是否包含子任务
- `file` (可选): 任务文件路径

**示例:**
```json
{
  "name": "listTasks",
  "arguments": {
    "projectRoot": "/Users/xxx/projects/my-app",
    "status": "pending",
    "withSubtasks": true
  }
}
```

### 3. showTask - 查看特定任务详情

**参数:**
- `id` (必需): 任务 ID (例如: 1 或 1.2 表示子任务)
- `projectRoot` (必需): 项目的根目录路径
- `file` (可选): 任务文件路径

### 4. setTaskStatus - 设置任务状态

**参数:**
- `id` (必需): 任务 ID
- `status` (必需): 新状态 (pending, done, deferred)
- `projectRoot` (必需): 项目的根目录路径
- `file` (可选): 任务文件路径

### 5. nextTask - 获取下一个待处理任务

**参数:**
- `projectRoot` (必需): 项目的根目录路径
- `file` (可选): 任务文件路径

### 6. expandTask - 展开任务为子任务

**参数:**
- `id` (必需): 要展开的任务 ID
- `projectRoot` (必需): 项目的根目录路径
- `num` (可选): 子任务数量
- `prompt` (可选): 额外的上下文提示
- `research` (可选): 是否使用研究支持
- `file` (可选): 任务文件路径

### 7. addTask - 添加新任务

**参数:**
- `prompt` (必需): 任务描述
- `projectRoot` (必需): 项目的根目录路径
- `dependencies` (可选): 依赖的任务 ID 列表
- `priority` (可选): 优先级 (high, medium, low)
- `file` (可选): 任务文件路径

## 完整工作流示例

### 示例 1: 新项目初始化

```
1. 用户: 我想为项目 /path/to/my-ugc-project 初始化任务管理

2. AI: 好的,我来帮你读取该项目的需求文档并生成任务
   [调用 parsePrd 工具]

3. AI: 已经从需求文档生成了 8 个任务:
   - 任务1: 用户认证系统 (pending)
   - 任务2: UGC 内容上传 (pending)
   - 任务3: 内容审核流程 (pending)
   ...

4. 用户: 请展开任务2

5. AI: [调用 expandTask 工具]
   任务2已拆分为 3 个子任务:
   - 2.1: 实现图片上传接口
   - 2.2: 实现视频上传接口
   - 2.3: 添加文件大小限制和格式验证
```

### 示例 2: 跨项目管理

```
用户: 我在项目 A 工作,但需要查看项目 B 的任务状态

AI: [调用 listTasks 工具,指定 projectRoot 为项目 B 的路径]
项目 B 当前有 15 个任务:
- 已完成: 8 个
- 进行中: 4 个
- 待处理: 3 个

下一个推荐任务: 任务 9 - 实现支付功能
```

## 在 Cursor AI 中使用

### 配置 MCP 服务器

在 Cursor 设置中添加 MCP 服务器配置:

**使用千问 (推荐):**
```json
{
  "mcpServers": {
    "ai-task-manager": {
      "command": "npx",
      "args": ["-y", "--package=ai-task-manager", "task-manager-mcp-server"],
      "env": {
        "QWEN_API_KEY": "YOUR_QWEN_API_KEY",
        "QWEN_MODEL": "qwen-plus"
      }
    }
  }
}
```

**使用 Gemini:**
```json
{
  "mcpServers": {
    "ai-task-manager": {
      "command": "npx",
      "args": ["-y", "--package=ai-task-manager", "task-manager-mcp-server"],
      "env": {
        "GOOGLE_API_KEY": "YOUR_GOOGLE_API_KEY",
        "GEMINI_BASE_URL": "可选的代理地址"
      }
    }
  }
}
```

### 在对话中使用

置完成后,你可以直接在 Cursor 的 AI 对话中:

1. **读取需求生成任务**
   ```
   请读取 /path/to/project/docs/需求文档.md 并生成任务
   ```

2. **查看任务状态**
   ```
   列出 /path/to/project 的所有任务
   ```

3. **更新任务状态**
   ```
   将项目 /path/to/project 的任务 3 标记为完成
   ```

4. **获取下一步建议**
   ```
   项目 /path/to/project 下一步应该做什么?
   ```

## 优势

1. **跨项目管理**: 在一个 AI 会话中管理多个项目的任务
2. **需求驱动**: 直接从需求文档生成结构化任务
3. **智能拆分**: AI 根据需求复杂度自动决定任务粒度
4. **业务上下文**: 支持引入业务知识库,生成更准确的任务
5. **状态追踪**: 实时查看和更新任务状态
6. **依赖管理**: 自动处理任务间的依赖关系

## 注意事项

1. 确保提供的项目路径是有效的
2. 需求文档应该是文本格式 (txt, md 等)
3. 首次使用需要在项目中运行 `task-manager init` 初始化
4. 任务数据存储在项目的 `tasks/tasks.json` 文件中
5. MCP 服务器需要配置正确的 AI 服务提供商 API Key

## 故障排查

### 问题 1: MCP 工具调用失败

**解决方案:**
- 检查 MCP 服务器配置是否正确
- 确认 API Key 是否有效
- 查看 Cursor 的开发者控制台获取详细错误信息

### 问题 2: 找不到需求文档

**解决方案:**
- 使用绝对路径而不是相对路径
- 确认文件路径和文件名拼写正确
- 检查文件权限

### 问题 3: 任务生成失败

**解决方案:**
- 确认项目已初始化 (存在 tasks 目录)
- 检查需求文档格式是否正确
- 查看错误日志获取详细信息
