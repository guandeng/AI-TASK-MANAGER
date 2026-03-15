package services

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/ai-task-manager/backend/internal/config"
	"go.uber.org/zap"
)

// KnowledgeService 知识库服务
type KnowledgeService struct {
	cfg    *config.KnowledgeConfig
	logger *zap.Logger
}

// NewKnowledgeService 创建知识库服务
func NewKnowledgeService(cfg *config.KnowledgeConfig, logger *zap.Logger) *KnowledgeService {
	return &KnowledgeService{
		cfg:    cfg,
		logger: logger,
	}
}

// LoadKnowledge 加载知识库内容
func (s *KnowledgeService) LoadKnowledge(paths []string) (string, error) {
	if !s.cfg.Enabled && len(paths) == 0 {
		return "", nil
	}

	// 如果传入路径，使用传入的；否则使用配置的
	usePaths := paths
	if len(usePaths) == 0 {
		usePaths = s.cfg.Paths
	}

	if len(usePaths) == 0 {
		return "", nil
	}

	var contents []string
	fileCount := 0

	for _, path := range usePaths {
		content, count, err := s.loadPath(path)
		if err != nil {
			s.logger.Warn("加载知识库路径失败", zap.String("path", path), zap.Error(err))
			continue
		}
		if content != "" {
			contents = append(contents, content)
			fileCount += count
		}

		// 检查文件数量限制
		if fileCount >= s.cfg.MaxFiles {
			break
		}
	}

	if len(contents) == 0 {
		return "", nil
	}

	return strings.Join(contents, "\n\n---\n\n"), nil
}

// loadPath 加载单个路径（文件或目录）
func (s *KnowledgeService) loadPath(path string) (string, int, error) {
	info, err := os.Stat(path)
	if err != nil {
		return "", 0, err
	}

	if info.IsDir() {
		return s.loadDirectory(path)
	}
	return s.loadFile(path)
}

// loadFile 加载单个文件
func (s *KnowledgeService) loadFile(path string) (string, int, error) {
	// 检查文件类型
	ext := strings.ToLower(filepath.Ext(path))
	if !s.isAllowedType(ext) {
		return "", 0, nil
	}

	// 检查文件大小
	info, err := os.Stat(path)
	if err != nil {
		return "", 0, err
	}

	maxSize := int64(s.cfg.MaxSize) * 1024 // KB to Bytes
	if info.Size() > maxSize {
		return "", 0, fmt.Errorf("文件大小超过限制: %s", path)
	}

	// 读取文件内容
	content, err := ioutil.ReadFile(path)
	if err != nil {
		return "", 0, err
	}

	// 添加文件名作为标题
	header := fmt.Sprintf("## 文件: %s\n\n", filepath.Base(path))
	return header + string(content), 1, nil
}

// loadDirectory 加载目录下的所有文件
func (s *KnowledgeService) loadDirectory(dirPath string) (string, int, error) {
	var contents []string
	fileCount := 0

	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// 跳过目录
		if info.IsDir() {
			// 跳过隐藏目录和常见的排除目录
			if strings.HasPrefix(info.Name(), ".") ||
				info.Name() == "node_modules" ||
				info.Name() == "vendor" {
				return filepath.SkipDir
			}
			return nil
		}

		// 检查文件数量限制
		if fileCount >= s.cfg.MaxFiles {
			return filepath.SkipAll
		}

		// 加载文件
		content, count, err := s.loadFile(path)
		if err != nil {
			s.logger.Debug("跳过文件", zap.String("path", path), zap.Error(err))
			return nil
		}
		if content != "" {
			contents = append(contents, content)
			fileCount += count
		}

		return nil
	})

	if err != nil && err != filepath.SkipAll {
		return "", 0, err
	}

	return strings.Join(contents, "\n\n"), fileCount, nil
}

// isAllowedType 检查文件类型是否允许
func (s *KnowledgeService) isAllowedType(ext string) bool {
	if len(s.cfg.FileTypes) == 0 {
		// 默认允许的类型
		defaultTypes := []string{".md", ".txt", ".json", ".yaml", ".yml"}
		for _, t := range defaultTypes {
			if ext == t {
				return true
			}
		}
		return false
	}

	for _, t := range s.cfg.FileTypes {
		if ext == t {
			return true
		}
	}
	return false
}

// BuildKnowledgeContext 构建知识库上下文
func (s *KnowledgeService) BuildKnowledgeContext(paths []string, additionalContext string) (string, error) {
	knowledge, err := s.LoadKnowledge(paths)
	if err != nil {
		return "", err
	}

	var context strings.Builder

	if knowledge != "" {
		context.WriteString("=== 业务知识库 ===\n\n")
		context.WriteString(knowledge)
		context.WriteString("\n\n")
	}

	if additionalContext != "" {
		context.WriteString("=== 补充上下文 ===\n\n")
		context.WriteString(additionalContext)
	}

	return context.String(), nil
}

// GetKnowledgeSummary 获取知识库摘要
func (s *KnowledgeService) GetKnowledgeSummary(paths []string) (map[string]interface{}, error) {
	usePaths := paths
	if len(usePaths) == 0 {
		usePaths = s.cfg.Paths
	}

	summary := map[string]interface{}{
		"enabled":    s.cfg.Enabled,
		"paths":      usePaths,
		"maxSize":    s.cfg.MaxSize,
		"maxFiles":   s.cfg.MaxFiles,
		"fileTypes":  s.cfg.FileTypes,
		"fileCount":  0,
		"totalSize":  0,
		"files":      []map[string]interface{}{},
	}

	fileCount := 0
	totalSize := int64(0)
	var files []map[string]interface{}

	for _, path := range usePaths {
		s.scanPath(path, &fileCount, &totalSize, &files)
		if fileCount >= s.cfg.MaxFiles {
			break
		}
	}

	summary["fileCount"] = fileCount
	summary["totalSize"] = totalSize
	summary["files"] = files

	return summary, nil
}

// scanPath 扫描路径统计信息
func (s *KnowledgeService) scanPath(path string, fileCount *int, totalSize *int64, files *[]map[string]interface{}) {
	info, err := os.Stat(path)
	if err != nil {
		return
	}

	if info.IsDir() {
		filepath.Walk(path, func(p string, i os.FileInfo, err error) error {
			if err != nil {
				return nil
			}
			if i.IsDir() {
				if strings.HasPrefix(i.Name(), ".") || i.Name() == "node_modules" || i.Name() == "vendor" {
					return filepath.SkipDir
				}
				return nil
			}
			if s.isAllowedType(strings.ToLower(filepath.Ext(p))) {
				*fileCount++
				*totalSize += i.Size()
				*files = append(*files, map[string]interface{}{
					"path": p,
					"size": i.Size(),
				})
			}
			return nil
		})
	} else if s.isAllowedType(strings.ToLower(filepath.Ext(path))) {
		*fileCount++
		*totalSize += info.Size()
		*files = append(*files, map[string]interface{}{
			"path": path,
			"size": info.Size(),
		})
	}
}
