// Package dto DIY 前端页面中台数据传输对象 - 组件（component 子域）
package dto

import (
	"time"

	"wuchang-tongcheng/internal/pkg/utils"
)

// ComponentInfo 组件详情响应
type ComponentInfo struct {
	ID          uint       `json:"id"`
	Name        string     `json:"name"`
	Code        string     `json:"code"`
	Category    string     `json:"category"`
	CategoryText string    `json:"category_text"`
	Description string     `json:"description"`
	Thumbnail   string     `json:"thumbnail"`
	Status      int        `json:"status"`
	StatusText  string     `json:"status_text"`
	Config      interface{} `json:"config"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// CreateComponentRequest 创建组件请求
type CreateComponentRequest struct {
	Name        string      `json:"name" binding:"required,max=64"`
	Code        string      `json:"code" binding:"required,max=64"`
	Category    string      `json:"category" binding:"omitempty,oneof=basic layout business"`
	Description string      `json:"description"`
	Thumbnail   string      `json:"thumbnail" binding:"max=500"`
	Config      interface{} `json:"config"`
	Status      int         `json:"status" binding:"omitempty,oneof=0 1"`
}

// UpdateComponentRequest 更新组件请求
type UpdateComponentRequest struct {
	Name        *string      `json:"name" binding:"omitempty,max=64"`
	Category    *string      `json:"category" binding:"omitempty,oneof=basic layout business"`
	Description *string      `json:"description"`
	Thumbnail   *string      `json:"thumbnail" binding:"omitempty,max=500"`
	Config      interface{}  `json:"config"`
	Status      *int         `json:"status" binding:"omitempty,oneof=0 1"`
}

// ComponentListRequest 组件列表请求
type ComponentListRequest struct {
	Keyword  string `form:"keyword" json:"keyword"`
	Category string `form:"category" json:"category"`
	Status   *int   `form:"status" json:"status"`
	utils.Pagination
}
