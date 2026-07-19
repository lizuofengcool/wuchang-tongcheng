// Package dto 风控审核中台精简版数据传输对象
package dto

import "time"

// ReportRequest 举报请求
type ReportRequest struct {
	ReportedUserID uint     `json:"reported_user_id" binding:"required"`           // 被举报人ID
	BizModule      string   `json:"biz_module" binding:"required"`                 // 业务模块
	BizID          string   `json:"biz_id" binding:"required"`                     // 业务ID
	ReportType     string   `json:"report_type" binding:"required,oneof=spam fraud porn violence contraband other"`
	Reason         string   `json:"reason" binding:"required,max=512"`             // 举报理由
	EvidenceImages []string `json:"evidence_images"`                               // 证据图片 URL 列表
}

// ReportInfo 举报信息
type ReportInfo struct {
	ID                uint       `json:"id"`
	ReportNo          string     `json:"report_no"`
	ReporterID        uint       `json:"reporter_id"`
	ReportedUserID    uint       `json:"reported_user_id"`
	ReportedBizModule string     `json:"reported_biz_module"`
	ReportedBizID     string     `json:"reported_biz_id"`
	ReportType        string     `json:"report_type"`
	Reason            string     `json:"reason"`
	EvidenceImages    string     `json:"evidence_images"`
	Status            int        `json:"status"`
	HandleResult      string     `json:"handle_result"`
	HandleRemark      string     `json:"handle_remark"`
	HandlerID         uint       `json:"handler_id"`
	HandledAt         *time.Time `json:"handled_at"`
	SLADeadline       *time.Time `json:"sla_deadline"`
	CreatedAt         time.Time  `json:"created_at"`
}

// HandleReportRequest 处理举报请求（M 端）
type HandleReportRequest struct {
	ReportID     uint   `json:"report_id" binding:"required"`
	HandleResult string `json:"handle_result" binding:"required,oneof=warning remove ban note"`
	HandleRemark string `json:"handle_remark" binding:"max=512"`
	// 处罚相关
	NeedPenalty    bool   `json:"need_penalty"`                                // 是否处罚被举报人
	PenaltyLevel   string `json:"penalty_level" binding:"omitempty,oneof=warning limit mute ban_1d ban_7d ban_forever"`
	PenaltyMinutes int    `json:"penalty_minutes"`                              // 处罚时长（分钟）
}

// ReportListRequest 举报列表请求
type ReportListRequest struct {
	Status      int    `form:"status" json:"status"`
	ReportType  string `form:"report_type" json:"report_type"`
	BizModule   string `form:"biz_module" json:"biz_module"`
	ReporterID  uint   `form:"reporter_id" json:"reporter_id"`
	Page        int    `form:"page" json:"page"`
	PageSize    int    `form:"page_size" json:"page_size"`
}

// SensitiveWordRequest 敏感词管理请求
type SensitiveWordRequest struct {
	Word        string `json:"word" binding:"required,max=64"`
	WordType    string `json:"word_type" binding:"required,oneof=politics porn violence contraband ad"`
	Category    string `json:"category"`
	Replacement string `json:"replacement"`
}

// CheckTextRequest 文本审核请求（敏感词检测）
type CheckTextRequest struct {
	Text        string `json:"text" binding:"required,max=10000"` // 待检测文本
	Replacement string `json:"replacement"`                       // 替换字符（默认 ***）
}

// CheckTextResponse 文本审核响应
type CheckTextResponse struct {
	Passed       bool     `json:"passed"`        // 是否通过
	HitWords     []string `json:"hit_words"`     // 命中敏感词
	CleanedText  string   `json:"cleaned_text"`  // 替换后文本
	HitCount     int      `json:"hit_count"`     // 命中数量
}

// BlacklistRequest 黑名单请求
type BlacklistRequest struct {
	TargetType  string `json:"target_type" binding:"required,oneof=user ip device"`
	TargetValue string `json:"target_value" binding:"required,max=128"`
	Reason      string `json:"reason" binding:"max=256"`
	ExpireAt    *time.Time `json:"expire_at"`                          // 过期时间，nil 表示永久
}

// BlacklistCheckRequest 黑名单检查请求
type BlacklistCheckRequest struct {
	TargetType  string `json:"target_type" binding:"required,oneof=user ip device"`
	TargetValue string `json:"target_value" binding:"required"`
}

// UserScoreInfo 用户风险分信息
type UserScoreInfo struct {
	UserID          uint       `json:"user_id"`
	Score           int        `json:"score"`
	Level           string     `json:"level"`
	ViolationCount  int        `json:"violation_count"`
	ReportCount     int        `json:"report_count"`
	LastViolationAt *time.Time `json:"last_violation_at"`
}

// ViolationInfo 违规处罚信息
type ViolationInfo struct {
	ID            uint       `json:"id"`
	UserID        uint       `json:"user_id"`
	ViolationType string     `json:"violation_type"`
	Level         string     `json:"level"`
	Reason        string     `json:"reason"`
	BizModule     string     `json:"biz_module"`
	BizID         string     `json:"biz_id"`
	ReportID      uint       `json:"report_id"`
	PenaltyStart  *time.Time `json:"penalty_start"`
	PenaltyEnd    *time.Time `json:"penalty_end"`
	Status        int        `json:"status"`
	AppealStatus  int        `json:"appeal_status"`
	CreatedAt     time.Time  `json:"created_at"`
}

// AppealRequest 申诉请求
type AppealRequest struct {
	ViolationID uint   `json:"violation_id" binding:"required"`
	Remark      string `json:"remark" binding:"required,max=512"`
}

// AuditResultRequest 内容审核请求（其他模块调用）
type AuditResultRequest struct {
	Module   string `json:"module" binding:"required"` // 业务模块
	BizID    string `json:"biz_id" binding:"required"` // 业务ID
	UserID   uint   `json:"user_id"`                   // 发布用户
	Title    string `json:"title"`                     // 标题
	Content  string `json:"content"`                   // 内容
	Images   []string `json:"images"`                  // 图片URL
}

// AuditResultResponse 内容审核响应
type AuditResultResponse struct {
	Passed     bool     `json:"passed"`      // 是否通过
	RiskLevel  string   `json:"risk_level"`  // 风险等级 safe/warning/danger
	HitWords   []string `json:"hit_words"`   // 命中敏感词
	Reasons    []string `json:"reasons"`     // 不通过原因
	Suggestion string   `json:"suggestion"`  // 处理建议
}
