// Package dto 同城商城 - 物流 DTO
package dto

import (
	"time"

	"wuchang-tongcheng/internal/pkg/utils"
)

// LogisticsInfo 物流详情响应
type LogisticsInfo struct {
	ID              uint       `json:"id"`
	OrderID         uint       `json:"order_id"`
	OrderNo         string     `json:"order_no"`
	UserID          uint       `json:"user_id"`
	ShopID          uint       `json:"shop_id"`
	Company         string     `json:"company"`
	CompanyCode     string     `json:"company_code"`
	TrackingNo      string     `json:"tracking_no"`
	CourierName     string     `json:"courier_name"`
	CourierPhone    string     `json:"courier_phone"`
	SenderName      string     `json:"sender_name"`
	SenderPhone     string     `json:"sender_phone"`
	SenderAddress   string     `json:"sender_address"`
	ReceiverName    string     `json:"receiver_name"`
	ReceiverPhone   string     `json:"receiver_phone"`
	ReceiverAddress string     `json:"receiver_address"`
	Status          int        `json:"status"`
	StatusText      string     `json:"status_text"`
	ShippedAt       *time.Time `json:"shipped_at"`
	InTransitAt     *time.Time `json:"in_transit_at"`
	DeliveredAt     *time.Time `json:"delivered_at"`
	ReceivedAt      *time.Time `json:"received_at"`
	ReturnedAt      *time.Time `json:"returned_at"`
	Traces          interface{} `json:"traces"`
	Weight          float64    `json:"weight"`
	Volume          float64    `json:"volume"`
	Pieces          int        `json:"pieces"`
	Freight         float64    `json:"freight"`
	InsuredFee      float64    `json:"insured_fee"`
	CodFee          float64    `json:"cod_fee"`
	RegionID        uint       `json:"region_id"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

// CreateLogisticsRequest 创建物流请求
type CreateLogisticsRequest struct {
	OrderID         uint    `json:"order_id" binding:"required"`
	Company         string  `json:"company" binding:"required,max=64"`
	CompanyCode     string  `json:"company_code" binding:"max=32"`
	TrackingNo      string  `json:"tracking_no" binding:"required,max=64"`
	CourierName     string  `json:"courier_name" binding:"max=32"`
	CourierPhone    string  `json:"courier_phone" binding:"max=32"`
	SenderName      string  `json:"sender_name" binding:"max=32"`
	SenderPhone     string  `json:"sender_phone" binding:"max=32"`
	SenderAddress   string  `json:"sender_address" binding:"max=500"`
	Weight          float64 `json:"weight"`
	Volume          float64 `json:"volume"`
	Pieces          int     `json:"pieces" binding:"omitempty,min=1"`
	Freight         float64 `json:"freight"`
	InsuredFee      float64 `json:"insured_fee"`
	CodFee          float64 `json:"cod_fee"`
}

// UpdateLogisticsRequest 更新物流请求
type UpdateLogisticsRequest struct {
	Company      *string  `json:"company" binding:"max=64"`
	CompanyCode  *string  `json:"company_code" binding:"max=32"`
	TrackingNo   *string  `json:"tracking_no" binding:"max=64"`
	CourierName  *string  `json:"courier_name" binding:"max=32"`
	CourierPhone *string  `json:"courier_phone" binding:"max=32"`
	Status       *int     `json:"status" binding:"omitempty,oneof=0 1 2 3 4 5"`
	Traces       interface{} `json:"traces"`
}

// LogisticsListRequest 物流列表请求
type LogisticsListRequest struct {
	utils.Pagination
	OrderID      uint   `form:"order_id" json:"order_id"`
	OrderNo      string `form:"order_no" json:"order_no"`
	ShopID       uint   `form:"shop_id" json:"shop_id"`
	UserID       uint   `form:"user_id" json:"user_id"`
	TrackingNo   string `form:"tracking_no" json:"tracking_no"`
	CompanyCode  string `form:"company_code" json:"company_code"`
	Status       *int   `form:"status" json:"status"`
	StartDate    string `form:"start_date" json:"start_date"`
	EndDate      string `form:"end_date" json:"end_date"`
	Keyword      string `form:"keyword" json:"keyword"`
	RegionID     uint   `form:"region_id" json:"region_id"`
}

// LogisticsTraceItem 物流轨迹条目
type LogisticsTraceItem struct {
	Time    string `json:"time"`
	Context string `json:"context"`
	Status  string `json:"status"`
	Location string `json:"location"`
}

// UpdateLogisticsStatusRequest 更新物流状态请求（对接物流回调时使用）
type UpdateLogisticsStatusRequest struct {
	OrderID    uint   `json:"order_id"`
	OrderNo    string `json:"order_no"`
	TrackingNo string `json:"tracking_no"`
	Status     int    `json:"status"`
	Traces     []LogisticsTraceItem `json:"traces"`
}
