// Package model 团购表（对标大众点评/美团 限时抢购）
// 限时抢购/数量限制/使用规则/有效期
package model

import (
	"time"

	"wuchang-tongcheng/internal/pkg/database"
)

// Dh114Groupbuy 团购表
type Dh114Groupbuy struct {
	database.RegionBaseModel // 含 id/region_id/created_at/updated_at/deleted_at
	GroupbuyNo string `gorm:"size:64;not null;uniqueIndex:uniq_dh114_groupbuys_no" json:"groupbuy_no"` // 团购单号
	Dh114ID    uint   `gorm:"not null;index" json:"dh114_id"`                                            // 商户 ID
	BusinessID uint   `gorm:"not null;default:0;index" json:"business_id"`                              // 商户详情 ID

	// === 基本信息 ===
	Title       string `gorm:"size:200;not null" json:"title"`                  // 团购标题
	Description string `gorm:"type:text" json:"description"`                    // 团购描述
	CoverImage  string `gorm:"size:255;not null;default:''" json:"cover_image"` // 封面图
	Images      JSONB  `gorm:"type:jsonb" json:"images"`                        // 图片列表 JSON

	// === 价格 ===
	OriginalPrice float64 `gorm:"type:decimal(12,2);default:0;index" json:"original_price"` // 原价
	GroupbuyPrice float64 `gorm:"type:decimal(12,2);default:0;index" json:"groupbuy_price"`   // 团购价
	Discount      float64 `gorm:"type:decimal(3,2);default:0" json:"discount"`                 // 折扣（如 0.80 表示 8 折）

	// === 库存 ===
	TotalCount   int `gorm:"not null;default:0" json:"total_count"`     // 总数量
	SoldCount    int `gorm:"not null;default:0;index" json:"sold_count"` // 已售数量
	PerUserLimit int `gorm:"not null;default:0" json:"per_user_limit"`   // 每人限购（0 不限）

	// === 时间 ===
	StartTime  *time.Time `gorm:"index" json:"start_time"`   // 开始时间
	EndTime    *time.Time `gorm:"index" json:"end_time"`     // 结束时间
	ValidStart *time.Time `gorm:"type:date;index" json:"valid_start"` // 使用开始日期
	ValidEnd   *time.Time `gorm:"type:date" json:"valid_end"`         // 使用结束日期
	ValidWeekdays JSONB   `gorm:"type:jsonb" json:"valid_weekdays"`   // 可用星期 JSON（如 [1,2,3,4,5] 工作日）

	// === 使用规则 ===
	UseInstructions JSONB `gorm:"type:jsonb" json:"use_instructions"` // 使用规则 JSON
	UseTimeRanges   JSONB `gorm:"type:jsonb" json:"use_time_ranges"`  // 可用时段 JSON
	NeedReservation bool   `gorm:"not null;default:false" json:"need_reservation"` // 是否需要预约

	// === 互动 ===
	ViewCount int `gorm:"not null;default:0" json:"view_count"` // 浏览数
	FavCount  int `gorm:"not null;default:0" json:"fav_count"`  // 收藏数

	// === 状态 ===
	Status      int    `gorm:"default:0;index" json:"status"`                       // 0草稿 1已发布 2已售罄 3已下架 4已过期
	AuditStatus int    `gorm:"default:0;index" json:"audit_status"`                  // 审核状态
	AuditReason string `gorm:"size:500;not null;default:''" json:"audit_reason"`    // 审核拒绝原因
	PublishedAt *time.Time `gorm:"index" json:"published_at"`                       // 发布时间

	// === 运营 ===
	Featured bool `gorm:"not null;default:false;index" json:"featured"` // 精选推荐
}

// TableName 表名
func (Dh114Groupbuy) TableName() string { return "dh114_groupbuys" }
