// Package model 数据统计 + 车源图片表
// CarStatistic 统计；CarImage 图片
package model

import (
	"time"

	"wuchang-tongcheng/internal/pkg/database"
)

// === 统计类型常量 ===
const (
	StatTypeCar        = "car"         // 单车统计
	StatTypeDealer     = "dealer"      // 车商统计
	StatTypeCategory   = "category"    // 分类统计
	StatTypeBrand      = "brand"       // 品牌统计
	StatTypeRegion     = "region"      // 地区统计
	StatTypePlatform   = "platform"    // 平台统计
)

// CarStatistic 数据统计表
type CarStatistic struct {
	database.RegionBaseModel // id/region_id/created_at/updated_at/deleted_at
	StatDate         time.Time `gorm:"type:date;not null;index;uniqueIndex:uniq_car_stats_date_type_target,priority:1" json:"stat_date"` // 统计日期
	StatType         string    `gorm:"size:32;not null;index;uniqueIndex:uniq_car_stats_date_type_target,priority:2" json:"stat_type"`     // car/dealer/category/brand/region/platform
	TargetID         uint      `gorm:"not null;default:0;index;uniqueIndex:uniq_car_stats_date_type_target,priority:3" json:"target_id"`     // 目标 ID
	TargetName       string    `gorm:"size:128;not null;default:''" json:"target_name"`                                                       // 目标名称
	ImpressionCount  int       `gorm:"not null;default:0" json:"impression_count"`                                                            // 曝光数
	ClickCount       int       `gorm:"not null;default:0" json:"click_count"`                                                                 // 点击数
	FavCount         int       `gorm:"not null;default:0" json:"fav_count"`                                                                   // 收藏数
	ContactCount     int       `gorm:"not null;default:0" json:"contact_count"`                                                               // 联系数
	TestDriveCount   int       `gorm:"not null;default:0" json:"test_drive_count"`                                                            // 试驾数
	DealCount        int       `gorm:"not null;default:0" json:"deal_count"`                                                                  // 成交数
	ConversionRate   float64   `gorm:"type:decimal(5,2);default:0" json:"conversion_rate"`                                                    // 转化率
	AvgPrice         float64   `gorm:"type:decimal(10,2);default:0" json:"avg_price"`                                                         // 平均价
	AvgDealDays      int       `gorm:"not null;default:0" json:"avg_deal_days"`                                                               // 平均成交周期
}

// TableName 表名（car_ 前缀）
func (CarStatistic) TableName() string { return "car_statistics" }

// === 图片类型常量 ===
const (
	ImageTypeExterior      = "exterior"      // 外观
	ImageTypeInterior      = "interior"      // 内饰
	ImageTypeEngine        = "engine"        // 发动机舱
	ImageTypeChassis       = "chassis"       // 底盘
	ImageTypeAccident      = "accident"      // 事故
	ImageTypeDocument      = "document"      // 证件
	ImageTypeDashboard     = "dashboard"     // 仪表盘
	ImageTypeWheel         = "wheel"         // 轮毂
	ImageTypeTrunk         = "trunk"         // 后备箱
	ImageTypeOther         = "other"         // 其他
)

// CarImage 车源图片表
type CarImage struct {
	database.RegionBaseModel // id/region_id/created_at/updated_at/deleted_at
	CarID       uint   `gorm:"not null;index" json:"car_id"`                                  // 车源 ID
	ListingID   uint   `gorm:"not null;default:0;index" json:"listing_id"`                    // 发布 ID
	ImageType   string `gorm:"size:32;not null;default:'exterior';index" json:"image_type"`   // exterior/interior/engine/chassis/accident/document/dashboard/wheel/trunk/other
	URL         string `gorm:"size:500;not null" json:"url"`                                  // 图片 URL
	Thumbnail   string `gorm:"size:500;not null;default:''" json:"thumbnail"`                 // 缩略图 URL
	Title       string `gorm:"size:128;not null;default:''" json:"title"`                     // 图片标题
	Description string `gorm:"size:500;not null;default:''" json:"description"`               // 图片描述
	Sort        int    `gorm:"not null;default:0;index" json:"sort"`                          // 排序
	IsCover     bool   `gorm:"not null;default:false;index" json:"is_cover"`                  // 是否封面
	Width       int    `gorm:"not null;default:0" json:"width"`                               // 宽度
	Height      int    `gorm:"not null;default:0" json:"height"`                              // 高度
	Size        int    `gorm:"not null;default:0" json:"size"`                                // 文件大小（字节）
	Tag         string `gorm:"size:64;not null;default:''" json:"tag"`                        // 标签
}

// TableName 表名（car_ 前缀）
func (CarImage) TableName() string { return "car_images" }
