# AI Task Manager

A modern AI-driven task management system with a separated frontend and backend architecture, supporting requirement management, task breakdown, subtask management, and MCP integration.

![License](https://img.shields.io/badge/license-AGPL--3.0-blue.svg)
![Go](https://img.shields.io/badge/Go-1.22+-00ADD8?logo=go&logoColor=white)
![Vue](https://img.shields.io/badge/Vue-3.5+-4FC08D?logo=vue.js&logoColor=white)
![MCP](https://img.shields.io/badge/MCP-Supported-FF6B6B?logo=github&logoColor=white)

## ✨ Features

### Core Features
- **Requirement Management** - Create, edit, upload documents, and track requirement status
- **Task Breakdown** - AI-assisted breakdown of requirements into executable tasks and subtasks
- **Task Management** - Complete task lifecycle management with dependencies, priorities, and categories
- **Subtask Management** - Fine-grained subtask breakdown with code interface definitions and acceptance criteria
- **Member Management** - Team member management, task assignment, workload viewing
- **Activity Tracking** - Complete operation logs and activity timelines
- **Message Notifications** - System messages and notification management
- **Template System** - Project and task templates with quick instantiation
- **Backup & Recovery** - Automatic requirement data backup and recovery
- **Complexity Analysis** - AI-driven task complexity assessment
- **Knowledge Base Integration** - Business knowledge base support for task breakdown
- **Multi-language Support** - Chinese and English bilingual task content
- **MCP Server** - Cursor and other AI editor integration

### Technical Features
- Separated frontend and backend architecture
- RESTful API design
- Graceful error handling and recovery
- MySQL database support
- Responsive web interface
- Dark mode support
- Comprehensive unit test coverage

## 🏗️ Technical Architecture

### Backend
| Technology | Description |
|------------|-------------|
| Go 1.22+ | Programming Language |
| Gin | Web Framework |
| GORM | ORM Framework |
| Zap | Logging Library |
| Viper | Configuration Management |
| MCP-Go | MCP Server SDK |
| MySQL | Database |

### Frontend
| Technology | Description |
|------------|-------------|
| Vue 3.5+ | Progressive Framework |
| TypeScript | Type System |
| Vite 7 | Build Tool |
| Naive UI | UI Component Library |
| Pinia | State Management |
| Vue Router | Routing Management |
| UnoCSS | Atomic CSS |
| ECharts | Data Visualization |

## 📦 Project Structure

```
AI-TASK-MANAGER/
├── backend/                     # Backend Service
│   ├── cmd/
│   │   ├── server/             # Main Service Entry
│   │   └── mcp-server/         # MCP Server Entry
│   ├── internal/
│   │   ├── config/             # Configuration
│   │   ├── database/           # Database Initialization
│   │   ├── handlers/           # HTTP Handlers
│   │   ├── middleware/         # Middleware
│   │   ├── models/             # Data Models
│   │   ├── repository/         # Data Access Layer
│   │   ├── services/           # Business Logic Layer
│   │   └── mcp/                # MCP Services
│   ├── pkg/
│   │   ├── ai/                 # AI Service Wrapper
│   │   └── response/           # Unified Response Format
│   └── test/                   # Test Files
├── frontend/                    # Frontend Application
│   ├── src/
│   │   ├── api/                # API Request Wrapper
│   │   ├── components/         # Components
│   │   ├── composables/        # Composable Functions
│   │   ├── layouts/            # Layouts
│   │   ├── router/             # Router Configuration
│   │   ├── store/              # State Management
│   │   ├── views/              # Page Views
│   │   └── utils/              # Utility Functions
│   └── public/                 # Static Assets
└── package.json                 # Project Configuration
```

## 🚀 Quick Start

### Prerequisites
- Go 1.22+
- Node.js 18.0+
- pnpm 10.5+
- MySQL 8.0+

### Installation

```bash
# Clone the repository
git clone https://github.com/skindhu/AI-TASK-MANAGER.git
cd AI-TASK-MANAGER

# Install frontend dependencies
cd frontend
pnpm install

# Install backend dependencies
cd ../backend
go mod download
```

### Configuration

#### Backend Configuration

Create a `.env` file in the `backend` directory or copy from `.env.example`:

```env
# Server Configuration
SERVER_PORT=8080
SERVER_MODE=debug

# Database Configuration
DB_HOST=localhost
DB_PORT=3306
DB_USER=root
DB_PASSWORD=your_password
DB_NAME=ai_task_manager

# AI Service Configuration (Optional)
AI_PROVIDER=qwen  # or gemini
QWEN_API_KEY=your_qwen_api_key
QWEN_MODEL=qwen-plus
QWEN_BASE_URL=https://dashscope.aliyuncs.com/compatible-mode/v1

# Or Gemini Configuration
# GOOGLE_API_KEY=your_google_api_key
# GEMINI_MODEL=gemini-2.5-pro
```

#### Database Initialization

```sql
CREATE DATABASE ai_task_manager CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
```

### Start Services

#### Development Mode

```bash
# Option 1: Use root scripts
# Start frontend (development mode)
npm run dev

# Start backend (new terminal)
npm run dev:backend

# Option 2: Start separately
# Frontend
cd frontend
pnpm dev

# Backend
cd backend
go run ./cmd/server
```

#### Production Build

```bash
# Build frontend
cd frontend
pnpm build

# Build backend
cd backend
go build -o bin/server ./cmd/server

# Start backend service
./bin/server
```

### Access the Application

- Frontend: http://localhost:5173
- Backend API: http://localhost:8080/api
- Health Check: http://localhost:8080/health

## 🔌 API Endpoints

### Requirement Management
| Method | Path | Description |
|--------|------|-------------|
| GET | /api/requirements | Get requirement list |
| GET | /api/requirements/:id | Get requirement details |
| POST | /api/requirements | Create requirement |
| POST | /api/requirements/:id/update | Update requirement |
| POST | /api/requirements/:id/delete | Delete requirement |
| POST | /api/requirements/:id/split-tasks | AI split tasks |
| GET | /api/requirements/:id/structure | Get requirement structure |

### Task Management
| Method | Path | Description |
|--------|------|-------------|
| GET | /api/tasks | Get task list |
| GET | /api/tasks/:id | Get task details |
| POST | /api/tasks | Create task |
| POST | /api/tasks/:id/update | Update task |
| POST | /api/tasks/:id/delete | Delete task |
| POST | /api/tasks/:id/expand | AI expand subtasks |
| GET | /api/tasks/ready | Get ready tasks |
| POST | /api/tasks/:id/score | Task quality score |

### Subtask Management
| Method | Path | Description |
|--------|------|-------------|
| POST | /api/tasks/:id/subtasks/:subtaskId/update | Update subtask |
| POST | /api/tasks/:id/subtasks/:subtaskId/delete | Delete subtask |
| POST | /api/tasks/:id/subtasks/reorder | Reorder subtasks |

### Member Management
| Method | Path | Description |
|--------|------|-------------|
| GET | /api/members | Get member list |
| POST | /api/members | Create member |
| POST | /api/members/:id/update | Update member |
| GET | /api/members/:id/workload | Workload |

### MCP Tools

The system provides the following MCP tools for AI editor integration:

| Tool | Description |
|------|-------------|
| list_tasks | List tasks |
| show_task | Show task details |
| set_task_status | Set task status |
| expand_task | Expand task into subtasks |
| next_task | Get next pending task |
| add_task | Add new task |
| update_task | Update task information |
| get_task_with_comments | Get task with comments |
| validate_dependencies | Validate dependencies |
| get_ready_tasks | Get executable tasks |
| search_requirements | Search requirements |
| get_requirement_tasks | Get tasks under requirement |

### Cursor MCP Configuration

Add the following configuration to Cursor settings:

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

## 📊 Data Models

### Core Entity Relationships

```
Requirement
    └── Task
            ├── Subtask
            ├── Assignment
            ├── Comment
            └── Dependency
```

### Main Models

#### Task
- `id` - Primary key
- `requirement_id` - Parent requirement
- `title` - Title
- `description` - Description
- `status` - Status (pending/in_progress/done/cancelled)
- `priority` - Priority (high/medium/low)
- `category` - Category (frontend/backend)
- `details` - Details
- `acceptance_criteria` - Acceptance criteria
- `module` - Module assignment

#### Subtask
- `id` - Primary key
- `task_id` - Parent task
- `title` - Title
- `description` - Description
- `status` - Status
- `code_interface` - Code interface definition (JSON)
- `acceptance_criteria` - Acceptance criteria
- `related_files` - Related files
- `code_hints` - Code hints

## 📸 Screenshots

### Requirement List
Requirement management interface supports creating, editing, deleting requirements, viewing requirement status and task statistics.

### Task Kanban
Kanban view displays tasks in different statuses with drag-and-drop support.

### Task Details
Task detail page shows task information, subtask list, comments, and assignments.

### Dependency Graph
Visual representation of task dependencies.

## 🧪 Testing

### Backend Tests
```bash
cd backend
go test ./... -v
```

### Frontend Tests
```bash
cd frontend
pnpm test
```

## 📝 Development Guidelines

### Go Code Standards
- Use `gofmt` for code formatting
- Follow Effective Go guidelines
- Use `any` instead of `interface{}`
- Use `:=` for short variable declarations

### Frontend Code Standards
- Use ESLint + Prettier for code formatting
- TypeScript strict mode
- Use Composition API for components
- Use ESLint for code linting

### Commit Convention
```bash
feat: New feature
fix: Bug fix
docs: Documentation update
style: Code style adjustment
refactor: Code refactoring
test: Test related
chore: Build/toolchain related
```

## 🔧 Configuration

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| SERVER_PORT | Server port | 8080 |
| SERVER_MODE | Run mode (debug/release) | debug |
| DB_HOST | Database host | localhost |
| DB_PORT | Database port | 3306 |
| DB_USER | Database user | root |
| DB_PASSWORD | Database password | - |
| DB_NAME | Database name | ai_task_manager |
| AI_PROVIDER | AI provider (qwen/gemini) | qwen |
| QWEN_API_KEY | Qwen API Key | - |
| QWEN_MODEL | Qwen model | qwen-plus |
| MAX_TOKENS | Maximum tokens | 8192 |
| TEMPERATURE | Model temperature | 0.7 |

## 🤝 Contributing

Issues and Pull Requests are welcome!

1. Fork the project
2. Create a feature branch (`git checkout -b feature/AmazingFeature`)
3. Commit your changes (`git commit -m 'Add some AmazingFeature'`)
4. Push to the branch (`git push origin feature/AmazingFeature`)
5. Open a Pull Request

### Contributor License Agreement (CLA)

By contributing to this project, you agree to the [Contributor License Agreement](CLA.md). Key terms:

- You grant the copyright holder permission to use your contribution in any form (including commercial)
- You represent that your contribution is original and does not infringe third-party rights
- The copyright holder may use your contribution for commercial licensing

📄 View full terms: [CLA.md](CLA.md)

## 📄 License

This project uses a **Dual Licensing Model**:

### AGPL-3.0 Open Source License

For the following use cases, this project is licensed under the **GNU Affero General Public License v3.0 (AGPL-3.0)**:

- ✅ Personal learning and research
- ✅ Non-commercial use
- ✅ Open source projects (complying with AGPL-3.0 terms)
- ✅ Internal use and testing

**AGPL-3.0 Key Requirements:**
- If you modify this project, you must release the modified source code
- If you provide network services, you must provide source code to service users
- You must retain copyright notices and license declarations
- Derivative works must also use AGPL-3.0

### Commercial License

For the following use cases, you need to obtain a **Commercial License**:

- ❌ Using this project for commercial products or services
- ❌ Offering this project as a SaaS service
- ❌ Integrating this project into proprietary commercial products
- ❌ Not wanting to comply with AGPL-3.0 open source obligations

**Commercial License Benefits:**
- No need to release your source code
- No need to comply with AGPL-3.0 terms
- Access to professional technical support
- Access to custom development services

📧 For commercial licensing inquiries, contact: [your-email@example.com]

View full license terms: [LICENSE](LICENSE) | [COMMERCIAL_LICENSE.md](COMMERCIAL_LICENSE.md)

## 👥 Authors

- **skindhu** - [GitHub](https://github.com/skindhu)

## 🙏 Acknowledgments

This project is built on the following open source projects:

- [Gin](https://github.com/gin-gonic/gin)
- [GORM](https://github.com/go-gorm/gorm)
- [Vue.js](https://github.com/vuejs/core)
- [Naive UI](https://github.com/tusen-ai/naive-ui)
- [Soybean Admin](https://github.com/soybeanjs/soybean-admin)
- [MCP Go](https://github.com/mark3labs/mcp-go)

## 📮 Contact

For questions or suggestions, please contact through:

- GitHub Issues: [Submit an issue](https://github.com/skindhu/AI-TASK-MANAGER/issues)
- Email: [Send email](mailto:your-email@example.com)

---

If this project helps you, please give it a ⭐️ Star!
