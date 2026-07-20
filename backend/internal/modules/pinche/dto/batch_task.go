// Package dto 同城拼车出行数据传输对象 - 批量任务
package dto

import (
	"time"
)

// BatchTaskInfo 批量任务详情
type BatchTaskInfo struct {
	ID            uint       `json:"id"`
	RegionID      uint       `json:"region_id"`
	TaskNo        string     `json:"task_no"`
	Name          string     `json:"name"`
	TaskType      string     `json:"task_type"`
	TargetType    string     `json:"target_type"`
	TargetIDs     interface{} `json:"target_ids,omitempty"`
	TargetCount   int        `json:"total_count"`
	Filters       interface{} `json:"filters,omitempty"`
	Action        string     `json:"action"`
	ActionParams  interface{} `json:"action_params,omitempty"`
	Status        int        `json:"status"`
	StatusText    string     `json:"status_text"`
	Progress      int        `json:"progress"`
	SuccessCount  int        `json:"success_count"`
	FailCount     int        `json:"fail_count"`
	SkipCount     int        `json:"skip_count"`
	FailReason    string     `json:"error_message,omitempty"`
	OperatorID    *uint      `json:"operator_id"`
	OperatorName  string     `json:"operator_name"`
	RunMode       string     `json:"run_mode"`
	OnFailure    string     `json:"on_failure"`
	Description   string     `json:"description,omitempty"`
	StartedAt     *time.Time `json:"started_at"`
	CompletedAt   *time.Time `json:"completed_at"`
	FinishedAt    *time.Time `json:"finished_at,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
}

// CreateBatchTaskRequest 创建批量任务请求
type CreateBatchTaskRequest struct {
	Name         string      `json:"name" binding:"required,max=100"`
	TargetType   string      `json:"target_type" binding:"required"`
	Action       string      `json:"action" binding:"required"`
	TargetIDs    []uint      `json:"target_ids"`
	TaskType     string      `json:"task_type"`
	Filters      interface{} `json:"filters"`
	ActionParams interface{} `json:"action_params"`
	RunMode      string      `json:"run_mode"`
	OnFailure    string      `json:"on_failure"`
	Description  string      `json:"description" binding:"max=500"`
}

// BatchTaskListRequest 批量任务列表查询请求
type BatchTaskListRequest struct {
	Page       int    `form:"page" json:"page"`
	PageSize   int    `form:"page_size" json:"page_size"`
	Keyword    string `form:"keyword" json:"keyword"`
	Action     string `form:"action" json:"action"`
	TargetType string `form:"target_type" json:"target_type"`
	Status     *int   `form:"status" json:"status"`
}

// BatchTaskStatsResponse 批量任务统计响应
type BatchTaskStatsResponse struct {
	Total     int64 `json:"total"`
	Pending   int64 `json:"pending"`
	Running   int64 `json:"running"`
	Completed int64 `json:"completed"`
}

// BatchTaskListResult 批量任务列表响应（带统计）
type BatchTaskListResult struct {
	List     []BatchTaskInfo        `json:"list"`
	Total    int64                  `json:"total"`
	Page     int                    `json:"page"`
	PageSize int                    `json:"page_size"`
	Stats    BatchTaskStatsResponse `json:"stats"`
}

// PreviewIDsRequest 预览 ID 请求
type PreviewIDsRequest struct {
	TargetType string `form:"target_type" json:"target_type"`
	TaskType   string `form:"task_type" json:"task_type"`
	Filters    string `form:"filters" json:"filters"`
	Limit      int    `form:"limit" json:"limit"`
}

// PreviewIDsResponse 预览 ID 响应
type PreviewIDsResponse struct {
	IDS   []uint `json:"ids"`
	Total int64  `json:"total"`
}
