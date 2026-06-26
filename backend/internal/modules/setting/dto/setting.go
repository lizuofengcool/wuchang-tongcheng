// Package dto 系统设置数据传输对象
package dto

// SettingInfo 配置信息
type SettingInfo struct {
	ID          uint   `json:"id"`
	Group       string `json:"group"`
	Key         string `json:"key"`
	Value       string `json:"value"`
	ValueType   string `json:"value_type"`
	Description string `json:"description"`
	Sort        int    `json:"sort"`
}

// CreateSettingRequest 创建配置请求
type CreateSettingRequest struct {
	Group       string `json:"group" binding:"required,max=50"`
	Key         string `json:"key" binding:"required,max=100"`
	Value       string `json:"value"`
	ValueType   string `json:"value_type" binding:"omitempty,oneof=string number bool json"`
	Description string `json:"description" binding:"max=255"`
	Sort        int    `json:"sort"`
}

// UpdateSettingRequest 更新配置请求
type UpdateSettingRequest struct {
	Value       string `json:"value"`
	Description string `json:"description" binding:"max=255"`
	Sort        int    `json:"sort"`
}

// BatchUpdateRequest 批量更新配置
type BatchUpdateRequest struct {
	Items []BatchItem `json:"items" binding:"required,min=1"`
}

// BatchItem 批量更新项
type BatchItem struct {
	Key   string `json:"key" binding:"required"`
	Value string `json:"value"`
}
