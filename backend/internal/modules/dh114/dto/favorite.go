// Package dto 同城114数据传输对象 - 收藏
package dto

import (
	"time"

	"wuchang-tongcheng/internal/pkg/utils"
)

// FavoriteInfo 收藏信息响应
type FavoriteInfo struct {
	ID           uint       `json:"id"`
	UserID       uint       `json:"user_id"`
	Dh114ID      uint       `json:"dh114_id"`
	BusinessID   uint       `json:"business_id"`
	FavoriteType string     `json:"favorite_type"`
	GroupID      uint       `json:"group_id"`
	Remark       string     `json:"remark"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

// CreateFavoriteRequest 创建收藏请求
type CreateFavoriteRequest struct {
	Dh114ID      uint   `json:"dh114_id" binding:"required"`
	FavoriteType string `json:"favorite_type" binding:"omitempty,oneof=business groupbuy coupon"`
	GroupID      uint   `json:"group_id"`
	Remark       string `json:"remark" binding:"max=200"`
}

// UpdateFavoriteRequest 更新收藏请求
type UpdateFavoriteRequest struct {
	GroupID *uint   `json:"group_id"`
	Remark  *string `json:"remark" binding:"max=200"`
}

// FavoriteListRequest 收藏列表请求
type FavoriteListRequest struct {
	FavoriteType string `form:"favorite_type" json:"favorite_type"`
	GroupID      uint   `form:"group_id" json:"group_id"`
	utils.Pagination
}

// FavoriteGroupInfo 收藏分组信息
type FavoriteGroupInfo struct {
	GroupID   uint   `json:"group_id"`
	GroupName string `json:"group_name"`
	Count     int    `json:"count"`
}

// CreateFavoriteGroupRequest 创建收藏分组请求
type CreateFavoriteGroupRequest struct {
	GroupName string `json:"group_name" binding:"required,max=64"`
}
