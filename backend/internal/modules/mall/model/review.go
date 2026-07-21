// Package model 同城商城 - 商品评价
// 依据需求文档 1.10：4 维数据隔离（region_id + user_id 买家 + product_id 商品 + order_id 订单）
package model

import (
	"time"

	"wuchang-tongcheng/internal/pkg/database"
)

// Review 评价表
type Review struct {
	database.RegionBaseModel // 含 id/region_id/created_at/updated_at/deleted_at

	// === 关联 ===
	ProductID    uint `gorm:"index;not null" json:"product_id"`    // 商品 SPU ID
	SkuID        uint `gorm:"index;not null;default:0" json:"sku_id"` // SKU ID
	OrderID      uint `gorm:"index;not null" json:"order_id"`      // 订单 ID
	OrderNo      string `gorm:"size:32;not null;default:''" json:"order_no"` // 订单号（冗余）
	OrderItemID  uint `gorm:"index;not null;default:0" json:"order_item_id"` // 订单明细 ID
	UserID       uint `gorm:"index;not null" json:"user_id"`       // 评价人 ID
	ShopID       uint `gorm:"index;not null" json:"shop_id"`       // 店铺 ID

	// === 评价人信息（冗余） ===
	UserName  string `gorm:"size:50;not null;default:''" json:"user_name"`   // 评价人昵称
	UserAvatar string `gorm:"size:255;not null;default:''" json:"user_avatar"` // 评价人头像
	IsAnonymous bool  `gorm:"default:false" json:"is_anonymous"`              // 是否匿名

	// === 评价内容 ===
	Rating     int    `gorm:"not null;default:5;index" json:"rating"`        // 评分 1-5
	Content    string `gorm:"type:text" json:"content"`                       // 评价内容
	Images     JSONB  `gorm:"type:jsonb" json:"images"`                       // 评价图片
	Video      string `gorm:"size:255;not null;default:''" json:"video"`      // 评价视频 URL

	// === 规格（冗余） ===
	SkuName  string `gorm:"size:200;not null;default:''" json:"sku_name"` // SKU 名（如：红色-XL）
	SkuSpecs string `gorm:"size:500;not null;default:''" json:"sku_specs"` // SKU 规格文本

	// === 卖家回复 ===
	Reply       string     `gorm:"type:text" json:"reply"`                       // 卖家回复内容
	ReplyAt     *time.Time `gorm:"index" json:"reply_at"`                        // 回复时间
	ReplyUserID uint       `gorm:"index" json:"reply_user_id"`                   // 回复人 ID

	// === 评价标签 ===
	Tags JSONB `gorm:"type:jsonb" json:"tags"` // 评价标签（如：好评/质量好/物流快）

	// === 状态 ===
	Status      int    `gorm:"default:0;index" json:"status"`         // 0待审 1通过 2拒绝 3隐藏
	AuditReason string `gorm:"size:500;not null;default:''" json:"audit_reason"` // 审核拒绝原因
	HasSellerReply bool `gorm:"default:false;index" json:"has_seller_reply"`     // 是否已回复

	// === 互动 ===
	LikeCount   int    `gorm:"not null;default:0" json:"like_count"`   // 点赞数
	DislikeCount int   `gorm:"not null;default:0" json:"dislike_count"` // 踩数
	ReplyCount  int    `gorm:"not null;default:0" json:"reply_count"`   // 追评数

	// === 追评 ===
	AppendContent string     `gorm:"type:text" json:"append_content"`     // 追评内容
	AppendImages  JSONB      `gorm:"type:jsonb" json:"append_images"`     // 追评图片
	AppendAt      *time.Time `gorm:"index" json:"append_at"`              // 追评时间

	// === 类型 ===
	Type string `gorm:"size:32;not null;default:'product';index" json:"type"` // product 商品评价 / logistics 物流评价 / service 服务评价

	// === 风控 ===
	ContentHash string `gorm:"size:64;index" json:"content_hash"` // 图文指纹
	RiskScore   int    `gorm:"default:0;index" json:"risk_score"` // 风险评分 0-100
}

// TableName 表名
func (Review) TableName() string { return "mall_reviews" }
