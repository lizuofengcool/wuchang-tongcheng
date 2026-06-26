// Package model 同城头条数据模型
package model

import (
	"time"

	"wuchang-tongcheng/internal/pkg/database"
)

// News 同城头条模型
type News struct {
	database.RegionBaseModel
	Title       string    `gorm:"size:200;not null" json:"title"`           // 标题
	Content     string    `gorm:"type:text" json:"content"`                // 内容
	CoverImage  string    `gorm:"size:255" json:"cover_image"`             // 封面图
	Summary     string    `gorm:"size:500" json:"summary"`                 // 摘要
	AuthorID    uint      `gorm:"index" json:"author_id"`                  // 作者ID
	AuthorName  string    `gorm:"size:50" json:"author_name"`             // 作者名
	CategoryID  uint      `gorm:"index" json:"category_id"`                // 分类ID
	Tags        string    `gorm:"size:255" json:"tags"`                   // 标签（逗号分隔）
	ViewCount   int       `gorm:"default:0" json:"view_count"`            // 浏览量
	LikeCount   int       `gorm:"default:0" json:"like_count"`            // 点赞数
	Status      int       `gorm:"default:0;index" json:"status"`          // 状态 0草稿 1已发布 2下架
	PublishedAt *time.Time `gorm:"index" json:"published_at"`             // 发布时间
}

// TableName 表名
func (News) TableName() string {
	return "news"
}
