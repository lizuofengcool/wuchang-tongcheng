// Package model 同城招聘求职数据模型
// 依据 v3.2.1 架构方案第四章：对标 BOSS直聘/拉勾/58招聘/智联/前程无忧
// 依据需求文档 1.10：4 维数据隔离（region_id 地区隔离 + user_id 用户隔离）
// 依据需求文档 7.1：通用字段 id/region_id/created_at/updated_at/deleted_at + status + audit_status
// 依据需求文档 7.2：主表 jobs（保持兼容已发布数据）
package model

import (
	"time"

	"wuchang-tongcheng/internal/pkg/database"
)

// === 状态常量 ===
const (
	StatusDraft     = 0 // 草稿
	StatusPublished = 1 // 已发布
	StatusClosed    = 2 // 已关闭
	StatusOffline   = 3 // 已下架
	StatusExpired   = 4 // 已过期
)

// === 审核状态常量 ===
const (
	AuditPending  = 0 // 待审
	AuditApproved = 1 // 通过
	AuditRejected = 2 // 拒绝
)

// === 招聘类型常量 ===
const (
	RecruitmentTypeFullTime  = "full_time"  // 全职
	RecruitmentTypePartTime  = "part_time"  // 兼职
	RecruitmentTypeIntern    = "internship"  // 实习
	RecruitmentTypeTemp      = "temp"        // 临时
	RecruitmentTypeOutsource = "outsource"  // 外包
	RecruitmentTypeGig       = "gig"         // 零工
)

// === 雇佣方式常量 ===
const (
	EmploymentTypeRegular     = "regular"       // 正式员工
	EmploymentTypeLaborDispatch = "labor_dispatch" // 劳务派遣
	EmploymentTypeOutsourcing = "outsourcing"   // 外包
	EmploymentTypeFreelance   = "freelance"     // 自由职业
)

// === 学历要求常量 ===
const (
	EducationUnlimited   = "unlimited"   // 不限
	EducationJuniorHigh  = "junior_high"  // 初中
	EducationHighSchool  = "high_school"  // 高中
	EducationCollege     = "college"      // 大专
	EducationBachelor    = "bachelor"     // 本科
	EducationMaster      = "master"       // 硕士
	EducationPhd         = "phd"          // 博士
)

// === 薪资单位常量 ===
const (
	SalaryUnitMonth = "month" // 月薪
	SalaryUnitYear  = "year"  // 年薪
	SalaryUnitHour  = "hour"  // 时薪
	SalaryUnitDay   = "day"   // 日薪
)

// === 性别要求常量 ===
const (
	GenderUnlimited = "unlimited" // 不限
	GenderMale      = "male"       // 男
	GenderFemale    = "female"     // 女
)

// === 出差频率常量 ===
const (
	TravelNone       = "none"        // 不出差
	TravelOccasional = "occasional"  // 偶尔出差
	TravelFrequent   = "frequent"    // 频繁出差
)

// === 加班情况常量 ===
const (
	OvertimeNo        = "no"         // 不加班
	OvertimeOccasional = "occasional" // 偶尔加班
	OvertimeFrequent  = "frequent"   // 经常加班
	OvertimeUnknown   = "unknown"    // 未知
)

// Job 招聘求职主表
type Job struct {
	database.RegionBaseModel // 含 id/region_id/created_at/updated_at/deleted_at（地区隔离）

	// === 基础信息 ===
	Title   string `gorm:"size:200;not null" json:"title"`   // 职位标题
	Content string `gorm:"type:text" json:"content"`         // 职位描述
	Summary string `gorm:"size:500" json:"summary"`          // 摘要

	// === 发布者（用户隔离，依据 1.10.1） ===
	UserID     uint   `gorm:"index;not null" json:"user_id"`        // 发布者ID
	UserName   string `gorm:"size:50" json:"user_name"`             // 发布者昵称
	UserPhone  string `gorm:"size:20" json:"user_phone"`            // 发布者手机
	UserAvatar string `gorm:"size:255" json:"user_avatar"`          // 发布者头像

	// === 薪资相关 ===
	SalaryMin        float64 `gorm:"type:decimal(12,2);default:0;index" json:"salary_min"`         // 薪资下限
	SalaryMax        float64 `gorm:"type:decimal(12,2);default:0;index" json:"salary_max"`         // 薪资上限
	SalaryUnit       string  `gorm:"size:16;default:'month'" json:"salary_unit"`                   // 薪资单位 month/year/hour/day
	SalaryMonthly    float64 `gorm:"type:decimal(12,2);default:0" json:"salary_monthly"`           // 月薪展示值
	SalaryNegotiable bool    `gorm:"default:true" json:"salary_negotiable"`                         // 是否面议
	SalaryRangeID    uint    `gorm:"index" json:"salary_range_id"`                                  // 关联薪资范围配置
	ShowSalary       bool    `gorm:"default:true" json:"show_salary"`                               // 是否公开薪资

	// === 学历/经验要求 ===
	Education      string `gorm:"size:32;default:'unlimited';index" json:"education"` // 学历要求
	WorkYearMin    int    `gorm:"default:0;index" json:"work_year_min"`               // 经验下限（年），0=不限
	WorkYearMax    int    `gorm:"default:0" json:"work_year_max"`                     // 经验上限（年），0=不限
	ExperienceText string `gorm:"size:64" json:"experience_text"`                    // 经验要求展示文本

	// === 工作地点 ===
	WorkAddress         string  `gorm:"size:255" json:"work_address"`           // 详细地址
	WorkLatitude        float64 `gorm:"type:decimal(10,7)" json:"work_latitude"`  // 纬度
	WorkLongitude       float64 `gorm:"type:decimal(10,7)" json:"work_longitude"` // 经度
	WorkCity            string  `gorm:"size:64;index" json:"work_city"`         // 工作城市
	WorkDistrict        string  `gorm:"size:64" json:"work_district"`            // 工作行政区
	WorkBusinessDistrict string  `gorm:"size:128" json:"work_business_district"` // 工作商圈

	// === 招聘类型与雇用方式 ===
	RecruitmentType    string `gorm:"size:32;default:'full_time';index" json:"recruitment_type"` // 招聘类型
	EmploymentType     string `gorm:"size:32;default:'regular';index" json:"employment_type"`     // 雇佣方式
	HiringCount        int    `gorm:"default:1" json:"hiring_count"`                              // 招聘人数
	Department         string `gorm:"size:64" json:"department"`                                  // 所属部门
	PositionTemplateID uint   `gorm:"index" json:"position_template_id"`                          // 职位模板 ID
	CategoryID         uint   `gorm:"index" json:"category_id"`                                   // 职位分类 ID
	CompanyID          uint   `gorm:"index" json:"company_id"`                                     // 所属公司 ID

	// === 福利/技能/标签 ===
	Benefits    JSONB `gorm:"type:jsonb" json:"benefits"`     // 福利标签 ID 数组
	Skills      JSONB `gorm:"type:jsonb" json:"skills"`       // 技能要求 ID 数组
	Tags        JSONB `gorm:"type:jsonb" json:"tags"`         // 职位标签数组
	WelfareTags JSONB `gorm:"type:jsonb" json:"welfare_tags"` // 福利文案标签

	// === 招聘者/紧急/置顶 ===
	RecruiterID       uint       `gorm:"index" json:"recruiter_id"`            // 招聘者用户 ID
	RecruiterName     string     `gorm:"size:50" json:"recruiter_name"`        // 招聘者昵称
	RecruiterAvatar   string     `gorm:"size:255" json:"recruiter_avatar"`     // 招聘者头像
	RecruiterPosition string     `gorm:"size:64" json:"recruiter_position"`   // 招聘者职位
	IsUrgent          bool       `gorm:"default:false;index" json:"is_urgent"` // 是否紧急招聘
	UrgentExpire      *time.Time `gorm:"index" json:"urgent_expire"`           // 紧急招聘到期时间
	IsTop             bool       `gorm:"default:false;index" json:"is_top"`    // 是否置顶
	TopExpire         *time.Time `gorm:"index" json:"top_expire"`              // 置顶到期时间

	// === 应聘要求 ===
	AgeMin                int    `gorm:"default:0" json:"age_min"`                            // 年龄下限
	AgeMax                int    `gorm:"default:0" json:"age_max"`                            // 年龄上限
	GenderRequirement     string `gorm:"size:16;default:'unlimited'" json:"gender_requirement"` // 性别要求
	Major                 string `gorm:"size:128" json:"major"`                              // 专业要求
	LanguageRequirement   string `gorm:"size:64" json:"language_requirement"`                 // 语言要求
	CertificateRequirement string `gorm:"size:255" json:"certificate_requirement"`           // 证书要求
	TravelFrequency       string `gorm:"size:16;default:'none'" json:"travel_frequency"`     // 出差频率

	// === 试用期/社保/公积金 ===
	ProbationMonths         int     `gorm:"default:0" json:"probation_months"`                       // 试用期月数
	ProbationSalaryRatio    float64 `gorm:"type:decimal(3,2);default:1.00" json:"probation_salary_ratio"` // 试用期薪资比例
	HasSocialInsurance      bool    `gorm:"default:false" json:"has_social_insurance"`                // 是否五险
	HasHousingFund          bool    `gorm:"default:false" json:"has_housing_fund"`                     // 是否一金
	Allowances              JSONB   `gorm:"type:jsonb" json:"allowances"`                              // 补贴 JSON
	PromotionChannels       JSONB   `gorm:"type:jsonb" json:"promotion_channels"`                     // 晋升通道 JSON
	WorkSchedule            string  `gorm:"size:64" json:"work_schedule"`                              // 工作时间
	OvertimeStatus          string  `gorm:"size:16;default:'unknown'" json:"overtime_status"`         // 加班情况
	AllowRemote             bool    `gorm:"default:false" json:"allow_remote"`                         // 是否支持远程

	// === 联系方式/期限 ===
	ContactName         string     `gorm:"size:50" json:"contact_name"`               // 联系人姓名
	ContactPhone        string     `gorm:"size:20" json:"contact_phone"`              // 联系电话
	ContactEmail        string     `gorm:"size:128" json:"contact_email"`             // 联系邮箱
	ContactWechat       string     `gorm:"size:50" json:"contact_wechat"`             // 联系微信
	ApplicationDeadline *time.Time `gorm:"index" json:"application_deadline"`         // 招聘截止时间
	NeedBgCheck         bool       `gorm:"default:false" json:"need_bg_check"`         // 是否需要背景调查
	NeedHealthCheck     bool       `gorm:"default:false" json:"need_health_check"`     // 是否需要体检

	// === 展示控制 ===
	ExpiryTime *time.Time `gorm:"index" json:"expiry_time"` // 过期时间

	// === 互动统计 ===
	ViewCount       int `gorm:"default:0" json:"view_count"`        // 浏览数
	FavCount        int `gorm:"default:0" json:"fav_count"`         // 收藏数
	DeliverCount    int `gorm:"default:0" json:"deliver_count"`     // 投递数
	InterviewCount  int `gorm:"default:0" json:"interview_count"`   // 面试数
	OfferCount      int `gorm:"default:0" json:"offer_count"`        // Offer 数
	MessageCount    int `gorm:"default:0" json:"message_count"`     // 消息数

	// === 状态（依据 7.1：status + audit_status） ===
	Status      int        `gorm:"default:0;index" json:"status"`         // 0草稿 1已发布 2已关闭 3下架 4过期
	AuditStatus int        `gorm:"default:0;index" json:"audit_status"`    // 0待审 1通过 2拒绝
	AuditReason string     `gorm:"size:500" json:"audit_reason"`           // 审核拒绝原因
	PublishedAt *time.Time `gorm:"index" json:"published_at"`              // 发布时间

	// === 风控 ===
	ContentHash string `gorm:"size:64;index" json:"content_hash"`  // 图文指纹（MD5/SHA256）
	RiskScore   int    `gorm:"default:0;index" json:"risk_score"` // 风险评分 0-100，<30 限制发布
	SameJobID   string `gorm:"size:64;index" json:"same_job_id"`   // 同职位识别 ID

	// === 视频支持 ===
	VideoURL    string `gorm:"size:255" json:"video_url"`     // 视频 URL
	VideoCover  string `gorm:"size:255" json:"video_cover"`  // 视频封面

	// === 运营字段 ===
	Featured       bool    `gorm:"default:false;index" json:"featured"`         // 精选推荐
	Picked         bool    `gorm:"default:false;index" json:"picked"`           // 运营甄选
	Verified       bool    `gorm:"default:false;index" json:"verified"`         // 官方验真
	PromotionLevel int     `gorm:"default:0;index" json:"promotion_level"`     // 推广等级 0-10
	TrafficWeight  float64 `gorm:"type:decimal(3,2);default:1.00" json:"traffic_weight"` // 流量权重 0.00-9.99

	// Distance 仅在"附近"查询时由 SQL 计算并回填，非持久化字段（公里）
	Distance float64 `gorm:"-" json:"-"`
}

// TableName 表名（依据需求文档 7.2：主表 {module}s，保持 jobs 表名兼容已发布数据）
func (Job) TableName() string { return "jobs" }

// JobImage 职位图片子表
type JobImage struct {
	ID        uint      `gorm:"primarykey" json:"id"`
	JobID     uint      `gorm:"not null;index" json:"job_id"`     // 关联职位 ID
	URL       string    `gorm:"size:255;not null" json:"url"`     // 图片 URL
	Sort      int       `gorm:"default:0" json:"sort"`            // 排序
	CreatedAt time.Time `json:"created_at"`
}

// TableName 表名
func (JobImage) TableName() string { return "job_images" }
