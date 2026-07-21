// Package model 多租户分站数据模型
// 依据架构设计 ershou-大模块架构方案.md 第 4.10 节：多租户分站中台（tenant）
// 职责：无限城市分站/独立配置/独立域名/独立运营权限/配置一键复制/数据隔离
package model

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"
)

// === 分站状态常量 ===
const (
	StationStatusDisabled = 0 // 已停用
	StationStatusEnabled  = 1 // 已启用
)

// === 员工角色常量 ===
const (
	StaffRoleOperator = "operator" // 运营员
	StaffRoleManager  = "manager"  // 管理员
)

// === 员工状态常量 ===
const (
	StaffStatusDisabled = 0 // 已停用
	StaffStatusEnabled  = 1 // 已启用
)

// === 域名 SSL 状态常量 ===
const (
	DomainSSLNone    = "none"    // 未配置
	DomainSSLPending = "pending" // 申请中
	DomainSSLActive  = "active"  // 已生效
	DomainSSLFailed  = "failed"  // 失败
)

// ============================================================
// JSONB 字段类型（兼容 GORM 与 PostgreSQL jsonb）
// 沿用 dh114 风格：实现 driver.Valuer / sql.Scanner / json.Marshaler / json.Unmarshaler
// 空值落库为 NULL，非空值落库为合法 JSON
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
	return errors.New("tenant.JSONB.Scan: unsupported source type")
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
		return errors.New("tenant.JSONB.UnmarshalJSON: nil pointer")
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
// Station 分站主表（tenant_stations）
// region_id 唯一：一个地区只能对应一个分站
// ============================================================

// Station 分站主表
type Station struct {
	ID          uint      `gorm:"primarykey" json:"id"`                                    // 主键
	RegionID    uint      `gorm:"uniqueIndex;not null" json:"region_id"`                  // 地区 ID（唯一，一个地区一个分站）
	Name        string    `gorm:"size:100;not null;default:''" json:"name"`               // 分站名称
	Domain      string    `gorm:"size:200;not null;default:'';index" json:"domain"`       // 主域名（冗余，便于快速查询）
	Logo        string    `gorm:"size:255;not null;default:''" json:"logo"`               // 分站 Logo
	Description string    `gorm:"type:text" json:"description"`                           // 分站描述
	Status      int       `gorm:"default:1;index" json:"status"`                          // 0已停用 1已启用
	Config      JSONB     `gorm:"type:jsonb" json:"config"`                               // 独立运营配置（JSONB）
	CreatedAt   time.Time `json:"created_at"`                                             // 创建时间
	UpdatedAt   time.Time `json:"updated_at"`                                             // 更新时间
}

// TableName 表名
func (Station) TableName() string { return "tenant_stations" }
