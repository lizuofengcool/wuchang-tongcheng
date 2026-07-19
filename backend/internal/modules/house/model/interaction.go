// Package model 收藏 + 浏览记录表
// HouseFavorite 房源/小区/经纪人收藏；HouseView 浏览记录
package model

import (
	"time"

	"wuchang-tongcheng/internal/pkg/database"
)

// === 收藏类型常量 ===
const (
	FavoriteTypeHouse      = "house"      // 收藏房源
	FavoriteTypeCommunity  = "community"  // 收藏小区
	FavoriteTypeAgent      = "agent"      // 收藏经纪人
	FavoriteTypeListing    = "listing"    // 收藏发布
)

// === 浏览来源常量 ===
const (
	ViewSourceList       = "list"        // 列表
	ViewSourceSearch     = "search"      // 搜索
	ViewSourceRecommend  = "recommend"   // 推荐
	ViewSourceShare      = "share"       // 分享
	ViewSourceQRCode     = "qrcode"      // 扫码
	ViewSourcePush       = "push"        // 推送
)

// HouseFavorite 房源收藏表
type HouseFavorite struct {
	database.BaseModel // id/created_at/updated_at/deleted_at
	UserID         uint   `gorm:"not null;index;uniqueIndex:uniq_house_fav_user_type_target" json:"user_id"`         // 用户 ID
	HouseID        uint   `gorm:"not null;default:0;index;uniqueIndex:uniq_house_fav_user_type_target" json:"house_id"`        // 房源 ID
	ListingID      uint   `gorm:"not null;default:0;index;uniqueIndex:uniq_house_fav_user_type_target" json:"listing_id"`      // 发布 ID
	CommunityID    uint   `gorm:"not null;default:0;index;uniqueIndex:uniq_house_fav_user_type_target" json:"community_id"`    // 小区 ID
	AgentID        uint   `gorm:"not null;default:0;index;uniqueIndex:uniq_house_fav_user_type_target" json:"agent_id"`        // 经纪人 ID
	FavoriteType   string `gorm:"size:16;not null;default:'house';index;uniqueIndex:uniq_house_fav_user_type_target" json:"favorite_type"` // house/community/agent/listing
	Notify         bool   `gorm:"not null;default:true" json:"notify"`                                                          // 价格变动通知
	Remark         string `gorm:"size:500;not null;default:''" json:"remark"`                                                   // 备注
}

// TableName 表名（house_ 前缀）
func (HouseFavorite) TableName() string { return "house_favorites" }

// HouseView 浏览记录表
type HouseView struct {
	database.BaseModel // id/created_at/updated_at/deleted_at
	RegionID         uint      `gorm:"not null;default:1;index" json:"region_id"`                       // 地区 ID
	UserID           uint      `gorm:"not null;index" json:"user_id"`                                   // 用户 ID
	HouseID          uint      `gorm:"not null;default:0;index" json:"house_id"`                        // 房源 ID
	ListingID        uint      `gorm:"not null;default:0;index" json:"listing_id"`                      // 发布 ID
	CommunityID      uint      `gorm:"not null;default:0;index" json:"community_id"`                    // 小区 ID
	AgentID          uint      `gorm:"not null;default:0;index" json:"agent_id"`                        // 经纪人 ID
	ViewType         string    `gorm:"size:16;not null;default:'house';index" json:"view_type"`         // house/community/agent/listing
	Source           string    `gorm:"size:32;not null;default:'list';index" json:"source"`             // list/search/recommend/share/qrcode/push
	IP               string    `gorm:"size:64;not null;default:''" json:"ip"`                           // 客户端 IP
	UserAgent        string    `gorm:"size:255;not null;default:''" json:"user_agent"`                  // User-Agent
	DurationSeconds  int       `gorm:"not null;default:0" json:"duration_seconds"`                      // 停留时长（秒）
	CreatedAt        time.Time `gorm:"index" json:"created_at"`                                         // 浏览时间
}

// TableName 表名（house_ 前缀）
func (HouseView) TableName() string { return "house_views" }
