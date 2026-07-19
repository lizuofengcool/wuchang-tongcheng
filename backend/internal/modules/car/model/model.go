// Package model 车型库表（对标懂车帝/汽车之家）
// 品牌/系列/年款/配置/指导价/折旧率
package model

import (
	"wuchang-tongcheng/internal/pkg/database"
)

// === 车型库状态常量 ===
const (
	ModelStatusDraft     = 0 // 草稿
	ModelStatusPublished = 1 // 已发布
	ModelStatusOffline   = 2 // 已下架
)

// CarModel 车型库表
type CarModel struct {
	database.BaseModel // id/created_at/updated_at/deleted_at
	Brand             string  `gorm:"size:64;not null;default:'';index" json:"brand"`                 // 品牌
	BrandLogo         string  `gorm:"size:255;not null;default:''" json:"brand_logo"`                 // 品牌 Logo
	Series            string  `gorm:"size:64;not null;default:'';index" json:"series"`                // 车系
	ModelName         string  `gorm:"size:128;not null;uniqueIndex:uniq_car_models_name_year_trim" json:"model_name"` // 车型名称
	Year              int     `gorm:"not null;default:0;index;uniqueIndex:uniq_car_models_name_year_trim" json:"year"` // 年款
	Trim              string  `gorm:"size:64;not null;default:'';uniqueIndex:uniq_car_models_name_year_trim" json:"trim"` // 配置款型
	CarType           string  `gorm:"size:32;not null;default:'sedan';index" json:"car_type"`         // sedan/suv/mpv/new_energy/sports/truck
	Displacement      float64 `gorm:"type:decimal(4,2);default:0" json:"displacement"`               // 排量
	Transmission      string  `gorm:"size:32;not null;default:''" json:"transmission"`               // 变速箱
	FuelType          string  `gorm:"size:32;not null;default:'gasoline';index" json:"fuel_type"`     // 燃油类型
	EmissionStandard  string  `gorm:"size:32;not null;default:''" json:"emission_standard"`           // 排放标准
	SeatCount         int     `gorm:"not null;default:5" json:"seat_count"`                           // 座位数
	DoorCount         int     `gorm:"not null;default:4" json:"door_count"`                           // 车门数
	ExteriorColor     string  `gorm:"size:32;not null;default:''" json:"exterior_color"`             // 默认外观色
	InteriorColor     string  `gorm:"size:32;not null;default:''" json:"interior_color"`             // 默认内饰色
	GuidePrice        float64 `gorm:"type:decimal(14,2);default:0" json:"guide_price"`               // 厂商指导价
	DepreciationRate  float64 `gorm:"type:decimal(5,2);default:0" json:"depreciation_rate"`           // 年折旧率（%）
	EngineType        string  `gorm:"size:64;not null;default:''" json:"engine_type"`                 // 发动机型号
	Horsepower        int     `gorm:"not null;default:0" json:"horsepower"`                           // 马力
	Description       string  `gorm:"type:text" json:"description"`                                   // 车型简介
	CoverImage        string  `gorm:"size:255;not null;default:''" json:"cover_image"`               // 封面图
	Status            int     `gorm:"default:1;index" json:"status"`                                  // 0草稿 1已发布 2下架
	Sort              int     `gorm:"not null;default:0;index" json:"sort"`                           // 排序
	UseCount          int     `gorm:"not null;default:0" json:"use_count"`                            // 引用次数
}

// TableName 表名（car_ 前缀）
func (CarModel) TableName() string { return "car_models" }
