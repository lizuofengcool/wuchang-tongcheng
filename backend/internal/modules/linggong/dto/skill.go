// Package dto 同城零工兼职数据传输对象 - 技能标签
package dto

import (
	"time"

	"wuchang-tongcheng/internal/pkg/utils"
)

// SkillInfo 技能标签详情响应
type SkillInfo struct {
	ID             uint      `json:"id"`
	Name           string    `json:"name"`
	Code           string    `json:"code"`
	Category       string    `json:"category"`
	CategoryText   string    `json:"category_text"`
	ParentID       uint      `json:"parent_id"`
	Level          int       `json:"level"`
	Icon           string    `json:"icon"`
	Color          string    `json:"color"`
	Description    string    `json:"description"`
	WorkerCount    int       `json:"worker_count"`
	LinggongCount  int       `json:"linggong_count"`
	AvgSalary      float64   `json:"avg_salary"`
	HotScore       int       `json:"hot_score"`
	Status         int       `json:"status"`
	StatusText     string    `json:"status_text"`
	Sort           int       `json:"sort"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// CreateSkillRequest 创建技能标签请求
type CreateSkillRequest struct {
	Name        string  `json:"name" binding:"required,max=64"`
	Code        string  `json:"code" binding:"max=64"`
	Category    string  `json:"category" binding:"omitempty,oneof=technical service art language sports driver labor design marketing writing catering other"`
	ParentID    uint    `json:"parent_id"`
	Level       int     `json:"level" binding:"min=1,max=3"`
	Icon        string  `json:"icon" binding:"max=255"`
	Color       string  `json:"color" binding:"max=32"`
	Description string  `json:"description" binding:"max=500"`
	Sort        int     `json:"sort"`
}

// UpdateSkillRequest 更新技能标签请求
type UpdateSkillRequest struct {
	Name        *string `json:"name" binding:"omitempty,max=64"`
	Code        *string `json:"code" binding:"omitempty,max=64"`
	Category    *string `json:"category" binding:"omitempty,oneof=technical service art language sports driver labor design marketing writing catering other"`
	ParentID    *uint   `json:"parent_id"`
	Level       *int    `json:"level" binding:"omitempty,min=1,max=3"`
	Icon        *string `json:"icon" binding:"omitempty,max=255"`
	Color       *string `json:"color" binding:"omitempty,max=32"`
	Description *string `json:"description" binding:"omitempty,max=500"`
	HotScore    *int    `json:"hot_score"`
	Status      *int    `json:"status" binding:"omitempty,oneof=0 1"`
	Sort        *int    `json:"sort"`
}

// SkillListRequest 技能列表请求
type SkillListRequest struct {
	Category string `form:"category" json:"category"`
	ParentID *uint  `form:"parent_id" json:"parent_id"`
	Level    *int   `form:"level" json:"level"`
	Status   *int   `form:"status" json:"status"`
	Keyword  string `form:"keyword" json:"keyword"`
	utils.Pagination
}

// SkillAdminListRequest 管理后台技能列表请求
type SkillAdminListRequest struct {
	Category string `form:"category" json:"category"`
	Status   *int   `form:"status" json:"status"`
	Keyword  string `form:"keyword" json:"keyword"`
	utils.Pagination
}
