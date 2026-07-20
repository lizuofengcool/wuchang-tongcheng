// Package model 信用分表（对标芝麻信用/猪八戒）
// 履约 +10 / 违约 -20 / 影响接单 + 历史变更记录
package model

import (
	"time"

	"wuchang-tongcheng/internal/pkg/database"
)

// === 信用变更原因常量 ===
const (
	CreditReasonFulfill         = "fulfill"           // 履约完成
	CreditReasonBreach          = "breach"            // 违约
	CreditReasonLate            = "late"               // 迟到
	CreditReasonAbsent          = "absent"            // 缺勤
	CreditReasonNoShow          = "no_show"           // 放鸽子
	CreditReasonGoodRating      = "good_rating"      // 好评
	CreditReasonBadRating       = "bad_rating"       // 差评
	CreditReasonVerified        = "verified"         // 实名认证
	CreditReasonSkillCert       = "skill_cert"       // 技能认证
	CreditReasonReport          = "report"           // 被举报
	CreditReasonAppeal          = "appeal"           // 申诉成功
	CreditReasonManual          = "manual"           // 人工调整
	CreditReasonInviteFriend    = "invite_friend"     // 邀请好友
	CreditReasonDailyLogin      = "daily_login"       // 每日登录
	CreditReasonCompleteProfile = "complete_profile"  // 完善资料
)

// === 信用变更类型常量 ===
const (
	CreditTypeAdd     = "add"     // 加分
	CreditTypeDeduct  = "deduct"  // 扣分
	CreditTypeReset   = "reset"   // 重置
)

// === 用户类型常量 ===
const (
	CreditUserTypeWorker   = "worker"   // 求职者
	CreditUserTypeEmployer = "employer" // 雇主
)

// LinggongCredit 信用分表
type LinggongCredit struct {
	database.RegionBaseModel // id/region_id/created_at/updated_at/deleted_at
	UserID        uint       `gorm:"not null;index" json:"user_id"`                              // 用户 ID
	UserType      string     `gorm:"size:16;not null;default:'worker';index" json:"user_type"`   // worker/employer
	Reason        string     `gorm:"size:32;not null;default:'manual';index" json:"reason"`      // 履约原因
	ChangeType    string     `gorm:"size:16;not null;default:'add'" json:"change_type"`           // add/deduct/reset
	ChangeScore   int        `gorm:"not null;default:0" json:"change_score"`                      // 变更分值（正负）
	BeforeScore   int        `gorm:"not null;default:0" json:"before_score"`                       // 变更前分数
	AfterScore    int        `gorm:"not null;default:0;index" json:"after_score"`                 // 变更后分数
	LinggongID   uint       `gorm:"not null;default:0;index" json:"linggong_id"`                  // 关联岗位 ID
	TaskID       uint       `gorm:"not null;default:0" json:"task_id"`                            // 关联任务 ID
	ApplicationID uint      `gorm:"not null;default:0" json:"application_id"`                    // 关联报名 ID
	RatingID     uint       `gorm:"not null;default:0" json:"rating_id"`                          // 关联评价 ID
	OperatorID   uint       `gorm:"not null;default:0" json:"operator_id"`                       // 操作人 ID（0 表示系统）
	OperatorName string     `gorm:"size:50;not null;default:''" json:"operator_name"`             // 操作人姓名
	Description  string     `gorm:"size:500;not null;default:''" json:"description"`             // 变更说明
	EvidenceURL  string     `gorm:"size:255;not null;default:''" json:"evidence_url"`            // 凭证 URL
	CreatedAt    time.Time  `gorm:"index" json:"created_at"`                                     // 兼容字段
}

// TableName 表名（linggong_ 前缀）
func (LinggongCredit) TableName() string { return "linggong_credits" }
