// Package model 求职者档案表（对标斗米/兼职猫）
// 求职意向 + 工作经历 + 教育背景 + 技能认证
package model

import (
	"time"

	"wuchang-tongcheng/internal/pkg/database"
)

// === 求职者档案状态常量 ===
const (
	WorkerStatusActive    = 1 // 活跃
	WorkerStatusInactive = 0 // 不活跃
	WorkerStatusBanned   = 2 // 已封禁
)

// === 求职意向常量 ===
const (
	WorkerIntentFullTime = "full_time" // 全职
	WorkerIntentPartTime = "part_time" // 兼职
	WorkerIntentTemp      = "temp"      // 临时
	WorkerIntentRemote   = "remote"   // 远程
	WorkerIntentAny      = "any"      // 任意
)

// === 学历常量 ===
const (
	EducationBelowHigh = "below_high"   // 高中以下
	EducationHigh      = "high"          // 高中
	EducationCollege  = "college"       // 大专
	EducationBachelor = "bachelor"       // 本科
	EducationMaster   = "master"         // 硕士
	EducationDoctor   = "doctor"         // 博士
)

// LinggongWorker 求职者档案表
type LinggongWorker struct {
	database.RegionBaseModel // id/region_id/created_at/updated_at/deleted_at
	UserID            uint       `gorm:"not null;index;uniqueIndex:uniq_linggong_workers_user" json:"user_id"` // 用户 ID
	RealName          string     `gorm:"size:50;not null;default:''" json:"real_name"`              // 真实姓名
	RealNameVerified  bool       `gorm:"not null;default:false;index" json:"real_name_verified"`    // 实名认证
	Gender            string     `gorm:"size:16;not null;default:'unknown'" json:"gender"`          // 性别
	Birthday          *time.Time `gorm:"type:date" json:"birthday"`                                  // 生日
	Age               int        `gorm:"not null;default:0" json:"age"`                              // 年龄
	IDCard            string     `gorm:"size:32;not null;default:''" json:"id_card"`                // 身份证号
	IDCardVerified    bool       `gorm:"not null;default:false;index" json:"id_card_verified"`      // 身份证认证
	IDCardFrontURL    string     `gorm:"size:255;not null;default:''" json:"id_card_front_url"`     // 身份证正面
	IDCardBackURL     string     `gorm:"size:255;not null;default:''" json:"id_card_back_url"`       // 身份证反面
	IDCardHandURL     string     `gorm:"size:255;not null;default:''" json:"id_card_hand_url"`      // 手持身份证
	Nickname         string     `gorm:"size:50;not null;default:''" json:"nickname"`               // 昵称
	Avatar           string     `gorm:"size:255;not null;default:''" json:"avatar"`                // 头像
	Phone            string     `gorm:"size:20;not null;default:'';index" json:"phone"`            // 手机号
	Email            string     `gorm:"size:128;not null;default:''" json:"email"`                 // 邮箱
	Wechat           string     `gorm:"size:64;not null;default:''" json:"wechat"`                 // 微信
	Province         string     `gorm:"size:64;not null;default:'';index" json:"province"`          // 省
	City             string     `gorm:"size:64;not null;default:'';index" json:"city"`              // 市
	District         string     `gorm:"size:64;not null;default:''" json:"district"`              // 区/县
	Address          string     `gorm:"size:500;not null;default:''" json:"address"`              // 详细地址
	Latitude         float64    `gorm:"type:decimal(10,7);default:0" json:"latitude"`               // 纬度
	Longitude        float64    `gorm:"type:decimal(10,7);default:0" json:"longitude"`             // 经度
	Education        string     `gorm:"size:32;not null;default:''" json:"education"`              // 学历
	School           string     `gorm:"size:128;not null;default:''" json:"school"`                // 学校
	Major            string     `gorm:"size:128;not null;default:''" json:"major"`                // 专业
	GraduationYear   int        `gorm:"not null;default:0" json:"graduation_year"`                // 毕业年份
	JobIntention     string     `gorm:"size:32;not null;default:'any';index" json:"job_intention"`  // 求职意向
	ExpectedSalary   float64    `gorm:"type:decimal(12,2);default:0" json:"expected_salary"`        // 期望薪资
	AvailableTime    string     `gorm:"size:64;not null;default:''" json:"available_time"`         // 可用时间
	AvailableNow     bool       `gorm:"not null;default:false;index" json:"available_now"`          // 立即到岗
	HealthCertURL    string     `gorm:"size:255;not null;default:''" json:"health_cert_url"`         // 健康证
	HealthCertValidUntil *time.Time `gorm:"type:date" json:"health_cert_valid_until"`                // 健康证有效期
	HasCriminalRecord bool      `gorm:"not null;default:false" json:"has_criminal_record"`          // 无犯罪记录
	CriminalRecordURL string    `gorm:"size:255;not null;default:''" json:"criminal_record_url"`    // 无犯罪证明
	BankAccount      string     `gorm:"size:64;not null;default:''" json:"bank_account"`            // 银行账号
	BankName         string     `gorm:"size:64;not null;default:''" json:"bank_name"`              // 开户行
	AlipayAccount    string     `gorm:"size:128;not null;default:''" json:"alipay_account"`         // 支付宝账号
	WechatPayAccount string     `gorm:"size:128;not null;default:''" json:"wechat_pay_account"`     // 微信支付账号
	Bio              string     `gorm:"type:text" json:"bio"`                                       // 个人简介
	SkillTags        JSONB      `gorm:"type:jsonb" json:"skill_tags"`                              // 技能标签 ID 数组
	CategoryTags     JSONB      `gorm:"type:jsonb" json:"category_tags"`                            // 求职分类
	WorkExperience   JSONB      `gorm:"type:jsonb" json:"work_experience"`                          // 工作经历 JSON
	EducationHistory JSONB      `gorm:"type:jsonb" json:"education_history"`                        // 教育背景 JSON
	Portfolio        JSONB      `gorm:"type:jsonb" json:"portfolio"`                                // 作品集 JSON
	CreditScore      int        `gorm:"not null;default:100;index" json:"credit_score"`              // 信用分
	Level            int        `gorm:"not null;default:1;index" json:"level"`                     // 用户等级
	Status           int        `gorm:"default:1;index" json:"status"`                            // 0不活跃 1活跃 2封禁
	AppliedCount     int        `gorm:"not null;default:0" json:"applied_count"`                   // 已报名次数
	CompletedCount   int        `gorm:"not null;default:0" json:"completed_count"`                // 已完成单数
	TotalWorkHours   int        `gorm:"not null;default:0" json:"total_work_hours"`                // 累计工作时长
	TotalEarnings    float64    `gorm:"type:decimal(12,2);default:0" json:"total_earnings"`          // 累计收入
	AvgRating        float64    `gorm:"type:decimal(3,2);default:0" json:"avg_rating"`              // 平均评分
	RatingCount      int        `gorm:"not null;default:0" json:"rating_count"`                    // 评分次数
	PunctualityRate  float64    `gorm:"type:decimal(3,2);default:0" json:"punctuality_rate"`        // 守时率
	CompletionRate   float64    `gorm:"type:decimal(3,2);default:0" json:"completion_rate"`          // 完成率
}

// TableName 表名（linggong_ 前缀）
func (LinggongWorker) TableName() string { return "linggong_workers" }
