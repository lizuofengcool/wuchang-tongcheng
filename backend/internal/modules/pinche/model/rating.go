// Package model 拼车评价数据模型
package model

import (
	"time"

	"wuchang-tongcheng/internal/pkg/database"
)

// PincheRating 评价
type PincheRating struct {
	database.RegionBaseModel

	PincheID  uint  `gorm:"index;not null" json:"pinche_id"`
	BookingID *uint `gorm:"index" json:"booking_id"`
	TripID    *uint `gorm:"index" json:"trip_id"`

	RaterID     uint   `gorm:"index;not null" json:"rater_id"`
	RaterName   string `gorm:"size:50" json:"rater_name"`
	RaterAvatar string `gorm:"size:255" json:"rater_avatar"`
	RateeID     uint   `gorm:"index;not null" json:"ratee_id"`
	RateeName   string `gorm:"size:50" json:"ratee_name"`
	RateeAvatar string `gorm:"size:255" json:"ratee_avatar"`

	RatingType string `gorm:"size:32;not null;default:'passenger_to_driver';index" json:"rating_type"` // passenger_to_driver/driver_to_passenger
	Rating     int    `gorm:"not null;default:5" json:"rating"`
	Content    string `gorm:"type:text" json:"content"`
	Images     JSONB  `gorm:"type:jsonb" json:"images"`
	Tags       JSONB  `gorm:"type:jsonb" json:"tags"`
	IsAnonymous bool  `gorm:"default:false" json:"is_anonymous"`

	Reply    string     `gorm:"type:text" json:"reply"`
	ReplyAt  *time.Time `gorm:"index" json:"reply_at"`
	LikeCount int       `gorm:"not null;default:0" json:"like_count"`

	Status int `gorm:"default:0;index" json:"status"` // 0待审 1通过 2拒绝 3隐藏
}

// TableName 表名
func (PincheRating) TableName() string { return "pinche_ratings" }
