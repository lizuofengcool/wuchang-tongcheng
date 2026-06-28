// Package model 同城分类信息数据模型
// 包含分类信息发布、点赞、收藏、评论、消息等核心数据模型
package model

import (
	"time"

	"wuchang-tongcheng/internal/pkg/database"
)

// 分类信息类型常量
const (
	ListingTypeSell    = "sell"    // 出售
	ListingTypeBuy     = "buy"     // 求购
	ListingTypeRent    = "rent"    // 出租
	ListingTypeService = "service" // 服务
	ListingTypeJob     = "job"     // 招聘
)

// 成色常量
const (
	ConditionNew  = "new"  // 全新
	ConditionUsed = "used" // 二手
)

// 信息状态常量
const (
	StatusDraft     = 0 // 草稿
	StatusPublished = 1 // 已发布
	StatusOffline   = 2 // 已下架
	StatusExpired   = 3 // 已过期
)

// News 同城分类信息模型（从纯文章扩展为真正的分类信息发布）
type News struct {
	database.RegionBaseModel
	Title      string `gorm:"size:200;not null" json:"title"` // 标题
	Content    string `gorm:"type:text" json:"content"`       // 详情描述
	CoverImage string `gorm:"size:255" json:"cover_image"`    // 封面图
	Images     string `gorm:"type:text" json:"images"`        // 图片集（JSON数组URL）
	Summary    string `gorm:"size:500" json:"summary"`        // 摘要
	AuthorID   uint   `gorm:"index" json:"author_id"`         // 发布者ID
	AuthorName string `gorm:"size:50" json:"author_name"`     // 发布者名
	CategoryID uint   `gorm:"index" json:"category_id"`       // 分类ID
	Tags       string `gorm:"size:255" json:"tags"`           // 标签（逗号分隔）

	// === 分类信息核心字段 ===
	Price       float64 `gorm:"type:decimal(12,2);default:0" json:"price"`         // 价格
	PriceUnit   string  `gorm:"size:20;default:'元'" json:"price_unit"`             // 价格单位：元/万元/面议
	ListingType string  `gorm:"size:20;index;default:'sell'" json:"listing_type"`  // 信息类型：sell/buy/rent/service/job
	Condition   string  `gorm:"size:20;default:'used'" json:"condition"`           // 成色：new/used

	// === 联系方式 ===
	ContactPhone  string `gorm:"size:20" json:"contact_phone"`  // 联系电话
	ContactWechat string `gorm:"size:50" json:"contact_wechat"` // 微信号
	ContactQQ     string `gorm:"size:20" json:"contact_qq"`     // QQ号

	// === 位置信息 ===
	Address   string  `gorm:"size:255" json:"address"`            // 详细地址
	Latitude  float64 `gorm:"type:decimal(10,7)" json:"latitude"` // 纬度
	Longitude float64 `gorm:"type:decimal(10,7)" json:"longitude"` // 经度

	// === 展示控制 ===
	IsUrgent     bool       `gorm:"default:false;index" json:"is_urgent"` // 是否置顶/加急
	ExpiryTime   *time.Time `gorm:"index" json:"expiry_time"`             // 过期时间
	ViewCount    int        `gorm:"default:0" json:"view_count"`          // 浏览量
	LikeCount    int        `gorm:"default:0" json:"like_count"`          // 点赞数
	FavCount     int        `gorm:"default:0" json:"fav_count"`           // 收藏数
	CommentCount int        `gorm:"default:0" json:"comment_count"`       // 评论数
	Status       int        `gorm:"default:0;index" json:"status"`        // 状态 0草稿 1已发布 2下架 3过期
	PublishedAt  *time.Time `gorm:"index" json:"published_at"`            // 发布时间
}

func (News) TableName() string { return "news" }

// NewsLike 点赞记录
type NewsLike struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    uint      `gorm:"not null;uniqueIndex:uniq_user_news" json:"user_id"`
	NewsID    uint      `gorm:"not null;uniqueIndex:uniq_user_news" json:"news_id"`
	CreatedAt time.Time `json:"created_at"`
}

func (NewsLike) TableName() string { return "news_likes" }

// NewsFavorite 收藏记录
type NewsFavorite struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    uint      `gorm:"not null;uniqueIndex:uniq_user_news_fav" json:"user_id"`
	NewsID    uint      `gorm:"not null;uniqueIndex:uniq_user_news_fav" json:"news_id"`
	CreatedAt time.Time `json:"created_at"`
}

func (NewsFavorite) TableName() string { return "news_favorites" }

// NewsComment 评论
type NewsComment struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	NewsID    uint      `gorm:"not null;index" json:"news_id"`
	UserID    uint      `gorm:"not null;index" json:"user_id"`
	UserName  string    `gorm:"size:50" json:"user_name"`
	Avatar    string    `gorm:"size:255" json:"avatar"`
	Content   string    `gorm:"type:text;not null" json:"content"`
	ParentID  *uint     `gorm:"index" json:"parent_id"`  // 父评论ID（回复）
	ReplyTo   string    `gorm:"size:50" json:"reply_to"` // 回复给谁
	Status    int       `gorm:"default:1;index" json:"status"` // 0删除 1正常
	CreatedAt time.Time `json:"created_at"`
}

func (NewsComment) TableName() string { return "news_comments" }

// Message 用户消息/通知
type Message struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	FromUserID uint      `gorm:"index" json:"from_user_id"`          // 发送者
	ToUserID   uint      `gorm:"not null;index" json:"to_user_id"`   // 接收者
	NewsID     *uint     `gorm:"index" json:"news_id"`               // 关联信息ID
	Type       string    `gorm:"size:20;not null;index" json:"type"` // 类型: comment/reply/like/fav/system
	Content    string    `gorm:"type:text" json:"content"`           // 消息内容
	IsRead     bool      `gorm:"default:false;index" json:"is_read"` // 已读
	CreatedAt  time.Time `json:"created_at"`
}

func (Message) TableName() string { return "messages" }