// Package dto 多租户分站数据传输对象 - 配置
package dto

import "wuchang-tongcheng/internal/pkg/utils"

// ConfigInfo 配置详情响应
type ConfigInfo struct {
	ID          uint   `json:"id"`
	StationID   uint   `json:"station_id"`
	BizModule   string `json:"biz_module"`
	ConfigKey   string `json:"config_key"`
	ConfigValue string `json:"config_value"`
	UpdatedAt   string `json:"updated_at"`
}

// UpsertConfigRequest 新增/更新配置请求（按 station_id + biz_module + config_key 唯一）
type UpsertConfigRequest struct {
	StationID   uint   `json:"station_id" binding:"required"`
	BizModule   string `json:"biz_module" binding:"required,max=50"`
	ConfigKey   string `json:"config_key" binding:"required,max=100"`
	ConfigValue string `json:"config_value"`
}

// UpdateConfigRequest 更新配置请求
type UpdateConfigRequest struct {
	ConfigValue *string `json:"config_value"`
}

// ConfigListRequest 配置列表请求
type ConfigListRequest struct {
	StationID uint   `form:"station_id" json:"station_id"`
	BizModule string `form:"biz_module" json:"biz_module"`
	ConfigKey string `form:"config_key" json:"config_key"`
	Keyword   string `form:"keyword" json:"keyword"`
	utils.Pagination
}

// BatchGetConfigRequest 批量获取配置请求
type BatchGetConfigRequest struct {
	StationID  uint     `json:"station_id" binding:"required"`
	BizModule  string   `json:"biz_module" binding:"required,max=50"`
	ConfigKeys []string `json:"config_keys"`
}

// ConfigKeyValue 配置键值对
type ConfigKeyValue struct {
	ConfigKey   string `json:"config_key"`
	ConfigValue string `json:"config_value"`
}
