// Package model 车源发布表（对标瓜子）
// 新车/二手/置换/租车 + 车商/个人 + 审核/检测状态
package model

import (
	"time"

	"wuchang-tongcheng/internal/pkg/database"
)

// === 发布状态常量 ===
const (
	ListingStatusDraft     = 0 // 草稿
	ListingStatusPublished = 1 // 已发布
	ListingStatusOffline   = 2 // 已下架
	ListingStatusExpired   = 3 // 已过期
	ListingStatusSold      = 4 // 已售出
)

// === 发布审核状态常量 ===
const (
	ListingAuditPending  = 0 // 待审
	ListingAuditApproved = 1 // 通过
	ListingAuditRejected = 2 // 拒绝
)

// === 检测状态常量 ===
const (
	ListingInspectionNone       = 0 // 未检测
	ListingInspectionPending    = 1 // 待检测
	ListingInspectionInProgress = 2 // 检测中
	ListingInspectionPassed     = 3 // 检测通过
	ListingInspectionFailed     = 4 // 检测不通过
)

// === 发布者类型常量 ===
const (
	PublisherTypePersonal = "personal" // 个人
	PublisherTypeDealer   = "dealer"   // 车商
	PublisherTypeAgent    = "agent"    // 经纪人
)

// CarListing 车源发布表
type CarListing struct {
	database.RegionBaseModel // id/region_id/created_at/updated_at/deleted_at
	ListingNo          string     `gorm:"size:64;not null;uniqueIndex" json:"listing_no"`                // 发布单号
	CarID              uint       `gorm:"not null;index" json:"car_id"`                                  // 车源 ID
	ModelID            uint       `gorm:"not null;default:0;index" json:"model_id"`                      // 车型 ID
	PublisherID        uint       `gorm:"not null;index" json:"publisher_id"`                            // 发布者 ID
	PublisherName      string     `gorm:"size:50;not null;default:''" json:"publisher_name"`             // 发布者姓名
	PublisherAvatar    string     `gorm:"size:255;not null;default:''" json:"publisher_avatar"`          // 发布者头像
	PublisherType      string     `gorm:"size:16;not null;default:'personal';index" json:"publisher_type"` // personal/dealer/agent
	DealerID           uint       `gorm:"not null;default:0;index" json:"dealer_id"`                     // 车商 ID
	DealerName         string     `gorm:"size:128;not null;default:''" json:"dealer_name"`               // 车商名
	ListingType        string     `gorm:"size:16;not null;default:'used';index" json:"listing_type"`     // new/used/replace/rental
	Title              string     `gorm:"size:200;not null;default:''" json:"title"`                     // 发布标题
	Description        string     `gorm:"type:text" json:"description"`                                  // 发布描述
	Price              float64    `gorm:"type:decimal(14,2);default:0;index" json:"price"`               // 售价
	OriginalPrice      float64    `gorm:"type:decimal(14,2);default:0" json:"original_price"`            // 原价
	PriceNegotiable    bool       `gorm:"not null;default:false" json:"price_negotiable"`                // 价格可议
	Status             int        `gorm:"default:0;index" json:"status"`                                 // 0草稿 1已发布 2下架 3过期 4已售
	AuditStatus        int        `gorm:"default:0;index" json:"audit_status"`                           // 0待审 1通过 2拒绝
	AuditReason        string     `gorm:"size:500;not null;default:''" json:"audit_reason"`              // 审核拒绝原因
	PublishedAt        *time.Time `gorm:"index" json:"published_at"`                                     // 发布时间
	OfflineAt          *time.Time `gorm:"index" json:"offline_at"`                                       // 下架时间
	ExpiredAt          *time.Time `gorm:"index" json:"expired_at"`                                       // 过期时间
	SoldAt             *time.Time `gorm:"index" json:"sold_at"`                                          // 售出时间
	ViewCount          int        `gorm:"not null;default:0" json:"view_count"`                          // 浏览数
	FavCount           int        `gorm:"not null;default:0" json:"fav_count"`                           // 收藏数
	ContactCount       int        `gorm:"not null;default:0" json:"contact_count"`                       // 联系数
	TestDriveCount     int        `gorm:"not null;default:0" json:"test_drive_count"`                    // 试驾预约数
	InspectionStatus   int        `gorm:"default:0;index" json:"inspection_status"`                      // 0未检测 1待检 2检测中 3通过 4不通过
	InspectionID       uint       `gorm:"not null;default:0" json:"inspection_id"`                       // 检测单 ID
	RealCarVerified    bool       `gorm:"not null;default:false;index" json:"real_car_verified"`          // 真车认证
	Featured           bool       `gorm:"not null;default:false;index" json:"featured"`                  // 精选
	PromotionLevel     int        `gorm:"not null;default:0" json:"promotion_level"`                     // 推广等级
}

// TableName 表名（car_ 前缀）
func (CarListing) TableName() string { return "car_listings" }
