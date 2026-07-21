// Package model 分销合伙人中台 - JSONB 字段类型
// 兼容 GORM 与 PostgreSQL jsonb，参考 dh114/model/dh114.go
package model

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
)

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
	return errors.New("distribution.JSONB.Scan: unsupported source type")
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
		return errors.New("distribution.JSONB.UnmarshalJSON: nil pointer")
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

// LevelExtraBenefit 等级额外权益项（用于 distribution_levels.extra_benefits JSONB）
type LevelExtraBenefit struct {
	Code string `json:"code"` // 权益代码
	Name string `json:"name"` // 权益名称
	Desc string `json:"desc"` // 权益描述
}

// BankInfoItem 提现银行信息（用于 distribution_withdrawals.bank_info JSONB）
type BankInfoItem struct {
	BankName    string `json:"bank_name"`     // 银行名称
	AccountName string `json:"account_name"`  // 账户姓名
	AccountNo   string `json:"account_no"`    // 账号
	Branch      string `json:"branch"`        // 支行
	Type        string `json:"type"`          // 类型：bank/alipay/wechat
}
