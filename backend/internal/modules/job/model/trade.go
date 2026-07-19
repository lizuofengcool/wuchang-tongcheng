// Package model 担保交易 + AI 推荐（对标 BOSS直聘）
// 招聘保证金/中介费托管 + AI 智能推荐（人岗匹配）
package model

import (
	"time"

	"wuchang-tongcheng/internal/pkg/database"
)

// === 担保类型常量 ===
const (
	EscrowTypeRecruitmentDeposit = "recruitment_deposit" // 招聘保证金
	EscrowTypeAgencyFee          = "agency_fee"          // 中介费托管
	EscrowTypeRecommendFee       = "recommend_fee"       // 推荐费
	EscrowTypeOnboardingBonus    = "onboarding_bonus"    // 入职奖金
	EscrowTypeDeposit            = "deposit"              // 押金
	EscrowTypePerformance         = "performance"          // 绩效奖金
)

// === 担保状态常量 ===
const (
	EscrowStatusNone       = 0 // 未启用
	EscrowStatusPending    = 1 // 待支付
	EscrowStatusFrozen     = 2 // 资金冻结中
	EscrowStatusReleased   = 3 // 资金已解冻（待放款）
	EscrowStatusPaid        = 4 // 已放款
	EscrowStatusRefunded   = 5 // 已退款
	EscrowStatusDispute    = 6 // 纠纷处理中
	EscrowStatusCanceled  = 7 // 已取消
	EscrowStatusArbitrated = 8 // 仲裁完成
)

// === 支付方式常量 ===
const (
	PayMethodWechat = "wechat" // 微信支付
	PayMethodAlipay = "alipay" // 支付宝
	PayMethodBank   = "bank"   // 银行转账
	PayMethodBalance = "balance" // 余额支付
	PayMethodOther  = "other"   // 其他
)

// JobEscrow 担保交易表
type JobEscrow struct {
	database.RegionBaseModel // id/region_id/created_at/updated_at/deleted_at
	EscrowNo        string     `gorm:"size:64;not null;uniqueIndex" json:"escrow_no"`        // 担保单号
	EscrowType      string     `gorm:"size:32;default:'recruitment_deposit';index" json:"escrow_type"` // 担保类型
	JobID           uint       `gorm:"index" json:"job_id"`                                  // 关联职位 ID
	ApplicationID   uint       `gorm:"index" json:"application_id"`                          // 关联投递记录 ID
	CompanyID       uint       `gorm:"index" json:"company_id"`                              // 所属公司 ID
	PayerID         uint       `gorm:"not null;index" json:"payer_id"`                       // 付款方 ID
	PayeeID         uint       `gorm:"not null;index" json:"payee_id"`                       // 收款方 ID
	Amount          float64    `gorm:"type:decimal(12,2);not null" json:"amount"`             // 担保金额
	PlatformFee     float64    `gorm:"type:decimal(12,2);default:0" json:"platform_fee"`      // 平台手续费
	PayeeAmount     float64    `gorm:"type:decimal(12,2);default:0" json:"payee_amount"`    // 收款方到账金额
	Status          int        `gorm:"default:1;index" json:"status"`                         // 9 状态机
	PayMethod       string     `gorm:"size:32;default:'wechat'" json:"pay_method"`           // 支付方式
	PayTradeNo      string     `gorm:"size:128;index" json:"pay_trade_no"`                  // 支付流水号
	PaidAt          *time.Time `gorm:"index" json:"paid_at"`                                  // 支付时间
	FrozenAt        *time.Time `gorm:"index" json:"frozen_at"`                                // 冻结时间
	ReleaseAt       *time.Time `gorm:"index" json:"release_at"`                              // 解冻时间
	RefundedAt      *time.Time `gorm:"index" json:"refunded_at"`                              // 退款时间
	AutoReleaseAt   *time.Time `gorm:"index" json:"auto_release_at"`                         // 自动放款时间
	DisputeReason   string     `gorm:"size:500" json:"dispute_reason"`                        // 纠纷原因
	ArbitrationResult string   `gorm:"type:text" json:"arbitration_result"`                  // 仲裁结果
	CompletedAt     *time.Time `gorm:"index" json:"completed_at"`                             // 完成时间
}

// TableName 表名（job_ 前缀）
func (JobEscrow) TableName() string { return "job_escrows" }

// === 推荐类型常量 ===
const (
	RecTypeJobToUser    = "job_to_user"    // 职位推荐给求职者
	RecTypeUserToJob    = "user_to_job"    // 求职者推荐给招聘者
	RecTypeCompanyToUser = "company_to_user" // 公司推荐给求职者
	RecTypeUserToCompany = "user_to_company" // 求职者推荐给公司
)

// === 推荐来源常量 ===
const (
	RecSourceAI       = "ai"        // AI 推荐
	RecSourceBehavior = "behavior" // 行为推荐
	RecSourceSimilar  = "similar"   // 相似推荐
	RecSourceHot      = "hot"        // 热门推荐
	RecSourceManual   = "manual"    // 人工推荐
)

// === 推荐状态常量 ===
const (
	RecStatusPending   = 0 // 待展示
	RecStatusShown     = 1 // 已展示
	RecStatusClicked   = 2 // 已点击
	RecStatusApplied   = 3 // 已投递
	RecStatusDismissed = 4 // 已忽略
	RecStatusExpired   = 5 // 已过期
)

// JobRecommendation AI 智能推荐记录表（对标 BOSS直聘）
type JobRecommendation struct {
	database.RegionBaseModel // id/region_id/created_at/updated_at/deleted_at
	UserID         uint       `gorm:"not null;index;uniqueIndex:uniq_job_recs_user_job_type" json:"user_id"` // 推荐给的用户 ID
	JobID          uint       `gorm:"not null;index;uniqueIndex:uniq_job_recs_user_job_type" json:"job_id"`  // 关联职位 ID
	RecType        string     `gorm:"size:32;default:'job_to_user';index;uniqueIndex:uniq_job_recs_user_job_type" json:"rec_type"` // 推荐类型
	Source         string     `gorm:"size:32;default:'ai';index" json:"source"`                              // 推荐来源
	Score          float64    `gorm:"type:decimal(5,2);default:0;index" json:"score"`                        // 综合评分
	Reason         string     `gorm:"size:500" json:"reason"`                                              // 推荐理由
	PositionMatch  float64    `gorm:"type:decimal(5,2);default:0" json:"position_match"`                   // 职位匹配度
	SalaryMatch    float64    `gorm:"type:decimal(5,2);default:0" json:"salary_match"`                     // 薪资匹配度
	LocationMatch  float64    `gorm:"type:decimal(5,2);default:0" json:"location_match"`                   // 地点匹配度
	SkillMatch     float64    `gorm:"type:decimal(5,2);default:0" json:"skill_match"`                     // 技能匹配度
	Status         int        `gorm:"default:0;index" json:"status"`                                         // 推荐状态
	ClickedAt      *time.Time `gorm:"index" json:"clicked_at"`                                              // 点击时间
	AppliedAt      *time.Time `gorm:"index" json:"applied_at"`                                              // 投递时间
	DismissedAt    *time.Time `gorm:"index" json:"dismissed_at"`                                            // 忽略时间
	ExpiredAt      *time.Time `gorm:"index" json:"expired_at"`                                              // 过期时间
}

// TableName 表名（job_ 前缀）
func (JobRecommendation) TableName() string { return "job_recommendations" }
