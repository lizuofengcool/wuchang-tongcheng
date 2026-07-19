// Package dto 同城车辆买卖数据传输对象 - 车辆评估
package dto

import (
	"time"

	"wuchang-tongcheng/internal/pkg/utils"
)

// EvaluationInfo 评估详情响应
type EvaluationInfo struct {
	ID                 uint       `json:"id"`
	EvaluationNo       string     `json:"evaluation_no"`
	CarID              uint       `json:"car_id"`
	ModelID            uint       `json:"model_id"`
	EvaluatorID        uint       `json:"evaluator_id"`
	EvaluatorName      string     `json:"evaluator_name"`
	EvaluatorLevel     string     `json:"evaluator_level"`
	EvaluationType     string     `json:"evaluation_type"`
	MarketPrice        float64    `json:"market_price"`
	TradeInPrice       float64    `json:"trade_in_price"`
	RetailPrice        float64    `json:"retail_price"`
	WholesalePrice     float64    `json:"wholesale_price"`
	DepreciationAmount float64    `json:"depreciation_amount"`
	DepreciationRate   float64    `json:"depreciation_rate"`
	FinalPrice         float64    `json:"final_price"`
	PriceRangeMin      float64    `json:"price_range_min"`
	PriceRangeMax      float64    `json:"price_range_max"`
	Factors            interface{} `json:"factors"`
	SimilarDeals       interface{} `json:"similar_deals"`
	Description        string     `json:"description"`
	ReportURL          string     `json:"report_url"`
	ValidUntil         *time.Time `json:"valid_until"`
	Status             int        `json:"status"`
	StatusText         string     `json:"status_text"`
	RegionID           uint       `json:"region_id"`
	CreatedAt          time.Time  `json:"created_at"`
	UpdatedAt          time.Time  `json:"updated_at"`
}

// CreateEvaluationRequest 创建评估请求
type CreateEvaluationRequest struct {
	CarID           uint   `json:"car_id" binding:"required"`
	ModelID         uint   `json:"model_id"`
	EvaluatorID     uint   `json:"evaluator_id"`
	EvaluatorName   string `json:"evaluator_name" binding:"max=50"`
	EvaluatorLevel  string `json:"evaluator_level" binding:"omitempty,oneof=junior senior master"`
	EvaluationType  string `json:"evaluation_type" binding:"omitempty,oneof=online offline ai manual"`
	MarketPrice     float64 `json:"market_price"`
	TradeInPrice    float64 `json:"trade_in_price"`
	RetailPrice     float64 `json:"retail_price"`
	WholesalePrice  float64 `json:"wholesale_price"`
	DepreciationAmount float64 `json:"depreciation_amount"`
	DepreciationRate   float64 `json:"depreciation_rate"`
	FinalPrice      float64 `json:"final_price"`
	PriceRangeMin   float64 `json:"price_range_min"`
	PriceRangeMax   float64 `json:"price_range_max"`
	Factors         interface{} `json:"factors"`
	SimilarDeals    interface{} `json:"similar_deals"`
	Description     string `json:"description"`
	ReportURL       string `json:"report_url" binding:"max=255"`
	ValidUntil      *time.Time `json:"valid_until"`
}

// UpdateEvaluationRequest 更新评估请求
type UpdateEvaluationRequest struct {
	EvaluatorID     *uint   `json:"evaluator_id"`
	EvaluatorName   *string `json:"evaluator_name" binding:"omitempty,max=50"`
	EvaluatorLevel  *string `json:"evaluator_level" binding:"omitempty,oneof=junior senior master"`
	EvaluationType  *string `json:"evaluation_type" binding:"omitempty,oneof=online offline ai manual"`
	MarketPrice     *float64 `json:"market_price"`
	TradeInPrice    *float64 `json:"trade_in_price"`
	RetailPrice     *float64 `json:"retail_price"`
	WholesalePrice  *float64 `json:"wholesale_price"`
	DepreciationAmount *float64 `json:"depreciation_amount"`
	DepreciationRate   *float64 `json:"depreciation_rate"`
	FinalPrice      *float64 `json:"final_price"`
	PriceRangeMin   *float64 `json:"price_range_min"`
	PriceRangeMax   *float64 `json:"price_range_max"`
	Factors         interface{} `json:"factors"`
	SimilarDeals    interface{} `json:"similar_deals"`
	Description     *string `json:"description"`
	ReportURL       *string `json:"report_url" binding:"omitempty,max=255"`
	ValidUntil      *time.Time `json:"valid_until"`
	Status          *int    `json:"status" binding:"omitempty,oneof=0 1 2 3"`
}

// EvaluationListRequest 评估列表请求
type EvaluationListRequest struct {
	CarID          uint   `form:"car_id" json:"car_id"`
	ModelID        uint   `form:"model_id" json:"model_id"`
	EvaluatorID    uint   `form:"evaluator_id" json:"evaluator_id"`
	EvaluationType string `form:"evaluation_type" json:"evaluation_type"`
	Status         *int   `form:"status" json:"status"`
	Keyword        string `form:"keyword" json:"keyword"`
	utils.Pagination
}

// OnlineEvaluationRequest 在线评估请求（C 端）
type OnlineEvaluationRequest struct {
	BrandID         *uint   `json:"brand_id" binding:"required"`
	ModelID         *uint   `json:"model_id"`
	BrandName       string  `json:"brand_name"`
	ModelName       string  `json:"model_name"`
	Year            int     `json:"year" binding:"required,min=1900,max=2100"`
	Mileage         float64 `json:"mileage" binding:"min=0"`
	Displacement    float64 `json:"displacement"`
	Transmission    string  `json:"transmission"`
	FuelType        string  `json:"fuel_type"`
	ConditionLevel  string  `json:"condition_level" binding:"omitempty,oneof=A B C D"`
	City            string  `json:"city"`
	TransferCount   int     `json:"transfer_count"`
	AccidentCount   int     `json:"accident_count"`
}

// OnlineEvaluationResponse 在线评估响应
type OnlineEvaluationResponse struct {
	MarketPrice      float64 `json:"market_price"`
	TradeInPrice     float64 `json:"trade_in_price"`
	RetailPrice      float64 `json:"retail_price"`
	PriceRangeMin    float64 `json:"price_range_min"`
	PriceRangeMax    float64 `json:"price_range_max"`
	DepreciationRate float64 `json:"depreciation_rate"`
	FinalPrice       float64 `json:"final_price"`
	Reason           string  `json:"reason"`
}
