// Package model 商户评价表（对标大众点评/美团 5星评价）
// 评价/评分/图片/视频/商家回复
package model

import (
	"time"

	"wuchang-tongcheng/internal/pkg/database"
)

// Dh114Review 评价表
type Dh114Review struct {
	database.RegionBaseModel // 含 id/region_id/created_at/updated_at/deleted_at
	ReviewNo  string `gorm:"size:64;not null;uniqueIndex:uniq_dh114_reviews_no" json:"review_no"` // 评价单号
	Dh114ID   uint   `gorm:"not null;index" json:"dh114_id"`                                       // 商户 ID
	BusinessID uint  `gorm:"not null;default:0;index" json:"business_id"`                          // 商户详情 ID

	// === 评价人 ===
	ReviewerID     uint   `gorm:"not null;index" json:"reviewer_id"`            // 评价人 ID
	ReviewerName   string `gorm:"size:50;not null;default:''" json:"reviewer_name"` // 评价人昵称
	ReviewerAvatar string `gorm:"size:255;not null;default:''" json:"reviewer_avatar"` // 评价人头像
	ReviewerPhone  string `gorm:"size:20;not null;default:''" json:"reviewer_phone"`  // 评价人手机

	// === 评分 ===
	Rating            int `gorm:"not null;default:5;index" json:"rating"`             // 综合评分 1-5
	TasteRating      int  `gorm:"not null;default:5" json:"taste_rating"`           // 口味评分 1-5
	ServiceRating    int  `gorm:"not null;default:5" json:"service_rating"`         // 服务评分 1-5
	EnvironmentRating int `gorm:"not null;default:5" json:"environment_rating"`     // 环境评分 1-5

	// === 内容 ===
	Content   string `gorm:"type:text" json:"content"`                                  // 评价内容
	Images    JSONB  `gorm:"type:jsonb" json:"images"`                                  // 评价图片 JSON
	VideoURL  string `gorm:"size:255;not null;default:''" json:"video_url"`            // 视频 URL
	VideoCover string `gorm:"size:255;not null;default:''" json:"video_cover"`        // 视频封面
	Tags      JSONB  `gorm:"type:jsonb" json:"tags"`                                    // 评价标签 JSON

	// === 商家回复（冗余字段，子表 dh114_review_replies 存多条）===
	Reply     string     `gorm:"type:text" json:"reply"`                              // 商家回复
	RepliedAt *time.Time `gorm:"index" json:"replied_at"`                            // 回复时间
	HasReply  bool       `gorm:"not null;default:false;index" json:"has_reply"`     // 是否已回复

	// === 互动 ===
	LikeCount int   `gorm:"not null;default:0" json:"like_count"`                  // 点赞数
	LikedBy   JSONB `gorm:"type:jsonb" json:"liked_by"`                              // 点赞用户 ID 列表 JSON

	// === 状态 ===
	Status      int    `gorm:"default:0;index" json:"status"`                         // 0待审 1通过 2拒绝 3隐藏
	AuditStatus int    `gorm:"default:0;index" json:"audit_status"`                   // 审核状态
	AuditReason string `gorm:"size:500;not null;default:''" json:"audit_reason"`     // 审核拒绝原因

	// === 关联订单 ===
	OrderID    uint       `gorm:"not null;default:0;index" json:"order_id"`          // 关联订单 ID
	ConsumedAt *time.Time `gorm:"type:date" json:"consumed_at"`                       // 消费时间

	// === 评价类型 ===
	ReviewType string `gorm:"size:32;not null;default:'general';index" json:"review_type"` // general/order/visit
}

// TableName 表名
func (Dh114Review) TableName() string { return "dh114_reviews" }
