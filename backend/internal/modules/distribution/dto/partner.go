// Package dto 分销合伙人中台 - 合伙人请求/响应 DTO
package dto

import (
	"time"

	"wuchang-tongcheng/internal/pkg/utils"
)

// PartnerInfo 合伙人详情响应
type PartnerInfo struct {
	ID                uint       `json:"id"`
	UserID            uint       `json:"user_id"`
	ParentID          uint       `json:"parent_id"`
	Level             int        `json:"level"`
	LevelText         string     `json:"level_text"`
	CommissionRate    float64    `json:"commission_rate"`
	TotalCommission   float64    `json:"total_commission"`
	SettledCommission float64    `json:"settled_commission"`
	PendingCommission float64    `json:"pending_commission"`
	Status            int        `json:"status"`
	StatusText        string     `json:"status_text"`
	JoinedAt          *time.Time `json:"joined_at"`
	RegionID          uint       `json:"region_id"`
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`

	// 扩展（列表/树形展示用）
	ParentName  string        `json:"parent_name,omitempty"`
	ChildCount  int           `json:"child_count,omitempty"`
	Children    []PartnerInfo `json:"children,omitempty"`
}

// PartnerApplyRequest 申请加入合伙人
type PartnerApplyRequest struct {
	ParentID       uint    `json:"parent_id"`                                     // 邀请人（上级）合伙人 ID，0=无
	Level          int     `json:"level" binding:"omitempty,oneof=1 2 3"`          // 申请等级
	CommissionRate float64 `json:"commission_rate" binding:"omitempty,min=0,max=1"` // 佣金比例（0-1）
}

// PartnerUpdateRequest 管理后台更新合伙人
type PartnerUpdateRequest struct {
	Level           *int     `json:"level" binding:"omitempty,oneof=1 2 3"`
	CommissionRate  *float64 `json:"commission_rate" binding:"omitempty,min=0,max=1"`
	ParentID        *uint    `json:"parent_id"`
	Status          *int     `json:"status" binding:"omitempty,oneof=0 1 2 3 4"`
}

// PartnerListRequest 合伙人列表请求
type PartnerListRequest struct {
	UserID   uint   `form:"user_id" json:"user_id"`
	ParentID uint   `form:"parent_id" json:"parent_id"`
	Level    *int   `form:"level" json:"level"`
	Status   *int   `form:"status" json:"status"`
	RegionID uint   `form:"region_id" json:"region_id"`
	Keyword  string `form:"keyword" json:"keyword"`
	utils.Pagination
}

// PartnerTreeRequest 上下级树请求
type PartnerTreeRequest struct {
	ParentID uint `form:"parent_id" json:"parent_id"` // 根节点（默认 0=全部顶级）
	Depth    int  `form:"depth" json:"depth"`         // 深度，默认 2
}
