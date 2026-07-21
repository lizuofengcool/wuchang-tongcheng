// Package model 多租户分站数据模型 - 分站员工
package model

import "time"

// Staff 分站员工表（tenant_staff）
// 一个分站可拥有多名员工（operator/manager），实现独立运营权限
type Staff struct {
	ID          uint      `gorm:"primarykey" json:"id"`                          // 主键
	StationID   uint      `gorm:"index;not null" json:"station_id"`             // 分站 ID
	UserID      uint      `gorm:"index;not null" json:"user_id"`                // 用户 ID
	Role        string    `gorm:"size:20;not null;default:'operator';index" json:"role"` // operator运营员 / manager管理员
	Permissions JSONB     `gorm:"type:jsonb" json:"permissions"`                // 权限列表（JSONB）
	Status      int       `gorm:"default:1;index" json:"status"`                // 0已停用 1已启用
	CreatedAt   time.Time `json:"created_at"`                                    // 创建时间
	UpdatedAt   time.Time `json:"updated_at"`                                    // 更新时间
}

// TableName 表名
func (Staff) TableName() string { return "tenant_staff" }
