// Package model 商户图片表
// 封面/环境/菜品/营业执照/其他图片
package model

import (
	"wuchang-tongcheng/internal/pkg/database"
)

// Dh114Image 商户图片表
type Dh114Image struct {
	database.BaseModel // id/created_at/updated_at/deleted_at
	Dh114ID    uint   `gorm:"not null;index" json:"dh114_id"`                                       // 商户 ID
	BusinessID uint   `gorm:"not null;default:0;index" json:"business_id"`                          // 商户详情 ID
	ImageType  string `gorm:"size:32;not null;default:'other';index" json:"image_type"`             // cover/environment/dish/license/other
	URL        string `gorm:"size:500;not null" json:"url"`                                          // 图片 URL
	Thumbnail  string `gorm:"size:500;not null;default:''" json:"thumbnail"`                       // 缩略图 URL
	Title      string `gorm:"size:128;not null;default:''" json:"title"`                            // 图片标题
	Description string `gorm:"size:500;not null;default:''" json:"description"`                     // 图片描述
	Sort       int    `gorm:"not null;default:0;index" json:"sort"`                                // 排序
	IsCover    bool   `gorm:"not null;default:false;index" json:"is_cover"`                         // 是否封面图
	Width      int    `gorm:"not null;default:0" json:"width"`                                      // 图片宽度
	Height     int    `gorm:"not null;default:0" json:"height"`                                     // 图片高度
	Size       int64  `gorm:"not null;default:0" json:"size"`                                       // 文件大小（字节）
	Tag        string `gorm:"size:64;not null;default:''" json:"tag"`                              // 标签
}

// TableName 表名
func (Dh114Image) TableName() string { return "dh114_images" }
