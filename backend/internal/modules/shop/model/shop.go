// Package model 商家模块数据模型
// 提供店铺、店铺相册、店铺评价的数据结构定义
package model

import (
	"time"

	"wuchang-tongcheng/internal/pkg/database"
)

// 营业状态
const (
	ShopStatusClosed  = 0 // 歇业
	ShopStatusOpen    = 1 // 营业中
	ShopStatusResting = 2 // 休息中
)

// 审核状态
const (
	AuditStatusPending = 0 // 待审核
	AuditStatusApproved = 1 // 通过
	AuditStatusRejected = 2 // 拒绝
)

// 评价审核状态
const (
	ReviewStatusPending  = 0 // 待审核
	ReviewStatusApproved = 1 // 通过
	ReviewStatusRejected = 2 // 拒绝
)

// Shop 店铺模型
type Shop struct {
	database.RegionBaseModel
	Name          string  `gorm:"size:100;not null" json:"name"`            // 店铺名称
	Logo          string  `gorm:"size:255" json:"logo"`                    // 店铺logo
	Description   string  `gorm:"type:text" json:"description"`           // 简介
	Phone         string  `gorm:"size:30" json:"phone"`                   // 联系电话
	Address       string  `gorm:"size:255" json:"address"`                // 地址
	Longitude     float64 `gorm:"type:decimal(10,6)" json:"longitude"`    // 经度
	Latitude      float64 `gorm:"type:decimal(10,6)" json:"latitude"`     // 纬度
	CategoryID    uint    `gorm:"index" json:"category_id"`               // 分类ID
	BusinessHours string  `gorm:"size:50" json:"business_hours"`          // 营业时间，如"08:00-22:00"
	Status        int     `gorm:"default:0;index" json:"status"`          // 营业状态 0歇业 1营业中 2休息中
	AuditStatus   int     `gorm:"default:0;index" json:"audit_status"`    // 审核状态 0待审核 1通过 2拒绝
	Rating        float32 `gorm:"type:decimal(2,1);default:0" json:"rating"` // 评分
	Views         int     `gorm:"default:0" json:"views"`                 // 浏览量
	IsRecommend   int     `gorm:"default:0;index" json:"is_recommend"`    // 是否推荐 0否 1是
	Sort          int     `gorm:"default:0" json:"sort"`                  // 排序（越大越靠前）
	UserID        uint    `gorm:"index" json:"user_id"`                   // 店铺所有者用户ID
}

// TableName 店铺表名
func (Shop) TableName() string {
	return "shops"
}

// ShopImage 店铺相册模型
type ShopImage struct {
	database.RegionBaseModel
	ShopID   uint   `gorm:"index;not null" json:"shop_id"` // 店铺ID
	ImageURL string `gorm:"size:255;not null" json:"image_url"` // 图片URL
	Sort     int    `gorm:"default:0" json:"sort"`         // 排序（越小越靠前）
}

// TableName 店铺相册表名
func (ShopImage) TableName() string {
	return "shop_images"
}

// ShopReview 店铺评价模型
type ShopReview struct {
	database.RegionBaseModel
	ShopID  uint       `gorm:"index;not null" json:"shop_id"`           // 店铺ID
	UserID  uint       `gorm:"index;not null" json:"user_id"`           // 评价用户ID
	Rating  int        `gorm:"not null" json:"rating"`                  // 评分 1-5
	Content string     `gorm:"type:text" json:"content"`                // 评价内容
	Reply   string     `gorm:"type:text" json:"reply"`                  // 商家回复
	ReplyAt *time.Time `json:"reply_at"`                                // 回复时间
	Status  int        `gorm:"default:0;index" json:"status"`           // 状态 0待审核 1通过 2拒绝
}

// TableName 店铺评价表名
func (ShopReview) TableName() string {
	return "shop_reviews"
}
