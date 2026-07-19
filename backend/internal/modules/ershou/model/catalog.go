// Package model 标签库 + 品牌库 + 型号库 + 分类属性配置（对标转转）
package model

import (
	"wuchang-tongcheng/internal/pkg/database"
)

// === 标签类型常量 ===
const (
	TagTypeSmart     = "smart"     // 智能标签（AI 自动识别）
	TagTypeOperation = "operation" // 运营标签（精选/新品/爆款/特价）
	TagTypeCustom    = "custom"    // 用户自定义标签
)

// ErshouTag 标签库表（对标转转）
type ErshouTag struct {
	database.BaseModel // id/created_at/updated_at/deleted_at
	Name       string `gorm:"size:64;not null;uniqueIndex:uniq_tag_name_type" json:"name"`     // 标签名称
	Type       string `gorm:"size:16;not null;uniqueIndex:uniq_tag_name_type;index" json:"type"` // smart/operation/custom
	Color      string `gorm:"size:16;default:'#409EFF'" json:"color"`                          // 标签颜色（HEX）
	Icon       string `gorm:"size:64" json:"icon"`                                              // 标签图标
	Background string `gorm:"size:32" json:"background"`                                         // 背景色（用于卡片展示）
	Status     int    `gorm:"default:1;index" json:"status"`                                     // 0禁用 1启用
	Sort       int    `gorm:"default:0" json:"sort"`                                               // 排序（越小越靠前）
	UseCount   int    `gorm:"default:0" json:"use_count"`                                         // 使用次数
	IsHot      bool   `gorm:"default:false;index" json:"is_hot"`                                  // 是否热门标签
	CreatorID  uint   `gorm:"index" json:"creator_id"`                                            // 创建人ID（运营/系统）
}

// TableName 表名（ers_ 前缀）
func (ErshouTag) TableName() string { return "ers_tags" }

// ErshouBrand 品牌库表（对标转转）
type ErshouBrand struct {
	database.BaseModel // id/created_at/updated_at/deleted_at
	Name           string `gorm:"size:128;not null;uniqueIndex:uniq_brand_name" json:"name"`     // 品牌名称
	Logo           string `gorm:"size:255" json:"logo"`                                            // 品牌 Logo URL
	EnglishName    string `gorm:"size:128" json:"english_name"`                                  // 英文名
	Description    string `gorm:"type:text" json:"description"`                                    // 品牌简介
	Country        string `gorm:"size:32" json:"country"`                                          // 国家
	OfficialVerified bool  `gorm:"default:false;index" json:"official_verified"`                  // 是否官方认证
	OfficialURL    string `gorm:"size:255" json:"official_url"`                                  // 官方网站
	CategoryIDs    JSONB  `gorm:"type:jsonb" json:"category_ids"`                                  // 关联分类ID数组 [1,2,3]
	Status         int    `gorm:"default:1;index" json:"status"`                                    // 0禁用 1启用
	Sort           int    `gorm:"default:0" json:"sort"`                                             // 排序
	UseCount       int    `gorm:"default:0" json:"use_count"`                                       // 使用次数
}

// TableName 表名（ers_ 前缀）
func (ErshouBrand) TableName() string { return "ers_brands" }

// ErshouModel 型号库表（对标转转）
// 如 iPhone 15 Pro Max 256G 深空黑
type ErshouModel struct {
	database.BaseModel // id/created_at/updated_at/deleted_at
	BrandID       uint   `gorm:"not null;index:idx_model_brand_name;uniqueIndex:uniq_model_name_brand" json:"brand_id"` // 关联品牌ID
	Name          string `gorm:"size:128;not null;uniqueIndex:uniq_model_name_brand;index:idx_model_brand_name" json:"name"` // 型号名称
	FullName      string `gorm:"size:255" json:"full_name"`                                                            // 完整名称（含规格）
	Specifications JSONB  `gorm:"type:jsonb" json:"specifications"`                                                  // 规格参数 {内存:256G,屏幕:6.7寸,处理器:A17Pro}
	Image         string `gorm:"size:255" json:"image"`                                                              // 型号图片
	ReleaseDate   string `gorm:"size:16" json:"release_date"`                                                       // 发布日期（YYYY-MM）
	Status        int    `gorm:"default:1;index" json:"status"`                                                      // 0禁用 1启用
	Sort          int    `gorm:"default:0" json:"sort"`                                                              // 排序
	UseCount      int    `gorm:"default:0" json:"use_count"`                                                          // 使用次数
	ReferencePrice float64 `gorm:"type:decimal(12,2);default:0" json:"reference_price"`                                // 参考价（市场价）
}

// TableName 表名（ers_ 前缀）
func (ErshouModel) TableName() string { return "ers_models" }

// === 属性类型常量 ===
const (
	CategoryAttrTypeString   = "string"   // 文本
	CategoryAttrTypeNumber   = "number"   // 数字
	CategoryAttrTypeSelect   = "select"   // 单选
	CategoryAttrTypeMultiSelect = "multi_select" // 多选
	CategoryAttrTypeDate     = "date"     // 日期
	CategoryAttrTypeBoolean  = "boolean"  // 布尔
)

// ErshouCategoryAttr 分类属性配置表（对标转转）
// 如：数码分类 → 内存/屏幕/处理器；服装分类 → 尺码/颜色/面料
type ErshouCategoryAttr struct {
	database.BaseModel // id/created_at/updated_at/deleted_at
	CategoryID    uint   `gorm:"not null;index:idx_cat_attr_category;uniqueIndex:uniq_cat_attr_name" json:"category_id"` // 关联分类ID
	AttrName      string `gorm:"size:64;not null;uniqueIndex:uniq_cat_attr_name;index:idx_cat_attr_category" json:"attr_name"` // 属性名（如 内存/屏幕/尺码）
	AttrKey       string `gorm:"size:64" json:"attr_key"`                                                                  // 属性键（英文/拼音，如 memory/screen/size）
	AttrType      string `gorm:"size:32;not null;default:'string'" json:"attr_type"`                                       // string/number/select/multi_select/date/boolean
	Options       JSONB  `gorm:"type:jsonb" json:"options"`                                                                // 可选值 JSON（如 ["8G","16G","32G"]）
	Unit          string `gorm:"size:32" json:"unit"`                                                                      // 单位（GB/英寸/cm）
	IsRequired    bool   `gorm:"default:false" json:"is_required"`                                                          // 是否必填
	IsFilterable  bool   `gorm:"default:false;index" json:"is_filterable"`                                                  // 是否可筛选（用于搜索过滤）
	IsSearchable  bool   `gorm:"default:false" json:"is_searchable"`                                                       // 是否可搜索
	DefaultValue  string `gorm:"size:255" json:"default_value"`                                                            // 默认值
	Placeholder   string `gorm:"size:255" json:"placeholder"`                                                              // 输入提示
	Description   string `gorm:"size:500" json:"description"`                                                              // 属性描述
	Status        int    `gorm:"default:1" json:"status"`                                                                    // 0禁用 1启用
	Sort          int    `gorm:"default:0" json:"sort"`                                                                       // 排序
}

// TableName 表名（ers_ 前缀）
func (ErshouCategoryAttr) TableName() string { return "ers_category_attrs" }
