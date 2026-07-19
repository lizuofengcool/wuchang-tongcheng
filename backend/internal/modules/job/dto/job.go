// Package dto 职位相关 DTO（发布/编辑/查询/列表/响应）
// 依据 v3.2.1 架构方案第四章：对标 BOSS直聘/拉勾/58招聘
package dto

import (
	"time"

	"wuchang-tongcheng/internal/pkg/utils"
)

// JobInfo 职位详情
type JobInfo struct {
	ID      uint   `json:"id"`
	Title   string `json:"title"`
	Content string `json:"content"`
	Summary string `json:"summary"`

	// 发布者
	UserID     uint   `json:"user_id"`
	UserName   string `json:"user_name"`
	UserPhone  string `json:"user_phone"`
	UserAvatar string `json:"user_avatar"`

	// 薪资
	SalaryMin        float64 `json:"salary_min"`
	SalaryMax        float64 `json:"salary_max"`
	SalaryUnit       string  `json:"salary_unit"`
	SalaryMonthly    float64 `json:"salary_monthly"`
	SalaryNegotiable bool    `json:"salary_negotiable"`
	SalaryRangeID    uint    `json:"salary_range_id"`
	ShowSalary       bool    `json:"show_salary"`

	// 学历/经验
	Education      string `json:"education"`
	WorkYearMin    int    `json:"work_year_min"`
	WorkYearMax    int    `json:"work_year_max"`
	ExperienceText string `json:"experience_text"`

	// 工作地点
	WorkAddress          string  `json:"work_address"`
	WorkLatitude         float64 `json:"work_latitude"`
	WorkLongitude        float64 `json:"work_longitude"`
	WorkCity             string  `json:"work_city"`
	WorkDistrict         string  `json:"work_district"`
	WorkBusinessDistrict string  `json:"work_business_district"`

	// 招聘类型
	RecruitmentType    string `json:"recruitment_type"`
	EmploymentType     string `json:"employment_type"`
	HiringCount        int    `json:"hiring_count"`
	Department         string `json:"department"`
	PositionTemplateID uint   `json:"position_template_id"`
	CategoryID         uint   `json:"category_id"`
	CompanyID          uint   `json:"company_id"`

	// 福利/技能/标签
	Benefits    []uint   `json:"benefits"`
	Skills      []uint   `json:"skills"`
	Tags        []string `json:"tags"`
	WelfareTags []string `json:"welfare_tags"`

	// 招聘者
	RecruiterID       uint       `json:"recruiter_id"`
	RecruiterName     string     `json:"recruiter_name"`
	RecruiterAvatar   string     `json:"recruiter_avatar"`
	RecruiterPosition string     `json:"recruiter_position"`
	IsUrgent          bool       `json:"is_urgent"`
	UrgentExpire      *time.Time `json:"urgent_expire"`
	IsTop             bool       `json:"is_top"`
	TopExpire         *time.Time `json:"top_expire"`

	// 应聘要求
	AgeMin                 int    `json:"age_min"`
	AgeMax                 int    `json:"age_max"`
	GenderRequirement      string `json:"gender_requirement"`
	Major                  string `json:"major"`
	LanguageRequirement    string `json:"language_requirement"`
	CertificateRequirement string `json:"certificate_requirement"`
	TravelFrequency        string `json:"travel_frequency"`

	// 试用期/社保
	ProbationMonths      int     `json:"probation_months"`
	ProbationSalaryRatio float64 `json:"probation_salary_ratio"`
	HasSocialInsurance   bool    `json:"has_social_insurance"`
	HasHousingFund       bool    `json:"has_housing_fund"`
	Allowances           map[string]interface{} `json:"allowances"`
	PromotionChannels    map[string]interface{} `json:"promotion_channels"`
	WorkSchedule         string `json:"work_schedule"`
	OvertimeStatus       string `json:"overtime_status"`
	AllowRemote          bool   `json:"allow_remote"`

	// 联系方式
	ContactName         string     `json:"contact_name"`
	ContactPhone        string     `json:"contact_phone"`
	ContactEmail        string     `json:"contact_email"`
	ContactWechat       string     `json:"contact_wechat"`
	ApplicationDeadline *time.Time `json:"application_deadline"`
	NeedBgCheck         bool       `json:"need_bg_check"`
	NeedHealthCheck     bool       `json:"need_health_check"`

	// 展示控制
	ExpiryTime *time.Time `json:"expiry_time"`

	// 互动统计
	ViewCount      int `json:"view_count"`
	FavCount       int `json:"fav_count"`
	DeliverCount   int `json:"deliver_count"`
	InterviewCount int `json:"interview_count"`
	OfferCount     int `json:"offer_count"`
	MessageCount   int `json:"message_count"`

	// 状态
	Status      int        `json:"status"`
	AuditStatus int        `json:"audit_status"`
	AuditReason string     `json:"audit_reason"`
	PublishedAt *time.Time `json:"published_at"`
	RegionID    uint       `json:"region_id"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`

	// 视频
	VideoURL   string `json:"video_url"`
	VideoCover string `json:"video_cover"`

	// 运营字段
	Featured       bool    `json:"featured"`
	Picked         bool    `json:"picked"`
	Verified       bool    `json:"verified"`
	PromotionLevel int     `json:"promotion_level"`
	TrafficWeight  float64 `json:"traffic_weight"`

	// 图片
	Images []string `json:"images"`

	// 关联冗余
	CompanyName    string `json:"company_name,omitempty"`
	CompanyLogo    string `json:"company_logo,omitempty"`
	CategoryName   string `json:"category_name,omitempty"`
	Distance       float64 `json:"distance,omitempty"` // 仅附近查询返回（公里）
	HasFaved       bool    `json:"has_faved,omitempty"`
	HasDelivered   bool    `json:"has_delivered,omitempty"`
}

// CreateJobRequest C 端发布职位请求
type CreateJobRequest struct {
	Title       string   `json:"title" binding:"required,max=200"`
	Content     string   `json:"content"`
	Summary     string   `json:"summary" binding:"max=500"`
	CategoryID  uint     `json:"category_id"`
	CompanyID   uint     `json:"company_id"`
	Images      []string `json:"images"`

	// 薪资
	SalaryMin        float64 `json:"salary_min" binding:"gte=0"`
	SalaryMax        float64 `json:"salary_max" binding:"gte=0"`
	SalaryUnit       string  `json:"salary_unit" binding:"omitempty,oneof=month year hour day"`
	SalaryNegotiable bool    `json:"salary_negotiable"`
	SalaryRangeID    uint    `json:"salary_range_id"`
	ShowSalary       bool    `json:"show_salary"`

	// 学历/经验
	Education      string `json:"education" binding:"omitempty,oneof=unlimited junior_high high_school college bachelor master phd"`
	WorkYearMin    int    `json:"work_year_min" binding:"gte=0"`
	WorkYearMax    int    `json:"work_year_max" binding:"gte=0"`
	ExperienceText string `json:"experience_text" binding:"max=64"`

	// 工作地点
	WorkAddress          string  `json:"work_address" binding:"max=255"`
	WorkLatitude         float64 `json:"work_latitude"`
	WorkLongitude        float64 `json:"work_longitude"`
	WorkCity             string  `json:"work_city" binding:"max=64"`
	WorkDistrict         string  `json:"work_district" binding:"max=64"`
	WorkBusinessDistrict string  `json:"work_business_district" binding:"max=128"`

	// 招聘类型
	RecruitmentType    string `json:"recruitment_type" binding:"omitempty,oneof=full_time part_time internship temp outsource gig"`
	EmploymentType     string `json:"employment_type" binding:"omitempty,oneof=regular labor_dispatch outsourcing freelance"`
	HiringCount        int    `json:"hiring_count" binding:"gte=1"`
	Department         string `json:"department" binding:"max=64"`
	PositionTemplateID uint   `json:"position_template_id"`

	// 福利/技能/标签
	Benefits    []uint   `json:"benefits"`
	Skills      []uint   `json:"skills"`
	Tags        []string `json:"tags"`
	WelfareTags []string `json:"welfare_tags"`

	// 招聘者
	RecruiterID       uint   `json:"recruiter_id"`
	RecruiterName     string `json:"recruiter_name" binding:"max=50"`
	RecruiterAvatar   string `json:"recruiter_avatar" binding:"max=255"`
	RecruiterPosition string `json:"recruiter_position" binding:"max=64"`
	IsUrgent          bool   `json:"is_urgent"`

	// 应聘要求
	AgeMin                 int    `json:"age_min" binding:"gte=0"`
	AgeMax                 int    `json:"age_max" binding:"gte=0"`
	GenderRequirement      string `json:"gender_requirement" binding:"omitempty,oneof=unlimited male female"`
	Major                  string `json:"major" binding:"max=128"`
	LanguageRequirement    string `json:"language_requirement" binding:"max=64"`
	CertificateRequirement string `json:"certificate_requirement" binding:"max=255"`
	TravelFrequency        string `json:"travel_frequency" binding:"omitempty,oneof=none occasional frequent"`

	// 试用期/社保
	ProbationMonths      int     `json:"probation_months" binding:"gte=0"`
	ProbationSalaryRatio float64 `json:"probation_salary_ratio" binding:"gte=0,lte=2"`
	HasSocialInsurance   bool    `json:"has_social_insurance"`
	HasHousingFund       bool    `json:"has_housing_fund"`
	Allowances           map[string]interface{} `json:"allowances"`
	PromotionChannels    map[string]interface{} `json:"promotion_channels"`
	WorkSchedule         string `json:"work_schedule" binding:"max=64"`
	OvertimeStatus       string `json:"overtime_status" binding:"omitempty,oneof=no occasional frequent unknown"`
	AllowRemote          bool   `json:"allow_remote"`

	// 联系方式
	ContactName         string     `json:"contact_name" binding:"max=50"`
	ContactPhone        string     `json:"contact_phone" binding:"max=20"`
	ContactEmail        string     `json:"contact_email" binding:"max=128"`
	ContactWechat       string     `json:"contact_wechat" binding:"max=50"`
	ApplicationDeadline *time.Time `json:"application_deadline"`
	NeedBgCheck         bool       `json:"need_bg_check"`
	NeedHealthCheck     bool       `json:"need_health_check"`

	// 视频
	VideoURL   string `json:"video_url" binding:"max=255"`
	VideoCover string `json:"video_cover" binding:"max=255"`

	// 期限/状态
	ExpireDays int `json:"expire_days"` // 过期天数（默认 30 天）
	Status     int `json:"status" binding:"oneof=0 1"`
}

// UpdateJobRequest 更新职位请求
type UpdateJobRequest struct {
	Title       string   `json:"title" binding:"max=200"`
	Content     string   `json:"content"`
	Summary     string   `json:"summary" binding:"max=500"`
	CategoryID  uint     `json:"category_id"`
	CompanyID   uint     `json:"company_id"`
	Images      []string `json:"images"`

	SalaryMin        float64 `json:"salary_min" binding:"gte=0"`
	SalaryMax        float64 `json:"salary_max" binding:"gte=0"`
	SalaryUnit       string  `json:"salary_unit" binding:"omitempty,oneof=month year hour day"`
	SalaryNegotiable *bool   `json:"salary_negotiable"`
	SalaryRangeID    uint    `json:"salary_range_id"`
	ShowSalary       *bool   `json:"show_salary"`

	Education      string `json:"education" binding:"omitempty,oneof=unlimited junior_high high_school college bachelor master phd"`
	WorkYearMin    int    `json:"work_year_min"`
	WorkYearMax    int    `json:"work_year_max"`
	ExperienceText string `json:"experience_text" binding:"max=64"`

	WorkAddress          string  `json:"work_address" binding:"max=255"`
	WorkLatitude         float64 `json:"work_latitude"`
	WorkLongitude        float64 `json:"work_longitude"`
	WorkCity             string  `json:"work_city" binding:"max=64"`
	WorkDistrict         string  `json:"work_district" binding:"max=64"`
	WorkBusinessDistrict string  `json:"work_business_district" binding:"max=128"`

	RecruitmentType    string `json:"recruitment_type" binding:"omitempty,oneof=full_time part_time internship temp outsource gig"`
	EmploymentType     string `json:"employment_type" binding:"omitempty,oneof=regular labor_dispatch outsourcing freelance"`
	HiringCount        int    `json:"hiring_count"`
	Department         string `json:"department" binding:"max=64"`
	PositionTemplateID uint   `json:"position_template_id"`

	Benefits    []uint   `json:"benefits"`
	Skills      []uint   `json:"skills"`
	Tags        []string `json:"tags"`
	WelfareTags []string `json:"welfare_tags"`

	RecruiterID       uint   `json:"recruiter_id"`
	RecruiterName     string `json:"recruiter_name" binding:"max=50"`
	RecruiterAvatar   string `json:"recruiter_avatar" binding:"max=255"`
	RecruiterPosition string `json:"recruiter_position" binding:"max=64"`
	IsUrgent          *bool  `json:"is_urgent"`

	AgeMin                 int    `json:"age_min"`
	AgeMax                 int    `json:"age_max"`
	GenderRequirement      string `json:"gender_requirement" binding:"omitempty,oneof=unlimited male female"`
	Major                  string `json:"major" binding:"max=128"`
	LanguageRequirement    string `json:"language_requirement" binding:"max=64"`
	CertificateRequirement string `json:"certificate_requirement" binding:"max=255"`
	TravelFrequency        string `json:"travel_frequency" binding:"omitempty,oneof=none occasional frequent"`

	ProbationMonths      int     `json:"probation_months"`
	ProbationSalaryRatio float64 `json:"probation_salary_ratio" binding:"gte=0,lte=2"`
	HasSocialInsurance   *bool   `json:"has_social_insurance"`
	HasHousingFund       *bool   `json:"has_housing_fund"`
	Allowances           map[string]interface{} `json:"allowances"`
	PromotionChannels    map[string]interface{} `json:"promotion_channels"`
	WorkSchedule         string `json:"work_schedule" binding:"max=64"`
	OvertimeStatus       string `json:"overtime_status" binding:"omitempty,oneof=no occasional frequent unknown"`
	AllowRemote          *bool   `json:"allow_remote"`

	ContactName         string     `json:"contact_name" binding:"max=50"`
	ContactPhone        string     `json:"contact_phone" binding:"max=20"`
	ContactEmail        string     `json:"contact_email" binding:"max=128"`
	ContactWechat       string     `json:"contact_wechat" binding:"max=50"`
	ApplicationDeadline *time.Time `json:"application_deadline"`
	NeedBgCheck         *bool      `json:"need_bg_check"`
	NeedHealthCheck     *bool      `json:"need_health_check"`

	VideoURL   string `json:"video_url" binding:"max=255"`
	VideoCover string `json:"video_cover" binding:"max=255"`

	ExpireDays int `json:"expire_days"`
	Status     int `json:"status" binding:"omitempty,oneof=0 1 2 3"`
}

// JobListRequest C 端职位列表查询
type JobListRequest struct {
	CategoryID      uint    `form:"category_id" json:"category_id"`
	Keyword         string  `form:"keyword" json:"keyword"`
	RecruitmentType string  `form:"recruitment_type" json:"recruitment_type"`
	EmploymentType  string  `form:"employment_type" json:"employment_type"`
	Education       string  `form:"education" json:"education"`
	WorkYearMin     int     `form:"work_year_min" json:"work_year_min"`
	WorkYearMax     int     `form:"work_year_max" json:"work_year_max"`
	SalaryMin       float64 `form:"salary_min" json:"salary_min"`
	SalaryMax       float64 `form:"salary_max" json:"salary_max"`
	WorkCity        string  `form:"work_city" json:"work_city"`
	CompanyID       uint    `form:"company_id" json:"company_id"`
	AllowRemote     *bool   `form:"allow_remote" json:"allow_remote"`
	IsUrgent        *bool   `form:"is_urgent" json:"is_urgent"`
	Featured        *bool   `form:"featured" json:"featured"`
	Verified        *bool   `form:"verified" json:"verified"`
	Sort            string  `form:"sort" json:"sort"` // latest/salary_asc/salary_desc/popular
	utils.Pagination
}

// JobNearbyRequest 附近职位查询
type JobNearbyRequest struct {
	Latitude  float64 `form:"latitude" binding:"required"`
	Longitude float64 `form:"longitude" binding:"required"`
	RadiusKm  float64 `form:"radius_km"`
	utils.Pagination
}

// JobSearchRequest 搜索职位（关键词）
type JobSearchRequest struct {
	Keyword string `form:"keyword" binding:"required,max=100"`
	utils.Pagination
}

// JobAdminListRequest 管理后台列表查询
type JobAdminListRequest struct {
	RegionID    uint   `form:"region_id" json:"region_id"`
	UserID      uint   `form:"user_id" json:"user_id"`
	CategoryID  uint   `form:"category_id" json:"category_id"`
	CompanyID   uint   `form:"company_id" json:"company_id"`
	Status      *int   `form:"status" json:"status"`
	AuditStatus *int   `form:"audit_status" json:"audit_status"`
	Keyword     string `form:"keyword" json:"keyword"`
	utils.Pagination
}

// UpdateJobStatusRequest 上下架请求
type UpdateJobStatusRequest struct {
	Status int `json:"status" binding:"oneof=1 2 3"` // 1发布 2关闭 3下架
}

// JobPromotionRequest 推广请求
type JobPromotionRequest struct {
	PromotionLevel int     `json:"promotion_level" binding:"gte=0,lte=10"`
	TrafficWeight  float64 `json:"traffic_weight" binding:"gte=0,lte=9.99"`
	Featured       bool    `json:"featured"`
	Picked         bool    `json:"picked"`
	Verified       bool    `json:"verified"`
	IsTop          bool    `json:"is_top"`
	TopDays        int     `json:"top_days"`     // 置顶天数
	IsUrgent       bool    `json:"is_urgent"`
	UrgentDays     int     `json:"urgent_days"`   // 紧急天数
}

// JobDetailResponse 聚合详情响应
type JobDetailResponse struct {
	JobInfo
	Company     *CompanyResponse     `json:"company,omitempty"`
	Reviews     []ReviewResponse     `json:"reviews,omitempty"`
	ReviewStats ReviewStats          `json:"review_stats"`
}
