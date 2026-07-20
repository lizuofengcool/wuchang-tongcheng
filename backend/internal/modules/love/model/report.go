// Package model love 相亲交友数据模型 - 举报表 LoveReport
// 对标陌陌/Soul：举报用户/动态/消息
// 依据需求文档 1.10：4 维数据隔离（region_id + user_id）
package model

import (
	"time"

	"wuchang-tongcheng/internal/pkg/database"
)

// === 处理结果 ===
const (
	ReportHandleResultValid      = "valid"      // 举报成立
	ReportHandleResultInvalid    = "invalid"    // 举报不成立
	ReportHandleResultProcessing = "processing" // 处理中
	ReportHandleResultClosed     = "closed"     // 已关闭
)

// === 处罚类型 ===
const (
	ReportPenaltyWarning   = "warning"   // 警告
	ReportPenaltyFreeze    = "freeze"    // 冻结
	ReportPenaltyBan       = "ban"       // 封禁
	ReportPenaltyDeleteContent = "delete_content" // 删除内容
	ReportPenaltyDeduction = "deduction" // 扣分
	ReportPenaltyNone      = "none"      // 无处罚
)

// LoveReport 举报表
// 举报人 reporter_user_id 举报目标 target_type + target_user_id
type LoveReport struct {
	database.RegionBaseModel

	ReportNo string `gorm:"size:64;not null;uniqueIndex" json:"report_no"` // 举报单号

	// 举报人
	ReporterUserID   uint   `gorm:"index;not null" json:"reporter_user_id"`
	ReporterLoveID   uint   `gorm:"index;not null" json:"reporter_love_id"`
	ReporterNickname string `gorm:"size:64;not null;default:''" json:"reporter_nickname"`
	ReporterAvatar   string `gorm:"size:255;not null;default:''" json:"reporter_avatar"`

	// 举报目标
	TargetType      string `gorm:"size:32;not null;default:'user';index" json:"target_type"` // user/story/message/impression/gift
	TargetUserID    uint   `gorm:"index;not null" json:"target_user_id"`
	TargetLoveID    uint   `gorm:"index;not null" json:"target_love_id"`
	TargetNickname  string `gorm:"size:64;not null;default:''" json:"target_nickname"`
	TargetAvatar    string `gorm:"size:255;not null;default:''" json:"target_avatar"`
	TargetID        uint   `gorm:"not null;default:0" json:"target_id"` // 目标资源 ID（动态/消息 ID）

	// 举报理由
	ReasonType   string `gorm:"size:32;not null;default:'';index" json:"reason_type"` // 参考 ReportReason*
	ReasonDetail string `gorm:"type:text" json:"reason_detail"`

	// 证据
	EvidenceImages JSONB `gorm:"type:jsonb" json:"evidence_images"`
	EvidenceVideos JSONB `gorm:"type:jsonb" json:"evidence_videos"`
	EvidenceText   string `gorm:"type:text" json:"evidence_text"`
	ChatSnapshot   JSONB `gorm:"type:jsonb" json:"chat_snapshot"`

	IP string `gorm:"size:64;not null;default:''" json:"ip"`

	// 处理
	Status       int        `gorm:"not null;default:0;index" json:"status"` // 0待处理 1处理中 2已处理 3已驳回
	HandledBy    uint       `gorm:"not null;default:0" json:"handled_by"`
	HandledAt    *time.Time `json:"handled_at"`
	HandleResult string     `gorm:"size:32;not null;default:''" json:"handle_result"` // valid/invalid/processing/closed
	HandleRemark string     `gorm:"type:text" json:"handle_remark"`

	// 处罚
	PenaltyType       string     `gorm:"size:32;not null;default:''" json:"penalty_type"` // warning/freeze/ban/delete_content/deduction/none
	PenaltyDuration   int        `gorm:"not null;default:0" json:"penalty_duration"`     // 处罚时长（小时）
	PenaltyExpiredAt  *time.Time `json:"penalty_expired_at"`

	// 申诉
	AppealStatus     int        `gorm:"not null;default:0;index" json:"appeal_status"` // 0未申诉 1申诉中 2申诉成功 3申诉驳回
	AppealReason     string     `gorm:"type:text" json:"appeal_reason"`
	AppealedAt       *time.Time `json:"appealed_at"`
	AppealHandledBy  uint       `gorm:"not null;default:0" json:"appeal_handled_by"`
	AppealHandledAt  *time.Time `json:"appeal_handled_at"`
	AppealResult     string     `gorm:"size:32;not null;default:''" json:"appeal_result"`
	AppealRemark     string     `gorm:"type:text" json:"appeal_remark"`

	// 风险评分
	RiskScore int `gorm:"not null;default:0" json:"risk_score"`
}

// TableName 表名
func (LoveReport) TableName() string { return "love_reports" }
