// Package dto 同城114数据传输对象 - 优惠券
package dto

import (
	"time"

	"wuchang-tongcheng/internal/pkg/utils"
)

// CouponInfo 优惠券详情响应
type CouponInfo struct {
	ID              uint       `json:"id"`
	CouponNo        string     `json:"coupon_no"`
	Dh114ID         uint       `json:"dh114_id"`
	BusinessID      uint       `json:"business_id"`
	Title           string     `json:"title"`
	Description     string     `json:"description"`
	CoverImage      string     `json:"cover_image"`
	CouponType      string     `json:"coupon_type"`
	CouponTypeText  string     `json:"coupon_type_text"`
	FaceValue       float64    `json:"face_value"`
	Threshold       float64    `json:"threshold"`
	Discount        float64    `json:"discount"`
	MaxDiscount     float64    `json:"max_discount"`
	TotalCount      int        `json:"total_count"`
	IssuedCount     int        `json:"issued_count"`
	UsedCount       int        `json:"used_count"`
	PerUserLimit    int        `json:"per_user_limit"`
	StartTime       *time.Time `json:"start_time"`
	EndTime         *time.Time `json:"end_time"`
	ValidStart      *time.Time `json:"valid_start"`
	ValidEnd        *time.Time `json:"valid_end"`
	ValidDays       int        `json:"valid_days"`
	UseInstructions interface{} `json:"use_instructions"`
	UseThreshold    float64    `json:"use_threshold"`
	Status          int        `json:"status"`
	StatusText      string     `json:"status_text"`
	AuditStatus     int        `json:"audit_status"`
	AuditReason     string     `json:"audit_reason"`
	PublishedAt     *time.Time `json:"published_at"`
	Featured        bool       `json:"featured"`
	RegionID        uint       `json:"region_id"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

// CreateCouponRequest 创建优惠券请求
type CreateCouponRequest struct {
	Dh114ID         uint       `json:"dh114_id" binding:"required"`
	Title           string     `json:"title" binding:"required,max=200"`
	Description     string     `json:"description"`
	CoverImage      string     `json:"cover_image" binding:"max=255"`
	CouponType      string     `json:"coupon_type" binding:"required,oneof=discount full_reduction cash gift"`
	FaceValue       float64    `json:"face_value" binding:"min=0"`
	Threshold       float64    `json:"threshold" binding:"min=0"`
	Discount        float64    `json:"discount" binding:"min=0,max=1"`
	MaxDiscount     float64    `json:"max_discount" binding:"min=0"`
	TotalCount      int        `json:"total_count" binding:"min=0"`
	PerUserLimit    int        `json:"per_user_limit" binding:"min=1"`
	StartTime       *time.Time `json:"start_time"`
	EndTime         *time.Time `json:"end_time"`
	ValidStart      *time.Time `json:"valid_start"`
	ValidEnd        *time.Time `json:"valid_end"`
	ValidDays       int        `json:"valid_days" binding:"min=0"`
	UseInstructions interface{} `json:"use_instructions"`
	UseThreshold    float64    `json:"use_threshold" binding:"min=0"`
	Status          int        `json:"status" binding:"omitempty,oneof=0 1"`
}

// UpdateCouponRequest 更新优惠券请求
type UpdateCouponRequest struct {
	Title           *string `json:"title" binding:"max=200"`
	Description     *string `json:"description"`
	CoverImage      *string `json:"cover_image" binding:"max=255"`
	FaceValue       *float64 `json:"face_value" binding:"min=0"`
	Threshold       *float64 `json:"threshold" binding:"min=0"`
	Discount        *float64 `json:"discount" binding:"min=0,max=1"`
	MaxDiscount     *float64 `json:"max_discount" binding:"min=0"`
	TotalCount      *int    `json:"total_count" binding:"min=0"`
	PerUserLimit    *int    `json:"per_user_limit" binding:"min=1"`
	StartTime       *time.Time `json:"start_time"`
	EndTime         *time.Time `json:"end_time"`
	ValidStart      *time.Time `json:"valid_start"`
	ValidEnd        *time.Time `json:"valid_end"`
	ValidDays       *int    `json:"valid_days" binding:"min=0"`
	UseInstructions interface{} `json:"use_instructions"`
	UseThreshold    *float64 `json:"use_threshold" binding:"min=0"`
	Status          *int    `json:"status" binding:"omitempty,oneof=0 1 2 3 4"`
}

// CouponListRequest 优惠券列表请求
type CouponListRequest struct {
	Dh114ID    uint   `form:"dh114_id" json:"dh114_id"`
	CouponType string `form:"coupon_type" json:"coupon_type"`
	Status     *int   `form:"status" json:"status"`
	Featured   *bool  `form:"featured" json:"featured"`
	Keyword    string `form:"keyword" json:"keyword"`
	utils.Pagination
}

// CouponAdminListRequest 管理后台优惠券列表请求
type CouponAdminListRequest struct {
	Dh114ID     uint   `form:"dh114_id" json:"dh114_id"`
	Status      *int   `form:"status" json:"status"`
	AuditStatus *int   `form:"audit_status" json:"audit_status"`
	Keyword     string `form:"keyword" json:"keyword"`
	utils.Pagination
}

// CouponReceiveRequest 领取优惠券请求
type CouponReceiveRequest struct {
	CouponID uint `json:"coupon_id" binding:"required"`
}
