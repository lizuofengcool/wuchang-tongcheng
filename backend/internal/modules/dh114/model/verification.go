// Package model 商户认证表（营业执照 + 实地认证 + 品牌授权）
// 对标大众点评/美团商户认证
package model

import (
	"time"

	"wuchang-tongcheng/internal/pkg/database"
)

// Dh114Verification 商户认证表
type Dh114Verification struct {
	database.RegionBaseModel // 含 id/region_id/created_at/updated_at/deleted_at
	VerificationNo string `gorm:"size:64;not null;uniqueIndex:uniq_dh114_verifications_no" json:"verification_no"` // 认证单号
	Dh114ID        uint   `gorm:"not null;index" json:"dh114_id"`                                            // 商户 ID
	BusinessID     uint   `gorm:"not null;default:0;index" json:"business_id"`                              // 商户详情 ID
	UserID         uint   `gorm:"not null;index" json:"user_id"`                                            // 申请用户 ID

	// === 认证类型 ===
	VerificationType string `gorm:"size:32;not null;default:'business_license';index" json:"verification_type"` // business_license/field/brand/license_field

	// === 营业执照信息 ===
	LicenseNo         string `gorm:"size:64;not null;default:'';index" json:"license_no"`         // 营业执照号
	LicenseImage      string `gorm:"size:255;not null;default:''" json:"license_image"`           // 营业执照图片
	BusinessName      string `gorm:"size:200;not null;default:''" json:"business_name"`           // 营业执照名称
	LegalPerson       string `gorm:"size:64;not null;default:''" json:"legal_person"`             // 法人代表
	LegalPersonIDCard string `gorm:"size:32;not null;default:''" json:"legal_person_id_card"`     // 法人身份证号
	BusinessScope     string `gorm:"type:text" json:"business_scope"`                              // 经营范围
	RegisteredAddress string `gorm:"size:500;not null;default:''" json:"registered_address"`      // 注册地址

	// === 实地认证信息 ===
	FieldPhotos    JSONB     `gorm:"type:jsonb" json:"field_photos"`                               // 实地照片 JSON
	FieldAddress  string    `gorm:"size:500;not null;default:''" json:"field_address"`            // 实地地址
	FieldLongitude float64  `gorm:"type:decimal(10,7);default:0" json:"field_longitude"`           // 实地经度
	FieldLatitude  float64  `gorm:"type:decimal(10,7);default:0" json:"field_latitude"`            // 实地纬度
	FieldDate      *time.Time `gorm:"type:date" json:"field_date"`                                // 实地考察日期
	InspectorID    uint      `gorm:"not null;default:0" json:"inspector_id"`                      // 考察员 ID
	InspectorName  string   `gorm:"size:50;not null;default:''" json:"inspector_name"`           // 考察员姓名

	// === 品牌授权信息 ===
	BrandName       string `gorm:"size:128;not null;default:''" json:"brand_name"`               // 品牌名称
	BrandAuthImage string `gorm:"size:255;not null;default:''" json:"brand_auth_image"`         // 品牌授权书图片

	// === 状态 ===
	Status      int        `gorm:"default:0;index" json:"status"`                                // 0待审 1通过 2拒绝 3过期
	AuditRemark string     `gorm:"size:500;not null;default:''" json:"audit_remark"`            // 审核备注
	AuditedBy   uint       `gorm:"not null;default:0;index" json:"audited_by"`                   // 审核人 ID
	AuditedAt   *time.Time `gorm:"index" json:"audited_at"`                                     // 审核时间
	ValidUntil  *time.Time `gorm:"type:date;index" json:"valid_until"`                          // 认证有效期
}

// TableName 表名
func (Dh114Verification) TableName() string { return "dh114_verifications" }
