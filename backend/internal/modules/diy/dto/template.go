// Package dto DIY 前端页面中台数据传输对象 - 模板（template 子域）
package dto

import (
	"time"

	"wuchang-tongcheng/internal/pkg/utils"
)

// TemplateInfo 模板详情响应
type TemplateInfo struct {
	ID          uint       `json:"id"`
	Name        string     `json:"name"`
	Thumbnail   string     `json:"thumbnail"`
	Description string     `json:"description"`
	Category    string     `json:"category"`
	CategoryText string    `json:"category_text"`
	Status      int        `json:"status"`
	StatusText  string     `json:"status_text"`
	Pages       interface{} `json:"pages"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// CreateTemplateRequest 创建模板请求
type CreateTemplateRequest struct {
	Name        string      `json:"name" binding:"required,max=100"`
	Thumbnail   string      `json:"thumbnail" binding:"max=500"`
	Description string      `json:"description"`
	Category    string      `json:"category" binding:"omitempty,oneof=home topic shop activity"`
	Pages       interface{} `json:"pages"`
	Status      int         `json:"status" binding:"omitempty,oneof=0 1"`
}

// UpdateTemplateRequest 更新模板请求
type UpdateTemplateRequest struct {
	Name        *string      `json:"name" binding:"omitempty,max=100"`
	Thumbnail   *string      `json:"thumbnail" binding:"omitempty,max=500"`
	Description *string      `json:"description"`
	Category    *string      `json:"category" binding:"omitempty,oneof=home topic shop activity"`
	Pages       interface{}  `json:"pages"`
	Status      *int         `json:"status" binding:"omitempty,oneof=0 1"`
}

// TemplateListRequest 模板列表请求
type TemplateListRequest struct {
	Keyword  string `form:"keyword" json:"keyword"`
	Category string `form:"category" json:"category"`
	Status   *int   `form:"status" json:"status"`
	utils.Pagination
}

// ApplyTemplateRequest 应用模板请求
type ApplyTemplateRequest struct {
	Title  string `json:"title" binding:"required,max=100"`
	Type   string `json:"type" binding:"omitempty,oneof=home topic shop activity"`
	Slug   string `json:"slug" binding:"max=100"`
	BizID  uint   `json:"biz_id"`
	Status int    `json:"status" binding:"omitempty,oneof=0 1"`
}
