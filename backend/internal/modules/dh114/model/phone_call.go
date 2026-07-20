// Package model 电话拨打记录表（一键拨号）
// 记录用户点击拨号/直接拨打的次数与设备信息
package model

import (
	"time"

	"wuchang-tongcheng/internal/pkg/database"
)

// Dh114PhoneCall 电话拨打记录表
type Dh114PhoneCall struct {
	database.RegionBaseModel // 含 id/region_id/created_at/updated_at/deleted_at
	CallNo   string `gorm:"size:64;not null;uniqueIndex:uniq_dh114_phone_calls_no" json:"call_no"` // 拨打单号
	Dh114ID  uint   `gorm:"not null;index" json:"dh114_id"`                                       // 商户 ID
	BusinessID uint  `gorm:"not null;default:0;index" json:"business_id"`                         // 商户详情 ID
	Phone    string `gorm:"size:32;not null;default:'';index" json:"phone"`                       // 被叫号码

	// === 主叫 ===
	CallerID   uint   `gorm:"not null;default:0;index" json:"caller_id"`           // 主叫用户 ID（0 表示游客）
	CallerPhone string `gorm:"size:20;not null;default:''" json:"caller_phone"`     // 主叫号码
	CallerName string `gorm:"size:50;not null;default:''" json:"caller_name"`       // 主叫昵称

	// === 拨打信息 ===
	CallType string `gorm:"size:16;not null;default:'click';index" json:"call_type"` // click/call
	Device   string `gorm:"size:32;not null;default:''" json:"device"`              // pc/wap/app/miniapp
	IP       string `gorm:"size:64;not null;default:''" json:"ip"`                  // IP 地址
	UserAgent string `gorm:"size:500;not null;default:''" json:"user_agent"`        // User-Agent

	// === 结果 ===
	Status   string `gorm:"size:16;not null;default:'success';index" json:"status"` // success/failed
	Duration int    `gorm:"not null;default:0" json:"duration"`                     // 通话时长（秒）
	CalledAt time.Time `gorm:"index" json:"called_at"`                              // 拨打时间
}

// TableName 表名
func (Dh114PhoneCall) TableName() string { return "dh114_phone_calls" }
