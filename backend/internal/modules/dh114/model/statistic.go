// Package model 统计表
// 日统计/商户统计/分类统计
package model

import (
	"time"

	"wuchang-tongcheng/internal/pkg/database"
)

// Dh114Statistic 统计表
type Dh114Statistic struct {
	database.RegionBaseModel // 含 id/region_id/created_at/updated_at/deleted_at
	StatDate   time.Time `gorm:"type:date;not null;index;uniqueIndex:uniq_dh114_statistics_date_type_target" json:"stat_date"` // 统计日期
	StatType   string    `gorm:"size:32;not null;default:'daily';index;uniqueIndex:uniq_dh114_statistics_date_type_target" json:"stat_type"` // daily/business/category
	Dh114ID    uint      `gorm:"not null;default:0;index;uniqueIndex:uniq_dh114_statistics_date_type_target" json:"dh114_id"` // 商户 ID（business 类型时使用）
	BusinessID uint      `gorm:"not null;default:0;index" json:"business_id"`                          // 商户详情 ID
	CategoryID *uint    `gorm:"index" json:"category_id"`                                            // 分类 ID（category 类型时使用）

	// === 互动统计 ===
	ViewCount    int64 `gorm:"not null;default:0" json:"view_count"`    // 浏览数
	FavCount     int64 `gorm:"not null;default:0" json:"fav_count"`     // 收藏数
	CallCount    int64 `gorm:"not null;default:0" json:"call_count"`    // 拨打数
	ShareCount   int64 `gorm:"not null;default:0" json:"share_count"`   // 分享数
	ContactCount int64 `gorm:"not null;default:0" json:"contact_count"` // 联系数
	VisitCount   int64 `gorm:"not null;default:0" json:"visit_count"`   // 到店数

	// === 评价统计 ===
	ReviewCount  int64   `gorm:"not null;default:0" json:"review_count"`  // 评价数
	NewReviewCount int64  `gorm:"not null;default:0" json:"new_review_count"` // 新增评价数
	AvgRating     float64 `gorm:"type:decimal(3,2);default:0" json:"avg_rating"` // 平均评分
	GoodRate      float64 `gorm:"type:decimal(5,2);default:0" json:"good_rate"` // 好评率

	// === 交易统计 ===
	GroupbuySold  int64   `gorm:"not null;default:0" json:"groupbuy_sold"`   // 团购销量
	GroupbuyAmount float64 `gorm:"type:decimal(14,2);default:0" json:"groupbuy_amount"` // 团购金额
	CouponIssued  int64   `gorm:"not null;default:0" json:"coupon_issued"`   // 优惠券领取数
	CouponUsed    int64   `gorm:"not null;default:0" json:"coupon_used"`     // 优惠券使用数
	OrderCount    int64   `gorm:"not null;default:0" json:"order_count"`    // 订单数
	OrderAmount   float64 `gorm:"type:decimal(14,2);default:0" json:"order_amount"` // 订单金额
}

// TableName 表名
func (Dh114Statistic) TableName() string { return "dh114_statistics" }
