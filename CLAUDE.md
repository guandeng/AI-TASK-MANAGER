# AI Task Manager - 项目概览

## 项目结构
```
AI-TASK-MANAGER/
├── backend/          # Go 后端 (Gin + GORM)
│   ├── CLAUDE.md     # 后端开发规范
│   └── ...
└── frontend/         # Vue 3 前端 (TypeScript + Element Plus)
    ├── CLAUDE.md     # 前端开发规范
    └── ...
```

## 开发规范
- 后端规范详见 `backend/CLAUDE.md`
- 前端规范详见 `frontend/CLAUDE.md`

## AI 助手行为准则
- 直接修改代码，无需详细解释
- 输出简洁，避免冗余
- 完成后简短回复"完成"或"已修改"
- 不要输出代码片段预览，直接编辑文件

## 技术栈
- **后端**: Go 1.22+ / Gin / GORM / MySQL
- **前端**: Vue 3 / TypeScript / Element Plus / Pinia
