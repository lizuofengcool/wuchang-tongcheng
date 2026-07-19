// Package dto 同城车辆买卖数据传输对象 - 车型库/分类
package dto

import (
	"wuchang-tongcheng/internal/pkg/utils"
)

// ModelInfo 车型详情响应
type ModelInfo struct {
	ID               uint    `json:"id"`
	Brand            string  `json:"brand"`
	BrandLogo        string  `json:"brand_logo"`
	Series           string  `json:"series"`
	ModelName        string  `json:"model_name"`
	Year             int     `json:"year"`
	Trim             string  `json:"trim"`
	CarType          string  `json:"car_type"`
	Displacement     float64 `json:"displacement"`
	Transmission     string  `json:"transmission"`
	FuelType         string  `json:"fuel_type"`
	EmissionStandard string  `json:"emission_standard"`
	SeatCount        int     `json:"seat_count"`
	DoorCount        int     `json:"door_count"`
	ExteriorColor    string  `json:"exterior_color"`
	InteriorColor    string  `json:"interior_color"`
	GuidePrice       float64 `json:"guide_price"`
	DepreciationRate float64 `json:"depreciation_rate"`
	EngineType       string  `json:"engine_type"`
	Horsepower       int     `json:"horsepower"`
	Description      string  `json:"description"`
	CoverImage       string  `json:"cover_image"`
	Status           int     `json:"status"`
	StatusText       string  `json:"status_text"`
	Sort             int     `json:"sort"`
	UseCount         int     `json:"use_count"`
	CreatedAt        string  `json:"created_at"`
	UpdatedAt        string  `json:"updated_at"`
}

// CreateModelRequest 创建车型请求
type CreateModelRequest struct {
	Brand            string  `json:"brand" binding:"required,max=64"`
	BrandLogo        string  `json:"brand_logo" binding:"max=255"`
	Series           string  `json:"series" binding:"max=64"`
	ModelName        string  `json:"model_name" binding:"required,max=128"`
	Year             int     `json:"year" binding:"required,min=1900,max=2100"`
	Trim             string  `json:"trim" binding:"max=64"`
	CarType          string  `json:"car_type" binding:"omitempty,oneof=sedan suv mpv new_energy sports truck van bus"`
	Displacement     float64 `json:"displacement"`
	Transmission     string  `json:"transmission" binding:"omitempty,oneof=manual auto cvt dct amt"`
	FuelType         string  `json:"fuel_type" binding:"omitempty,oneof=gasoline diesel hybrid pure_electric range_extender hydrogen"`
	EmissionStandard string  `json:"emission_standard"`
	SeatCount        int     `json:"seat_count"`
	DoorCount        int     `json:"door_count"`
	ExteriorColor    string  `json:"exterior_color" binding:"max=32"`
	InteriorColor    string  `json:"interior_color" binding:"max=32"`
	GuidePrice       float64 `json:"guide_price"`
	DepreciationRate float64 `json:"depreciation_rate"`
	EngineType       string  `json:"engine_type" binding:"max=64"`
	Horsepower       int     `json:"horsepower"`
	Description      string  `json:"description"`
	CoverImage       string  `json:"cover_image" binding:"max=255"`
	Status           int     `json:"status" binding:"omitempty,oneof=0 1 2"`
	Sort             int     `json:"sort"`
}

// UpdateModelRequest 更新车型请求
type UpdateModelRequest struct {
	Brand            *string  `json:"brand" binding:"omitempty,max=64"`
	BrandLogo        *string  `json:"brand_logo" binding:"omitempty,max=255"`
	Series           *string  `json:"series" binding:"omitempty,max=64"`
	ModelName        *string  `json:"model_name" binding:"omitempty,max=128"`
	Year             *int     `json:"year" binding:"omitempty,min=1900,max=2100"`
	Trim             *string  `json:"trim" binding:"omitempty,max=64"`
	CarType          *string  `json:"car_type" binding:"omitempty,oneof=sedan suv mpv new_energy sports truck van bus"`
	Displacement     *float64 `json:"displacement"`
	Transmission     *string  `json:"transmission" binding:"omitempty,oneof=manual auto cvt dct amt"`
	FuelType         *string  `json:"fuel_type" binding:"omitempty,oneof=gasoline diesel hybrid pure_electric range_extender hydrogen"`
	EmissionStandard *string  `json:"emission_standard"`
	SeatCount        *int     `json:"seat_count"`
	DoorCount        *int     `json:"door_count"`
	ExteriorColor    *string  `json:"exterior_color" binding:"omitempty,max=32"`
	InteriorColor    *string  `json:"interior_color" binding:"omitempty,max=32"`
	GuidePrice       *float64 `json:"guide_price"`
	DepreciationRate *float64 `json:"depreciation_rate"`
	EngineType       *string  `json:"engine_type" binding:"omitempty,max=64"`
	Horsepower       *int     `json:"horsepower"`
	Description      *string  `json:"description"`
	CoverImage       *string  `json:"cover_image" binding:"omitempty,max=255"`
	Status           *int     `json:"status" binding:"omitempty,oneof=0 1 2"`
	Sort             *int     `json:"sort"`
}

// ModelListRequest 车型列表请求
type ModelListRequest struct {
	Brand           string `form:"brand" json:"brand"`
	Series          string `form:"series" json:"series"`
	CarType         string `form:"car_type" json:"car_type"`
	FuelType        string `form:"fuel_type" json:"fuel_type"`
	Keyword         string `form:"keyword" json:"keyword"`
	Status          *int   `form:"status" json:"status"`
	utils.Pagination
}

// CategoryInfo 车型分类详情响应
type CategoryInfo struct {
	ID          uint   `json:"id"`
	Name        string `json:"name"`
	Code        string `json:"code"`
	ParentID    uint   `json:"parent_id"`
	Level       int    `json:"level"`
	CarType     string `json:"car_type"`
	Icon        string `json:"icon"`
	Color       string `json:"color"`
	Description string `json:"description"`
	Sort        int    `json:"sort"`
	Status      int    `json:"status"`
	StatusText  string `json:"status_text"`
	CarCount    int    `json:"car_count"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

// CreateCategoryRequest 创建分类请求
type CreateCategoryRequest struct {
	Name        string `json:"name" binding:"required,max=64"`
	Code        string `json:"code" binding:"required,max=64"`
	ParentID    uint   `json:"parent_id"`
	Level       int    `json:"level" binding:"omitempty,oneof=1 2 3"`
	CarType     string `json:"car_type" binding:"omitempty,oneof=sedan suv mpv new_energy sports truck van bus"`
	Icon        string `json:"icon" binding:"max=64"`
	Color       string `json:"color" binding:"max=32"`
	Description string `json:"description" binding:"max=500"`
	Sort        int    `json:"sort"`
	Status      int    `json:"status" binding:"omitempty,oneof=0 1 2"`
}

// UpdateCategoryRequest 更新分类请求
type UpdateCategoryRequest struct {
	Name        *string `json:"name" binding:"omitempty,max=64"`
	Code        *string `json:"code" binding:"omitempty,max=64"`
	ParentID    *uint   `json:"parent_id"`
	Level       *int    `json:"level" binding:"omitempty,oneof=1 2 3"`
	CarType     *string `json:"car_type" binding:"omitempty,oneof=sedan suv mpv new_energy sports truck van bus"`
	Icon        *string `json:"icon" binding:"omitempty,max=64"`
	Color       *string `json:"color" binding:"omitempty,max=32"`
	Description *string `json:"description" binding:"omitempty,max=500"`
	Sort        *int    `json:"sort"`
	Status      *int    `json:"status" binding:"omitempty,oneof=0 1 2"`
}

// CategoryListRequest 分类列表请求
type CategoryListRequest struct {
	ParentID uint   `form:"parent_id" json:"parent_id"`
	Level    int    `form:"level" json:"level"`
	CarType  string `form:"car_type" json:"car_type"`
	Status   *int   `form:"status" json:"status"`
	utils.Pagination
}

// BrandInfo 品牌信息（聚合，从 car_models 提取）
type BrandInfo struct {
	Brand     string `json:"brand"`
	BrandLogo string `json:"brand_logo"`
	SeriesCount int  `json:"series_count"`
	ModelCount int  `json:"model_count"`
	UseCount  int    `json:"use_count"`
}

// BrandListRequest 品牌列表请求
type BrandListRequest struct {
	Keyword string `form:"keyword" json:"keyword"`
	CarType string `form:"car_type" json:"car_type"`
	utils.Pagination
}
