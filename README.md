# AI Task Manager

一个现代化的 AI 驱动任务管理系统，采用前后端分离架构，支持需求管理、任务拆分、子任务管理、MCP 集成等功能。

![License](https://img.shields.io/badge/license-MIT-blue.svg)
![Go](https://img.shields.io/badge/Go-1.22+-00ADD8?logo=go&logoColor=white)
![Vue](https://img.shields.io/badge/Vue-3.5+-4FC08D?logo=vue.js&logoColor=white)
![MCP](https://img.shields.io/badge/MCP-Supported-FF6B6B?logo=github&logoColor=white)

## ✨ 特性

### 核心功能
- **需求管理** - 支持需求创建、编辑、文档上传、状态追踪
- **任务拆分** - AI 辅助将需求拆分为可执行的任务和子任务
- **任务管理** - 完整的任务生命周期管理，支持依赖关系、优先级、分类
- **子任务管理** - 细粒度的子任务拆分，支持代码接口定义、验收标准
- **成员管理** - 团队成员管理、任务分配、工作负载查看
- **活动追踪** - 完整的操作日志和活动时间线
- **消息通知** - 系统消息和通知管理
- **模板系统** - 项目模板和任务模板，支持快速实例化
- **备份恢复** - 需求数据自动备份和恢复功能
- **复杂度分析** - AI 驱动的任务复杂度评估
- **知识库集成** - 支持业务知识库辅助任务拆分
- **多语言支持** - 中英双语任务内容
- **MCP 服务器** - 支持 Cursor 等 AI 编辑器集成

### 技术特性
- 前后端分离架构
- RESTful API 设计
- 优雅的错误处理和恢复机制
- 支持 MySQL 数据库
- 响应式 Web 界面
- 支持暗色模式
- 完整的单元测试覆盖

## 🏗️ 技术架构

### 后端
| 技术 | 说明 |
|------|------|
| Go 1.22+ | 编程语言 |
| Gin | Web 框架 |
| GORM | ORM 框架 |
| Zap | 日志库 |
| Viper | 配置管理 |
| MCP-Go | MCP 服务器 SDK |
| MySQL | 数据库 |

### 前端
| 技术 | 说明 |
|------|------|
| Vue 3.5+ | 渐进式框架 |
| TypeScript | 类型系统 |
| Vite 7 | 构建工具 |
| Naive UI | UI 组件库 |
| Pinia | 状态管理 |
| Vue Router | 路由管理 |
| UnoCSS | 原子化 CSS |
| ECharts | 数据可视化 |

## 📦 项目结构

```
AI-TASK-MANAGER/
├── backend/                     # 后端服务
│   ├── cmd/
│   │   ├── server/             # 主服务入口
│   │   └── mcp-server/         # MCP 服务器入口
│   ├── internal/
│   │   ├── config/             # 配置管理
│   │   ├── database/           # 数据库初始化
│   │   ├── handlers/           # HTTP 处理器
│   │   ├── middleware/         # 中间件
│   │   ├── models/             # 数据模型
│   │   ├── repository/         # 数据访问层
│   │   ├── services/           # 业务逻辑层
│   │   └── mcp/                # MCP 服务
│   ├── pkg/
│   │   ├── ai/                 # AI 服务封装
│   │   └── response/           # 统一响应格式
│   └── test/                   # 测试文件
├── frontend/                    # 前端应用
│   ├── src/
│   │   ├── api/                # API 请求封装
│   │   ├── components/         # 组件
│   │   ├── composables/        # 组合式函数
│   │   ├── layouts/            # 布局
│   │   ├── router/             # 路由配置
│   │   ├── store/              # 状态管理
│   │   ├── views/              # 页面视图
│   │   └── utils/              # 工具函数
│   └── public/                 # 静态资源
└── package.json                 # 项目配置
```

## 🚀 快速开始

### 环境要求
- Go 1.22+
- Node.js 18.0+
- pnpm 10.5+
- MySQL 8.0+

### 安装

```bash
# 克隆项目
git clone https://github.com/skindhu/AI-TASK-MANAGER.git
cd AI-TASK-MANAGER

# 安装前端依赖
cd frontend
pnpm install

# 安装后端依赖
cd ../backend
go mod download
```

### 配置

#### 后端配置

在 `backend` 目录下创建 `.env` 文件或复制 `.env.example`：

```env
# 服务器配置
SERVER_PORT=8080
SERVER_MODE=debug

# 数据库配置
DB_HOST=localhost
DB_PORT=3306
DB_USER=root
DB_PASSWORD=your_password
DB_NAME=ai_task_manager

# AI 服务配置 (可选)
AI_PROVIDER=qwen  # 或 gemini
QWEN_API_KEY=your_qwen_api_key
QWEN_MODEL=qwen-plus
QWEN_BASE_URL=https://dashscope.aliyuncs.com/compatible-mode/v1

# 或 Gemini 配置
# GOOGLE_API_KEY=your_google_api_key
# GEMINI_MODEL=gemini-2.5-pro
```

#### 数据库初始化

```sql
CREATE DATABASE ai_task_manager CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
```

### 启动服务

#### 开发模式

```bash
# 方式一：使用根目录脚本
# 启动前端（开发模式）
npm run dev

# 启动后端（新开终端）
npm run dev:backend

# 方式二：分别启动
# 前端
cd frontend
pnpm dev

# 后端
cd backend
go run ./cmd/server
```

#### 生产构建

```bash
# 构建前端
cd frontend
pnpm build

# 构建后端
cd backend
go build -o bin/server ./cmd/server

# 启动后端服务
./bin/server
```

### 访问应用

- 前端：http://localhost:5173
- 后端 API：http://localhost:8080/api
- 健康检查：http://localhost:8080/health

## 🔌 API 接口

### 需求管理
| 方法 | 路径 | 说明 |
|------|------|------|
| GET | /api/requirements | 获取需求列表 |
| GET | /api/requirements/:id | 获取需求详情 |
| POST | /api/requirements | 创建需求 |
| POST | /api/requirements/:id/update | 更新需求 |
| POST | /api/requirements/:id/delete | 删除需求 |
| POST | /api/requirements/:id/split-tasks | AI 拆分任务 |
| GET | /api/requirements/:id/structure | 获取需求结构 |

### 任务管理
| 方法 | 路径 | 说明 |
|------|------|------|
| GET | /api/tasks | 获取任务列表 |
| GET | /api/tasks/:id | 获取任务详情 |
| POST | /api/tasks | 创建任务 |
| POST | /api/tasks/:id/update | 更新任务 |
| POST | /api/tasks/:id/delete | 删除任务 |
| POST | /api/tasks/:id/expand | AI 展开子任务 |
| GET | /api/tasks/ready | 获取可执行任务 |
| POST | /api/tasks/:id/score | 任务质量评分 |

### 子任务管理
| 方法 | 路径 | 说明 |
|------|------|------|
| POST | /api/tasks/:id/subtasks/:subtaskId/update | 更新子任务 |
| POST | /api/tasks/:id/subtasks/:subtaskId/delete | 删除子任务 |
| POST | /api/tasks/:id/subtasks/reorder | 重排子任务 |

### 成员管理
| 方法 | 路径 | 说明 |
|------|------|------|
| GET | /api/members | 获取成员列表 |
| POST | /api/members | 创建成员 |
| POST | /api/members/:id/update | 更新成员 |
| GET | /api/members/:id/workload | 工作负载 |

### MCP 工具

系统提供以下 MCP 工具供 AI 编辑器调用：

| 工具 | 说明 |
|------|------|
| list_tasks | 列出任务 |
| show_task | 显示任务详情 |
| set_task_status | 设置任务状态 |
| expand_task | 展开任务为子任务 |
| next_task | 获取下一个待办任务 |
| add_task | 添加新任务 |
| update_task | 更新任务信息 |
| get_task_with_comments | 获取任务及评论 |
| validate_dependencies | 验证依赖关系 |
| get_ready_tasks | 获取可执行任务 |
| search_requirements | 搜索需求 |
| get_requirement_tasks | 获取需求下的任务 |

### Cursor MCP 配置

在 Cursor 设置中添加以下配置：

```json
{
  "mcpServers": {
    "ai-task-manager": {
      "command": "node",
      "args": ["/path/to/backend/cmd/mcp-server/main.go"],
      "env": {
        "QWEN_API_KEY": "your_api_key"
      }
    }
  }
}
```

## 📊 数据模型

### 核心实体关系

```
Requirement (需求)
    └── Task (任务)
            ├── Subtask (子任务)
            ├── Assignment (分配)
            ├── Comment (评论)
            └── Dependency (依赖)
```

### 主要模型

#### Task (任务)
- `id` - 主键
- `requirement_id` - 所属需求
- `title` - 标题
- `description` - 描述
- `status` - 状态 (pending/in_progress/done/cancelled)
- `priority` - 优先级 (high/medium/low)
- `category` - 分类 (frontend/backend)
- `details` - 详情
- `acceptance_criteria` - 验收标准
- `module` - 模块归属

#### Subtask (子任务)
- `id` - 主键
- `task_id` - 所属任务
- `title` - 标题
- `description` - 描述
- `status` - 状态
- `code_interface` - 代码接口定义 (JSON)
- `acceptance_criteria` - 验收标准
- `related_files` - 关联文件
- `code_hints` - 代码提示

## 📸 界面预览

### 需求列表
需求管理界面支持创建、编辑、删除需求，查看需求状态和任务统计。

### 任务看板
看板视图展示任务的不同状态，支持拖拽操作。

### 任务详情
任务详情页展示任务信息、子任务列表、评论、分配情况等。

### 依赖关系图
可视化展示任务之间的依赖关系。

## 🧪 测试

### 后端测试
```bash
cd backend
go test ./... -v
```

### 前端测试
```bash
cd frontend
pnpm test
```

## 📝 开发规范

### Go 代码规范
- 使用 `gofmt` 格式化代码
- 遵循 Effective Go 指南
- 使用 `any` 代替 `interface{}`
- 使用 `:=` 短变量声明

### 前端代码规范
- 使用 ESLint + Prettier 代码格式
- TypeScript 严格模式
- 组件使用 Composition API
- 使用 ESLint 进行代码检查

### 提交规范
```bash
feat: 新功能
fix: 修复 bug
docs: 文档更新
style: 代码格式调整
refactor: 重构代码
test: 测试相关
chore: 构建/工具链相关
```

## 🔧 配置说明

### 环境变量

| 变量 | 说明 | 默认值 |
|------|------|--------|
| SERVER_PORT | 服务端口 | 8080 |
| SERVER_MODE | 运行模式 (debug/release) | debug |
| DB_HOST | 数据库主机 | localhost |
| DB_PORT | 数据库端口 | 3306 |
| DB_USER | 数据库用户 | root |
| DB_PASSWORD | 数据库密码 | - |
| DB_NAME | 数据库名 | ai_task_manager |
| AI_PROVIDER | AI 提供商 (qwen/gemini) | qwen |
| QWEN_API_KEY | 千问 API Key | - |
| QWEN_MODEL | 千问模型 | qwen-plus |
| MAX_TOKENS | 最大 Token 数 | 8192 |
| TEMPERATURE | 模型温度 | 0.7 |

## 🤝 参与贡献

欢迎提交 Issue 和 Pull Request！

1. Fork 本项目
2. 创建特性分支 (`git checkout -b feature/AmazingFeature`)
3. 提交更改 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 创建 Pull Request

## 📄 开源协议

本项目采用 MIT 协议开源 - 查看 [LICENSE](LICENSE) 文件了解详情。

## 👥 作者

- **skindhu** - [GitHub](https://github.com/skindhu)

## 🙏 致谢

本项目基于以下开源项目构建：

- [Gin](https://github.com/gin-gonic/gin)
- [GORM](https://github.com/go-gorm/gorm)
- [Vue.js](https://github.com/vuejs/core)
- [Naive UI](https://github.com/tusen-ai/naive-ui)
- [Soybean Admin](https://github.com/soybeanjs/soybean-admin)
- [MCP Go](https://github.com/mark3labs/mcp-go)

## 📮 联系方式

如有问题或建议，请通过以下方式联系：

- GitHub Issues: [提交问题](https://github.com/skindhu/AI-TASK-MANAGER/issues)
- Email: [发送邮件](mailto:your-email@example.com)

---

如果这个项目对你有帮助，请给一个 ⭐️ Star 支持！
