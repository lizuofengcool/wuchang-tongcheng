// Package model 同城车辆买卖数据模型
// 依据 v3.2.1 架构方案第六章：对标瓜子/人人车/懂车帝/毛豆新车/易鑫车贷
// 依据需求文档 1.10：4 维数据隔离（region_id 地区隔离 + user_id 用户隔离）
// 依据需求文档 7.1：通用字段 id/region_id/created_at/updated_at/deleted_at + status + audit_status
// 依据需求文档 7.2：主表 cars（保持兼容已发布数据）
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
	StatusSold      = 5 // 已售出
)

// === 审核状态常量（依据需求文档 7.1） ===
const (
	AuditPending  = 0 // 待审
	AuditApproved = 1 // 通过
	AuditRejected = 2 // 拒绝
)

// === 发布类型常量 ===
const (
	ListingTypeNew    = "new"    // 新车
	ListingTypeUsed   = "used"   // 二手
	ListingTypeReplace = "replace" // 置换
	ListingTypeRental = "rental"  // 租车
)

// === 来源类型常量 ===
const (
	SourceTypePersonal    = "personal"    // 个人
	SourceTypeDealer      = "dealer"      // 车商
	SourceTypeManufacturer = "manufacturer" // 厂商
)

// === 车型分类常量 ===
const (
	CarTypeSedan       = "sedan"       // 轿车
	CarTypeSUV         = "suv"         // SUV
	CarTypeMPV         = "mpv"         // MPV
	CarTypeNewEnergy   = "new_energy"  // 新能源
	CarTypeSports      = "sports"      // 跑车
	CarTypeTruck       = "truck"       // 皮卡
	CarTypeVan         = "van"         // 面包车
	CarTypeBus         = "bus"         // 客车
)

// === 变速箱常量 ===
const (
	TransmissionManual = "manual" // 手动
	TransmissionAuto   = "auto"   // 自动
	TransmissionCVT    = "cvt"    // CVT
	TransmissionDCT    = "dct"    // 双离合
	TransmissionAMT    = "amt"    // 手自一体
)

// === 燃油类型常量 ===
const (
	FuelTypeGasoline       = "gasoline"       // 汽油
	FuelTypeDiesel         = "diesel"         // 柴油
	FuelTypeHybrid         = "hybrid"         // 混动
	FuelTypePureElectric   = "pure_electric"  // 纯电
	FuelTypeRangeExtender  = "range_extender" // 增程
	FuelTypeHydrogen       = "hydrogen"       // 氢能源
)

// === 排放标准常量 ===
const (
	EmissionStandardChina1 = "china_1"
	EmissionStandardChina2 = "china_2"
	EmissionStandardChina3 = "china_3"
	EmissionStandardChina4 = "china_4"
	EmissionStandardChina5 = "china_5"
	EmissionStandardChina6 = "china_6"
)

// === 车况等级常量 ===
const (
	ConditionLevelA      = "A" // 极佳
	ConditionLevelB      = "B" // 良好
	ConditionLevelC      = "C" // 一般
	ConditionLevelD      = "D" // 较差
)

// === 年检状态常量 ===
const (
	AnnualInspectionValid   = "valid"   // 有效
	AnnualInspectionExpiring = "expiring" // 即将到期
	AnnualInspectionExpired = "expired" // 已过期
	AnnualInspectionNone    = "none"    // 无需年检
)

// === 保险状态常量 ===
const (
	InsuranceValid    = "valid"    // 有效
	InsuranceExpiring = "expiring" // 即将到期
	InsuranceExpired  = "expired"  // 已过期
	InsuranceNone     = "none"     // 无保险
)

// === 使用性质常量 ===
const (
	UseTypeNonOperational = "non_operational" // 非营运
	UseTypeOperational    = "operational"     // 营运
)

// === 里程单位常量 ===
const (
	MileageUnitKM  = "km"  // 公里
	MileageUnitMile = "mile" // 英里
)

// Car 车辆买卖主表（保持 cars 表名兼容已发布数据）
type Car struct {
	database.RegionBaseModel // 含 id/region_id/created_at/updated_at/deleted_at（地区隔离）

	// === 基础信息 ===
	Title       string     `gorm:"size:200;not null" json:"title"`         // 标题
	Content     string     `gorm:"type:text" json:"content"`              // 详情描述
	CoverImage  string     `gorm:"size:255" json:"cover_image"`           // 封面图
	UserID      uint       `gorm:"index;not null" json:"user_id"`         // 发布者 ID
	UserName    string     `gorm:"size:50" json:"user_name"`              // 发布者昵称
	UserPhone   string     `gorm:"size:20" json:"user_phone"`             // 发布者手机
	UserAvatar  string     `gorm:"size:255" json:"user_avatar"`           // 发布者头像
	Status      int        `gorm:"default:0;index" json:"status"`         // 0草稿 1已发布 2下架 3过期 4删除 5已售
	AuditStatus int        `gorm:"default:0;index" json:"audit_status"`   // 0待审 1通过 2拒绝
	AuditReason string     `gorm:"size:500" json:"audit_reason"`          // 审核拒绝原因
	PublishedAt *time.Time `gorm:"index" json:"published_at"`             // 发布时间

	// === 发布类型/来源类型 ===
	ListingType string `gorm:"size:16;not null;default:'used';index" json:"listing_type"`  // new/used/replace/rental
	SourceType  string `gorm:"size:16;not null;default:'personal';index" json:"source_type"` // personal/dealer/manufacturer
	CarType     string `gorm:"size:32;not null;default:'sedan';index" json:"car_type"`     // sedan/suv/mpv/new_energy/sports/truck

	// === 品牌型号关联 ===
	BrandID    *uint  `gorm:"index" json:"brand_id"`                                              // 品牌 ID（car_models.id）
	BrandName  string `gorm:"size:64;not null;default:'';index" json:"brand_name"`              // 品牌名（冗余）
	ModelID    *uint  `gorm:"index" json:"model_id"`                                              // 型号 ID（car_models.id）
	ModelName  string `gorm:"size:128;not null;default:''" json:"model_name"`                    // 型号名（冗余）
	Series     string `gorm:"size:64;not null;default:''" json:"series"`                         // 车系
	CategoryID *uint  `gorm:"index" json:"category_id"`                                           // 分类 ID（car_categories.id）

	// === 价格 ===
	Price          float64 `gorm:"type:decimal(14,2);default:0;index" json:"price"`           // 售价（元）
	OriginalPrice  float64 `gorm:"type:decimal(14,2);default:0" json:"original_price"`        // 原价/指导价
	AveragePrice   float64 `gorm:"type:decimal(10,2);default:0;index" json:"average_price"`   // 同款均价
	PriceNegotiable bool   `gorm:"not null;default:false" json:"price_negotiable"`             // 价格可议
	DealerPrice    float64 `gorm:"type:decimal(14,2);default:0" json:"dealer_price"`           // 车商报价

	// === 上牌时间/里程 ===
	RegistrationYear        int        `gorm:"not null;default:0;index" json:"registration_year"`            // 上牌年
	RegistrationMonth       int        `gorm:"not null;default:0" json:"registration_month"`                  // 上牌月
	FirstRegistrationDate   *time.Time `gorm:"type:date" json:"first_registration_date"`                      // 首次上牌日期
	Mileage                 float64    `gorm:"type:decimal(10,1);default:0;index" json:"mileage"`             // 里程（公里）
	MileageUnit             string     `gorm:"size:16;not null;default:'km'" json:"mileage_unit"`             // 里程单位 km/mile

	// === 排量/变速/燃油 ===
	Displacement       float64 `gorm:"type:decimal(4,2);default:0;index" json:"displacement"`           // 排量（L）
	Transmission       string  `gorm:"size:32;not null;default:'';index" json:"transmission"`           // manual/auto/cvt/dct/amt
	FuelType           string  `gorm:"size:32;not null;default:'gasoline';index" json:"fuel_type"`      // gasoline/diesel/hybrid/pure_electric/range_extender
	EmissionStandard   string  `gorm:"size:32;not null;default:'';index" json:"emission_standard"`      // china_1~china_6
	EngineType         string  `gorm:"size:64;not null;default:''" json:"engine_type"`                  // 发动机型号
	Horsepower         int     `gorm:"not null;default:0" json:"horsepower"`                            // 马力（PS）

	// === 颜色/座位/车门 ===
	ExteriorColor string `gorm:"size:32;not null;default:''" json:"exterior_color"` // 外观颜色
	InteriorColor string `gorm:"size:32;not null;default:''" json:"interior_color"` // 内饰颜色
	SeatCount     int    `gorm:"not null;default:5" json:"seat_count"`              // 座位数
	DoorCount     int    `gorm:"not null;default:4" json:"door_count"`              // 车门数

	// === 车况 ===
	ConditionLevel  string `gorm:"size:16;not null;default:'A';index" json:"condition_level"` // A/B/C/D
	ConditionScore  int    `gorm:"not null;default:0" json:"condition_score"`                 // 车况评分 0-100
	AccidentCount   int    `gorm:"not null;default:0" json:"accident_count"`                  // 事故次数

	// === 过户/年检/保险 ===
	TransferCount             int        `gorm:"not null;default:0;index" json:"transfer_count"`              // 过户次数
	LastTransferDate          *time.Time `gorm:"type:date" json:"last_transfer_date"`                          // 最近过户日期
	AnnualInspectionDue       *time.Time `gorm:"type:date;index" json:"annual_inspection_due"`                 // 年检到期
	AnnualInspectionStatus    string     `gorm:"size:16;not null;default:'valid'" json:"annual_inspection_status"` // valid/expiring/expired/none
	InsuranceDue              *time.Time `gorm:"type:date;index" json:"insurance_due"`                        // 交强险到期
	InsuranceStatus           string     `gorm:"size:16;not null;default:'valid'" json:"insurance_status"`     // valid/expiring/expired/none
	CommercialInsuranceDue    *time.Time `gorm:"type:date" json:"commercial_insurance_due"`                    // 商业险到期

	// === 车架号/车牌 ===
	VIN             string `gorm:"size:32;not null;default:'';index" json:"vin"`             // 车架号（17位）
	LicensePlate    string `gorm:"size:32;not null;default:'';index" json:"license_plate"`   // 车牌号
	LicenseLocation string `gorm:"size:64;not null;default:''" json:"license_location"`      // 上牌地
	EngineNo        string `gorm:"size:64;not null;default:''" json:"engine_no"`             // 发动机号

	// === 使用性质 ===
	UseType                string  `gorm:"size:32;not null;default:'non_operational'" json:"use_type"` // non_operational/operational
	MaintenanceCount       int     `gorm:"not null;default:0" json:"maintenance_count"`                 // 保养次数
	LastMaintenanceMileage float64 `gorm:"type:decimal(10,1);default:0" json:"last_maintenance_mileage"` // 最近保养里程

	// === 地理位置 ===
	City             string  `gorm:"size:64;not null;default:'';index" json:"city"`             // 城市
	District         string  `gorm:"size:64;not null;default:'';index" json:"district"`         // 行政区
	BusinessDistrict string  `gorm:"size:128;not null;default:''" json:"business_district"`     // 商圈
	Address          string  `gorm:"size:500;not null;default:''" json:"address"`              // 详细地址
	Latitude         float64 `gorm:"type:decimal(10,7);default:0" json:"latitude"`             // 纬度
	Longitude        float64 `gorm:"type:decimal(10,7);default:0" json:"longitude"`            // 经度

	// === 互动统计 ===
	ViewCount      int        `gorm:"default:0" json:"view_count"`      // 浏览数
	FavCount       int        `gorm:"default:0" json:"fav_count"`        // 收藏数
	ContactCount   int        `gorm:"default:0" json:"contact_count"`    // 联系数
	ShareCount     int        `gorm:"default:0" json:"share_count"`      // 分享数
	TestDriveCount int        `gorm:"default:0" json:"test_drive_count"` // 试驾预约数
	LastTestDriveAt *time.Time `gorm:"index" json:"last_test_drive_at"`  // 最近试驾时间

	// === 风控 ===
	ContentHash string `gorm:"size:64;index" json:"content_hash"`   // 图文指纹
	RiskScore   int    `gorm:"default:0;index" json:"risk_score"`   // 风险评分 0-100
	SameCarID   string `gorm:"size:64;index" json:"same_car_id"`    // 同车识别 ID

	// === 视频/360° ===
	VideoURL       string `gorm:"size:255" json:"video_url"`        // 视频 URL
	VideoCover     string `gorm:"size:255" json:"video_cover"`      // 视频封面
	Panorama360URL string `gorm:"size:255" json:"panorama_360_url"` // 360° 全景图 URL

	// === 配置/特征（JSONB）===
	Features         JSONB `gorm:"type:jsonb" json:"features"`          // 配置特征 JSON
	Tags             JSONB `gorm:"type:jsonb" json:"tags"`              // 标签数组（最多 5 个）
	InspectionItems  JSONB `gorm:"type:jsonb" json:"inspection_items"`  // 254项检测报告 JSON
	AccidentHistory  JSONB `gorm:"type:jsonb" json:"accident_history"`  // 事故历史 JSON

	// === 运营字段 ===
	Featured       bool    `gorm:"default:false;index" json:"featured"`                       // 精选推荐
	Picked         bool    `gorm:"default:false;index" json:"picked"`                         // 运营甄选
	Verified       bool    `gorm:"default:false;index" json:"verified"`                       // 官方验真
	PromotionLevel int     `gorm:"default:0" json:"promotion_level"`                          // 推广等级 0-10
	TrafficWeight  float64 `gorm:"type:decimal(3,2);default:1.00" json:"traffic_weight"`      // 流量权重 0.00-9.99

	// === 真车认证 ===
	RealCarVerified   bool       `gorm:"default:false;index" json:"real_car_verified"`     // 真车认证
	RealCarVerifiedAt *time.Time `gorm:"index" json:"real_car_verified_at"`                // 真车认证时间

	// Distance 仅在"附近"查询时由 SQL 计算并回填，非持久化字段（公里）
	Distance float64 `gorm:"-" json:"-"`
}

// TableName 表名（依据需求文档 7.2：主表 {module}s）
// 注意：保持现有 cars 表名不变以兼容已发布数据
func (Car) TableName() string { return "cars" }
