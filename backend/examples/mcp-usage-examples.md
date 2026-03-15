# MCP 使用示例

这个文件包含了各种 MCP 工具的使用示例,帮助你快速上手。

## 示例 1: 从需求文档生成任务

### 场景
你有一个 UGC 应用的需求文档,想要生成任务。

### AI 对话
```
用户: 请读取 /Users/guandeng/projects/ugc-app/docs/UGC需求1.0版本.md
     并在 /Users/guandeng/projects/ugc-app 项目中生成任

AI: 好的,我来解析需求文档并生成任务...

[MCP 调用: parsePrd]

✅ 已成功从需求文档生成 8 个任务:

1. 用户认证系统 (high priority)
   - 状态: pending
   - 描述: 实现用户注册、登录、JWT认证

2. UGC 内容上传 (high priority)
   - 状态: pending
   - 描述: 支持图片、视频上传,大小限制100MB

3. 内容审核流程 (medium priority)
   - 状态: pending
   - 描述: 实现人工审核和AI自动审核

... 更多任务

下一步建议: 开始处理任务 1 - 用户认证系统
```

### MCP 工具调用
```json
{
  "name": "parsePrd",
  "arguments": {
    "prdFile": "/Users/guandeng/projects/ugc-app/docs/UGC需求1.0版本.md",
    "projectRoot": "/Users/guandeng/projects/ugc-app",
    "numTasks": 10
  }
}
```

---

## 示例 2: 查看其他项目的任务

### 场景
你在项目 A 工作,想要查看项目 B 的某个任务详情。

### AI 对话
```
用户: 帮我看看 /Users/guandeng/projects/shop-app 的任务 3

AI: [MCP 调用: showTask]

📋 任务 3: 支付系统集成
──────────────────────────────────
状态: in-progress
优先级: high
依赖: [1, 2] ✅ 已完成

描述:
集成第三方支付系统,支持微信支付和支付宝

实现细节:
1. 对接微信支付API v3
2. 对接支付宝开放平台API
3. 实现统一的支付抽象层
4. 处理支付回调通知
5. 实现订单超时取消

测试策略:
- 单元测试: Mock支付接口
- 集成测试: 使用沙箱环境
- 压力测试: 模拟1000并发支付

子任务:
├─ 3.1 微信支付集成 (done)
├─ 3.2 支付宝集成 (pending)
└─ 3.3 支付回调处理 (pending)
```

### MCP 工具调用
```json
{
  "name": "showTask",
  "arguments": {
    "id": "3",
    "projectRoot": "/Users/guandeng/projects/shop-app"
  }
}
```

---

## 示例 3: 获取其他项目的子任务实现

### 场景
你想要参考另一个项目中"文件上传"子任务的实现方式。

### AI 对话
```
用户: 我想看看 /Users/guandeng/projects/old-app 任务 2.3 的实现细节

AI: [MCP 调用: getSubtask]

📋 子任务来源: /Users/guandeng/projects/old-app
────────────────────────────────────────────────

任务 2.3: 文件上传进度显示
状态: done
优先级: medium

描述:
实现文件上传时的实时进度条显示,支持大文件和批量上传

实现细节:
1. 使用 XMLHttpRequest 替代 fetch 以支持上传进度监听
   - 监听 xhr.upload.onprogress 事件
   - 计算 percentage = (loaded / total) * 100

2. 添加进度条 UI 组件
   - 使用 NProgress 库显示顶部进度条
   - 支持多个文件同时上传时分别显示进度

3. 处理上传失败和重试逻辑
   - 失败后自动重试 3 次
   - 指数退避: 1s, 2s, 4s

4. 支持多文件并发上传
   - 使用 Promise.all 并发处理
   - 限制最大并发数为 3

测试策略:
- ✅ 测试大文件上传 (100MB+)
- ✅ 测试网络中断恢复
- ✅ 测试并发 10 个文件上传
- ✅ 测试进度条准确性

关键代码片段:
```javascript
const upload = (file, onProgress) => {
  return new Promise((resolve, reject) => {
    const xhr = new XMLHttpRequest();
    xhr.upload.onprogress = (e) => {
      if (e.lengthComputable) {
        const percent = (e.loaded / e.total) * 100;
        onProgress(percent);
      }
    };
    xhr.onload = () => resolve(xhr.response);
    xhr.onerror = () => reject(xhr.statusText);
    xhr.open('POST', '/api/upload', true);
    xhr.send(file);
  });
};
```

💡 建议: 可以在当前项目中复用这个实现模式
```

### MCP 工具调用
```json
{
  "name": "getSubtask",
  "arguments": {
    "sourceProject": "/Users/guandeng/projects/old-app",
    "taskId": "2.3",
    "includeDetails": true,
    "includeImplementation": true,
    "format": "detailed"
  }
}
```

---

## 示例 4: 按条件过滤任务

### 场景
想要查看所有高优先级的待处理任务。

### AI 对话
```
用户: 列出 /Users/guandeng/projects/my-app 中所有高优先级待处理任务

AI: [MCP 调用: getTaskByFilter]

找到 3 个高优先级待处理任务:

🔴 任务 7: 数据库优化 (high)
   依赖: 无
   描述: 优化慢查询,添加索引

🔴 任务 12: 安全漏洞修复 (high)
   依赖: 无
   描述: 修复 XSS 和 CSRF 漏洞

🔴 任务 15: 支付功能集成 (high)
   依赖: [7, 12] ⏱️ 待处理
   描述: 集成第三方支付

建议处理顺序:
1. 任务 7 或 任务 12 (无依赖,可并行)
2. 任务 15 (需要先完成 7 和 12)
```

### MCP 工具调用
```json
{
  "name": "getTaskByFilter",
  "arguments": {
    "projectRoot": "/Users/guandeng/projects/my-app",
    "status": "pending",
    "priority": "high"
  }
}
```

---

## 示例 5: 按关键词搜索任务

### 场景
想要查找所有包含"支付"相关内容的任务。

### AI 对话
```
用户: 搜索项目 /Users/guandeng/projects/my-app 中所有包含"支付"的任务

AI: [MCP 调用: getTaskByFilter with keyword]

找到 4 个包含"支付"的任务:

任务 6: 支付宝集成
  状态: done ✅
  描述: 集成支付宝支付功能

任务 7: 微信支付集成
  状态: in-progress 🔄
  描述: 集成微信支付功能

任务 15: 支付回调处理
  状态: pending ⏱️
  描述: 处理支付成功/失败的回调通知

任务 3.2: 支付失败重试机制
  状态: pending ⏱️
  描述: 实现支付失败后的自动重试逻辑
```

### MCP 工具调用
```json
{
  "name": "getTaskByFilter",
  "arguments": {
    "projectRoot": "/Users/guandeng/projects/my-app",
    "keyword": "支付",
    "includeSubtasks": true
  }
}
```

---

## 示例 6: 获取 JSON 格式的任务数据

### 场景
想要以 JSON 格式获取任务数据,便于程序处理。

### AI 对话
```
用户: 以 JSON 格式返回 /Users/guandeng/projects/my-app 任务 2 的所有子任务

AI: [MCP 调用: getSubtask with format=json]

```json
[
  {
    "id": 1,
    "title": "Design database schema",
    "description": "Create database tables and relationships",
    "status": "done",
    "priority": "high",
    "dependencies": [],
    "details": "Use PostgreSQL with proper indexing",
    "testStrategy": "Verify schema with sample data"
  },
  {
    "id": 2,
    "title": "Implement API endpoints",
    "description": "Create RESTful API endpoints",
    "status": "pending",
    "priority": "high",
    "dependencies": [1],
    "details": "Use Express.js with proper error handling",
    "testStrategy": "Unit tests for each endpoint"
  },
  {
    "id": 3,
    "title": "Add authentication",
    "description": "Implement JWT authentication",
    "status": "pending",
    "priority": "medium",
    "dependencies": [2],
    "details": "Use passport-jwt strategy",
    "testStrategy": "Test with valid and invalid tokens"
  }
]
```
```

### MCP 工具调用
```json
{
  "name": "getSubtask",
  "arguments": {
    "sourceProject": "/Users/guandeng/projects/my-app",
    "taskId": "2",
    "format": "json"
  }
}
```

---

## 完整工作流示例

### 场景: 新项目初始化和任务管理

```
第 1 步: 从需求生成任务
用户: 请从 /projects/new-app/docs/prd.md 生成任务
AI: [parsePrd] 已生成 10 个任务

第 2 步: 查看生成的任务
用户: 列出所有任务
AI: [listTasks] 显示 10 个任务列表

第 3 步: 展开复杂任务
用户: 任务 3 看起来很复杂,请展开它
AI: [expandTask] 任务 3 已拆分为 4 个子任务

第 4 步: 开始工作
用户: 我该从哪个任务开始?
AI: [nextTask] 建议从任务 1 开始,因为没有依赖

第 5 步: 参考其他项目
用户: 看看 /projects/old-app 的任务 2.1 是怎么实现的
AI: [getSubtask] 显示任务 2.1 的实现细节

第 6 步: 更新状态
用户: 任务 1 已完成
AI: [setTaskStatus] 任务 1 已标记为完成

第 7 步: 继续下一个
用户: 下一步做什么?
AI: [nextTask] 任务 2 现在可以开始了
```

---

## 提示和技巧

### 💡 使用绝对路径
```
✅ /Users/username/projects/my-app
❌ ../my-app 或 ~/projects/my-app
```

### 💡 善用关键词搜索
```
# 同时搜索中英文
"搜索包含 'payment' 或 '支付' 的任务"

# 使用技术术语
"搜索包含 'JWT' 的任务"
```

### 💡 结合多个工具
```
1. 先用 getTaskByFilter 找到感兴趣的任务
2. 再用 showTask 查看详情
3. 最后用 getSubtask 参考其他项目的实现
```

### 💡 定期获取建议
```
# 每天开始工作时
"我应该从哪个任务开始?"

# 完成一个任务后
"下一步做什么?"
```
