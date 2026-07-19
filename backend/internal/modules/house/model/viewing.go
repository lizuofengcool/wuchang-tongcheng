// Package model 看房预约表（对标贝壳/链家）
// 预约时间/经纪人/用户/在线线下/结果/评分
package model

import (
	"time"

	"wuchang-tongcheng/internal/pkg/database"
)

// === 看房状态常量 ===
const (
	ViewingStatusPending    = 0 // 待确认
	ViewingStatusConfirmed  = 1 // 已确认
	ViewingStatusInProgress = 2 // 进行中
	ViewingStatusCompleted  = 3 // 已完成
	ViewingStatusCanceled   = 4 // 已取消
	ViewingStatusNoShow     = 5 // 爽约
)

// === 看房类型常量 ===
const (
	ViewingTypeOffline = "offline" // 线下看房
	ViewingTypeOnline  = "online"  // 在线看房（视频/VR）
	ViewingTypeVR      = "vr"      // VR 看房
)

// === 看房结果常量 ===
const (
	ViewingResultPending    = "pending"     // 待反馈
	ViewingResultSatisfied  = "satisfied"   // 满意
	ViewingResultNeutral    = "neutral"     // 一般
	ViewingResultDissatisfied = "dissatisfied" // 不满意
	ViewingResultNoShow     = "no_show"     // 爽约
)

// HouseViewing 看房预约表
type HouseViewing struct {
	database.RegionBaseModel // id/region_id/created_at/updated_at/deleted_at
	ViewingNo        string     `gorm:"size:64;not null;uniqueIndex" json:"viewing_no"`                       // 预约单号
	HouseID          uint       `gorm:"not null;index" json:"house_id"`                                       // 关联房源 ID
	ListingID        uint       `gorm:"not null;default:0;index" json:"listing_id"`                           // 关联发布 ID
	CommunityID      uint       `gorm:"not null;default:0;index" json:"community_id"`                         // 关联小区 ID
	UserID           uint       `gorm:"not null;index:idx_house_viewings_user" json:"user_id"`                // 用户 ID
	UserName         string     `gorm:"size:50;not null;default:''" json:"user_name"`                         // 用户姓名
	UserPhone        string     `gorm:"size:20;not null;default:''" json:"user_phone"`                        // 用户手机
	UserAvatar       string     `gorm:"size:255;not null;default:''" json:"user_avatar"`                      // 用户头像
	AgentID          uint       `gorm:"not null;default:0;index:idx_house_viewings_agent" json:"agent_id"`    // 经纪人 ID
	AgentName        string     `gorm:"size:50;not null;default:''" json:"agent_name"`                        // 经纪人姓名
	AgentPhone       string     `gorm:"size:20;not null;default:''" json:"agent_phone"`                       // 经纪人手机
	ScheduledAt      *time.Time `gorm:"index" json:"scheduled_at"`                                            // 预约时间
	DurationMinutes  int        `gorm:"not null;default:30" json:"duration_minutes"`                          // 时长（分钟）
	ViewingType      string     `gorm:"size:32;not null;default:'offline';index" json:"viewing_type"`         // offline/online/vr
	OnlineURL        string     `gorm:"size:500;not null;default:''" json:"online_url"`                       // 在线看房 URL
	OnlinePassword   string     `gorm:"size:64;not null;default:''" json:"online_password"`                   // 在线看房密码
	MeetLocation     string     `gorm:"size:255;not null;default:''" json:"meet_location"`                    // 线下看房集合地点
	Remark           string     `gorm:"size:500;not null;default:''" json:"remark"`                           // 备注
	Status           int        `gorm:"default:0;index:idx_house_viewings_user;index:idx_house_viewings_agent;index" json:"status"` // 0待确认 1已确认 2进行中 3完成 4取消 5爽约
	Result           string     `gorm:"size:32;not null;default:'pending';index" json:"result"`               // pending/satisfied/neutral/dissatisfied/no_show
	Feedback         string     `gorm:"type:text" json:"feedback"`                                             // 反馈内容
	Rating           int        `gorm:"not null;default:0" json:"rating"`                                      // 评分 1-5
	AttendedAt       *time.Time `gorm:"index" json:"attended_at"`                                              // 实际到场时间
	CompletedAt      *time.Time `gorm:"index" json:"completed_at"`                                             // 完成时间
	CanceledAt       *time.Time `gorm:"index" json:"canceled_at"`                                              // 取消时间
	CanceledReason   string     `gorm:"size:500;not null;default:''" json:"canceled_reason"`                  // 取消原因
	CanceledBy       uint       `gorm:"not null;default:0" json:"canceled_by"`                                 // 取消人 ID
	ReminderSent     bool       `gorm:"not null;default:false" json:"reminder_sent"`                          // 是否已发提醒
	ReminderSentAt   *time.Time `gorm:"index" json:"reminder_sent_at"`                                        // 提醒发送时间
}

// TableName 表名（house_ 前缀）
func (HouseViewing) TableName() string { return "house_viewings" }
