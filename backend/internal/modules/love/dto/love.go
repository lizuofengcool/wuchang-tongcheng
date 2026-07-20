// Package dto love 相亲交友数据传输对象 - 主表 Love
// 依据 v3.2.1 架构方案第六章：对标 Soul / 陌陌 / 探探 / 百合网
package dto

import (
	"time"

	"wuchang-tongcheng/internal/pkg/utils"
)

// LoveInfo 用户资料详情响应
type LoveInfo struct {
	ID         uint   `json:"id"`
	UserID     uint   `json:"user_id"`
	Nickname   string `json:"nickname"`
	Avatar     string `json:"avatar"`
	Gender     int    `json:"gender"`
	GenderText string `json:"gender_text"`
	Age        int    `json:"age"`
	Birthday   *time.Time `json:"birthday"`
	Height     int    `json:"height"`
	Weight     int    `json:"weight"`
	Constellation string `json:"constellation"`
	Zodiac     string `json:"zodiac"`
	Hometown   string `json:"hometown"`
	Residence  string `json:"residence"`
	Education  string `json:"education"`
	Occupation string `json:"occupation"`
	Income     string `json:"income"`
	Marriage   string `json:"marriage"`
	House      string `json:"house"`
	Car        string `json:"car"`
	Drinking   string `json:"drinking"`
	Smoking    string `json:"smoking"`
	WantKids   string `json:"want_kids"`
	Bio        string `json:"bio"`
	VoiceIntroURL string `json:"voice_intro_url"`
	CoverImage string `json:"cover_image"`

	PhotoVerified     bool `json:"photo_verified"`
	VideoVerified     bool `json:"video_verified"`
	EducationVerified bool `json:"education_verified"`
	RealNameVerified  bool `json:"real_name_verified"`

	Status      int    `json:"status"`
	StatusText  string `json:"status_text"`
	AuditStatus int    `json:"audit_status"`
	AuditStatusText string `json:"audit_status_text"`
	AuditReason string `json:"audit_reason"`

	MemberLevel     int        `json:"member_level"`
	MemberLevelText string     `json:"member_level_text"`
	MemberExpiredAt *time.Time `json:"member_expired_at"`
	Credits         float64    `json:"credits"`

	LastActiveAt *time.Time `json:"last_active_at"`
	Online       bool       `json:"online"`
	Longitude    float64    `json:"longitude,omitempty"`
	Latitude     float64    `json:"latitude,omitempty"`
	Distance     float64    `json:"distance,omitempty"`

	HideOnline   bool `json:"hide_online"`
	HideLocation bool `json:"hide_location"`
	HideAge      bool `json:"hide_age"`
	HideDistance bool `json:"hide_distance"`

	ViewCount       int     `json:"view_count"`
	LikeCount       int     `json:"like_count"`
	LikedCount      int     `json:"liked_count"`
	MatchCount      int     `json:"match_count"`
	VisitorCount    int     `json:"visitor_count"`
	StoryCount      int     `json:"story_count"`
	GiftCount       int     `json:"gift_count"`
	ImpressionCount int     `json:"impression_count"`
	PopularityScore float64 `json:"popularity_score"`

	Featured bool `json:"featured"`
	Picked   bool `json:"picked"`

	Tags             interface{} `json:"tags"`
	Interests        interface{} `json:"interests"`
	Personality      interface{} `json:"personality"`
	Values           interface{} `json:"values"`
	PhotoUrls        interface{} `json:"photo_urls"`
	MatchPreferences interface{} `json:"match_preferences"`

	HasLiked  bool `json:"has_liked"`
	HasBlocked bool `json:"has_blocked"`

	RegionID  uint      `json:"region_id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// CreateLoveRequest 注册/创建资料请求
type CreateLoveRequest struct {
	Nickname string `json:"nickname" binding:"required,max=64"`
	Avatar   string `json:"avatar" binding:"max=255"`
	Gender   int    `json:"gender" binding:"oneof=0 1 2"`
	Birthday *time.Time `json:"birthday"`
	Height   int    `json:"height" binding:"min=0,max=300"`
	Weight   int    `json:"weight" binding:"min=0,max=500"`
	Hometown string `json:"hometown" binding:"max=128"`
	Residence string `json:"residence" binding:"max=128"`
	Education string `json:"education" binding:"max=32"`
	Occupation string `json:"occupation" binding:"max=64"`
	Income   string `json:"income" binding:"max=32"`
	Marriage string `json:"marriage" binding:"max=32"`
	Bio      string `json:"bio"`
}

// UpdateLoveRequest 更新资料请求
type UpdateLoveRequest struct {
	Nickname    *string `json:"nickname" binding:"omitempty,max=64"`
	Avatar      *string `json:"avatar" binding:"omitempty,max=255"`
	Gender      *int    `json:"gender" binding:"omitempty,oneof=0 1 2"`
	Birthday    *time.Time `json:"birthday"`
	Height      *int    `json:"height" binding:"omitempty,min=0,max=300"`
	Weight      *int    `json:"weight" binding:"omitempty,min=0,max=500"`
	Constellation *string `json:"constellation" binding:"omitempty,max=16"`
	Zodiac      *string `json:"zodiac" binding:"omitempty,max=16"`
	Hometown    *string `json:"hometown" binding:"omitempty,max=128"`
	Residence   *string `json:"residence" binding:"omitempty,max=128"`
	Education   *string `json:"education" binding:"omitempty,max=32"`
	Occupation  *string `json:"occupation" binding:"omitempty,max=64"`
	Income      *string `json:"income" binding:"omitempty,max=32"`
	Marriage    *string `json:"marriage" binding:"omitempty,max=32"`
	House       *string `json:"house" binding:"omitempty,max=32"`
	Car         *string `json:"car" binding:"omitempty,max=32"`
	Drinking    *string `json:"drinking" binding:"omitempty,max=32"`
	Smoking     *string `json:"smoking" binding:"omitempty,max=32"`
	WantKids    *string `json:"want_kids" binding:"omitempty,max=32"`
	Bio         *string `json:"bio"`
	VoiceIntroURL *string `json:"voice_intro_url" binding:"omitempty,max=255"`
	CoverImage  *string `json:"cover_image" binding:"omitempty,max=255"`
}

// LoveListRequest 用户列表请求
type LoveListRequest struct {
	Keyword     string `form:"keyword" json:"keyword"`
	Gender      *int   `form:"gender" json:"gender"`
	MinAge      int    `form:"min_age" json:"min_age"`
	MaxAge      int    `form:"max_age" json:"max_age"`
	Education   string `form:"education" json:"education"`
	Residence   string `form:"residence" json:"residence"`
	Hometown    string `form:"hometown" json:"hometown"`
	MemberLevel *int   `form:"member_level" json:"member_level"`
	Status      *int   `form:"status" json:"status"`
	AuditStatus *int   `form:"audit_status" json:"audit_status"`
	Featured    *bool  `form:"featured" json:"featured"`
	Picked      *bool  `form:"picked" json:"picked"`
	Verified    *bool  `form:"verified" json:"verified"`
	Sort        string `form:"sort" json:"sort"` // latest/popular/age/active
	utils.Pagination
}

// LoveNearbyRequest 附近的人请求
type LoveNearbyRequest struct {
	Latitude  float64 `form:"latitude" json:"latitude" binding:"required,latitude"`
	Longitude float64 `form:"longitude" json:"longitude" binding:"required,longitude"`
	RadiusKm  float64 `form:"radius_km" json:"radius_km"`
	Gender    *int    `form:"gender" json:"gender"`
	MinAge    int     `form:"min_age" json:"min_age"`
	MaxAge    int     `form:"max_age" json:"max_age"`
	utils.Pagination
}

// UpdateLocationRequest 更新位置请求
type UpdateLocationRequest struct {
	Latitude  float64 `json:"latitude" binding:"required,latitude"`
	Longitude float64 `json:"longitude" binding:"required,longitude"`
}

// UpdateVoiceIntroRequest 更新语音介绍请求
type UpdateVoiceIntroRequest struct {
	VoiceIntroURL string `json:"voice_intro_url" binding:"required,max=255"`
}

// LoveMatchScoreRequest 灵魂匹配评分请求
type LoveMatchScoreRequest struct {
	TargetUserID uint `json:"target_user_id" binding:"required"`
}

// LoveMatchScoreResponse 灵魂匹配评分响应
type LoveMatchScoreResponse struct {
	TotalScore       float64 `json:"total_score"`
	InterestMatch    float64 `json:"interest_match"`
	PersonalityMatch float64 `json:"personality_match"`
	ValueMatch       float64 `json:"value_match"`
	LocationMatch    float64 `json:"location_match"`
	AgeMatch         float64 `json:"age_match"`
	Reason           string  `json:"reason"`
}
