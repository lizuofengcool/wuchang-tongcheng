// Package model 同城114商户黄页数据模型
// 依据 v3.2.1 架构方案：对标大众点评/美团/58同城
// 依据需求文档 1.10：4 维数据隔离（region_id 地区隔离 + user_id 用户隔离）
// 依据需求文档 7.1：通用字段 id/region_id/created_at/updated_at/deleted_at + status + audit_status
// 依据需求文档 7.2：主表 dh114s（黄页商户列表）
package model

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
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

// === 审核状态常量 ===
const (
	AuditPending  = 0 // 待审
	AuditApproved = 1 // 通过
	AuditRejected = 2 // 拒绝
)

// === 商户类型常量 ===
const (
	BusinessTypeRestaurant = "restaurant" // 餐饮
	BusinessTypeRetail     = "retail"     // 零售
	BusinessTypeService    = "service"    // 服务
	BusinessTypeEntertain  = "entertain"  // 娱乐
	BusinessTypeHotel      = "hotel"      // 酒店
	BusinessTypeMedical    = "medical"    // 医疗
	BusinessTypeEducation  = "education"  // 教育
	BusinessTypeLife       = "life"       // 生活服务
	BusinessTypeOther      = "other"      // 其他
)

// === 来源类型常量 ===
const (
	SourceTypePersonal = "personal" // 个人
	SourceTypeMerchant = "merchant" // 商户
	SourceTypeChain    = "chain"    // 连锁
)

// === 认证类型常量 ===
const (
	VerificationTypeLicense       = "business_license" // 营业执照
	VerificationTypeField         = "field"             // 实地认证
	VerificationTypeBrand         = "brand"             // 品牌授权
	VerificationTypeLicenseAndField = "license_field"  // 营业执照+实地
)

// === 认证状态常量 ===
const (
	VerificationStatusPending  = 0 // 待审
	VerificationStatusApproved = 1 // 通过
	VerificationStatusRejected = 2 // 拒绝
	VerificationStatusExpired  = 3 // 已过期
)

// === 评价状态常量 ===
const (
	ReviewStatusPending  = 0 // 待审
	ReviewStatusApproved = 1 // 通过
	ReviewStatusRejected = 2 // 拒绝
	ReviewStatusHidden   = 3 // 隐藏
)

// === 团购状态常量 ===
const (
	GroupbuyStatusDraft     = 0 // 草稿
	GroupbuyStatusPublished = 1 // 已发布
	GroupbuyStatusSoldOut   = 2 // 已售罄
	GroupbuyStatusOffline   = 3 // 已下架
	GroupbuyStatusExpired   = 4 // 已过期
)

// === 优惠券类型常量 ===
const (
	CouponTypeDiscount    = "discount"    // 折扣券
	CouponTypeFullReduction = "full_reduction" // 满减券
	CouponTypeCash         = "cash"        // 代金券
	CouponTypeGift         = "gift"        // 礼品券
)

// === 优惠券状态常量 ===
const (
	CouponStatusDraft     = 0 // 草稿
	CouponStatusPublished = 1 // 已发布
	CouponStatusSoldOut   = 2 // 已抢完
	CouponStatusOffline   = 3 // 已下架
	CouponStatusExpired   = 4 // 已过期
)

// === 收藏类型常量 ===
const (
	FavoriteTypeBusiness = "business" // 商户
	FavoriteTypeGroupbuy = "groupbuy" // 团购
	FavoriteTypeCoupon   = "coupon"   // 优惠券
)

// === 图片类型常量 ===
const (
	ImageTypeCover       = "cover"       // 封面
	ImageTypeEnvironment = "environment" // 环境
	ImageTypeDish        = "dish"        // 菜品
	ImageTypeLicense    = "license"     // 营业执照
	ImageTypeOther      = "other"       // 其他
)

// === 菜单类型常量 ===
const (
	MenuTypeDish    = "dish"    // 菜品
	MenuTypeService = "service" // 服务项目
)

// === 标签类型常量 ===
const (
	TagTypeBusiness = "business" // 商户标签
	TagTypeReview   = "review"   // 评价标签
	TagTypeFood     = "food"     // 美食标签
	TagTypeService  = "service"  // 服务标签
)

// === 推荐类型常量 ===
const (
	RecommendTypeHome         = "home"         // 首页推荐
	RecommendTypeCategory     = "category"     // 分类推荐
	RecommendTypeNearby       = "nearby"       // 附近推荐
	RecommendTypePersonalized = "personalized" // 个性化推荐
)

// === 推荐状态常量 ===
const (
	RecommendStatusShown     = 0 // 已展示
	RecommendStatusClicked   = 1 // 已点击
	RecommendStatusContacted = 2 // 已联系
	RecommendStatusDismissed = 3 // 已忽略
)

// === 审核规则类型常量 ===
const (
	AuditRuleTypeSensitiveWord = "sensitive_word" // 敏感词
	AuditRuleTypeProhibited    = "prohibited"     // 违禁内容
	AuditRuleTypeContact       = "contact"        // 联系方式
	AuditRuleTypePriceCheck    = "price_check"    // 价格校验
	AuditRuleTypeFrequency     = "frequency"       // 频率
)

// === 审核动作常量 ===
const (
	AuditActionReject  = "reject"  // 拒绝
	AuditActionApproval = "approval" // 通过
	AuditActionFilter  = "filter"  // 过滤
	AuditActionLimit   = "limit"   // 限制
)

// === 统计类型常量 ===
const (
	StatTypeDaily     = "daily"     // 日统计
	StatTypeBusiness  = "business"  // 商户统计
	StatTypeCategory  = "category"  // 分类统计
)

// === 电话拨打类型常量 ===
const (
	CallTypeClick = "click" // 点击拨号
	CallTypeCall  = "call"  // 直接拨打
)

// === 电话拨打状态常量 ===
const (
	CallStatusSuccess = "success" // 成功
	CallStatusFailed  = "failed"  // 失败
)

// === 设备类型常量 ===
const (
	VisitDevicePC      = "pc"      // PC
	VisitDeviceWAP     = "wap"      // 移动 web
	VisitDeviceAPP     = "app"      // APP
	VisitDeviceMiniAPP = "miniapp"  // 小程序
)

// === 来源常量 ===
const (
	VisitSourceSearch    = "search"    // 搜索
	VisitSourceCategory  = "category"  // 分类
	VisitSourceRecommend = "recommend" // 推荐
	VisitSourceDirect    = "direct"    // 直接访问
	VisitSourceShare     = "share"     // 分享
)

// ============================================================
// JSONB 字段类型（兼容 GORM 与 PostgreSQL jsonb）
// ============================================================

// JSONB 包装 []byte 以便与 PostgreSQL jsonb 类型交互
// 实现 driver.Valuer 与 sql.Scanner 接口，支持 GORM 自动映射
// 空值落库为 NULL，非空值落库为合法 JSON
type JSONB []byte

// Value 实现 driver.Valuer 接口
func (j JSONB) Value() (driver.Value, error) {
	if j == nil || len(j) == 0 {
		return nil, nil
	}
	return []byte(j), nil
}

// Scan 实现 sql.Scanner 接口
func (j *JSONB) Scan(value interface{}) error {
	if value == nil {
		*j = nil
		return nil
	}
	switch v := value.(type) {
	case []byte:
		*j = append((*j)[:0], v...)
		return nil
	case string:
		*j = []byte(v)
		return nil
	}
	return errors.New("dh114.JSONB.Scan: unsupported source type")
}

// MarshalJSON 实现 json.Marshaler
func (j JSONB) MarshalJSON() ([]byte, error) {
	if j == nil || len(j) == 0 {
		return []byte("null"), nil
	}
	return j, nil
}

// UnmarshalJSON 实现 json.Unmarshaler
func (j *JSONB) UnmarshalJSON(data []byte) error {
	if j == nil {
		return errors.New("dh114.JSONB.UnmarshalJSON: nil pointer")
	}
	*j = append((*j)[:0], data...)
	return nil
}

// Bytes 返回底层字节切片的只读副本
func (j JSONB) Bytes() []byte {
	if j == nil {
		return nil
	}
	out := make([]byte, len(j))
	copy(out, j)
	return out
}

// String 返回字符串形式
func (j JSONB) String() string {
	if j == nil || len(j) == 0 {
		return ""
	}
	return string(j)
}

// Parse 尝试反序列化为目标对象
func (j JSONB) Parse(v interface{}) error {
	if j == nil || len(j) == 0 {
		return nil
	}
	return json.Unmarshal(j, v)
}

// FromJSON 从任意 Go 对象构造 JSONB
func FromJSON(v interface{}) (JSONB, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	return JSONB(b), nil
}

// ============================================================
// 专用结构定义（用于 JSONB 字段反序列化）
// ============================================================

// BusinessImageItem 商户图片项（用于 dh114s.images 冗余字段）
type BusinessImageItem struct {
	URL       string `json:"url"`
	Thumbnail string `json:"thumbnail"`
	Type      string `json:"type"`
	Title     string `json:"title"`
	Sort      int    `json:"sort"`
}

// BusinessTagItem 商户标签项
type BusinessTagItem struct {
	Text  string `json:"text"`
	Color string `json:"color"`
}

// BusinessHourItem 营业时间项（用于 dh114s.business_hours 冗余字段）
type BusinessHourItem struct {
	Weekday  int    `json:"weekday"`
	OpenTime string `json:"open_time"`
	CloseTime string `json:"close_time"`
	IsOpen   bool   `json:"is_open"`
	Is24H    bool   `json:"is_24h"`
}

// ReviewImageItem 评价图片项
type ReviewImageItem struct {
	URL       string `json:"url"`
	Thumbnail string `json:"thumbnail"`
}

// AuditRuleThreshold 审核规则阈值
type AuditRuleThreshold struct {
	Min         float64 `json:"min"`
	Max         float64 `json:"max"`
	MaxCount    int     `json:"max_count"`
	WindowSec   int     `json:"window_sec"`
	Description string  `json:"description"`
}

// FacilityItem 设施项（用于 dh114_business.facilities JSONB）
type FacilityItem struct {
	Code string `json:"code"`
	Name string `json:"name"`
	Has  bool   `json:"has"`
}

// MenuTagItem 菜单标签项
type MenuTagItem struct {
	Text  string `json:"text"`
	Color string `json:"color"`
}

// Dh114 同城114黄页商户主表
type Dh114 struct {
	database.RegionBaseModel // 含 id/region_id/created_at/updated_at/deleted_at（地区隔离）

	// === 基础信息 ===
	Title       string     `gorm:"size:200;not null" json:"title"`         // 商户名
	Content     string     `gorm:"type:text" json:"content"`              // 简介
	CoverImage  string     `gorm:"size:255" json:"cover_image"`           // 封面图
	UserID      uint       `gorm:"index;not null" json:"user_id"`         // 发布者 ID
	UserName    string     `gorm:"size:50" json:"user_name"`              // 发布者昵称
	UserPhone   string     `gorm:"size:20" json:"user_phone"`             // 发布者手机
	UserAvatar  string     `gorm:"size:255" json:"user_avatar"`           // 发布者头像
	Status      int        `gorm:"default:0;index" json:"status"`         // 0草稿 1已发布 2下架 3过期 4删除
	AuditStatus int        `gorm:"default:0;index" json:"audit_status"`  // 0待审 1通过 2拒绝
	AuditReason string     `gorm:"size:500" json:"audit_reason"`          // 审核拒绝原因
	PublishedAt *time.Time `gorm:"index" json:"published_at"`             // 发布时间

	// === 分类关联 ===
	CategoryID   *uint  `gorm:"index" json:"category_id"`                               // 分类 ID
	CategoryName string `gorm:"size:64;not null;default:'';index" json:"category_name"` // 分类名（冗余）
	BusinessType string `gorm:"size:32;not null;default:'other';index" json:"business_type"` // restaurant/retail/service/entertain/hotel/medical/education/life/other

	// === 来源类型 ===
	SourceType string `gorm:"size:16;not null;default:'personal';index" json:"source_type"` // personal/merchant/chain

	// === 联系方式 ===
	Phone    string `gorm:"size:32;not null;default:'';index" json:"phone"`     // 联系电话
	AltPhone string `gorm:"size:32;not null;default:''" json:"alt_phone"`       // 备用电话
	Website  string `gorm:"size:255;not null;default:''" json:"website"`        // 官网
	Wechat   string `gorm:"size:64;not null;default:''" json:"wechat"`          // 微信号

	// === 地理位置 ===
	City             string  `gorm:"size:64;not null;default:'';index" json:"city"`             // 城市
	District         string  `gorm:"size:64;not null;default:'';index" json:"district"`         // 行政区
	BusinessDistrict string  `gorm:"size:128;not null;default:''" json:"business_district"`    // 商圈
	Address          string  `gorm:"size:500;not null;default:''" json:"address"`              // 详细地址
	Latitude         float64 `gorm:"type:decimal(10,7);default:0;index" json:"latitude"`      // 纬度
	Longitude        float64 `gorm:"type:decimal(10,7);default:0;index" json:"longitude"`       // 经度

	// === 评分统计 ===
	Rating      float64 `gorm:"type:decimal(3,2);default:0;index" json:"rating"`      // 综合评分 0-5
	ReviewCount int     `gorm:"not null;default:0;index" json:"review_count"`          // 评价数
	PriceAvg    float64 `gorm:"type:decimal(10,2);default:0" json:"price_avg"`         // 人均消费

	// === 互动统计 ===
	ViewCount   int        `gorm:"default:0" json:"view_count"`     // 浏览数
	FavCount    int        `gorm:"default:0" json:"fav_count"`       // 收藏数
	ContactCount int       `gorm:"default:0" json:"contact_count"`  // 联系数
	ShareCount  int        `gorm:"default:0" json:"share_count"`     // 分享数
	CallCount   int        `gorm:"default:0" json:"call_count"`      // 拨打数
	LastCallAt  *time.Time `gorm:"index" json:"last_call_at"`        // 最近拨打电话时间

	// === 风控 ===
	ContentHash string `gorm:"size:64;index" json:"content_hash"` // 图文指纹
	RiskScore   int    `gorm:"default:0;index" json:"risk_score"` // 风险评分 0-100

	// === 视频/VR ===
	VideoURL   string `gorm:"size:255" json:"video_url"`     // 视频 URL
	VideoCover string `gorm:"size:255" json:"video_cover"`   // 视频封面
	VRURL      string `gorm:"size:255" json:"vr_url"`         // VR 全景 URL

	// === JSONB 字段 ===
	Images        JSONB `gorm:"type:jsonb" json:"images"`         // 图片列表 JSON
	Tags          JSONB `gorm:"type:jsonb" json:"tags"`           // 标签数组
	BusinessHours JSONB `gorm:"type:jsonb" json:"business_hours"` // 营业时间 JSON
	Features      JSONB `gorm:"type:jsonb" json:"features"`       // 特色服务 JSON

	// === 运营字段 ===
	Featured       bool    `gorm:"default:false;index" json:"featured"`                   // 精选推荐
	Picked         bool    `gorm:"default:false;index" json:"picked"`                     // 运营甄选
	Verified       bool    `gorm:"default:false;index" json:"verified"`                  // 官方认证
	PromotionLevel int     `gorm:"default:0" json:"promotion_level"`                     // 推广等级 0-10
	TrafficWeight  float64 `gorm:"type:decimal(3,2);default:1.00" json:"traffic_weight"` // 流量权重

	// === 认证信息 ===
	VerifiedAt *time.Time `gorm:"index" json:"verified_at"` // 认证时间

	// Distance 仅在"附近"查询时由 SQL 计算并回填，非持久化字段（公里）
	Distance float64 `gorm:"-" json:"-"`
}

// TableName 表名（依据需求文档 7.2：主表 {module}s）
func (Dh114) TableName() string { return "dh114s" }
