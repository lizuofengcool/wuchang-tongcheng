// Package dto 风控中台扩展数据传输对象
package dto

import "time"

// EvidenceInfo 证据信息
type EvidenceInfo struct {
	ID           uint      `json:"id"`
	ReportID     uint      `json:"report_id"`
	EvidenceType string    `json:"evidence_type"`
	URL          string    `json:"url"`
	Description  string    `json:"description"`
	UploaderID   uint      `json:"uploader_id"`
	FileSize     int64     `json:"file_size"`
	FileHash     string    `json:"file_hash"`
	Extra        string    `json:"extra"`
	CreatedAt    time.Time `json:"created_at"`
}

// AddEvidenceRequest 添加证据请求
type AddEvidenceRequest struct {
	ReportID     uint   `json:"report_id" binding:"required"`
	EvidenceType string `json:"evidence_type" binding:"required,oneof=image video text audio"`
	URL          string `json:"url" binding:"required,max=512"`
	Description  string `json:"description" binding:"max=256"`
	FileSize     int64  `json:"file_size"`
	FileHash     string `json:"file_hash" binding:"max=64"`
	Extra        string `json:"extra"`
}

// AppealInfo 申诉信息
type AppealInfo struct {
	ID             uint       `json:"id"`
	AppealNo       string     `json:"appeal_no"`
	ViolationID    uint       `json:"violation_id"`
	UserID         uint       `json:"user_id"`
	Reason         string     `json:"reason"`
	EvidenceImages string     `json:"evidence_images"`
	Status         int        `json:"status"`
	HandlerID      uint       `json:"handler_id"`
	HandleRemark   string     `json:"handle_remark"`
	HandledAt      *time.Time `json:"handled_at"`
	CreatedAt      time.Time  `json:"created_at"`
}

// CreateAppealRequest 创建申诉请求
type CreateAppealRequest struct {
	ViolationID    uint     `json:"violation_id" binding:"required"`
	Reason         string   `json:"reason" binding:"required,max=512"`
	EvidenceImages []string `json:"evidence_images"`
}

// HandleAppealRequest 处理申诉请求（M 端）
type HandleAppealRequest struct {
	AppealID     uint   `json:"appeal_id" binding:"required"`
	Status       int    `json:"status" binding:"required,oneof=1 2"`
	HandleRemark string `json:"handle_remark" binding:"max=512"`
}

// RuleInfo 风控规则信息
type RuleInfo struct {
	ID          uint      `json:"id"`
	RuleName    string    `json:"rule_name"`
	RuleType    string    `json:"rule_type"`
	Description string    `json:"description"`
	Config      string    `json:"config"`
	Action      string    `json:"action"`
	Priority   int        `json:"priority"`
	Status     int        `json:"status"`
	HitCount   int64      `json:"hit_count"`
	LastHitAt  *time.Time `json:"last_hit_at"`
	CreatedAt  time.Time `json:"created_at"`
}

// CreateRuleRequest 创建风控规则请求
type CreateRuleRequest struct {
	RuleName    string `json:"rule_name" binding:"required,max=64"`
	RuleType    string `json:"rule_type" binding:"required,oneof=frequency amount content behavior sensitive_word price contraband"`
	Description string `json:"description" binding:"max=256"`
	Config      string `json:"config"`
	Action      string `json:"action" binding:"required,oneof=block review warn log"`
	Priority    int    `json:"priority"`
	Status      int    `json:"status"`
}

// UpdateRuleRequest 更新风控规则请求
type UpdateRuleRequest struct {
	Description string `json:"description" binding:"max=256"`
	Config      string `json:"config"`
	Action      string `json:"action" binding:"omitempty,oneof=block review warn log"`
	Priority    int    `json:"priority"`
	Status      int    `json:"status"`
}

// ScoreRecordInfo 风险评分记录信息
type ScoreRecordInfo struct {
	ID          uint      `json:"id"`
	UserID      uint      `json:"user_id"`
	TargetType  string    `json:"target_type"`
	TargetValue string    `json:"target_value"`
	ContentType string    `json:"content_type"`
	ContentID   string    `json:"content_id"`
	Score       int       `json:"score"`
	Level       string    `json:"level"`
	Reasons     string    `json:"reasons"`
	RuleIDs     string    `json:"rule_ids"`
	ActionTaken string    `json:"action_taken"`
	CreatedAt   time.Time `json:"created_at"`
}

// AuditLogInfo 审核日志信息
type AuditLogInfo struct {
	ID           uint      `json:"id"`
	AuditorID    uint      `json:"auditor_id"`
	AuditorName  string    `json:"auditor_name"`
	Action       string    `json:"action"`
	TargetType   string    `json:"target_type"`
	TargetID     uint      `json:"target_id"`
	BizModule    string    `json:"biz_module"`
	BizID        string    `json:"biz_id"`
	BeforeStatus string    `json:"before_status"`
	AfterStatus  string    `json:"after_status"`
	Remark       string    `json:"remark"`
	IP           string    `json:"ip"`
	UserAgent    string    `json:"user_agent"`
	CreatedAt    time.Time `json:"created_at"`
}

// RiskStatisticsResponse 风控统计响应
type RiskStatisticsResponse struct {
	TotalReports      int64 `json:"total_reports"`
	PendingReports    int64 `json:"pending_reports"`
	HandledReports    int64 `json:"handled_reports"`
	TotalAppeals      int64 `json:"total_appeals"`
	PendingAppeals    int64 `json:"pending_appeals"`
	TotalViolations   int64 `json:"total_violations"`
	ActiveViolations  int64 `json:"active_violations"`
	BlacklistCount    int64 `json:"blacklist_count"`
	SensitiveWords    int64 `json:"sensitive_words"`
	RulesCount        int64 `json:"rules_count"`
	AuditLogsCount    int64 `json:"audit_logs_count"`
}
