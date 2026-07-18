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

// UserOAuth 第三方账号绑定关系
//
// 一个本地用户可绑定多个 provider，同一 (provider, open_id) 全局唯一。
// 自动注册的 OAuth 用户 username 形如 "wechat_<openid>"，密码为随机占位（无法走密码登录）。
type UserOAuth struct {
	database.BaseModel
	UserID   uint   `gorm:"index;not null" json:"user_id"`
	Provider string `gorm:"size:32;not null;uniqueIndex:idx_provider_openid,priority:1" json:"provider"` // wechat ...
	OpenID   string `gorm:"size:128;not null;uniqueIndex:idx_provider_openid,priority:2" json:"open_id"`
	UnionID  string `gorm:"size:128;index" json:"union_id"` // 联合 ID，同主体多应用唯一，可为空
	Nickname string `gorm:"size:50" json:"nickname"`
	Avatar   string `gorm:"size:255" json:"avatar"`
}

// TableName 表名
func (UserOAuth) TableName() string {
	return "user_oauths"
}
