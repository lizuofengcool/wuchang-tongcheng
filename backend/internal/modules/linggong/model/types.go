// Package model linggong 零工兼职模块通用类型定义
// 提供 JSONB 字段类型 + 兼职专用结构，兼容 GORM 与 PostgreSQL jsonb 类型
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
// nil 或长度为 0 时返回 nil（落库 NULL），否则原样返回字节切片
func (j JSONB) Value() (driver.Value, error) {
	if j == nil || len(j) == 0 {
		return nil, nil
	}
	return []byte(j), nil
}

// Scan 实现 sql.Scanner 接口
// 接受 []byte / string / nil 三种来源数据，统一转为 []byte
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
	return errors.New("linggong.JSONB.Scan: unsupported source type")
}

// MarshalJSON 实现 json.Marshaler，输出原始 JSON 字节
func (j JSONB) MarshalJSON() ([]byte, error) {
	if j == nil || len(j) == 0 {
		return []byte("null"), nil
	}
	return j, nil
}

// UnmarshalJSON 实现 json.Unmarshaler，原样保存字节
func (j *JSONB) UnmarshalJSON(data []byte) error {
	if j == nil {
		return errors.New("linggong.JSONB.UnmarshalJSON: nil pointer")
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

// String 返回字符串形式（用于日志/打印）
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

// FromJSON 从任意 Go 对象构造 JSONB（出错时返回 nil 与错误）
func FromJSON(v interface{}) (JSONB, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	return JSONB(b), nil
}

// ============================================================
// 兼职专用结构定义（用于 JSONB 字段反序列化）
// ============================================================

// WorkExperienceItem 工作经历项（用于 workers.work_experience JSONB）
type WorkExperienceItem struct {
	Company     string `json:"company"`
	Position    string `json:"position"`
	Description string `json:"description"`
	StartDate   string `json:"start_date"`
	EndDate     string `json:"end_date"`
	Location    string `json:"location"`
	Salary      float64 `json:"salary"`
	Reason      string `json:"reason"`
}

// EducationItem 教育背景项（用于 workers.education_history JSONB）
type EducationItem struct {
	School      string `json:"school"`
	Major       string `json:"major"`
	Degree      string `json:"degree"`
	StartDate   string `json:"start_date"`
	EndDate     string `json:"end_date"`
	Description string `json:"description"`
}

// PortfolioItem 作品集项（用于 workers.portfolio JSONB）
type PortfolioItem struct {
	Title       string `json:"title"`
	URL         string `json:"url"`
	Thumbnail   string `json:"thumbnail"`
	Type        string `json:"type"`
	Description string `json:"description"`
	CreatedAt   string `json:"created_at"`
}

// LinggongTag 标签项（用于 linggongs.tags / linggong_tasks.tags JSONB）
type LinggongTag struct {
	Text  string `json:"text"`
	Color string `json:"color"`
}

// LinggongFeature 岗位特征项（用于 linggongs.features JSONB）
type LinggongFeature struct {
	Category string `json:"category"` // work/welfare/requirement/equipment
	Code     string `json:"code"`
	Name     string `json:"name"`
	Value    string `json:"value"`
}

// ContractAttachment 合同附件项（用于 linggong_contracts.attachments JSONB）
type ContractAttachment struct {
	Type string `json:"type"` // id_card/contract_pdf/license/other
	Name string `json:"name"`
	URL  string `json:"url"`
	Size int64 `json:"size"`
}

// EvidenceImage 证据图项（用于 linggong_disputes.evidence_images JSONB）
type EvidenceImage struct {
	URL       string `json:"url"`
	Thumbnail string `json:"thumbnail"`
	Title     string `json:"title"`
}

// RatingImage 评价图片项（用于 linggong_ratings.images / append_images JSONB）
type RatingImage struct {
	URL       string `json:"url"`
	Thumbnail string `json:"thumbnail"`
}

// AuditRuleThreshold 审核规则阈值（用于 linggong_audit_rules.threshold JSONB）
type AuditRuleThreshold struct {
	Level       string  `json:"level"`        // high/medium/low
	MinSalary   float64 `json:"min_salary"`    // 最低薪资
	MaxSalary   float64 `json:"max_salary"`    // 最高薪资
	MaxCount    int     `json:"max_count"`     // 最大次数
	WindowSec   int     `json:"window_sec"`    // 时间窗口（秒）
	Ratio       float64 `json:"ratio"`         // 倍率
	Description string  `json:"description"`   // 描述
}

// TaskRequirement 任务要求项（用于 linggong_tasks.requirements JSONB）
type TaskRequirement struct {
	Name     string `json:"name"`
	Type     string `json:"type"`     // text/image/file/video
	Required bool   `json:"required"`
	Value    string `json:"value"`
	Description string `json:"description"`
}
