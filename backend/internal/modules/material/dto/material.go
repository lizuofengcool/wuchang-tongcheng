// Package dto 素材存储中台精简版数据传输对象
package dto

import "time"

// FileInfo 文件信息
type FileInfo struct {
	ID            uint      `json:"id"`
	FileID        string    `json:"file_id"`
	UserID        uint      `json:"user_id"`
	FileType      string    `json:"file_type"`
	FileURL       string    `json:"file_url"`
	FileSize      int64     `json:"file_size"`
	MimeType      string    `json:"mime_type"`
	FileHash      string    `json:"file_hash"`
	OriginalName  string    `json:"original_name"`
	Category      string    `json:"category"`
	StorageDriver string    `json:"storage_driver"`
	Extra         string    `json:"extra"`
	RegionID      uint      `json:"region_id"`
	CreatedAt     time.Time `json:"created_at"`
}

// UploadRequest 上传请求（普通表单字段，文件由 FormFile 提供）
type UploadRequest struct {
	Category string `form:"category" json:"category"` // user/merchant/operation
	FileType string `form:"file_type" json:"file_type"` // image/video/document
}

// UploadResponse 上传响应
type UploadResponse struct {
	FileID        string `json:"file_id"`
	FileURL       string `json:"file_url"`
	OriginalName  string `json:"original_name"`
	FileSize      int64  `json:"file_size"`
	MimeType      string `json:"mime_type"`
	FileHash      string `json:"file_hash"`
	Width         int    `json:"width,omitempty"`
	Height        int    `json:"height,omitempty"`
	Duration      int    `json:"duration,omitempty"`
	Thumbnails    string `json:"thumbnails,omitempty"`
}

// SearchByImageRequest 以图搜图请求
type SearchByImageRequest struct {
	FileID  string `json:"file_id" binding:"required"` // 参考图片的 file_id
	Limit   int    `json:"limit"`                       // 返回数量，默认 10
	RegionID uint  `json:"region_id"`
}

// SimilarImage 相似图片信息
type SimilarImage struct {
	FileID    string  `json:"file_id"`
	FileURL   string  `json:"file_url"`
	Similarity float64 `json:"similarity"` // 相似度 0-1
}

// WatermarkRequest 加水印请求
type WatermarkRequest struct {
	FileID    string `json:"file_id" binding:"required"`
	Text      string `json:"text" binding:"required,max=64"` // 水印文字
	Position  string `json:"position" binding:"omitempty,oneof=top-left top-right bottom-left bottom-right center"`
}

// ThumbnailRequest 生成缩略图请求
type ThumbnailRequest struct {
	FileID string `json:"file_id" binding:"required"`
	Sizes  []string `json:"sizes"` // 如 ["100x100","300x300","800x800"]
}

// FileInfoListRequest 文件列表请求
type FileInfoListRequest struct {
	UserID    uint   `form:"user_id" json:"user_id"`
	FileType  string `form:"file_type" json:"file_type"`
	Category  string `form:"category" json:"category"`
	Page      int    `form:"page" json:"page"`
	PageSize  int    `form:"page_size" json:"page_size"`
}
