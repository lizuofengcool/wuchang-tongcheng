// Package model 商家回复评价表
// 支持一条评价多条回复（追问/追答）
package model

import (
	"wuchang-tongcheng/internal/pkg/database"
)

// Dh114ReviewReply 商家回复评价表
type Dh114ReviewReply struct {
	database.BaseModel // id/created_at/updated_at/deleted_at
	ReviewID    uint   `gorm:"not null;index" json:"review_id"`                                // 评价 ID
	Dh114ID     uint   `gorm:"not null;index" json:"dh114_id"`                                 // 商户 ID
	BusinessID  uint   `gorm:"not null;default:0;index" json:"business_id"`                    // 商户详情 ID
	ReplierID   uint   `gorm:"not null;index" json:"replier_id"`                              // 回复人 ID
	ReplierName string `gorm:"size:50;not null;default:''" json:"replier_name"`               // 回复人昵称
	ReplierAvatar string `gorm:"size:255;not null;default:''" json:"replier_avatar"`         // 回复人头像
	ReplierType  string `gorm:"size:16;not null;default:'merchant';index" json:"replier_type"` // merchant/user/admin
	Content     string `gorm:"type:text" json:"content"`                                       // 回复内容
	Images      JSONB  `gorm:"type:jsonb" json:"images"`                                       // 回复图片 JSON
	ParentID    uint   `gorm:"not null;default:0;index" json:"parent_id"`                      // 父回复 ID（用于追问）
	Status      int    `gorm:"default:1;index" json:"status"`                                   // 0隐藏 1显示
	LikeCount   int    `gorm:"not null;default:0" json:"like_count"`                            // 点赞数
}

// TableName 表名
func (Dh114ReviewReply) TableName() string { return "dh114_review_replies" }
