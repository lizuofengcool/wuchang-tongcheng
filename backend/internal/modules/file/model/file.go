// Package model 文件数据模型
package model

import "wuchang-tongcheng/internal/pkg/database"

// FileUpload 文件上传记录
type FileUpload struct {
	database.RegionBaseModel
	UserID   uint   `gorm:"index" json:"user_id"`                  // 上传者ID
	FileName string `gorm:"size:255" json:"file_name"`             // 原始文件名
	FileURL  string `gorm:"size:500;not null" json:"file_url"`    // 访问URL
	FileSize int64  `gorm:"default:0" json:"file_size"`             // 文件大小（字节）
	FileType string `gorm:"size:50" json:"file_type"`              // 文件类型（image/video/doc等）
	MimeType string `gorm:"size:100" json:"mime_type"`             // MIME类型
}

// TableName 表名
func (FileUpload) TableName() string {
	return "file_uploads"
}
