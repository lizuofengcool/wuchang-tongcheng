// Package model 面试邀约表（对标 BOSS直聘）
// 多轮 + 在线/线下 + 反馈 + Offer
package model

import (
	"time"

	"wuchang-tongcheng/internal/pkg/database"
)

// === 面试状态常量 ===
const (
	InterviewStatusPending    = 0 // 待确认
	InterviewStatusConfirmed  = 1 // 已确认
	InterviewStatusRescheduled = 2 // 已改期
	InterviewStatusAttended   = 3 // 已参加
	InterviewStatusCompleted  = 4 // 已完成
	InterviewStatusCanceled   = 5 // 已取消
	InterviewStatusNoShow     = 6 // 未到面
)

// === 面试结果常量 ===
const (
	InterviewResultPending  = "pending"   // 待定
	InterviewResultPass     = "pass"       // 通过
	InterviewResultReject   = "reject"     // 不通过
	InterviewResultNextRound = "next_round" // 进入下一轮
	InterviewResultOffer    = "offer"      // 发放 Offer
)

// === 面试方式常量 ===
const (
	InterviewTypeOnsite = "onsite" // 现场面试
	InterviewTypeOnline = "online" // 在线视频
	InterviewTypePhone  = "phone"  // 电话面试
	InterviewTypeVideo  = "video"   // 视频面试
)

// JobInterview 面试邀约表
type JobInterview struct {
	database.RegionBaseModel // id/region_id/created_at/updated_at/deleted_at
	InterviewNo        string     `gorm:"size:64;not null;uniqueIndex" json:"interview_no"`      // 面试单号
	ApplicationID      uint       `gorm:"not null;index" json:"application_id"`                  // 关联投递记录 ID
	JobID              uint       `gorm:"not null;index" json:"job_id"`                           // 关联职位 ID
	ApplicantID       uint       `gorm:"not null;index:idx_job_interview_applicant" json:"applicant_id"` // 求职者 ID
	RecruiterID       uint       `gorm:"not null;index:idx_job_interview_recruiter" json:"recruiter_id"` // 招聘者 ID
	CompanyID          uint       `gorm:"index" json:"company_id"`                                 // 所属公司 ID
	Round              int        `gorm:"default:1" json:"round"`                                 // 面试轮次（1=初面/2=复面/3=终面...）
	InterviewType      string     `gorm:"size:32;default:'onsite';index" json:"interview_type"`   // 面试方式 onsite/online/phone/video
	ScheduledAt        *time.Time `gorm:"index" json:"scheduled_at"`                              // 计划面试时间
	DurationMinutes   int        `gorm:"default:60" json:"duration_minutes"`                     // 预计时长（分钟）
	Location          string     `gorm:"size:255" json:"location"`                                // 线下地址
	OnlineURL         string     `gorm:"size:500" json:"online_url"`                              // 在线会议链接
	OnlinePassword    string     `gorm:"size:64" json:"online_password"`                          // 会议密码
	InterviewerName   string     `gorm:"size:50" json:"interviewer_name"`                         // 面试官姓名
	InterviewerPosition string   `gorm:"size:64" json:"interviewer_position"`                    // 面试官职位
	ContactPhone      string     `gorm:"size:20" json:"contact_phone"`                            // 联系电话
	Status            int        `gorm:"default:0;index:idx_job_interview_applicant;index:idx_job_interview_recruiter" json:"status"` // 7 状态
	Result            string     `gorm:"size:32;default:'pending';index" json:"result"`         // 面试结果
	Feedback          string     `gorm:"type:text" json:"feedback"`                              // 面试反馈
	Rating            int        `gorm:"default:0" json:"rating"`                                // 面试官评分 1-5
	SalaryOffered     float64    `gorm:"type:decimal(12,2);default:0" json:"salary_offered"`     // Offer 薪资
	PositionOffered   string     `gorm:"size:128" json:"position_offered"`                       // Offer 职位
	EntryDate         *time.Time `gorm:"type:date" json:"entry_date"`                             // 入职日期
	Attachments       JSONB      `gorm:"type:jsonb" json:"attachments"`                          // 附件 [{type,name,url}]
	ConfirmedAt       *time.Time `gorm:"index" json:"confirmed_at"`                              // 求职者确认时间
	AttendedAt        *time.Time `gorm:"index" json:"attended_at"`                               // 实际到面时间
	CompletedAt       *time.Time `gorm:"index" json:"completed_at"`                              // 完成时间
	CanceledAt        *time.Time `gorm:"index" json:"canceled_at"`                               // 取消时间
	CanceledReason    string     `gorm:"size:500" json:"canceled_reason"`                       // 取消原因
}

// TableName 表名（job_ 前缀）
func (JobInterview) TableName() string { return "job_interviews" }
