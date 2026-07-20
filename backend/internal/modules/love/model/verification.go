// Package model love 相亲交友数据模型 - 实名/视频/学历认证表 LoveVerification
// 对标百合网/陌陌：三重认证（实名 + 视频 + 学历）
// 依据需求文档 1.10：4 维数据隔离（region_id + user_id）
package model

import (
	"time"

	"wuchang-tongcheng/internal/pkg/database"
)

// LoveVerification 认证表
// 一个用户可有多条认证记录（每种 type 一条）
type LoveVerification struct {
	database.RegionBaseModel

	VerifyNo string `gorm:"size:64;not null;uniqueIndex" json:"verify_no"` // 认证单号
	LoveID   uint   `gorm:"index;not null" json:"love_id"`
	UserID   uint   `gorm:"index;not null" json:"user_id"`
	Type     string `gorm:"size:32;not null;default:'real_name';index" json:"type"` // real_name/photo/video/education/property/car

	// 实名认证
	RealName    string `gorm:"size:64;not null;default:''" json:"real_name"`
	IDCardNo    string `gorm:"size:32;not null;default:''" json:"id_card_no"`
	IDCardFront string `gorm:"size:255;not null;default:''" json:"id_card_front"`
	IDCardBack  string `gorm:"size:255;not null;default:''" json:"id_card_back"`
	IDCardHold  string `gorm:"size:255;not null;default:''" json:"id_card_hold"`
	FaceImage   string `gorm:"size:255;not null;default:''" json:"face_image"`

	// 视频认证
	VideoURL      string `gorm:"size:255;not null;default:''" json:"video_url"`
	VideoCover    string `gorm:"size:255;not null;default:''" json:"video_cover"`
	VideoDuration int    `gorm:"not null;default:0" json:"video_duration"`

	// 学历认证
	SchoolName      string `gorm:"size:128;not null;default:''" json:"school_name"`
	DiplomaImage    string `gorm:"size:255;not null;default:''" json:"diploma_image"`
	DiplomaNo       string `gorm:"size:64;not null;default:''" json:"diploma_no"`
	Education       string `gorm:"size:32;not null;default:''" json:"education"`
	GraduationYear  int    `gorm:"not null;default:0" json:"graduation_year"`

	// 房产认证
	PropertyImage string `gorm:"size:255;not null;default:''" json:"property_image"`
	PropertyNo    string `gorm:"size:64;not null;default:''" json:"property_no"`

	// 第三方认证
	ThirdPartyToken string  `gorm:"size:255;not null;default:''" json:"third_party_token"`
	ThirdPartyScore float64 `gorm:"type:decimal(5,2);not null;default:0" json:"third_party_score"`

	// 状态/审核
	Status       int        `gorm:"not null;default:0;index" json:"status"` // 0待审 1通过 2拒绝
	RejectReason string     `gorm:"size:500;not null;default:''" json:"reject_reason"`
	VerifiedAt   *time.Time `gorm:"index" json:"verified_at"`
	VerifiedBy   uint       `gorm:"not null;default:0" json:"verified_by"`
	VerifiedName string     `gorm:"size:64;not null;default:''" json:"verified_name"`
	ExpiredAt    *time.Time `json:"expired_at"`
}

// TableName 表名
func (LoveVerification) TableName() string { return "love_verifications" }
