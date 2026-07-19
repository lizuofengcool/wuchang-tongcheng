// Package model 举报工单 + 交易评价 + 审核规则 + 用户信用（对标转转）
package model

import (
	"time"

	"wuchang-tongcheng/internal/pkg/database"
)

// === 举报类型常量 ===
const (
	ReportTypePorn        = "porn"        // 色情低俗
	ReportTypeScam        = "scam"        // 诈骗欺诈
	ReportTypeFake        = "fake"        // 虚假信息
	ReportTypeProhibited  = "prohibited"  // 违禁品
	ReportTypeInfringement = "infringement" // 侵权
	ReportTypeSpam        = "spam"        // 垃圾广告
	ReportTypeOther       = "other"       // 其他
)

// === 举报状态常量 ===
const (
	ReportStatusPending        = 0 // 待处理
	ReportStatusProcessing     = 1 // 处理中
	ReportStatusResolved       = 2 // 已处理（成立）
	ReportStatusRejected       = 3 // 已处理（不成立）
	ReportStatusAppealed       = 4 // 申诉中
	ReportStatusClosed         = 5 // 已关闭
)

// ErshouReport 举报工单表
type ErshouReport struct {
	database.BaseModel // id/created_at/updated_at/deleted_at
	ReportNo       string     `gorm:"size:64;not null;uniqueIndex" json:"report_no"`              // 举报单号
	ErshouID       uint       `gorm:"not null;index" json:"ershou_id"`                              // 被举报的物品ID
	ReporterID     uint       `gorm:"not null;index" json:"reporter_id"`                            // 举报人ID
	ReporterName   string     `gorm:"size:50" json:"reporter_name"`                                // 举报人昵称
	ReportedUserID uint       `gorm:"not null;index" json:"reported_user_id"`                       // 被举报人ID
	ReportedUserName string   `gorm:"size:50" json:"reported_user_name"`                           // 被举报人昵称
	ReportType     string     `gorm:"size:32;not null;index" json:"report_type"`                    // 举报类型 porn/scam/fake/prohibited/infringement/spam/other
	Reason         string     `gorm:"size:500;not null" json:"reason"`                             // 举报原因
	Description    string     `gorm:"type:text" json:"description"`                                 // 详细描述
	EvidenceImages JSONB       `gorm:"type:jsonb" json:"evidence_images"`                            // 证据图片 URL 数组
	Status         int        `gorm:"default:0;index" json:"status"`                              // 6状态
	HandlerID      uint       `gorm:"index" json:"handler_id"`                                    // 处理人ID（M端审核员）
	HandlerName    string     `gorm:"size:50" json:"handler_name"`                                // 处理人昵称
	HandleResult   string     `gorm:"type:text" json:"handle_result"`                              // 处理结果
	PenaltyType    string     `gorm:"size:32" json:"penalty_type"`                                // 处罚类型 warning/limit/ban1d/ban7d/banForever
	PenaltyUserID  uint       `gorm:"index" json:"penalty_user_id"`                                // 被处罚用户ID
	SLADeadline    *time.Time `gorm:"index" json:"sla_deadline"`                                    // SLA 截止时间（24h响应/72h处理）
	HandledAt      *time.Time `gorm:"index" json:"handled_at"`                                     // 处理时间
	AppealReason   string     `gorm:"type:text" json:"appeal_reason"`                              // 申诉理由
	AppealedAt     *time.Time `gorm:"index" json:"appealed_at"`                                    // 申诉时间
	AppealResult   string     `gorm:"type:text" json:"appeal_result"`                              // 申诉结果
	AppealHandlerID uint      `gorm:"index" json:"appeal_handler_id"`                              // 申诉处理人ID
	AppealHandledAt *time.Time `gorm:"index" json:"appeal_handled_at"`                              // 申诉处理时间
}

// TableName 表名（ers_ 前缀）
func (ErshouReport) TableName() string { return "ers_reports" }

// ErshouReview 交易评价表（对标转转）
type ErshouReview struct {
	database.BaseModel // id/created_at/updated_at/deleted_at
	OrderID        uint       `gorm:"not null;uniqueIndex:uniq_review_order_user" json:"order_id"`  // 关联订单ID
	ErshouID       uint       `gorm:"not null;index" json:"ershou_id"`                              // 关联二手物品ID
	ReviewerID     uint       `gorm:"not null;uniqueIndex:uniq_review_order_user;index" json:"reviewer_id"` // 评价人ID（买家）
	ReviewerName   string     `gorm:"size:50" json:"reviewer_name"`                                // 评价人昵称
	ReviewerAvatar string     `gorm:"size:255" json:"reviewer_avatar"`                            // 评价人头像
	RevieweeID     uint       `gorm:"not null;index" json:"reviewee_id"`                            // 被评价人ID（卖家）
	ReviewType     string     `gorm:"size:16;default:'buyer_to_seller'" json:"review_type"`        // buyer_to_seller/seller_to_buyer
	Rating         int        `gorm:"not null;default:5;index" json:"rating"`                       // 评分 1-5 星
	Content        string     `gorm:"type:text" json:"content"`                                    // 评价内容
	Images         JSONB       `gorm:"type:jsonb" json:"images"`                                    // 评价图片 URL 数组
	VideoURL       string     `gorm:"size:255" json:"video_url"`                                  // 评价视频 URL
	IsAnonymous    bool       `gorm:"default:false" json:"is_anonymous"`                          // 是否匿名
	IsRecommended  bool       `gorm:"default:true;index" json:"is_recommended"`                   // 是否推荐
	Tags           JSONB       `gorm:"type:jsonb" json:"tags"`                                       // 评价标签 ["描述相符","发货快"]
	Reply          string     `gorm:"type:text" json:"reply"`                                       // 卖家回复
	ReplyAt        *time.Time `gorm:"index" json:"reply_at"`                                        // 回复时间
	AppendContent  string     `gorm:"type:text" json:"append_content"`                              // 追评内容
	AppendImages   JSONB       `gorm:"type:jsonb" json:"append_images"`                              // 追评图片
	AppendAt       *time.Time `gorm:"index" json:"append_at"`                                       // 追评时间
	LikeCount      int        `gorm:"default:0" json:"like_count"`                                  // 点赞数
}

// TableName 表名（ers_ 前缀）
func (ErshouReview) TableName() string { return "ers_reviews" }

// === 审核规则类型常量 ===
const (
	AuditRuleTypeSensitiveWord = "sensitive_word" // 敏感词
	AuditRuleTypePriceCheck    = "price_check"    // 价格异常
	AuditRuleTypeFrequency     = "frequency"       // 频率限制
	AuditRuleTypeProhibited    = "prohibited"      // 违禁品
	AuditRuleTypeContent       = "content"          // 内容审核
)

// === 审核规则状态常量 ===
const (
	AuditRuleStatusDisabled = 0 // 禁用
	AuditRuleStatusEnabled  = 1 // 启用
)

// ErshouAuditRule 审核规则表
type ErshouAuditRule struct {
	database.BaseModel // id/created_at/updated_at/deleted_at
	RuleName    string `gorm:"size:128;not null" json:"rule_name"`                       // 规则名称
	RuleType    string `gorm:"size:32;not null;index" json:"rule_type"`                  // 类型 sensitive_word/price_check/frequency/prohibited/content
	RuleKey     string `gorm:"size:64;index" json:"rule_key"`                            // 规则键（如敏感词级别：politics/porn/violence）
	Pattern     string `gorm:"type:text" json:"pattern"`                                  // 匹配模式（正则/关键词列表 JSON）
	Threshold   JSONB  `gorm:"type:jsonb" json:"threshold"`                                // 阈值配置 {min_price,max_price,max_count_per_min...}
	Action      string `gorm:"size:32;default:'reject'" json:"action"`                   // 触发动作：reject/approval/warning/manual
	PenaltyType string `gorm:"size:32" json:"penalty_type"`                               // 处罚类型 warning/limit/ban1d/ban7d/banForever
	Severity    int    `gorm:"default:1;index" json:"severity"`                            // 严重等级 1-5
	Status      int    `gorm:"default:1;index" json:"status"`                              // 0禁用 1启用
	Description string `gorm:"size:500" json:"description"`                                // 规则描述
	Sort        int    `gorm:"default:0" json:"sort"`                                       // 排序
}

// TableName 表名（ers_ 前缀）
func (ErshouAuditRule) TableName() string { return "ers_audit_rules" }

// === 用户信用等级常量 ===
const (
	CreditLevelNewbie   = 0 // 新手
	CreditLevelBronze   = 1 // 青铜
	CreditLevelSilver   = 2 // 白银
	CreditLevelGold     = 3 // 黄金
	CreditLevelPlatinum = 4 // 铂金
	CreditLevelDiamond  = 5 // 钻石
)

// ErshouUserCredit 用户信用表（对标转转）
type ErshouUserCredit struct {
	database.BaseModel // id/created_at/updated_at/deleted_at
	UserID            uint   `gorm:"not null;uniqueIndex" json:"user_id"`                     // 用户ID（一对一）
	CreditScore       int    `gorm:"default:100;index" json:"credit_score"`                    // 信用分 0-1000（初始 100，<30 限制交易）
	CreditLevel       int    `gorm:"default:0;index" json:"credit_level"`                       // 信用等级 0新手 1青铜 2白银 3黄金 4铂金 5钻石
	TotalTransactions int    `gorm:"default:0" json:"total_transactions"`                     // 历史交易总数
	SuccessTransactions int  `gorm:"default:0" json:"success_transactions"`                   // 成功交易数
	CancelTransactions int  `gorm:"default:0" json:"cancel_transactions"`                     // 取消交易数
	GoodReviews       int    `gorm:"default:0" json:"good_reviews"`                            // 好评数（4-5星）
	MediumReviews     int    `gorm:"default:0" json:"medium_reviews"`                          // 中评数（3星）
	BadReviews        int    `gorm:"default:0" json:"bad_reviews"`                              // 差评数（1-2星）
	GoodRate          float64 `gorm:"type:decimal(5,2);default:100.00" json:"good_rate"`       // 好评率
	Disputes          int    `gorm:"default:0" json:"disputes"`                                  // 纠纷数
	Reports           int    `gorm:"default:0" json:"reports"`                                  // 被举报次数
	Penalties         int    `gorm:"default:0" json:"penalties"`                                // 处罚次数
	LastTransactionAt *time.Time `gorm:"index" json:"last_transaction_at"`                       // 最近交易时间
	FrozenReason      string `gorm:"size:500" json:"frozen_reason"`                            // 冻结原因
	FrozenUntil      *time.Time `gorm:"index" json:"frozen_until"`                            // 冻结截止时间
}

// TableName 表名（ers_ 前缀）
func (ErshouUserCredit) TableName() string { return "ers_user_credit" }
