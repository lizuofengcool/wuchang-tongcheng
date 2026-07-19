// Package model 举报 + 评价 + 审核规则表
// HouseReport 举报工单；HouseReview 评价；HouseAuditRule 审核规则
package model

import (
	"time"

	"wuchang-tongcheng/internal/pkg/database"
)

// === 举报目标类型常量 ===
const (
	ReportTargetTypeHouse      = "house"      // 举报房源
	ReportTargetTypeListing    = "listing"    // 举报发布
	ReportTargetTypeAgent      = "agent"      // 举报经纪人
	ReportTargetTypeCommunity  = "community"  // 举报小区
	ReportTargetTypeReview     = "review"     // 举报评价
)

// === 举报类型常量 ===
const (
	ReportTypeFakeHouse       = "fake_house"        // 虚假房源
	ReportTypePriceFraud      = "price_fraud"       // 价格欺诈
	ReportTypeSoldStillListed = "sold_still_listed" // 已售仍挂牌
	ReportTypeContactInvalid  = "contact_invalid"   // 联系方式无效
	ReportTypeScam            = "scam"              // 诈骗
	ReportTypePorn            = "porn"              // 色情
	ReportTypeIllegal         = "illegal"           // 违法
	ReportTypeInfringement    = "infringement"      // 侵权
	ReportTypeOther           = "other"             // 其他
)

// === 举报状态常量 ===
const (
	ReportStatusPending      = 0 // 待处理
	ReportStatusProcessing   = 1 // 处理中
	ReportStatusHandled      = 2 // 已处理
	ReportStatusRejected     = 3 // 已驳回
	ReportStatusAppealed     = 4 // 申诉中
	ReportStatusAppealHandled = 5 // 申诉已处理
)

// === 处罚类型常量 ===
const (
	PenaltyTypeNone        = ""            // 无处罚
	PenaltyTypeWarning     = "warning"     // 警告
	PenaltyTypeRemove      = "remove"      // 下架
	PenaltyTypeBan7d       = "ban7d"       // 封禁 7 天
	PenaltyTypeBan30d      = "ban30d"      // 封禁 30 天
	PenaltyTypeBanForever  = "ban_forever" // 永久封禁
)

// === 评价目标类型常量 ===
const (
	ReviewTargetTypeAgent     = "agent"     // 评价经纪人
	ReviewTargetTypeCommunity = "community" // 评价小区
	ReviewTargetTypeHouse     = "house"     // 评价房源
)

// === 评价者类型常量 ===
const (
	ReviewTypeTenant  = "tenant"  // 租客
	ReviewTypeBuyer   = "buyer"   // 买家
	ReviewTypeSeller  = "seller"  // 卖家
	ReviewTypeLandlord = "landlord" // 房东
)

// === 评价状态常量 ===
const (
	ReviewStatusHidden   = 0 // 隐藏
	ReviewStatusVisible  = 1 // 显示
	ReviewStatusReported = 2 // 被举报
)

// === 审核规则类型常量 ===
const (
	RuleTypeSensitiveWord = "sensitive_word" // 敏感词
	RuleTypePriceCheck    = "price_check"    // 价格异常
	RuleTypeFrequency     = "frequency"      // 频率限制
	RuleTypeFakeHouse     = "fake_house"     // 虚假房源
	RuleTypeProhibited    = "prohibited"     // 违禁内容
	RuleTypeContact       = "contact"        // 联系方式校验
)

// === 审核动作常量 ===
const (
	RuleActionReject    = "reject"    // 拒绝
	RuleActionApproval  = "approval"  // 转人工审核
	RuleActionLimit     = "limit"     // 限流
	RuleActionWarning   = "warning"   // 警告
)

// HouseReport 举报工单表
type HouseReport struct {
	database.BaseModel // id/created_at/updated_at/deleted_at
	ReportNo          string     `gorm:"size:64;not null;uniqueIndex" json:"report_no"`                       // 举报单号
	TargetType        string     `gorm:"size:32;not null;index:idx_house_reports_target_type_target" json:"target_type"` // house/listing/agent/community/review
	TargetID          uint       `gorm:"not null;index:idx_house_reports_target_type_target" json:"target_id"` // 目标 ID
	TargetUserID      uint       `gorm:"not null;index" json:"target_user_id"`                                // 目标用户 ID
	ReporterID        uint       `gorm:"not null;index" json:"reporter_id"`                                   // 举报人 ID
	ReporterName      string     `gorm:"size:50;not null;default:''" json:"reporter_name"`                    // 举报人姓名
	ReportedUserID    uint       `gorm:"not null;index" json:"reported_user_id"`                              // 被举报用户 ID
	ReportedUserName  string     `gorm:"size:50;not null;default:''" json:"reported_user_name"`               // 被举报用户姓名
	ReportType        string     `gorm:"size:32;not null;index" json:"report_type"`                           // fake_house/price_fraud/scam/porn/illegal 等
	Reason            string     `gorm:"size:500;not null" json:"reason"`                                     // 举报原因
	Description       string     `gorm:"type:text" json:"description"`                                        // 详细描述
	EvidenceImages    JSONB      `gorm:"type:jsonb" json:"evidence_images"`                                   // 证据图片
	Status            int        `gorm:"default:0;index" json:"status"`                                       // 0待处理 1处理中 2已处理 3驳回 4申诉中 5申诉处理
	HandlerID         uint       `gorm:"not null;default:0;index" json:"handler_id"`                          // 处理人 ID
	HandlerName       string     `gorm:"size:50;not null;default:''" json:"handler_name"`                     // 处理人姓名
	HandleResult      string     `gorm:"type:text" json:"handle_result"`                                      // 处理结果
	PenaltyType       string     `gorm:"size:32;not null;default:''" json:"penalty_type"`                     // 处罚类型
	PenaltyUserID     uint       `gorm:"not null;default:0;index" json:"penalty_user_id"`                     // 被处罚用户 ID
	SLADeadline       *time.Time `gorm:"index" json:"sla_deadline"`                                           // SLA 截止时间
	HandledAt         *time.Time `gorm:"index" json:"handled_at"`                                             // 处理时间
	AppealReason      string     `gorm:"type:text" json:"appeal_reason"`                                      // 申诉理由
	AppealedAt        *time.Time `gorm:"index" json:"appealed_at"`                                            // 申诉时间
	AppealResult      string     `gorm:"type:text" json:"appeal_result"`                                      // 申诉结果
	AppealHandlerID   uint       `gorm:"not null;default:0" json:"appeal_handler_id"`                         // 申诉处理人 ID
	AppealHandledAt   *time.Time `gorm:"index" json:"appeal_handled_at"`                                      // 申诉处理时间
}

// TableName 表名（house_ 前缀）
func (HouseReport) TableName() string { return "house_reports" }

// HouseReview 评价表
type HouseReview struct {
	database.RegionBaseModel // id/region_id/created_at/updated_at/deleted_at
	TargetType        string     `gorm:"size:32;not null;default:'agent';index:idx_house_reviews_target" json:"target_type"` // agent/community/house
	TargetID          uint       `gorm:"not null;index:idx_house_reviews_target" json:"target_id"`                // 目标 ID
	ReviewerID        uint       `gorm:"not null;index" json:"reviewer_id"`                                       // 评价人 ID
	ReviewerName      string     `gorm:"size:50;not null;default:''" json:"reviewer_name"`                       // 评价人姓名
	ReviewerAvatar    string     `gorm:"size:255;not null;default:''" json:"reviewer_avatar"`                    // 评价人头像
	ReviewType        string     `gorm:"size:32;not null;default:'tenant';index" json:"review_type"`             // tenant/buyer/seller/landlord
	Rating            int        `gorm:"not null;default:5;index" json:"rating"`                                 // 总评分 1-5
	Content           string     `gorm:"type:text" json:"content"`                                                // 评价内容
	Images            JSONB      `gorm:"type:jsonb" json:"images"`                                                // 评价图片
	VideoURL          string     `gorm:"size:255;not null;default:''" json:"video_url"`                          // 评价视频
	IsAnonymous       bool       `gorm:"not null;default:false" json:"is_anonymous"`                             // 是否匿名
	IsRecommended     bool       `gorm:"not null;default:true;index" json:"is_recommended"`                      // 是否推荐
	Tags              JSONB      `gorm:"type:jsonb" json:"tags"`                                                  // 标签
	DealAmount        float64    `gorm:"type:decimal(14,2);default:0" json:"deal_amount"`                        // 成交金额
	ServiceAttitude   int        `gorm:"not null;default:5" json:"service_attitude"`                             // 服务态度评分
	ProfessionalSkill int        `gorm:"not null;default:5" json:"professional_skill"`                           // 专业能力评分
	Reply             string     `gorm:"type:text" json:"reply"`                                                  // 回复内容
	ReplyAt           *time.Time `gorm:"index" json:"reply_at"`                                                   // 回复时间
	AppendContent     string     `gorm:"type:text" json:"append_content"`                                         // 追评内容
	AppendImages      JSONB      `gorm:"type:jsonb" json:"append_images"`                                         // 追评图片
	AppendAt          *time.Time `gorm:"index" json:"append_at"`                                                  // 追评时间
	LikeCount         int        `gorm:"not null;default:0" json:"like_count"`                                   // 点赞数
	Status            int        `gorm:"default:1;index" json:"status"`                                           // 0隐藏 1显示 2被举报
}

// TableName 表名（house_ 前缀）
func (HouseReview) TableName() string { return "house_reviews" }

// HouseAuditRule 审核规则表
type HouseAuditRule struct {
	database.BaseModel // id/created_at/updated_at/deleted_at
	RuleName    string `gorm:"size:128;not null" json:"rule_name"`                              // 规则名
	RuleType    string `gorm:"size:32;not null;index" json:"rule_type"`                        // sensitive_word/price_check/frequency/fake_house/prohibited/contact
	RuleKey     string `gorm:"size:64;not null;default:'';index" json:"rule_key"`              // 规则键
	Pattern     string `gorm:"type:text" json:"pattern"`                                        // 匹配模式
	Threshold   JSONB  `gorm:"type:jsonb" json:"threshold"`                                     // 阈值 JSON
	Action      string `gorm:"size:32;not null;default:'reject'" json:"action"`                // reject/approval/limit/warning
	PenaltyType string `gorm:"size:32;not null;default:''" json:"penalty_type"`               // 处罚类型
	Severity    int    `gorm:"not null;default:1;index" json:"severity"`                        // 严重等级 1-5
	Status      int    `gorm:"default:1;index" json:"status"`                                   // 0禁用 1启用
	Description string `gorm:"size:500;not null;default:''" json:"description"`               // 描述
	Sort        int    `gorm:"not null;default:0;index" json:"sort"`                            // 排序
}

// TableName 表名（house_ 前缀）
func (HouseAuditRule) TableName() string { return "house_audit_rules" }

// ReviewStatsData 评价统计聚合数据（不落库，仅用于聚合查询返回）
type ReviewStatsData struct {
	TotalReviews int64   `json:"total_reviews"` // 总评价数
	AvgRating    float64 `json:"avg_rating"`    // 平均评分
	GoodRate     float64 `json:"good_rate"`     // 好评率（rating >= 4）
	MediumRate   float64 `json:"medium_rate"`   // 中评率（rating == 3）
	BadRate      float64 `json:"bad_rate"`      // 差评率（rating <= 2）
}
