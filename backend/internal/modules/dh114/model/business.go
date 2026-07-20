// Package model 商户详细信息表（1:1 关联 dh114s）
// 营业执照/法人/注册资本/营业范围/营业时间汇总/设施/价格区间
package model

import (
	"time"

	"wuchang-tongcheng/internal/pkg/database"
)

// Dh114Business 商户详细信息表
type Dh114Business struct {
	database.RegionBaseModel // 含 id/region_id/created_at/updated_at/deleted_at
	Dh114ID uint `gorm:"not null;index;uniqueIndex:uniq_dh114_business_dh114_id" json:"dh114_id"` // 关联 dh114s.id（1:1）

	// === 营业执照信息 ===
	BusinessName    string `gorm:"size:200;not null;default:''" json:"business_name"`       // 营业执照名称
	LicenseNo       string `gorm:"size:64;not null;default:'';index" json:"license_no"`     // 营业执照号
	LicenseImage    string `gorm:"size:255;not null;default:''" json:"license_image"`       // 营业执照图片
	LegalPerson     string `gorm:"size:64;not null;default:''" json:"legal_person"`         // 法人代表
	LegalPersonIDCard string `gorm:"size:32;not null;default:''" json:"legal_person_id_card"` // 法人身份证号
	BusinessScope   string `gorm:"type:text" json:"business_scope"`                          // 经营范围
	RegisteredCapital float64 `gorm:"type:decimal(14,2);default:0" json:"registered_capital"` // 注册资本
	EstablishedDate *time.Time `gorm:"type:date" json:"established_date"`                   // 成立日期
	RegisteredAddress string `gorm:"size:500;not null;default:''" json:"registered_address"` // 注册地址

	// === 营业时间 ===
	OpeningHours   string `gorm:"size:32;not null;default:''" json:"opening_hours"`   // 营业开始时间（如 09:00）
	ClosingHours   string `gorm:"size:32;not null;default:''" json:"closing_hours"`   // 营业结束时间（如 22:00）
	OpenAllDay     bool   `gorm:"not null;default:false" json:"open_all_day"`         // 全天营业
	ClosedDays     JSONB  `gorm:"type:jsonb" json:"closed_days"`                       // 休息日 JSON（如 [0,6] 表示周日、周六）

	// === 价格信息 ===
	PriceAvg       float64 `gorm:"type:decimal(10,2);default:0" json:"price_avg"`        // 人均消费
	PriceRangeMin  float64 `gorm:"type:decimal(10,2);default:0" json:"price_range_min"`  // 最低消费
	PriceRangeMax  float64 `gorm:"type:decimal(10,2);default:0" json:"price_range_max"`  // 最高消费

	// === 联系方式扩展 ===
	Website   string `gorm:"size:255;not null;default:''" json:"website"`   // 官网
	Wechat    string `gorm:"size:64;not null;default:''" json:"wechat"`     // 微信号
	WechatQR  string `gorm:"size:255;not null;default:''" json:"wechat_qr"` // 微信二维码
	Email     string `gorm:"size:128;not null;default:''" json:"email"`     // 邮箱

	// === 设施服务（JSONB）===
	Facilities JSONB `gorm:"type:jsonb" json:"facilities"` // 设施列表 JSON（WiFi/停车场/包间/刷卡等）

	// === 认证状态 ===
	VerificationStatus int       `gorm:"default:0;index" json:"verification_status"` // 0待审 1通过 2拒绝 3过期
	VerifiedAt         *time.Time `gorm:"index" json:"verified_at"`                  // 认证时间
	ValidUntil         *time.Time `gorm:"type:date;index" json:"valid_until"`         // 认证有效期
}

// TableName 表名
func (Dh114Business) TableName() string { return "dh114_business" }
