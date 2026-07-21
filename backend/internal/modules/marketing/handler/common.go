// Package handler 营销活动中台 HTTP 处理层 - 通用辅助函数
// 依据 v3.2.1 架构方案：4 维数据隔离（region_id + user_id）
package handler

import (
	"net/http"
	"strconv"
	"strings"

	"wuchang-tongcheng/internal/core/middleware"
	"wuchang-tongcheng/internal/core/plugin"
	"wuchang-tongcheng/internal/core/response"
)

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

// parseIntPtr 解析 *int 类型 query 参数
func parseIntPtr(ctx plugin.Context, key string) *int {
	v := ctx.Query(key)
	if v == "" {
		return nil
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return nil
	}
	return &n
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

// unauthorized 未登录响应
func unauthorized(ctx plugin.Context) {
	ctx.JSON(http.StatusOK, response.Unauthorized("请先登录"))
}

// badRequest 参数错误响应
func badRequest(ctx plugin.Context, msg string) {
	ctx.JSON(http.StatusOK, response.BadRequest(msg))
}
