// Package dto 商户中台数据传输对象 - 认证
// 支持企业认证（business）与个人认证（personal）
package dto

import (
	"time"

	"wuchang-tongcheng/internal/pkg/utils"
)

// VerificationInfo 认证详情响应
type VerificationInfo struct {
	ID            uint       `json:"id"`
	ShopID        uint       `json:"shop_id"`
	RegionID      uint       `json:"region_id"`
	Type          string     `json:"type"`
	TypeText      string     `json:"type_text"`
	LicenseNo     string     `json:"license_no"`
	LicenseImage  string     `json:"license_image"`
	LegalPerson   string     `json:"legal_person"`
	LegalPersonID string     `json:"legal_person_id"`
	Status        int        `json:"status"`
	StatusText    string     `json:"status_text"`
	AuditReason   string     `json:"audit_reason"`
	AuditedAt     *time.Time `json:"audited_at"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

// CreateVerificationRequest 提交认证请求
type CreateVerificationRequest struct {
	ShopID        uint   `json:"shop_id" binding:"required"`
	Type          string `json:"type" binding:"required,oneof=business personal"`
	LicenseNo     string `json:"license_no" binding:"max=64"`
	LicenseImage  string `json:"license_image" binding:"max=255"`
	LegalPerson   string `json:"legal_person" binding:"max=64"`
	LegalPersonID string `json:"legal_person_id" binding:"max=32"`
}

// UpdateVerificationRequest 更新认证请求
type UpdateVerificationRequest struct {
	Type          *string `json:"type" binding:"omitempty,oneof=business personal"`
	LicenseNo     *string `json:"license_no" binding:"max=64"`
	LicenseImage  *string `json:"license_image" binding:"max=255"`
	LegalPerson   *string `json:"legal_person" binding:"max=64"`
	LegalPersonID *string `json:"legal_person_id" binding:"max=32"`
}

// VerificationListRequest 认证列表请求
type VerificationListRequest struct {
	ShopID  uint   `form:"shop_id" json:"shop_id"`
	Type    string `form:"type" json:"type"`
	Status  *int   `form:"status" json:"status"`
	utils.Pagination
}

// VerificationAuditRequest 审核请求
type VerificationAuditRequest struct {
	Status      int    `json:"status" binding:"oneof=1 2"` // 1通过 2拒绝
	AuditReason string `json:"audit_reason" binding:"max=500"`
}
