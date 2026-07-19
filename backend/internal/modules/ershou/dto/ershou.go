// Package dto 同城二手物品数据传输对象
// 依据需求文档 2.2.A.10：商品发布/分类/搜索/留言/交易
package dto

import (
	"time"

	"wuchang-tongcheng/internal/pkg/utils"
)

// ErshouInfo 二手物品详情
type ErshouInfo struct {
	ID          uint       `json:"id"`
	Title       string     `json:"title"`
	Content     string     `json:"content"`
	CoverImage  string     `json:"cover_image"`
	Images      []string   `json:"images"`           // 图片URL列表（由 service 拼装）
	Summary     string     `json:"summary"`

	// 发布者
	UserID     uint   `json:"user_id"`
	UserName   string `json:"user_name"`
	UserPhone  string `json:"user_phone"`
	UserAvatar string `json:"user_avatar"`

	// 分类
	CategoryID   uint   `json:"category_id"`
	CategoryName string `json:"category_name"`

	// 二手核心字段
	Price         float64 `json:"price"`
	OriginalPrice float64 `json:"original_price"`
	PriceUnit     string  `json:"price_unit"`
	Condition     string  `json:"condition"`
	Brand         string  `json:"brand"`

	// 联系方式
	ContactPhone  string `json:"contact_phone"`
	ContactWechat string `json:"contact_wechat"`

	// 位置
	Address   string  `json:"address"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Distance  float64 `json:"distance,omitempty"` // 仅附近查询返回（公里）

	// 交易方式
	DeliveryMethod string `json:"delivery_method"`

	// 展示控制
	IsUrgent     bool       `json:"is_urgent"`
	ExpiryTime   *time.Time `json:"expiry_time"`
	ViewCount    int        `json:"view_count"`
	FavCount     int        `json:"fav_count"`
	MessageCount int        `json:"message_count"`

	// 状态
	Status      int        `json:"status"`
	AuditStatus int        `json:"audit_status"`
	AuditReason string     `json:"audit_reason"`
	PublishedAt *time.Time `json:"published_at"`
	RegionID    uint       `json:"region_id"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`

	// 当前用户是否已收藏（仅登录态列表/详情返回）
	HasFaved bool `json:"has_faved,omitempty"`
}

// CreateErshouRequest C端发布二手物品请求
type CreateErshouRequest struct {
	Title       string   `json:"title" binding:"required,max=200"`
	Content     string   `json:"content"`
	CoverImage  string   `json:"cover_image" binding:"max=255"`
	Images      []string `json:"images"`              // 图片URL列表
	Summary     string   `json:"summary" binding:"max=500"`
	CategoryID  uint     `json:"category_id"`
	Price       float64  `json:"price" binding:"gte=0"`
	OriginalPrice float64 `json:"original_price" binding:"gte=0"`
	PriceUnit   string   `json:"price_unit" binding:"omitempty,oneof=元 万元 面议 免费"`
	Condition   string   `json:"condition" binding:"omitempty,oneof=new almost_new used broken"`
	Brand       string   `json:"brand" binding:"max=100"`
	ContactPhone  string `json:"contact_phone" binding:"max=20"`
	ContactWechat string `json:"contact_wechat" binding:"max=50"`
	Address     string   `json:"address" binding:"max=255"`
	Latitude    float64  `json:"latitude"`
	Longitude   float64  `json:"longitude"`
	DeliveryMethod string `json:"delivery_method" binding:"omitempty,oneof=face self express"`
	IsUrgent    bool     `json:"is_urgent"`
	ExpireDays  int      `json:"expire_days"`  // 过期天数（默认30天）
	Status      int      `json:"status" binding:"oneof=0 1"`  // 0草稿 1直接发布
}

// UpdateErshouRequest 更新二手物品请求
type UpdateErshouRequest struct {
	Title       string   `json:"title" binding:"max=200"`
	Content     string   `json:"content"`
	CoverImage  string   `json:"cover_image" binding:"max=255"`
	Images      []string `json:"images"`
	Summary     string   `json:"summary" binding:"max=500"`
	CategoryID  uint     `json:"category_id"`
	Price       float64  `json:"price" binding:"gte=0"`
	OriginalPrice float64 `json:"original_price" binding:"gte=0"`
	PriceUnit   string   `json:"price_unit" binding:"omitempty,oneof=元 万元 面议 免费"`
	Condition   string   `json:"condition" binding:"omitempty,oneof=new almost_new used broken"`
	Brand       string   `json:"brand" binding:"max=100"`
	ContactPhone  string `json:"contact_phone" binding:"max=20"`
	ContactWechat string `json:"contact_wechat" binding:"max=50"`
	Address     string   `json:"address" binding:"max=255"`
	Latitude    float64  `json:"latitude"`
	Longitude   float64  `json:"longitude"`
	DeliveryMethod string `json:"delivery_method" binding:"omitempty,oneof=face self express"`
	IsUrgent    *bool    `json:"is_urgent"`
	ExpireDays  int      `json:"expire_days"`
	Status      int      `json:"status" binding:"omitempty,oneof=0 1 3"` // 0草稿 1发布 3下架
}

// ErshouListRequest 列表查询（C端）
type ErshouListRequest struct {
	CategoryID uint    `form:"category_id" json:"category_id"`
	Keyword    string  `form:"keyword" json:"keyword"`
	MinPrice   float64 `form:"min_price" json:"min_price"`
	MaxPrice   float64 `form:"max_price" json:"max_price"`
	Condition  string  `form:"condition" json:"condition"`
	Brand      string  `form:"brand" json:"brand"`
	IsUrgent   *bool   `form:"is_urgent" json:"is_urgent"`
	Sort       string  `form:"sort" json:"sort"` // latest/price_asc/price_desc/popular
	utils.Pagination
}

// ErshouNearbyRequest 附近查询
type ErshouNearbyRequest struct {
	Latitude  float64 `form:"latitude" binding:"required"`
	Longitude float64 `form:"longitude" binding:"required"`
	RadiusKm  float64 `form:"radius_km"` // 默认 5 公里
	utils.Pagination
}

// ErshouSearchRequest 搜索（走 Elasticsearch）
type ErshouSearchRequest struct {
	Keyword string `form:"keyword" binding:"required,max=100"`
	utils.Pagination
}

// ErshouAdminListRequest 管理后台列表查询（M端）
// Status/AuditStatus 使用 *int 指针，nil 表示不传该过滤条件（返回全部），
// 区分"未传"（nil，不过滤）和"传0"（筛选草稿/待审）。
type ErshouAdminListRequest struct {
	RegionID    uint   `form:"region_id" json:"region_id"`
	UserID      uint   `form:"user_id" json:"user_id"`
	CategoryID  uint   `form:"category_id" json:"category_id"`
	Status      *int   `form:"status" json:"status"`         // nil全部 0草稿 1已发布 2已售 3下架 4过期
	AuditStatus *int   `form:"audit_status" json:"audit_status"` // nil全部 0待审 1通过 2拒绝
	Keyword     string `form:"keyword" json:"keyword"`
	utils.Pagination
}

// AuditRequest 审核操作请求（M端）
type AuditRequest struct {
	AuditStatus int    `json:"audit_status" binding:"oneof=0 1 2"` // 0待审 1通过 2拒绝
	AuditReason string `json:"audit_reason" binding:"max=500"`
}

// AdminUpdateStatusRequest 管理后台强制下架/恢复
type AdminUpdateStatusRequest struct {
	Status int `json:"status" binding:"oneof=1 3 4"` // 1发布 3下架 4过期
}

// CreateMessageRequest C端用户留言请求
type CreateMessageRequest struct {
	Content string `json:"content" binding:"required,max=500"`
}

// MessageInfo 留言信息
type MessageInfo struct {
	ID         uint      `json:"id"`
	ErshouID   uint      `json:"ershou_id"`
	FromUserID uint      `json:"from_user_id"`
	FromName   string    `json:"from_name"`
	FromAvatar string    `json:"from_avatar"`
	Content    string    `json:"content"`
	IsRead     bool      `json:"is_read"`
	CreatedAt  time.Time `json:"created_at"`
}

// FavResponse 收藏操作响应
type FavResponse struct {
	HasFaved bool `json:"has_faved"`
	FavCount int  `json:"fav_count"`
}
