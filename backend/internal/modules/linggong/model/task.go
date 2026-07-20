// Package model 任务包表（对标斗米任务制 + 猪八戒威客）
// 长短期任务分类 + 按件/按时/按日计费 + 任务领取/交付
package model

import (
	"time"

	"wuchang-tongcheng/internal/pkg/database"
)

// === 任务包状态常量 ===
const (
	TaskStatusDraft      = 0 // 草稿
	TaskStatusPublished  = 1 // 已发布
	TaskStatusInProgress = 2 // 进行中
	TaskStatusCompleted  = 3 // 已完成
	TaskStatusCanceled   = 4 // 已取消
	TaskStatusExpired     = 5 // 已过期
)

// === 任务类型常量 ===
const (
	TaskTypeSingle     = "single"     // 单一任务
	TaskTypeBatch      = "batch"      // 批量任务
	TaskTypeProject    = "project"    // 项目制
	TaskTypeContest    = "contest"    // 比赛制
	TaskTypeBounty     = "bounty"     // 悬赏制
)

// === 任务难度常量 ===
const (
	TaskDifficultyEasy   = "easy"   // 简单
	TaskDifficultyMedium = "medium" // 中等
	TaskDifficultyHard   = "hard"   // 困难
	TaskDifficultyExpert = "expert" // 专家级
)

// === 任务交付方式常量 ===
const (
	TaskDeliveryOnline  = "online"  // 在线交付
	TaskDeliveryOffline = "offline" // 线下交付
	TaskDeliveryBoth    = "both"    // 均可
)

// LinggongTask 任务包表
type LinggongTask struct {
	database.RegionBaseModel // id/region_id/created_at/updated_at/deleted_at
	TaskNo          string     `gorm:"size:64;not null;uniqueIndex" json:"task_no"`               // 任务单号
	LinggongID     uint       `gorm:"not null;index" json:"linggong_id"`                          // 关联岗位 ID
	EmployerID     uint       `gorm:"not null;index" json:"employer_id"`                          // 雇主 ID
	EmployerName   string     `gorm:"size:128;not null;default:''" json:"employer_name"`          // 雇主名称
	Title          string     `gorm:"size:200;not null;default:''" json:"title"`                  // 任务标题
	Description   string     `gorm:"type:text" json:"description"`                               // 任务描述
	TaskType       string     `gorm:"size:32;not null;default:'single';index" json:"task_type"`    // single/batch/project/contest/bounty
	Difficulty     string     `gorm:"size:16;not null;default:'easy';index" json:"difficulty"`      // easy/medium/hard/expert
	DeliveryMethod string     `gorm:"size:16;not null;default:'online'" json:"delivery_method"`    // online/offline/both
	BillingType   string     `gorm:"size:32;not null;default:'by_piece';index" json:"billing_type"` // by_piece/by_hour/by_day/fixed
	UnitPrice     float64    `gorm:"type:decimal(12,2);default:0;index" json:"unit_price"`        // 单价
	TotalCount    int        `gorm:"not null;default:1" json:"total_count"`                       // 任务总数
	ClaimedCount  int        `gorm:"not null;default:0" json:"claimed_count"`                     // 已领取数
	CompletedCount int       `gorm:"not null;default:0" json:"completed_count"`                  // 已完成数
	VerifiedCount int        `gorm:"not null;default:0" json:"verified_count"`                    // 已验收数
	MaxClaimPerUser int      `gorm:"not null;default:1" json:"max_claim_per_user"`                // 单人最大领取数
	StartTime     *time.Time `gorm:"index" json:"start_time"`                                    // 任务开始时间
	EndTime       *time.Time `gorm:"index" json:"end_time"`                                       // 任务结束时间
	ClaimDeadline *time.Time `gorm:"index" json:"claim_deadline"`                                // 领取截止时间
	SubmitDeadline *time.Time `gorm:"index" json:"submit_deadline"`                               // 交付截止时间
	TotalAmount   float64    `gorm:"type:decimal(12,2);default:0" json:"total_amount"`            // 任务总额
	PaidAmount    float64    `gorm:"type:decimal(12,2);default:0" json:"paid_amount"`              // 已支付金额
	Status        int        `gorm:"default:0;index" json:"status"`                                // 0草稿 1已发布 2进行中 3完成 4取消 5过期
	AttachmentURL string     `gorm:"size:255;not null;default:''" json:"attachment_url"`           // 附件 URL
	Tags          JSONB      `gorm:"type:jsonb" json:"tags"`                                       // 标签 JSON
	Requirements  JSONB      `gorm:"type:jsonb" json:"requirements"`                               // 详细要求 JSON
	PublishedAt   *time.Time `gorm:"index" json:"published_at"`                                    // 发布时间
	CompletedAt   *time.Time `gorm:"index" json:"completed_at"`                                    // 完成时间
}

// TableName 表名（linggong_ 前缀）
func (LinggongTask) TableName() string { return "linggong_tasks" }
