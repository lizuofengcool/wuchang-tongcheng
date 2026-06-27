// Package dto 文件模块数据传输对象
package dto

// FileInfo 文件信息
type FileInfo struct {
	ID        uint   `json:"id"`
	UserID    uint   `json:"user_id"`
	FileName  string `json:"file_name"`
	FileURL   string `json:"file_url"`
	FileSize  int64  `json:"file_size"`
	FileType  string `json:"file_type"`
	MimeType  string `json:"mime_type"`
	RegionID  uint   `json:"region_id"`
	CreatedAt string `json:"created_at"`
}

// ListFilesRequest 文件列表查询请求
type ListFilesRequest struct {
	Page     int    `form:"page" json:"page"`
	PageSize int    `form:"page_size" json:"page_size"`
	FileType string `form:"file_type" json:"file_type"` // image/video/doc/archive/audio
	Keyword  string `form:"keyword" json:"keyword"`     // 文件名关键词
}
