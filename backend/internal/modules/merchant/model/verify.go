// Package model 商户中台 - 商户认证表（merchant_verifications）
// 对标美团/大众点评商户资质认证
// 支持企业认证（business）与个人认证（personal）
package model

import (
	"time"

	"wuchang-tongcheng/internal/pkg/database"
)

// Verification 商户认证表
type Verification struct {
	database.RegionBaseModel // 含 id/region_id/created_at/updated_at/deleted_at（地区隔离）

	// === 关联 ===
	ShopID uint   `gorm:"index;not null" json:"shop_id"` // 所属商户 ID

	// === 认证类型 ===
	Type string `gorm:"size:20;not null;default:'business';index" json:"type"` // business/personal

	// === 营业执照信息 ===
	LicenseNo     string `gorm:"size:64;not null;default:'';index" json:"license_no"`     // 营业执照号
	LicenseImage  string `gorm:"size:255;not null;default:''" json:"license_image"`       // 营业执照图片
	LegalPerson   string `gorm:"size:64;not null;default:''" json:"legal_person"`         // 法人代表
	LegalPersonID string `gorm:"size:32;not null;default:''" json:"legal_person_id"`       // 法人身份证号

	// === 状态 ===
	Status      int        `gorm:"default:0;index" json:"status"`        // 0待审 1通过 2拒绝
	AuditReason string     `gorm:"size:500;not null;default:''" json:"audit_reason"` // 审核备注
	AuditedAt   *time.Time `gorm:"index" json:"audited_at"`               // 审核时间
}

// TableName 表名
func (Verification) TableName() string { return "merchant_verifications" }
