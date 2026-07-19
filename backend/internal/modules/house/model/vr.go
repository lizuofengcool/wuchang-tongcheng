// Package model VR 看房记录表（对标贝壳）
// 720°全景/虚拟看房/录制信息/场景 JSON
package model

import (
	"time"

	"wuchang-tongcheng/internal/pkg/database"
)

// === VR 类型常量 ===
const (
	VRTypePanorama = "panorama" // 360°全景
	VRTypeVR       = "vr"       // VR 看房
	VRTypeVideo    = "video"    // 视频看房
	VRType3D       = "3d"       // 3D 模型
)

// === VR 状态常量 ===
const (
	VRStatusDraft     = 0 // 草稿
	VRStatusPublished = 1 // 已发布
	VRStatusOffline   = 2 // 已下架
	VRStatusRejected  = 3 // 已拒绝
)

// HouseVRTour VR 看房记录表
type HouseVRTour struct {
	database.BaseModel // id/created_at/updated_at/deleted_at
	HouseID         uint       `gorm:"not null;index" json:"house_id"`                       // 关联房源 ID
	ListingID       uint       `gorm:"not null;default:0;index" json:"listing_id"`           // 关联发布 ID
	CommunityID     uint       `gorm:"not null;default:0;index" json:"community_id"`         // 关联小区 ID
	VRNo            string     `gorm:"size:64;not null;uniqueIndex" json:"vr_no"`            // VR 单号
	Title           string     `gorm:"size:200;not null;default:''" json:"title"`            // 标题
	Description     string     `gorm:"type:text" json:"description"`                          // 描述
	VRType          string     `gorm:"size:32;not null;default:'panorama';index" json:"vr_type"` // panorama/vr/video/3d
	VRURL           string     `gorm:"size:500;not null" json:"vr_url"`                       // VR URL
	CoverImage      string     `gorm:"size:500;not null;default:''" json:"cover_image"`      // 封面图
	Scenes          JSONB      `gorm:"type:jsonb" json:"scenes"`                              // 场景列表 JSON
	DurationSeconds int        `gorm:"not null;default:0" json:"duration_seconds"`            // 时长（秒）
	ViewCount       int        `gorm:"not null;default:0" json:"view_count"`                  // 浏览数
	ShareCount      int        `gorm:"not null;default:0" json:"share_count"`                 // 分享数
	Status          int        `gorm:"default:0;index" json:"status"`                         // 0草稿 1已发布 2下架 3拒绝
	RecorderID      uint       `gorm:"not null;default:0;index" json:"recorder_id"`          // 录制人 ID
	RecorderName    string     `gorm:"size:50;not null;default:''" json:"recorder_name"`     // 录制人姓名
	RecordedAt      *time.Time `gorm:"index" json:"recorded_at"`                              // 录制时间
	PublishedAt     *time.Time `gorm:"index" json:"published_at"`                             // 发布时间
	OfflineAt       *time.Time `gorm:"index" json:"offline_at"`                               // 下线时间
	Equipment       string     `gorm:"size:64;not null;default:''" json:"equipment"`         // 录制设备
	Resolution      string     `gorm:"size:32;not null;default:''" json:"resolution"`        // 分辨率
	FileSize        int64      `gorm:"not null;default:0" json:"file_size"`                   // 文件大小（字节）
}

// TableName 表名（house_ 前缀）
func (HouseVRTour) TableName() string { return "house_vr_tours" }
