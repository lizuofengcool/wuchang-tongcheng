// Package model 报名记录表（对标斗米/兼职猫）
// 报名状态机 + 报名审核 + 报名取消/拒绝
package model

import (
	"time"

	"wuchang-tongcheng/internal/pkg/database"
)

// === 报名状态常量 ===
const (
	ApplicationStatusPending   = 0 // 待审核
	ApplicationStatusApproved  = 1 // 已通过
	ApplicationStatusRejected   = 2 // 已拒绝
	ApplicationStatusCanceled   = 3 // 已取消
	ApplicationStatusExpired    = 4 // 已过期
	ApplicationStatusConfirmed  = 5 // 已确认（已报到）
	ApplicationStatusNoShow     = 6 // 未到岗
	ApplicationStatusWorking    = 7 // 工作中
	ApplicationStatusCompleted  = 8 // 已完成
	ApplicationStatusQuit       = 9 // 已离职
	ApplicationStatusFired      = 10 // 已辞退
)

// === 报名来源常量 ===
const (
	ApplicationSourceSearch     = "search"     // 搜索
	ApplicationSourceRecommend  = "recommend"  // 推荐
	ApplicationSourceShare      = "share"      // 分享
	ApplicationSourceDirect     = "direct"     // 直接报名
	ApplicationSourceInvite     = "invite"     // 雇主邀请
	ApplicationSourceFavorite   = "favorite"   // 收藏触发
)

// === 报名方式常量 ===
const (
	ApplicationMethodOnline  = "online"  // 在线报名
	ApplicationMethodPhone   = "phone"   // 电话报名
	ApplicationMethodOnsite  = "onsite"  // 现场报名
)

// LinggongApplication 报名记录表
type LinggongApplication struct {
	database.RegionBaseModel // id/region_id/created_at/updated_at/deleted_at
	ApplicationNo   string     `gorm:"size:64;not null;uniqueIndex" json:"application_no"`         // 报名单号
	LinggongID      uint       `gorm:"not null;index" json:"linggong_id"`                          // 岗位 ID
	TaskID          uint       `gorm:"not null;default:0;index" json:"task_id"`                   // 任务包 ID
	EmployerID      uint       `gorm:"not null;index" json:"employer_id"`                          // 雇主 ID
	EmployerName    string     `gorm:"size:128;not null;default:''" json:"employer_name"`          // 雇主名称
	WorkerID       uint       `gorm:"not null;index" json:"worker_id"`                            // 求职者 ID
	WorkerName     string     `gorm:"size:50;not null;default:''" json:"worker_name"`              // 求职者姓名
	WorkerAvatar   string     `gorm:"size:255;not null;default:''" json:"worker_avatar"`           // 求职者头像
	WorkerPhone    string     `gorm:"size:20;not null;default:''" json:"worker_phone"`              // 求职者手机
	WorkerAge      int        `gorm:"not null;default:0" json:"worker_age"`                        // 求职者年龄
	WorkerGender   string     `gorm:"size:16;not null;default:'unknown'" json:"worker_gender"`     // 求职者性别
	WorkerCity     string     `gorm:"size:64;not null;default:''" json:"worker_city"`              // 求职者所在城市
	WorkerCreditScore int    `gorm:"not null;default:0" json:"worker_credit_score"`               // 求职者信用分
	WorkerProfileID uint      `gorm:"not null;default:0;index" json:"worker_profile_id"`           // 求职者档案 ID
	Source         string     `gorm:"size:32;not null;default:'direct';index" json:"source"`       // search/recommend/share/direct/invite/favorite
	Method         string     `gorm:"size:16;not null;default:'online'" json:"method"`             // online/phone/onsite
	Status         int        `gorm:"default:0;index" json:"status"`                              // 0待审 1通过 2拒绝 3取消 4过期 5确认 6未到岗 7工作中 8完成 9离职 10辞退
	AppliedCount   int        `gorm:"not null;default:1" json:"applied_count"`                    // 报名人数
	CoverLetter    string     `gorm:"type:text" json:"cover_letter"`                                // 求职信
	EmployerRemark string     `gorm:"type:text" json:"employer_remark"`                            // 雇主备注
	RejectReason   string     `gorm:"size:500;not null;default:''" json:"reject_reason"`          // 拒绝原因
	CancelReason   string     `gorm:"size:500;not null;default:''" json:"cancel_reason"`          // 取消原因
	ReviewedAt     *time.Time `gorm:"index" json:"reviewed_at"`                                    // 审核时间
	ConfirmedAt    *time.Time `gorm:"index" json:"confirmed_at"`                                  // 确认时间
	OnboardedAt    *time.Time `gorm:"index" json:"onboarded_at"`                                  // 到岗时间
	CompletedAt    *time.Time `gorm:"index" json:"completed_at"`                                  // 完成时间
	CanceledAt     *time.Time `gorm:"index" json:"canceled_at"`                                   // 取消时间
	Evaluated      bool       `gorm:"not null;default:false;index" json:"evaluated"`              // 是否已评价
	AttachmentURL  string     `gorm:"size:255;not null;default:''" json:"attachment_url"`          // 附件 URL
}

// TableName 表名（linggong_ 前缀）
func (LinggongApplication) TableName() string { return "linggong_applications" }
