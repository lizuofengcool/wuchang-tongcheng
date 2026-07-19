// Package model 收藏 + 浏览记录表
// CarFavorite 收藏；CarView 浏览
package model

import (
	"time"

	"wuchang-tongcheng/internal/pkg/database"
)

// === 收藏类型常量 ===
const (
	FavoriteTypeCar     = "car"     // 车源
	FavoriteTypeDealer  = "dealer"  // 车商
	FavoriteTypeSearch  = "search"  // 搜索条件
)

// CarFavorite 车源收藏表
type CarFavorite struct {
	database.BaseModel // id/created_at/updated_at/deleted_at
	UserID       uint   `gorm:"not null;index;uniqueIndex:uniq_car_favorites_user_car_type" json:"user_id"`         // 用户 ID
	CarID        uint   `gorm:"not null;index;uniqueIndex:uniq_car_favorites_user_car_type" json:"car_id"`        // 车源 ID
	ListingID    uint   `gorm:"not null;default:0;index" json:"listing_id"`                                          // 发布 ID
	FavoriteType string `gorm:"size:32;not null;default:'car';index;uniqueIndex:uniq_car_favorites_user_car_type" json:"favorite_type"` // car/dealer/search
	Remark       string `gorm:"size:200;not null;default:''" json:"remark"`                                          // 备注
}

// TableName 表名（car_ 前缀）
func (CarFavorite) TableName() string { return "car_favorites" }

// === 设备类型常量 ===
const (
	ViewDevicePC     = "pc"     // PC
	ViewDeviceWAP    = "wap"    // 移动 web
	ViewDeviceAPP    = "app"    // APP
	ViewDeviceMiniAPP = "miniapp" // 小程序
)

// === 来源常量 ===
const (
	ViewSourceSearch   = "search"   // 搜索
	ViewSourceCategory = "category" // 分类
	ViewSourceRecommend = "recommend" // 推荐
	ViewSourceDirect   = "direct"   // 直接访问
	ViewSourceShare    = "share"    // 分享
)

// CarView 浏览记录表
type CarView struct {
	database.BaseModel // id/created_at/updated_at/deleted_at
	RegionID   uint      `gorm:"not null;default:1;index" json:"region_id"`                       // 地区 ID
	UserID     uint      `gorm:"not null;default:0;index" json:"user_id"`                        // 用户 ID（0 表示未登录）
	CarID      uint      `gorm:"not null;index" json:"car_id"`                                   // 车源 ID
	ListingID  uint      `gorm:"not null;default:0;index" json:"listing_id"`                     // 发布 ID
	IP         string    `gorm:"size:64;not null;default:''" json:"ip"`                          // IP 地址
	UserAgent  string    `gorm:"size:500;not null;default:''" json:"user_agent"`                 // User-Agent
	Referer    string    `gorm:"size:500;not null;default:''" json:"referer"`                    // 来源页
	Device     string    `gorm:"size:32;not null;default:'';index" json:"device"`                // pc/wap/app/miniapp
	Source     string    `gorm:"size:32;not null;default:'';index" json:"source"`                // search/category/recommend/direct/share
	Duration   int       `gorm:"not null;default:0" json:"duration"`                             // 停留时长（秒）
	Longitude  float64   `gorm:"type:decimal(10,7);default:0" json:"longitude"`                  // 经度
	Latitude   float64   `gorm:"type:decimal(10,7);default:0" json:"latitude"`                   // 纬度
	CreatedAt  time.Time `gorm:"index" json:"created_at"`                                       // 兼容字段
}

// TableName 表名（car_ 前缀）
func (CarView) TableName() string { return "car_views" }
