// Package dto 同城分类信息数据传输对象
package dto

import "time"

// NewsInfo 分类信息详情
type NewsInfo struct {
	ID          uint       `json:"id"`
	Title       string     `json:"title"`
	Content     string     `json:"content"`
	CoverImage  string     `json:"cover_image"`
	Images      string     `json:"images"`
	Summary     string     `json:"summary"`
	AuthorID    uint       `json:"author_id"`
	AuthorName  string     `json:"author_name"`
	CategoryID  uint       `json:"category_id"`
	Tags        string     `json:"tags"`

	// 分类信息核心字段
	Price       float64 `json:"price"`
	PriceUnit   string  `json:"price_unit"`
	ListingType string  `json:"listing_type"`
	Condition   string  `json:"condition"`

	// 联系方式
	ContactPhone  string `json:"contact_phone"`
	ContactWechat string `json:"contact_wechat"`
	ContactQQ     string `json:"contact_qq"`

	// 位置信息
	Address   string  `json:"address"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`

	// 展示控制
	IsUrgent     bool       `json:"is_urgent"`
	ExpiryTime   *time.Time `json:"expiry_time"`
	ViewCount    int        `json:"view_count"`
	LikeCount    int        `json:"like_count"`
	FavCount     int        `json:"fav_count"`
	CommentCount int        `json:"comment_count"`
	Status       int        `json:"status"`
	RegionID     uint       `json:"region_id"`
	PublishedAt  *time.Time `json:"published_at"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

// CreateNewsRequest 创建分类信息请求
type CreateNewsRequest struct {
	Title      string  `json:"title" binding:"required,max=200"`
	Content    string  `json:"content"`
	CoverImage string  `json:"cover_image" binding:"max=255"`
	Images     string  `json:"images"`
	Summary    string  `json:"summary" binding:"max=500"`
	CategoryID uint    `json:"category_id"`
	Tags       string  `json:"tags" binding:"max=255"`
	Status     int     `json:"status" binding:"oneof=0 1"`

	// 分类信息核心字段
	Price       float64 `json:"price"`
	PriceUnit   string  `json:"price_unit"`
	ListingType string  `json:"listing_type" binding:"oneof=sell buy rent service job"`
	Condition   string  `json:"condition" binding:"oneof=new used"`

	// 联系方式
	ContactPhone  string `json:"contact_phone" binding:"max=20"`
	ContactWechat string `json:"contact_wechat" binding:"max=50"`
	ContactQQ     string `json:"contact_qq" binding:"max=20"`

	// 位置信息
	Address   string  `json:"address" binding:"max=255"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`

	// 过期天数（默认30天）
	ExpireDays int `json:"expire_days"`
}

// UpdateNewsRequest 更新分类信息请求
type UpdateNewsRequest struct {
	Title      string  `json:"title" binding:"max=200"`
	Content    string  `json:"content"`
	CoverImage string  `json:"cover_image" binding:"max=255"`
	Images     string  `json:"images"`
	Summary    string  `json:"summary" binding:"max=500"`
	CategoryID uint    `json:"category_id"`
	Tags       string  `json:"tags" binding:"max=255"`
	Status     int     `json:"status" binding:"omitempty,oneof=0 1 2 3"`

	Price       float64 `json:"price"`
	PriceUnit   string  `json:"price_unit"`
	ListingType string  `json:"listing_type" binding:"omitempty,oneof=sell buy rent service job"`
	Condition   string  `json:"condition" binding:"omitempty,oneof=new used"`

	ContactPhone  string `json:"contact_phone" binding:"max=20"`
	ContactWechat string `json:"contact_wechat" binding:"max=50"`
	ContactQQ     string `json:"contact_qq" binding:"max=20"`

	Address   string  `json:"address" binding:"max=255"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`

	IsUrgent   bool       `json:"is_urgent"`
	ExpiryTime *time.Time `json:"expiry_time"`
}

// NewsListRequest 分类信息列表查询请求
type NewsListRequest struct {
	Page        int    `form:"page"`
	PageSize    int    `form:"page_size"`
	CategoryID  uint   `form:"category_id"`
	Status      int    `form:"status"`
	Keyword     string `form:"keyword"`
	ListingType string `form:"listing_type"`
	MinPrice    float64 `form:"min_price"`
	MaxPrice    float64 `form:"max_price"`
	IsUrgent    *bool  `form:"is_urgent"`
	Sort        string `form:"sort"` // time/price/views
}

// LikeResponse 点赞操作/状态响应
type LikeResponse struct {
	Liked     bool `json:"liked"`
	LikeCount int  `json:"like_count"`
}

// FavResponse 收藏操作/状态响应
type FavResponse struct {
	Faved    bool `json:"faved"`
	FavCount int  `json:"fav_count"`
}

// NewsSearchRequest 全文检索请求
type NewsSearchRequest struct {
	Page       int    `form:"page"`
	PageSize   int    `form:"page_size"`
	Keyword    string `form:"keyword"`
	CategoryID uint   `form:"category_id"`
	ListingType string `form:"listing_type"`
}

// CommentInfo 评论信息
type CommentInfo struct {
	ID        uint      `json:"id"`
	NewsID    uint      `json:"news_id"`
	UserID    uint      `json:"user_id"`
	UserName  string    `json:"user_name"`
	Avatar    string    `json:"avatar"`
	Content   string    `json:"content"`
	ParentID  *uint     `json:"parent_id"`
	ReplyTo   string    `json:"reply_to"`
	CreatedAt time.Time `json:"created_at"`
}

// CreateCommentRequest 创建评论请求
type CreateCommentRequest struct {
	Content  string `json:"content" binding:"required,max=500"`
	ParentID *uint  `json:"parent_id"`
	ReplyTo  string `json:"reply_to" binding:"max=50"`
}

// MessageInfo 消息信息
type MessageInfo struct {
	ID         uint      `json:"id"`
	FromUserID uint      `json:"from_user_id"`
	ToUserID   uint      `json:"to_user_id"`
	NewsID     *uint     `json:"news_id"`
	Type       string    `json:"type"`
	Content    string    `json:"content"`
	IsRead     bool      `json:"is_read"`
	CreatedAt  time.Time `json:"created_at"`
}