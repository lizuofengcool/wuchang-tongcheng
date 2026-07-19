// Package model 风控审核中台精简版数据模型
// 依据 ershou 模块依赖：举报 + 敏感词 + 审核规则 + 黑名单 + 风险分 + 违规处罚
// 依据需求文档 1.10：4 维数据隔离（region_id + user_id）
package model

import (
	"time"

	"wuchang-tongcheng/internal/pkg/database"
)

// === 举报状态 ===
const (
	ReportStatusPending    = 0 // 待处理
	ReportStatusProcessing = 1 // 处理中
	ReportStatusHandled    = 2 // 已处理
	ReportStatusArbitrated = 3 // 已仲裁
	ReportStatusRevoked    = 4 // 已撤销
)

// === 举报类型 ===
const (
	ReportTypeSpam       = "spam"       // 垃圾信息
	ReportTypeFraud      = "fraud"      // 欺诈
	ReportTypePorn       = "porn"       // 色情
	ReportTypeViolence   = "violence"   // 暴力
	ReportTypeContraband = "contraband" // 违禁品
	ReportTypeOther      = "other"      // 其他
)

// === 处理结果 ===
const (
	HandleResultWarning = "warning" // 警告
	HandleResultRemove  = "remove"  // 删除内容
	HandleResultBan     = "ban"     // 封禁
	HandleResultNote    = "note"    // 备注
)

// === 敏感词类型 ===
const (
	WordTypePolitics   = "politics"   // 政治
	WordTypePorn       = "porn"       // 色情
	WordTypeViolence   = "violence"   // 暴力
	WordTypeContraband = "contraband" // 违禁
	WordTypeAd         = "ad"         // 广告
)

// === 审核规则类型 ===
const (
	RuleTypeSensitiveWord = "sensitive_word" // 敏感词
	RuleTypePrice         = "price"          // 价格
	RuleTypeFrequency     = "frequency"      // 频率
	RuleTypeContraband    = "contraband"     // 违禁
)

// === 黑名单目标类型 ===
const (
	BlacklistTargetUser   = "user"   // 用户
	BlacklistTargetIP     = "ip"     // IP
	BlacklistTargetDevice = "device" // 设备
)

// === 风险等级 ===
const (
	RiskLevelSafe    = "safe"    // 安全
	RiskLevelWarning = "warning" // 警告
	RiskLevelDanger  = "danger"  // 危险
)

// === 违规类型 ===
const (
	ViolationTypeSpam       = "spam"
	ViolationTypeFraud      = "fraud"
	ViolationTypePorn       = "porn"
	ViolationTypeViolence   = "violence"
	ViolationTypeContraband = "contraband"
	ViolationTypeAd         = "ad"
)

// === 处罚级别 ===
const (
	PenaltyLevelWarning    = "warning"     // 警告
	PenaltyLevelLimit      = "limit"       // 限制
	PenaltyLevelMute       = "mute"        // 禁言
	PenaltyLevelBan1d      = "ban_1d"      // 封禁1天
	PenaltyLevelBan7d      = "ban_7d"      // 封禁7天
	PenaltyLevelBanForever = "ban_forever" // 永久封禁
)

// === 违规状态 ===
const (
	ViolationStatusActive    = 1 // 生效中
	ViolationStatusEnded     = 0 // 已结束
	ViolationStatusAppealed  = 2 // 已申诉撤销
)

// === 申诉状态 ===
const (
	AppealStatusNone       = 0 // 未申诉
	AppealStatusProcessing = 1 // 申诉中
	AppealStatusSuccess    = 2 // 申诉成功
	AppealStatusFailed     = 3 // 申诉失败
)

// Report 举报工单
type Report struct {
	database.RegionBaseModel

	ReportNo          string     `gorm:"size:64;not null;uniqueIndex" json:"report_no"`              // 举报单号
	ReporterID        uint       `gorm:"index;not null" json:"reporter_id"`                          // 举报人ID
	ReportedUserID    uint       `gorm:"index" json:"reported_user_id"`                              // 被举报人ID
	ReportedBizModule string     `gorm:"size:32;index" json:"reported_biz_module"`                   // 被举报业务模块
	ReportedBizID     string     `gorm:"size:128;index" json:"reported_biz_id"`                      // 被举报业务ID
	ReportType        string     `gorm:"size:32" json:"report_type"`                                 // 举报类型
	Reason            string     `gorm:"size:512" json:"reason"`                                     // 举报理由
	EvidenceImages    string     `gorm:"type:jsonb;default:'[]'::jsonb" json:"evidence_images"`      // 证据图片 JSON
	Status            int        `gorm:"default:0;index" json:"status"`                              // 状态
	HandleResult      string     `gorm:"size:32" json:"handle_result"`                               // 处理结果
	HandleRemark      string     `gorm:"type:text" json:"handle_remark"`                             // 处理备注
	HandlerID         uint       `gorm:"default:0" json:"handler_id"`                                // 处理人ID
	HandledAt         *time.Time `json:"handled_at"`                                                  // 处理时间
	SLADeadline       *time.Time `json:"sla_deadline"`                                                // SLA到期时间
}

// TableName 表名
func (Report) TableName() string { return "risk_reports" }

// SensitiveWord 敏感词
type SensitiveWord struct {
	ID          uint      `gorm:"primarykey" json:"id"`
	Word        string    `gorm:"size:64;not null;uniqueIndex" json:"word"`         // 敏感词
	WordType    string    `gorm:"size:32;index;default:'politics'" json:"word_type"` // 类型
	Category    string    `gorm:"size:32" json:"category"`                          // 分类
	Replacement string    `gorm:"size:32;default:'***'" json:"replacement"`         // 替换字符
	Status      int       `gorm:"default:1;index" json:"status"`                    // 状态 1启用 0禁用
	CreatedAt   time.Time `gorm:"not null;default:now()" json:"created_at"`
	UpdatedAt   time.Time `gorm:"not null;default:now()" json:"updated_at"`
}

// TableName 表名
func (SensitiveWord) TableName() string { return "risk_sensitive_words" }

// AuditRule 审核规则
type AuditRule struct {
	ID          uint      `gorm:"primarykey" json:"id"`
	RuleName    string    `gorm:"size:64;not null;uniqueIndex" json:"rule_name"` // 规则名
	RuleType    string    `gorm:"size:32;index" json:"rule_type"`                // 规则类型
	Config      string    `gorm:"type:jsonb;default:'{}'::jsonb" json:"config"`  // 配置 JSON
	Description string    `gorm:"size:256" json:"description"`                  // 描述
	Status      int       `gorm:"default:1;index" json:"status"`                 // 状态
	CreatedAt   time.Time `gorm:"not null;default:now()" json:"created_at"`
	UpdatedAt   time.Time `gorm:"not null;default:now()" json:"updated_at"`
}

// TableName 表名
func (AuditRule) TableName() string { return "risk_audit_rules" }

// Blacklist 黑名单
type Blacklist struct {
	ID         uint      `gorm:"primarykey" json:"id"`
	RegionID   uint      `gorm:"index;not null;default:1" json:"region_id"`
	TargetType string    `gorm:"size:16;default:'user';index" json:"target_type"` // 目标类型
	TargetValue string   `gorm:"size:128;index" json:"target_value"`             // 目标值
	Reason     string    `gorm:"size:256" json:"reason"`                         // 原因
	OperatorID uint      `gorm:"default:0" json:"operator_id"`                   // 操作人ID
	ExpireAt   *time.Time `json:"expire_at"`                                      // 过期时间
	Status     int       `gorm:"default:1;index" json:"status"`                  // 状态 1生效 0失效
	CreatedAt  time.Time `gorm:"not null;default:now()" json:"created_at"`
	UpdatedAt  time.Time `gorm:"not null;default:now()" json:"updated_at"`
}

// TableName 表名
func (Blacklist) TableName() string { return "risk_blacklist" }

// UserScore 用户风险分
type UserScore struct {
	ID             uint      `gorm:"primarykey" json:"id"`
	RegionID       uint      `gorm:"index;not null;default:1" json:"region_id"`
	UserID         uint      `gorm:"not null;uniqueIndex" json:"user_id"`           // 用户ID
	Score          int       `gorm:"default:100;index" json:"score"`                // 风险分 0-100
	Level          string    `gorm:"size:16;default:'safe'" json:"level"`           // 风险等级
	ViolationCount int       `gorm:"default:0" json:"violation_count"`              // 违规次数
	ReportCount    int       `gorm:"default:0" json:"report_count"`                 // 被举报次数
	LastViolationAt *time.Time `json:"last_violation_at"`                           // 最后违规时间
	CreatedAt      time.Time `gorm:"not null;default:now()" json:"created_at"`
	UpdatedAt      time.Time `gorm:"not null;default:now()" json:"updated_at"`
}

// TableName 表名
func (UserScore) TableName() string { return "risk_user_scores" }

// Violation 违规处罚
type Violation struct {
	database.RegionBaseModel

	UserID         uint       `gorm:"index;not null" json:"user_id"`                              // 用户ID
	ViolationType  string     `gorm:"size:32" json:"violation_type"`                              // 违规类型
	Level          string     `gorm:"size:16;default:'warning'" json:"level"`                     // 处罚级别
	Reason         string     `gorm:"size:512" json:"reason"`                                     // 原因
	BizModule      string     `gorm:"size:32" json:"biz_module"`                                  // 业务模块
	BizID          string     `gorm:"size:128" json:"biz_id"`                                     // 业务ID
	ReportID       uint       `gorm:"default:0" json:"report_id"`                                 // 关联举报ID
	PenaltyStart   *time.Time `json:"penalty_start"`                                              // 处罚开始
	PenaltyEnd     *time.Time `json:"penalty_end"`                                                // 处罚结束
	Status         int        `gorm:"default:1;index" json:"status"`                              // 状态
	AppealStatus   int        `gorm:"default:0" json:"appeal_status"`                              // 申诉状态
	AppealRemark   string     `gorm:"type:text" json:"appeal_remark"`                             // 申诉备注
}

// TableName 表名
func (Violation) TableName() string { return "risk_violations" }
