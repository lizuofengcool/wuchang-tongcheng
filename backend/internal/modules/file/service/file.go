// Package service 文件业务逻辑层
package service

import (
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"wuchang-tongcheng/internal/modules/file/model"
	"wuchang-tongcheng/internal/pkg/storage"

	"gorm.io/gorm"
)

var (
	ErrFileEmpty       = errors.New("文件为空")
	ErrFileTypeInvalid = errors.New("不支持的文件类型")
	ErrFileTooLarge    = errors.New("文件过大")
)

// 允许的文件扩展名
var allowedExtensions = map[string]string{
	".jpg": "image", ".jpeg": "image", ".png": "image", ".gif": "image", ".webp": "image",
	".mp4": "video", ".mov": "video", ".avi": "video",
	".pdf": "doc", ".doc": "doc", ".docx": "doc", ".xls": "doc", ".xlsx": "doc", ".ppt": "doc", ".pptx": "doc",
	".txt":  "doc",
	".zip":  "archive", ".rar": "archive", ".7z": "archive",
	".mp3":  "audio", ".wav": "audio",
}

// 最大文件大小 50MB
const maxFileSize = 50 * 1024 * 1024

// FileService 文件业务逻辑接口
type FileService interface {
	Upload(regionID uint, userID uint, filename string, mimeType string, size int64, reader io.Reader) (*model.FileUpload, error)
}

type fileService struct {
	db *gorm.DB
}

// NewFileService 创建文件服务
func NewFileService(db *gorm.DB) FileService {
	return &fileService{db: db}
}

// Upload 上传文件
func (s *fileService) Upload(regionID uint, userID uint, filename string, mimeType string, size int64, reader io.Reader) (*model.FileUpload, error) {
	if size <= 0 {
		return nil, ErrFileEmpty
	}
	if size > maxFileSize {
		return nil, ErrFileTooLarge
	}

	// 校验文件类型
	ext := strings.ToLower(filepath.Ext(filename))
	fileType, ok := allowedExtensions[ext]
	if !ok {
		return nil, ErrFileTypeInvalid
	}

	// 调用存储层保存
	storage := storage.GetStorage()
	url, err := storage.Save(filename, reader)
	if err != nil {
		return nil, fmt.Errorf("保存文件失败: %w", err)
	}

	// 写入数据库记录
	record := &model.FileUpload{
		UserID:   userID,
		FileName: filename,
		FileURL:  url,
		FileSize: size,
		FileType: fileType,
		MimeType: mimeType,
	}
	record.RegionID = regionID

	if err := s.db.Create(record).Error; err != nil {
		// 数据库写入失败，尝试回滚已保存的文件
		_ = storage.Delete(url)
		return nil, fmt.Errorf("记录文件信息失败: %w", err)
	}

	return record, nil
}
