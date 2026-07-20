// Package dto 同城零工兼职数据传输对象 - 岗位主表
package dto

import (
	"time"

	"wuchang-tongcheng/internal/pkg/utils"
)

// LinggongInfo 岗位详情响应
type LinggongInfo struct {
	ID                uint       `json:"id"`
	Title             string     `json:"title"`
	Content           string     `json:"content"`
	CoverImage        string     `json:"cover_image"`
	UserID            uint       `json:"user_id"`
	UserName          string     `json:"user_name"`
	UserPhone         string     `json:"user_phone"`
	UserAvatar        string     `json:"user_avatar"`
	Status            int        `json:"status"`
	StatusText        string     `json:"status_text"`
	AuditStatus       int        `json:"audit_status"`
	AuditStatusText   string     `json:"audit_status_text"`
	AuditReason       string     `json:"audit_reason"`
	PublishedAt       *time.Time `json:"published_at"`
	LinggongType      string     `json:"linggong_type"`
	LinggongTypeText  string     `json:"linggong_type_text"`
	PublisherType     string     `json:"publisher_type"`
	PublisherTypeText string     `json:"publisher_type_text"`
	EmployerID        uint       `json:"employer_id"`
	CompanyName       string     `json:"company_name"`
	ContactName       string     `json:"contact_name"`
	ContactPhone      string     `json:"contact_phone"`
	ContactWechat     string     `json:"contact_wechat"`
	BillingType       string     `json:"billing_type"`
	BillingTypeText   string     `json:"billing_type_text"`
	SalaryMin         float64    `json:"salary_min"`
	SalaryMax         float64    `json:"salary_max"`
	SalaryUnit        string     `json:"salary_unit"`
	SalaryNegotiable  bool       `json:"salary_negotiable"`
	Settlement        string     `json:"settlement"`
	SettlementText    string     `json:"settlement_text"`
	Currency          string     `json:"currency"`
	WorkStartDate     *time.Time `json:"work_start_date"`
	WorkEndDate       *time.Time `json:"work_end_date"`
	WorkDays          int        `json:"work_days"`
	WorkHours         int        `json:"work_hours"`
	WorkTimeStart     string     `json:"work_time_start"`
	WorkTimeEnd       string     `json:"work_time_end"`
	WorkWeekdays       string     `json:"work_weekdays"`
	WorkIntensity     string     `json:"work_intensity"`
	RecruitCount      int        `json:"recruit_count"`
	AppliedCount      int        `json:"applied_count"`
	ConfirmedCount    int        `json:"confirmed_count"`
	NeedGender        string     `json:"need_gender"`
	MinAge            int        `json:"min_age"`
	MaxAge            int        `json:"max_age"`
	Education         string     `json:"education"`
	Experience        string     `json:"experience"`
	NeedHealthCert     bool       `json:"need_health_cert"`
	NeedIDCard         bool       `json:"need_id_card"`
	MinCreditScore    int        `json:"min_credit_score"`
	Province          string     `json:"province"`
	City              string     `json:"city"`
	District          string     `json:"district"`
	BusinessDistrict  string     `json:"business_district"`
	Address           string     `json:"address"`
	Latitude          float64    `json:"latitude"`
	Longitude         float64    `json:"longitude"`
	WorkLocationType  string     `json:"work_location_type"`
	TaskID            uint       `json:"task_id"`
	TotalTaskCount    int        `json:"total_task_count"`
	ClaimedCount      int        `json:"claimed_count"`
	CompletedTaskCount int       `json:"completed_task_count"`
	ViewCount         int        `json:"view_count"`
	FavCount          int        `json:"fav_count"`
	ContactCount      int        `json:"contact_count"`
	ShareCount        int        `json:"share_count"`
	ApplicationCount  int        `json:"application_count"`
	LastAppliedAt     *time.Time `json:"last_applied_at"`
	ContentHash       string     `json:"content_hash"`
	RiskScore         int        `json:"risk_score"`
	VideoURL          string     `json:"video_url"`
	VideoCover        string     `json:"video_cover"`
	Features          interface{} `json:"features"`
	Tags              interface{} `json:"tags"`
	SkillTags         interface{} `json:"skill_tags"`
	WelfareTags       interface{} `json:"welfare_tags"`
	Images            interface{} `json:"images"`
	Requirements      interface{} `json:"requirements"`
	Featured          bool       `json:"featured"`
	Picked            bool       `json:"picked"`
	Verified          bool       `json:"verified"`
	PromotionLevel    int        `json:"promotion_level"`
	TrafficWeight     float64    `json:"traffic_weight"`
	EmployerVerified  bool       `json:"employer_verified"`
	EmployerVerifiedAt *time.Time `json:"employer_verified_at"`
	RegionID          uint       `json:"region_id"`
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`
}

// CreateLinggongRequest 创建岗位请求
type CreateLinggongRequest struct {
	Title             string     `json:"title" binding:"required,max=200"`
	Content           string     `json:"content"`
	CoverImage        string     `json:"cover_image"`
	LinggongType      string     `json:"linggong_type" binding:"omitempty,oneof=short_term long_term task hourly daily temp"`
	PublisherType     string     `json:"publisher_type" binding:"omitempty,oneof=personal company agent headhunter"`
	EmployerID        uint       `json:"employer_id"`
	CompanyName       string     `json:"company_name" binding:"max=128"`
	ContactName       string     `json:"contact_name" binding:"max=50"`
	ContactPhone      string     `json:"contact_phone" binding:"max=20"`
	ContactWechat     string     `json:"contact_wechat" binding:"max=64"`
	BillingType       string     `json:"billing_type" binding:"omitempty,oneof=by_piece by_hour by_day by_week by_month fixed negotiable"`
	SalaryMin         float64    `json:"salary_min" binding:"min=0"`
	SalaryMax         float64    `json:"salary_max" binding:"min=0"`
	SalaryUnit        string     `json:"salary_unit" binding:"max=32"`
	SalaryNegotiable  bool       `json:"salary_negotiable"`
	Settlement        string     `json:"settlement" binding:"omitempty,oneof=T+0 T+1 T+3 T+7 M+1 project"`
	WorkStartDate     *time.Time `json:"work_start_date" time_format:"2006-01-02"`
	WorkEndDate       *time.Time `json:"work_end_date" time_format:"2006-01-02"`
	WorkDays          int        `json:"work_days"`
	WorkHours         int        `json:"work_hours"`
	WorkTimeStart     string     `json:"work_time_start" binding:"max=16"`
	WorkTimeEnd       string     `json:"work_time_end" binding:"max=16"`
	WorkWeekdays       string     `json:"work_weekdays" binding:"max=32"`
	WorkIntensity     string     `json:"work_intensity" binding:"omitempty,oneof=light medium heavy extreme"`
	RecruitCount      int        `json:"recruit_count" binding:"min=1"`
	NeedGender        string     `json:"need_gender" binding:"omitempty,oneof=any male female"`
	MinAge            int        `json:"min_age"`
	MaxAge            int        `json:"max_age"`
	Education         string     `json:"education" binding:"max=32"`
	Experience        string     `json:"experience" binding:"max=64"`
	NeedHealthCert     bool       `json:"need_health_cert"`
	NeedIDCard         bool       `json:"need_id_card"`
	MinCreditScore    int        `json:"min_credit_score"`
	Province          string     `json:"province" binding:"max=64"`
	City              string     `json:"city" binding:"max=64"`
	District          string     `json:"district" binding:"max=64"`
	BusinessDistrict  string     `json:"business_district" binding:"max=128"`
	Address           string     `json:"address" binding:"max=500"`
	Latitude          float64    `json:"latitude"`
	Longitude         float64    `json:"longitude"`
	WorkLocationType  string     `json:"work_location_type" binding:"omitempty,oneof=onsite remote hybrid"`
	TaskID            uint       `json:"task_id"`
	TotalTaskCount    int        `json:"total_task_count"`
	VideoURL          string     `json:"video_url" binding:"max=255"`
	VideoCover        string     `json:"video_cover" binding:"max=255"`
	Features          interface{} `json:"features"`
	Tags              interface{} `json:"tags"`
	SkillTags         interface{} `json:"skill_tags"`
	WelfareTags       interface{} `json:"welfare_tags"`
	Images            interface{} `json:"images"`
	Requirements      interface{} `json:"requirements"`
}

// UpdateLinggongRequest 更新岗位请求
type UpdateLinggongRequest struct {
	Title             *string  `json:"title" binding:"omitempty,max=200"`
	Content           *string  `json:"content"`
	CoverImage        *string  `json:"cover_image"`
	Status            *int     `json:"status" binding:"omitempty,oneof=0 1 2 3 5 6 7"`
	CompanyName       *string  `json:"company_name" binding:"omitempty,max=128"`
	ContactName       *string  `json:"contact_name" binding:"omitempty,max=50"`
	ContactPhone      *string  `json:"contact_phone" binding:"omitempty,max=20"`
	ContactWechat     *string  `json:"contact_wechat" binding:"omitempty,max=64"`
	BillingType       *string  `json:"billing_type" binding:"omitempty,oneof=by_piece by_hour by_day by_week by_month fixed negotiable"`
	SalaryMin         *float64 `json:"salary_min" binding:"omitempty,min=0"`
	SalaryMax         *float64 `json:"salary_max" binding:"omitempty,min=0"`
	SalaryUnit        *string  `json:"salary_unit" binding:"omitempty,max=32"`
	SalaryNegotiable  *bool    `json:"salary_negotiable"`
	Settlement        *string  `json:"settlement" binding:"omitempty,oneof=T+0 T+1 T+3 T+7 M+1 project"`
	WorkStartDate     *time.Time `json:"work_start_date" time_format:"2006-01-02"`
	WorkEndDate       *time.Time `json:"work_end_date" time_format:"2006-01-02"`
	RecruitCount      *int     `json:"recruit_count" binding:"omitempty,min=1"`
	Province          *string  `json:"province" binding:"omitempty,max=64"`
	City              *string  `json:"city" binding:"omitempty,max=64"`
	District          *string  `json:"district" binding:"omitempty,max=64"`
	Address           *string  `json:"address" binding:"omitempty,max=500"`
	VideoURL          *string  `json:"video_url" binding:"omitempty,max=255"`
	Features          interface{} `json:"features"`
	Tags              interface{} `json:"tags"`
}

// LinggongListRequest 岗位列表请求
type LinggongListRequest struct {
	LinggongType   string `form:"linggong_type" json:"linggong_type"`
	PublisherType  string `form:"publisher_type" json:"publisher_type"`
	BillingType    string `form:"billing_type" json:"billing_type"`
	Settlement     string `form:"settlement" json:"settlement"`
	MinSalary      float64 `form:"min_salary" json:"min_salary"`
	MaxSalary      float64 `form:"max_salary" json:"max_salary"`
	Status         *int   `form:"status" json:"status"`
	AuditStatus    *int   `form:"audit_status" json:"audit_status"`
	EmployerID     uint   `form:"employer_id" json:"employer_id"`
	Province       string `form:"province" json:"province"`
	City           string `form:"city" json:"city"`
	District       string `form:"district" json:"district"`
	WorkLocationType string `form:"work_location_type" json:"work_location_type"`
	Featured       *bool  `form:"featured" json:"featured"`
	Picked         *bool  `form:"picked" json:"picked"`
	Verified       *bool  `form:"verified" json:"verified"`
	EmployerVerified *bool `form:"employer_verified" json:"employer_verified"`
	NeedGender     string `form:"need_gender" json:"need_gender"`
	Education      string `form:"education" json:"education"`
	Keyword        string `form:"keyword" json:"keyword"`
	Sort           string `form:"sort" json:"sort"`
	utils.Pagination
}

// LinggongAdminListRequest 管理后台岗位列表请求
type LinggongAdminListRequest struct {
	RegionID      uint   `form:"region_id" json:"region_id"`
	UserID        uint   `form:"user_id" json:"user_id"`
	EmployerID    uint   `form:"employer_id" json:"employer_id"`
	LinggongType  string `form:"linggong_type" json:"linggong_type"`
	PublisherType string `form:"publisher_type" json:"publisher_type"`
	BillingType   string `form:"billing_type" json:"billing_type"`
	Status        *int   `form:"status" json:"status"`
	AuditStatus   *int   `form:"audit_status" json:"audit_status"`
	Keyword       string `form:"keyword" json:"keyword"`
	utils.Pagination
}

// LinggongAuditRequest 岗位审核请求
type LinggongAuditRequest struct {
	AuditStatus int    `json:"audit_status" binding:"oneof=0 1 2"`
	AuditReason string `json:"audit_reason" binding:"max=500"`
}

// LinggongSearchRequest 关键词搜索请求
type LinggongSearchRequest struct {
	Keyword string `form:"keyword" json:"keyword" binding:"required"`
	utils.Pagination
}

// LinggongNearbyRequest 附近岗位请求
type LinggongNearbyRequest struct {
	Latitude  float64 `form:"latitude" json:"latitude" binding:"required"`
	Longitude float64 `form:"longitude" json:"longitude" binding:"required"`
	RadiusKm  float64 `form:"radius_km" json:"radius_km"`
	utils.Pagination
}
