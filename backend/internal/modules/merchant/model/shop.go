// Package model 商户中台数据模型
// 依据架构设计 4.4：商家商户中台（merchant）
// 职责：入驻/认领/店铺管理/CRM/商家权限/商家结算/商家营销工具
// 说明：所有"经纪人/企业/师傅/商家"角色统一复用此中台
//       （房产经纪人、招聘企业、二手商家、到家师傅都是 merchant）
package model

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"

	"wuchang-tongcheng/internal/pkg/database"
)

// === 店铺状态常量 ===
const (
	ShopStatusPending = 0 // 审核中
	ShopStatusActive  = 1 // 正常
	ShopStatusStopped = 2 // 停用
)

// === 员工角色常量 ===
const (
	StaffRoleOwner   = "owner"   // 店主
	StaffRoleManager = "manager" // 管理员
	StaffRoleClerk   = "clerk"   // 店员
)

// === 员工状态常量 ===
const (
	StaffStatusActive  = 1 // 在职
	StaffStatusStopped = 2 // 停用
)

// === 结算状态常量 ===
const (
	SettleStatusPending  = 0 // 待结算
	SettleStatusSettled  = 1 // 已结算
	SettleStatusPaid     = 2 // 已提现
	SettleStatusCanceled = 3 // 已撤销
)

// === 类目状态常量 ===
const (
	CategoryStatusDisabled = 0 // 禁用
	CategoryStatusEnabled  = 1 // 启用
)

// === 认证类型常量 ===
const (
	VerifyTypeBusiness  = "business"  // 企业认证
	VerifyTypePersonal  = "personal"  // 个人认证
)

// === 认证状态常量 ===
const (
	VerifyStatusPending  = 0 // 待审
	VerifyStatusApproved = 1 // 通过
	VerifyStatusRejected = 2 // 拒绝
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
	return errors.New("merchant.JSONB.Scan: unsupported source type")
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
		return errors.New("merchant.JSONB.UnmarshalJSON: nil pointer")
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

// StaffPermissionItem 员工权限项
type StaffPermissionItem struct {
	Code  string `json:"code"`  // 权限编码
	Name  string `json:"name"`  // 权限名称
	Scope string `json:"scope"` // 权限范围 all/read/write
}

// Shop 商户店铺主表（merchant_shops）
// 依据架构设计 4.4 merchant_shops
type Shop struct {
	database.RegionBaseModel // 含 id/region_id/created_at/updated_at/deleted_at（地区隔离）

	// === 关联 ===
	OwnerID    uint   `gorm:"index;not null" json:"owner_id"`          // 店主用户 ID
	Name       string `gorm:"size:100;not null;default:'';index" json:"name"` // 商户名
	Logo       string `gorm:"size:500;not null;default:''" json:"logo"`        // 商户 LOGO
	Intro      string `gorm:"type:text" json:"intro"`                       // 商户简介
	CategoryID *uint  `gorm:"index" json:"category_id"`                    // 主营类目 ID

	// === 状态 ===
	Status      int        `gorm:"default:0;index" json:"status"`        // 0审核中 1正常 2停用
	CreditScore int        `gorm:"not null;default:100;index" json:"credit_score"` // 信用分
	Level       int        `gorm:"not null;default:1;index" json:"level"` // 商户等级
	SettledAt   *time.Time `gorm:"index" json:"settled_at"`               // 入驻时间
}

// TableName 表名
func (Shop) TableName() string { return "merchant_shops" }
