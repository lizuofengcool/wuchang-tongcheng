// Package module 模块注册表
// 维护所有已注册插件的元信息与开关状态，支撑模块总控面板与运行时启停
package module

import "time"

// Module 模块注册表 GORM 模型
// 每条记录对应一个已注册的插件，由 SyncAllFromManager 在启动时同步
type Module struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	Name         string    `gorm:"uniqueIndex;size:64;not null" json:"name"`        // 模块唯一标识（如 ershou）
	DisplayName  string    `gorm:"size:128" json:"display_name"`                    // 中文展示名
	Category     string    `gorm:"size:32;index" json:"category"`                   // 分类：system/business/marketing/user/community/middleware
	Description  string    `gorm:"size:512" json:"description"`                     // 模块描述
	Version      string    `gorm:"size:32" json:"version"`                          // 版本号
	Dependencies string    `gorm:"size:512" json:"dependencies"`                    // 依赖的其他模块名（JSON 数组字符串）
	Icon         string    `gorm:"size:256" json:"icon"`                            // 图标 URL 或 icon class
	Author       string    `gorm:"size:64" json:"author"`                           // 作者
	Homepage     string    `gorm:"size:256" json:"homepage"`                        // 主页 URL
	Enabled      bool      `gorm:"default:true;index" json:"enabled"`               // 是否启用
	InstalledAt  time.Time `gorm:"autoCreateTime" json:"installed_at"`              // 安装时间
	UpdatedAt    time.Time `gorm:"autoUpdateTime" json:"updated_at"`                // 更新时间
	CreatedAt    time.Time `json:"created_at"`                                       // 创建时间
}

// TableName 表名
func (Module) TableName() string {
	return "modules"
}
