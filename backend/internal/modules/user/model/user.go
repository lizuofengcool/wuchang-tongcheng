// Package model 用户数据模型
package model

import "wuchang-tongcheng/internal/pkg/database"

// User 用户模型
type User struct {
	database.RegionBaseModel
	Username string `gorm:"size:50;uniqueIndex;not null" json:"username"` // 用户名
	Password string `gorm:"size:100;not null" json:"-"`                   // 密码（bcrypt哈希，不输出）
	Nickname string `gorm:"size:50" json:"nickname"`                      // 昵称
	Avatar   string `gorm:"size:255" json:"avatar"`                        // 头像
	Phone    string `gorm:"size:20;index" json:"phone"`                    // 手机号
	Email    string `gorm:"size:100" json:"email"`                         // 邮箱
	Gender   int    `gorm:"default:0" json:"gender"`                      // 性别 0未知 1男 2女
	Status   int    `gorm:"default:1" json:"status"`                      // 状态 1正常 0禁用
}

// TableName 表名
func (User) TableName() string {
	return "users"
}
