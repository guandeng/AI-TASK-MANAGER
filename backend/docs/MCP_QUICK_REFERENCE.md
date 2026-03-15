# MCP 多维度任务读取 - 快速参考

## 🎯 三种读取方式

### 1️⃣ 从需求读取 → 生成任务
```javascript
// AI 对话
"读取 /path/to/需求.md 并生成任务"

// MCP 调用
parsePrd({
  prdFile: "/path/to/需求.md",
  projectRoot: "/path/to/project",
  numTasks: 10  // 可选
})
```

### 2️⃣ 从其他项目读取任务
```javascript
// AI 对话
"查看 /other/project 的任务 3"

// MCP 调用
showTask({
  id: "3",
  projectRoot: "/other/project"
})
```

### 3️⃣ 从其他项目读取子任务 ⭐ NEW
```javascript
// AI 对话
"看看 /other/project 任务 2.3 的实现"

// MCP 调用
getSubtask({
  sourceProject: "/other/project",
  taskId: "2.3",
  format: "detailed"  // summary | detailed | json
})
```

## 🔍 过滤和搜索

```javascript
// 按状态筛选
getTaskByFilter({
  projectRoot: "/project",
  status: "pending"  // pending | done | deferred
})

// 按优先级筛选
getTaskByFilter({
  projectRoot: "/project",
  priority: "high"  // high | medium | low
})

// 按关键词搜索
getTaskByFilter({
  projectRoot: "/project",
  keyword: "支付"
})

// 组合筛选
getTaskByFilter({
  projectRoot: "/project",
  status: "pending",
  priority: "high",
  keyword: "API"
})
```

## 📋 常用操作

| 操作 | AI 对话示例 |
|------|------------|
| 生成任务 | "从 docs/需求.md 生成任务" |
| 查看任务 | "显示任务 5 的详情" |
| 查看子任务 | "查看任务 3 的所有子任务" |
| 跨项目参考 | "看看项目 B 的任务 2.1" |
| 搜索 | "搜索所有包含'认证'的任务" |
| 下一步 | "我接下来该做什么?" |
| 更新状态 | "将任务 3 标记为完成" |

## ⚡ 状态感知功能 🆕

所有工具都会**自动检查任务状态**并给出智能提示!

### 状态图标
- ⏱️ pending - 待处理
- 🔄 in-progress - 进行中
- ✅ done - 已完成
- 📌 deferred - 延期
- 🚫 blocked - 阻塞

### 示例输出
```
用户: 查看 /projects/shop 的任务 3

AI: ⏱️ **状态提示**: 任务 "支付系统" 当前状态为 **pending**

    ⚠️ **注意**: 此任务尚未开始。是否需要:
    1. 查看详细信息?
    2. 开始执行?
    3. 查看依赖?

    ---
    [任务详情...]
```

📖 详细说明: [状态感知功能](STATUS_AWARENESS.md)

## 🚀 快速开始

1. 配置 MCP 服务器 (在 Cursor/Claude Code 中)
2. 初始化项目: `task-manager init`
3. 创建需求文档
4. 开始与 AI 对话管理任务!

## 💡 最佳实践

✅ 使用绝对路径确保跨项目访问
✅ 利用知识库增强任务生成质量
✅ 参考已完成项目的子任务实现
✅ 定期查看 nextTask 获取建议
✅ 使用中文关键词搜索 (需设置 USE_CHINESE=true)

## 📚 详细文档

- 完整指南: [docs/mcp-multi-dimension-reading.md](mcp-multi-dimension-reading.md)
- 需求读取: [docs/mcp-requirement-reading.md](mcp-requirement-reading.md)
- README: [README.md](../README.md)
