// Package model 拍卖信息 + 出价记录（对标闲鱼）
// 支持起拍价 / 加价幅度 / 保留价 / 截拍时间 / 自动延拍
package model

import (
	"time"

	"wuchang-tongcheng/internal/pkg/database"
)

// === 拍卖状态常量 ===
const (
	AuctionStatusPending  = 0 // 待开拍
	AuctionStatusActive   = 1 // 进行中
	AuctionStatusEnded    = 2 // 已截拍
	AuctionStatusCanceled = 3 // 已取消
	AuctionStatusSold     = 4 // 已成交
	AuctionStatusFailed   = 5 // 流拍
)

// ErshouAuction 拍卖信息表
type ErshouAuction struct {
	database.RegionBaseModel // id/region_id/created_at/updated_at/deleted_at
	ErshouID        uint       `gorm:"not null;uniqueIndex" json:"ershou_id"`                        // 关联二手物品ID（一对一）
	StartPrice      float64    `gorm:"type:decimal(12,2);not null" json:"start_price"`                // 起拍价
	StepPrice       float64    `gorm:"type:decimal(12,2);default:1" json:"step_price"`                // 加价幅度
	ReservePrice    float64    `gorm:"type:decimal(12,2);default:0" json:"reserve_price"`             // 保留价（0=无保留价）
	BondAmount      float64    `gorm:"type:decimal(12,2);default:0" json:"bond_amount"`              // 参拍保证金（0=无需保证金）
	CurrentBidPrice float64    `gorm:"type:decimal(12,2);default:0;index" json:"current_bid_price"`   // 当前最高出价
	CurrentBidUserID uint      `gorm:"index" json:"current_bid_user_id"`                              // 当前最高出价者ID
	BidCount        int        `gorm:"default:0" json:"bid_count"`                                    // 出价次数
	WatcherCount    int        `gorm:"default:0" json:"watcher_count"`                                  // 围观人数
	Status          int        `gorm:"default:0;index" json:"status"`                                   // 0待开拍 1进行中 2已截拍 3取消 4成交 5流拍
	StartTime       *time.Time `gorm:"index" json:"start_time"`                                         // 开拍时间
	EndTime         *time.Time `gorm:"index" json:"end_time"`                                           // 截拍时间
	AutoExtendEnabled bool     `gorm:"default:true" json:"auto_extend_enabled"`                         // 是否启用自动延拍（最后1分钟出价自动延长5分钟）
	WinnerID        uint       `gorm:"index" json:"winner_id"`                                          // 成交者ID
	WinnerPrice     float64    `gorm:"type:decimal(12,2);default:0" json:"winner_price"`                // 成交价
	ClosedAt        *time.Time `gorm:"index" json:"closed_at"`                                          // 成交/关闭时间
}

// TableName 表名（ers_ 前缀）
func (ErshouAuction) TableName() string { return "ers_auctions" }

// ErshouAuctionBid 拍卖出价记录表
type ErshouAuctionBid struct {
	database.BaseModel // id/created_at/updated_at/deleted_at
	AuctionID uint    `gorm:"not null;index:idx_auction_bid_auction;uniqueIndex:uniq_auction_bid" json:"auction_id"` // 关联拍卖ID
	UserID    uint    `gorm:"not null;index:idx_auction_bid_auction" json:"user_id"`                                          // 出价用户ID
	BidPrice  float64 `gorm:"type:decimal(12,2);not null;uniqueIndex:uniq_auction_bid" json:"bid_price"`                     // 出价金额
	BidTime   time.Time `gorm:"not null" json:"bid_time"`                                                                          // 出价时间
	IsWinner  bool   `gorm:"default:false;index" json:"is_winner"`                                                                 // 是否为最终成交者
	IP        string `gorm:"size:64" json:"ip"`                                                                                    // 出价IP（防恶意）
	UserAgent string `gorm:"size:255" json:"user_agent"`                                                                          // 出价UA
	IsInvalid bool   `gorm:"default:false;index" json:"is_invalid"`                                                                // 是否为无效出价（违规/作弊）
	InvalidReason string `gorm:"size:500" json:"invalid_reason"`                                                                  // 无效原因
}

// TableName 表名（ers_ 前缀）
func (ErshouAuctionBid) TableName() string { return "ers_auction_bids" }
