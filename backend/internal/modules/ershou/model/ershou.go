// Package model 同城二手物品数据模型
// 依据需求文档 2.2.A.10：商品发布/分类/搜索/留言/交易
// 依据需求文档 1.10：4 维数据隔离（region_id 地区隔离 + user_id 用户隔离）
// 依据需求文档 7.1：通用字段 id/region_id/created_at/updated_at/deleted_at + status + audit_status
// 依据需求文档 7.2：表命名 erhous / ershou_images / ershou_favorites / ershou_messages
// 依据 v3.2.1 架构方案第二章：对标闲鱼/转转/58同城/瓜子/贝壳/趣店等头部平台
package model

import (
	"time"

	"wuchang-tongcheng/internal/pkg/database"
)

// === 状态常量 ===
const (
	StatusDraft     = 0 // 草稿
	StatusPublished = 1 // 已发布
	StatusSold      = 2 // 已售出
	StatusOffline   = 3 // 已下架
	StatusExpired   = 4 // 已过期
)

// === 审核状态常量（依据需求文档 7.1） ===
const (
	AuditPending  = 0 // 待审
	AuditApproved = 1 // 通过
	AuditRejected = 2 // 拒绝
)

// === 成色常量 ===
const (
	ConditionNew       = "new"        // 全新
	ConditionAlmostNew = "almost_new" // 9成新
	ConditionUsed      = "used"       // 二手
	ConditionBroken    = "broken"     // 有瑕疵
)

// === 交易方式常量 ===
const (
	DeliveryFace    = "face"    // 当面交易
	DeliverySelf    = "self"    // 自提
	DeliveryExpress = "express" // 快递
)

// === 议价类型常量（v3.2.1 新增，对标闲鱼） ===
const (
	BargainNone   = "none"   // 不议价
	BargainSmall  = "small"  // 小幅议价
	BargainBig    = "big"    // 大幅议价
	BargainFixed  = "fixed"  // 一口价
)

// === 隐私可见性常量（v3.2.1 新增，对标闲鱼） ===
const (
	VisibilityPublic         = "public"          // 公开
	VisibilityFriends        = "friends"         // 仅好友
	VisibilityCityOnly       = "city_only"       // 仅同城
	VisibilityFollowersOnly  = "followers_only"  // 仅粉丝
)

// === 物品来源常量（v3.2.1 新增） ===
const (
	ItemSourcePersonal = "personal" // 个人闲置
	ItemSourceMerchant = "merchant" // 商家商品
	ItemSourceOverseas = "overseas" // 海外代购
	ItemSourceTaobao   = "taobao"   // 淘宝搬运
)

// Ershou 同城二手物品主表
type Ershou struct {
	database.RegionBaseModel // 含 id/region_id/created_at/updated_at/deleted_at（地区隔离）

	// === 基础信息 ===
	Title      string `gorm:"size:200;not null" json:"title"`       // 标题
	Content    string `gorm:"type:text" json:"content"`             // 详情描述
	CoverImage string `gorm:"size:255" json:"cover_image"`          // 封面图
	Summary    string `gorm:"size:500" json:"summary"`              // 摘要

	// === 发布者（用户隔离，依据 1.10.1） ===
	UserID     uint   `gorm:"index;not null" json:"user_id"`        // 发布者ID
	UserName   string `gorm:"size:50" json:"user_name"`             // 发布者昵称
	UserPhone  string `gorm:"size:20" json:"user_phone"`            // 发布者手机
	UserAvatar string `gorm:"size:255" json:"user_avatar"`          // 发布者头像

	// === 分类 ===
	CategoryID   uint   `gorm:"index" json:"category_id"`           // 分类ID（通用 category 表）
	CategoryName string `gorm:"size:50" json:"category_name"`        // 分类名（冗余，便于列表展示）

	// === 二手核心字段 ===
	Price         float64 `gorm:"type:decimal(12,2);default:0" json:"price"`           // 售价
	OriginalPrice float64 `gorm:"type:decimal(12,2);default:0" json:"original_price"`  // 原价（可选，用于展示折扣）
	PriceUnit     string  `gorm:"size:20;default:'元'" json:"price_unit"`               // 价格单位：元/面议/免费
	Condition     string  `gorm:"size:20;default:'used'" json:"condition"`              // 成色：new/almost_new/used/broken
	Brand         string  `gorm:"size:100" json:"brand"`                              // 品牌（可选，文本冗余）

	// === 联系方式 ===
	ContactPhone  string `gorm:"size:20" json:"contact_phone"`   // 联系电话
	ContactWechat string `gorm:"size:50" json:"contact_wechat"`  // 微信号（可选）

	// === 位置信息（PostGIS 附近查询用） ===
	Address   string  `gorm:"size:255" json:"address"`            // 详细地址
	Latitude  float64 `gorm:"type:decimal(10,7)" json:"latitude"`  // 纬度
	Longitude float64 `gorm:"type:decimal(10,7)" json:"longitude"` // 经度

	// === 交易方式 ===
	DeliveryMethod string `gorm:"size:50;default:'face'" json:"delivery_method"` // face/self/express

	// === 展示控制 ===
	IsUrgent     bool       `gorm:"default:false;index" json:"is_urgent"`  // 是否加急置顶
	ExpiryTime   *time.Time `gorm:"index" json:"expiry_time"`              // 过期时间
	ViewCount    int        `gorm:"default:0" json:"view_count"`           // 浏览量
	FavCount     int        `gorm:"default:0" json:"fav_count"`            // 收藏数
	MessageCount int        `gorm:"default:0" json:"message_count"`      // 留言数

	// === 状态（依据 7.1：status + audit_status） ===
	Status      int        `gorm:"default:0;index" json:"status"`         // 0草稿 1已发布 2已售出 3下架 4过期
	AuditStatus int        `gorm:"default:0;index" json:"audit_status"`   // 0待审 1通过 2拒绝
	AuditReason string     `gorm:"size:500" json:"audit_reason"`          // 审核拒绝原因
	PublishedAt *time.Time `gorm:"index" json:"published_at"`             // 发布时间

	// === v3.2.1 新增：SKU 规格（对标闲鱼/转转） ===
	SKUEnabled bool `gorm:"default:false;index" json:"sku_enabled"` // 是否启用多规格 SKU

	// === v3.2.1 新增：拍卖（对标闲鱼） ===
	IsAuction       bool       `gorm:"default:false;index" json:"is_auction"`        // 是否为拍卖商品
	AuctionStartTime *time.Time `gorm:"index" json:"auction_start_time"`             // 拍卖开始时间
	AuctionEndTime   *time.Time `gorm:"index" json:"auction_end_time"`               // 拍卖截拍时间

	// === v3.2.1 新增：议价（对标闲鱼） ===
	BargainType string `gorm:"size:20;default:'small'" json:"bargain_type"` // none/small/big/fixed

	// === v3.2.1 新增：物流（对标转转） ===
	DeliveryFee         float64 `gorm:"type:decimal(12,2);default:0" json:"delivery_fee"`       // 运费
	FreeShipping        bool    `gorm:"default:false;index" json:"free_shipping"`                // 是否包邮
	LogisticsTemplateID uint    `gorm:"index" json:"logistics_template_id"`                     // 物流模板ID

	// === v3.2.1 新增：担保交易（对标闲鱼/瓜子） ===
	EscrowEnabled bool `gorm:"default:false;index" json:"escrow_enabled"` // 是否启用担保交易

	// === v3.2.1 新增：分期付款（对标趣店/花呗） ===
	InstallmentEnabled bool `gorm:"default:false;index" json:"installment_enabled"` // 是否支持分期付款

	// === v3.2.1 新增：隐私设置（对标闲鱼） ===
	Visibility string `gorm:"size:20;default:'public'" json:"visibility"` // public/friends/city_only/followers_only

	// === v3.2.1 新增：互动开关 ===
	AllowComment bool `gorm:"default:true" json:"allow_comment"`  // 是否允许留言
	AllowShare   bool `gorm:"default:true" json:"allow_share"`    // 是否允许转发

	// === v3.2.1 新增：物品信息 ===
	SellReason         string `gorm:"size:50" json:"sell_reason"`                  // 出手原因：换新/搬家/缺钱/不喜欢
	ItemSource         string `gorm:"size:50;default:'personal'" json:"item_source"` // personal/merchant/overseas/taobao
	CertificationType  string `gorm:"size:50" json:"certification_type"`             // 鉴定类型：official/third_party/none
	CertificationNo    string `gorm:"size:100" json:"certification_no"`              // 鉴定证书编号

	// === v3.2.1 新增：包装附件 ===
	HasOriginalBox   bool `gorm:"default:false" json:"has_original_box"`     // 是否有原包装
	HasAccessories   bool `gorm:"default:false" json:"has_accessories"`     // 是否有配件
	HasWarrantyCard  bool `gorm:"default:false" json:"has_warranty_card"`   // 是否有保修卡
	HasInvoice       bool `gorm:"default:false" json:"has_invoice"`          // 是否有发票

	// === v3.2.1 新增：使用情况 ===
	UseDuration    string     `gorm:"size:50" json:"use_duration"`     // 使用时长：1个月/3个月/半年/1年/2年+
	RepairHistory  string     `gorm:"type:text" json:"repair_history"`  // 维修历史
	WarrantyExpire *time.Time `gorm:"index" json:"warranty_expire"`     // 保修到期日期

	// === v3.2.1 新增：估值参考（对标转转 AI 估值） ===
	EstimatedValue float64 `gorm:"type:decimal(12,2);default:0" json:"estimated_value"` // 估值参考价

	// === v3.2.1 新增：推广加权 ===
	PromotionLevel int     `gorm:"default:0;index" json:"promotion_level"`        // 推广等级 0-10
	TrafficWeight float64 `gorm:"type:decimal(3,2);default:1.00" json:"traffic_weight"` // 流量权重 0.00-9.99

	// === v3.2.1 新增：风控（对标转转） ===
	ContentHash string `gorm:"size:64;index" json:"content_hash"`  // 图文指纹（MD5/SHA256），用于重复发布检测
	RiskScore   int    `gorm:"default:0;index" json:"risk_score"` // 风险评分 0-100，<30 限制交易
	SameItemID  string `gorm:"size:64;index" json:"same_item_id"`  // 同款识别 ID（用于聚合）

	// === v3.2.1 新增：关联 ===
	ShopID  uint `gorm:"index" json:"shop_id"`  // 关联店铺ID（个人闲置为 0）
	BrandID uint `gorm:"index" json:"brand_id"` // 关联品牌库ID
	ModelID uint `gorm:"index" json:"model_id"` // 关联型号库ID

	// === v3.2.1 新增：视频支持 ===
	VideoURL    string `gorm:"size:255" json:"video_url"`     // 视频 URL
	VideoCover  string `gorm:"size:255" json:"video_cover"`  // 视频封面

	// === v3.2.1 新增：360° 展示 ===
	PanoramaURL string `gorm:"size:255" json:"panorama_url"` // 360° 全景图 URL

	// === v3.2.1 新增：多地交易 ===
	TradeLocations  JSONB `gorm:"type:jsonb" json:"trade_locations"`   // 多个 POI 交易地点 [{name,address,lat,lng}]
	PickupTimeSlots JSONB `gorm:"type:jsonb" json:"pickup_time_slots"` // 自提时间段 ["09:00-12:00","14:00-18:00"]

	// === v3.2.1 新增：标签冗余 ===
	Tags JSONB `gorm:"type:jsonb" json:"tags"` // 标签数组（最多5个），冗余存储便于查询

	// === v3.2.1 新增：运营字段 ===
	Featured bool `gorm:"default:false;index" json:"featured"` // 精选推荐
	Picked   bool `gorm:"default:false;index" json:"picked"`   // 运营甄选
	Verified bool `gorm:"default:false;index" json:"verified"`  // 官方验真

	// Distance 仅在"附近"查询时由 SQL 计算并回填，非持久化字段（公里）
	Distance float64 `gorm:"-" json:"-"`
}

// TableName 表名（依据需求文档 7.2：主表 {module}s）
// 注意：保持现有 erhous 表名不变以兼容已发布数据
func (Ershou) TableName() string { return "ershous" }

// ErshouImage 二手物品图片子表（依据 7.2：子表 {module}_{sub}）
type ErshouImage struct {
	ID        uint      `gorm:"primarykey" json:"id"`
	ErshouID  uint      `gorm:"not null;index" json:"ershou_id"`  // 关联二手物品ID
	URL       string    `gorm:"size:255;not null" json:"url"`     // 图片URL
	Sort      int       `gorm:"default:0" json:"sort"`            // 排序（越小越靠前）
	CreatedAt time.Time `json:"created_at"`
}

// TableName 表名
func (ErshouImage) TableName() string { return "ershou_images" }

// ErshouFavorite 收藏关联表（依据 7.2：关联表 {module}_{rel}）
type ErshouFavorite struct {
	ID        uint      `gorm:"primarykey" json:"id"`
	UserID    uint      `gorm:"not null;uniqueIndex:uniq_user_ershou_fav" json:"user_id"`
	ErshouID  uint      `gorm:"not null;uniqueIndex:uniq_user_ershou_fav" json:"ershou_id"`
	CreatedAt time.Time `json:"created_at"`
}

// TableName 表名
func (ErshouFavorite) TableName() string { return "ershou_favorites" }

// ErshouMessage 留言/咨询关联表（C端用户向发布者留言）
type ErshouMessage struct {
	ID         uint      `gorm:"primarykey" json:"id"`
	ErshouID   uint      `gorm:"not null;index" json:"ershou_id"`    // 关联二手物品ID
	FromUserID uint      `gorm:"not null;index" json:"from_user_id"` // 留言者ID
	FromName   string    `gorm:"size:50" json:"from_name"`           // 留言者昵称
	FromAvatar string    `gorm:"size:255" json:"from_avatar"`        // 留言者头像
	Content    string    `gorm:"type:text;not null" json:"content"`  // 留言内容
	IsRead     bool      `gorm:"default:false;index" json:"is_read"` // 发布者是否已读
	Status     int       `gorm:"default:1;index" json:"status"`      // 0删除 1正常
	CreatedAt  time.Time `json:"created_at"`
}

// TableName 表名
func (ErshouMessage) TableName() string { return "ershou_messages" }
