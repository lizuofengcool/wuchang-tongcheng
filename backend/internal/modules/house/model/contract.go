// Package model 合同电子化表（对标贝壳/链家）
// 租约/买卖合同/电子签/三方协议/状态机
package model

import (
	"time"

	"wuchang-tongcheng/internal/pkg/database"
)

// === 合同类型常量 ===
const (
	ContractTypeRent      = "rent"      // 租赁合同
	ContractTypeSale      = "sale"      // 买卖合同
	ContractTypeDeposit   = "deposit"   // 定金合同
	ContractTypeAgency    = "agency"    // 居间合同
	ContractTypeTransfer  = "transfer"  // 转让合同
)

// === 合同状态常量 ===
const (
	ContractStatusDraft        = 0 // 草稿
	ContractStatusPendingSign  = 1 // 待签署
	ContractStatusPartSigned   = 2 // 部分签署
	ContractStatusSigned       = 3 // 已签署
	ContractStatusEffective    = 4 // 已生效
	ContractStatusTerminated   = 5 // 已终止
	ContractStatusArchived     = 6 // 已归档
	ContractStatusCanceled     = 7 // 已取消
)

// === 签署方式常量 ===
const (
	SignMethodOnline  = "online"  // 在线电子签
	SignMethodOffline = "offline" // 线下签署
	SignMethodMixed   = "mixed"   // 混合
)

// === 佣金支付方常量 ===
const (
	CommissionPayerBoth    = "both"    // 双方
	CommissionPayerBuyer   = "buyer"   // 买方/承租方
	CommissionPayerSeller  = "seller"  // 卖方/出租方
	CommissionPayerAgent   = "agent"   // 经纪人
)

// HouseContract 合同电子化表
type HouseContract struct {
	database.RegionBaseModel // id/region_id/created_at/updated_at/deleted_at
	ContractNo       string     `gorm:"size:64;not null;uniqueIndex" json:"contract_no"`                       // 合同编号
	ContractType     string     `gorm:"size:32;not null;default:'rent';index" json:"contract_type"`            // rent/sale/deposit/agency/transfer
	HouseID          uint       `gorm:"not null;index" json:"house_id"`                                       // 关联房源 ID
	ListingID        uint       `gorm:"not null;default:0;index" json:"listing_id"`                           // 关联发布 ID
	CommunityID      uint       `gorm:"not null;default:0;index" json:"community_id"`                         // 关联小区 ID
	PartyAID         uint       `gorm:"not null;index:idx_house_contracts_party_a" json:"party_a_id"`          // 甲方 ID
	PartyAName       string     `gorm:"size:50;not null;default:''" json:"party_a_name"`                      // 甲方姓名
	PartyAPhone      string     `gorm:"size:20;not null;default:''" json:"party_a_phone"`                     // 甲方手机
	PartyAIDCard     string     `gorm:"size:32;not null;default:''" json:"party_a_id_card"`                   // 甲方身份证
	PartyBID         uint       `gorm:"not null;index:idx_house_contracts_party_b" json:"party_b_id"`          // 乙方 ID
	PartyBName       string     `gorm:"size:50;not null;default:''" json:"party_b_name"`                      // 乙方姓名
	PartyBPhone      string     `gorm:"size:20;not null;default:''" json:"party_b_phone"`                     // 乙方手机
	PartyBIDCard     string     `gorm:"size:32;not null;default:''" json:"party_b_id_card"`                   // 乙方身份证
	AgentID          uint       `gorm:"not null;default:0;index" json:"agent_id"`                             // 经纪人 ID
	AgentName        string     `gorm:"size:50;not null;default:''" json:"agent_name"`                        // 经纪人姓名
	Title            string     `gorm:"size:200;not null;default:''" json:"title"`                            // 合同标题
	Content          string     `gorm:"type:text" json:"content"`                                              // 合同正文
	Attachments      JSONB      `gorm:"type:jsonb" json:"attachments"`                                         // 附件
	Amount           float64    `gorm:"type:decimal(14,2);default:0" json:"amount"`                           // 合同金额
	Deposit          float64    `gorm:"type:decimal(14,2);default:0" json:"deposit"`                          // 押金/定金
	Commission       float64    `gorm:"type:decimal(14,2);default:0" json:"commission"`                       // 佣金
	CommissionPayer  string     `gorm:"size:16;not null;default:'both'" json:"commission_payer"`              // 佣金支付方
	StartDate        *time.Time `gorm:"type:date;index" json:"start_date"`                                    // 起始日期
	EndDate          *time.Time `gorm:"type:date;index" json:"end_date"`                                      // 终止日期
	PaymentMethod    string     `gorm:"size:32;not null;default:'monthly'" json:"payment_method"`             // 付款方式
	SignMethod       string     `gorm:"size:32;not null;default:'online'" json:"sign_method"`                 // 签署方式
	PartyASignedAt   *time.Time `gorm:"index" json:"party_a_signed_at"`                                        // 甲方签署时间
	PartyBSignedAt   *time.Time `gorm:"index" json:"party_b_signed_at"`                                        // 乙方签署时间
	AgentSignedAt    *time.Time `gorm:"index" json:"agent_signed_at"`                                          // 经纪人签署时间
	Status           int        `gorm:"default:0;index:idx_house_contracts_party_a;index:idx_house_contracts_party_b;index" json:"status"` // 0草稿 1待签 2部分签 3已签 4生效 5终止 6归档 7取消
	EffectiveAt      *time.Time `gorm:"index" json:"effective_at"`                                             // 生效时间
	TerminatedAt     *time.Time `gorm:"index" json:"terminated_at"`                                            // 终止时间
	TerminatedReason string     `gorm:"size:500;not null;default:''" json:"terminated_reason"`                // 终止原因
	ArchivedAt       *time.Time `gorm:"index" json:"archived_at"`                                              // 归档时间
}

// TableName 表名（house_ 前缀）
func (HouseContract) TableName() string { return "house_contracts" }
