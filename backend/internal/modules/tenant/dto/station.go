// Package dto 多租户分站数据传输对象 - 分站
package dto

import "wuchang-tongcheng/internal/pkg/utils"

// StationInfo 分站详情响应
type StationInfo struct {
	ID          uint   `json:"id"`
	RegionID    uint   `json:"region_id"`
	Name        string `json:"name"`
	Domain      string `json:"domain"`
	Logo        string `json:"logo"`
	Description string `json:"description"`
	Status      int    `json:"status"`
	StatusText  string `json:"status_text"`
	Config      any    `json:"config"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

// CreateStationRequest 创建分站请求
type CreateStationRequest struct {
	RegionID    uint   `json:"region_id" binding:"required"`
	Name        string `json:"name" binding:"required,max=100"`
	Domain      string `json:"domain" binding:"max=200"`
	Logo        string `json:"logo" binding:"max=255"`
	Description string `json:"description" binding:"max=1000"`
	Status      int    `json:"status" binding:"omitempty,oneof=0 1"`
	Config      any    `json:"config"`
}

// UpdateStationRequest 更新分站请求
type UpdateStationRequest struct {
	Name        *string `json:"name" binding:"omitempty,max=100"`
	Domain      *string `json:"domain" binding:"omitempty,max=200"`
	Logo        *string `json:"logo" binding:"omitempty,max=255"`
	Description *string `json:"description" binding:"omitempty,max=1000"`
	Status      *int    `json:"status" binding:"omitempty,oneof=0 1"`
	Config      any     `json:"config"`
}

// StationListRequest 分站列表请求
type StationListRequest struct {
	RegionID uint   `form:"region_id" json:"region_id"`
	Name     string `form:"name" json:"name"`
	Domain   string `form:"domain" json:"domain"`
	Status   *int   `form:"status" json:"status"`
	Keyword  string `form:"keyword" json:"keyword"`
	utils.Pagination
}

// UpdateStationStatusRequest 更新分站状态请求（启停）
type UpdateStationStatusRequest struct {
	Status int `json:"status" binding:"oneof=0 1"`
}

// CopyConfigRequest 配置复制请求（从源分站复制配置到目标分站）
type CopyConfigRequest struct {
	SourceStationID uint   `json:"source_station_id" binding:"required"`
	TargetStationID uint   `json:"target_station_id" binding:"required"`
	BizModule       string `json:"biz_module" binding:"max=50"` // 可选，空则复制全部模块
}

// CopyConfigResult 配置复制结果
type CopyConfigResult struct {
	SourceStationID uint   `json:"source_station_id"`
	TargetStationID uint   `json:"target_station_id"`
	BizModule       string `json:"biz_module"`
	CopiedCount     int    `json:"copied_count"`
}
