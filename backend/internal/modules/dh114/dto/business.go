// Package dto 同城114数据传输对象 - 商户详细信息
package dto

import (
	"time"

	"wuchang-tongcheng/internal/pkg/utils"
)

// BusinessInfo 商户详细信息响应
type BusinessInfo struct {
	ID                uint       `json:"id"`
	Dh114ID           uint       `json:"dh114_id"`
	BusinessName      string     `json:"business_name"`
	LicenseNo         string     `json:"license_no"`
	LicenseImage      string     `json:"license_image"`
	LegalPerson       string     `json:"legal_person"`
	BusinessScope     string     `json:"business_scope"`
	RegisteredCapital float64    `json:"registered_capital"`
	EstablishedDate  *time.Time `json:"established_date"`
	RegisteredAddress string     `json:"registered_address"`
	OpeningHours      string     `json:"opening_hours"`
	ClosingHours      string     `json:"closing_hours"`
	OpenAllDay        bool       `json:"open_all_day"`
	ClosedDays        interface{} `json:"closed_days"`
	PriceAvg          float64    `json:"price_avg"`
	PriceRangeMin     float64    `json:"price_range_min"`
	PriceRangeMax     float64    `json:"price_range_max"`
	Website           string     `json:"website"`
	Wechat            string     `json:"wechat"`
	WechatQR          string     `json:"wechat_qr"`
	Email             string     `json:"email"`
	Facilities        interface{} `json:"facilities"`
	VerificationStatus int       `json:"verification_status"`
	VerificationStatusText string `json:"verification_status_text"`
	VerifiedAt        *time.Time `json:"verified_at"`
	ValidUntil        *time.Time `json:"valid_until"`
	RegionID          uint       `json:"region_id"`
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`
}

// CreateBusinessRequest 创建商户详情请求
type CreateBusinessRequest struct {
	Dh114ID           uint    `json:"dh114_id" binding:"required"`
	BusinessName      string  `json:"business_name" binding:"max=200"`
	LicenseNo         string  `json:"license_no" binding:"max=64"`
	LicenseImage      string  `json:"license_image" binding:"max=255"`
	LegalPerson       string  `json:"legal_person" binding:"max=64"`
	LegalPersonIDCard string  `json:"legal_person_id_card" binding:"max=32"`
	BusinessScope     string  `json:"business_scope"`
	RegisteredCapital float64 `json:"registered_capital"`
	EstablishedDate   *time.Time `json:"established_date"`
	RegisteredAddress string  `json:"registered_address" binding:"max=500"`
	OpeningHours      string  `json:"opening_hours" binding:"max=32"`
	ClosingHours      string  `json:"closing_hours" binding:"max=32"`
	OpenAllDay        bool    `json:"open_all_day"`
	ClosedDays        interface{} `json:"closed_days"`
	PriceAvg          float64 `json:"price_avg"`
	PriceRangeMin     float64 `json:"price_range_min"`
	PriceRangeMax     float64 `json:"price_range_max"`
	Website           string  `json:"website" binding:"max=255"`
	Wechat            string  `json:"wechat" binding:"max=64"`
	WechatQR          string  `json:"wechat_qr" binding:"max=255"`
	Email             string  `json:"email" binding:"max=128"`
	Facilities        interface{} `json:"facilities"`
}

// UpdateBusinessRequest 更新商户详情请求
type UpdateBusinessRequest struct {
	BusinessName      *string `json:"business_name" binding:"max=200"`
	LicenseNo         *string `json:"license_no" binding:"max=64"`
	LicenseImage      *string `json:"license_image" binding:"max=255"`
	LegalPerson       *string `json:"legal_person" binding:"max=64"`
	LegalPersonIDCard *string `json:"legal_person_id_card" binding:"max=32"`
	BusinessScope     *string `json:"business_scope"`
	RegisteredCapital *float64 `json:"registered_capital"`
	EstablishedDate   *time.Time `json:"established_date"`
	RegisteredAddress *string `json:"registered_address" binding:"max=500"`
	OpeningHours      *string `json:"opening_hours" binding:"max=32"`
	ClosingHours      *string `json:"closing_hours" binding:"max=32"`
	OpenAllDay        *bool   `json:"open_all_day"`
	ClosedDays        interface{} `json:"closed_days"`
	PriceAvg          *float64 `json:"price_avg"`
	PriceRangeMin     *float64 `json:"price_range_min"`
	PriceRangeMax     *float64 `json:"price_range_max"`
	Website           *string `json:"website" binding:"max=255"`
	Wechat            *string `json:"wechat" binding:"max=64"`
	WechatQR          *string `json:"wechat_qr" binding:"max=255"`
	Email             *string `json:"email" binding:"max=128"`
	Facilities        interface{} `json:"facilities"`
}

// BusinessListRequest 商户详情列表请求
type BusinessListRequest struct {
	Dh114ID            uint   `form:"dh114_id" json:"dh114_id"`
	VerificationStatus *int   `form:"verification_status" json:"verification_status"`
	Keyword            string `form:"keyword" json:"keyword"`
	utils.Pagination
}
