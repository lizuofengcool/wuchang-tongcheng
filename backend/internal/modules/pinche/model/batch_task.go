// Package model 拼车批量任务数据模型
package model

import (
	"time"

	"wuchang-tongcheng/internal/pkg/database"
)

// 批量任务状态常量
const (
	BatchTaskStatusPending  = 0 // 待执行
	BatchTaskStatusRunning  = 1 // 执行中
	BatchTaskStatusDone     = 2 // 已完成
	BatchTaskStatusFailed   = 3 // 失败
	BatchTaskStatusCanceled = 4 // 已取消
)

// PincheBatchTask 批量任务
type PincheBatchTask struct {
	database.RegionBaseModel

	TaskNo       string `gorm:"size:32;index" json:"task_no"`
	TaskName     string `gorm:"size:100" json:"task_name"`
	TaskType     string `gorm:"size:32;index" json:"task_type"`
	TargetIDs    JSONB  `gorm:"type:jsonb" json:"target_ids"`
	TargetCount  int    `gorm:"not null;default:0" json:"target_count"`
	Filters      JSONB  `gorm:"type:jsonb" json:"filters"`
	Action       string `gorm:"size:32" json:"action"`
	ActionParams JSONB  `gorm:"type:jsonb" json:"action_params"`

	Status       int        `gorm:"default:0;index" json:"status"`
	Progress     int        `gorm:"not null;default:0" json:"progress"`
	SuccessCount int        `gorm:"not null;default:0" json:"success_count"`
	FailCount    int        `gorm:"not null;default:0" json:"fail_count"`
	FailReason   string     `gorm:"type:text" json:"fail_reason"`

	OperatorID   *uint      `gorm:"index" json:"operator_id"`
	OperatorName string     `gorm:"size:50" json:"operator_name"`
	StartedAt    *time.Time `gorm:"index" json:"started_at"`
	FinishedAt   *time.Time `gorm:"index" json:"finished_at"`
}

// TableName 表名
func (PincheBatchTask) TableName() string { return "pinche_batch_tasks" }
