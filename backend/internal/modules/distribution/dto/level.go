// Package dto 分销合伙人中台 - 等级请求/响应 DTO
package dto

import (
	"time"

	"wuchang-tongcheng/internal/pkg/utils"
)

// LevelInfo 等级详情响应
type LevelInfo struct {
	ID              uint      `json:"id"`
	Level           int       `json:"level"`
	Name            string    `json:"name"`
	RequiredAmount  float64   `json:"required_amount"`
	CommissionRate  float64   `json:"commission_rate"`
	ExtraBenefits   interface{} `json:"extra_benefits"`
	Status          int       `json:"status"`
	StatusText      string    `json:"status_text"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// LevelCreateRequest 创建等级请求
type LevelCreateRequest struct {
	Level          int         `json:"level" binding:"required,oneof=1 2 3"`
	Name           string      `json:"name" binding:"required,max=64"`
	RequiredAmount float64     `json:"required_amount" binding:"omitempty,min=0"`
	CommissionRate float64     `json:"commission_rate" binding:"omitempty,min=0,max=1"`
	ExtraBenefits  interface{} `json:"extra_benefits"`
	Status         int         `json:"status" binding:"omitempty,oneof=0 1"`
}

// LevelUpdateRequest 更新等级请求
type LevelUpdateRequest struct {
	Name           *string      `json:"name" binding:"omitempty,max=64"`
	RequiredAmount *float64     `json:"required_amount" binding:"omitempty,min=0"`
	CommissionRate *float64     `json:"commission_rate" binding:"omitempty,min=0,max=1"`
	ExtraBenefits  interface{}  `json:"extra_benefits"`
	Status         *int         `json:"status" binding:"omitempty,oneof=0 1"`
}

// LevelListRequest 等级列表请求
type LevelListRequest struct {
	Level  *int  `form:"level" json:"level"`
	Status *int  `form:"status" json:"status"`
	utils.Pagination
}

// LevelCheckUpgradeRequest 检查自动升级请求
type LevelCheckUpgradeRequest struct {
	PartnerID uint `json:"partner_id" binding:"required"`
}

// LevelCheckUpgradeResponse 检查自动升级响应
type LevelCheckUpgradeResponse struct {
	PartnerID       uint    `json:"partner_id"`
	CurrentLevel    int     `json:"current_level"`
	TargetLevel     int     `json:"target_level"`      // 建议升级到的等级（=当前等级表示不升级）
	TotalCommission float64 `json:"total_commission"`  // 累计佣金
	Upgraded        bool    `json:"upgraded"`          // 是否已执行升级
}
