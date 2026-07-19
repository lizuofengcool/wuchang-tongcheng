// Package model 房源发布表（与 houses 主表 1:1 冗余发布信息）
// 对标贝壳/链家：发布人/发布类型/发布时间/有效期/发布状态/审核
package model

import (
	"time"

	"wuchang-tongcheng/internal/pkg/database"
)

// === 发布状态常量（与 houses.status 同步，便于冗余查询） ===
const (
	ListingStatusDraft     = 0 // 草稿
	ListingStatusPublished = 1 // 已发布
	ListingStatusOffline   = 2 // 已下架
	ListingStatusExpired   = 3 // 已过期
	ListingStatusRejected  = 4 // 已拒绝
)

// === 发布者类型常量 ===
const (
	PublisherTypePersonal  = "personal"  // 个人
	PublisherTypeAgent     = "agent"     // 经纪人
	PublisherTypeDeveloper = "developer" // 开发商
)

// HouseListing 房源发布表
type HouseListing struct {
	database.RegionBaseModel // id/region_id/created_at/updated_at/deleted_at
	ListingNo        string     `gorm:"size:64;not null;uniqueIndex" json:"listing_no"`                       // 发布单号
	HouseID          uint       `gorm:"not null;index" json:"house_id"`                                       // 关联房源 ID
	CommunityID      uint       `gorm:"not null;default:0;index" json:"community_id"`                         // 关联小区 ID
	AgentID          uint       `gorm:"not null;default:0;index" json:"agent_id"`                             // 关联经纪人 ID
	PublisherID      uint       `gorm:"not null;index:idx_house_listings_publisher" json:"publisher_id"`       // 发布人 ID
	PublisherName    string     `gorm:"size:50;not null;default:''" json:"publisher_name"`                    // 发布人昵称
	PublisherPhone   string     `gorm:"size:20;not null;default:''" json:"publisher_phone"`                   // 发布人手机
	PublisherAvatar  string     `gorm:"size:255;not null;default:''" json:"publisher_avatar"`                 // 发布人头像
	PublisherType    string     `gorm:"size:16;not null;default:'personal'" json:"publisher_type"`            // 发布者类型
	ListingType      string     `gorm:"size:16;not null;default:'rent';index" json:"listing_type"`            // rent/sale/transfer
	Title            string     `gorm:"size:200;not null;default:''" json:"title"`                            // 发布标题
	Description      string     `gorm:"type:text" json:"description"`                                          // 发布描述
	Price            float64    `gorm:"type:decimal(14,2);default:0" json:"price"`                            // 价格
	PriceUnit        string     `gorm:"size:16;not null;default:'month'" json:"price_unit"`                   // 价格单位
	Decoration       string     `gorm:"size:16;not null;default:'rough'" json:"decoration"`                    // 装修
	Orientation      string     `gorm:"size:32;not null;default:''" json:"orientation"`                       // 朝向
	Layout           string     `gorm:"size:32;not null;default:''" json:"layout"`                            // 户型
	BuildingArea     float64    `gorm:"type:decimal(10,2);default:0" json:"building_area"`                    // 建筑面积
	Status           int        `gorm:"default:0;index:idx_house_listings_publisher;index" json:"status"`     // 0草稿 1已发布 2下架 3过期 4拒绝
	AuditStatus      int        `gorm:"default:0;index" json:"audit_status"`                                  // 0待审 1通过 2拒绝
	AuditReason      string     `gorm:"size:500;not null;default:''" json:"audit_reason"`                     // 审核拒绝原因
	PublishedAt      *time.Time `gorm:"index" json:"published_at"`                                            // 发布时间
	ExpiredAt        *time.Time `gorm:"index" json:"expired_at"`                                              // 过期时间
	RefreshedAt      *time.Time `gorm:"index" json:"refreshed_at"`                                            // 最近刷新时间
	OfflineAt        *time.Time `gorm:"index" json:"offline_at"`                                              // 下线时间
	RefreshCount     int        `gorm:"not null;default:0" json:"refresh_count"`                              // 刷新次数
	ViewCount        int        `gorm:"not null;default:0" json:"view_count"`                                 // 浏览数
	FavCount         int        `gorm:"not null;default:0" json:"fav_count"`                                  // 收藏数
	ContactCount     int        `gorm:"not null;default:0" json:"contact_count"`                              // 联系数
}

// TableName 表名（house_ 前缀）
func (HouseListing) TableName() string { return "house_listings" }
