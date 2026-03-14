# 生成变更日志

根据 git commits 生成 CHANGELOG。

## 用法

```bash
/changelog [选项]
```

## 选项

| 选项 | 说明 |
|-----|------|
| `--from=<tag>` | 起始版本标签 |
| `--to=<tag>` | 结束版本标签（默认 HEAD） |
| `--output=<file>` | 输出文件路径 |
| `--version=<ver>` | 新版本号 |

---

## 示例

```bash
# 生成最近一个版本的 changelog
/changelog

# 生成指定版本的 changelog
/changelog --from=v1.2.0 --to=v1.3.0 --version=1.3.0
```

---

## 步骤

### 1. 获取提交历史

```bash
# 获取最近一个 tag 以来的 commits
git log $(git describe --tags --abbrev=0)..HEAD --oneline
```

### 2. 分析提交类型

按照 Conventional Commits 分类：

| 类型 | 说明 | 分类 |
|-----|------|------|
| `feat` | 新功能 | Features |
| `fix` | Bug 修复 | Bug Fixes |
| `perf` | 性能优化 | Performance Improvements |
| `refactor` | 重构 | Code Refactoring |
| `docs` | 文档更新 | Documentation |
| `chore` | 构建/工具 | Chores |

### 3. 生成 CHANGELOG

```markdown
# 变更日志

## [1.3.0] - 2024-01-15

### ✨ 新功能
- 添加任务状态筛选功能
- 支持批量导出任务

### 🐛 Bug 修复
- 修复任务列表分页错误

### 🔧 其他
- 更新依赖版本
```

---

## 输出格式

```
== 变更日志 (v1.2.0 → v1.3.0) ==

📊 统计:
  - 新特性：5
  - Bug 修复：8
  - 重构：3
  - 文档：2
  - 总计：22 commits

📝 详细变更:
### ✨ 新功能 (5)
- 添加任务状态筛选功能 (abc1234)
...

完整内容已保存到 CHANGELOG.md
```
