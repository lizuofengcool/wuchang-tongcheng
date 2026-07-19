// Package model 素材存储中台精简版数据模型
// 依据 ershou 模块依赖：图片/视频 + 以图搜图
// 依据需求文档 1.10：4 维数据隔离（region_id + user_id）
package model

import (
	"time"

	"wuchang-tongcheng/internal/pkg/database"
)

// === 文件类型常量 ===
const (
	FileTypeImage    = "image"    // 图片
	FileTypeVideo    = "video"    // 视频
	FileTypeDocument = "document" // 文档
)

// === 素材分类常量 ===
const (
	CategoryUser      = "user"      // 用户素材
	CategoryMerchant  = "merchant"  // 商家素材
	CategoryOperation = "operation" // 运营素材
)

// === 存储驱动常量 ===
const (
	StorageLocal = "local" // 本地
	StorageMinio = "minio" // MinIO
	StorageQiniu = "qiniu" // 七牛
)

// === 转码状态常量 ===
const (
	TranscodeStatusPending   = 0 // 待转码
	TranscodeStatusRunning   = 1 // 转码中
	TranscodeStatusCompleted = 2 // 已完成
	TranscodeStatusFailed    = 3 // 失败
)

// File 文件主表
type File struct {
	database.RegionBaseModel

	FileID        string `gorm:"size:64;not null;uniqueIndex" json:"file_id"`           // 文件ID
	UserID        uint   `gorm:"index" json:"user_id"`                                  // 上传者ID
	FileType      string `gorm:"size:16;default:'image';index" json:"file_type"`        // 文件类型
	FileURL       string `gorm:"size:512" json:"file_url"`                              // 文件URL
	FileSize      int64  `gorm:"default:0" json:"file_size"`                            // 文件大小（字节）
	MimeType      string `gorm:"size:64" json:"mime_type"`                              // MIME 类型
	FileHash      string `gorm:"size:64;index" json:"file_hash"`                        // 文件哈希（SHA256）
	OriginalName  string `gorm:"size:256" json:"original_name"`                         // 原始文件名
	Category      string `gorm:"size:32;default:'user';index" json:"category"`          // 素材分类
	StorageDriver string `gorm:"size:32;default:'local'" json:"storage_driver"`         // 存储驱动
	Extra         string `gorm:"type:jsonb;default:'{}'::jsonb" json:"extra"`           // 扩展字段
}

// TableName 表名
func (File) TableName() string { return "mat_files" }

// Image 图片元数据
type Image struct {
	ID             uint      `gorm:"primarykey" json:"id"`
	FileID         string    `gorm:"size:64;not null;uniqueIndex" json:"file_id"`           // 文件ID
	RegionID       uint      `gorm:"index;not null;default:1" json:"region_id"`             // 地区ID
	Width          int       `gorm:"default:0" json:"width"`                                // 宽度
	Height         int       `gorm:"default:0" json:"height"`                               // 高度
	Thumbnails     string    `gorm:"type:jsonb;default:'{}'::jsonb" json:"thumbnails"`     // 缩略图 JSON
	PHash          string    `gorm:"column:phash;size:64;index" json:"phash"`              // 感知哈希
	ColorHistogram string    `gorm:"type:jsonb;default:'[]'::jsonb" json:"color_histogram"` // 颜色直方图
	Watermarked    bool      `gorm:"default:false" json:"watermarked"`                      // 是否已加水印
	Extra          string    `gorm:"type:jsonb;default:'{}'::jsonb" json:"extra"`          // 扩展字段
	CreatedAt      time.Time `gorm:"not null;default:now()" json:"created_at"`
	UpdatedAt      time.Time `gorm:"not null;default:now()" json:"updated_at"`
}

// TableName 表名
func (Image) TableName() string { return "mat_images" }

// Video 视频元数据
type Video struct {
	ID              uint      `gorm:"primarykey" json:"id"`
	FileID          string    `gorm:"size:64;not null;uniqueIndex" json:"file_id"`           // 文件ID
	RegionID        uint      `gorm:"index;not null;default:1" json:"region_id"`             // 地区ID
	Duration        int       `gorm:"default:0" json:"duration"`                             // 时长（秒）
	Resolution      string    `gorm:"size:16" json:"resolution"`                             // 分辨率
	Codec           string    `gorm:"size:16" json:"codec"`                                  // 编码
	CoverURL        string    `gorm:"size:512" json:"cover_url"`                             // 封面 URL
	TranscodeStatus int       `gorm:"default:0;index" json:"transcode_status"`               // 转码状态
	TranscodeJobs   string    `gorm:"type:jsonb;default:'[]'::jsonb" json:"transcode_jobs"`  // 转码任务
	Extra           string    `gorm:"type:jsonb;default:'{}'::jsonb" json:"extra"`           // 扩展字段
	CreatedAt       time.Time `gorm:"not null;default:now()" json:"created_at"`
	UpdatedAt       time.Time `gorm:"not null;default:now()" json:"updated_at"`
}

// TableName 表名
func (Video) TableName() string { return "mat_videos" }

// ImageFeature 图片特征向量
type ImageFeature struct {
	ID             uint      `gorm:"primarykey" json:"id"`
	ImageID        uint      `gorm:"index;not null;uniqueIndex" json:"image_id"`             // 图片记录ID
	FileID         string    `gorm:"size:64" json:"file_id"`                                 // 文件ID
	RegionID       uint      `gorm:"index;not null;default:1" json:"region_id"`              // 地区ID
	PHash          string    `gorm:"column:phash;size:64;index" json:"phash"`                // pHash 哈希
	FeatureVector  string    `gorm:"type:jsonb;default:'[]'::jsonb" json:"feature_vector"`   // 特征向量
	ColorHistogram string    `gorm:"type:jsonb;default:'[]'::jsonb" json:"color_histogram"`  // 颜色直方图
	CreatedAt      time.Time `gorm:"not null;default:now()" json:"created_at"`
	UpdatedAt      time.Time `gorm:"not null;default:now()" json:"updated_at"`
}

// TableName 表名
func (ImageFeature) TableName() string { return "mat_image_features" }
