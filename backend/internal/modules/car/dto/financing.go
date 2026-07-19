// Package dto 同城车辆买卖数据传输对象 - 分期付款方案
package dto

import (
	"wuchang-tongcheng/internal/pkg/utils"
)

// FinancingInfo 分期方案详情响应
type FinancingInfo struct {
	ID             uint    `json:"id"`
	Name           string  `json:"name"`
	Code           string  `json:"code"`
	FinancingType  string  `json:"financing_type"`
	MinDownPayment float64 `json:"min_down_payment"`
	MaxDownPayment float64 `json:"max_down_payment"`
	InterestRate   float64 `json:"interest_rate"`
	AnnualRate     float64 `json:"annual_rate"`
	MinPeriods     int     `json:"min_periods"`
	MaxPeriods     int     `json:"max_periods"`
	MaxAmount      float64 `json:"max_amount"`
	Provider       string  `json:"provider"`
	Description    string  `json:"description"`
	Sort           int     `json:"sort"`
	Status         int     `json:"status"`
	StatusText     string  `json:"status_text"`
	IsHot          bool    `json:"is_hot"`
	UseCount       int     `json:"use_count"`
	CreatedAt      string  `json:"created_at"`
	UpdatedAt      string  `json:"updated_at"`
}

// CreateFinancingRequest 创建分期方案请求
type CreateFinancingRequest struct {
	Name           string  `json:"name" binding:"required,max=64"`
	Code           string  `json:"code" binding:"required,max=64"`
	FinancingType  string  `json:"financing_type" binding:"omitempty,oneof=loan lease balloon zero_down replace_loan"`
	MinDownPayment float64 `json:"min_down_payment"`
	MaxDownPayment float64 `json:"max_down_payment"`
	InterestRate   float64 `json:"interest_rate"`
	AnnualRate     float64 `json:"annual_rate"`
	MinPeriods     int     `json:"min_periods"`
	MaxPeriods     int     `json:"max_periods"`
	MaxAmount      float64 `json:"max_amount"`
	Provider       string  `json:"provider" binding:"max=128"`
	Description    string  `json:"description" binding:"max=500"`
	Sort           int     `json:"sort"`
	Status         int     `json:"status" binding:"omitempty,oneof=0 1 2"`
	IsHot          bool    `json:"is_hot"`
}

// UpdateFinancingRequest 更新分期方案请求
type UpdateFinancingRequest struct {
	Name           *string  `json:"name" binding:"omitempty,max=64"`
	Code           *string  `json:"code" binding:"omitempty,max=64"`
	FinancingType  *string  `json:"financing_type" binding:"omitempty,oneof=loan lease balloon zero_down replace_loan"`
	MinDownPayment *float64 `json:"min_down_payment"`
	MaxDownPayment *float64 `json:"max_down_payment"`
	InterestRate   *float64 `json:"interest_rate"`
	AnnualRate     *float64 `json:"annual_rate"`
	MinPeriods     *int     `json:"min_periods"`
	MaxPeriods     *int     `json:"max_periods"`
	MaxAmount      *float64 `json:"max_amount"`
	Provider       *string  `json:"provider" binding:"omitempty,max=128"`
	Description    *string  `json:"description" binding:"omitempty,max=500"`
	Sort           *int     `json:"sort"`
	Status         *int     `json:"status" binding:"omitempty,oneof=0 1 2"`
	IsHot          *bool    `json:"is_hot"`
}

// FinancingListRequest 分期方案列表请求
type FinancingListRequest struct {
	FinancingType string `form:"financing_type" json:"financing_type"`
	Provider      string `form:"provider" json:"provider"`
	Status        *int   `form:"status" json:"status"`
	IsHot         *bool  `form:"is_hot" json:"is_hot"`
	Keyword       string `form:"keyword" json:"keyword"`
	utils.Pagination
}

// FinancingCalculateRequest 分期计算请求（C 端）
type FinancingCalculateRequest struct {
	FinancingID uint    `json:"financing_id" binding:"required"`
	CarPrice    float64 `json:"car_price" binding:"required,min=0"`
	DownPayment float64 `json:"down_payment" binding:"min=0"`
	Periods     int     `json:"periods" binding:"required,min=1"`
}

// FinancingCalculateResponse 分期计算响应
type FinancingCalculateResponse struct {
	CarPrice      float64 `json:"car_price"`
	DownPayment   float64 `json:"down_payment"`
	LoanAmount    float64 `json:"loan_amount"`
	Periods       int     `json:"periods"`
	InterestRate  float64 `json:"interest_rate"`
	AnnualRate    float64 `json:"annual_rate"`
	MonthlyPayment float64 `json:"monthly_payment"`
	TotalPayment  float64 `json:"total_payment"`
	TotalInterest float64 `json:"total_interest"`
}
