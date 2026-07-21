// Package dto 商户中台数据传输对象 - 类目
// 树形 CRUD
package dto

import (
	"wuchang-tongcheng/internal/pkg/utils"
)

// CategoryInfo 类目详情响应
type CategoryInfo struct {
	ID        uint         `json:"id"`
	ParentID  uint         `json:"parent_id"`
	Name      string       `json:"name"`
	Icon      string       `json:"icon"`
	Sort      int          `json:"sort"`
	Status    int          `json:"status"`
	StatusText string      `json:"status_text"`
	Children  []CategoryInfo `json:"children,omitempty"`
	CreatedAt string       `json:"created_at"`
	UpdatedAt string       `json:"updated_at"`
}

// CreateCategoryRequest 创建类目请求
type CreateCategoryRequest struct {
	ParentID uint   `json:"parent_id"`
	Name     string `json:"name" binding:"required,max=64"`
	Icon     string `json:"icon" binding:"max=255"`
	Sort     int    `json:"sort"`
	Status   int    `json:"status" binding:"omitempty,oneof=0 1"`
}

// UpdateCategoryRequest 更新类目请求
type UpdateCategoryRequest struct {
	ParentID *uint   `json:"parent_id"`
	Name     *string `json:"name" binding:"omitempty,max=64"`
	Icon     *string `json:"icon" binding:"max=255"`
	Sort     *int    `json:"sort"`
	Status   *int    `json:"status" binding:"omitempty,oneof=0 1"`
}

// CategoryListRequest 类目列表请求
type CategoryListRequest struct {
	ParentID *uint  `form:"parent_id" json:"parent_id"`
	Status   *int   `form:"status" json:"status"`
	Keyword  string `form:"keyword" json:"keyword"`
	utils.Pagination
}

// CategoryStatusUpdateRequest 状态变更请求
type CategoryStatusUpdateRequest struct {
	Status int `json:"status" binding:"oneof=0 1"`
}
