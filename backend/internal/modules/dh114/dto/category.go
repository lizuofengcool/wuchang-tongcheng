// Package dto 同城114数据传输对象 - 商家分类
package dto

import (
	"wuchang-tongcheng/internal/pkg/utils"
)

// CategoryInfo 分类详情响应
type CategoryInfo struct {
	ID            uint   `json:"id"`
	Name          string `json:"name"`
	Code          string `json:"code"`
	ParentID      uint   `json:"parent_id"`
	Level         int    `json:"level"`
	BusinessType  string `json:"business_type"`
	Icon          string `json:"icon"`
	Color         string `json:"color"`
	Description   string `json:"description"`
	Sort          int    `json:"sort"`
	Status        int    `json:"status"`
	StatusText    string `json:"status_text"`
	BusinessCount int    `json:"business_count"`
	CreatedAt     string `json:"created_at"`
	UpdatedAt     string `json:"updated_at"`
}

// CreateCategoryRequest 创建分类请求
type CreateCategoryRequest struct {
	Name         string `json:"name" binding:"required,max=64"`
	Code         string `json:"code" binding:"required,max=64"`
	ParentID     uint   `json:"parent_id"`
	Level        int    `json:"level" binding:"omitempty,min=1,max=3"`
	BusinessType string `json:"business_type" binding:"omitempty,oneof=restaurant retail service entertain hotel medical education life other"`
	Icon         string `json:"icon" binding:"max=64"`
	Color        string `json:"color" binding:"max=32"`
	Description  string `json:"description" binding:"max=500"`
	Sort         int    `json:"sort"`
	Status       int    `json:"status" binding:"omitempty,oneof=0 1 2"`
}

// UpdateCategoryRequest 更新分类请求
type UpdateCategoryRequest struct {
	Name         *string `json:"name" binding:"omitempty,max=64"`
	Code         *string `json:"code" binding:"omitempty,max=64"`
	ParentID     *uint   `json:"parent_id"`
	Level        *int    `json:"level" binding:"omitempty,min=1,max=3"`
	BusinessType *string `json:"business_type" binding:"omitempty,oneof=restaurant retail service entertain hotel medical education life other"`
	Icon         *string `json:"icon" binding:"max=64"`
	Color        *string `json:"color" binding:"max=32"`
	Description  *string `json:"description" binding:"max=500"`
	Sort         *int    `json:"sort"`
	Status       *int    `json:"status" binding:"omitempty,oneof=0 1 2"`
}

// CategoryListRequest 分类列表请求
type CategoryListRequest struct {
	ParentID     uint   `form:"parent_id" json:"parent_id"`
	Level        int    `form:"level" json:"level"`
	BusinessType string `form:"business_type" json:"business_type"`
	Status       *int   `form:"status" json:"status"`
	Keyword      string `form:"keyword" json:"keyword"`
	utils.Pagination
}
