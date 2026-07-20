// Package dto 同城零工兼职数据传输对象 - 任务包（对标斗米任务制/猪八戒威客）
package dto

import (
	"time"

	"wuchang-tongcheng/internal/pkg/utils"
)

// TaskInfo 任务包详情响应
type TaskInfo struct {
	ID              uint       `json:"id"`
	TaskNo          string     `json:"task_no"`
	LinggongID      uint       `json:"linggong_id"`
	EmployerID      uint       `json:"employer_id"`
	EmployerName    string     `json:"employer_name"`
	Title           string     `json:"title"`
	Description     string     `json:"description"`
	TaskType        string     `json:"task_type"`
	TaskTypeText    string     `json:"task_type_text"`
	Difficulty      string     `json:"difficulty"`
	DifficultyText  string     `json:"difficulty_text"`
	DeliveryMethod  string     `json:"delivery_method"`
	BillingType     string     `json:"billing_type"`
	UnitPrice       float64    `json:"unit_price"`
	TotalCount      int        `json:"total_count"`
	ClaimedCount    int        `json:"claimed_count"`
	CompletedCount  int        `json:"completed_count"`
	VerifiedCount   int        `json:"verified_count"`
	MaxClaimPerUser int        `json:"max_claim_per_user"`
	StartTime       *time.Time `json:"start_time"`
	EndTime         *time.Time `json:"end_time"`
	ClaimDeadline   *time.Time `json:"claim_deadline"`
	SubmitDeadline  *time.Time `json:"submit_deadline"`
	TotalAmount     float64    `json:"total_amount"`
	PaidAmount      float64    `json:"paid_amount"`
	Status          int        `json:"status"`
	StatusText      string     `json:"status_text"`
	AttachmentURL   string     `json:"attachment_url"`
	Tags            interface{} `json:"tags"`
	Requirements    interface{} `json:"requirements"`
	PublishedAt     *time.Time `json:"published_at"`
	CompletedAt     *time.Time `json:"completed_at"`
	RegionID        uint       `json:"region_id"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

// CreateTaskRequest 创建任务包请求
type CreateTaskRequest struct {
	LinggongID      uint       `json:"linggong_id" binding:"required"`
	Title           string     `json:"title" binding:"required,max=200"`
	Description     string     `json:"description"`
	TaskType        string     `json:"task_type" binding:"omitempty,oneof=single batch project contest bounty"`
	Difficulty      string     `json:"difficulty" binding:"omitempty,oneof=easy medium hard expert"`
	DeliveryMethod  string     `json:"delivery_method" binding:"omitempty,oneof=online offline both"`
	BillingType     string     `json:"billing_type" binding:"omitempty,oneof=by_piece by_hour by_day by_week by_month fixed negotiable"`
	UnitPrice       float64    `json:"unit_price" binding:"min=0"`
	TotalCount      int        `json:"total_count" binding:"min=1"`
	MaxClaimPerUser int        `json:"max_claim_per_user" binding:"min=1"`
	StartTime       *time.Time `json:"start_time"`
	EndTime         *time.Time `json:"end_time"`
	ClaimDeadline   *time.Time `json:"claim_deadline"`
	SubmitDeadline  *time.Time `json:"submit_deadline"`
	TotalAmount     float64    `json:"total_amount" binding:"min=0"`
	AttachmentURL   string     `json:"attachment_url" binding:"max=255"`
	Tags            interface{} `json:"tags"`
	Requirements    interface{} `json:"requirements"`
}

// UpdateTaskRequest 更新任务包请求
type UpdateTaskRequest struct {
	Title           *string  `json:"title" binding:"omitempty,max=200"`
	Description     *string `json:"description"`
	TaskType        *string  `json:"task_type" binding:"omitempty,oneof=single batch project contest bounty"`
	Difficulty      *string  `json:"difficulty" binding:"omitempty,oneof=easy medium hard expert"`
	DeliveryMethod  *string  `json:"delivery_method" binding:"omitempty,oneof=online offline both"`
	BillingType     *string  `json:"billing_type" binding:"omitempty,oneof=by_piece by_hour by_day by_week by_month fixed negotiable"`
	UnitPrice       *float64 `json:"unit_price" binding:"omitempty,min=0"`
	TotalCount      *int     `json:"total_count" binding:"omitempty,min=1"`
	MaxClaimPerUser *int     `json:"max_claim_per_user" binding:"omitempty,min=1"`
	StartTime       *time.Time `json:"start_time"`
	EndTime         *time.Time `json:"end_time"`
	ClaimDeadline   *time.Time `json:"claim_deadline"`
	SubmitDeadline  *time.Time `json:"submit_deadline"`
	TotalAmount     *float64 `json:"total_amount" binding:"omitempty,min=0"`
	AttachmentURL   *string  `json:"attachment_url" binding:"omitempty,max=255"`
	Tags            interface{} `json:"tags"`
	Requirements    interface{} `json:"requirements"`
}

// TaskListRequest 任务列表请求
type TaskListRequest struct {
	LinggongID  uint   `form:"linggong_id" json:"linggong_id"`
	EmployerID  uint   `form:"employer_id" json:"employer_id"`
	TaskType    string `form:"task_type" json:"task_type"`
	Difficulty  string `form:"difficulty" json:"difficulty"`
	Status      *int   `form:"status" json:"status"`
	Keyword     string `form:"keyword" json:"keyword"`
	utils.Pagination
}

// TaskAdminListRequest 管理后台任务列表请求
type TaskAdminListRequest struct {
	RegionID    uint   `form:"region_id" json:"region_id"`
	LinggongID  uint   `form:"linggong_id" json:"linggong_id"`
	EmployerID  uint   `form:"employer_id" json:"employer_id"`
	TaskType    string `form:"task_type" json:"task_type"`
	Status      *int   `form:"status" json:"status"`
	Keyword     string `form:"keyword" json:"keyword"`
	utils.Pagination
}

// TaskClaimRequest 任务领取请求
type TaskClaimRequest struct {
	Count int `json:"count" binding:"required,min=1"`
}

// TaskSubmitRequest 任务交付请求
type TaskSubmitRequest struct {
	Count        int    `json:"count" binding:"required,min=1"`
	AttachmentURL string `json:"attachment_url" binding:"max=255"`
	Remark       string `json:"remark"`
}

// TaskVerifyRequest 任务验收请求
type TaskVerifyRequest struct {
	Count       int    `json:"count" binding:"required,min=1"`
	Pass        bool   `json:"pass"`
	RejectReason string `json:"reject_reason" binding:"max=500"`
}
