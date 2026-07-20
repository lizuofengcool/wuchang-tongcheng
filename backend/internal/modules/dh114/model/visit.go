// Package model 浏览记录表
// 用户浏览商户/团购/优惠券的记录
package model

import (
	"time"

	"wuchang-tongcheng/internal/pkg/database"
)

// Dh114Visit 浏览记录表
type Dh114Visit struct {
	database.BaseModel // id/created_at/updated_at/deleted_at
	RegionID   uint      `gorm:"not null;default:1;index" json:"region_id"`                       // 地区 ID
	UserID     uint      `gorm:"not null;default:0;index" json:"user_id"`                        // 用户 ID（0 表示未登录）
	Dh114ID    uint      `gorm:"not null;index" json:"dh114_id"`                                  // 商户 ID
	BusinessID uint      `gorm:"not null;default:0;index" json:"business_id"`                     // 商户详情 ID
	VisitType  string    `gorm:"size:32;not null;default:'business';index" json:"visit_type"`     // business/groupbuy/coupon
	IP         string    `gorm:"size:64;not null;default:''" json:"ip"`                          // IP 地址
	UserAgent  string    `gorm:"size:500;not null;default:''" json:"user_agent"`                 // User-Agent
	Referer    string    `gorm:"size:500;not null;default:''" json:"referer"`                    // 来源页
	Device     string    `gorm:"size:32;not null;default:'';index" json:"device"`                // pc/wap/app/miniapp
	Source     string    `gorm:"size:32;not null;default:'';index" json:"source"`                // search/category/recommend/direct/share
	Duration   int       `gorm:"not null;default:0" json:"duration"`                              // 停留时长（秒）
	Longitude  float64   `gorm:"type:decimal(10,7);default:0" json:"longitude"`                  // 经度
	Latitude   float64   `gorm:"type:decimal(10,7);default:0" json:"latitude"`                   // 纬度
	CreatedAt  time.Time `gorm:"index" json:"created_at"`                                       // 兼容字段
}

// TableName 表名
func (Dh114Visit) TableName() string { return "dh114_visits" }
