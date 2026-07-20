// Package model 同城拼车出行数据模型
// 依据 v3.2.1 架构方案：对标哈啰出行/嘀嗒出行/滴滴顺风车
// 依据需求文档 1.10：4 维数据隔离（region_id 地区隔离 + user_id 用户隔离）
// 依据需求文档 7.1：通用字段 id/region_id/created_at/updated_at/deleted_at + status + audit_status
// 依据需求文档 7.2：主表 pinches（保持兼容已发布数据）
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
	PincheStatusDraft     = 0 // 草稿
	PincheStatusPublished = 1 // 已发布
	PincheStatusFinished  = 2 // 已结束
	PincheStatusCancelled = 3 // 已取消
	PincheStatusOngoing   = 4 // 进行中
)

// === 审核状态常量 ===
const (
	PincheAuditPending  = 0 // 待审
	PincheAuditApproved = 1 // 通过
	PincheAuditRejected = 2 // 拒绝
)

// === 行程类型常量 ===
const (
	TripTypeShunfeng = "shunfeng" // 顺风车
	TripTypePinche   = "pinche"   // 拼车
	TripTypeBaoche   = "baoche"   // 包车
)

// === 发布者角色常量 ===
const (
	RoleDriver    = "driver"    // 车主
	RolePassenger = "passenger" // 乘客
)

// === 支付方式常量 ===
const (
	PaymentMethodCash    = "cash"    // 现金
	PaymentMethodWechat  = "wechat"  // 微信
	PaymentMethodAlipay  = "alipay"  // 支付宝
	PaymentMethodBalance = "balance" // 余额
	PaymentMethodETC     = "etc"     // ETC
)

// === 预订状态常量 ===
const (
	BookingStatusPending   = 0 // 待支付
	BookingStatusPaid      = 1 // 已支付
	BookingStatusBoarded   = 2 // 已上车
	BookingStatusCompleted = 3 // 已完成
	BookingStatusCancelled = 4 // 已取消
	BookingStatusRefunded  = 5 // 已退款
)

// === 车主认证状态常量 ===
const (
	DriverStatusPending  = 0 // 待审
	DriverStatusApproved = 1 // 通过
	DriverStatusRejected = 2 // 拒绝
	DriverStatusExpired  = 3 // 已过期
)

// === 车辆状态常量 ===
const (
	VehicleStatusPending  = 0 // 待审
	VehicleStatusApproved = 1 // 通过
	VehicleStatusRejected = 2 // 拒绝
)

// === 保险状态常量 ===
const (
	InsuranceStatusPending  = 0 // 待生效
	InsuranceStatusActive   = 1 // 生效中
	InsuranceStatusExpired  = 2 // 已结束
	InsuranceStatusClaimed  = 3 // 已理赔
)

// === 保险类型常量 ===
const (
	InsuranceTypePassenger = "passenger" // 乘客险
	InsuranceTypeDriver    = "driver"    // 司机险
	InsuranceTypeBoth      = "both"      // 双重
)

// === 支付状态常量 ===
const (
	PaymentStatusPending = 0 // 待支付
	PaymentStatusPaid    = 1 // 已支付
	PaymentStatusRefund  = 2 // 已退款
	PaymentStatusFailed  = 3 // 已失败
)

// === 评价类型常量 ===
const (
	RatingTypePassengerToDriver = "passenger_to_driver" // 乘客评价车主
	RatingTypeDriverToPassenger = "driver_to_passenger" // 车主评价乘客
)

// === 评价状态常量 ===
const (
	RatingStatusPending = 0 // 待审
	RatingStatusApproved = 1 // 通过
	RatingStatusRejected = 2 // 拒绝
	RatingStatusHidden   = 3 // 隐藏
)

// === 报警类型常量 ===
const (
	AlertTypeSOS       = "sos"       // 一键报警
	AlertTypeShare     = "share"     // 行程分享
	AlertTypePeriodic  = "periodic"  // 定时上报
)

// === 报警状态常量 ===
const (
	AlertStatusPending  = 0 // 未处理
	AlertStatusHandling = 1 // 处理中
	AlertStatusHandled  = 2 // 已处理
)

// === 行程状态常量 ===
const (
	TripStatusOngoing   = 0 // 进行中
	TripStatusCompleted = 1 // 已完成
	TripStatusAbnormal  = 2 // 异常结束
)

// === 消息类型常量 ===
const (
	MessageTypeText     = "text"     // 文本
	MessageTypeImage    = "image"    // 图片
	MessageTypeVoice    = "voice"    // 语音
	MessageTypeSystem   = "system"   // 系统
	MessageTypeLocation = "location" // 位置
)

// === 取消角色/类型常量 ===
const (
	CancelRoleDriver    = "driver"
	CancelRolePassenger = "passenger"
	CancelTypeUser      = "user"
	CancelTypeSystem    = "system"
	CancelTypeTimeout   = "timeout"
)

// === 退款状态常量 ===
const (
	RefundStatusPending = 0 // 待退
	RefundStatusDone    = 1 // 已退
	RefundStatusFailed  = 2 // 失败
)

// === 投诉状态常量 ===
const (
	ComplaintStatusPending  = 0 // 待处理
	ComplaintStatusHandling = 1 // 处理中
	ComplaintStatusHandled  = 2 // 已处理
	ComplaintStatusRejected = 3 // 已驳回
)

// === 审核规则动作常量 ===
const (
	AuditRuleActionPass          = "pass"
	AuditRuleActionReject        = "reject"
	AuditRuleActionManualReview  = "manual_review"
)

// === 统计类型常量 ===
const (
	StatTypeDaily   = "daily"
	StatTypeWeekly  = "weekly"
	StatTypeMonthly = "monthly"
	StatTypeTotal   = "total"
)

// ============================================================
// JSONB 类型定义（兼容 PostgreSQL jsonb 与 GORM）
// ============================================================

// JSONB 包装 []byte 以便与 PostgreSQL jsonb 类型交互
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
	return errors.New("pinche.JSONB.Scan: unsupported source type")
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
		return errors.New("pinche.JSONB.UnmarshalJSON: nil pointer")
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
// 主表 Pinche
// ============================================================

// Pinche 拼车行程主表
type Pinche struct {
	database.RegionBaseModel // id/region_id/created_at/updated_at/deleted_at

	// 发布者信息
	UserID     uint   `gorm:"index;not null" json:"user_id"`
	UserName   string `gorm:"size:50" json:"user_name"`
	UserPhone  string `gorm:"size:20" json:"user_phone"`
	UserAvatar string `gorm:"size:255" json:"user_avatar"`

	// 行程类型与角色
	TripType string `gorm:"size:16;not null;default:'shunfeng';index" json:"trip_type"` // shunfeng/pinche/baoche
	Role     string `gorm:"size:16;not null;default:'driver';index" json:"role"`         // driver/passenger

	// 基础信息
	Title      string `gorm:"size:200" json:"title"`
	Content    string `gorm:"type:text" json:"content"`
	CoverImage string `gorm:"size:255" json:"cover_image"`

	// 状态
	Status      int        `gorm:"default:0;index" json:"status"`         // 0草稿 1已发布 2已结束 3已取消 4进行中
	AuditStatus int        `gorm:"default:0;index" json:"audit_status"`   // 0待审 1通过 2拒绝
	AuditReason string     `gorm:"size:500" json:"audit_reason"`
	PublishedAt *time.Time `gorm:"index" json:"published_at"`

	// 出发信息
	DepartureTime   *time.Time `gorm:"index" json:"departure_time"`
	PickupLocation  string     `gorm:"size:255" json:"pickup_location"`
	PickupLat       float64    `gorm:"type:decimal(10,7);default:0" json:"pickup_lat"`
	PickupLng       float64    `gorm:"type:decimal(10,7);default:0" json:"pickup_lng"`
	DropoffLocation string     `gorm:"size:255" json:"dropoff_location"`
	DropoffLat      float64    `gorm:"type:decimal(10,7);default:0" json:"dropoff_lat"`
	DropoffLng      float64    `gorm:"type:decimal(10,7);default:0" json:"dropoff_lng"`

	// 行程距离/时长
	DistanceKm  float64 `gorm:"type:decimal(10,2);default:0" json:"distance_km"`
	DurationMin int     `gorm:"not null;default:0" json:"duration_min"`

	// 座位
	TotalSeats     int `gorm:"not null;default:4" json:"total_seats"`
	AvailableSeats int `gorm:"not null;default:4" json:"available_seats"`
	BookedSeats    int `gorm:"not null;default:0" json:"booked_seats"`

	// 金额
	PricePerSeat float64 `gorm:"type:decimal(12,2);default:0;index" json:"price_per_seat"`
	TotalAmount  float64 `gorm:"type:decimal(12,2);default:0" json:"total_amount"`
	TollFee      float64 `gorm:"type:decimal(12,2);default:0" json:"toll_fee"`

	// 关联
	VehicleID         *uint `gorm:"index" json:"vehicle_id"`
	DriverID          *uint `gorm:"index" json:"driver_id"`
	RouteID           *uint `gorm:"index" json:"route_id"`
	InsuranceID       *uint `gorm:"index" json:"insurance_id"`
	TripID            *uint `gorm:"index" json:"trip_id"`
	EmergencyContactID *uint `gorm:"index" json:"emergency_contact_id"`

	// 行程分享与支付
	ShareToken    string `gorm:"size:64" json:"share_token"`
	PaymentMethod string `gorm:"size:16;not null;default:'cash'" json:"payment_method"`

	// 配置/特征（JSONB）
	Features JSONB `gorm:"type:jsonb" json:"features"`
	Tags     JSONB `gorm:"type:jsonb" json:"tags"`

	// 互动统计
	ViewCount    int `gorm:"default:0" json:"view_count"`
	FavCount     int `gorm:"default:0" json:"fav_count"`
	ContactCount int `gorm:"default:0" json:"contact_count"`
	ShareCount   int `gorm:"default:0" json:"share_count"`

	// 运营字段
	Featured       bool `gorm:"default:false;index" json:"featured"`
	Picked         bool `gorm:"default:false;index" json:"picked"`
	Verified       bool `gorm:"default:false;index" json:"verified"`
	PromotionLevel int  `gorm:"default:0" json:"promotion_level"`

	// 风控
	ContentHash string `gorm:"size:64;index" json:"content_hash"`
	RiskScore   int    `gorm:"default:0;index" json:"risk_score"`

	// 时间节点
	StartedAt   *time.Time `gorm:"index" json:"started_at"`
	CompletedAt *time.Time `gorm:"index" json:"completed_at"`
	CancelledAt *time.Time `gorm:"index" json:"cancelled_at"`

	// Distance 仅在"附近"查询时回填
	Distance float64 `gorm:"-" json:"-"`
}

// TableName 表名
func (Pinche) TableName() string { return "pinches" }
