// Package dto love 相亲交友数据传输对象 - 认证
package dto

import (
	"time"

	"wuchang-tongcheng/internal/pkg/utils"
)

// LoveVerificationInfo 认证记录响应
type LoveVerificationInfo struct {
	ID              uint       `json:"id"`
	VerifyNo        string     `json:"verify_no"`
	LoveID          uint       `json:"love_id"`
	UserID          uint       `json:"user_id"`
	Type            string     `json:"type"`
	TypeText        string     `json:"type_text"`
	RealName        string     `json:"real_name"`
	IdCardMasked    string     `json:"id_card_masked"`
	IdCardFront     string     `json:"id_card_front"`
	IdCardBack      string     `json:"id_card_back"`
	IdCardHold      string     `json:"id_card_hold"`
	FaceImage       string     `json:"face_image"`
	VideoURL        string     `json:"video_url"`
	VideoCover      string     `json:"video_cover"`
	VideoDuration   int        `json:"video_duration"`
	SchoolName      string     `json:"school_name"`
	DiplomaImage    string     `json:"diploma_image"`
	DiplomaNo       string     `json:"diploma_no"`
	Education       string     `json:"education"`
	GraduationYear  int        `json:"graduation_year"`
	PropertyImage   string     `json:"property_image"`
	PropertyNo      string     `json:"property_no"`
	ThirdPartyScore float64    `json:"third_party_score"`
	Status          int        `json:"status"`
	StatusText      string     `json:"status_text"`
	RejectReason    string     `json:"reject_reason"`
	VerifiedAt      *time.Time `json:"verified_at"`
	VerifiedName    string     `json:"verified_name"`
	ExpiredAt       *time.Time `json:"expired_at"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

// CreateLoveVerificationRequest 提交认证请求
type CreateLoveVerificationRequest struct {
	Type          string `json:"type" binding:"required,oneof=real_name photo video education property car"`
	RealName      string `json:"real_name" binding:"omitempty,max=64"`
	IdCardNo      string `json:"id_card_no" binding:"omitempty,max=32"`
	IdCardFront   string `json:"id_card_front" binding:"omitempty,max=255"`
	IdCardBack    string `json:"id_card_back" binding:"omitempty,max=255"`
	IdCardHold    string `json:"id_card_hold" binding:"omitempty,max=255"`
	FaceImage     string `json:"face_image" binding:"omitempty,max=255"`
	VideoURL      string `json:"video_url" binding:"omitempty,max=255"`
	VideoCover    string `json:"video_cover" binding:"omitempty,max=255"`
	VideoDuration int    `json:"video_duration"`
	SchoolName    string `json:"school_name" binding:"omitempty,max=128"`
	DiplomaImage  string `json:"diploma_image" binding:"omitempty,max=255"`
	DiplomaNo     string `json:"diploma_no" binding:"omitempty,max=64"`
	Education     string `json:"education" binding:"omitempty,max=32"`
	GraduationYear int   `json:"graduation_year"`
	PropertyImage string `json:"property_image" binding:"omitempty,max=255"`
	PropertyNo    string `json:"property_no" binding:"omitempty,max=64"`
}

// LoveVerificationListRequest 认证列表请求
type LoveVerificationListRequest struct {
	UserID uint   `form:"user_id" json:"user_id"`
	LoveID uint   `form:"love_id" json:"love_id"`
	Type   string `form:"type" json:"type"`
	Status *int   `form:"status" json:"status"`
	utils.Pagination
}

// LoveVerificationAuditRequest 认证审核请求
type LoveVerificationAuditRequest struct {
	ID           uint   `json:"id" binding:"required"`
	Status       int    `json:"status" binding:"oneof=1 2"` // 1通过 2拒绝
	RejectReason string `json:"reject_reason" binding:"max=500"`
}
