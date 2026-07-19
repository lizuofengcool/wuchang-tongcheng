// Package model 职位/公司收藏 + 浏览记录（对标 BOSS直聘）
package model

import (
	"time"

	"wuchang-tongcheng/internal/pkg/database"
)

// === 收藏类型常量 ===
const (
	FavoriteTypeJob     = "job"     // 职位收藏
	FavoriteTypeCompany = "company" // 公司收藏
	FavoriteTypeResume  = "resume"  // 简历收藏（招聘者收藏求职者简历）
	FavoriteTypeSearch  = "search"  // 搜索条件收藏
)

// JobFavorite 收藏表
type JobFavorite struct {
	database.BaseModel // id/created_at/updated_at/deleted_at
	UserID       uint   `gorm:"not null;index;uniqueIndex:uniq_job_fav_user_type_target" json:"user_id"` // 收藏人 ID
	JobID        uint   `gorm:"index;uniqueIndex:uniq_job_fav_user_type_target" json:"job_id"`            // 关联职位 ID
	CompanyID    uint   `gorm:"index;uniqueIndex:uniq_job_fav_user_type_target" json:"company_id"`         // 关联公司 ID
	FavoriteType string `gorm:"size:16;default:'job';index;uniqueIndex:uniq_job_fav_user_type_target" json:"favorite_type"` // 收藏类型
	Notify       bool   `gorm:"default:true" json:"notify"`                                                // 是否通知更新
}

// TableName 表名（job_ 前缀）
func (JobFavorite) TableName() string { return "job_favorites" }

// === 浏览类型常量 ===
const (
	ViewTypeJob     = "job"     // 浏览职位
	ViewTypeCompany = "company" // 浏览公司
	ViewTypeResume  = "resume"  // 浏览简历
	ViewTypeRecruiter = "recruiter" // 浏览招聘者主页
)

// === 浏览来源常量 ===
const (
	ViewSourceList     = "list"     // 列表
	ViewSourceSearch   = "search"   // 搜索
	ViewSourceRecommend = "recommend" // 推荐
	ViewSourcePush     = "push"     // 推送
	ViewSourceShare    = "share"    // 分享
	ViewSourceQRCode   = "qrcode"   // 二维码
	ViewSourceOther    = "other"     // 其他
)

// JobView 浏览记录表
type JobView struct {
	ID              uint      `gorm:"primarykey" json:"id"`
	RegionID        uint      `gorm:"index;not null;default:1" json:"region_id"`                            // 地区 ID
	CreatedAt       time.Time `gorm:"index" json:"created_at"`                                              // 创建时间
	UserID          uint      `gorm:"not null;index" json:"user_id"`                                       // 浏览者 ID
	JobID           uint      `gorm:"index" json:"job_id"`                                                  // 关联职位 ID
	CompanyID       uint      `gorm:"index" json:"company_id"`                                              // 关联公司 ID
	ResumeID        uint      `gorm:"index" json:"resume_id"`                                                // 关联简历 ID
	RecruiterID     uint      `gorm:"index" json:"recruiter_id"`                                            // 关联招聘者 ID
	ViewType        string    `gorm:"size:16;default:'job';index" json:"view_type"`                        // 浏览类型
	Source          string    `gorm:"size:32;default:'list';index" json:"source"`                           // 来源
	IP              string    `gorm:"size:64" json:"ip"`                                                     // IP 地址
	UserAgent       string    `gorm:"size:255" json:"user_agent"`                                          // User-Agent
	DurationSeconds int       `gorm:"default:0" json:"duration_seconds"`                                    // 停留时长（秒）
}

// TableName 表名（job_ 前缀）
func (JobView) TableName() string { return "job_views" }
