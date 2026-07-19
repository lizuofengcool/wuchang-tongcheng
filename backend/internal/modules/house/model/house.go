// Package model 同城房屋租售数据模型
// 依据 v3.2.1 架构方案第五章：对标贝壳/链家/安居客/我爱我家/58房产
// 依据需求文档 1.10：4 维数据隔离（region_id 地区隔离 + user_id 用户隔离）
// 依据需求文档 7.1：通用字段 id/region_id/created_at/updated_at/deleted_at + status + audit_status
// 依据需求文档 7.2：主表 houses（保持兼容已发布数据）
package model

import (
	"time"

	"wuchang-tongcheng/internal/pkg/database"
)

// === 状态常量 ===
const (
	StatusDraft     = 0 // 草稿
	StatusPublished = 1 // 已发布
	StatusOffline   = 2 // 已下架
	StatusExpired   = 3 // 已过期
	StatusDeleted   = 4 // 已删除
)

// === 审核状态常量（依据需求文档 7.1） ===
const (
	AuditPending  = 0 // 待审
	AuditApproved = 1 // 通过
	AuditRejected = 2 // 拒绝
)

// === 发布类型常量 ===
const (
	ListingTypeRent     = "rent"     // 出租
	ListingTypeSale     = "sale"     // 出售
	ListingTypeTransfer = "transfer" // 转让
)

// === 物业类型常量 ===
const (
	PropertyTypeResidential = "residential" // 住宅
	PropertyTypeApartment   = "apartment"   // 公寓
	PropertyTypeVilla       = "villa"       // 别墅
	PropertyTypeLoft        = "loft"        // LOFT
	PropertyTypeOffice      = "office"      // 写字楼
	PropertyTypeShop        = "shop"        // 商铺
)

// === 来源类型常量 ===
const (
	SourceTypePersonal  = "personal"  // 个人
	SourceTypeAgent     = "agent"     // 经纪人
	SourceTypeDeveloper = "developer" // 开发商
)

// === 租赁类型常量 ===
const (
	RentTypeEntire = "entire" // 整租
	RentTypeShared = "shared" // 合租
)

// === 押金类型常量 ===
const (
	DepositTypeNone               = "none"                   // 无押金
	DepositTypeOneMonth           = "one_month"              // 押一
	DepositTypeTwoMonth           = "two_month"              // 押二
	DepositTypeThreeMonth         = "three_month"            // 押三
	DepositTypePayOneDepositThree = "pay_one_deposit_three"  // 押一付三
)

// === 付款方式常量 ===
const (
	PaymentMethodMonthly   = "monthly"   // 月付
	PaymentMethodQuarterly = "quarterly" // 季付
	PaymentMethodHalfYear  = "half_year" // 半年付
	PaymentMethodYearly    = "yearly"    // 年付
	PaymentMethodOneTime   = "one_time"  // 一次性
)

// === 租金单位常量 ===
const (
	RentUnitMonth = "month" // 月
	RentUnitYear  = "year"  // 年
	RentUnitDay   = "day"   // 日
)

// === 楼层段常量 ===
const (
	FloorTypeLow  = "low"  // 低楼层
	FloorTypeMid  = "mid"  // 中楼层
	FloorTypeHigh = "high" // 高楼层
)

// === 朝向常量 ===
const (
	OrientationEast       = "east"
	OrientationSouth      = "south"
	OrientationWest       = "west"
	OrientationNorth      = "north"
	OrientationSouthNorth = "south_north"
	OrientationEastWest   = "east_west"
	OrientationNorthEast  = "north_east"
	OrientationNorthWest  = "north_west"
	OrientationSouthEast  = "south_east"
	OrientationSouthWest  = "south_west"
)

// === 装修类型常量 ===
const (
	DecorationRough  = "rough"  // 毛坯
	DecorationSimple = "simple" // 简装
	DecorationFine   = "fine"   // 精装
	DecorationLuxury = "luxury" // 豪装
)

// === 产权类型常量 ===
const (
	PropertyOwnershipCommercial = "commercial"     // 商品房
	PropertyOwnershipReformed   = "reformed"       // 房改房
	PropertyOwnershipAffordable = "affordable"     // 经济适用房
	PropertyOwnershipSmallProp  = "small_property" // 小产权
)

// House 房屋租售主表（保持 houses 表名兼容已发布数据）
type House struct {
	database.RegionBaseModel // 含 id/region_id/created_at/updated_at/deleted_at（地区隔离）

	// === 基础信息 ===
	Title       string     `gorm:"size:200;not null" json:"title"`         // 标题
	Content     string     `gorm:"type:text" json:"content"`              // 详情描述
	CoverImage  string     `gorm:"size:255" json:"cover_image"`           // 封面图
	UserID      uint       `gorm:"index;not null" json:"user_id"`         // 发布者 ID
	UserName    string     `gorm:"size:50" json:"user_name"`              // 发布者昵称
	UserPhone   string     `gorm:"size:20" json:"user_phone"`             // 发布者手机
	UserAvatar  string     `gorm:"size:255" json:"user_avatar"`           // 发布者头像
	Status      int        `gorm:"default:0;index" json:"status"`         // 0草稿 1已发布 2下架 3过期 4删除
	AuditStatus int        `gorm:"default:0;index" json:"audit_status"`   // 0待审 1通过 2拒绝
	AuditReason string     `gorm:"size:500" json:"audit_reason"`          // 审核拒绝原因
	PublishedAt *time.Time `gorm:"index" json:"published_at"`             // 发布时间

	// === 交易类型/发布类型 ===
	ListingType  string `gorm:"size:16;not null;default:'rent';index" json:"listing_type"`            // rent/sale/transfer
	PropertyType string `gorm:"size:32;not null;default:'residential';index" json:"property_type"`    // residential/apartment/villa/loft/office/shop
	SourceType   string `gorm:"size:16;not null;default:'personal';index" json:"source_type"`         // personal/agent/developer

	// === 租金相关 ===
	RentPrice      float64 `gorm:"type:decimal(12,2);default:0;index" json:"rent_price"`     // 租金
	RentUnit       string  `gorm:"size:16;not null;default:'month'" json:"rent_unit"`        // month/year/day
	RentType       string  `gorm:"size:16;not null;default:'entire';index" json:"rent_type"` // entire/shared
	DepositType    string  `gorm:"size:32;not null;default:'one_month'" json:"deposit_type"` // none/one_month/two_month/three_month/pay_one_deposit_three
	PaymentMethod  string  `gorm:"size:16;not null;default:'monthly'" json:"payment_method"`  // monthly/quarterly/half_year/yearly/one_time
	RentNegotiable bool    `gorm:"not null;default:false" json:"rent_negotiable"`            // 租金可议
	RentMinMonths  int     `gorm:"not null;default:0" json:"rent_min_months"`                // 最短租期（月）
	RentMaxMonths  int     `gorm:"not null;default:0" json:"rent_max_months"`                // 最长租期（月）

	// === 售价相关 ===
	SalePrice      float64 `gorm:"type:decimal(14,2);default:0;index" json:"sale_price"`     // 售价（元）
	SaleNegotiable bool    `gorm:"not null;default:false" json:"sale_negotiable"`            // 售价可议
	AveragePrice   float64 `gorm:"type:decimal(10,2);default:0;index" json:"average_price"`  // 每平米均价
	OriginalPrice  float64 `gorm:"type:decimal(14,2);default:0" json:"original_price"`       // 原价/挂牌价

	// === 户型 ===
	Rooms     int    `gorm:"not null;default:0;index" json:"rooms"`     // 室
	Halls     int    `gorm:"not null;default:0" json:"halls"`           // 厅
	Bathrooms int    `gorm:"not null;default:0" json:"bathrooms"`       // 卫
	Kitchens int    `gorm:"not null;default:0" json:"kitchens"`         // 厨
	Balconies int   `gorm:"not null;default:0" json:"balconies"`        // 阳台数
	Layout    string `gorm:"size:32;not null;default:''" json:"layout"` // 户型文本（3室2厅2卫）

	// === 面积 ===
	BuildingArea float64 `gorm:"type:decimal(10,2);default:0;index" json:"building_area"` // 建筑面积（㎡）
	InnerArea    float64 `gorm:"type:decimal(10,2);default:0" json:"inner_area"`          // 套内面积（㎡）
	PoolRatio    float64 `gorm:"type:decimal(5,2);default:0" json:"pool_ratio"`           // 公摊比例 0.00-1.00
	UsableArea   float64 `gorm:"type:decimal(10,2);default:0" json:"usable_area"`         // 使用面积（㎡）

	// === 楼层/朝向 ===
	Floor       int    `gorm:"not null;default:0;index" json:"floor"`                    // 所在楼层
	TotalFloor  int    `gorm:"not null;default:0;index" json:"total_floor"`              // 总楼层
	FloorType   string `gorm:"size:16;not null;default:'mid';index" json:"floor_type"`   // low/mid/high
	Orientation string `gorm:"size:32;not null;default:'';index" json:"orientation"`     // east/south/west/north/south_north 等
	HasElevator bool   `gorm:"not null;default:false" json:"has_elevator"`               // 是否有电梯

	// === 装修/产权/年限 ===
	Decoration        string `gorm:"size:16;not null;default:'rough';index" json:"decoration"`                    // rough/simple/fine/luxury
	PropertyOwnership string `gorm:"size:32;not null;default:'commercial';index" json:"property_ownership"`        // commercial/reformed/affordable/small_property
	PropertyYears     int    `gorm:"not null;default:70" json:"property_years"`                                    // 产权年限 70/50/40
	BuildingYear      int    `gorm:"not null;default:0;index" json:"building_year"`                                // 建造年代
	BuildingAge       int    `gorm:"not null;default:0" json:"building_age"`                                       // 房龄（年）

	// === 关联 ID ===
	CommunityID *uint `gorm:"index" json:"community_id"` // 关联小区 ID（house_communities.id）
	AgentID     *uint `gorm:"index" json:"agent_id"`     // 关联经纪人 ID（house_agents.id）
	CategoryID  *uint `gorm:"index" json:"category_id"`  // 关联房源分类 ID（house_categories.id）

	// === 地理位置冗余（小区信息冗余存储以便 LBS 检索）===
	City             string  `gorm:"size:64;not null;default:'';index" json:"city"`             // 城市
	District         string  `gorm:"size:64;not null;default:'';index" json:"district"`         // 行政区
	BusinessDistrict string  `gorm:"size:128;not null;default:''" json:"business_district"`     // 商圈
	Address          string  `gorm:"size:500;not null;default:''" json:"address"`              // 详细地址
	Latitude         float64 `gorm:"type:decimal(10,7);default:0" json:"latitude"`             // 纬度
	Longitude        float64 `gorm:"type:decimal(10,7);default:0" json:"longitude"`            // 经度

	// === 互动统计 ===
	ViewCount    int        `gorm:"default:0" json:"view_count"`      // 浏览数
	FavCount     int        `gorm:"default:0" json:"fav_count"`        // 收藏数
	ContactCount int        `gorm:"default:0" json:"contact_count"`    // 联系数
	ShareCount   int        `gorm:"default:0" json:"share_count"`      // 分享数
	ViewingCount int        `gorm:"default:0" json:"viewing_count"`    // 看房预约数
	LastViewingAt *time.Time `gorm:"index" json:"last_viewing_at"`     // 最近看房时间

	// === 风控 ===
	ContentHash string `gorm:"size:64;index" json:"content_hash"`   // 图文指纹
	RiskScore   int    `gorm:"default:0;index" json:"risk_score"`   // 风险评分 0-100
	SameHouseID string `gorm:"size:64;index" json:"same_house_id"`  // 同房源识别 ID

	// === 视频/VR/全景 ===
	VideoURL    string `gorm:"size:255" json:"video_url"`    // 视频 URL
	VideoCover  string `gorm:"size:255" json:"video_cover"`  // 视频封面
	VRURL       string `gorm:"size:255" json:"vr_url"`       // VR 看房 URL
	PanoramaURL string `gorm:"size:255" json:"panorama_url"` // 360° 全景图 URL

	// === 配套设施/标签（JSONB）===
	Facilities JSONB `gorm:"type:jsonb" json:"facilities"`  // 配套设施 ID 数组 JSON
	Tags       JSONB `gorm:"type:jsonb" json:"tags"`        // 标签数组（最多 5 个）
	NearbyPOIs JSONB `gorm:"type:jsonb" json:"nearby_pois"` // 附近 POI JSON（地铁/学校/医院/商超等）

	// === 运营字段 ===
	Featured       bool    `gorm:"default:false;index" json:"featured"`                       // 精选推荐
	Picked         bool    `gorm:"default:false;index" json:"picked"`                         // 运营甄选
	Verified       bool    `gorm:"default:false;index" json:"verified"`                       // 官方验真
	PromotionLevel int     `gorm:"default:0" json:"promotion_level"`                          // 推广等级 0-10
	TrafficWeight  float64 `gorm:"type:decimal(3,2);default:1.00" json:"traffic_weight"`      // 流量权重 0.00-9.99

	// === 真房源认证 ===
	RealHouseVerified   bool       `gorm:"default:false;index" json:"real_house_verified"`     // 真房源认证
	RealHouseVerifiedAt *time.Time `gorm:"index" json:"real_house_verified_at"`                // 真房源认证时间

	// Distance 仅在"附近"查询时由 SQL 计算并回填，非持久化字段（公里）
	Distance float64 `gorm:"-" json:"-"`
}

// TableName 表名（依据需求文档 7.2：主表 {module}s）
// 注意：保持现有 houses 表名不变以兼容已发布数据
func (House) TableName() string { return "houses" }
