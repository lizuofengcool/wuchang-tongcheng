// Package dto 营销活动中台数据传输对象 - 通用 DTO
package dto

import (
	"wuchang-tongcheng/internal/pkg/utils"
)

// ===== 通用响应 =====

// IDResponse 仅返回 ID 的响应
type IDResponse struct {
	ID uint `json:"id"`
}

// ===== 状态/批量 =====

// AdminUpdateStatusRequest 管理后台更新状态
type AdminUpdateStatusRequest struct {
	Status int `json:"status"`
}

// BatchStatusUpdateRequest 批量状态变更请求
type BatchStatusUpdateRequest struct {
	IDs    []uint `json:"ids" binding:"required,min=1"`
	Status int    `json:"status"`
}

// BatchDeleteRequest 批量删除请求
type BatchDeleteRequest struct {
	IDs []uint `json:"ids" binding:"required,min=1"`
}

// BatchResultResponse 批量操作结果
type BatchResultResponse struct {
	Total     int    `json:"total"`
	Success   int    `json:"success"`
	Failed    int    `json:"failed"`
	FailedIDs []uint `json:"failed_ids,omitempty"`
}

// ===== 通用列表分页基类 =====

// ListRequest 通用列表请求（嵌入分页）
type ListRequest struct {
	utils.Pagination
}
