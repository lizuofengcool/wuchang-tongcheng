// Package model 同城商城数据模型
// 依据产品需求文档 5.3：阶段2 商家服务扩展 - mall 同城商城
// 依据需求文档 1.10：4 维数据隔离（region_id + user_id + shop_id）
// 依据需求文档 7.1：通用字段 id/region_id/created_at/updated_at/deleted_at + status + audit_status
// 依据需求文档 7.2：主表 mall_shops（店铺列表）
// 对标：淘宝 / 京东 / 拼多多 / dismall 同城商城
package model

import (
	"database/sql/driver"
	"encoding/json"
	"errors"

	"wuchang-tongcheng/internal/pkg/database"
)

// === 店铺状态常量 ===
const (
	ShopStatusDraft     = 0 // 草稿
	ShopStatusOpened    = 1 // 已开业
	ShopStatusClosed    = 2 // 已关闭
	ShopStatusFrozen    = 3 // 已冻结
	ShopStatusExpired   = 4 // 已过期
)

// === 商品状态常量 ===
const (
	ProductStatusDraft     = 0 // 草稿
	ProductStatusOnSale    = 1 // 在售
	ProductStatusOffShelf  = 2 // 下架
	ProductStatusSoldOut   = 3 // 售罄
	ProductStatusRecycled  = 4 // 回收站
)

// === 商品审核状态常量 ===
const (
	ProductAuditPending  = 0 // 待审
	ProductAuditApproved = 1 // 通过
	ProductAuditRejected = 2 // 拒绝
)

// === 分类状态常量 ===
const (
	CategoryStatusDisabled = 0 // 禁用
	CategoryStatusEnabled  = 1 // 启用
)

// === 购物车选中状态 ===
const (
	CartUnselected = 0 // 未选中
	CartSelected   = 1 // 已选中
)

// === 订单状态常量 ===
const (
	OrderStatusPending   = 0 // 待付款
	OrderStatusPaid      = 1 // 已付款
	OrderStatusShipped   = 2 // 已发货
	OrderStatusReceived  = 3 // 已收货
	OrderStatusCompleted = 4 // 已完成
	OrderStatusCancelled = 5 // 已取消
	OrderStatusRefunded  = 6 // 已退款
	OrderStatusClosed    = 7 // 已关闭
)

// === 支付状态常量 ===
const (
	PaymentStatusPending  = 0 // 待支付
	PaymentStatusSuccess  = 1 // 成功
	PaymentStatusFailed   = 2 // 失败
	PaymentStatusRefunded = 3 // 已退款
	PaymentStatusClosed   = 4 // 已关闭
)

// === 支付方式常量 ===
const (
	PaymentMethodWechat   = "wechat"   // 微信支付
	PaymentMethodAlipay   = "alipay"   // 支付宝
	PaymentMethodBalance  = "balance"  // 余额支付
	PaymentMethodCod      = "cod"      // 货到付款
	PaymentMethodBankcard = "bankcard" // 银行卡
)

// === 退款状态常量 ===
const (
	RefundStatusPending  = 0 // 待审核
	RefundStatusApproved = 1 // 已同意
	RefundStatusRejected = 2 // 已拒绝
	RefundStatusRefunded = 3 // 已退款
	RefundStatusClosed   = 4 // 已关闭
)

// === 退款类型常量 ===
const (
	RefundTypeOnlyRefund = "only_refund" // 仅退款
	RefundTypeReturn     = "return"      // 退货退款
)

// === 评价状态常量 ===
const (
	ReviewStatusPending  = 0 // 待审
	ReviewStatusApproved = 1 // 通过
	ReviewStatusRejected = 2 // 拒绝
	ReviewStatusHidden   = 3 // 隐藏
)

// === 物流状态常量 ===
const (
	LogisticsStatusPending    = 0 // 待发货
	LogisticsStatusShipped    = 1 // 已发货
	LogisticsStatusInTransit  = 2 // 运输中
	LogisticsStatusDelivered  = 3 // 已派送
	LogisticsStatusReceived   = 4 // 已签收
	LogisticsStatusReturned   = 5 // 已退回
)

// === 收货地址默认标记 ===
const (
	AddressNotDefault = 0 // 非默认
	AddressIsDefault  = 1 // 默认
)

// === 审核规则类型常量 ===
const (
	AuditRuleTypeSensitiveWord = "sensitive_word" // 敏感词
	AuditRuleTypeProhibited    = "prohibited"     // 违禁内容
	AuditRuleTypeContact       = "contact"        // 联系方式
	AuditRuleTypePriceCheck    = "price_check"    // 价格校验
	AuditRuleTypeFrequency     = "frequency"      // 频率
)

// === 审核动作常量 ===
const (
	AuditActionReject   = "reject"   // 拒绝
	AuditActionApproval = "approval" // 通过
	AuditActionFilter   = "filter"   // 过滤
	AuditActionLimit    = "limit"    // 限制
)

// === 统计类型常量 ===
const (
	StatTypeDaily   = "daily"   // 日统计
	StatTypeShop    = "shop"    // 店铺统计
	StatTypeProduct = "product" // 商品统计
	StatTypeCategory = "category" // 分类统计
)

// === 举报状态常量 ===
const (
	ReportStatusPending     = 0 // 待处理
	ReportStatusWarning     = 1 // 已核实警告
	ReportStatusTakedown    = 2 // 已下架
	ReportStatusBanned      = 3 // 已封号
	ReportStatusDismissed   = 4 // 已驳回
	ReportStatusTransferred = 5 // 已转交
)

// === 举报目标类型常量 ===
const (
	ReportTargetShop    = "shop"    // 店铺
	ReportTargetProduct = "product" // 商品
	ReportTargetReview  = "review"  // 评价
	ReportTargetOrder   = "order"   // 订单
	ReportTargetUser    = "user"    // 用户
)

// === 举报类型常量 ===
const (
	ReportTypePorn         = "porn"         // 色情低俗
	ReportTypeScam         = "scam"         // 欺诈
	ReportTypeFake         = "fake"         // 售假
	ReportTypeProhibited   = "prohibited"   // 违禁品
	ReportTypeInfringement = "infringement" // 侵权
	ReportTypeSpam         = "spam"         // 垃圾信息
	ReportTypeOther        = "other"        // 其他
)

// === 店铺类型常量 ===
const (
	ShopTypePersonal  = "personal"  // 个人店铺
	ShopTypeEnterprise = "enterprise" // 企业店铺
	ShopTypeFlagship  = "flagship"  // 旗舰店
)

// === 商品类型常量 ===
const (
	ProductTypePhysical = "physical" // 实物商品
	ProductTypeVirtual  = "virtual"  // 虚拟商品
	ProductTypeService  = "service"  // 服务商品
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
	return errors.New("mall.JSONB.Scan: unsupported source type")
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
		return errors.New("mall.JSONB.UnmarshalJSON: nil pointer")
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

// ProductImageItem 商品图片项
type ProductImageItem struct {
	URL       string `json:"url"`
	Thumbnail string `json:"thumbnail"`
	Type      string `json:"type"` // main/detail/sku
	Sort      int    `json:"sort"`
}

// ProductSpecItem 商品规格项（如：颜色/尺寸）
type ProductSpecItem struct {
	Name   string   `json:"name"`   // 规格名（颜色）
	Values []string `json:"values"` // 规格值（红/黄/蓝）
}

// SkuSpecValue SKU 规格值（如：颜色=红）
type SkuSpecValue struct {
	Name  string `json:"name"`  // 规格名
	Value string `json:"value"` // 规格值
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

// OrderItemSnapshot 订单商品快照（用于订单详情 JSONB 冗余存储）
type OrderItemSnapshot struct {
	ProductID   uint   `json:"product_id"`
	SkuID       uint   `json:"sku_id"`
	Name        string `json:"name"`
	Image       string `json:"image"`
	Price       string `json:"price"`
	Quantity    int    `json:"quantity"`
	Specs       string `json:"specs"`
}

// 确保 database 包被引用（model 文件中使用 RegionBaseModel / BaseModel）
var _ = database.RegionBaseModel{}
