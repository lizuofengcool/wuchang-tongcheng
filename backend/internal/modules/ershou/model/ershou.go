// Package model 同城二手物品数据模型
// 依据需求文档 2.2.A.10：商品发布/分类/搜索/留言/交易
// 依据需求文档 1.10：4 维数据隔离（region_id 地区隔离 + user_id 用户隔离）
// 依据需求文档 7.1：通用字段 id/region_id/created_at/updated_at/deleted_at + status + audit_status
// 依据需求文档 7.2：表命名 erhous / ershou_images / ershou_favorites / ershou_messages
package model

import (
	"time"

	"wuchang-tongcheng/internal/pkg/database"
)

// === 状态常量 ===
const (
	StatusDraft     = 0 // 草稿
	StatusPublished = 1 // 已发布
	StatusSold      = 2 // 已售出
	StatusOffline   = 3 // 已下架
	StatusExpired   = 4 // 已过期
)

// === 审核状态常量（依据需求文档 7.1） ===
const (
	AuditPending = 0 // 待审
	AuditApproved = 1 // 通过
	AuditRejected = 2 // 拒绝
)

// === 成色常量 ===
const (
	ConditionNew        = "new"         // 全新
	ConditionAlmostNew  = "almost_new"  // 9成新
	ConditionUsed       = "used"        // 二手
	ConditionBroken     = "broken"      // 有瑕疵
)

// === 交易方式常量 ===
const (
	DeliveryFace    = "face"    // 当面交易
	DeliverySelf    = "self"    // 自提
	DeliveryExpress = "express" // 快递
)

// Ershou 同城二手物品主表
type Ershou struct {
	database.RegionBaseModel // 含 id/region_id/created_at/updated_at/deleted_at（地区隔离）

	// === 基础信息 ===
	Title      string `gorm:"size:200;not null" json:"title"`       // 标题
	Content    string `gorm:"type:text" json:"content"`             // 详情描述
	CoverImage string `gorm:"size:255" json:"cover_image"`          // 封面图
	Summary    string `gorm:"size:500" json:"summary"`              // 摘要

	// === 发布者（用户隔离，依据 1.10.1） ===
	UserID     uint   `gorm:"index;not null" json:"user_id"`        // 发布者ID
	UserName   string `gorm:"size:50" json:"user_name"`             // 发布者昵称
	UserPhone  string `gorm:"size:20" json:"user_phone"`            // 发布者手机
	UserAvatar string `gorm:"size:255" json:"user_avatar"`          // 发布者头像

	// === 分类 ===
	CategoryID uint   `gorm:"index" json:"category_id"`             // 分类ID（通用 category 表）
	CategoryName string `gorm:"size:50" json:"category_name"`       // 分类名（冗余，便于列表展示）

	// === 二手核心字段 ===
	Price         float64 `gorm:"type:decimal(12,2);default:0" json:"price"`           // 售价
	OriginalPrice float64 `gorm:"type:decimal(12,2);default:0" json:"original_price"`  // 原价（可选，用于展示折扣）
	PriceUnit     string  `gorm:"size:20;default:'元'" json:"price_unit"`              // 价格单位：元/面议/免费
	Condition     string  `gorm:"size:20;default:'used'" json:"condition"`             // 成色：new/almost_new/used/broken
	Brand         string  `gorm:"size:100" json:"brand"`                              // 品牌（可选）

	// === 联系方式 ===
	ContactPhone  string `gorm:"size:20" json:"contact_phone"`   // 联系电话
	ContactWechat string `gorm:"size:50" json:"contact_wechat"`  // 微信号（可选）

	// === 位置信息（PostGIS 附近查询用） ===
	Address   string  `gorm:"size:255" json:"address"`            // 详细地址
	Latitude  float64 `gorm:"type:decimal(10,7)" json:"latitude"`  // 纬度
	Longitude float64 `gorm:"type:decimal(10,7)" json:"longitude"` // 经度

	// === 交易方式 ===
	DeliveryMethod string `gorm:"size:50;default:'face'" json:"delivery_method"` // face/self/express

	// === 展示控制 ===
	IsUrgent     bool       `gorm:"default:false;index" json:"is_urgent"`  // 是否加急置顶
	ExpiryTime   *time.Time `gorm:"index" json:"expiry_time"`              // 过期时间
	ViewCount    int        `gorm:"default:0" json:"view_count"`           // 浏览量
	FavCount     int        `gorm:"default:0" json:"fav_count"`            // 收藏数
	MessageCount int        `gorm:"default:0" json:"message_count"`        // 留言数

	// === 状态（依据 7.1：status + audit_status） ===
	Status      int        `gorm:"default:0;index" json:"status"`         // 0草稿 1已发布 2已售出 3下架 4过期
	AuditStatus int        `gorm:"default:0;index" json:"audit_status"`   // 0待审 1通过 2拒绝
	AuditReason string     `gorm:"size:500" json:"audit_reason"`          // 审核拒绝原因
	PublishedAt *time.Time `gorm:"index" json:"published_at"`             // 发布时间

	// Distance 仅在"附近"查询时由 SQL 计算并回填，非持久化字段（公里）
	Distance float64 `gorm:"-" json:"-"`
}

// TableName 表名（依据需求文档 7.2：主表 {module}s）
func (Ershou) TableName() string { return "ershous" }

// ErshouImage 二手物品图片子表（依据 7.2：子表 {module}_{sub}）
type ErshouImage struct {
	ID        uint      `gorm:"primarykey" json:"id"`
	ErshouID  uint      `gorm:"not null;index" json:"ershou_id"`  // 关联二手物品ID
	URL       string    `gorm:"size:255;not null" json:"url"`     // 图片URL
	Sort      int       `gorm:"default:0" json:"sort"`            // 排序（越小越靠前）
	CreatedAt time.Time `json:"created_at"`
}

// TableName 表名
func (ErshouImage) TableName() string { return "ershou_images" }

// ErshouFavorite 收藏关联表（依据 7.2：关联表 {module}_{rel}）
type ErshouFavorite struct {
	ID        uint      `gorm:"primarykey" json:"id"`
	UserID    uint      `gorm:"not null;uniqueIndex:uniq_user_ershou_fav" json:"user_id"`
	ErshouID  uint      `gorm:"not null;uniqueIndex:uniq_user_ershou_fav" json:"ershou_id"`
	CreatedAt time.Time `json:"created_at"`
}

// TableName 表名
func (ErshouFavorite) TableName() string { return "ershou_favorites" }

// ErshouMessage 留言/咨询关联表（C端用户向发布者留言）
type ErshouMessage struct {
	ID         uint      `gorm:"primarykey" json:"id"`
	ErshouID   uint      `gorm:"not null;index" json:"ershou_id"`    // 关联二手物品ID
	FromUserID uint      `gorm:"not null;index" json:"from_user_id"` // 留言者ID
	FromName   string    `gorm:"size:50" json:"from_name"`           // 留言者昵称
	FromAvatar string    `gorm:"size:255" json:"from_avatar"`        // 留言者头像
	Content    string    `gorm:"type:text;not null" json:"content"`  // 留言内容
	IsRead     bool      `gorm:"default:false;index" json:"is_read"` // 发布者是否已读
	Status     int       `gorm:"default:1;index" json:"status"`      // 0删除 1正常
	CreatedAt  time.Time `json:"created_at"`
}

// TableName 表名
func (ErshouMessage) TableName() string { return "ershou_messages" }
