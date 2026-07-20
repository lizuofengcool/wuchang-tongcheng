// Package dto 同城114数据传输对象 - 商户认证
package dto

import (
	"time"

	"wuchang-tongcheng/internal/pkg/utils"
)

// VerificationInfo 认证信息响应
type VerificationInfo struct {
	ID                uint       `json:"id"`
	VerificationNo    string     `json:"verification_no"`
	Dh114ID           uint       `json:"dh114_id"`
	BusinessID        uint       `json:"business_id"`
	UserID            uint       `json:"user_id"`
	VerificationType   string     `json:"verification_type"`
	VerificationTypeText string `json:"verification_type_text"`
	LicenseNo          string     `json:"license_no"`
	LicenseImage      string     `json:"license_image"`
	BusinessName      string     `json:"business_name"`
	LegalPerson       string     `json:"legal_person"`
	BusinessScope     string     `json:"business_scope"`
	RegisteredAddress string     `json:"registered_address"`
	FieldPhotos       interface{} `json:"field_photos"`
	FieldAddress      string     `json:"field_address"`
	FieldDate         *time.Time `json:"field_date"`
	InspectorName     string     `json:"inspector_name"`
	BrandName         string     `json:"brand_name"`
	BrandAuthImage   string     `json:"brand_auth_image"`
	Status            int        `json:"status"`
	StatusText        string     `json:"status_text"`
	AuditRemark       string     `json:"audit_remark"`
	AuditedBy         uint       `json:"audited_by"`
	AuditedAt         *time.Time `json:"audited_at"`
	ValidUntil        *time.Time `json:"valid_until"`
	RegionID          uint       `json:"region_id"`
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`
}

// CreateVerificationRequest 创建认证请求
type CreateVerificationRequest struct {
	Dh114ID          uint   `json:"dh114_id" binding:"required"`
	VerificationType string `json:"verification_type" binding:"required,oneof=business_license field brand license_field"`
	LicenseNo        string `json:"license_no" binding:"max=64"`
	LicenseImage     string `json:"license_image" binding:"max=255"`
	BusinessName     string `json:"business_name" binding:"max=200"`
	LegalPerson      string `json:"legal_person" binding:"max=64"`
	LegalPersonIDCard string `json:"legal_person_id_card" binding:"max=32"`
	BusinessScope    string `json:"business_scope"`
	RegisteredAddress string `json:"registered_address" binding:"max=500"`
	FieldPhotos      interface{} `json:"field_photos"`
	FieldAddress    string `json:"field_address" binding:"max=500"`
	FieldLongitude  float64 `json:"field_longitude"`
	FieldLatitude   float64 `json:"field_latitude"`
	FieldDate       *time.Time `json:"field_date"`
	BrandName       string `json:"brand_name" binding:"max=128"`
	BrandAuthImage  string `json:"brand_auth_image" binding:"max=255"`
	ValidUntil      *time.Time `json:"valid_until"`
}

// UpdateVerificationRequest 更新认证请求
type UpdateVerificationRequest struct {
	LicenseNo         *string `json:"license_no" binding:"max=64"`
	LicenseImage     *string `json:"license_image" binding:"max=255"`
	BusinessName     *string `json:"business_name" binding:"max=200"`
	LegalPerson      *string `json:"legal_person" binding:"max=64"`
	LegalPersonIDCard *string `json:"legal_person_id_card" binding:"max=32"`
	BusinessScope    *string `json:"business_scope"`
	RegisteredAddress *string `json:"registered_address" binding:"max=500"`
	FieldPhotos      interface{} `json:"field_photos"`
	FieldAddress     *string `json:"field_address" binding:"max=500"`
	FieldLongitude   *float64 `json:"field_longitude"`
	FieldLatitude    *float64 `json:"field_latitude"`
	FieldDate        *time.Time `json:"field_date"`
	BrandName        *string `json:"brand_name" binding:"max=128"`
	BrandAuthImage   *string `json:"brand_auth_image" binding:"max=255"`
	ValidUntil       *time.Time `json:"valid_until"`
}

// VerificationListRequest 认证列表请求
type VerificationListRequest struct {
	Dh114ID          uint   `form:"dh114_id" json:"dh114_id"`
	UserID           uint   `form:"user_id" json:"user_id"`
	VerificationType string `form:"verification_type" json:"verification_type"`
	Status           *int   `form:"status" json:"status"`
	utils.Pagination
}

// VerificationAuditRequest 审核认证请求
type VerificationAuditRequest struct {
	Status      int    `json:"status" binding:"oneof=1 2 3"`
	AuditRemark string `json:"audit_remark" binding:"max=500"`
	ValidUntil  *time.Time `json:"valid_until"`
}
