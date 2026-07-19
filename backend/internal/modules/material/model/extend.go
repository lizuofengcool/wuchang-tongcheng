// Package model 素材中台扩展数据模型
// 依据 014_material_full.sql：分类/标签/图片标签/搜索历史/相似结果/OCR结果
package model

import (
	"time"
)

// === 标签类型 ===
const (
	TagTypeGeneral = "general"
	TagTypeObject  = "object"
	TagTypeScene   = "scene"
	TagTypeColor   = "color"
	TagTypeShape   = "shape"
)

// === 搜索类型 ===
const (
	SearchTypeImage   = "image"
	SearchTypeText    = "text"
	SearchTypeFeature = "feature"
)

// === 标签来源 ===
const (
	TagSourceManual = "manual"
	TagSourceAuto   = "auto"
)

// === OCR 引擎 ===
const (
	OCREngineLocal   = "local"
	OCREngineAliyun  = "aliyun"
	OCREngineTencent = "tencent"
	OCREngineBaidu   = "baidu"
)

// === 特征算法 ===
const (
	AlgoPHash     = "phash"
	AlgoCNN       = "cnn"
	AlgoColorHist = "color_hist"
)

// Category 图片分类
type Category struct {
	ID          uint      `gorm:"primarykey" json:"id"`
	RegionID    uint      `gorm:"index;not null;default:1" json:"region_id"`
	Name        string    `gorm:"size:64;not null;uniqueIndex:uk_material_categories,priority:1" json:"name"`
	ParentID    uint      `gorm:"not null;default:0;uniqueIndex:uk_material_categories,priority:2;index" json:"parent_id"`
	Level       int       `gorm:"default:1" json:"level"`
	Icon        string    `gorm:"size:256;not null;default:''" json:"icon"`
	Description string    `gorm:"size:256;not null;default:''" json:"description"`
	Sort        int       `gorm:"default:0" json:"sort"`
	ImageCount  int       `gorm:"default:0" json:"image_count"`
	Status      int       `gorm:"default:1;index" json:"status"`
	CreatedAt   time.Time `gorm:"not null;default:now()" json:"created_at"`
	UpdatedAt   time.Time `gorm:"not null;default:now()" json:"updated_at"`
}

// TableName 表名
func (Category) TableName() string { return "material_categories" }

// Tag 图片标签
type Tag struct {
	ID          uint      `gorm:"primarykey" json:"id"`
	RegionID    uint      `gorm:"index;not null;default:1" json:"region_id"`
	Name        string    `gorm:"size:64;not null;uniqueIndex" json:"name"`
	TagType     string    `gorm:"size:32;default:'general';index" json:"tag_type"`
	Description string    `gorm:"size:256;not null;default:''" json:"description"`
	Icon        string    `gorm:"size:256;not null;default:''" json:"icon"`
	UsageCount  int       `gorm:"default:0;index" json:"usage_count"`
	Status      int       `gorm:"default:1;index" json:"status"`
	CreatedAt   time.Time `gorm:"not null;default:now()" json:"created_at"`
	UpdatedAt   time.Time `gorm:"not null;default:now()" json:"updated_at"`
}

// TableName 表名
func (Tag) TableName() string { return "material_tags" }

// ImageTag 图片-标签关联
type ImageTag struct {
	ID         uint    `gorm:"primarykey" json:"id"`
	RegionID   uint    `gorm:"index;not null;default:1" json:"region_id"`
	ImageID    uint    `gorm:"not null;index;uniqueIndex:uk_material_image_tags,priority:1" json:"image_id"`
	TagID      uint    `gorm:"not null;index;uniqueIndex:uk_material_image_tags,priority:2" json:"tag_id"`
	Source     string  `gorm:"size:16;not null;default:'manual'" json:"source"`
	Confidence float64 `gorm:"type:decimal(5,2);default:1.00" json:"confidence"`
	CreatedAt  time.Time `gorm:"not null;default:now()" json:"created_at"`
}

// TableName 表名
func (ImageTag) TableName() string { return "material_image_tags" }

// SearchHistory 搜索历史
type SearchHistory struct {
	ID              uint      `gorm:"primarykey" json:"id"`
	RegionID        uint      `gorm:"index;not null;default:1" json:"region_id"`
	UserID         uint      `gorm:"index;not null" json:"user_id"`
	SearchType     string    `gorm:"size:16;default:'image'" json:"search_type"`
	QueryImageID   uint      `gorm:"default:0" json:"query_image_id"`
	QueryText      string    `gorm:"size:256;not null;default:''" json:"query_text"`
	ResultCount    int       `gorm:"default:0" json:"result_count"`
	TopSimilarity  float64   `gorm:"type:decimal(5,2);default:0.00" json:"top_similarity"`
	CostMs         int       `gorm:"default:0" json:"cost_ms"`
	CreatedAt      time.Time `gorm:"not null;default:now();index" json:"created_at"`
}

// TableName 表名
func (SearchHistory) TableName() string { return "material_search_history" }

// SimilarResult 相似图搜索结果
type SimilarResult struct {
	ID              uint      `gorm:"primarykey" json:"id"`
	RegionID        uint      `gorm:"index;not null;default:1" json:"region_id"`
	SearchHistoryID uint      `gorm:"default:0" json:"search_history_id"`
	SourceImageID   uint      `gorm:"not null;index" json:"source_image_id"`
	TargetImageID   uint      `gorm:"not null;index" json:"target_image_id"`
	Similarity      float64   `gorm:"type:decimal(5,2);default:0.00" json:"similarity"`
	FeatureAlgo     string    `gorm:"size:32;default:'phash';index" json:"feature_algo"`
	CreatedAt       time.Time `gorm:"not null;default:now()" json:"created_at"`
}

// TableName 表名
func (SimilarResult) TableName() string { return "material_similar_results" }

// OCRResult OCR 识别结果
type OCRResult struct {
	ID          uint      `gorm:"primarykey" json:"id"`
	RegionID    uint      `gorm:"index;not null;default:1" json:"region_id"`
	ImageID     uint      `gorm:"not null;index" json:"image_id"`
	FileID      string    `gorm:"size:64;not null;default:'';index" json:"file_id"`
	OCREngine   string    `gorm:"size:32;default:'local';index" json:"ocr_engine"`
	TextContent string    `gorm:"type:text;not null;default:''" json:"text_content"`
	TextBlocks  string    `gorm:"type:jsonb;default:'[]'::jsonb" json:"text_blocks"`
	Language    string    `gorm:"size:16;not null;default:'zh'" json:"language"`
	Confidence  float64   `gorm:"type:decimal(5,2);default:0.00" json:"confidence"`
	CostMs      int       `gorm:"default:0" json:"cost_ms"`
	CreatedAt   time.Time `gorm:"not null;default:now()" json:"created_at"`
	UpdatedAt   time.Time `gorm:"not null;default:now()" json:"updated_at"`
}

// TableName 表名
func (OCRResult) TableName() string { return "material_ocr_results" }
