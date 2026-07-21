// Package dto 同城商城 - 收货地址 DTO
package dto

import (
	"wuchang-tongcheng/internal/pkg/utils"
)

// AddressInfo 地址详情响应
type AddressInfo struct {
	ID           uint    `json:"id"`
	UserID       uint    `json:"user_id"`
	Name         string  `json:"name"`
	Phone        string  `json:"phone"`
	ZipCode      string  `json:"zip_code"`
	Province     string  `json:"province"`
	City         string  `json:"city"`
	District     string  `json:"district"`
	ProvinceCode string  `json:"province_code"`
	CityCode     string  `json:"city_code"`
	DistrictCode string  `json:"district_code"`
	Detail       string  `json:"detail"`
	Latitude     float64 `json:"latitude"`
	Longitude    float64 `json:"longitude"`
	Tag          string  `json:"tag"`
	IsDefault    bool    `json:"is_default"`
	Status       int     `json:"status"`
	RegionID     uint    `json:"region_id"`
}

// CreateAddressRequest 创建地址请求
type CreateAddressRequest struct {
	Name         string  `json:"name" binding:"required,max=64"`
	Phone        string  `json:"phone" binding:"required,max=32"`
	ZipCode      string  `json:"zip_code" binding:"max=16"`
	Province     string  `json:"province" binding:"required,max=64"`
	City         string  `json:"city" binding:"required,max=64"`
	District     string  `json:"district" binding:"max=64"`
	ProvinceCode string  `json:"province_code" binding:"max=16"`
	CityCode     string  `json:"city_code" binding:"max=16"`
	DistrictCode string  `json:"district_code" binding:"max=16"`
	Detail       string  `json:"detail" binding:"required,max=500"`
	Latitude     float64 `json:"latitude"`
	Longitude    float64 `json:"longitude"`
	Tag          string  `json:"tag" binding:"max=32"` // 家/公司/学校
	IsDefault    bool    `json:"is_default"`
}

// UpdateAddressRequest 更新地址请求
type UpdateAddressRequest struct {
	Name         *string  `json:"name" binding:"omitempty,max=64"`
	Phone        *string  `json:"phone" binding:"omitempty,max=32"`
	ZipCode      *string  `json:"zip_code" binding:"max=16"`
	Province     *string  `json:"province" binding:"max=64"`
	City         *string  `json:"city" binding:"max=64"`
	District     *string  `json:"district" binding:"max=64"`
	ProvinceCode *string  `json:"province_code" binding:"max=16"`
	CityCode     *string  `json:"city_code" binding:"max=16"`
	DistrictCode *string  `json:"district_code" binding:"max=16"`
	Detail       *string  `json:"detail" binding:"max=500"`
	Latitude     *float64 `json:"latitude"`
	Longitude    *float64 `json:"longitude"`
	Tag          *string  `json:"tag" binding:"max=32"`
	IsDefault    *bool    `json:"is_default"`
	Status       *int     `json:"status" binding:"omitempty,oneof=0 1"`
}

// AddressListRequest 地址列表请求
type AddressListRequest struct {
	utils.Pagination
	Keyword string `form:"keyword" json:"keyword"`
	UserID  uint   `form:"user_id" json:"user_id"`
	RegionID uint  `form:"region_id" json:"region_id"`
}

// SetDefaultAddressRequest 设默认地址请求
type SetDefaultAddressRequest struct {
	ID uint `json:"id" binding:"required"`
}
