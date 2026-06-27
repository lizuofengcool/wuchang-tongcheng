// Package dto 同城头条数据传输对象
package dto

import "time"

// NewsInfo 头条信息
type NewsInfo struct {
	ID          uint       `json:"id"`
	Title       string     `json:"title"`
	Content     string     `json:"content"`
	CoverImage  string     `json:"cover_image"`
	Summary     string     `json:"summary"`
	AuthorID    uint       `json:"author_id"`
	AuthorName  string     `json:"author_name"`
	CategoryID  uint       `json:"category_id"`
	Tags        string     `json:"tags"`
	ViewCount   int        `json:"view_count"`
	LikeCount   int        `json:"like_count"`
	Status      int        `json:"status"`
	PublishedAt *time.Time `json:"published_at"`
	CreatedAt   time.Time  `json:"created_at"`
}

// CreateNewsRequest 创建头条请求
type CreateNewsRequest struct {
	Title      string `json:"title" binding:"required,max=200"`
	Content    string `json:"content"`
	CoverImage string `json:"cover_image" binding:"max=255"`
	Summary    string `json:"summary" binding:"max=500"`
	CategoryID uint   `json:"category_id"`
	Tags       string `json:"tags" binding:"max=255"`
	Status     int    `json:"status" binding:"oneof=0 1"` // 0草稿 1发布
}

// UpdateNewsRequest 更新头条请求
type UpdateNewsRequest struct {
	Title      string `json:"title" binding:"max=200"`
	Content    string `json:"content"`
	CoverImage string `json:"cover_image" binding:"max=255"`
	Summary    string `json:"summary" binding:"max=500"`
	CategoryID uint   `json:"category_id"`
	Tags       string `json:"tags" binding:"max=255"`
	Status     int    `json:"status" binding:"omitempty,oneof=0 1 2"`
}

// NewsListRequest 头条列表查询请求
type NewsListRequest struct {
	Page       int    `form:"page"`
	PageSize   int    `form:"page_size"`
	CategoryID uint   `form:"category_id"`
	Status     int    `form:"status"`
	Keyword    string `form:"keyword"`
}

// LikeResponse 点赞操作/状态响应
type LikeResponse struct {
	Liked     bool `json:"liked"`      // 当前用户是否已点赞
	LikeCount int  `json:"like_count"` // 该头条的总点赞数
}

// NewsSearchRequest 头条全文检索请求（走 Elasticsearch，ES 不可用时降级到 DB LIKE）
type NewsSearchRequest struct {
	Page     int    `form:"page"`
	PageSize int    `form:"page_size"`
	Keyword  string `form:"keyword"`  // 检索关键词（匹配 title/content/summary/tags）
	CategoryID uint `form:"category_id"` // 可选过滤
}
