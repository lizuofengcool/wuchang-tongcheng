// Package model 举报 + 评价 + 审核规则表
// CarReport 举报；CarReview 评价；CarAuditRule 审核规则
package model

import (
	"time"

	"wuchang-tongcheng/internal/pkg/database"
)

// === 举报状态常量 ===
const (
	ReportStatusPending    = 0 // 待处理
	ReportStatusProcessing = 1 // 处理中
	ReportStatusHandled    = 2 // 已处理
	ReportStatusRejected   = 3 // 已驳回
	ReportStatusCanceled   = 4 // 已取消
)

// === 举报类型常量 ===
const (
	ReportTypeFakeCar       = "fake_car"        // 虚假车源
	ReportTypeFraud         = "fraud"           // 诈骗
	ReportTypePorn          = "porn"            // 色情
	ReportTypeInfringement  = "infringement"    // 侵权
	ReportTypeIllegalCar    = "illegal_car"     // 违法车辆
	ReportTypeFalseMileage  = "false_mileage"   // 虚假里程
	ReportTypeAccidentHide  = "accident_hide"   // 隐瞒事故
	ReportTypeOther         = "other"           // 其他
)

// === 举报目标类型常量 ===
const (
	ReportTargetTypeCar     = "car"     // 车源
	ReportTargetTypeListing = "listing" // 发布
	ReportTargetTypeDealer  = "dealer"  // 车商
	ReportTargetTypeUser    = "user"    // 用户
	ReportTargetTypeReview  = "review"  // 评价
)

// === 处罚类型常量 ===
const (
	PenaltyTypeWarning     = "warning"     // 警告
	PenaltyTypeBan24h      = "ban24h"      // 封禁 24 小时
	PenaltyTypeBan7d       = "ban7d"       // 封禁 7 天
	PenaltyTypeBan30d      = "ban30d"      // 封禁 30 天
	PenaltyTypeBanForever  = "ban_forever" // 永久封禁
	PenaltyTypeDeleteCar   = "delete_car"  // 删除车源
	PenaltyTypeLimit       = "limit"       // 限流
)

// CarReport 举报工单表
type CarReport struct {
	database.BaseModel // id/created_at/updated_at/deleted_at
	ReportNo          string     `gorm:"size:64;not null;uniqueIndex" json:"report_no"`                  // 举报单号
	TargetType        string     `gorm:"size:32;not null;index:idx_car_reports_target_type_target,priority:1" json:"target_type"` // car/listing/dealer/user/review
	TargetID          uint       `gorm:"not null;index:idx_car_reports_target_type_target,priority:2" json:"target_id"`         // 目标 ID
	TargetUserID      uint       `gorm:"not null;index" json:"target_user_id"`                           // 目标用户 ID
	ReporterID        uint       `gorm:"not null;index" json:"reporter_id"`                              // 举报人 ID
	ReporterName      string     `gorm:"size:50;not null;default:''" json:"reporter_name"`               // 举报人姓名
	ReportedUserID    uint       `gorm:"not null;index" json:"reported_user_id"`                         // 被举报人 ID
	ReportedUserName  string     `gorm:"size:50;not null;default:''" json:"reported_user_name"`          // 被举报人姓名
	ReportType        string     `gorm:"size:32;not null;index" json:"report_type"`                      // fake_car/fraud/porn/infringement 等
	Reason            string     `gorm:"size:500;not null" json:"reason"`                                // 举报原因
	Description       string     `gorm:"type:text" json:"description"`                                   // 详细描述
	EvidenceImages    JSONB      `gorm:"type:jsonb" json:"evidence_images"`                              // 证据图片
	Status            int        `gorm:"default:0;index" json:"status"`                                  // 0待处理 1处理中 2已处理 3驳回 4取消
	HandlerID         uint       `gorm:"not null;default:0;index" json:"handler_id"`                     // 处理人 ID
	HandlerName       string     `gorm:"size:50;not null;default:''" json:"handler_name"`                // 处理人姓名
	HandleResult      string     `gorm:"type:text" json:"handle_result"`                                 // 处理结果
	PenaltyType       string     `gorm:"size:32;not null;default:''" json:"penalty_type"`                // warning/ban24h/ban7d/ban30d/ban_forever/delete_car/limit
	PenaltyUserID     uint       `gorm:"not null;default:0;index" json:"penalty_user_id"`                // 被处罚用户 ID
	SLADeadline       *time.Time `gorm:"index" json:"sla_deadline"`                                      // SLA 截止时间
	HandledAt         *time.Time `gorm:"index" json:"handled_at"`                                        // 处理时间
	AppealReason      string     `gorm:"type:text" json:"appeal_reason"`                                 // 申诉原因
	AppealedAt        *time.Time `gorm:"index" json:"appealed_at"`                                       // 申诉时间
	AppealResult      string     `gorm:"type:text" json:"appeal_result"`                                 // 申诉结果
	AppealHandlerID   uint       `gorm:"not null;default:0" json:"appeal_handler_id"`                    // 申诉处理人 ID
	AppealHandledAt   *time.Time `gorm:"index" json:"appeal_handled_at"`                                 // 申诉处理时间
}

// TableName 表名（car_ 前缀）
func (CarReport) TableName() string { return "car_reports" }

// === 评价状态常量 ===
const (
	ReviewStatusPending  = 0 // 待审核
	ReviewStatusApproved = 1 // 已通过
	ReviewStatusRejected = 2 // 已拒绝
	ReviewStatusHidden   = 3 // 已隐藏
)

// === 评价目标类型常量 ===
const (
	ReviewTargetTypeDealer = "dealer" // 车商
	ReviewTargetTypeCar    = "car"    // 车源
	ReviewTargetTypeSales  = "sales"  // 销售
)

// === 评价者类型常量 ===
const (
	ReviewerTypeBuyer  = "buyer"  // 买家
	ReviewerTypeSeller = "seller" // 卖家
)

// CarReview 评价表
type CarReview struct {
	database.RegionBaseModel // id/region_id/created_at/updated_at/deleted_at
	TargetType        string     `gorm:"size:32;not null;default:'dealer';index:idx_car_reviews_target,priority:1" json:"target_type"` // dealer/car/sales
	TargetID          uint       `gorm:"not null;index:idx_car_reviews_target,priority:2" json:"target_id"`                            // 目标 ID
	ReviewerID        uint       `gorm:"not null;index" json:"reviewer_id"`                                                              // 评价人 ID
	ReviewerName      string     `gorm:"size:50;not null;default:''" json:"reviewer_name"`                                              // 评价人姓名
	ReviewerAvatar    string     `gorm:"size:255;not null;default:''" json:"reviewer_avatar"`                                           // 评价人头像
	ReviewType        string     `gorm:"size:32;not null;default:'buyer';index" json:"review_type"`                                     // buyer/seller
	Rating            int        `gorm:"not null;default:5;index" json:"rating"`                                                        // 总体评分 1-5
	Content           string     `gorm:"type:text" json:"content"`                                                                       // 评价内容
	Images            JSONB      `gorm:"type:jsonb" json:"images"`                                                                       // 评价图片
	VideoURL          string     `gorm:"size:255;not null;default:''" json:"video_url"`                                                 // 评价视频
	IsAnonymous       bool       `gorm:"not null;default:false" json:"is_anonymous"`                                                    // 匿名评价
	IsRecommended     bool       `gorm:"not null;default:true;index" json:"is_recommended"`                                             // 是否推荐
	Tags              JSONB      `gorm:"type:jsonb" json:"tags"`                                                                        // 标签
	DealAmount        float64    `gorm:"type:decimal(14,2);default:0" json:"deal_amount"`                                               // 成交金额
	ExteriorRating    int        `gorm:"not null;default:5" json:"exterior_rating"`                                                     // 外观评分
	InteriorRating    int        `gorm:"not null;default:5" json:"interior_rating"`                                                     // 内饰评分
	EngineRating      int        `gorm:"not null;default:5" json:"engine_rating"`                                                       // 发动机评分
	PaperworkRating   int        `gorm:"not null;default:5" json:"paperwork_rating"`                                                    // 手续评分
	ServiceAttitude   int        `gorm:"not null;default:5" json:"service_attitude"`                                                    // 服务态度
	ProfessionalSkill int        `gorm:"not null;default:5" json:"professional_skill"`                                                  // 专业能力
	Reply             string     `gorm:"type:text" json:"reply"`                                                                        // 回复内容
	ReplyAt           *time.Time `gorm:"index" json:"reply_at"`                                                                         // 回复时间
	AppendContent     string     `gorm:"type:text" json:"append_content"`                                                               // 追评内容
	AppendImages      JSONB      `gorm:"type:jsonb" json:"append_images"`                                                               // 追评图片
	AppendAt          *time.Time `gorm:"index" json:"append_at"`                                                                        // 追评时间
	LikeCount         int        `gorm:"not null;default:0" json:"like_count"`                                                         // 点赞数
	Status            int        `gorm:"default:1;index" json:"status"`                                                                 // 0待审 1通过 2拒绝 3隐藏
}

// TableName 表名（car_ 前缀）
func (CarReview) TableName() string { return "car_reviews" }

// === 审核规则类型常量 ===
const (
	AuditRuleTypeSensitiveWord = "sensitive_word" // 敏感词
	AuditRuleTypePriceCheck    = "price_check"    // 价格异常
	AuditRuleTypeFrequency     = "frequency"      // 频率限制
	AuditRuleTypeFakeCar       = "fake_car"       // 虚假车源
	AuditRuleTypeMileageCheck  = "mileage_check"  // 里程异常
	AuditRuleTypeVINCheck      = "vin_check"      // 车架号校验
	AuditRuleTypeProhibited    = "prohibited"     // 违禁内容
	AuditRuleTypeContact       = "contact"        // 联系方式校验
)

// === 审核动作常量 ===
const (
	AuditActionReject    = "reject"    // 拒绝
	AuditActionApproval  = "approval"  // 转人工
	AuditActionFilter    = "filter"    // 过滤敏感词
	AuditActionLimit     = "limit"     // 限流
)

// CarAuditRule 审核规则表
type CarAuditRule struct {
	database.BaseModel // id/created_at/updated_at/deleted_at
	RuleName    string `gorm:"size:128;not null" json:"rule_name"`                              // 规则名
	RuleType    string `gorm:"size:32;not null;index" json:"rule_type"`                         // sensitive_word/price_check/frequency/fake_car/mileage_check/vin_check/prohibited/contact
	RuleKey     string `gorm:"size:64;not null;default:'';index" json:"rule_key"`               // 规则键
	Pattern     string `gorm:"type:text" json:"pattern"`                                        // 正则/关键词模式
	Threshold   JSONB  `gorm:"type:jsonb" json:"threshold"`                                     // 阈值 JSON
	Action      string `gorm:"size:32;not null;default:'reject'" json:"action"`                 // reject/approval/filter/limit
	PenaltyType string `gorm:"size:32;not null;default:''" json:"penalty_type"`                // 处罚类型
	Severity    int    `gorm:"not null;default:1;index" json:"severity"`                        // 严重等级 1-5
	Status      int    `gorm:"default:1;index" json:"status"`                                   // 0禁用 1启用
	Description string `gorm:"size:500;not null;default:''" json:"description"`                 // 描述
	Sort        int    `gorm:"not null;default:0" json:"sort"`                                  // 排序
}

// TableName 表名（car_ 前缀）
func (CarAuditRule) TableName() string { return "car_audit_rules" }
