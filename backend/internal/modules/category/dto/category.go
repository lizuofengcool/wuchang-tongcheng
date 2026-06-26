// Package dto 分类信息数据传输对象
package dto

// CategoryInfo 分类信息
type CategoryInfo struct {
	ID       uint   `json:"id"`
	Name     string `json:"name"`
	Icon     string `json:"icon"`
	ParentID uint   `json:"parent_id"`
	Level    int    `json:"level"`
	Sort     int    `json:"sort"`
	Status   int    `json:"status"`
}

// CategoryTree 分类树形结构
type CategoryTree struct {
	CategoryInfo
	Children []CategoryTree `json:"children,omitempty"`
}

// CreateCategoryRequest 创建分类请求
type CreateCategoryRequest struct {
	Name     string `json:"name" binding:"required,max=50"`
	Icon     string `json:"icon" binding:"max=255"`
	ParentID uint   `json:"parent_id"`
	Level    int    `json:"level" binding:"oneof=1 2 3"`
	Sort     int    `json:"sort"`
	Status   int    `json:"status" binding:"omitempty,oneof=0 1"`
}

// UpdateCategoryRequest 更新分类请求
type UpdateCategoryRequest struct {
	Name   string `json:"name" binding:"max=50"`
	Icon   string `json:"icon" binding:"max=255"`
	Sort   int    `json:"sort"`
	Status int    `json:"status" binding:"omitempty,oneof=0 1"`
}
