// Package dto 公司相关 DTO
// 依据 v3.2.1 架构方案第四章：对标 BOSS直聘公司主页/认证
package dto

import (
	"time"

	"wuchang-tongcheng/internal/pkg/utils"
)

// CompanyResponse 公司详情响应
type CompanyResponse struct {
	ID                uint       `json:"id"`
	UserID            uint       `json:"user_id"`
	Name              string     `json:"name"`
	ShortName         string     `json:"short_name"`
	Logo              string     `json:"logo"`
	Banner            string     `json:"banner"`
	Description       string     `json:"description"`
	Industry          string     `json:"industry"`
	Scale             string     `json:"scale"`
	Level             int        `json:"level"`
	LevelText         string     `json:"level_text"`
	Status            int        `json:"status"`
	StatusText        string     `json:"status_text"`
	ContactName       string     `json:"contact_name"`
	ContactPhone      string     `json:"contact_phone"`
	ContactEmail      string     `json:"contact_email"`
	ContactWechat     string     `json:"contact_wechat"`
	Address           string     `json:"address"`
	Latitude          float64    `json:"latitude"`
	Longitude         float64    `json:"longitude"`
	BusinessLicense   string     `json:"business_license"`
	LicenseNo         string     `json:"license_no"`
	IDCardFront       string     `json:"id_card_front"`
	IDCardBack        string     `json:"id_card_back"`
	LegalPerson       string     `json:"legal_person"`
	LegalPersonIDCard string     `json:"legal_person_id_card"`
	RegisteredCapital float64    `json:"registered_capital"`
	FoundedAt         *time.Time `json:"founded_at"`
	Website           string     `json:"website"`
	VerifiedAt        *time.Time `json:"verified_at"`
	ApprovedAt        *time.Time `json:"approved_at"`
	RejectedReason    string     `json:"rejected_reason"`
	ClosedAt          *time.Time `json:"closed_at"`
	FollowerCount     int        `json:"follower_count"`
	JobCount          int        `json:"job_count"`
	EmployeeCount     int        `json:"employee_count"`
	ActiveJobCount    int        `json:"active_job_count"`
	TotalHiredCount   int        `json:"total_hired_count"`
	GoodRate          float64    `json:"good_rate"`
	Deposit           float64    `json:"deposit"`
	Tags              []string   `json:"tags"`
	IsFollowing       bool       `json:"is_following,omitempty"`
	CreatedAt         time.Time  `json:"created_at"`
}

// CompanyCreateRequest 创建公司请求
type CompanyCreateRequest struct {
	Name              string   `json:"name" binding:"required,max=128"`
	ShortName         string   `json:"short_name" binding:"max=64"`
	Logo              string   `json:"logo" binding:"max=255"`
	Banner            string   `json:"banner" binding:"max=255"`
	Description       string   `json:"description"`
	Industry          string   `json:"industry" binding:"max=64"`
	Scale             string   `json:"scale" binding:"omitempty,oneof=0-20 20-100 100-500 500-1000 1000-10000 10000+"`
	ContactName       string   `json:"contact_name" binding:"max=50"`
	ContactPhone      string   `json:"contact_phone" binding:"max=20"`
	ContactEmail      string   `json:"contact_email" binding:"max=128"`
	ContactWechat     string   `json:"contact_wechat" binding:"max=50"`
	Address           string   `json:"address" binding:"max=500"`
	Latitude          float64  `json:"latitude"`
	Longitude         float64  `json:"longitude"`
	BusinessLicense   string   `json:"business_license" binding:"max=255"`
	LicenseNo         string   `json:"license_no" binding:"max=64"`
	IDCardFront       string   `json:"id_card_front" binding:"max=255"`
	IDCardBack        string   `json:"id_card_back" binding:"max=255"`
	LegalPerson       string   `json:"legal_person" binding:"max=50"`
	LegalPersonIDCard string   `json:"legal_person_id_card" binding:"max=32"`
	RegisteredCapital float64  `json:"registered_capital" binding:"gte=0"`
	FoundedAt         *time.Time `json:"founded_at"`
	Website           string   `json:"website" binding:"max=255"`
	Tags              []string `json:"tags"`
	Deposit           float64  `json:"deposit" binding:"gte=0"`
}

// CompanyUpdateRequest 更新公司请求
type CompanyUpdateRequest struct {
	Name              string   `json:"name" binding:"max=128"`
	ShortName         string   `json:"short_name" binding:"max=64"`
	Logo              string   `json:"logo" binding:"max=255"`
	Banner            string   `json:"banner" binding:"max=255"`
	Description       string   `json:"description"`
	Industry          string   `json:"industry" binding:"max=64"`
	Scale             string   `json:"scale" binding:"omitempty,oneof=0-20 20-100 100-500 500-1000 1000-10000 10000+"`
	ContactName       string   `json:"contact_name" binding:"max=50"`
	ContactPhone      string   `json:"contact_phone" binding:"max=20"`
	ContactEmail      string   `json:"contact_email" binding:"max=128"`
	ContactWechat     string   `json:"contact_wechat" binding:"max=50"`
	Address           string   `json:"address" binding:"max=500"`
	Latitude          float64  `json:"latitude"`
	Longitude         float64  `json:"longitude"`
	Website           string   `json:"website" binding:"max=255"`
	Tags              []string `json:"tags"`
}

// CompanyListQuery 公司列表查询
type CompanyListQuery struct {
	UserID   uint   `form:"user_id" json:"user_id"`
	Status   *int   `form:"status" json:"status"`
	Level    *int   `form:"level" json:"level"`
	Industry string `form:"industry" json:"industry"`
	Keyword  string `form:"keyword" json:"keyword"`
	utils.Pagination
}

// CompanyAuditRequest 公司审核请求
type CompanyAuditRequest struct {
	Status  int    `json:"status" binding:"oneof=1 2 3 4"` // 1通过 2拒绝 3冻结 4关闭
	Level   int    `json:"level" binding:"oneof=0 1 2 3"`
	Reason  string `json:"reason" binding:"max=500"`
}

// CompanyFollowRequest 关注公司请求
type CompanyFollowRequest struct {
	Notify bool `json:"notify"`
}

// ===== 企业认证 =====

// CertificationResponse 认证详情响应
type CertificationResponse struct {
	ID                uint       `json:"id"`
	CompanyID         uint       `json:"company_id"`
	UserID            uint       `json:"user_id"`
	CertType          string     `json:"cert_type"`
	CertNo            string     `json:"cert_no"`
	CertName          string     `json:"cert_name"`
	CertImage         string     `json:"cert_image"`
	LegalPerson       string     `json:"legal_person"`
	LegalPersonIDCard string     `json:"legal_person_id_card"`
	RegisteredCapital float64    `json:"registered_capital"`
	BusinessScope     string     `json:"business_scope"`
	ValidFrom         *time.Time `json:"valid_from"`
	ValidTo           *time.Time `json:"valid_to"`
	Status            int        `json:"status"`
	StatusText        string     `json:"status_text"`
	VerifiedAt        *time.Time `json:"verified_at"`
	VerifierID        uint       `json:"verifier_id"`
	VerifierName      string     `json:"verifier_name"`
	RejectReason      string     `json:"reject_reason"`
	ExpiredAt         *time.Time `json:"expired_at"`
	CreatedAt         time.Time  `json:"created_at"`
}

// CertificationCreateRequest 创建认证请求
type CertificationCreateRequest struct {
	CompanyID         uint       `json:"company_id" binding:"required"`
	CertType          string     `json:"cert_type" binding:"omitempty,oneof=business_license legal_person_id organization_code tax_registration opening_permit social_credit_code industry_license brand_authorization"`
	CertNo            string     `json:"cert_no" binding:"max=64"`
	CertName          string     `json:"cert_name" binding:"max=128"`
	CertImage         string     `json:"cert_image" binding:"max=255"`
	LegalPerson       string     `json:"legal_person" binding:"max=50"`
	LegalPersonIDCard string     `json:"legal_person_id_card" binding:"max=32"`
	RegisteredCapital float64    `json:"registered_capital" binding:"gte=0"`
	BusinessScope     string     `json:"business_scope"`
	ValidFrom         *time.Time `json:"valid_from"`
	ValidTo           *time.Time `json:"valid_to"`
}

// CertificationProcessRequest 认证审核请求
type CertificationProcessRequest struct {
	Status       int    `json:"status" binding:"oneof=1 2 4"` // 1通过 2拒绝 4撤销
	RejectReason string `json:"reject_reason" binding:"max=500"`
}

// CertificationListQuery 认证列表查询
type CertificationListQuery struct {
	CompanyID uint   `form:"company_id" json:"company_id"`
	UserID    uint   `form:"user_id" json:"user_id"`
	Status    *int   `form:"status" json:"status"`
	CertType  string `form:"cert_type" json:"cert_type"`
	utils.Pagination
}
