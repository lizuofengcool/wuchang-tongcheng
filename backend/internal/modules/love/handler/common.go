// Package handler love 相亲交友 HTTP 处理层 - 通用辅助函数
// 依据 v3.2.1 架构方案：对标 Soul / 陌陌 / 探探 / 百合网
// 依据需求文档 1.10：4 维数据隔离（region_id 由 Region 中间件注入，user_id 由 JWT 解析）
package handler

import (
	"net/http"
	"strconv"
	"strings"

	"wuchang-tongcheng/internal/core/middleware"
	"wuchang-tongcheng/internal/core/plugin"
	"wuchang-tongcheng/internal/core/response"
)

// ===== 辅助函数（love 模块通用，供所有 handler 文件复用） =====

// getUserProfile 从上下文获取登录用户的完整信息（id/name/avatar）
// avatar 来自 JWT 冗余字段（见 jwt.GenerateTokenWithProfile），
// 用于发布时冗余存储到 love 表，避免每次发布都查 users 表。
func getUserProfile(ctx plugin.Context) (userID uint, username, avatar string) {
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
	if v, ok := ctx.Get(middleware.ContextUserAvatar); ok {
		if a, ok := v.(string); ok {
			avatar = a
		}
	}
	return
}

// getRegionID 从上下文获取地区ID（由 Region 中间件注入）
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

// parseSubID 解析 URL 中的子资源 ID（如 :message_id / :session_id）
func parseSubID(ctx plugin.Context, param string) (uint, error) {
	idStr := ctx.Param(param)
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		return 0, err
	}
	return uint(id), nil
}

// parsePagination 从 query 解析分页参数
func parsePagination(ctx plugin.Context) (page, pageSize int) {
	pageStr := ctx.Query("page")
	if pageStr == "" {
		pageStr = "1"
	}
	pageSizeStr := ctx.Query("page_size")
	if pageSizeStr == "" {
		pageSizeStr = "10"
	}
	page, _ = strconv.Atoi(pageStr)
	pageSize, _ = strconv.Atoi(pageSizeStr)
	return
}

// getClientIP 从上下文获取客户端 IP（用于浏览记录）
func getClientIP(ctx plugin.Context) string {
	// 优先从 X-Forwarded-For 取
	if ip := ctx.GetHeader("X-Forwarded-For"); ip != "" {
		// 取第一个
		if idx := strings.Index(ip, ","); idx > 0 {
			return strings.TrimSpace(ip[:idx])
		}
		return strings.TrimSpace(ip)
	}
	if ip := ctx.GetHeader("X-Real-IP"); ip != "" {
		return ip
	}
	// plugin.Context 接口未暴露 RemoteAddr，兜底返回空串
	return ""
}

// requireLogin 校验登录状态，返回 userID。若未登录则写入 401 响应并返回 0
func requireLogin(ctx plugin.Context) (uint, bool) {
	userID, _, _ := getUserProfile(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusOK, response.Unauthorized("请先登录"))
		return 0, false
	}
	return userID, true
}

// parseUint 解析字符串为 uint
func parseUint(s string) (uint, error) {
	v, err := strconv.ParseUint(s, 10, 32)
	if err != nil {
		return 0, err
	}
	return uint(v), nil
}
