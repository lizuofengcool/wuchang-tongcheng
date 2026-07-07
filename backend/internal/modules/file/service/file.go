// Package service 文件业务逻辑层
package service

import (
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"time"

	"wuchang-tongcheng/internal/modules/file/dto"
	"wuchang-tongcheng/internal/modules/file/model"
	"wuchang-tongcheng/internal/modules/file/repository"
	"wuchang-tongcheng/internal/pkg/storage"
	"wuchang-tongcheng/internal/pkg/utils"

	"gorm.io/gorm"
)

var (
	ErrFileEmpty       = errors.New("文件为空")
	ErrFileTypeInvalid = errors.New("不支持的文件类型")
	ErrFileTooLarge    = errors.New("文件过大")
	ErrFileNotFound    = errors.New("文件不存在")
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

// 预签名直传 URL 有效期（S3 协议上限 7 天，这里取 15 分钟兼顾安全与可用性）
const presignExpiry = 15 * time.Minute

// FileService 文件业务逻辑接口
type FileService interface {
	Upload(regionID uint, userID uint, filename string, mimeType string, size int64, reader io.Reader) (*model.FileUpload, error)
	List(regionID uint, req *dto.ListFilesRequest) (*utils.Pagination, []model.FileUpload, error)
	Delete(id uint) error
	// PresignUpload 生成预签名 PUT 上传 URL（前端直传对象存储），仅 S3/MinIO 等支持。
	PresignUpload(regionID uint, userID uint, filename string) (*dto.PresignUploadResponse, error)
	// CommitUpload 前端直传完成后提交文件记录（按 object_name 由后端重新拼装访问 URL）。
	CommitUpload(regionID uint, userID uint, filename, objectName, mimeType string, size int64) (*model.FileUpload, error)
}

type fileService struct {
	repo repository.FileRepository
}

// NewFileService 创建文件服务
func NewFileService(repo repository.FileRepository) FileService {
	return &fileService{repo: repo}
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

	if err := s.repo.Create(record); err != nil {
		// 数据库写入失败，尝试回滚已保存的文件
		_ = storage.Delete(url)
		return nil, fmt.Errorf("记录文件信息失败: %w", err)
	}

	return record, nil
}

// List 文件列表（按地区隔离）
func (s *fileService) List(regionID uint, req *dto.ListFilesRequest) (*utils.Pagination, []model.FileUpload, error) {
	pagination := utils.NewPagination(req.Page, req.PageSize)
	list, total, err := s.repo.List(regionID, pagination, req.FileType, req.Keyword)
	if err != nil {
		return nil, nil, err
	}
	pagination.Total = total
	return pagination, list, nil
}

// Delete 删除文件（同时删除存储文件）
func (s *fileService) Delete(id uint) error {
	record, err := s.repo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrFileNotFound
		}
		return err
	}
	// 删除存储文件（失败不阻塞记录删除）
	_ = storage.GetStorage().Delete(record.FileURL)
	return s.repo.Delete(id)
}

// PresignUpload 生成预签名 PUT 上传 URL，供前端直传对象存储。
//
// 流程：校验文件类型 → 调用存储层 PresignPut（MinIO 本地 SigV4 签名，无网络请求）→ 返回上传URL/对象名/访问URL。
// 本地存储/七牛云等不支持预签名时返回 storage.ErrPresignNotSupported，handler 据此回 501。
func (s *fileService) PresignUpload(regionID uint, userID uint, filename string) (*dto.PresignUploadResponse, error) {
	// regionID/userID 当前仅用于鉴权留痕，预签名 URL 本身不绑定用户；
	// 真正落库在 CommitUpload 阶段（携带 regionID/userID）。
	_ = regionID
	_ = userID

	ext := strings.ToLower(filepath.Ext(filename))
	fileType, ok := allowedExtensions[ext]
	if !ok {
		return nil, ErrFileTypeInvalid
	}

	st := storage.GetStorage()
	uploadURL, objectName, accessURL, err := st.PresignPut(filename, presignExpiry)
	if err != nil {
		// 不支持预签名的存储后端：原样透出 sentinel，由 handler 映射为 501
		if errors.Is(err, storage.ErrPresignNotSupported) {
			return nil, storage.ErrPresignNotSupported
		}
		return nil, fmt.Errorf("生成预签名 URL 失败: %w", err)
	}

	return &dto.PresignUploadResponse{
		UploadURL:  uploadURL,
		AccessURL:  accessURL,
		ObjectName: objectName,
		ExpiresIn:  int(presignExpiry.Seconds()),
		FileName:   filename,
		FileType:   fileType,
	}, nil
}

// CommitUpload 前端直传完成后提交文件记录。
//
// 后端按 object_name 重新拼装访问 URL（避免前端伪造 URL），校验类型/大小后写入 file_uploads 表。
// 注意：此处不校验对象是否真实存在于桶中（跨服务 HEAD 请求成本高，且直传成功后理应存在），
// 调用方需保证先完成 PUT 上传再调用本接口。
func (s *fileService) CommitUpload(regionID uint, userID uint, filename, objectName, mimeType string, size int64) (*model.FileUpload, error) {
	if size <= 0 {
		return nil, ErrFileEmpty
	}
	if size > maxFileSize {
		return nil, ErrFileTooLarge
	}

	ext := strings.ToLower(filepath.Ext(filename))
	fileType, ok := allowedExtensions[ext]
	if !ok {
		return nil, ErrFileTypeInvalid
	}
	if mimeType == "" {
		mimeType = guessMIMEByExt(ext)
	}

	st := storage.GetStorage()
	accessURL, err := st.AccessURL(objectName)
	if err != nil {
		if errors.Is(err, storage.ErrPresignNotSupported) {
			return nil, storage.ErrPresignNotSupported
		}
		return nil, fmt.Errorf("构造访问 URL 失败: %w", err)
	}

	record := &model.FileUpload{
		UserID:   userID,
		FileName: filename,
		FileURL:  accessURL,
		FileSize: size,
		FileType: fileType,
		MimeType: mimeType,
	}
	record.RegionID = regionID

	if err := s.repo.Create(record); err != nil {
		return nil, fmt.Errorf("记录文件信息失败: %w", err)
	}
	return record, nil
}

// guessMIMEByExt 根据扩展名推断 MIME 类型（CommitUpload 兜底用）。
// 与 handler.guessMIME 保持一致，避免引入循环依赖。
func guessMIMEByExt(ext string) string {
	switch ext {
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".png":
		return "image/png"
	case ".gif":
		return "image/gif"
	case ".webp":
		return "image/webp"
	case ".mp4":
		return "video/mp4"
	case ".mov":
		return "video/quicktime"
	case ".mp3":
		return "audio/mpeg"
	case ".wav":
		return "audio/wav"
	case ".pdf":
		return "application/pdf"
	case ".zip":
		return "application/zip"
	case ".doc", ".docx":
		return "application/msword"
	case ".xls", ".xlsx":
		return "application/vnd.ms-excel"
	case ".ppt", ".pptx":
		return "application/vnd.ms-powerpoint"
	case ".txt":
		return "text/plain"
	default:
		return "application/octet-stream"
	}
}
