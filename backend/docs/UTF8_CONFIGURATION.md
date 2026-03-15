# 项目 UTF-8 编码配置说明

本文档说明项目的 UTF-8 编码配置,确保所有开发者在任何环境下都能正确处理中文和其他 Unicode 字符。

## 📋 已配置的项目

### 1. EditorConfig (`.editorconfig`)
- 跨编辑器统一的编码和格式配置
- 所有文本文件强制使用 UTF-8 编码
- 支持 VSCode、WebStorm、Sublime、Vim 等主流编辑器

### 2. Git Attributes (`.gitattributes`)
- 强制 Git 使用 UTF-8 处理文本文件
- 自动转换 CRLF 为 LF(跨平台兼容)
- 防止 Git 损坏中文文件名和内容

### 3. VSCode 配置 (`.vscode/settings.json`)
- 项目级 VSCode 设置
- 强制 UTF-8 编码,禁用自动编码猜测
- 配置格式化工具

### 4. Git 本地配置
```bash
core.quotepath=false      # 正确显示中文文件名
i18n.commitEncoding=utf-8  # Commit 信息使用 UTF-8
i18n.logOutputEncoding=utf-8  # Git log 输出使用 UTF-8
gui.encoding=utf-8        # Git GUI 使用 UTF-8
```

## 🔧 使用方法

### VSCode 用户
1. 安装推荐扩展:
   - EditorConfig for VS Code
   - Prettier
   - ESLint
   - Vue - Official

2. 重新加载窗口 (`Cmd+Shift+P` → "Reload Window")

3. 检查编码: 右下角状态栏应显示 "UTF-8"

### 其他编辑器用户
- **WebStorm/IntelliJ**: 自动读取 `.editorconfig`
- **Vim/Neovim**: 安装 `editorconfig-vim` 插件
- **Sublime Text**: 安装 `EditorConfig` 插件

## ✅ 验证配置

### 检查文件编码
```bash
# 检查单个文件
file -I filename.md

# 检查所有 Markdown 文件
find . -name "*.md" -exec file -I {} \;
```

### 检查 Git 配置
```bash
git config --local --list | grep -E '(quotepath|i18n|gui.encoding)'
```

### 测试中文支持
创建测试文件:
```bash
echo "# 测试中文编码" > test-encoding.md
git add test-encoding.md
git status  # 应正确显示中文文件名
```

## 🚨 常见问题

### Q: 文件已经乱码怎么办?
```bash
# 转换为 UTF-8 (macOS/Linux)
iconv -f GBK -t UTF-8 input.md > output.md

# 或使用 VSCode
# 1. 打开文件
# 2. 点击右下角编码
# 3. 选择 "Reopen with Encoding"
# 4. 选择原始编码(如 GBK)
# 5. 再次点击编码 → "Save with Encoding" → UTF-8
```

### Q: Git 显示中文文件名为 `\xxx\xxx`?
```bash
git config --global core.quotepath false
```

### Q: Commit 信息乱码?
```bash
git config --global i18n.commitEncoding utf-8
git config --global i18n.logOutputEncoding utf-8
```

### Q: VSCode 仍然打开文件乱码?
1. 检查 VSCode 设置: `"files.encoding": "utf8"`
2. 禁用自动猜测: `"files.autoGuessEncoding": false`
3. 手动重新打开: 点击右下角编码 → "Reopen with Encoding" → UTF-8

## 📚 参考资料

- [EditorConfig 官方文档](https://editorconfig.org)
- [Git Attributes 文档](https://git-scm.com/docs/gitattributes)
- [UTF-8 编码说明](https://en.wikipedia.org/wiki/UTF-8)
- [VSCode 编码配置](https://code.visualstudio.com/docs/editor/codebasics#_file-encoding)

---

**最后更新**: 2026-03-14
**维护者**: 项目团队
