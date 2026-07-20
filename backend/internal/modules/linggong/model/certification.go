// Package model 资质证书表（对标猪八戒/斗米认证）
// 求职者证书 + 雇主认证 + 平台认证
package model

import (
	"time"

	"wuchang-tongcheng/internal/pkg/database"
)

// === 证书类型常量 ===
const (
	CertTypeIDCard        = "id_card"        // 身份证
	CertTypeHealthCert    = "health_cert"    // 健康证
	CertTypeSkillCert     = "skill_cert"     // 技能证书
	CertTypeEducationCert = "education_cert" // 学历证书
	CertTypeWorkCert      = "work_cert"      // 工作证明
	CertTypeDriverLicense = "driver_license" // 驾驶证
	CertTypeLanguageCert  = "language_cert" // 语言证书
	CertTypeProfessionCert = "profession_cert" // 职业资格证
	CertTypeSafetyCert    = "safety_cert"    // 安全证书
	CertTypeOther         = "other"          // 其他
)

// === 证书状态常量 ===
const (
	CertStatusPending  = 0 // 待审核
	CertStatusApproved = 1 // 已通过
	CertStatusRejected = 2 // 已拒绝
	CertStatusExpired   = 3 // 已过期
	CertStatusRevoked   = 4 // 已撤销
)

// === 颁发机构类型常量 ===
const (
	IssuerTypeGovernment = "government" // 政府机构
	IssuerTypeInstitution = "institution" // 事业单位
	IssuerTypeEnterprise  = "enterprise"  // 企业
	IssuerTypePlatform    = "platform"     // 平台
	IssuerTypeOther       = "other"       // 其他
)

// LinggongCertification 资质证书表
type LinggongCertification struct {
	database.RegionBaseModel // id/region_id/created_at/updated_at/deleted_at
	CertNo         string     `gorm:"size:64;not null;uniqueIndex" json:"cert_no"`               // 证书编号
	UserID         uint       `gorm:"not null;index" json:"user_id"`                              // 持有人 ID
	UserType       string     `gorm:"size:16;not null;default:'worker';index" json:"user_type"`   // worker/employer
	WorkerID       uint       `gorm:"not null;default:0;index" json:"worker_id"`                  // 求职者档案 ID
	EmployerID     uint       `gorm:"not null;default:0;index" json:"employer_id"`                // 雇主认证 ID
	CertType       string     `gorm:"size:32;not null;default:'id_card';index" json:"cert_type"` // 证书类型
	CertName       string     `gorm:"size:128;not null" json:"cert_name"`                          // 证书名称
	CertCode       string     `gorm:"size:128;not null;default:''" json:"cert_code"`              // 证书编码
	IssuerName     string     `gorm:"size:128;not null;default:''" json:"issuer_name"`            // 颁发机构名称
	IssuerType     string     `gorm:"size:32;not null;default:'other'" json:"issuer_type"`         // 颁发机构类型
	IssueDate      *time.Time `gorm:"type:date;index" json:"issue_date"`                            // 颁发日期
	ValidFrom      *time.Time `gorm:"type:date;index" json:"valid_from"`                           // 有效期起
	ValidUntil     *time.Time `gorm:"type:date;index" json:"valid_until"`                          // 有效期止
	ImageURL       string     `gorm:"size:255;not null;default:''" json:"image_url"`              // 证书图片
	ImageBackURL   string     `gorm:"size:255;not null;default:''" json:"image_back_url"`          // 证书反面图片
	SkillID        uint       `gorm:"not null;default:0;index" json:"skill_id"`                  // 关联技能 ID
	SkillName      string     `gorm:"size:64;not null;default:''" json:"skill_name"`              // 关联技能名（冗余）
	Level          string     `gorm:"size:32;not null;default:''" json:"level"`                    // 证书等级
	Score          float64    `gorm:"type:decimal(5,2);default:0" json:"score"`                    // 证书分数
	Verified       bool       `gorm:"not null;default:false;index" json:"verified"`                // 是否已验证
	VerifiedAt     *time.Time `gorm:"index" json:"verified_at"`                                    // 验证时间
	VerifiedBy     uint       `gorm:"not null;default:0" json:"verified_by"`                        // 验证人 ID
	VerifiedByName string     `gorm:"size:50;not null;default:''" json:"verified_by_name"`         // 验证人姓名
	Status         int        `gorm:"default:0;index" json:"status"`                              // 0待审 1通过 2拒绝 3过期 4撤销
	RejectReason   string     `gorm:"size:500;not null;default:''" json:"reject_reason"`          // 拒绝原因
	Description    string     `gorm:"type:text" json:"description"`                                // 描述
}

// TableName 表名（linggong_ 前缀）
func (LinggongCertification) TableName() string { return "linggong_certifications" }
