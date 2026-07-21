// Package dto 多租户分站数据传输对象 - 域名
package dto

import "wuchang-tongcheng/internal/pkg/utils"

// DomainInfo 域名详情响应
type DomainInfo struct {
	ID        uint   `json:"id"`
	StationID uint   `json:"station_id"`
	Domain    string `json:"domain"`
	IsPrimary bool   `json:"is_primary"`
	SSLStatus string `json:"ssl_status"`
	SSLText   string `json:"ssl_text"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

// CreateDomainRequest 绑定域名请求
type CreateDomainRequest struct {
	StationID uint   `json:"station_id" binding:"required"`
	Domain    string `json:"domain" binding:"required,max=200"`
	IsPrimary bool   `json:"is_primary"`
	SSLStatus string `json:"ssl_status" binding:"omitempty,oneof=none pending active failed"`
}

// UpdateDomainRequest 更新域名请求
type UpdateDomainRequest struct {
	SSLStatus *string `json:"ssl_status" binding:"omitempty,oneof=none pending active failed"`
}

// DomainListRequest 域名列表请求
type DomainListRequest struct {
	StationID uint   `form:"station_id" json:"station_id"`
	Domain    string `form:"domain" json:"domain"`
	SSLStatus string `form:"ssl_status" json:"ssl_status"`
	utils.Pagination
}

// SetPrimaryRequest 设置主域名请求
type SetPrimaryRequest struct {
	// 域名 ID 通过 URL :id 传入，无需 body 字段
}

// UpdateSSLStatusRequest 更新 SSL 状态请求
type UpdateSSLStatusRequest struct {
	SSLStatus string `json:"ssl_status" binding:"required,oneof=none pending active failed"`
}
