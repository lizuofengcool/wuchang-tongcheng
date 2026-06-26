// Package dto 地区模块数据传输对象
package dto

// RegionInfo 地区信息
type RegionInfo struct {
	ID       uint   `json:"id"`
	Name     string `json:"name"`
	Code     string `json:"code"`
	ParentID uint   `json:"parent_id"`
	Level    int    `json:"level"`
	Sort     int    `json:"sort"`
	Status   int    `json:"status"`
}

// RegionTree 地区树形结构
type RegionTree struct {
	RegionInfo
	Children []RegionTree `json:"children,omitempty"`
}

// CreateRegionRequest 创建地区请求
type CreateRegionRequest struct {
	Name     string `json:"name" binding:"required,max=50"`
	Code     string `json:"code" binding:"required,max=20"`
	ParentID uint   `json:"parent_id"`
	Level    int    `json:"level" binding:"oneof=1 2 3"`
	Sort     int    `json:"sort"`
	Status   int    `json:"status" binding:"oneof=0 1"`
}

// UpdateRegionRequest 更新地区请求
type UpdateRegionRequest struct {
	Name   string `json:"name" binding:"max=50"`
	Sort   int    `json:"sort"`
	Status int    `json:"status" binding:"omitempty,oneof=0 1"`
}
