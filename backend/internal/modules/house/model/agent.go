// Package model 经纪人表（对标贝壳/链家）
// 姓名/手机/门店/评分/成交量/认证状态
package model

import (
	"time"

	"wuchang-tongcheng/internal/pkg/database"
)

// === 经纪人状态常量 ===
const (
	AgentStatusPending  = 0 // 待审核
	AgentStatusApproved = 1 // 已通过
	AgentStatusRejected = 2 // 已拒绝
	AgentStatusFrozen   = 3 // 已冻结
	AgentStatusRevoked  = 4 // 已撤销
)

// === 经纪人等级常量 ===
const (
	AgentLevelTrainee  = 0 // 实习
	AgentLevelJunior   = 1 // 初级
	AgentLevelSenior   = 2 // 高级
	AgentLevelMaster   = 3 // 资深
	AgentLevelPartner  = 4 // 合伙人
)

// === 在线状态常量 ===
const (
	AgentOffline    = 0 // 离线
	AgentOnline     = 1 // 在线
	AgentBusy       = 2 // 忙碌
	AgentAway       = 3 // 离开
)

// HouseAgent 经纪人表
type HouseAgent struct {
	database.RegionBaseModel // id/region_id/created_at/updated_at/deleted_at
	UserID           uint       `gorm:"not null;default:0;index" json:"user_id"`                              // 用户 ID
	Name             string     `gorm:"size:50;not null;index" json:"name"`                                   // 姓名
	Phone            string     `gorm:"size:20;not null;uniqueIndex" json:"phone"`                            // 手机
	Avatar           string     `gorm:"size:255;not null;default:''" json:"avatar"`                           // 头像
	Gender           string     `gorm:"size:16;not null;default:'unlimited'" json:"gender"`                   // 性别
	StoreID          uint       `gorm:"not null;default:0;index" json:"store_id"`                             // 门店 ID
	StoreName        string     `gorm:"size:128;not null;default:'';index" json:"store_name"`                 // 门店名
	Company          string     `gorm:"size:128;not null;default:'';index" json:"company"`                    // 公司名
	Title            string     `gorm:"size:64;not null;default:''" json:"title"`                             // 职位
	Level            int        `gorm:"not null;default:0;index" json:"level"`                                // 等级
	LicenseNo        string     `gorm:"size:64;not null;default:'';index" json:"license_no"`                  // 执业证号
	LicenseImage     string     `gorm:"size:255;not null;default:''" json:"license_image"`                    // 执业证图片
	IDCardFront      string     `gorm:"size:255;not null;default:''" json:"id_card_front"`                    // 身份证正面
	IDCardBack       string     `gorm:"size:255;not null;default:''" json:"id_card_back"`                     // 身份证反面
	BusinessCard     string     `gorm:"size:255;not null;default:''" json:"business_card"`                    // 名片
	Description      string     `gorm:"type:text" json:"description"`                                          // 个人简介
	GoodAt           JSONB      `gorm:"type:jsonb" json:"good_at"`                                             // 擅长领域
	ServiceArea      JSONB      `gorm:"type:jsonb" json:"service_area"`                                        // 服务区域
	Rating           float64    `gorm:"type:decimal(3,2);default:5.00;index" json:"rating"`                    // 评分
	RatingCount      int        `gorm:"not null;default:0" json:"rating_count"`                                // 评分人数
	ListingCount     int        `gorm:"not null;default:0" json:"listing_count"`                               // 房源数
	DealCount        int        `gorm:"not null;default:0" json:"deal_count"`                                  // 成交量
	TotalAmount      float64    `gorm:"type:decimal(14,2);default:0" json:"total_amount"`                      // 总成交额
	ResponseTime     int        `gorm:"not null;default:0" json:"response_time"`                               // 平均响应时间（秒）
	ResponseRate     float64    `gorm:"type:decimal(5,2);default:0" json:"response_rate"`                      // 响应率
	OnlineStatus     int        `gorm:"not null;default:0" json:"online_status"`                               // 在线状态
	LastActiveAt     *time.Time `gorm:"index" json:"last_active_at"`                                           // 最近活跃时间
	VerifiedAt       *time.Time `gorm:"index" json:"verified_at"`                                              // 实名认证时间
	ApprovedAt       *time.Time `gorm:"index" json:"approved_at"`                                              // 审核通过时间
	RejectedReason   string     `gorm:"size:500;not null;default:''" json:"rejected_reason"`                  // 拒绝原因
	Status           int        `gorm:"default:0;index" json:"status"`                                         // 0待审 1通过 2拒绝 3冻结 4撤销
	FollowerCount    int        `gorm:"default:0" json:"follower_count"`                                       // 粉丝数
	Tags             JSONB      `gorm:"type:jsonb" json:"tags"`                                                // 标签
}

// TableName 表名（house_ 前缀）
func (HouseAgent) TableName() string { return "house_agents" }
