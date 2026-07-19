// Package module 模块注册表 DTO
package module

import (
	"encoding/json"
	"time"
)

// ModuleInfo 模块信息响应 DTO
type ModuleInfo struct {
	ID           uint     `json:"id"`
	Name         string   `json:"name"`
	DisplayName  string   `json:"display_name"`
	Category     string   `json:"category"`
	Description  string   `json:"description"`
	Version      string   `json:"version"`
	Dependencies []string `json:"dependencies"`
	Icon         string   `json:"icon"`
	Author       string   `json:"author"`
	Homepage     string   `json:"homepage"`
	Enabled      bool     `json:"enabled"`
	InstalledAt  string   `json:"installed_at"`
	UpdatedAt    string   `json:"updated_at"`
}

// UpdateModuleRequest 更新模块元信息请求 DTO
// 仅允许更新展示类字段，name/enabled 不通过此接口修改
type UpdateModuleRequest struct {
	DisplayName *string `json:"display_name"`
	Description *string `json:"description"`
	Icon        *string `json:"icon"`
	Author      *string `json:"author"`
	Homepage    *string `json:"homepage"`
}

// toInfo 将 Module 模型转换为 ModuleInfo DTO
// Dependencies 字段从 JSON 字符串反序列化为 []string
func toInfo(m *Module) ModuleInfo {
	deps := parseDependencies(m.Dependencies)
	return ModuleInfo{
		ID:           m.ID,
		Name:         m.Name,
		DisplayName:  m.DisplayName,
		Category:     m.Category,
		Description:  m.Description,
		Version:      m.Version,
		Dependencies: deps,
		Icon:         m.Icon,
		Author:       m.Author,
		Homepage:     m.Homepage,
		Enabled:      m.Enabled,
		InstalledAt:  formatTime(m.InstalledAt),
		UpdatedAt:    formatTime(m.UpdatedAt),
	}
}

// formatTime 格式化时间为 RFC3339 字符串，零值返回空串
func formatTime(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Format(time.RFC3339)
}

// parseDependencies 将 JSON 数组字符串解析为 []string
// 空串或解析失败返回 nil（不报错，降级为无依赖）
func parseDependencies(s string) []string {
	if s == "" {
		return nil
	}
	var deps []string
	if err := json.Unmarshal([]byte(s), &deps); err != nil {
		return nil
	}
	return deps
}

// toJSON 将 []string 序列化为 JSON 数组字符串
// 空切片返回 "[]"（modules 表 dependencies 字段为 JSONB 类型，空串会触发 "类型json的输入语法无效" 错误）
func toJSON(deps []string) string {
	if len(deps) == 0 {
		return "[]"
	}
	b, err := json.Marshal(deps)
	if err != nil {
		return "[]"
	}
	return string(b)
}
