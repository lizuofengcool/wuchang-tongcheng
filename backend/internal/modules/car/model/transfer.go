// Package model 过户办理表（对标瓜子）
// 材料/流程/状态/新牌照/新登记证
package model

import (
	"time"

	"wuchang-tongcheng/internal/pkg/database"
)

// === 过户状态常量 ===
const (
	TransferStatusPending    = 0 // 待提交
	TransferStatusSubmitted  = 1 // 已提交
	TransferStatusReviewing  = 2 // 审核中
	TransferStatusApproved   = 3 // 审核通过
	TransferStatusRejected   = 4 // 审核拒绝
	TransferStatusInProgress = 5 // 办理中
	TransferStatusCompleted  = 6 // 已完成
	TransferStatusCanceled   = 7 // 已取消
)

// === 过户类型常量 ===
const (
	TransferTypeSale       = "sale"       // 买卖过户
	TransferTypeReplace    = "replace"    // 置换过户
	TransferTypeGift       = "gift"       // 赠予过户
	TransferTypeInheritance = "inheritance" // 继承过户
	TransferTypeRelocation = "relocation" // 迁移过户
)

// CarTransfer 过户办理表
type CarTransfer struct {
	database.RegionBaseModel // id/region_id/created_at/updated_at/deleted_at
	TransferNo          string     `gorm:"size:64;not null;uniqueIndex" json:"transfer_no"`                // 过户单号
	CarID               uint       `gorm:"not null;index" json:"car_id"`                                   // 车源 ID
	ContractID          uint       `gorm:"not null;default:0;index" json:"contract_id"`                    // 合同 ID
	ListingID           uint       `gorm:"not null;default:0;index" json:"listing_id"`                     // 发布 ID
	SellerID            uint       `gorm:"not null;index" json:"seller_id"`                                // 卖方 ID
	SellerName          string     `gorm:"size:50;not null;default:''" json:"seller_name"`                 // 卖方姓名
	BuyerID             uint       `gorm:"not null;index" json:"buyer_id"`                                 // 买方 ID
	BuyerName           string     `gorm:"size:50;not null;default:''" json:"buyer_name"`                  // 买方姓名
	AgentID             uint       `gorm:"not null;default:0;index" json:"agent_id"`                       // 经办人 ID
	AgentName           string     `gorm:"size:50;not null;default:''" json:"agent_name"`                  // 经办人姓名
	TransferType        string     `gorm:"size:32;not null;default:'sale';index" json:"transfer_type"`     // sale/replace/gift/inheritance/relocation
	VehicleRegistration JSONB      `gorm:"type:jsonb" json:"vehicle_registration"`                         // 车辆登记信息 JSON
	Documents           JSONB      `gorm:"type:jsonb" json:"documents"`                                    // 过户材料 JSON
	TransferFee         float64    `gorm:"type:decimal(14,2);default:0" json:"transfer_fee"`               // 过户费
	TaxFee              float64    `gorm:"type:decimal(14,2);default:0" json:"tax_fee"`                    // 交易税
	OtherFee            float64    `gorm:"type:decimal(14,2);default:0" json:"other_fee"`                  // 其他费用
	Location            string     `gorm:"size:255;not null;default:''" json:"location"`                   // 办理地点
	AppointmentDate     *time.Time `gorm:"type:date;index" json:"appointment_date"`                        // 预约日期
	AppointmentTime     string     `gorm:"size:32;not null;default:''" json:"appointment_time"`            // 预约时段
	SubmittedAt         *time.Time `gorm:"index" json:"submitted_at"`                                      // 提交时间
	ReviewedAt          *time.Time `gorm:"index" json:"reviewed_at"`                                       // 审核时间
	CompletedAt         *time.Time `gorm:"index" json:"completed_at"`                                      // 完成时间
	CanceledAt          *time.Time `gorm:"index" json:"canceled_at"`                                       // 取消时间
	NewLicensePlate     string     `gorm:"size:32;not null;default:''" json:"new_license_plate"`           // 新车牌号
	NewRegistrationCert string     `gorm:"size:255;not null;default:''" json:"new_registration_cert"`      // 新登记证书 URL
	Status              int        `gorm:"default:0;index" json:"status"`                                  // 0待提交 1提交 2审核中 3通过 4拒绝 5办理中 6完成 7取消
	Remark              string     `gorm:"type:text" json:"remark"`                                        // 备注
}

// TableName 表名（car_ 前缀）
func (CarTransfer) TableName() string { return "car_transfer" }
