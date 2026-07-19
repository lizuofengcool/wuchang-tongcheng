// Package service 素材存储中台精简版业务逻辑层
// 依据 ershou 模块依赖：图片/视频 + 以图搜图
// 暴露 MaterialService 接口供其他模块直接 import 调用（不通过 HTTP）
package service

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"time"

	"wuchang-tongcheng/internal/modules/material/dto"
	"wuchang-tongcheng/internal/modules/material/model"
	"wuchang-tongcheng/internal/modules/material/repository"
	"wuchang-tongcheng/internal/pkg/utils"

	"gorm.io/gorm"
)

var (
	ErrFileNotFound     = errors.New("文件不存在")
	ErrImageNotFound    = errors.New("图片不存在")
	ErrVideoNotFound    = errors.New("视频不存在")
	ErrFeatureNotFound  = errors.New("图片特征不存在")
	ErrUploadFailed     = errors.New("文件上传失败")
	ErrUnsupportedType  = errors.New("不支持的文件类型")
)

// MaterialService 素材中台业务接口
// 暴露给其他模块直接 import 调用，不通过 HTTP
type MaterialService interface {
	// 文件上传
	Upload(regionID uint, userID uint, req *dto.UploadRequest, filename string, size int64, mimeType string, reader io.Reader) (*dto.UploadResponse, error)

	// 文件查询
	GetFile(fileID string) (*dto.FileInfo, error)
	ListFiles(req *dto.FileInfoListRequest) ([]dto.FileInfo, int64, error)
	DeleteFile(fileID string) error

	// 图片处理
	GenerateThumbnail(req *dto.ThumbnailRequest) (string, error) // 返回缩略图 JSON
	AddWatermark(req *dto.WatermarkRequest) error

	// 以图搜图
	SearchByImage(req *dto.SearchByImageRequest) ([]dto.SimilarImage, error)

	// 视频转码（精简版：仅标记状态）
	UpdateTranscodeStatus(fileID string, status int, jobs string) error
}

type materialService struct {
	repo          repository.MaterialRepository
	storageDriver string // 存储驱动
	storageBaseURL string // 存储 baseURL
}

// NewMaterialService 创建 service 实例
func NewMaterialService(repo repository.MaterialRepository) MaterialService {
	return &materialService{
		repo:           repo,
		storageDriver:  model.StorageLocal,
		storageBaseURL: "/uploads",
	}
}

// Upload 文件上传
// 精简版：仅计算哈希、生成 FileID 和 URL，实际存储由 file 模块或 storage 包处理
// 这里返回 fileURL 形如 /uploads/{fileID}.{ext}
func (s *materialService) Upload(regionID uint, userID uint, req *dto.UploadRequest, filename string, size int64, mimeType string, reader io.Reader) (*dto.UploadResponse, error) {
	// 计算文件哈希
	hasher := sha256.New()
	if _, err := io.Copy(hasher, reader); err != nil {
		return nil, ErrUploadFailed
	}
	hash := hex.EncodeToString(hasher.Sum(nil))

	// 文件类型推断
	fileType := req.FileType
	if fileType == "" {
		fileType = inferFileType(mimeType)
	}
	if fileType == "" {
		return nil, ErrUnsupportedType
	}

	// 幂等：相同哈希直接返回已有记录
	if existing, err := s.repo.FindFileByHash(hash); err == nil {
		return s.buildUploadResponse(existing), nil
	}

	// 生成 FileID 和 URL
	fileID := generateFileID()
	ext := utils.GetFileExt(filename)
	fileURL := fmt.Sprintf("%s/%s%s", s.storageBaseURL, fileID, ext)

	category := req.Category
	if category == "" {
		category = model.CategoryUser
	}

	f := &model.File{
		FileID:        fileID,
		UserID:        userID,
		FileType:      fileType,
		FileURL:       fileURL,
		FileSize:      size,
		MimeType:      mimeType,
		FileHash:      hash,
		OriginalName:  filename,
		Category:      category,
		StorageDriver: s.storageDriver,
		Extra:         "{}",
	}
	f.RegionID = regionID

	if err := s.repo.CreateFile(f); err != nil {
		return nil, err
	}

	// 图片：创建图片元数据
	if fileType == model.FileTypeImage {
		img := &model.Image{
			FileID:         fileID,
			RegionID:       regionID,
			Thumbnails:     "{}",
			PHash:          computePHash(hash), // 精简版：使用 hash 前 16 字符作为 pHash
			ColorHistogram: "[]",
		}
		if err := s.repo.CreateImage(img); err != nil {
			return nil, err
		}

		// 创建特征向量
		feature := &model.ImageFeature{
			ImageID:        img.ID,
			FileID:         fileID,
			RegionID:       regionID,
			PHash:          img.PHash,
			FeatureVector:  "[]",
			ColorHistogram: "[]",
		}
		_ = s.repo.CreateImageFeature(feature)
	}

	// 视频：创建视频元数据
	if fileType == model.FileTypeVideo {
		v := &model.Video{
			FileID:          fileID,
			RegionID:        regionID,
			TranscodeStatus: model.TranscodeStatusPending,
			TranscodeJobs:   "[]",
			Extra:           "{}",
		}
		if err := s.repo.CreateVideo(v); err != nil {
			return nil, err
		}
	}

	return &dto.UploadResponse{
		FileID:       fileID,
		FileURL:      fileURL,
		OriginalName: filename,
		FileSize:     size,
		MimeType:     mimeType,
		FileHash:     hash,
	}, nil
}

// GetFile 查询文件
func (s *materialService) GetFile(fileID string) (*dto.FileInfo, error) {
	f, err := s.repo.FindFileByID(fileID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrFileNotFound
		}
		return nil, err
	}
	return toFileInfo(f), nil
}

// ListFiles 文件列表
func (s *materialService) ListFiles(req *dto.FileInfoListRequest) ([]dto.FileInfo, int64, error) {
	page := req.Page
	if page <= 0 {
		page = 1
	}
	pageSize := req.PageSize
	if pageSize <= 0 {
		pageSize = 20
	}
	list, total, err := s.repo.ListFiles(req.UserID, req.FileType, req.Category, page, pageSize)
	if err != nil {
		return nil, 0, err
	}
	result := make([]dto.FileInfo, 0, len(list))
	for i := range list {
		result = append(result, *toFileInfo(&list[i]))
	}
	return result, total, nil
}

// DeleteFile 删除文件
func (s *materialService) DeleteFile(fileID string) error {
	f, err := s.repo.FindFileByID(fileID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrFileNotFound
		}
		return err
	}
	return s.repo.DeleteFile(f.ID)
}

// GenerateThumbnail 生成缩略图（精简版：仅记录 URLs 到 thumbnails JSON）
// 实际生成应由图片处理库处理，此处仅生成 URL 模板
func (s *materialService) GenerateThumbnail(req *dto.ThumbnailRequest) (string, error) {
	f, err := s.repo.FindFileByID(req.FileID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", ErrFileNotFound
		}
		return "", err
	}

	sizes := req.Sizes
	if len(sizes) == 0 {
		sizes = []string{"100x100", "300x300", "800x800"}
	}

	// 精简版：构造缩略图 URL（实际应由图片处理服务生成）
	thumbnailMap := map[string]string{}
	for _, size := range sizes {
		ext := urlExt(f.FileURL)
		thumbnailMap[size] = fmt.Sprintf("%s_%s%s", f.FileURL[:len(f.FileURL)-len(ext)], size, ext)
	}

	// 序列化为 JSON
	thumbnailsJSON := mapToJSON(thumbnailMap)

	// 更新图片表
	img, err := s.repo.FindImageByFileID(req.FileID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", ErrImageNotFound
		}
		return "", err
	}
	if err := s.repo.UpdateImageFields(img.ID, map[string]interface{}{
		"thumbnails": thumbnailsJSON,
	}); err != nil {
		return "", err
	}
	return thumbnailsJSON, nil
}

// AddWatermark 添加水印（精简版：仅标记 watermarked=true）
func (s *materialService) AddWatermark(req *dto.WatermarkRequest) error {
	img, err := s.repo.FindImageByFileID(req.FileID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrImageNotFound
		}
		return err
	}
	return s.repo.UpdateImageFields(img.ID, map[string]interface{}{
		"watermarked": true,
	})
}

// SearchByImage 以图搜图
// 精简版：基于 pHash 前缀匹配
func (s *materialService) SearchByImage(req *dto.SearchByImageRequest) ([]dto.SimilarImage, error) {
	limit := req.Limit
	if limit <= 0 {
		limit = 10
	}

	// 查询参考图片的 pHash
	img, err := s.repo.FindImageByFileID(req.FileID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrImageNotFound
		}
		return nil, err
	}
	if img.PHash == "" {
		return []dto.SimilarImage{}, nil
	}

	// 查找相似图片
	features, err := s.repo.FindSimilarByPHash(img.PHash, limit)
	if err != nil {
		return nil, err
	}

	result := make([]dto.SimilarImage, 0, len(features))
	for _, f := range features {
		if f.FileID == req.FileID {
			continue // 排除自身
		}
		file, err := s.repo.FindFileByID(f.FileID)
		if err != nil {
			continue
		}
		// 精简版相似度计算：基于 pHash 前缀匹配长度
		similarity := computeSimilarity(img.PHash, f.PHash)
		result = append(result, dto.SimilarImage{
			FileID:     f.FileID,
			FileURL:    file.FileURL,
			Similarity: similarity,
		})
	}
	return result, nil
}

// UpdateTranscodeStatus 更新转码状态（视频）
func (s *materialService) UpdateTranscodeStatus(fileID string, status int, jobs string) error {
	v, err := s.repo.FindVideoByFileID(fileID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrVideoNotFound
		}
		return err
	}
	return s.repo.UpdateVideoFields(v.ID, map[string]interface{}{
		"transcode_status": status,
		"transcode_jobs":   jobs,
	})
}

// ===== 工具函数 =====

// inferFileType 根据 MIME 推断文件类型
func inferFileType(mimeType string) string {
	switch {
	case startsWith(mimeType, "image/"):
		return model.FileTypeImage
	case startsWith(mimeType, "video/"):
		return model.FileTypeVideo
	case startsWith(mimeType, "application/") || startsWith(mimeType, "text/"):
		return model.FileTypeDocument
	}
	return ""
}

func startsWith(s, prefix string) bool {
	return len(s) >= len(prefix) && s[:len(prefix)] == prefix
}

// generateFileID 生成文件 ID
func generateFileID() string {
	return fmt.Sprintf("F%s%s", time.Now().Format("20060102150405"), utils.RandomNumber(6))
}

// computePHash 精简版 pHash：使用文件哈希前 16 字符
func computePHash(hash string) string {
	if len(hash) >= 16 {
		return hash[:16]
	}
	return hash
}

// computeSimilarity 精简版相似度计算：基于 pHash 前缀匹配长度
func computeSimilarity(a, b string) float64 {
	if a == b {
		return 1.0
	}
	minLen := len(a)
	if len(b) < minLen {
		minLen = len(b)
	}
	if minLen == 0 {
		return 0
	}
	match := 0
	for i := 0; i < minLen; i++ {
		if a[i] == b[i] {
			match++
		}
	}
	return float64(match) / float64(minLen)
}

// urlExt 返回 URL 的扩展名（含点）
func urlExt(url string) string {
	for i := len(url) - 1; i >= 0; i-- {
		if url[i] == '/' {
			break
		}
		if url[i] == '.' {
			return url[i:]
		}
	}
	return ""
}

// mapToJSON 简单 map 转 JSON 字符串
func mapToJSON(m map[string]string) string {
	if len(m) == 0 {
		return "{}"
	}
	result := "{"
	first := true
	for k, v := range m {
		if !first {
			result += ","
		}
		result += fmt.Sprintf(`"%s":"%s"`, k, v)
		first = false
	}
	result += "}"
	return result
}

// buildUploadResponse 从 File model 构造上传响应
func (s *materialService) buildUploadResponse(f *model.File) *dto.UploadResponse {
	resp := &dto.UploadResponse{
		FileID:       f.FileID,
		FileURL:      f.FileURL,
		OriginalName: f.OriginalName,
		FileSize:     f.FileSize,
		MimeType:     f.MimeType,
		FileHash:     f.FileHash,
	}
	// 补充图片元数据
	if f.FileType == model.FileTypeImage {
		if img, err := s.repo.FindImageByFileID(f.FileID); err == nil {
			resp.Thumbnails = img.Thumbnails
		}
	}
	return resp
}

// toFileInfo model → dto
func toFileInfo(f *model.File) *dto.FileInfo {
	return &dto.FileInfo{
		ID:            f.ID,
		FileID:        f.FileID,
		UserID:        f.UserID,
		FileType:      f.FileType,
		FileURL:       f.FileURL,
		FileSize:      f.FileSize,
		MimeType:      f.MimeType,
		FileHash:      f.FileHash,
		OriginalName:  f.OriginalName,
		Category:      f.Category,
		StorageDriver: f.StorageDriver,
		Extra:         f.Extra,
		RegionID:      f.RegionID,
		CreatedAt:     f.CreatedAt,
	}
}
