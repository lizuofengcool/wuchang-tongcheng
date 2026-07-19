// Package model car 车辆买卖模块通用类型定义
// 提供 JSONB 字段类型 + 车辆专用结构，兼容 GORM 与 PostgreSQL jsonb 类型
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
	return errors.New("car.JSONB.Scan: unsupported source type")
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
		return errors.New("car.JSONB.UnmarshalJSON: nil pointer")
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
// 车辆专用结构定义（用于 JSONB 字段反序列化）
// ============================================================

// CarInspectionItem 254项检测项（用于 cars.inspection_items / car_inspections.items JSONB）
// 对标瓜子：分类/项名/结果/扣分/图片
type CarInspectionItem struct {
	Category    string `json:"category"`    // 分类：exterior/interior/engine/chassis/electronics/safety
	Name        string `json:"name"`        // 检测项名称
	Result      string `json:"result"`      // 结果：pass/fail/warning
	Score       int    `json:"score"`       // 评分 0-100
	DeductScore int    `json:"deductScore"` // 扣分
	Description string `json:"description"` // 描述
	Images      []CarInspectionImage `json:"images"` // 检测图片
}

// CarInspectionImage 检测图片
type CarInspectionImage struct {
	URL       string `json:"url"`
	Thumbnail string `json:"thumbnail"`
	Title     string `json:"title"`
}

// CarInspectionItems 检测项集合
type CarInspectionItems []CarInspectionItem

// CarFeature 配置特征项（用于 cars.features JSONB）
// 对标懂车帝：天窗/导航/真皮/倒车雷达/座椅加热等
type CarFeature struct {
	Category string `json:"category"` // 分类：safety/comfort/entertainment/exterior/interior
	Code     string `json:"code"`     // 编码：sunroof/navigation/leather_seat
	Name     string `json:"name"`     // 名称
	Value    string `json:"value"`    // 取值（如：选装/标配/无）
}

// CarFeatures 配置特征集合
type CarFeatures []CarFeature

// CarTag 标签项（用于 cars.tags JSONB，最多 5 个）
type CarTag struct {
	Text  string `json:"text"`  // 标签文本
	Color string `json:"color"` // 标签颜色
}

// CarAccidentHistory 事故历史项（用于 cars.accident_history / car_inspections.accident_history JSONB）
// 对标瓜子：事故类型/位置/严重程度/维修日期
type CarAccidentHistory struct {
	Type         string `json:"type"`         // 类型：collision/flood/fire/overhaul/scratch
	Location     string `json:"location"`     // 事故位置
	Severity     string `json:"severity"`     // 严重程度：minor/moderate/severe
	Description  string `json:"description"`  // 详细描述
	RepairDate   string `json:"repairDate"`   // 维修日期
	RepairShop   string `json:"repairShop"`   // 维修厂
	RepairCost   float64 `json:"repairCost"`  // 维修费用
	HasPhoto     bool   `json:"hasPhoto"`     // 是否有照片
	PhotoURLs    []string `json:"photoURLs"`  // 照片 URL
}

// CarAccidentHistoryList 事故历史集合
type CarAccidentHistoryList []CarAccidentHistory

// CarImageItem 车源图片项（用于 car_images 表的扩展元数据，也用于 cars.images 冗余）
type CarImageItem struct {
	URL       string `json:"url"`
	Thumbnail string `json:"thumbnail"`
	Type      string `json:"type"`  // exterior/interior/engine/chassis/accident
	Title     string `json:"title"`
	Sort      int    `json:"sort"`
}

// ContractAttachment 合同附件项（用于 car_contracts.attachments JSONB）
type ContractAttachment struct {
	Type string `json:"type"` // 类型：id_card/vehicle_cert/contract_pdf/insurance/other
	Name string `json:"name"`
	URL  string `json:"url"`
	Size int64 `json:"size"`
}

// EvidenceImage 举报证据图项（用于 car_reports.evidence_images JSONB）
type EvidenceImage struct {
	URL       string `json:"url"`
	Thumbnail string `json:"thumbnail"`
	Title     string `json:"title"`
}

// ReviewImage 评价图片项（用于 car_reviews.images / append_images JSONB）
type ReviewImage struct {
	URL       string `json:"url"`
	Thumbnail string `json:"thumbnail"`
}

// AuditRuleThreshold 审核规则阈值（用于 car_audit_rules.threshold JSONB）
type AuditRuleThreshold struct {
	Level       string  `json:"level"`        // high/medium/low
	MinPrice    float64 `json:"min_price"`    // 最低价
	MaxPrice    float64 `json:"max_price"`    // 最高价
	MaxCount    int     `json:"max_count"`    // 最大次数
	WindowSec   int     `json:"window_sec"`   // 时间窗口（秒）
	Ratio       float64 `json:"ratio"`        // 倍率
	Description string  `json:"description"`  // 描述
}

// EvaluationFactor 评估因子（用于 car_evaluations.factors JSONB）
type EvaluationFactor struct {
	Name    string  `json:"name"`    // 因子名：品牌/型号/年份/里程/车况/区域
	Weight  float64 `json:"weight"`  // 权重
	Score   float64 `json:"score"`   // 评分
	Comment string  `json:"comment"` // 说明
}

// SimilarDeal 相似成交（用于 car_evaluations.similar_deals JSONB）
type SimilarDeal struct {
	Model       string  `json:"model"`
	Year        int     `json:"year"`
	Mileage     float64 `json:"mileage"`
	DealPrice   float64 `json:"dealPrice"`
	DealDate    string  `json:"dealDate"`
	Location    string  `json:"location"`
	Similarity  float64 `json:"similarity"` // 相似度 0-1
}

// TransferDocument 过户材料项（用于 car_transfer.documents JSONB）
type TransferDocument struct {
	Type     string `json:"type"`     // 类型：id_card/vehicle_cert/insurance/policy/other
	Name     string `json:"name"`
	URL      string `json:"url"`
	Status   string `json:"status"`   // pending/submitted/verified/rejected
	Remark   string `json:"remark"`
}
