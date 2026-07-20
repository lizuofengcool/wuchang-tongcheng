// Package dto 同城零工兼职数据传输对象 - 求职者档案 + 资质证书
package dto

import (
	"time"

	"wuchang-tongcheng/internal/pkg/utils"
)

// WorkerInfo 求职者档案详情响应
type WorkerInfo struct {
	ID                  uint       `json:"id"`
	UserID              uint       `json:"user_id"`
	RealName            string     `json:"real_name"`
	RealNameVerified    bool       `json:"real_name_verified"`
	Gender              string     `json:"gender"`
	Birthday            *time.Time `json:"birthday"`
	Age                 int        `json:"age"`
	IDCard              string     `json:"id_card"`
	IDCardVerified      bool       `json:"id_card_verified"`
	IDCardFrontURL      string     `json:"id_card_front_url"`
	IDCardBackURL       string     `json:"id_card_back_url"`
	IDCardHandURL       string     `json:"id_card_hand_url"`
	Nickname            string     `json:"nickname"`
	Avatar              string     `json:"avatar"`
	Phone               string     `json:"phone"`
	Email               string     `json:"email"`
	Wechat              string     `json:"wechat"`
	Province            string     `json:"province"`
	City                string     `json:"city"`
	District            string     `json:"district"`
	Address             string     `json:"address"`
	Latitude            float64    `json:"latitude"`
	Longitude           float64    `json:"longitude"`
	Education            string     `json:"education"`
	School              string     `json:"school"`
	Major               string     `json:"major"`
	GraduationYear      int        `json:"graduation_year"`
	JobIntention        string     `json:"job_intention"`
	JobIntentionText    string     `json:"job_intention_text"`
	ExpectedSalary      float64    `json:"expected_salary"`
	AvailableTime       string     `json:"available_time"`
	AvailableNow        bool       `json:"available_now"`
	HealthCertURL       string     `json:"health_cert_url"`
	HealthCertValidUntil *time.Time `json:"health_cert_valid_until"`
	HasCriminalRecord   bool       `json:"has_criminal_record"`
	CriminalRecordURL   string     `json:"criminal_record_url"`
	BankAccount         string     `json:"bank_account"`
	BankName            string     `json:"bank_name"`
	AlipayAccount       string     `json:"alipay_account"`
	WechatPayAccount    string     `json:"wechat_pay_account"`
	Bio                 string     `json:"bio"`
	SkillTags           interface{} `json:"skill_tags"`
	CategoryTags        interface{} `json:"category_tags"`
	WorkExperience      interface{} `json:"work_experience"`
	EducationHistory    interface{} `json:"education_history"`
	Portfolio           interface{} `json:"portfolio"`
	CreditScore         int        `json:"credit_score"`
	Level               int        `json:"level"`
	Status              int        `json:"status"`
	StatusText          string     `json:"status_text"`
	AppliedCount        int        `json:"applied_count"`
	CompletedCount      int        `json:"completed_count"`
	TotalWorkHours      int        `json:"total_work_hours"`
	TotalEarnings       float64    `json:"total_earnings"`
	AvgRating           float64    `json:"avg_rating"`
	RatingCount         int        `json:"rating_count"`
	PunctualityRate     float64    `json:"punctuality_rate"`
	CompletionRate      float64    `json:"completion_rate"`
	RegionID            uint       `json:"region_id"`
	CreatedAt           time.Time  `json:"created_at"`
	UpdatedAt           time.Time  `json:"updated_at"`
}

// CreateWorkerRequest 创建求职者档案请求
type CreateWorkerRequest struct {
	RealName          string     `json:"real_name" binding:"max=50"`
	Gender            string     `json:"gender" binding:"omitempty,oneof=male female unknown"`
	Birthday          *time.Time `json:"birthday"`
	IDCard            string     `json:"id_card" binding:"max=32"`
	IDCardFrontURL    string     `json:"id_card_front_url" binding:"max=255"`
	IDCardBackURL     string     `json:"id_card_back_url" binding:"max=255"`
	IDCardHandURL     string     `json:"id_card_hand_url" binding:"max=255"`
	Nickname          string     `json:"nickname" binding:"max=50"`
	Avatar            string     `json:"avatar" binding:"max=255"`
	Phone             string     `json:"phone" binding:"max=20"`
	Email             string     `json:"email" binding:"max=128"`
	Wechat            string     `json:"wechat" binding:"max=64"`
	Province          string     `json:"province" binding:"max=64"`
	City              string     `json:"city" binding:"max=64"`
	District          string     `json:"district" binding:"max=64"`
	Address           string     `json:"address" binding:"max=500"`
	Latitude          float64    `json:"latitude"`
	Longitude         float64    `json:"longitude"`
	Education          string     `json:"education" binding:"max=32"`
	School            string     `json:"school" binding:"max=128"`
	Major             string     `json:"major" binding:"max=128"`
	GraduationYear    int        `json:"graduation_year"`
	JobIntention      string     `json:"job_intention" binding:"omitempty,oneof=full_time part_time temp remote any"`
	ExpectedSalary    float64    `json:"expected_salary"`
	AvailableTime     string     `json:"available_time" binding:"max=64"`
	AvailableNow      bool       `json:"available_now"`
	HealthCertURL     string     `json:"health_cert_url" binding:"max=255"`
	HealthCertValidUntil *time.Time `json:"health_cert_valid_until"`
	HasCriminalRecord bool       `json:"has_criminal_record"`
	CriminalRecordURL string     `json:"criminal_record_url" binding:"max=255"`
	BankAccount       string     `json:"bank_account" binding:"max=64"`
	BankName          string     `json:"bank_name" binding:"max=64"`
	AlipayAccount     string     `json:"alipay_account" binding:"max=128"`
	WechatPayAccount  string     `json:"wechat_pay_account" binding:"max=128"`
	Bio               string     `json:"bio"`
	SkillTags         interface{} `json:"skill_tags"`
	CategoryTags      interface{} `json:"category_tags"`
	WorkExperience    interface{} `json:"work_experience"`
	EducationHistory  interface{} `json:"education_history"`
	Portfolio         interface{} `json:"portfolio"`
}

// UpdateWorkerRequest 更新求职者档案请求
type UpdateWorkerRequest struct {
	RealName             *string `json:"real_name" binding:"omitempty,max=50"`
	Gender               *string `json:"gender" binding:"omitempty,oneof=male female unknown"`
	Birthday             *time.Time `json:"birthday"`
	Nickname             *string `json:"nickname" binding:"omitempty,max=50"`
	Avatar               *string `json:"avatar" binding:"omitempty,max=255"`
	Phone                *string `json:"phone" binding:"omitempty,max=20"`
	Email                *string `json:"email" binding:"omitempty,max=128"`
	Wechat               *string `json:"wechat" binding:"omitempty,max=64"`
	Province             *string `json:"province" binding:"omitempty,max=64"`
	City                 *string `json:"city" binding:"omitempty,max=64"`
	District             *string `json:"district" binding:"omitempty,max=64"`
	Address              *string `json:"address" binding:"omitempty,max=500"`
	Latitude             *float64 `json:"latitude"`
	Longitude            *float64 `json:"longitude"`
	Education            *string `json:"education" binding:"omitempty,max=32"`
	School               *string `json:"school" binding:"omitempty,max=128"`
	Major                *string `json:"major" binding:"omitempty,max=128"`
	GraduationYear       *int    `json:"graduation_year"`
	JobIntention         *string `json:"job_intention" binding:"omitempty,oneof=full_time part_time temp remote any"`
	ExpectedSalary       *float64 `json:"expected_salary"`
	AvailableTime        *string `json:"available_time" binding:"omitempty,max=64"`
	AvailableNow         *bool   `json:"available_now"`
	HealthCertURL        *string `json:"health_cert_url" binding:"omitempty,max=255"`
	HealthCertValidUntil *time.Time `json:"health_cert_valid_until"`
	BankAccount          *string `json:"bank_account" binding:"omitempty,max=64"`
	BankName             *string `json:"bank_name" binding:"omitempty,max=64"`
	AlipayAccount        *string `json:"alipay_account" binding:"omitempty,max=128"`
	WechatPayAccount     *string `json:"wechat_pay_account" binding:"omitempty,max=128"`
	Bio                  *string `json:"bio"`
	SkillTags            interface{} `json:"skill_tags"`
	CategoryTags         interface{} `json:"category_tags"`
	WorkExperience       interface{} `json:"work_experience"`
	EducationHistory     interface{} `json:"education_history"`
	Portfolio            interface{} `json:"portfolio"`
}

// WorkerListRequest 求职者列表请求
type WorkerListRequest struct {
	JobIntention string `form:"job_intention" json:"job_intention"`
	Education    string `form:"education" json:"education"`
	City         string `form:"city" json:"city"`
	AvailableNow *bool  `form:"available_now" json:"available_now"`
	SkillID      uint   `form:"skill_id" json:"skill_id"`
	MinCreditScore int  `form:"min_credit_score" json:"min_credit_score"`
	Status       *int   `form:"status" json:"status"`
	Keyword      string `form:"keyword" json:"keyword"`
	utils.Pagination
}

// WorkerAdminListRequest 管理后台求职者列表请求
type WorkerAdminListRequest struct {
	RegionID     uint   `form:"region_id" json:"region_id"`
	UserID       uint   `form:"user_id" json:"user_id"`
	Status       *int   `form:"status" json:"status"`
	JobIntention string `form:"job_intention" json:"job_intention"`
	Keyword      string `form:"keyword" json:"keyword"`
	utils.Pagination
}

// CertificationInfo 资质证书信息
type CertificationInfo struct {
	ID              uint       `json:"id"`
	CertNo          string     `json:"cert_no"`
	UserID          uint       `json:"user_id"`
	WorkerID        uint       `json:"worker_id"`
	WorkerName      string     `json:"worker_name"`
	EmployerID      uint       `json:"employer_id"`
	CertType        string     `json:"cert_type"`
	CertTypeText    string     `json:"cert_type_text"`
	CertName        string     `json:"cert_name"`
	CertCode        string     `json:"cert_code"`
	IssuerName      string     `json:"issuer_name"`
	IssuerType      string     `json:"issuer_type"`
	IssueDate       *time.Time `json:"issue_date"`
	ValidFrom       *time.Time `json:"valid_from"`
	ValidUntil      *time.Time `json:"valid_until"`
	ImageURL        string     `json:"image_url"`
	ImageBackURL    string     `json:"image_back_url"`
	SkillID         uint       `json:"skill_id"`
	SkillName       string     `json:"skill_name"`
	Level           string     `json:"level"`
	Score           float64    `json:"score"`
	Verified        bool       `json:"verified"`
	VerifiedAt      *time.Time `json:"verified_at"`
	VerifiedBy      uint       `json:"verified_by"`
	VerifiedByName  string     `json:"verified_by_name"`
	Status          int        `json:"status"`
	StatusText      string     `json:"status_text"`
	RejectReason    string     `json:"reject_reason"`
	Description     string     `json:"description"`
	RegionID        uint       `json:"region_id"`
	CreatedAt       time.Time  `json:"created_at"`
}

// CreateCertificationRequest 创建资质证书请求
type CreateCertificationRequest struct {
	CertType     string     `json:"cert_type" binding:"omitempty,oneof=id_card health_cert skill_cert education_cert work_cert driver_license language_cert profession_cert safety_cert other"`
	CertName     string     `json:"cert_name" binding:"required,max=128"`
	CertCode     string     `json:"cert_code" binding:"max=128"`
	IssuerName   string     `json:"issuer_name" binding:"max=128"`
	IssuerType   string     `json:"issuer_type" binding:"max=32"`
	IssueDate    *time.Time `json:"issue_date"`
	ValidFrom    *time.Time `json:"valid_from"`
	ValidUntil   *time.Time `json:"valid_until"`
	ImageURL     string     `json:"image_url" binding:"max=255"`
	ImageBackURL string     `json:"image_back_url" binding:"max=255"`
	SkillID      uint       `json:"skill_id"`
	SkillName    string     `json:"skill_name" binding:"max=64"`
	Level        string     `json:"level" binding:"max=32"`
	Score        float64    `json:"score"`
	Description  string     `json:"description"`
}

// CertificationListRequest 资质证书列表请求
type CertificationListRequest struct {
	UserID     uint   `form:"user_id" json:"user_id"`
	WorkerID   uint   `form:"worker_id" json:"worker_id"`
	CertType   string `form:"cert_type" json:"cert_type"`
	SkillID    uint   `form:"skill_id" json:"skill_id"`
	Status     *int   `form:"status" json:"status"`
	Verified   *bool  `form:"verified" json:"verified"`
	utils.Pagination
}

// CertVerifyRequest 证书审核请求
type CertVerifyRequest struct {
	Status       int    `json:"status" binding:"oneof=0 1 2 3 4"`
	RejectReason string `json:"reject_reason" binding:"max=500"`
}
