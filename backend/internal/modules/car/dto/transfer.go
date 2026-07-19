// Package dto 同城车辆买卖数据传输对象 - 过户办理
package dto

import (
	"time"

	"wuchang-tongcheng/internal/pkg/utils"
)

// TransferInfo 过户办理详情响应
type TransferInfo struct {
	ID                 uint       `json:"id"`
	TransferNo         string     `json:"transfer_no"`
	CarID              uint       `json:"car_id"`
	ContractID         uint       `json:"contract_id"`
	ListingID          uint       `json:"listing_id"`
	SellerID           uint       `json:"seller_id"`
	SellerName         string     `json:"seller_name"`
	BuyerID            uint       `json:"buyer_id"`
	BuyerName          string     `json:"buyer_name"`
	AgentID            uint       `json:"agent_id"`
	AgentName          string     `json:"agent_name"`
	TransferType       string     `json:"transfer_type"`
	VehicleRegistration interface{} `json:"vehicle_registration"`
	Documents          interface{} `json:"documents"`
	TransferFee        float64    `json:"transfer_fee"`
	TaxFee             float64    `json:"tax_fee"`
	OtherFee           float64    `json:"other_fee"`
	Location           string     `json:"location"`
	AppointmentDate    *time.Time `json:"appointment_date"`
	AppointmentTime    string     `json:"appointment_time"`
	SubmittedAt        *time.Time `json:"submitted_at"`
	ReviewedAt         *time.Time `json:"reviewed_at"`
	CompletedAt        *time.Time `json:"completed_at"`
	CanceledAt         *time.Time `json:"canceled_at"`
	NewLicensePlate    string     `json:"new_license_plate"`
	NewRegistrationCert string     `json:"new_registration_cert"`
	Status             int        `json:"status"`
	StatusText         string     `json:"status_text"`
	Remark             string     `json:"remark"`
	RegionID           uint       `json:"region_id"`
	CreatedAt          time.Time  `json:"created_at"`
	UpdatedAt          time.Time  `json:"updated_at"`
}

// CreateTransferRequest 创建过户请求
type CreateTransferRequest struct {
	CarID              uint   `json:"car_id" binding:"required"`
	ContractID         uint   `json:"contract_id"`
	ListingID          uint   `json:"listing_id"`
	SellerID           uint   `json:"seller_id" binding:"required"`
	SellerName         string `json:"seller_name" binding:"required,max=50"`
	BuyerID            uint   `json:"buyer_id" binding:"required"`
	BuyerName          string `json:"buyer_name" binding:"required,max=50"`
	AgentID            uint   `json:"agent_id"`
	AgentName          string `json:"agent_name" binding:"max=50"`
	TransferType       string `json:"transfer_type" binding:"omitempty,oneof=sale replace gift inheritance relocation"`
	VehicleRegistration interface{} `json:"vehicle_registration"`
	Documents          interface{} `json:"documents"`
	TransferFee        float64 `json:"transfer_fee"`
	TaxFee             float64 `json:"tax_fee"`
	OtherFee           float64 `json:"other_fee"`
	Location           string `json:"location" binding:"max=255"`
	AppointmentDate    *time.Time `json:"appointment_date"`
	AppointmentTime    string `json:"appointment_time" binding:"max=32"`
	Remark             string `json:"remark"`
}

// UpdateTransferRequest 更新过户请求
type UpdateTransferRequest struct {
	AgentID            *uint   `json:"agent_id"`
	AgentName          *string `json:"agent_name" binding:"omitempty,max=50"`
	TransferType       *string `json:"transfer_type" binding:"omitempty,oneof=sale replace gift inheritance relocation"`
	VehicleRegistration interface{} `json:"vehicle_registration"`
	Documents          interface{} `json:"documents"`
	TransferFee        *float64 `json:"transfer_fee"`
	TaxFee             *float64 `json:"tax_fee"`
	OtherFee           *float64 `json:"other_fee"`
	Location           *string `json:"location" binding:"omitempty,max=255"`
	AppointmentDate    *time.Time `json:"appointment_date"`
	AppointmentTime    *string `json:"appointment_time" binding:"omitempty,max=32"`
	NewLicensePlate    *string `json:"new_license_plate" binding:"omitempty,max=32"`
	NewRegistrationCert *string `json:"new_registration_cert" binding:"omitempty,max=255"`
	Status             *int    `json:"status" binding:"omitempty,oneof=0 1 2 3 4 5 6 7"`
	Remark             *string `json:"remark"`
}

// TransferListRequest 过户列表请求
type TransferListRequest struct {
	CarID        uint   `form:"car_id" json:"car_id"`
	SellerID     uint   `form:"seller_id" json:"seller_id"`
	BuyerID      uint   `form:"buyer_id" json:"buyer_id"`
	AgentID      uint   `form:"agent_id" json:"agent_id"`
	TransferType string `form:"transfer_type" json:"transfer_type"`
	Status       *int   `form:"status" json:"status"`
	Keyword      string `form:"keyword" json:"keyword"`
	utils.Pagination
}

// TransferStatusUpdateRequest 过户状态更新请求
type TransferStatusUpdateRequest struct {
	Status              int    `json:"status" binding:"oneof=0 1 2 3 4 5 6 7"`
	NewLicensePlate     string `json:"new_license_plate" binding:"max=32"`
	NewRegistrationCert string `json:"new_registration_cert" binding:"max=255"`
	Remark              string `json:"remark"`
}
