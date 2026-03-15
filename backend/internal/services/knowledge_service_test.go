package services

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/ai-task-manager/backend/internal/config"
	"go.uber.org/zap"
)

func setupKnowledgeTest(cfg *config.KnowledgeConfig) *KnowledgeService {
	if cfg == nil {
		cfg = &config.KnowledgeConfig{
			Enabled:   true,
			MaxSize:   500,
			MaxFiles:  50,
			FileTypes: []string{".md", ".txt", ".json"},
		}
	}
	return NewKnowledgeService(cfg, zap.NewNop())
}

func TestKnowledgeService_LoadKnowledge_Empty(t *testing.T) {
	service := setupKnowledgeTest(nil)

	// 无路径时返回空
	content, err := service.LoadKnowledge(nil)
	if err != nil {
		t.Errorf("不期望错误: %v", err)
	}
	if content != "" {
		t.Errorf("期望空内容, 实际: %s", content)
	}
}

func TestKnowledgeService_LoadKnowledge_SingleFile(t *testing.T) {
	// 创建临时文件
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.md")
	content := "# 测试文档\n\n这是一个测试文档。"
	if err := os.WriteFile(tmpFile, []byte(content), 0644); err != nil {
		t.Fatalf("创建临时文件失败: %v", err)
	}

	service := setupKnowledgeTest(nil)

	loaded, err := service.LoadKnowledge([]string{tmpFile})
	if err != nil {
		t.Errorf("加载失败: %v", err)
	}

	if loaded == "" {
		t.Error("期望加载内容")
	}

	if !contains(loaded, "测试文档") {
		t.Error("应包含文档内容")
	}
}

func TestKnowledgeService_LoadKnowledge_Directory(t *testing.T) {
	// 创建临时目录和文件
	tmpDir := t.TempDir()

	files := map[string]string{
		"doc1.md":  "# 文档1\n内容1",
		"doc2.txt": "文档2内容",
		"data.json": `{"key": "value"}`,
	}

	for name, content := range files {
		path := filepath.Join(tmpDir, name)
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatalf("创建文件 %s 失败: %v", name, err)
		}
	}

	service := setupKnowledgeTest(nil)

	loaded, err := service.LoadKnowledge([]string{tmpDir})
	if err != nil {
		t.Errorf("加载失败: %v", err)
	}

	// 应该加载所有允许的文件类型
	if !contains(loaded, "文档1") {
		t.Error("应包含 doc1.md 内容")
	}
	if !contains(loaded, "文档2") {
		t.Error("应包含 doc2.txt 内容")
	}
	if !contains(loaded, "value") {
		t.Error("应包含 data.json 内容")
	}
}

func TestKnowledgeService_LoadKnowledge_FileTypeFilter(t *testing.T) {
	// 创建临时目录
	tmpDir := t.TempDir()

	// 创建不同类型的文件
	allowedFile := filepath.Join(tmpDir, "allowed.md")
	ignoredFile := filepath.Join(tmpDir, "ignored.exe")

	os.WriteFile(allowedFile, []byte("允许的内容"), 0644)
	os.WriteFile(ignoredFile, []byte("忽略的内容"), 0644)

	cfg := &config.KnowledgeConfig{
		Enabled:   true,
		FileTypes: []string{".md"},
		MaxFiles:  50,
		MaxSize:   500,
	}
	service := setupKnowledgeTest(cfg)

	loaded, err := service.LoadKnowledge([]string{tmpDir})
	if err != nil {
		t.Errorf("加载失败: %v", err)
	}

	if !contains(loaded, "允许的内容") {
		t.Error("应包含 .md 文件内容")
	}
	if contains(loaded, "忽略的内容") {
		t.Error("不应包含 .exe 文件内容")
	}
}

func TestKnowledgeService_LoadKnowledge_MaxSize(t *testing.T) {
	tmpDir := t.TempDir()

	// 创建大文件
	largeFile := filepath.Join(tmpDir, "large.md")
	largeContent := make([]byte, 600*1024) // 600KB
	for i := range largeContent {
		largeContent[i] = 'a'
	}
	os.WriteFile(largeFile, largeContent, 0644)

	cfg := &config.KnowledgeConfig{
		Enabled:   true,
		FileTypes: []string{".md"},
		MaxFiles:  50,
		MaxSize:   500, // 500KB 限制
	}
	service := setupKnowledgeTest(cfg)

	_, err := service.LoadKnowledge([]string{tmpDir})
	// 大文件应该被跳过，不应返回错误
	if err != nil {
		t.Errorf("不应返回错误: %v", err)
	}
}

func TestKnowledgeService_LoadKnowledge_MaxFiles(t *testing.T) {
	tmpDir := t.TempDir()

	// 创建多个文件
	for i := 0; i < 10; i++ {
		file := filepath.Join(tmpDir, filepath.Join("file", string(rune('0'+i))+".md"))
		os.WriteFile(file, []byte("内容"), 0644)
	}

	cfg := &config.KnowledgeConfig{
		Enabled:   true,
		FileTypes: []string{".md"},
		MaxFiles:  5, // 最多 5 个文件
		MaxSize:   500,
	}
	service := setupKnowledgeTest(cfg)

	// 由于文件数量限制，只应加载部分文件
	// 这里主要测试不会崩溃
	_, err := service.LoadKnowledge([]string{tmpDir})
	if err != nil {
		t.Errorf("不应返回错误: %v", err)
	}
}

func TestKnowledgeService_IsAllowedType(t *testing.T) {
	tests := []struct {
		name      string
		fileTypes []string
		ext       string
		expected  bool
	}{
		{"允许的 .md", []string{".md", ".txt"}, ".md", true},
		{"允许的 .txt", []string{".md", ".txt"}, ".txt", true},
		{"不允许的 .exe", []string{".md", ".txt"}, ".exe", false},
		{"大写扩展名", []string{".md"}, ".MD", false}, // 扩展名检查是大小写敏感的
		{"空配置使用默认", []string{}, ".md", true},
		{"空配置不匹配", []string{}, ".exe", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.KnowledgeConfig{
				FileTypes: tt.fileTypes,
			}
			service := setupKnowledgeTest(cfg)

			result := service.isAllowedType(tt.ext)
			if result != tt.expected {
				t.Errorf("期望 %v, 实际 %v", tt.expected, result)
			}
		})
	}
}

func TestKnowledgeService_BuildKnowledgeContext(t *testing.T) {
	service := setupKnowledgeTest(nil)

	// 创建临时文件
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.md")
	os.WriteFile(tmpFile, []byte("知识库内容"), 0644)

	tests := []struct {
		name              string
		paths             []string
		additionalContext string
		expectKnowledge   bool
		expectAdditional  bool
	}{
		{
			name:              "只有知识库",
			paths:             []string{tmpFile},
			additionalContext: "",
			expectKnowledge:   true,
			expectAdditional:  false,
		},
		{
			name:              "只有补充上下文",
			paths:             nil,
			additionalContext: "补充信息",
			expectKnowledge:   false,
			expectAdditional:  true,
		},
		{
			name:              "两者都有",
			paths:             []string{tmpFile},
			additionalContext: "补充信息",
			expectKnowledge:   true,
			expectAdditional:  true,
		},
		{
			name:              "都没有",
			paths:             nil,
			additionalContext: "",
			expectKnowledge:   false,
			expectAdditional:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			context, err := service.BuildKnowledgeContext(tt.paths, tt.additionalContext)
			if err != nil {
				t.Errorf("不期望错误: %v", err)
				return
			}

			hasKnowledge := contains(context, "知识库")
			hasAdditional := contains(context, "补充")

			if tt.expectKnowledge != hasKnowledge {
				t.Errorf("知识库存在: 期望 %v, 实际 %v", tt.expectKnowledge, hasKnowledge)
			}
			if tt.expectAdditional != hasAdditional {
				t.Errorf("补充上下文存在: 期望 %v, 实际 %v", tt.expectAdditional, hasAdditional)
			}
		})
	}
}

func TestKnowledgeService_GetKnowledgeSummary(t *testing.T) {
	// 创建临时文件
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.md")
	os.WriteFile(tmpFile, []byte("测试内容"), 0644)

	cfg := &config.KnowledgeConfig{
		Enabled:   true,
		Paths:     []string{tmpDir},
		MaxSize:   500,
		MaxFiles:  50,
		FileTypes: []string{".md"},
	}
	service := setupKnowledgeTest(cfg)

	summary, err := service.GetKnowledgeSummary(nil)
	if err != nil {
		t.Errorf("不期望错误: %v", err)
		return
	}

	if !summary["enabled"].(bool) {
		t.Error("应显示启用状态")
	}

	paths, ok := summary["paths"].([]string)
	if !ok || len(paths) == 0 {
		t.Error("应包含路径")
	}
}

func TestKnowledgeService_SkipHiddenDirectories(t *testing.T) {
	tmpDir := t.TempDir()

	// 创建普通目录和文件
	normalDir := filepath.Join(tmpDir, "normal")
	os.Mkdir(normalDir, 0755)
	os.WriteFile(filepath.Join(normalDir, "doc.md"), []byte("正常内容"), 0644)

	// 创建隐藏目录（应跳过）
	hiddenDir := filepath.Join(tmpDir, ".hidden")
	os.Mkdir(hiddenDir, 0755)
	os.WriteFile(filepath.Join(hiddenDir, "secret.md"), []byte("隐藏内容"), 0644)

	// 创建 node_modules（应跳过）
	nodeDir := filepath.Join(tmpDir, "node_modules")
	os.Mkdir(nodeDir, 0755)
	os.WriteFile(filepath.Join(nodeDir, "package.md"), []byte("node内容"), 0644)

	service := setupKnowledgeTest(nil)

	loaded, err := service.LoadKnowledge([]string{tmpDir})
	if err != nil {
		t.Errorf("加载失败: %v", err)
	}

	if !contains(loaded, "正常内容") {
		t.Error("应包含正常目录内容")
	}
	if contains(loaded, "隐藏内容") {
		t.Error("不应包含隐藏目录内容")
	}
	if contains(loaded, "node内容") {
		t.Error("不应包含 node_modules 内容")
	}
}

func TestKnowledgeService_NonExistentPath(t *testing.T) {
	service := setupKnowledgeTest(nil)

	// 不存在的路径应被忽略
	loaded, err := service.LoadKnowledge([]string{"/non/existent/path"})
	if err != nil {
		t.Errorf("不应返回错误: %v", err)
	}
	if loaded != "" {
		t.Errorf("不存在路径应返回空: %s", loaded)
	}
}
