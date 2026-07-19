// Package dto 同城车辆买卖数据传输对象 - 车险配置
package dto

import (
	"wuchang-tongcheng/internal/pkg/utils"
)

// InsuranceInfo 车险详情响应
type InsuranceInfo struct {
	ID            uint    `json:"id"`
	Name          string  `json:"name"`
	Code          string  `json:"code"`
	InsuranceType string  `json:"insurance_type"`
	Provider      string  `json:"provider"`
	Coverage      string  `json:"coverage"`
	CoverageAmount float64 `json:"coverage_amount"`
	Premium       float64 `json:"premium"`
	Deductible    int     `json:"deductible"`
	Description   string  `json:"description"`
	Sort          int     `json:"sort"`
	Status        int     `json:"status"`
	StatusText    string  `json:"status_text"`
	IsHot         bool    `json:"is_hot"`
	UseCount      int     `json:"use_count"`
	CreatedAt     string  `json:"created_at"`
	UpdatedAt     string  `json:"updated_at"`
}

// CreateInsuranceRequest 创建车险请求
type CreateInsuranceRequest struct {
	Name           string  `json:"name" binding:"required,max=64"`
	Code           string  `json:"code" binding:"required,max=64"`
	InsuranceType  string  `json:"insurance_type" binding:"omitempty,oneof=compulsory commercial third_party theft fire glass engine no_deduct passenger scratch"`
	Provider       string  `json:"provider" binding:"max=128"`
	Coverage       string  `json:"coverage" binding:"max=64"`
	CoverageAmount float64 `json:"coverage_amount"`
	Premium        float64 `json:"premium"`
	Deductible     int     `json:"deductible"`
	Description    string  `json:"description" binding:"max=500"`
	Sort           int     `json:"sort"`
	Status         int     `json:"status" binding:"omitempty,oneof=0 1 2"`
	IsHot          bool    `json:"is_hot"`
}

// UpdateInsuranceRequest 更新车险请求
type UpdateInsuranceRequest struct {
	Name           *string  `json:"name" binding:"omitempty,max=64"`
	Code           *string  `json:"code" binding:"omitempty,max=64"`
	InsuranceType  *string  `json:"insurance_type" binding:"omitempty,oneof=compulsory commercial third_party theft fire glass engine no_deduct passenger scratch"`
	Provider       *string  `json:"provider" binding:"omitempty,max=128"`
	Coverage       *string  `json:"coverage" binding:"omitempty,max=64"`
	CoverageAmount *float64 `json:"coverage_amount"`
	Premium        *float64 `json:"premium"`
	Deductible     *int     `json:"deductible"`
	Description    *string  `json:"description" binding:"omitempty,max=500"`
	Sort           *int     `json:"sort"`
	Status         *int     `json:"status" binding:"omitempty,oneof=0 1 2"`
	IsHot          *bool    `json:"is_hot"`
}

// InsuranceListRequest 车险列表请求
type InsuranceListRequest struct {
	InsuranceType string `form:"insurance_type" json:"insurance_type"`
	Provider      string `form:"provider" json:"provider"`
	Status        *int   `form:"status" json:"status"`
	IsHot         *bool  `form:"is_hot" json:"is_hot"`
	Keyword       string `form:"keyword" json:"keyword"`
	utils.Pagination
}

// InsuranceQuoteRequest 车险报价请求（C 端）
type InsuranceQuoteRequest struct {
	CarID         uint   `json:"car_id" binding:"required"`
	InsuranceIDs  []uint `json:"insurance_ids" binding:"required,min=1"`
	CarPrice      float64 `json:"car_price"`
	RegistrationYear int  `json:"registration_year"`
}

// InsuranceQuoteResponse 车险报价响应
type InsuranceQuoteResponse struct {
	TotalPremium  float64         `json:"total_premium"`
	TotalCoverage float64         `json:"total_coverage"`
	Items         []InsuranceItem `json:"items"`
}

// InsuranceItem 单项车险
type InsuranceItem struct {
	ID            uint    `json:"id"`
	Name          string  `json:"name"`
	InsuranceType string  `json:"insurance_type"`
	Provider      string  `json:"provider"`
	Coverage      string  `json:"coverage"`
	CoverageAmount float64 `json:"coverage_amount"`
	Premium       float64 `json:"premium"`
	Deductible    int     `json:"deductible"`
}
