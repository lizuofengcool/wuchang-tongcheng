// Package dto 同城商城 - 分类 DTO
package dto

import (
	"time"

	"wuchang-tongcheng/internal/pkg/utils"
)

// CategoryInfo 分类详情响应
type CategoryInfo struct {
	ID            uint      `json:"id"`
	ParentID      uint      `json:"parent_id"`
	Name          string    `json:"name"`
	Icon          string    `json:"icon"`
	Cover         string    `json:"cover"`
	Level         int       `json:"level"`
	Path          string    `json:"path"`
	Sort          int       `json:"sort"`
	Status        int       `json:"status"`
	StatusText    string    `json:"status_text"`
	IsShow        bool      `json:"is_show"`
	ProductCount  int       `json:"product_count"`
	Keywords      string    `json:"keywords"`
	Description   string    `json:"description"`
	Children      []CategoryInfo `json:"children,omitempty"` // 子分类（树形结构）
	RegionID      uint      `json:"region_id"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// CreateCategoryRequest 创建分类请求
type CreateCategoryRequest struct {
	ParentID    uint   `json:"parent_id"`
	Name        string `json:"name" binding:"required,max=64"`
	Icon        string `json:"icon" binding:"max=255"`
	Cover       string `json:"cover" binding:"max=255"`
	Sort        int    `json:"sort"`
	Status      int    `json:"status"`
	IsShow      *bool  `json:"is_show"`
	Keywords    string `json:"keywords" binding:"max=255"`
	Description string `json:"description" binding:"max=500"`
}

// UpdateCategoryRequest 更新分类请求
type UpdateCategoryRequest struct {
	Name        *string `json:"name" binding:"max=64"`
	Icon        *string `json:"icon" binding:"max=255"`
	Cover       *string `json:"cover" binding:"max=255"`
	Sort        *int    `json:"sort"`
	Status      *int    `json:"status"`
	IsShow      *bool   `json:"is_show"`
	Keywords    *string `json:"keywords" binding:"max=255"`
	Description *string `json:"description" binding:"max=500"`
}

// CategoryListRequest 分类列表请求
type CategoryListRequest struct {
	utils.Pagination
	ParentID *uint  `form:"parent_id" json:"parent_id"`
	Level    *int   `form:"level" json:"level"`
	Status   *int   `form:"status" json:"status"`
	Keyword  string `form:"keyword" json:"keyword"`
	IsTree   bool   `form:"is_tree" json:"is_tree"` // 是否返回树形结构
}
