// Package handler 多租户分站 HTTP 处理层 - 公共辅助函数与错误码
// 依据架构设计第 4.10 节：多租户分站中台（tenant）
// 辅助函数沿用 dh114 风格（getUserProfile/getRegionID/parseID/parsePagination 等）
package handler

import (
	"net/http"
	"strconv"
	"strings"

	"wuchang-tongcheng/internal/core/middleware"
	"wuchang-tongcheng/internal/core/plugin"
	"wuchang-tongcheng/internal/core/response"
)

// ============================================================
// 错误码（5601-5630，tenant 模块专用，不修改 error_code.go）
// ============================================================
const (
	CodeTenantError         = 5601 // 分站通用错误
	CodeTenantNotFound      = 5602 // 分站不存在
	CodeTenantRegionExists  = 5603 // 该地区已存在分站
	CodeTenantDomainExists  = 5604 // 域名已被占用
	CodeTenantStatusInvalid = 5605 // 分站状态不允许此操作
	CodeTenantStaffError    = 5606 // 员工通用错误
	CodeTenantStaffNotFound = 5607 // 员工不存在
	CodeTenantStaffExists   = 5608 // 员工已存在
	CodeTenantConfigError   = 5609 // 配置通用错误
	CodeTenantConfigNotFound = 5610 // 配置不存在
	CodeTenantDomainError   = 5611 // 域名通用错误
	CodeTenantDomainNotFound = 5612 // 域名不存在
	CodeTenantCopyError     = 5613 // 配置复制错误
	CodeTenantSSLInvalid    = 5614 // SSL 状态无效
	CodeTenantDomainExists2 = 5615 // 域名已被绑定
)

// ============================================================
// 公共辅助函数（沿用 dh114 风格）
// ============================================================

// getUserProfile 从上下文获取登录用户信息（id/name/phone/avatar）
func getUserProfile(ctx plugin.Context) (userID uint, username, phone, avatar string) {
	if v, ok := ctx.Get(middleware.ContextUserID); ok {
		if id, ok := v.(uint); ok {
			userID = id
		}
	}
	if v, ok := ctx.Get(middleware.ContextUsername); ok {
		if name, ok := v.(string); ok {
			username = name
		}
	}
	if v, ok := ctx.Get(middleware.ContextUserPhone); ok {
		if p, ok := v.(string); ok {
			phone = p
		}
	}
	if v, ok := ctx.Get(middleware.ContextUserAvatar); ok {
		if a, ok := v.(string); ok {
			avatar = a
		}
	}
	return
}

// getRegionID 从上下文获取地区 ID（由 Region 中间件注入）
func getRegionID(ctx plugin.Context) uint {
	if v, ok := ctx.Get(middleware.RegionIDKey); ok {
		if id, ok := v.(uint); ok {
			return id
		}
	}
	return middleware.DefaultRegionID
}

// parseID 解析 URL 中的 :id 参数
func parseID(ctx plugin.Context) (uint, error) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		return 0, err
	}
	return uint(id), nil
}

// parseSubID 解析 URL 中的指定参数（如 :station_id 等）
func parseSubID(ctx plugin.Context, key string) (uint, error) {
	idStr := ctx.Param(key)
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		return 0, err
	}
	return uint(id), nil
}

// parsePagination 从 query 解析分页参数
func parsePagination(ctx plugin.Context) (page, pageSize int) {
	pageStr := ctx.DefaultQuery("page", "1")
	pageSizeStr := ctx.DefaultQuery("page_size", "10")
	page, _ = strconv.Atoi(pageStr)
	pageSize, _ = strconv.Atoi(pageSizeStr)
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}
	return
}

// getClientIP 获取客户端 IP
func getClientIP(ctx plugin.Context) string {
	if ip := ctx.GetHeader("X-Forwarded-For"); ip != "" {
		if idx := strings.Index(ip, ","); idx > 0 {
			return strings.TrimSpace(ip[:idx])
		}
		return ip
	}
	if ip := ctx.GetHeader("X-Real-IP"); ip != "" {
		return ip
	}
	return ""
}

// badRequest 参数错误响应
func badRequest(ctx plugin.Context, msg string) {
	ctx.JSON(http.StatusOK, response.BadRequest(msg))
}

// invalidID 无效 ID 响应
func invalidID(ctx plugin.Context) {
	ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
}
