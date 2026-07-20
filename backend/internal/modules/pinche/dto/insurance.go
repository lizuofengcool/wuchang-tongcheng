// Package dto 同城拼车出行数据传输对象 - 顺风车保险
package dto

import (
	"time"

	"wuchang-tongcheng/internal/pkg/utils"
)

// InsuranceInfo 保险详情响应
type InsuranceInfo struct {
	ID                uint       `json:"id"`
	RegionID          uint       `json:"region_id"`
	PincheID          uint       `json:"pinche_id"`
	BookingID         *uint      `json:"booking_id"`
	PolicyNo          string     `json:"policy_no"`
	InsuranceCompany  string     `json:"insurance_company"`
	InsuranceType     string     `json:"insurance_type"`
	InsuranceTypeText string     `json:"insurance_type_text"`
	CoverageAmount    float64    `json:"coverage_amount"`
	Premium           float64    `json:"premium"`
	InsuredName       string     `json:"insured_name"`
	InsuredIDCard     string     `json:"insured_id_card"`
	StartTime         *time.Time `json:"start_time"`
	EndTime           *time.Time `json:"end_time"`
	Status            int        `json:"status"`
	StatusText        string     `json:"status_text"`
	ClaimAmount       float64    `json:"claim_amount"`
	ClaimReason       string     `json:"claim_reason"`
	ClaimedAt         *time.Time `json:"claimed_at"`
	Beneficiaries     interface{} `json:"beneficiaries"`
	CreatedAt         time.Time  `json:"created_at"`
}

// CreateInsuranceRequest 创建保险请求
type CreateInsuranceRequest struct {
	PincheID         uint   `json:"pinche_id" binding:"required"`
	BookingID        *uint  `json:"booking_id"`
	InsuranceCompany string `json:"insurance_company" binding:"required,max=64"`
	InsuranceType    string `json:"insurance_type" binding:"omitempty,oneof=passenger driver both"`
	CoverageAmount   float64 `json:"coverage_amount" binding:"min=0"`
	Premium          float64 `json:"premium" binding:"min=0"`
	InsuredName      string `json:"insured_name" binding:"required,max=50"`
	InsuredIDCard    string `json:"insured_id_card" binding:"required,len=18"`
	StartTime        *time.Time `json:"start_time"`
	EndTime          *time.Time `json:"end_time"`
	Beneficiaries    interface{} `json:"beneficiaries"`
}

// InsuranceClaimRequest 保险理赔请求
type InsuranceClaimRequest struct {
	ClaimAmount float64 `json:"claim_amount" binding:"min=0"`
	ClaimReason string  `json:"claim_reason" binding:"required,max=500"`
}

// InsuranceListRequest 保险列表查询请求
type InsuranceListRequest struct {
	PincheID   uint   `form:"pinche_id" json:"pinche_id"`
	BookingID  uint   `form:"booking_id" json:"booking_id"`
	Status     *int   `form:"status" json:"status"`
	PolicyNo   string `form:"policy_no" json:"policy_no"`
	utils.Pagination
}

// InsuranceQuoteRequest 保险报价请求
type InsuranceQuoteRequest struct {
	PincheID      uint   `json:"pinche_id" binding:"required"`
	InsuranceType string `json:"insurance_type" binding:"omitempty,oneof=passenger driver both"`
}
