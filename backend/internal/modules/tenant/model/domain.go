// Package model 多租户分站数据模型 - 域名绑定
package model

import "time"

// Domain 域名绑定表（tenant_domains）
// 一个分站可绑定多个域名，其中一个为主域名；SSL 状态独立管理
type Domain struct {
	ID        uint      `gorm:"primarykey" json:"id"`                          // 主键
	StationID uint      `gorm:"index;not null" json:"station_id"`             // 分站 ID
	Domain    string    `gorm:"size:200;not null;uniqueIndex" json:"domain"`  // 域名（唯一）
	IsPrimary bool      `gorm:"default:false;index" json:"is_primary"`        // 是否主域名
	SSLStatus string    `gorm:"size:20;not null;default:'none';index" json:"ssl_status"` // none/pending/active/failed
	CreatedAt time.Time `json:"created_at"`                                    // 创建时间
	UpdatedAt time.Time `json:"updated_at"`                                    // 更新时间
}

// TableName 表名
func (Domain) TableName() string { return "tenant_domains" }
