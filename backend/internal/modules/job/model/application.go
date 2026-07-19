// Package model 投递记录 + 投递附加项（对标 BOSS直聘）
// 投递状态机：0已投递/1已读/2不合适/3面试邀约/4面试中/5Offer/6已入职/7已撤回/8已过期
package model

import (
	"time"

	"wuchang-tongcheng/internal/pkg/database"
)

// === 投递状态常量（9 状态机，对标 BOSS直聘） ===
const (
	ApplicationStatusDelivered   = 0 // 已投递
	ApplicationStatusRead        = 1 // 已读
	ApplicationStatusUnsuitable  = 2 // 不合适
	ApplicationStatusInterview   = 3 // 面试邀约
	ApplicationStatusInterviewing = 4 // 面试中
	ApplicationStatusOffer       = 5 // 已发 Offer
	ApplicationStatusOnboarded   = 6 // 已入职
	ApplicationStatusWithdrawn   = 7 // 已撤回
	ApplicationStatusExpired     = 8 // 已过期
)

// === 投递来源常量 ===
const (
	ApplicationSourceProactive = "proactive" // 主动投递
	ApplicationSourceRecruiter = "recruiter" // 招聘者邀约
	ApplicationSourceRecommend = "recommend" // 系统推荐
	ApplicationSourceSearch    = "search"     // 搜索投递
	ApplicationSourceAd        = "ad"         // 广告位
)

// JobApplication 投递记录表
type JobApplication struct {
	database.RegionBaseModel // id/region_id/created_at/updated_at/deleted_at
	ApplicationNo    string     `gorm:"size:64;not null;uniqueIndex" json:"application_no"`      // 投递单号（业务唯一）
	JobID            uint       `gorm:"not null;index" json:"job_id"`                            // 关联职位 ID
	ResumeID         uint       `gorm:"not null;index" json:"resume_id"`                         // 关联简历 ID
	ApplicantID      uint       `gorm:"not null;index:idx_job_app_applicant" json:"applicant_id"` // 投递人 ID
	RecruiterID      uint       `gorm:"not null;index:idx_job_app_recruiter" json:"recruiter_id"` // 招聘者 ID
	CompanyID        uint       `gorm:"index" json:"company_id"`                                  // 所属公司 ID
	PositionName     string     `gorm:"size:128" json:"position_name"`                          // 职位名快照
	PositionSnapshot JSONB      `gorm:"type:jsonb" json:"position_snapshot"`                     // 职位快照（投递时的状态）
	ResumeSnapshot   JSONB      `gorm:"type:jsonb" json:"resume_snapshot"`                       // 简历快照
	Status           int        `gorm:"default:0;index:idx_job_app_applicant;index:idx_job_app_recruiter" json:"status"` // 9 状态机
	Source           string     `gorm:"size:32;default:'proactive';index" json:"source"`        // 投递来源
	CoverLetter      string     `gorm:"type:text" json:"cover_letter"`                            // 求职信/附言
	Attachments      JSONB      `gorm:"type:jsonb" json:"attachments"`                            // 附件 [{type,name,url,size}]
	ReadAt          *time.Time `gorm:"index" json:"read_at"`                                    // 招聘者已读时间
	RepliedAt       *time.Time `gorm:"index" json:"replied_at"`                                 // 招聘者回复时间
	InterviewCount   int        `gorm:"default:0" json:"interview_count"`                       // 面试次数
	OfferAt         *time.Time `gorm:"index" json:"offer_at"`                                    // Offer 发出时间
	OfferAmount     float64    `gorm:"type:decimal(12,2);default:0" json:"offer_amount"`       // Offer 金额
	RejectedReason   string     `gorm:"size:500" json:"rejected_reason"`                        // 拒绝原因
	RejectedAt      *time.Time `gorm:"index" json:"rejected_at"`                                // 拒绝时间
	WithdrawnAt     *time.Time `gorm:"index" json:"withdrawn_at"`                              // 撤回时间
	WithdrawnReason string     `gorm:"size:500" json:"withdrawn_reason"`                       // 撤回原因
	CompletedAt     *time.Time `gorm:"index" json:"completed_at"`                              // 完成时间
	ExpiredAt       *time.Time `gorm:"index" json:"expired_at"`                                 // 过期时间
	SLADeadline     *time.Time `gorm:"index" json:"sla_deadline"`                              // SLA 截止时间
}

// TableName 表名（job_ 前缀）
func (JobApplication) TableName() string { return "job_applications" }
