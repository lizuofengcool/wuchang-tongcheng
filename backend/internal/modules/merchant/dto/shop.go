// Package dto 商户中台数据传输对象 - 店铺
// 依据架构设计 4.4：商家入驻/认领/店铺管理
package dto

import (
	"time"

	"wuchang-tongcheng/internal/pkg/utils"
)

// ShopInfo 店铺详情响应
type ShopInfo struct {
	ID          uint       `json:"id"`
	RegionID    uint       `json:"region_id"`
	OwnerID     uint       `json:"owner_id"`
	Name        string     `json:"name"`
	Logo        string     `json:"logo"`
	Intro       string     `json:"intro"`
	CategoryID  *uint      `json:"category_id"`
	Status      int        `json:"status"`
	StatusText  string     `json:"status_text"`
	CreditScore int        `json:"credit_score"`
	Level       int        `json:"level"`
	SettledAt   *time.Time `json:"settled_at"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// CreateShopRequest 商户入驻请求
type CreateShopRequest struct {
	Name       string `json:"name" binding:"required,max=100"`
	Logo       string `json:"logo" binding:"max=500"`
	Intro      string `json:"intro"`
	CategoryID *uint  `json:"category_id"`
}

// UpdateShopRequest 更新店铺请求
type UpdateShopRequest struct {
	Name       *string `json:"name" binding:"omitempty,max=100"`
	Logo       *string `json:"logo" binding:"max=500"`
	Intro      *string `json:"intro"`
	CategoryID *uint   `json:"category_id"`
}

// ShopListRequest C 端店铺列表请求
type ShopListRequest struct {
	Keyword    string `form:"keyword" json:"keyword"`
	CategoryID uint   `form:"category_id" json:"category_id"`
	OwnerID    uint   `form:"owner_id" json:"owner_id"`
	Status     *int   `form:"status" json:"status"`
	utils.Pagination
}

// ShopAdminListRequest 管理后台店铺列表请求
type ShopAdminListRequest struct {
	RegionID   uint   `form:"region_id" json:"region_id"`
	OwnerID    uint   `form:"owner_id" json:"owner_id"`
	CategoryID uint   `form:"category_id" json:"category_id"`
	Status     *int   `form:"status" json:"status"`
	Keyword    string `form:"keyword" json:"keyword"`
	utils.Pagination
}

// ShopStatusUpdateRequest 状态变更请求
type ShopStatusUpdateRequest struct {
	Status int `json:"status" binding:"oneof=0 1 2"`
}

// ShopCreditAdjustRequest 信用分调整请求
type ShopCreditAdjustRequest struct {
	Delta  int    `json:"delta" binding:"required"`
	Reason string `json:"reason" binding:"max=500"`
}

// ShopLevelUpdateRequest 等级调整请求
type ShopLevelUpdateRequest struct {
	Level int `json:"level" binding:"min=1,max=10"`
}

// ShopClaimRequest 认领店铺请求
type ShopClaimRequest struct {
	ShopID uint `json:"shop_id" binding:"required"`
}
