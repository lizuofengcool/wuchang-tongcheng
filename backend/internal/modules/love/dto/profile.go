// Package dto love 相亲交友数据传输对象 - 详细资料
package dto

import (
	"time"

	"wuchang-tongcheng/internal/pkg/utils"
)

// LoveProfileInfo 详细资料响应
type LoveProfileInfo struct {
	ID             uint   `json:"id"`
	LoveID         uint   `json:"love_id"`
	UserID         uint   `json:"user_id"`
	Nickname       string `json:"nickname"`
	Avatar         string `json:"avatar"`
	Gender         int    `json:"gender"`
	Age            int    `json:"age"`
	Height         int    `json:"height"`
	Weight         int    `json:"weight"`
	City           string `json:"city"`
	District       string `json:"district"`
	Occupation     string `json:"occupation"`
	Company        string `json:"company"`
	Industry       string `json:"industry"`
	Education      string `json:"education"`
	School         string `json:"school"`
	Income         string `json:"income"`
	Marriage       string `json:"marriage"`
	ChildrenStatus string `json:"children_status"`
	HouseStatus    string `json:"house_status"`
	CarStatus      string `json:"car_status"`
	Drinking       string `json:"drinking"`
	Smoking        string `json:"smoking"`
	Exercise       string `json:"exercise"`
	Diet           string `json:"diet"`
	Sleep          string `json:"sleep"`
	Pets           string `json:"pets"`
	Languages      interface{} `json:"languages"`
	Interests      interface{} `json:"interests"`
	Skills         interface{} `json:"skills"`
	SelfIntro      string `json:"self_intro"`
	IdealPartner   string `json:"ideal_partner"`
	IdealAgeMin    int    `json:"ideal_age_min"`
	IdealAgeMax    int    `json:"ideal_age_max"`
	IdealHeightMin int    `json:"ideal_height_min"`
	IdealHeightMax int    `json:"ideal_height_max"`
	IdealCities    interface{} `json:"ideal_cities"`
	IdealEducation interface{} `json:"ideal_education"`
	IdealIncome    string `json:"ideal_income"`
	IdealMarriage  string `json:"ideal_marriage"`
	IdealHouse     string `json:"ideal_house"`
	IdealCar       string `json:"ideal_car"`
	IdealSmoking   string `json:"ideal_smoking"`
	IdealDrinking  string `json:"ideal_drinking"`
	VoiceIntroURL     string `json:"voice_intro_url"`
	VoiceIntroDuration int   `json:"voice_intro_duration"`
	VideoIntroURL  string `json:"video_intro_url"`
	VideoCover     string `json:"video_cover"`
	PhotoUrls      interface{} `json:"photo_urls"`
	PhotoCount     int    `json:"photo_count"`
	ProfileScore   int    `json:"profile_score"`
	CompletedStep  int    `json:"completed_step"`
	CompletedAt    *time.Time `json:"completed_at"`
	Status         int    `json:"status"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// CreateLoveProfileRequest 创建详细资料请求
type CreateLoveProfileRequest struct {
	City           string `json:"city" binding:"max=64"`
	District       string `json:"district" binding:"max=64"`
	Occupation     string `json:"occupation" binding:"max=64"`
	Company        string `json:"company" binding:"max=128"`
	Industry       string `json:"industry" binding:"max=64"`
	Education      string `json:"education" binding:"max=32"`
	School         string `json:"school" binding:"max=128"`
	Income         string `json:"income" binding:"max=32"`
	Marriage       string `json:"marriage" binding:"max=32"`
	ChildrenStatus string `json:"children_status" binding:"max=32"`
	HouseStatus    string `json:"house_status" binding:"max=32"`
	CarStatus      string `json:"car_status" binding:"max=32"`
	Drinking       string `json:"drinking" binding:"max=32"`
	Smoking        string `json:"smoking" binding:"max=32"`
	Exercise       string `json:"exercise" binding:"max=32"`
	Diet           string `json:"diet" binding:"max=32"`
	Sleep          string `json:"sleep" binding:"max=32"`
	Pets           string `json:"pets" binding:"max=32"`
	Languages      interface{} `json:"languages"`
	Interests      interface{} `json:"interests"`
	Skills         interface{} `json:"skills"`
	SelfIntro      string `json:"self_intro"`
	IdealPartner   string `json:"ideal_partner"`
	IdealAgeMin    int    `json:"ideal_age_min"`
	IdealAgeMax    int    `json:"ideal_age_max"`
	IdealHeightMin int    `json:"ideal_height_min"`
	IdealHeightMax int    `json:"ideal_height_max"`
	IdealCities    interface{} `json:"ideal_cities"`
	IdealEducation interface{} `json:"ideal_education"`
	IdealIncome    string `json:"ideal_income" binding:"max=32"`
	IdealMarriage  string `json:"ideal_marriage" binding:"max=32"`
	IdealHouse     string `json:"ideal_house" binding:"max=32"`
	IdealCar       string `json:"ideal_car" binding:"max=32"`
	IdealSmoking   string `json:"ideal_smoking" binding:"max=32"`
	IdealDrinking  string `json:"ideal_drinking" binding:"max=32"`
	VoiceIntroURL     string `json:"voice_intro_url" binding:"max=255"`
	VoiceIntroDuration int   `json:"voice_intro_duration"`
	VideoIntroURL  string `json:"video_intro_url" binding:"max=255"`
	VideoCover     string `json:"video_cover" binding:"max=255"`
	PhotoUrls      interface{} `json:"photo_urls"`
}

// UpdateLoveProfileRequest 更新详细资料请求
type UpdateLoveProfileRequest struct {
	City           *string `json:"city" binding:"omitempty,max=64"`
	District       *string `json:"district" binding:"omitempty,max=64"`
	Occupation     *string `json:"occupation" binding:"omitempty,max=64"`
	Company        *string `json:"company" binding:"omitempty,max=128"`
	Industry       *string `json:"industry" binding:"omitempty,max=64"`
	Education      *string `json:"education" binding:"omitempty,max=32"`
	School         *string `json:"school" binding:"omitempty,max=128"`
	Income         *string `json:"income" binding:"omitempty,max=32"`
	Marriage       *string `json:"marriage" binding:"omitempty,max=32"`
	ChildrenStatus *string `json:"children_status" binding:"omitempty,max=32"`
	HouseStatus    *string `json:"house_status" binding:"omitempty,max=32"`
	CarStatus      *string `json:"car_status" binding:"omitempty,max=32"`
	Drinking       *string `json:"drinking" binding:"omitempty,max=32"`
	Smoking        *string `json:"smoking" binding:"omitempty,max=32"`
	Exercise       *string `json:"exercise" binding:"omitempty,max=32"`
	Diet           *string `json:"diet" binding:"omitempty,max=32"`
	Sleep          *string `json:"sleep" binding:"omitempty,max=32"`
	Pets           *string `json:"pets" binding:"omitempty,max=32"`
	Languages      interface{} `json:"languages"`
	Interests      interface{} `json:"interests"`
	Skills         interface{} `json:"skills"`
	SelfIntro      *string `json:"self_intro"`
	IdealPartner   *string `json:"ideal_partner"`
	IdealAgeMin    *int   `json:"ideal_age_min"`
	IdealAgeMax    *int   `json:"ideal_age_max"`
	IdealHeightMin *int   `json:"ideal_height_min"`
	IdealHeightMax *int   `json:"ideal_height_max"`
	IdealCities    interface{} `json:"ideal_cities"`
	IdealEducation interface{} `json:"ideal_education"`
	IdealIncome    *string `json:"ideal_income" binding:"omitempty,max=32"`
	IdealMarriage  *string `json:"ideal_marriage" binding:"omitempty,max=32"`
	IdealHouse     *string `json:"ideal_house" binding:"omitempty,max=32"`
	IdealCar       *string `json:"ideal_car" binding:"omitempty,max=32"`
	IdealSmoking   *string `json:"ideal_smoking" binding:"omitempty,max=32"`
	IdealDrinking  *string `json:"ideal_drinking" binding:"omitempty,max=32"`
	VoiceIntroURL     *string `json:"voice_intro_url" binding:"omitempty,max=255"`
	VoiceIntroDuration *int  `json:"voice_intro_duration"`
	VideoIntroURL  *string `json:"video_intro_url" binding:"omitempty,max=255"`
	VideoCover     *string `json:"video_cover" binding:"omitempty,max=255"`
	PhotoUrls      interface{} `json:"photo_urls"`
}

// LoveProfileListRequest 详细资料列表请求
type LoveProfileListRequest struct {
	City       string `form:"city" json:"city"`
	Gender     *int   `form:"gender" json:"gender"`
	MinAge     int    `form:"min_age" json:"min_age"`
	MaxAge     int    `form:"max_age" json:"max_age"`
	Education  string `form:"education" json:"education"`
	Keyword    string `form:"keyword" json:"keyword"`
	utils.Pagination
}

// UpdateProfileStepRequest 更新资料填写步骤请求
type UpdateProfileStepRequest struct {
	Step int `json:"step" binding:"min=0,max=20"`
}
