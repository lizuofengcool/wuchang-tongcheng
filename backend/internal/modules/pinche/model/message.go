// Package model 拼车行程内消息数据模型
package model

import (
	"time"

	"wuchang-tongcheng/internal/pkg/database"
)

// PincheMessage 行程内消息
type PincheMessage struct {
	database.RegionBaseModel

	PincheID  uint  `gorm:"index;not null" json:"pinche_id"`
	BookingID *uint `gorm:"index" json:"booking_id"`
	TripID    *uint `gorm:"index" json:"trip_id"`

	SenderID     uint   `gorm:"index;not null" json:"sender_id"`
	SenderName   string `gorm:"size:50" json:"sender_name"`
	SenderAvatar string `gorm:"size:255" json:"sender_avatar"`
	ReceiverID   uint   `gorm:"index;not null" json:"receiver_id"`
	ReceiverName string `gorm:"size:50" json:"receiver_name"`

	MessageType string `gorm:"size:16;not null;default:'text'" json:"message_type"` // text/image/voice/system/location
	Content     string `gorm:"type:text" json:"content"`
	ImageURL    string `gorm:"size:255" json:"image_url"`
	VoiceURL    string `gorm:"size:255" json:"voice_url"`
	VoiceDuration int  `gorm:"not null;default:0" json:"voice_duration"`
	LocationLat  float64 `gorm:"type:decimal(10,7);default:0" json:"location_lat"`
	LocationLng  float64 `gorm:"type:decimal(10,7);default:0" json:"location_lng"`
	LocationAddress string `gorm:"size:255" json:"location_address"`

	IsRead      bool       `gorm:"default:false;index" json:"is_read"`
	ReadAt      *time.Time `gorm:"index" json:"read_at"`
	SystemEvent string     `gorm:"size:32" json:"system_event"`
}

// TableName 表名
func (PincheMessage) TableName() string { return "pinche_messages" }
