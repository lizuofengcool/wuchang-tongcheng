// Package model 风控中台扩展数据模型
// 依据 015_risk_full.sql：举报证据/申诉/规则/评分记录/审核日志
package model

import (
	"time"

	"wuchang-tongcheng/internal/pkg/database"
)

// === 证据类型 ===
const (
	EvidenceTypeImage = "image"
	EvidenceTypeVideo = "video"
	EvidenceTypeText  = "text"
	EvidenceTypeAudio = "audio"
)

// === 申诉状态 ===
const (
	AppealStatusPending  = 0 // 待审
	AppealStatusApproved = 1 // 通过
	AppealStatusRejected = 2 // 拒绝
)

// === 规则类型（实时风控扩展） ===
// 注：RuleTypeFrequency 已在 risk.go 中声明，此处仅新增其他类型
const (
	RuleTypeAmount   = "amount"
	RuleTypeContent  = "content"
	RuleTypeBehavior = "behavior"
)

// === 规则动作 ===
const (
	RuleActionBlock  = "block"
	RuleActionReview = "review"
	RuleActionWarn   = "warn"
	RuleActionLog    = "log"
)

// === 评分目标类型 ===
const (
	ScoreTargetUser    = "user"
	ScoreTargetContent = "content"
	ScoreTargetIP      = "ip"
	ScoreTargetDevice  = "device"
)

// === 审核动作 ===
const (
	AuditActionApprove  = "approve"
	AuditActionReject   = "reject"
	AuditActionEscalate = "escalate"
	AuditActionBan      = "ban"
	AuditActionUnban    = "unban"
)

// ReportEvidence 举报证据
type ReportEvidence struct {
	ID          uint      `gorm:"primarykey" json:"id"`
	RegionID    uint      `gorm:"index;not null;default:1" json:"region_id"`
	ReportID    uint      `gorm:"not null;index" json:"report_id"`
	EvidenceType string   `gorm:"size:16;not null;default:'image';index" json:"evidence_type"`
	URL         string    `gorm:"size:512;not null;default:''" json:"url"`
	Description string    `gorm:"size:256;not null;default:''" json:"description"`
	UploaderID  uint      `gorm:"default:0" json:"uploader_id"`
	FileSize    int64     `gorm:"default:0" json:"file_size"`
	FileHash    string    `gorm:"size:64;not null;default:''" json:"file_hash"`
	Extra       string    `gorm:"type:jsonb;default:'{}'::jsonb" json:"extra"`
	CreatedAt   time.Time `gorm:"not null;default:now()" json:"created_at"`
}

// TableName 表名
func (ReportEvidence) TableName() string { return "risk_report_evidence" }

// Appeal 违规申诉
type Appeal struct {
	database.RegionBaseModel

	AppealNo      string     `gorm:"size:64;not null;uniqueIndex" json:"appeal_no"`
	ViolationID   uint       `gorm:"not null;index" json:"violation_id"`
	UserID        uint       `gorm:"not null;index" json:"user_id"`
	Reason        string     `gorm:"type:text;not null;default:''" json:"reason"`
	EvidenceImages string    `gorm:"type:jsonb;default:'[]'::jsonb" json:"evidence_images"`
	Status        int        `gorm:"default:0;index" json:"status"`
	HandlerID     uint       `gorm:"default:0" json:"handler_id"`
	HandleRemark  string     `gorm:"type:text;not null;default:''" json:"handle_remark"`
	HandledAt     *time.Time `json:"handled_at"`
}

// TableName 表名
func (Appeal) TableName() string { return "risk_appeals" }

// Rule 风控规则（实时决策）
type Rule struct {
	ID         uint       `gorm:"primarykey" json:"id"`
	RegionID   uint       `gorm:"index;not null;default:1" json:"region_id"`
	RuleName   string     `gorm:"size:64;not null;uniqueIndex" json:"rule_name"`
	RuleType   string     `gorm:"size:32;not null;default:'';index" json:"rule_type"`
	Description string    `gorm:"size:256;not null;default:''" json:"description"`
	Config     string     `gorm:"type:jsonb;default:'{}'::jsonb" json:"config"`
	Action     string     `gorm:"size:32;not null;default:'block'" json:"action"`
	Priority   int        `gorm:"default:100;index" json:"priority"`
	Status     int        `gorm:"default:1;index" json:"status"`
	HitCount   int64      `gorm:"default:0" json:"hit_count"`
	LastHitAt  *time.Time `json:"last_hit_at"`
	CreatedAt  time.Time  `gorm:"not null;default:now()" json:"created_at"`
	UpdatedAt  time.Time  `gorm:"not null;default:now()" json:"updated_at"`
}

// TableName 表名
func (Rule) TableName() string { return "risk_rules" }

// ScoreRecord 风险评分记录
type ScoreRecord struct {
	ID           uint      `gorm:"primarykey" json:"id"`
	RegionID     uint      `gorm:"index;not null;default:1" json:"region_id"`
	UserID       uint      `gorm:"not null;index" json:"user_id"`
	TargetType   string    `gorm:"size:32;not null;default:'user'" json:"target_type"`
	TargetValue  string    `gorm:"size:128;not null;default:''" json:"target_value"`
	ContentType  string    `gorm:"size:32;not null;default:''" json:"content_type"`
	ContentID    string    `gorm:"size:128;not null;default:''" json:"content_id"`
	Score        int       `gorm:"default:0" json:"score"`
	Level        string    `gorm:"size:16;not null;default:'safe';index" json:"level"`
	Reasons      string    `gorm:"type:jsonb;default:'[]'::jsonb" json:"reasons"`
	RuleIDs      string    `gorm:"type:jsonb;default:'[]'::jsonb" json:"rule_ids"`
	ActionTaken  string    `gorm:"size:32;not null;default:''" json:"action_taken"`
	CreatedAt    time.Time `gorm:"not null;default:now();index" json:"created_at"`
}

// TableName 表名
func (ScoreRecord) TableName() string { return "risk_score_records" }

// AuditLog 审核日志
type AuditLog struct {
	ID           uint      `gorm:"primarykey" json:"id"`
	RegionID     uint      `gorm:"index;not null;default:1" json:"region_id"`
	AuditorID    uint      `gorm:"not null;index" json:"auditor_id"`
	AuditorName  string    `gorm:"size:64;not null;default:''" json:"auditor_name"`
	Action       string    `gorm:"size:32;not null;default:'';index" json:"action"`
	TargetType   string    `gorm:"size:32;not null;default:''" json:"target_type"`
	TargetID     uint      `gorm:"default:0" json:"target_id"`
	BizModule    string    `gorm:"size:32;not null;default:''" json:"biz_module"`
	BizID        string    `gorm:"size:128;not null;default:''" json:"biz_id"`
	BeforeStatus string    `gorm:"size:32;not null;default:''" json:"before_status"`
	AfterStatus  string    `gorm:"size:32;not null;default:''" json:"after_status"`
	Remark       string    `gorm:"type:text;not null;default:''" json:"remark"`
	IP           string    `gorm:"size:64;not null;default:''" json:"ip"`
	UserAgent    string    `gorm:"size:256;not null;default:''" json:"user_agent"`
	CreatedAt    time.Time `gorm:"not null;default:now();index" json:"created_at"`
}

// TableName 表名
func (AuditLog) TableName() string { return "risk_audit_logs" }
