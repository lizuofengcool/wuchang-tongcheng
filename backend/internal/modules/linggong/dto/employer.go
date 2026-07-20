// Package dto 同城零工兼职数据传输对象 - 雇主认证
package dto

import (
	"time"

	"wuchang-tongcheng/internal/pkg/utils"
)

// EmployerInfo 雇主详情响应
type EmployerInfo struct {
	ID                uint       `json:"id"`
	UserID            uint       `json:"user_id"`
	EmployerType      string     `json:"employer_type"`
	EmployerTypeText  string     `json:"employer_type_text"`
	CompanyName       string     `json:"company_name"`
	CompanyShortName  string     `json:"company_short_name"`
	ContactName       string     `json:"contact_name"`
	ContactPhone      string     `json:"contact_phone"`
	ContactEmail      string     `json:"contact_email"`
	ContactWechat     string     `json:"contact_wechat"`
	LicenseNo         string     `json:"license_no"`
	LicenseURL        string     `json:"license_url"`
	LegalPerson       string     `json:"legal_person"`
	LegalPersonIDCard string     `json:"legal_person_id_card"`
	LegalPersonIDCardURL string  `json:"legal_person_id_card_url"`
	BankAccount       string     `json:"bank_account"`
	BankName          string     `json:"bank_name"`
	BrandAuthURL      string     `json:"brand_auth_url"`
	CompanyAddress    string     `json:"company_address"`
	CompanyLatitude   float64    `json:"company_latitude"`
	CompanyLongitude  float64    `json:"company_longitude"`
	CompanyDescription string     `json:"company_description"`
	CompanyLogo       string     `json:"company_logo"`
	CompanyCover      string     `json:"company_cover"`
	Industry          string     `json:"industry"`
	CompanySize       string     `json:"company_size"`
	Level             int        `json:"level"`
	LevelText         string     `json:"level_text"`
	CreditScore       int        `json:"credit_score"`
	Status            int        `json:"status"`
	StatusText        string     `json:"status_text"`
	RejectReason      string     `json:"reject_reason"`
	VerifiedAt        *time.Time `json:"verified_at"`
	VerifiedBy        uint       `json:"verified_by"`
	VerifiedByName    string     `json:"verified_by_name"`
	PublishedCount    int        `json:"published_count"`
	OngoingCount      int        `json:"ongoing_count"`
	CompletedCount    int        `json:"completed_count"`
	TotalWorkers      int        `json:"total_workers"`
	TotalPaid         float64    `json:"total_paid"`
	AvgRating         float64    `json:"avg_rating"`
	RatingCount       int        `json:"rating_count"`
	RegionID          uint       `json:"region_id"`
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`
}

// CreateEmployerRequest 创建雇主认证请求
type CreateEmployerRequest struct {
	EmployerType         string  `json:"employer_type" binding:"omitempty,oneof=personal company agent"`
	CompanyName          string  `json:"company_name" binding:"max=128"`
	CompanyShortName     string  `json:"company_short_name" binding:"max=64"`
	ContactName          string  `json:"contact_name" binding:"max=50"`
	ContactPhone         string  `json:"contact_phone" binding:"max=20"`
	ContactEmail         string  `json:"contact_email" binding:"max=128"`
	ContactWechat        string  `json:"contact_wechat" binding:"max=64"`
	LicenseNo            string  `json:"license_no" binding:"max=64"`
	LicenseURL           string  `json:"license_url" binding:"max=255"`
	LegalPerson          string  `json:"legal_person" binding:"max=50"`
	LegalPersonIDCard    string  `json:"legal_person_id_card" binding:"max=32"`
	LegalPersonIDCardURL string  `json:"legal_person_id_card_url" binding:"max=255"`
	BankAccount          string  `json:"bank_account" binding:"max=64"`
	BankName             string  `json:"bank_name" binding:"max=64"`
	BrandAuthURL         string  `json:"brand_auth_url" binding:"max=255"`
	CompanyAddress       string  `json:"company_address" binding:"max=500"`
	CompanyLatitude      float64 `json:"company_latitude"`
	CompanyLongitude     float64 `json:"company_longitude"`
	CompanyDescription   string  `json:"company_description"`
	CompanyLogo          string  `json:"company_logo" binding:"max=255"`
	CompanyCover         string  `json:"company_cover" binding:"max=255"`
	Industry             string  `json:"industry" binding:"max=64"`
	CompanySize          string  `json:"company_size" binding:"max=32"`
}

// UpdateEmployerRequest 更新雇主认证请求
type UpdateEmployerRequest struct {
	EmployerType         *string  `json:"employer_type" binding:"omitempty,oneof=personal company agent"`
	CompanyName          *string `json:"company_name" binding:"omitempty,max=128"`
	CompanyShortName     *string `json:"company_short_name" binding:"omitempty,max=64"`
	ContactName          *string `json:"contact_name" binding:"omitempty,max=50"`
	ContactPhone         *string `json:"contact_phone" binding:"omitempty,max=20"`
	ContactEmail         *string `json:"contact_email" binding:"omitempty,max=128"`
	ContactWechat        *string `json:"contact_wechat" binding:"omitempty,max=64"`
	LicenseNo            *string `json:"license_no" binding:"omitempty,max=64"`
	LicenseURL           *string `json:"license_url" binding:"omitempty,max=255"`
	LegalPerson          *string `json:"legal_person" binding:"omitempty,max=50"`
	LegalPersonIDCard    *string `json:"legal_person_id_card" binding:"omitempty,max=32"`
	LegalPersonIDCardURL *string `json:"legal_person_id_card_url" binding:"omitempty,max=255"`
	BankAccount          *string `json:"bank_account" binding:"omitempty,max=64"`
	BankName             *string `json:"bank_name" binding:"omitempty,max=64"`
	BrandAuthURL         *string `json:"brand_auth_url" binding:"omitempty,max=255"`
	CompanyAddress       *string `json:"company_address" binding:"omitempty,max=500"`
	CompanyLatitude      *float64 `json:"company_latitude"`
	CompanyLongitude     *float64 `json:"company_longitude"`
	CompanyDescription   *string `json:"company_description"`
	CompanyLogo          *string `json:"company_logo" binding:"omitempty,max=255"`
	CompanyCover         *string `json:"company_cover" binding:"omitempty,max=255"`
	Industry             *string `json:"industry" binding:"omitempty,max=64"`
	CompanySize          *string `json:"company_size" binding:"omitempty,max=32"`
}

// EmployerListRequest 雇主列表请求
type EmployerListRequest struct {
	EmployerType string `form:"employer_type" json:"employer_type"`
	CompanyName  string `form:"company_name" json:"company_name"`
	Status       *int   `form:"status" json:"status"`
	Level        *int   `form:"level" json:"level"`
	Industry     string `form:"industry" json:"industry"`
	Keyword      string `form:"keyword" json:"keyword"`
	utils.Pagination
}

// EmployerAdminListRequest 管理后台雇主列表请求
type EmployerAdminListRequest struct {
	RegionID    uint   `form:"region_id" json:"region_id"`
	UserID      uint   `form:"user_id" json:"user_id"`
	EmployerType string `form:"employer_type" json:"employer_type"`
	Status      *int   `form:"status" json:"status"`
	Level       *int   `form:"level" json:"level"`
	Keyword     string `form:"keyword" json:"keyword"`
	utils.Pagination
}

// EmployerAuditRequest 雇主认证审核请求
type EmployerAuditRequest struct {
	Status       int    `json:"status" binding:"oneof=0 1 2 3 4"`
	RejectReason string `json:"reject_reason" binding:"max=500"`
}
