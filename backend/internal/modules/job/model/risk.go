// Package model 举报工单 + 公司评价 + 审核规则（对标 BOSS直聘/看准）
// 虚假招聘/诈骗/色情/侵权 + 5星+优缺点+追评 + 敏感词/薪资异常/频率限制
package model

import (
	"time"

	"wuchang-tongcheng/internal/pkg/database"
)

// === 举报类型常量 ===
const (
	ReportTypeFake          = "fake"          // 虚假招聘
	ReportTypeScam          = "scam"          // 诈骗欺诈
	ReportTypePorn          = "porn"          // 色情低俗
	ReportTypeProhibited    = "prohibited"    // 违禁品
	ReportTypeInfringement  = "infringement"  // 侵权
	ReportTypeSpam          = "spam"          // 垃圾广告
	ReportTypeHarassment    = "harassment"    // 骚扰
	ReportTypePrivacy       = "privacy"       // 隐私泄露
	ReportTypeSalaryAnomaly = "salary_anomaly" // 薪资异常
	ReportTypeOther         = "other"         // 其他
)

// === 举报状态常量 ===
const (
	ReportStatusPending    = 0 // 待处理
	ReportStatusProcessing = 1 // 处理中
	ReportStatusResolved   = 2 // 已处理（成立）
	ReportStatusRejected   = 3 // 已处理（不成立）
	ReportStatusAppealed   = 4 // 申诉中
	ReportStatusClosed     = 5 // 已关闭
)

// === 处罚类型常量 ===
const (
	PenaltyTypeWarning     = "warning"     // 警告
	PenaltyTypeLimit       = "limit"        // 限制发布
	PenaltyTypeBan1d       = "ban1d"        // 封禁 1 天
	PenaltyTypeBan7d       = "ban7d"        // 封禁 7 天
	PenaltyTypeBan30d      = "ban30d"       // 封禁 30 天
	PenaltyTypeBanForever  = "ban_forever" // 永久封禁
	PenaltyTypeCloseJob    = "close_job"    // 关闭职位
	PenaltyTypeFreezeCompany = "freeze_company" // 冻结公司
)

// JobReport 举报工单表
type JobReport struct {
	database.BaseModel // id/created_at/updated_at/deleted_at
	ReportNo          string     `gorm:"size:64;not null;uniqueIndex" json:"report_no"`          // 举报单号
	TargetType        string     `gorm:"size:32;not null;index:idx_job_report_target" json:"target_type"` // 目标类型 job/company/resume/recruiter/review
	TargetID          uint       `gorm:"not null;index:idx_job_report_target" json:"target_id"`  // 目标 ID
	TargetUserID      uint       `gorm:"not null;index" json:"target_user_id"`                   // 目标所属用户 ID
	ReporterID        uint       `gorm:"not null;index" json:"reporter_id"`                     // 举报人 ID
	ReporterName      string     `gorm:"size:50" json:"reporter_name"`                          // 举报人昵称
	ReportedUserID    uint       `gorm:"not null;index" json:"reported_user_id"`                 // 被举报人 ID
	ReportedUserName  string     `gorm:"size:50" json:"reported_user_name"`                    // 被举报人昵称
	ReportType        string     `gorm:"size:32;not null;index" json:"report_type"`             // 举报类型
	Reason            string     `gorm:"size:500;not null" json:"reason"`                       // 举报原因
	Description       string     `gorm:"type:text" json:"description"`                          // 详细描述
	EvidenceImages    JSONB      `gorm:"type:jsonb" json:"evidence_images"`                    // 证据图片 URL 数组
	Status            int        `gorm:"default:0;index" json:"status"`                        // 6 状态
	HandlerID         uint       `gorm:"index" json:"handler_id"`                              // 处理人 ID（M 端审核员）
	HandlerName       string     `gorm:"size:50" json:"handler_name"`                          // 处理人昵称
	HandleResult      string     `gorm:"type:text" json:"handle_result"`                        // 处理结果
	PenaltyType       string     `gorm:"size:32" json:"penalty_type"`                          // 处罚类型
	PenaltyUserID     uint       `gorm:"index" json:"penalty_user_id"`                          // 被处罚用户 ID
	SLADeadline       *time.Time `gorm:"index" json:"sla_deadline"`                              // SLA 截止时间
	HandledAt         *time.Time `gorm:"index" json:"handled_at"`                               // 处理时间
	AppealReason      string     `gorm:"type:text" json:"appeal_reason"`                       // 申诉理由
	AppealedAt        *time.Time `gorm:"index" json:"appealed_at"`                              // 申诉时间
	AppealResult      string     `gorm:"type:text" json:"appeal_result"`                       // 申诉结果
	AppealHandlerID   uint       `gorm:"index" json:"appeal_handler_id"`                        // 申诉处理人 ID
	AppealHandledAt   *time.Time `gorm:"index" json:"appeal_handled_at"`                        // 申诉处理时间
}

// TableName 表名（job_ 前缀）
func (JobReport) TableName() string { return "job_reports" }

// === 评价类型常量 ===
const (
	ReviewTypeEmployee      = "employee"      // 在职员工评价
	ReviewTypeFormerEmployee = "former_employee" // 离职员工评价
	ReviewTypeInterviewee  = "interviewee"    // 面试者评价
	ReviewTypeCandidate     = "candidate"     // 候选人评价
)

// === 评价状态常量 ===
const (
	ReviewStatusPending  = 0 // 待审核
	ReviewStatusApproved = 1 // 已通过
	ReviewStatusRejected = 2 // 已拒绝
	ReviewStatusHidden   = 3 // 已隐藏
)

// JobReview 公司评价表（对标看准/BOSS直聘）
type JobReview struct {
	database.RegionBaseModel // id/region_id/created_at/updated_at/deleted_at
	CompanyID       uint       `gorm:"not null;index;uniqueIndex:uniq_job_reviews_company_reviewer" json:"company_id"` // 公司 ID
	ReviewerID      uint       `gorm:"not null;index;uniqueIndex:uniq_job_reviews_company_reviewer" json:"reviewer_id"` // 评价人 ID
	ReviewerName    string     `gorm:"size:50" json:"reviewer_name"`                          // 评价人昵称
	ReviewerAvatar  string     `gorm:"size:255" json:"reviewer_avatar"`                      // 评价人头像
	ReviewType      string     `gorm:"size:32;default:'employee';index" json:"review_type"`   // 评价类型
	Rating          int        `gorm:"default:5;index" json:"rating"`                         // 评分 1-5 星
	Content         string     `gorm:"type:text" json:"content"`                              // 评价内容
	Images          JSONB      `gorm:"type:jsonb" json:"images"`                              // 评价图片 URL 数组
	VideoURL        string     `gorm:"size:255" json:"video_url"`                            // 视频链接
	IsAnonymous     bool       `gorm:"default:false" json:"is_anonymous"`                    // 是否匿名
	IsRecommended   bool       `gorm:"default:true;index" json:"is_recommended"`            // 是否推荐
	Tags            JSONB      `gorm:"type:jsonb" json:"tags"`                                // 评价标签
	Position        string     `gorm:"size:64" json:"position"`                              // 评价人职位
	Department      string     `gorm:"size:64" json:"department"`                            // 评价人部门
	WorkDuration    string     `gorm:"size:32" json:"work_duration"`                          // 工作时长
	SalaryRange     string     `gorm:"size:64" json:"salary_range"`                           // 薪资范围
	Pros            string     `gorm:"type:text" json:"pros"`                                  // 优点
	Cons            string     `gorm:"type:text" json:"cons"`                                  // 缺点
	Advice          string     `gorm:"type:text" json:"advice"`                                // 给管理层的建议
	Reply           string     `gorm:"type:text" json:"reply"`                                  // 公司回复
	ReplyAt         *time.Time `gorm:"index" json:"reply_at"`                                  // 回复时间
	AppendContent   string     `gorm:"type:text" json:"append_content"`                        // 追评内容
	AppendImages    JSONB      `gorm:"type:jsonb" json:"append_images"`                       // 追评图片
	AppendAt        *time.Time `gorm:"index" json:"append_at"`                                  // 追评时间
	LikeCount       int        `gorm:"default:0" json:"like_count"`                            // 点赞数
	Status          int        `gorm:"default:1;index" json:"status"`                          // 0待审 1通过 2拒绝 3隐藏
}

// TableName 表名（job_ 前缀）
func (JobReview) TableName() string { return "job_reviews" }

// === 审核规则类型常量 ===
const (
	AuditRuleTypeSensitiveWord  = "sensitive_word"  // 敏感词
	AuditRuleTypeSalaryCheck    = "salary_check"     // 薪资异常
	AuditRuleTypeFrequency      = "frequency"        // 频率限制
	AuditRuleTypeFakeRecruit    = "fake_recruit"      // 虚假招聘
	AuditRuleTypeProhibited     = "prohibited"        // 违禁内容
	AuditRuleTypeContent        = "content"           // 内容审核
	AuditRuleTypeCompany        = "company"           // 公司信息审核
	AuditRuleTypeResume         = "resume"            // 简历审核
)

// === 审核规则动作常量 ===
const (
	AuditRuleActionReject    = "reject"    // 拒绝
	AuditRuleActionApproval = "approval"  // 人工审核
	AuditRuleActionWarning  = "warning"    // 警告
	AuditRuleActionManual   = "manual"    // 人工处理
)

// === 审核规则状态常量 ===
const (
	AuditRuleStatusDisabled = 0 // 禁用
	AuditRuleStatusEnabled  = 1 // 启用
)

// JobAuditRule 审核规则表
type JobAuditRule struct {
	database.BaseModel // id/created_at/updated_at/deleted_at
	RuleName    string `gorm:"size:128;not null" json:"rule_name"`                       // 规则名称
	RuleType    string `gorm:"size:32;not null;index" json:"rule_type"`                  // 类型
	RuleKey     string `gorm:"size:64;index" json:"rule_key"`                            // 规则键（如敏感词级别：politics/porn/violence）
	Pattern     string `gorm:"type:text" json:"pattern"`                                  // 匹配模式（正则/关键词列表 JSON）
	Threshold   JSONB  `gorm:"type:jsonb" json:"threshold"`                               // 阈值配置 {min_salary,max_salary,max_count_per_min...}
	Action      string `gorm:"size:32;default:'reject'" json:"action"`                   // 触发动作：reject/approval/warning/manual
	PenaltyType string `gorm:"size:32" json:"penalty_type"`                               // 处罚类型
	Severity    int    `gorm:"default:1;index" json:"severity"`                            // 严重等级 1-5
	Status      int    `gorm:"default:1;index" json:"status"`                              // 0禁用 1启用
	Description string `gorm:"size:500" json:"description"`                                // 规则描述
	Sort        int    `gorm:"default:0;index" json:"sort"`                                // 排序
}

// TableName 表名（job_ 前缀）
func (JobAuditRule) TableName() string { return "job_audit_rules" }
