// Package storage 对象存储抽象层
// 支持 local 本地存储，预留 minio/qiniu 实现
package storage

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"wuchang-tongcheng/internal/pkg/config"
)

// Storage 存储接口
type Storage interface {
	// Save 保存文件，返回可访问的URL
	Save(filename string, reader io.Reader) (url string, err error)
	// Delete 删除文件
	Delete(url string) error
}

// 全局存储实例
var storage Storage

// Init 初始化存储
func Init(cfg *config.StorageConfig) error {
	switch cfg.Type {
	case "local", "":
		s, err := NewLocalStorage(cfg)
		if err != nil {
			return err
		}
		storage = s
	case "minio":
		// 预留MinIO实现（生产环境对接）
		return errors.New("minio storage not implemented yet, use local")
	case "qiniu":
		return errors.New("qiniu storage not implemented yet, use local")
	default:
		return fmt.Errorf("unsupported storage type: %s", cfg.Type)
	}
	return nil
}

// GetStorage 获取存储实例
func GetStorage() Storage {
	if storage == nil {
		// 兜底：使用本地存储默认配置
		storage, _ = NewLocalStorage(&config.StorageConfig{
			Type:   "local",
			Domain: "http://localhost:8080",
		})
	}
	return storage
}

// LocalStorage 本地磁盘存储
type LocalStorage struct {
	domain      string // 访问域名
	basePath    string // 存储根目录
	urlPrefix   string // URL访问前缀
}

// NewLocalStorage 创建本地存储
func NewLocalStorage(cfg *config.StorageConfig) (*LocalStorage, error) {
	basePath := "./uploads"
	if cfg.Bucket != "" {
		basePath = cfg.Bucket
	}
	// 确保目录存在
	if err := os.MkdirAll(basePath, 0755); err != nil {
		return nil, fmt.Errorf("create storage dir failed: %w", err)
	}

	domain := cfg.Domain
	if domain == "" {
		domain = "http://localhost:8080"
	}
	// 规范化domain，去掉末尾斜杠
	domain = strings.TrimSuffix(domain, "/")

	return &LocalStorage{
		domain:    domain,
		basePath:  basePath,
		urlPrefix: "/uploads", // 通过静态路由访问
	}, nil
}

// Save 保存文件
func (s *LocalStorage) Save(filename string, reader io.Reader) (string, error) {
	// 按日期分目录：uploads/2026/06/xxx.jpg
	dateDir := time.Now().Format("2006/01")
	dir := filepath.Join(s.basePath, dateDir)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("create date dir failed: %w", err)
	}

	// 生成唯一文件名，保留原扩展名
	ext := filepath.Ext(filename)
	newName := fmt.Sprintf("%d%s", time.Now().UnixNano(), ext)
	relPath := filepath.Join(dateDir, newName)
	absPath := filepath.Join(s.basePath, relPath)

	// 写入文件
	dst, err := os.Create(absPath)
	if err != nil {
		return "", fmt.Errorf("create file failed: %w", err)
	}
	defer dst.Close()

	if _, err := io.Copy(dst, reader); err != nil {
		return "", fmt.Errorf("write file failed: %w", err)
	}

	// 返回可访问URL（通过静态路由 /uploads/ 暴露）
	// 路径分隔符统一为 /
	urlPath := s.urlPrefix + "/" + filepath.ToSlash(relPath)
	return urlPath, nil
}

// Delete 删除文件
func (s *LocalStorage) Delete(url string) error {
	// 从URL中提取相对路径
	if !strings.HasPrefix(url, s.urlPrefix) {
		return errors.New("invalid file url")
	}
	relPath := strings.TrimPrefix(url, s.urlPrefix+"/")
	absPath := filepath.Join(s.basePath, relPath)
	if err := os.Remove(absPath); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}
