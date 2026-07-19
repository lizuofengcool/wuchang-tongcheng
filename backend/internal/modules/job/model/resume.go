// Package model 简历表（对标 BOSS直聘）
// 教育/工作/项目/技能/期望 + 多 JSONB 字段
package model

import (
	"time"

	"wuchang-tongcheng/internal/pkg/database"
)

// === 简历状态常量 ===
const (
	ResumeStatusDraft     = 0 // 草稿
	ResumeStatusPublished = 1 // 已发布
	ResumeStatusHidden    = 2 // 已隐藏
	ResumeStatusDeleted   = 3 // 已删除
)

// 注：性别常量 GenderUnlimited/GenderMale/GenderFemale 已在 job.go 中定义，此处复用

// JobResume 简历表
type JobResume struct {
	database.RegionBaseModel // id/region_id/created_at/updated_at/deleted_at
	UserID              uint       `gorm:"not null;index" json:"user_id"`                          // 用户 ID
	Name                string     `gorm:"size:50" json:"name"`                                    // 真实姓名
	Gender              string     `gorm:"size:16;default:'unlimited'" json:"gender"`             // 性别
	BirthDate           *time.Time `gorm:"type:date" json:"birth_date"`                            // 出生日期
	Phone               string     `gorm:"size:20" json:"phone"`                                   // 手机
	Email               string     `gorm:"size:128" json:"email"`                                  // 邮箱
	Avatar              string     `gorm:"size:255" json:"avatar"`                                 // 头像
	EducationLevel      string     `gorm:"size:32;default:'unlimited';index" json:"education_level"` // 最高学历
	School              string     `gorm:"size:128" json:"school"`                                 // 毕业学校
	Major               string     `gorm:"size:128" json:"major"`                                 // 所学专业
	GraduateDate        *time.Time `gorm:"type:date" json:"graduate_date"`                         // 毕业时间
	WorkYears           int        `gorm:"default:0;index" json:"work_years"`                       // 工作年限
	CurrentCompany      string     `gorm:"size:128" json:"current_company"`                       // 当前公司
	CurrentPosition     string     `gorm:"size:64" json:"current_position"`                         // 当前职位
	CurrentSalary       float64    `gorm:"type:decimal(12,2);default:0" json:"current_salary"`     // 当前薪资
	ExpectSalaryMin     float64    `gorm:"type:decimal(12,2);default:0" json:"expect_salary_min"` // 期望薪资下限
	ExpectSalaryMax     float64    `gorm:"type:decimal(12,2);default:0" json:"expect_salary_max"` // 期望薪资上限
	ExpectCity          string     `gorm:"size:64;index" json:"expect_city"`                       // 期望城市
	ExpectPosition      string     `gorm:"size:128;index" json:"expect_position"`                 // 期望职位
	ExpectIndustry      string     `gorm:"size:64" json:"expect_industry"`                          // 期望行业
	ExpectJobType       string     `gorm:"size:32;default:'full_time';index" json:"expect_job_type"` // 期望工作类型
	ExpectEmploymentType string    `gorm:"size:32;default:'regular'" json:"expect_employment_type"` // 期望雇佣方式
	Status              int        `gorm:"default:1;index" json:"status"`                            // 0草稿 1已发布 2隐藏 3删除
	Completeness        int        `gorm:"default:0" json:"completeness"`                          // 完整度 0-100
	IsPublic            bool       `gorm:"default:true;index" json:"is_public"`                    // 是否公开
	IsDefault           bool       `gorm:"default:false;index" json:"is_default"`                  // 是否默认简历
	ViewCount           int        `gorm:"default:0" json:"view_count"`                            // 浏览数
	DeliverCount        int        `gorm:"default:0" json:"deliver_count"`                          // 投递数
	InterviewCount      int        `gorm:"default:0" json:"interview_count"`                        // 面试数
	OfferCount          int        `gorm:"default:0" json:"offer_count"`                            // Offer 数
	SelfIntroduction    string     `gorm:"type:text" json:"self_introduction"`                      // 自我介绍
	Advantage           string     `gorm:"type:text" json:"advantage"`                              // 个人优势
	Disadvantage        string     `gorm:"type:text" json:"disadvantage"`                          // 个人劣势
	Attachments         JSONB      `gorm:"type:jsonb" json:"attachments"`                          // 附件 [{type,name,url,size}]
	Educations          JSONB      `gorm:"type:jsonb" json:"educations"`                            // 教育经历 [{school,major,degree,...}]
	WorkExperiences     JSONB      `gorm:"type:jsonb" json:"work_experiences"`                     // 工作经历 [{company,position,...}]
	Projects            JSONB      `gorm:"type:jsonb" json:"projects"`                              // 项目经历 [{name,role,...}]
	Skills              JSONB      `gorm:"type:jsonb" json:"skills"`                                // 技能 [{skill_id,name,level,years}]
	Certificates        JSONB      `gorm:"type:jsonb" json:"certificates"`                          // 证书 [{name,issuer,...}]
	Languages           JSONB      `gorm:"type:jsonb" json:"languages"`                            // 语言能力 [{language,level,...}]
	Tags                JSONB      `gorm:"type:jsonb" json:"tags"`                                  // 标签
}

// TableName 表名（job_ 前缀）
func (JobResume) TableName() string { return "job_resumes" }
