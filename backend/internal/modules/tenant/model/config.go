// Package model 多租户分站数据模型 - 分站配置
package model

import "time"

// Config 分站配置表（tenant_configs）
// 按 biz_module + config_key 维度存储分站级独立配置，支持配置继承（缺失时回退默认）
type Config struct {
	ID          uint      `gorm:"primarykey" json:"id"`                          // 主键
	StationID   uint      `gorm:"index;not null" json:"station_id"`             // 分站 ID
	BizModule   string    `gorm:"size:50;not null;default:'';index" json:"biz_module"`   // 业务模块（如 dh114/mall/ershou）
	ConfigKey   string    `gorm:"size:100;not null;default:'';index" json:"config_key"` // 配置键
	ConfigValue string    `gorm:"type:text" json:"config_value"`                // 配置值（TEXT）
	UpdatedAt   time.Time `json:"updated_at"`                                    // 更新时间
	CreatedAt   time.Time `json:"created_at"`                                    // 创建时间
}

// TableName 表名
func (Config) TableName() string { return "tenant_configs" }
