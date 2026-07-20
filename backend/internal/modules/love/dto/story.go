// Package dto love 相亲交友数据传输对象 - 动态广场
package dto

import (
	"time"

	"wuchang-tongcheng/internal/pkg/utils"
)

// LoveStoryInfo 动态响应
type LoveStoryInfo struct {
	ID            uint       `json:"id"`
	StoryNo       string     `json:"story_no"`
	LoveID        uint       `json:"love_id"`
	UserID        uint       `json:"user_id"`
	UserNickname  string     `json:"user_nickname"`
	UserAvatar    string     `json:"user_avatar"`
	UserGender    int        `json:"user_gender"`
	UserVerified  bool       `json:"user_verified"`
	Title         string     `json:"title"`
	Content       string     `json:"content"`
	MediaType     string     `json:"media_type"`
	ImageUrls     interface{} `json:"image_urls"`
	VideoURL      string     `json:"video_url"`
	VideoCover    string     `json:"video_cover"`
	VideoDuration int        `json:"video_duration"`
	VoiceURL      string     `json:"voice_url"`
	VoiceDuration int        `json:"voice_duration"`
	Location      string     `json:"location"`
	Latitude      float64    `json:"latitude"`
	Longitude     float64    `json:"longitude"`
	Tags          interface{} `json:"tags"`
	Topic         string     `json:"topic"`
	ViewCount     int        `json:"view_count"`
	LikeCount     int        `json:"like_count"`
	CommentCount  int        `json:"comment_count"`
	ShareCount    int        `json:"share_count"`
	ReportCount   int        `json:"report_count"`
	Featured      bool       `json:"featured"`
	Status        int        `json:"status"`
	StatusText    string     `json:"status_text"`
	AuditStatus   int        `json:"audit_status"`
	AuditStatusText string   `json:"audit_status_text"`
	AuditReason   string     `json:"audit_reason"`
	PublishedAt   *time.Time `json:"published_at"`
	HotScore      float64    `json:"hot_score"`
	HasLiked      bool       `json:"has_liked"`
	HasFaved      bool       `json:"has_faved"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

// CreateLoveStoryRequest 发布动态请求
type CreateLoveStoryRequest struct {
	Title         string      `json:"title" binding:"max=200"`
	Content       string      `json:"content" binding:"max=5000"`
	MediaType     string      `json:"media_type" binding:"omitempty,oneof=image video voice"`
	ImageUrls     interface{} `json:"image_urls"`
	VideoURL      string      `json:"video_url" binding:"omitempty,max=255"`
	VideoCover    string      `json:"video_cover" binding:"omitempty,max=255"`
	VideoDuration int         `json:"video_duration"`
	VoiceURL      string      `json:"voice_url" binding:"omitempty,max=255"`
	VoiceDuration int         `json:"voice_duration"`
	Location      string      `json:"location" binding:"max=128"`
	Latitude      float64     `json:"latitude"`
	Longitude     float64     `json:"longitude"`
	Tags          interface{} `json:"tags"`
	Topic         string      `json:"topic" binding:"max=64"`
}

// UpdateLoveStoryRequest 更新动态请求
type UpdateLoveStoryRequest struct {
	Title         *string `json:"title" binding:"omitempty,max=200"`
	Content       *string `json:"content" binding:"omitempty,max=5000"`
	ImageUrls     interface{} `json:"image_urls"`
	VideoURL      *string `json:"video_url" binding:"omitempty,max=255"`
	VideoCover    *string `json:"video_cover" binding:"omitempty,max=255"`
	VideoDuration *int    `json:"video_duration"`
	VoiceURL      *string `json:"voice_url" binding:"omitempty,max=255"`
	VoiceDuration *int    `json:"voice_duration"`
	Location      *string `json:"location" binding:"omitempty,max=128"`
	Latitude      *float64 `json:"latitude"`
	Longitude     *float64 `json:"longitude"`
	Tags          interface{} `json:"tags"`
	Topic         *string `json:"topic" binding:"omitempty,max=64"`
}

// LoveStoryListRequest 动态列表请求
type LoveStoryListRequest struct {
	Keyword    string `form:"keyword" json:"keyword"`
	LoveID     uint   `form:"love_id" json:"love_id"`
	UserID     uint   `form:"user_id" json:"user_id"`
	MediaType  string `form:"media_type" json:"media_type"`
	Topic      string `form:"topic" json:"topic"`
	Status     *int   `form:"status" json:"status"`
	AuditStatus *int  `form:"audit_status" json:"audit_status"`
	Featured   *bool  `form:"featured" json:"featured"`
	Sort       string `form:"sort" json:"sort"` // latest/hot/likes
	utils.Pagination
}

// LoveStoryLikeRequest 动态点赞请求
type LoveStoryLikeRequest struct {
	StoryID uint `json:"story_id" binding:"required"`
}

// LoveStoryCommentRequest 动态评论请求
type LoveStoryCommentRequest struct {
	Content string `json:"content" binding:"required,max=500"`
}
