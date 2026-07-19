// Package model 商家店铺 + 店铺粉丝（对标转转商家版）
package model

import (
	"time"

	"wuchang-tongcheng/internal/pkg/database"
)

// === 店铺状态常量 ===
const (
	ShopStatusPending  = 0 // 待审核
	ShopStatusApproved = 1 // 已通过
	ShopStatusRejected = 2 // 已拒绝
	ShopStatusFrozen   = 3 // 已冻结
	ShopStatusClosed   = 4 // 已关闭
)

// === 店铺等级常量 ===
const (
	ShopLevelNormal  = 0 // 普通商家
	ShopLevelVerified = 1 // 认证商家
	ShopLevelGold    = 2 // 金牌商家
	ShopLevelDiamond = 3 // 钻石商家
)

// ErshouShop 商家店铺表（对标转转商家版）
type ErshouShop struct {
	database.RegionBaseModel // id/region_id/created_at/updated_at/deleted_at
	UserID         uint       `gorm:"not null;uniqueIndex" json:"user_id"`                       // 店主用户ID（一对一）
	ShopName       string     `gorm:"size:128;not null;index" json:"shop_name"`                // 店铺名称
	Logo           string     `gorm:"size:255" json:"logo"`                                       // 店铺 Logo URL
	Banner         string     `gorm:"size:255" json:"banner"`                                    // 店铺 Banner URL
	Description    string     `gorm:"type:text" json:"description"`                              // 店铺简介
	Level          int        `gorm:"default:0;index" json:"level"`                              // 0普通 1认证 2金牌 3钻石
	Status         int        `gorm:"default:0;index" json:"status"`                              // 0待审 1通过 2拒绝 3冻结 4关闭
	ContactName    string     `gorm:"size:50" json:"contact_name"`                              // 联系人姓名
	ContactPhone   string     `gorm:"size:20" json:"contact_phone"`                            // 联系电话
	ContactWechat  string     `gorm:"size:50" json:"contact_wechat"`                            // 微信号
	Address        string     `gorm:"size:500" json:"address"`                                    // 店铺地址
	Latitude       float64    `gorm:"type:decimal(10,7)" json:"latitude"`                         // 纬度
	Longitude      float64    `gorm:"type:decimal(10,7)" json:"longitude"`                        // 经度
	BusinessLicense string    `gorm:"size:255" json:"business_license"`                          // 营业执照 URL
	LicenseNo      string     `gorm:"size:64;index" json:"license_no"`                            // 执照编号
	IDCardFront    string     `gorm:"size:255" json:"id_card_front"`                              // 身份证正面 URL
	IDCardBack     string     `gorm:"size:255" json:"id_card_back"`                               // 身份证背面 URL
	VerifiedAt     *time.Time `gorm:"index" json:"verified_at"`                                    // 认证通过时间
	FollowerCount  int        `gorm:"default:0" json:"follower_count"`                              // 粉丝数
	ItemCount      int        `gorm:"default:0" json:"item_count"`                                  // 在售商品数
	SoldCount      int        `gorm:"default:0" json:"sold_count"`                                  // 累计售出数
	TotalAmount    float64    `gorm:"type:decimal(14,2);default:0" json:"total_amount"`             // 累计成交额
	GoodRate       float64    `gorm:"type:decimal(5,2);default:100.00" json:"good_rate"`           // 好评率
	Deposit        float64    `gorm:"type:decimal(12,2);default:0" json:"deposit"`                // 保证金
	Tags           JSONB      `gorm:"type:jsonb" json:"tags"`                                       // 店铺标签 ["官方","正品","极速发货"]
	ApprovedAt     *time.Time `gorm:"index" json:"approved_at"`                                    // 审核通过时间
	RejectedReason string     `gorm:"size:500" json:"rejected_reason"`                            // 拒绝原因
	ClosedAt       *time.Time `gorm:"index" json:"closed_at"`                                       // 关闭时间
}

// TableName 表名（ers_ 前缀）
func (ErshouShop) TableName() string { return "ers_shops" }

// ErshouShopFollower 店铺粉丝关联表
type ErshouShopFollower struct {
	ID        uint      `gorm:"primarykey" json:"id"`
	ShopID    uint      `gorm:"not null;uniqueIndex:uniq_shop_follower;index" json:"shop_id"`     // 店铺ID
	UserID    uint      `gorm:"not null;uniqueIndex:uniq_shop_follower;index" json:"user_id"`     // 粉丝用户ID
	Notify    bool      `gorm:"default:true" json:"notify"`                                       // 是否接收上新通知
	CreatedAt time.Time `json:"created_at"`
}

// TableName 表名（ers_ 前缀）
func (ErshouShopFollower) TableName() string { return "ers_shop_followers" }
