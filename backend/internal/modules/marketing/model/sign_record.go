// Package model 营销活动中台 - 签到记录模型（sign 子域）
// 依据架构设计 4.6：每日签到/连续签到奖励/积分
package model

import (
	"time"

	"wuchang-tongcheng/internal/pkg/database"
)

// SignRecord 签到记录模型（sign_records 表）
type SignRecord struct {
	database.BaseModel // 含 id/created_at/updated_at/deleted_at（用户隔离）

	UserID         uint      `gorm:"not null;index:idx_sign_records_user" json:"user_id"`         // 用户 ID
	SignDate       time.Time `gorm:"type:date;not null;index:idx_sign_records_date" json:"sign_date"` // 签到日期
	ContinuousDays int       `gorm:"not null;default:1" json:"continuous_days"`                    // 连续签到天数
	Points         int       `gorm:"not null;default:0" json:"points"`                            // 本次签到获得积分
}

// TableName 表名
func (SignRecord) TableName() string { return "sign_records" }
