// Package model 企业认证表（对标 BOSS直聘）
// 营业执照/法人/认证状态/有效期
package model

import (
	"time"

	"wuchang-tongcheng/internal/pkg/database"
)

// === 认证类型常量 ===
const (
	CertTypeBusinessLicense     = "business_license"     // 营业执照
	CertTypeLegalPersonID       = "legal_person_id"      // 法人身份证
	CertTypeOrganizationCode    = "organization_code"    // 组织机构代码
	CertTypeTaxRegistration     = "tax_registration"     // 税务登记
	CertTypeOpeningPermit       = "opening_permit"        // 开户许可证
	CertTypeSocialCreditCode   = "social_credit_code"    // 统一社会信用代码
	CertTypeIndustryLicense    = "industry_license"      // 行业许可证
	CertTypeBrandAuthorization = "brand_authorization"   // 品牌授权书
)

// === 认证状态常量 ===
const (
	CertStatusPending  = 0 // 待审核
	CertStatusApproved = 1 // 已通过
	CertStatusRejected = 2 // 已拒绝
	CertStatusExpired  = 3 // 已过期
	CertStatusRevoked  = 4 // 已撤销
)

// JobCertification 企业认证表
type JobCertification struct {
	database.BaseModel // id/created_at/updated_at/deleted_at
	CompanyID         uint       `gorm:"not null;index" json:"company_id"`                     // 关联公司 ID
	UserID            uint       `gorm:"not null;index" json:"user_id"`                        // 提交用户 ID
	CertType          string     `gorm:"size:32;default:'business_license';index" json:"cert_type"` // 认证类型
	CertNo            string     `gorm:"size:64;index" json:"cert_no"`                          // 证书编号
	CertName          string     `gorm:"size:128" json:"cert_name"`                             // 证书名称
	CertImage         string     `gorm:"size:255" json:"cert_image"`                            // 证书图片 URL
	LegalPerson       string     `gorm:"size:50" json:"legal_person"`                           // 法人代表
	LegalPersonIDCard string     `gorm:"size:32" json:"legal_person_id_card"`                  // 法人身份证号
	RegisteredCapital float64    `gorm:"type:decimal(14,2);default:0" json:"registered_capital"` // 注册资本
	BusinessScope     string     `gorm:"type:text" json:"business_scope"`                       // 经营范围
	ValidFrom         *time.Time `gorm:"type:date" json:"valid_from"`                            // 有效期起
	ValidTo           *time.Time `gorm:"type:date;index" json:"valid_to"`                       // 有效期止
	Status            int        `gorm:"default:0;index" json:"status"`                         // 0待审 1通过 2拒绝 3过期 4撤销
	VerifiedAt        *time.Time `gorm:"index" json:"verified_at"`                              // 审核通过时间
	VerifierID        uint       `gorm:"index" json:"verifier_id"`                               // 审核员 ID
	VerifierName      string     `gorm:"size:50" json:"verifier_name"`                          // 审核员昵称
	RejectReason      string     `gorm:"size:500" json:"reject_reason"`                         // 拒绝原因
	ExpiredAt         *time.Time `gorm:"index" json:"expired_at"`                               // 失效时间
}

// TableName 表名（job_ 前缀）
func (JobCertification) TableName() string { return "job_certifications" }
