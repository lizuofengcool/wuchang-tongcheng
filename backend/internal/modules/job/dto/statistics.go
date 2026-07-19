// Package dto 数据统计相关 DTO
// 依据 v3.2.1 架构方案：M 端运营数据 + C 端招聘者数据
package dto

import "time"

// StatisticsResponse 平台总览（M 端）
type StatisticsResponse struct {
	TotalJobs          int64   `json:"total_jobs"`           // 在招职位总数
	TodayNewJobs       int64   `json:"today_new_jobs"`       // 今日新增职位
	TotalCompanies     int64   `json:"total_companies"`      // 公司总数
	ActiveRecruiters   int64   `json:"active_recruiters"`    // 活跃招聘者数
	TotalResumes       int64   `json:"total_resumes"`        // 简历总数
	TotalApplications  int64   `json:"total_applications"`   // 投递总数
	TodayApplications  int64   `json:"today_applications"`   // 今日投递数
	TotalInterviews    int64   `json:"total_interviews"`     // 面试总数
	TotalOffers        int64   `json:"total_offers"`         // Offer 总数
	TotalOnboarded     int64   `json:"total_onboarded"`      // 入职总数
	AvgSalary          float64 `json:"avg_salary"`           // 平均薪资
	OfferRate          float64 `json:"offer_rate"`           // 投递→Offer 转化率
	OnboardRate        float64 `json:"onboard_rate"`         // Offer→入职转化率
	TotalReports       int64   `json:"total_reports"`        // 举报总数
	PendingReports     int64   `json:"pending_reports"`      // 待处理举报数
}

// RecruiterOverviewResponse 招聘者数据总览（C 端）
type RecruiterOverviewResponse struct {
	UserID            uint    `json:"user_id"`
	TotalJobs         int64   `json:"total_jobs"`          // 发布职位数
	ActiveJobs        int64   `json:"active_jobs"`         // 在招职位数
	TotalApplications int64   `json:"total_applications"`  // 收到投递数
	TodayApplications int64   `json:"today_applications"`  // 今日投递数
	UnreadApplications int64  `json:"unread_applications"` // 未读投递数
	TotalInterviews   int64   `json:"total_interviews"`    // 面试数
	TotalOffers       int64   `json:"total_offers"`        // 已发 Offer 数
	TotalOnboarded    int64   `json:"total_onboarded"`     // 入职数
	OfferRate         float64 `json:"offer_rate"`          // 投递→Offer 率
	OnboardRate       float64 `json:"onboard_rate"`        // Offer→入职 率
	AvgSalary         float64 `json:"avg_salary"`          // 平均薪资
	TotalFavs         int64   `json:"total_favs"`          // 职位被收藏数
	TotalViews        int64   `json:"total_views"`         // 职位被浏览数
}

// ApplicantOverviewResponse 求职者数据总览（C 端）
type ApplicantOverviewResponse struct {
	UserID            uint   `json:"user_id"`
	TotalApplications int64  `json:"total_applications"`  // 投递总数
	TodayApplications int64  `json:"today_applications"`  // 今日投递数
	TotalInterviews   int64  `json:"total_interviews"`    // 面试数
	TotalOffers       int64  `json:"total_offers"`        // 收到 Offer 数
	TotalOnboarded    int64  `json:"total_onboarded"`     // 入职数
	TotalFavs         int64  `json:"total_favs"`          // 收藏职位数
	TotalViews        int64  `json:"total_views"`         // 浏览职位数
	UnreadMessages    int64  `json:"unread_messages"`     // 未读消息数
}

// HotJobResponse 热门职位响应
type HotJobResponse struct {
	JobID         uint    `json:"job_id"`
	Title         string  `json:"title"`
	CompanyName   string  `json:"company_name"`
	WorkCity      string  `json:"work_city"`
	SalaryMin     float64 `json:"salary_min"`
	SalaryMax     float64 `json:"salary_max"`
	SalaryUnit    string  `json:"salary_unit"`
	ViewCount     int     `json:"view_count"`
	FavCount      int     `json:"fav_count"`
	DeliverCount  int     `json:"deliver_count"`
	InterviewCount int    `json:"interview_count"`
	OfferCount    int     `json:"offer_count"`
	Rank          int     `json:"rank"`
}

// SalaryTrendResponse 薪资趋势响应
type SalaryTrendResponse struct {
	CategoryID   uint      `json:"category_id"`
	CategoryName string    `json:"category_name"`
	Dates        []string  `json:"dates"`
	AvgSalaries  []float64 `json:"avg_salaries"`
	MedianSalaries []float64 `json:"median_salaries"`
}

// ConversionStatsResponse 转化漏斗统计
type ConversionStatsResponse struct {
	ImpressionCount  int64   `json:"impression_count"`
	ClickCount       int64   `json:"click_count"`
	FavCount         int64   `json:"fav_count"`
	DeliverCount     int64   `json:"deliver_count"`
	InterviewCount   int64   `json:"interview_count"`
	OfferCount       int64   `json:"offer_count"`
	OnboardingCount  int64   `json:"onboarding_count"`
	ClickRate        float64 `json:"click_rate"`        // 点击率
	FavRate          float64 `json:"fav_rate"`          // 收藏率
	DeliverRate      float64 `json:"deliver_rate"`      // 投递率
	InterviewRate    float64 `json:"interview_rate"`    // 面试率
	OfferRate        float64 `json:"offer_rate"`        // Offer 率
	OnboardingRate   float64 `json:"onboarding_rate"`   // 入职率
}

// CategoryStatResponse 分类统计
type CategoryStatResponse struct {
	CategoryID       uint    `json:"category_id"`
	CategoryName     string  `json:"category_name"`
	JobCount         int64   `json:"job_count"`
	ApplicationCount int64   `json:"application_count"`
	AvgSalary        float64 `json:"avg_salary"`
}

// RegionStatResponse 地区统计
type RegionStatResponse struct {
	RegionID         uint    `json:"region_id"`
	RegionName       string  `json:"region_name"`
	JobCount         int64   `json:"job_count"`
	CompanyCount     int64   `json:"company_count"`
	ApplicationCount int64   `json:"application_count"`
	AvgSalary        float64 `json:"avg_salary"`
}

// TrendItem 趋势项
type TrendItem struct {
	Date  string `json:"date"`
	Count int64  `json:"count"`
}

// JobTrendResponse 职位趋势响应
type JobTrendResponse struct {
	Dates      []string `json:"dates"`
	NewJobs    []int64  `json:"new_jobs"`
	NewApps    []int64  `json:"new_applications"`
	NewOffers  []int64  `json:"new_offers"`
}

// DateTrendRequest 日期趋势请求
type DateTrendRequest struct {
	StartDate time.Time `form:"start_date" json:"start_date" time_format:"2006-01-02"`
	EndDate   time.Time `form:"end_date" json:"end_date" time_format:"2006-01-02"`
	RegionID  uint      `form:"region_id" json:"region_id"`
}
