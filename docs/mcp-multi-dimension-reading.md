# MCP 多维度任务读取 - 完整指南

## 概述

AI Task Manager 的 MCP 服务器现在支持**多维度任务读取**,可以从不同角度、不同项目获取任务信息:

1. ✅ **从需求文档获取** - 直接解析 PRD 生成任务
2. ✅ **从其他项目获取任务** - 跨项目查看和管理任务
3. ✅ **从其他项目获取子任务** - 参考 or 复其他项目的子任务实现
4. ✅ **按条件过滤获取** - 根据状态、优先级、关键词等筛选

---

## 使用场景详解

### 场景 1: 从需求文档生成任务

**场景描述**: 你有一个"UGC需求1.0版本.md"文档,想要生成任务

**AI 对话示例**:
```
用户: 请读取 /projects/ugc-app/docs/UGC需求1.0版本.md 这个需求文档,
     并在 /projects/ugc-app 项目中生成任务

AI: [调用 parsePrd 工具]
    已成功从需求文档生成 8 个任务:
    1. 用户认证系统 (high priority)
    2. UGC 内容上传功能
    3. 内容审核流程
    4. 用户积分系统
    ...
```

**MCP 工具调用**:
```json
{
  "name": "parsePrd",
  "arguments": {
    "prdFile": "/projects/ugc-app/docs/UGC需求1.0版本.md",
    "projectRoot": "/projects/ugc-app",
    "knowledgeBase": "/projects/ugc-app/docs",
    "numTasks": 10
  }
}
```

---

### 场景 2: 从其他项目获取特定任务

**场景描述**: 你在项目 A 工作,想要查看项目 B 的任务 3 的详情

**AI 对话示例**:
```
用户: 帮我查看 /projects/project-b 项目中的任务 3

AI: [调用 showTask 工具]
    任务 3: 支付系统集成
    状态: in-progress
    优先级: high
    依赖: [1, 2]
    描述: 集成第三方支付系统,支持微信支付和支付宝

    子任务:
    3.1: 微信支付集成 (done)
    3.2: 支付宝集成 (pending)
    3.3: 支付回调处理 (pending)
```

**MCP 工具调用**:
```json
{
  "name": "showTask",
  "arguments": {
    "id": "3",
    "projectRoot": "/projects/project-b"
  }
}
```

---

### 场景 3: 从其他项目获取特定子任务 (新功能!)

**场景描述**: 你想要参考项目 B 中任务 2 的子任务 3 的实现细节

**AI 对话示例**:
```
用户: 我想看看 /projects/old-project 项目中任务 2.3 的实现细节

AI: [调用 getSubtask 工具]
    📋 子任务来源: /projects/old-project
    ────────────────────────────────────────────────────

    任务 2.3: 文件上传进度显示
    状态: done
    描述: 实现文件上传时的进度条显示

    实现细节:
    - 使用 XMLHttpRequest 替代 fetch 以支持进度监听
    - 添加进度条 UI 组件
    - 处理上传失败和重试逻辑
    - 支持多文件并发上传

    测试策略:
    - 测试大文件上传 (100MB+)
    - 测试网络中断恢复
    - 测试并发 10 个文件上传
```

**MCP 工具调用**:
```json
{
  "name": "getSubtask",
  "arguments": {
    "sourceProject": "/projects/old-project",
    "taskId": "2.3",
    "includeDetails": true,
    "includeImplementation": true,
    "format": "detailed"
  }
}
```

---

### 场景 4: 获取某任务的所有子任务

**场景描述**: 想要查看项目 B 中任务 5 的所有子任务

**AI 对话示例**:
```
用户: 列出 /projects/shop-app 项目中任务 5 的所有子任务

AI: [调用 getSubtask 工具]
    📋 子任务列表来源: /projects/shop-app
    ────────────────────────────────────────────────────

    任务 5: 购物车功能 的子任务:

    5.1: 购物车数据结构设计 (done)
         优先级: high

    5.2: 添加商品到购物车 (done)
         优先级: high

    5.3: 购物车数量修改 (pending)
         优先级: medium

    5.4: 购物车结算功能 (pending)
         优先级: high

    5.5: 购物车持久化存储 (pending)
         优先级: low
```

**MCP 工具调用**:
```json
{
  "name": "getSubtask",
  "arguments": {
    "sourceProject": "/projects/shop-app",
    "taskId": "5",
    "format": "summary"
  }
}
```

---

### 场景 5: 按条件过滤任务

**场景描述**: 想要查看所有高优先级的待处理任务

**AI 对话示例**:
```
用户: 列出 /projects/my-app 项目中所有高优先级的待处理任务

AI: [调用 getTaskByFilter 工具]
    找到 3 个高优先级待处理任务:

    任务 7: 数据库优化 (high)
    任务 12: 安全漏洞修复 (high)
    任务 15: 支付功能集成 (high)
```

**MCP 工具调用**:
```json
{
  "name": "getTaskByFilter",
  "arguments": {
    "projectRoot": "/projects/my-app",
    "status": "pending",
    "priority": "high"
  }
}
```

---

### 场景 6: 按关键词搜索任务

**场景描述**: 想要查找所有包含"支付"的任务

**AI 对话示例**:
```
用户: 搜索 /projects/my-app 项目中所有包含"支付"的任务

AI: [调用 getTaskByFilter 工具]
    找到 4 个包含"支付"的任务:

    任务 6: 支付宝集成
    任务 7: 微信支付集成
    任务 15: 支付回调处理
    任务 3.2: 支付失败重试机制
```

**MCP 工具调用**:
```json
{
  "name": "getTaskByFilter",
  "arguments": {
    "projectRoot": "/projects/my-app",
    "keyword": "支付",
    "includeSubtasks": true
  }
}
```

---

## 完整工具列表

### 核心工具

| 工具名称        | 功能               | 主要参数                                      |
| --------------- | ------------------ | --------------------------------------------- |
| `parsePrd`      | 从需求文档生成任务 | prdFile, projectRoot, numTasks, knowledgeBase |
| `listTasks`     | 列出所有任务       | projectRoot, status, withSubtasks             |
| `showTask`      | 查看特定任务       | id, projectRoot                               |
| `nextTask`      | 获取下一个推荐任务 | projectRoot                                   |
| `setTaskStatus` | 设置任务状态       | id, status, projectRoot                       |
| `expandTask`    | 展开任务为子任务   | id, projectRoot, num                          |
| `addTask`       | 添加新任务         | prompt, projectRoot, dependencies             |

### 高级工具

| 工具名称          | 功能                 | 主要参数                                       |
| ----------------- | -------------------- | ---------------------------------------------- |
| `getTaskByFilter` | 按条件过滤任务       | projectRoot, status, priority, keyword, taskId |
| `getSubtask`      | 获取其他项目的子任务 | sourceProject, taskId, format, includeDetails  |

---

## 实战工作流

### 工作流 1: 新项目初始化

```
步骤 1: AI,请从 /project/new-app/docs/prd.md 生成任务
步骤 2: 列出所有生成的任务
步骤 3: 展开任务 3,它看起来比较复杂
步骤 4: 将任务 1 标记为开始处理
```

### 工作流 2: 跨项目参考

```
步骤 1: 查看项目 A 中任务 5 的实现
步骤 2: 获取项目 A 中任务 5.2 的详细实现细节
步骤 3: 在当前项目中创建类似任务
步骤 4: 参考 5.2 的实现方式展开当前任务
```

### 工作流 3: 任务管理

```
步骤 1: 列出所有高优先级待处理任务
步骤 2: 获取下一个推荐处理的任务
步骤 3: 查看该任务的依赖是否都已完成
步骤 4: 开始处理任务并更新状态
```

---

## 配置示例

### Cursor AI 配置

```json
{
  "mcpServers": {
    "ai-task-manager": {
      "command": "npx",
      "args": ["-y", "--package=ai-task-manager", "task-manager-mcp-server"],
      "env": {
        "QWEN_API_KEY": "sk-xxx",
        "QWEN_MODEL": "qwen-plus",
        "USE_CHINESE": "true"
      }
    }
  }
}
```

### Claude Code 配置

```json
{
  "mcpServers": {
    "ai-task-manager": {
      "command": "npx",
      "args": ["-y", "--package=ai-task-manager", "task-manager-mcp-server"],
      "env": {
        "GOOGLE_API_KEY": "xxx",
        "GEMINI_MODEL": "gemini-2.5-flash",
        "USE_CHINESE": "true"
      }
    }
  }
}
```

---

## 最佳实践

### 1. 需求文档管理
- 将需求文档统一放在项目的 `docs/` 目录
- 使用清晰的文件命名,如 `需求-v1.0.md`, `PRD-2024-01.md`
- 配合知识库目录提供业务背景

### 2. 跨项目参考
- 建立一个"参考项目"存放通用的任务模板
- 对于类似功能,先查看其他项目的实现
- 使用 JSON 格式获取结构化数据便于程序处理

### 3. 任务过滤
- 使用关键词快速定位相关任务
- 结合状态和优先级筛选,聚焦重要任务
- 定期查看 nextTask 获取智能建议

### 4. 子任务管理
- 复杂任务及时拆分为子任务
- 参考已完成项目的子任务拆分方式
- 保持子任务粒度适中 (2-4小时可完成)

---

## 故障排查

### 问题: 无法读取其他项目的任务

**解决方案**:
- 确认项目路径正确 (使用绝对路径)
- 检查项目中是否存在 `tasks/tasks.json`
- 确认文件权限

### 问题: 需求文档解析失败

**解决方案**:
- 检查文档格式 (支持 txt, md)
- 确认 API Key 有效
- 查看错误日志获取详细信息

### 问题: 中文搜索不生效

**解决方案**:
- 确保设置了 `USE_CHINESE=true`
- 使用准确的关键词
- 尝试同时搜索中英文关键词

---

## 更新日志

### v1.1.0 (2025-03-14)
- ✨ 新增 `parsePrd` 工具 - 支持从需求文档生成任务
- ✨ 新增 `getTaskByFilter` 工具 - 支持多维度过滤任务
- ✨ 新增 `getSubtask` 工具 - 支持跨项目获取子任务
- 🎨 支持中英文双语关键词搜索
- 📝 完善文档和使用示例

### v1.0.0
- 初始版本
- 基础任务管理功能
- MCP 服务器支持
