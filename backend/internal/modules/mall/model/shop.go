// Package model 同城商城 - 店铺主表
// 依据需求文档 1.10：4 维数据隔离（region_id + user_id 商家）
// 对标淘宝/京东/拼多多商家店铺
package model

import (
	"time"

	"wuchang-tongcheng/internal/pkg/database"
)

// Shop 店铺主表
type Shop struct {
	database.RegionBaseModel // 含 id/region_id/created_at/updated_at/deleted_at

	// === 关联 ===
	UserID uint `gorm:"index;not null" json:"user_id"` // 店主用户 ID

	// === 基本信息 ===
	ShopName    string `gorm:"size:200;not null;default:'';index" json:"shop_name"`    // 店铺名
	Logo        string `gorm:"size:255;not null;default:''" json:"logo"`               // 店铺 LOGO
	Description string `gorm:"type:text" json:"description"`                           // 店铺简介
	ShopType    string `gorm:"size:32;not null;default:'personal';index" json:"shop_type"` // personal/enterprise/flagship

	// === 状态 ===
	Status      int        `gorm:"default:0;index" json:"status"`        // 0草稿 1已开业 2已关闭 3已冻结 4已过期
	AuditStatus int        `gorm:"default:0;index" json:"audit_status"`  // 0待审 1通过 2拒绝
	OpenedAt    *time.Time `gorm:"index" json:"opened_at"`               // 开业时间

	// === 联系方式 ===
	ContactName  string `gorm:"size:50;not null;default:''" json:"contact_name"`   // 联系人
	ContactPhone string `gorm:"size:32;not null;default:''" json:"contact_phone"`  // 联系电话
	ContactEmail string `gorm:"size:128;not null;default:''" json:"contact_email"` // 联系邮箱
	Wechat       string `gorm:"size:64;not null;default:''" json:"wechat"`         // 微信号
	QQ           string `gorm:"size:32;not null;default:''" json:"qq"`             // QQ 号

	// === 地址 ===
	Province  string  `gorm:"size:64;not null;default:'';index" json:"province"`   // 省
	City      string  `gorm:"size:64;not null;default:'';index" json:"city"`       // 市
	District  string  `gorm:"size:64;not null;default:''" json:"district"`         // 区/县
	Address   string  `gorm:"size:500;not null;default:''" json:"address"`         // 详细地址
	Latitude  float64 `gorm:"default:0" json:"latitude"`                            // 纬度
	Longitude float64 `gorm:"default:0" json:"longitude"`                           // 经度

	// === 资质 ===
	LicenseNo      string `gorm:"size:64;not null;default:''" json:"license_no"`        // 营业执照号
	LicenseImage   string `gorm:"size:255;not null;default:''" json:"license_image"`    // 营业执照图片
	LegalPerson    string `gorm:"size:64;not null;default:''" json:"legal_person"`      // 法人姓名
	LegalPersonID  string `gorm:"size:32;not null;default:''" json:"legal_person_id"`   // 法人身份证号

	// === 统计 ===
	ProductCount  int     `gorm:"not null;default:0" json:"product_count"`   // 商品数
	OrderCount    int64   `gorm:"not null;default:0" json:"order_count"`     // 订单数
	SaleAmount    float64 `gorm:"type:decimal(14,2);default:0" json:"sale_amount"` // 销售额
	Rating        float64 `gorm:"type:decimal(3,2);default:0" json:"rating"` // 评分
	ReviewCount   int     `gorm:"not null;default:0" json:"review_count"`    // 评价数
	FavoriteCount int     `gorm:"not null;default:0" json:"favorite_count"`  // 收藏数
	ViewCount     int     `gorm:"not null;default:0" json:"view_count"`      // 浏览数

	// === 推广 ===
	Featured       bool       `gorm:"default:false;index" json:"featured"`        // 精选
	Verified       bool       `gorm:"default:false;index" json:"verified"`        // 已认证
	PromotionLevel int        `gorm:"not null;default:0" json:"promotion_level"`  // 推广等级 0-10
	TrafficWeight  float64    `gorm:"type:decimal(4,2);default:1.0" json:"traffic_weight"` // 流量权重
	VerifiedAt     *time.Time `gorm:"index" json:"verified_at"`                   // 认证时间

	// === 扩展信息（JSONB） ===
	Banners       JSONB `gorm:"type:jsonb" json:"banners"`        // 店铺轮播图
	Tags          JSONB `gorm:"type:jsonb" json:"tags"`           // 店铺标签
	BusinessHours JSONB `gorm:"type:jsonb" json:"business_hours"` // 营业时间
	Facilities    JSONB `gorm:"type:jsonb" json:"facilities"`     // 店铺设施

	// === 风控 ===
	ContentHash string `gorm:"size:64;not null;default:''" json:"content_hash"` // 内容哈希
	RiskScore   int    `gorm:"not null;default:0" json:"risk_score"`            // 风险评分
}

// TableName 表名
func (Shop) TableName() string { return "mall_shops" }
