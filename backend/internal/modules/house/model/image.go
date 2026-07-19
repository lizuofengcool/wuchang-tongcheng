// Package model 房源图片表（对标贝壳）
// 户型图/实景图/客厅/卧室/厨房/卫生间/小区环境
package model

import (
	"wuchang-tongcheng/internal/pkg/database"
)

// === 图片类型常量 ===
const (
	ImageTypeFloorPlan   = "floor_plan"   // 户型图
	ImageTypeReal        = "real"         // 实景图
	ImageTypeLivingRoom  = "living_room"  // 客厅
	ImageTypeBedroom     = "bedroom"      // 卧室
	ImageTypeKitchen     = "kitchen"      // 厨房
	ImageTypeBathroom    = "bathroom"     // 卫生间
	ImageTypeBalcony     = "balcony"      // 阳台
	ImageTypeCommunity   = "community"    // 小区环境
	ImageTypeCertificate = "certificate"  // 证件
)

// === 图片状态常量 ===
const (
	ImageStatusDisabled = 0 // 禁用
	ImageStatusEnabled  = 1 // 启用
)

// HouseImage 房源图片表
type HouseImage struct {
	database.BaseModel // id/created_at/updated_at/deleted_at
	HouseID     uint   `gorm:"not null;index" json:"house_id"`                       // 关联房源 ID
	ListingID   uint   `gorm:"not null;default:0;index" json:"listing_id"`           // 关联发布 ID
	URL         string `gorm:"size:500;not null" json:"url"`                         // 图片 URL
	Thumbnail   string `gorm:"size:500;not null;default:''" json:"thumbnail"`        // 缩略图 URL
	ImageType   string `gorm:"size:32;not null;default:'real';index" json:"image_type"` // floor_plan/real/living_room/bedroom/kitchen/bathroom/balcony/community/certificate
	Title       string `gorm:"size:128;not null;default:''" json:"title"`            // 标题
	Description string `gorm:"size:500;not null;default:''" json:"description"`     // 描述
	Width       int    `gorm:"not null;default:0" json:"width"`                      // 宽度
	Height      int    `gorm:"not null;default:0" json:"height"`                     // 高度
	Size        int64  `gorm:"not null;default:0" json:"size"`                       // 文件大小（字节）
	Sort        int    `gorm:"not null;default:0;index" json:"sort"`                 // 排序
	IsCover     bool   `gorm:"not null;default:false;index" json:"is_cover"`         // 是否封面
	Status      int    `gorm:"default:1;index" json:"status"`                        // 0禁用 1启用
	UploaderID  uint   `gorm:"not null;default:0" json:"uploader_id"`                // 上传人 ID
}

// TableName 表名（house_ 前缀）
func (HouseImage) TableName() string { return "house_images" }
