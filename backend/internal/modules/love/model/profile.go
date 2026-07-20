// Package model love 相亲交友数据模型 - 详细资料表 LoveProfile
// 对标 Soul/陌陌：扩展补充信息（自我介绍/择偶要求/语音视频）
// 依据需求文档 1.10：4 维数据隔离（region_id + user_id）
package model

import (
	"time"

	"wuchang-tongcheng/internal/pkg/database"
)

// LoveProfile 用户详细资料表
// 一对一关联 Loves 主表，承载完整择偶资料与匹配偏好
type LoveProfile struct {
	database.RegionBaseModel

	// 关联
	LoveID uint `gorm:"index;not null;uniqueIndex" json:"love_id"` // 主表 LoveID
	UserID uint `gorm:"index;not null" json:"user_id"`              // 用户 ID

	// 冗余基础信息（避免连表查询）
	Nickname string `gorm:"size:64;not null;default:''" json:"nickname"`
	Avatar   string `gorm:"size:255;not null;default:''" json:"avatar"`
	Gender   int    `gorm:"not null;default:0;index" json:"gender"`
	Age      int    `gorm:"not null;default:0;index" json:"age"`
	Height   int    `gorm:"not null;default:0" json:"height"`
	Weight   int    `gorm:"not null;default:0" json:"weight"`

	// 工作生活
	City        string `gorm:"size:64;not null;default:'';index" json:"city"`
	District    string `gorm:"size:64;not null;default:''" json:"district"`
	Occupation  string `gorm:"size:64;not null;default:''" json:"occupation"`
	Company     string `gorm:"size:128;not null;default:''" json:"company"`
	Industry    string `gorm:"size:64;not null;default:''" json:"industry"`
	Education   string `gorm:"size:32;not null;default:''" json:"education"`
	School      string `gorm:"size:128;not null;default:''" json:"school"`
	Income      string `gorm:"size:32;not null;default:''" json:"income"`
	Marriage    string `gorm:"size:32;not null;default:''" json:"marriage"`
	ChildrenStatus string `gorm:"size:32;not null;default:''" json:"children_status"`
	HouseStatus string `gorm:"size:32;not null;default:''" json:"house_status"`
	CarStatus   string `gorm:"size:32;not null;default:''" json:"car_status"`
	Drinking    string `gorm:"size:32;not null;default:''" json:"drinking"`
	Smoking     string `gorm:"size:32;not null;default:''" json:"smoking"`
	Exercise    string `gorm:"size:32;not null;default:''" json:"exercise"`
	Diet        string `gorm:"size:32;not null;default:''" json:"diet"`
	Sleep       string `gorm:"size:32;not null;default:''" json:"sleep"`
	Pets        string `gorm:"size:32;not null;default:''" json:"pets"`

	// 多值字段（JSONB）
	Languages  JSONB `gorm:"type:jsonb" json:"languages"`
	Interests  JSONB `gorm:"type:jsonb" json:"interests"`
	Skills     JSONB `gorm:"type:jsonb" json:"skills"`

	// 自我介绍 / 择偶要求
	SelfIntro     string `gorm:"type:text" json:"self_intro"`
	IdealPartner  string `gorm:"type:text" json:"ideal_partner"`
	IdealAgeMin   int    `gorm:"not null;default:0" json:"ideal_age_min"`
	IdealAgeMax   int    `gorm:"not null;default:0" json:"ideal_age_max"`
	IdealHeightMin int   `gorm:"not null;default:0" json:"ideal_height_min"`
	IdealHeightMax int   `gorm:"not null;default:0" json:"ideal_height_max"`

	// 择偶偏好 JSONB
	IdealCities     JSONB `gorm:"type:jsonb" json:"ideal_cities"`
	IdealEducation  JSONB `gorm:"type:jsonb" json:"ideal_education"`
	IdealIncome     string `gorm:"size:32;not null;default:''" json:"ideal_income"`
	IdealMarriage   string `gorm:"size:32;not null;default:''" json:"ideal_marriage"`
	IdealHouse      string `gorm:"size:32;not null;default:''" json:"ideal_house"`
	IdealCar        string `gorm:"size:32;not null;default:''" json:"ideal_car"`
	IdealSmoking    string `gorm:"size:32;not null;default:''" json:"ideal_smoking"`
	IdealDrinking   string `gorm:"size:32;not null;default:''" json:"ideal_drinking"`

	// 多媒体介绍
	VoiceIntroURL      string `gorm:"size:255;not null;default:''" json:"voice_intro_url"`
	VoiceIntroDuration int    `gorm:"not null;default:0" json:"voice_intro_duration"`
	VideoIntroURL      string `gorm:"size:255;not null;default:''" json:"video_intro_url"`
	VideoCover         string `gorm:"size:255;not null;default:''" json:"video_cover"`

	// 相册
	PhotoUrls   JSONB `gorm:"type:jsonb" json:"photo_urls"`
	PhotoCount  int   `gorm:"not null;default:0" json:"photo_count"`

	// 完整度
	ProfileScore  int        `gorm:"not null;default:0" json:"profile_score"`
	CompletedStep int        `gorm:"not null;default:0" json:"completed_step"`
	CompletedAt   *time.Time `json:"completed_at"`
	Status        int        `gorm:"not null;default:1;index" json:"status"`
}

// TableName 表名
func (LoveProfile) TableName() string { return "love_profiles" }
