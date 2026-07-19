// Package model house 房屋租售模块通用类型定义
// 提供 JSONB 字段类型 + 房屋专用结构，兼容 GORM 与 PostgreSQL jsonb 类型
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
	return errors.New("house.JSONB.Scan: unsupported source type")
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
		return errors.New("house.JSONB.UnmarshalJSON: nil pointer")
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
// 房屋专用结构定义（用于 JSONB 字段反序列化）
// ============================================================

// HouseFacilityItem 配套设施项（用于 houses.facilities JSONB）
// 存储房源已有的配套设施 ID 与名称
type HouseFacilityItem struct {
	ID   uint   `json:"id"`   // 设施 ID（house_facilities.id）
	Name string `json:"name"` // 设施名称
	Code string `json:"code"` // 设施编码
	Icon string `json:"icon"` // 图标
}

// HouseTagItem 标签项（用于 houses.tags JSONB，最多 5 个）
type HouseTagItem struct {
	Text  string `json:"text"`  // 标签文本
	Color string `json:"color"` // 标签颜色
}

// HouseNearbyPOI 附近 POI 项（用于 houses.nearby_pois JSONB）
// 对标贝壳：地铁/学校/医院/商超等周边配套
type HouseNearbyPOI struct {
	Type      string  `json:"type"`      // POI 类型：subway/school/hospital/mall/park/bank
	Name      string  `json:"name"`      // 名称
	Distance  int     `json:"distance"`  // 距离（米）
	Latitude  float64 `json:"latitude"`  // 纬度
	Longitude float64 `json:"longitude"` // 经度
	Address   string  `json:"address"`   // 地址
}

// VRScene VR 场景项（用于 house_vr_tours.scenes JSONB）
// 对标贝壳：720°全景/虚拟看房，多场景串联
type VRScene struct {
	ID       string `json:"id"`       // 场景 ID
	Name     string `json:"name"`     // 场景名称（如：客厅/主卧/厨房）
	ImageURL string `json:"imageURL"` // 全景图 URL
	ThumbURL string `json:"thumbURL"` // 缩略图 URL
	Hotspots []VRHotspot `json:"hotspots"` // 热点（跳转到其他场景）
}

// VRHotspot VR 场景热点
type VRHotspot struct {
	X         float64 `json:"x"`         // X 坐标 0-1
	Y         float64 `json:"y"`         // Y 坐标 0-1
	TargetID  string  `json:"targetID"`  // 跳转目标场景 ID
	Action    string  `json:"action"`    // 热点类型：jump/info/link
	Title     string  `json:"title"`     // 热点标题
	Content   string  `json:"content"`   // 热点内容
}

// CommunityImage 小区图片项（用于 house_communities.images JSONB）
type CommunityImage struct {
	URL       string `json:"url"`
	Thumbnail string `json:"thumbnail"`
	Sort      int    `json:"sort"`
	Title     string `json:"title"`
}

// ContractAttachment 合同附件项（用于 house_contracts.attachments JSONB）
type ContractAttachment struct {
	Type string `json:"type"` // 类型：id_card/property_cert/contract_pdf/other
	Name string `json:"name"`
	URL  string `json:"url"`
	Size int64 `json:"size"`
}

// EvidenceImage 举报证据图项（用于 house_reports.evidence_images JSONB）
type EvidenceImage struct {
	URL       string `json:"url"`
	Thumbnail string `json:"thumbnail"`
	Title     string `json:"title"`
}

// ReviewImage 评价图片项（用于 house_reviews.images / append_images JSONB）
type ReviewImage struct {
	URL       string `json:"url"`
	Thumbnail string `json:"thumbnail"`
}

// AuditRuleThreshold 审核规则阈值（用于 house_audit_rules.threshold JSONB）
type AuditRuleThreshold struct {
	Level       string  `json:"level"`        // high/medium/low
	MinPrice    float64 `json:"min_price"`    // 最低价
	MaxPrice    float64 `json:"max_price"`    // 最高价
	MaxCount    int     `json:"max_count"`    // 最大次数
	WindowSec   int     `json:"window_sec"`   // 时间窗口（秒）
	Ratio       float64 `json:"ratio"`        // 倍率
	Description string  `json:"description"`  // 描述
}

// AgentGoodAt 经纪人擅长领域（用于 house_agents.good_at JSONB）
type AgentGoodAt struct {
	Category string `json:"category"` // 类型：rent/sale/new_house/commercial
	Text     string `json:"text"`     // 描述
}

// AgentServiceArea 经纪人服务区域（用于 house_agents.service_area JSONB）
type AgentServiceArea struct {
	City      string `json:"city"`
	District  string `json:"district"`
	Business  string `json:"business"` // 商圈
}
