// Package handler 分销合伙人中台 HTTP 处理层 - 通用辅助函数与错误码
// 依据 dh114/merchant handler/common.go（复制并裁剪）
//
// 错误码范围：5901-5930（与 merchant 5701-5730、mall 5401-5430、dh114 5301-5340 等并列）
package handler

import (
	"net/http"
	"strconv"
	"strings"

	"wuchang-tongcheng/internal/core/middleware"
	"wuchang-tongcheng/internal/core/plugin"
	"wuchang-tongcheng/internal/core/response"
)

// ===== 分销合伙人中台错误码（5901-5930） =====
const (
	CodeDistributionError              = 5901 // 分销中台通用错误
	CodeDistributionPartnerNotFound    = 5902 // 合伙人不存在
	CodeDistributionPartnerExists      = 5903 // 该用户已是合伙人
	CodeDistributionPartnerStatusInvalid = 5904 // 合伙人状态不允许此操作
	CodeDistributionPartnerLevelInvalid = 5905 // 合伙人等级无效
	CodeDistributionPartnerParentInvalid = 5906 // 上级合伙人无效
	CodeDistributionPartnerRateInvalid = 5907 // 佣金比例无效
	CodeDistributionChannelNotFound    = 5908 // 渠道不存在
	CodeDistributionChannelCodeExists  = 5909 // 渠道码已存在
	CodeDistributionChannelNoPermission = 5910 // 无权操作此渠道
	CodeDistributionCommissionNotFound  = 5911 // 佣金记录不存在
	CodeDistributionCommissionStatusInvalid = 5912 // 佣金记录状态不允许此操作
	CodeDistributionCommissionAmountInvalid = 5913 // 佣金金额无效
	CodeDistributionLevelNotFound       = 5914 // 等级不存在
	CodeDistributionLevelExists         = 5915 // 该等级已存在
	CodeDistributionLevelHasPartners    = 5916 // 该等级下存在合伙人
	CodeDistributionLevelStatusInvalid  = 5917 // 等级状态无效
	CodeDistributionWithdrawalNotFound  = 5918 // 提现记录不存在
	CodeDistributionWithdrawalStatusInvalid = 5919 // 提现状态不允许此操作
	CodeDistributionWithdrawalAmountInvalid = 5920 // 提现金额无效
	CodeDistributionWithdrawalInsufficient  = 5921 // 可提现余额不足
	CodeDistributionParamError          = 5922 // 参数错误
	CodeDistributionNoPermission        = 5923 // 权限不足
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

// parseSubID 解析 URL 中的指定参数（如 :partner_id/:order_id 等）
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

// parseUintPtr 解析 *uint 类型 query 参数
func parseUintPtr(ctx plugin.Context, key string) *uint {
	v := ctx.Query(key)
	if v == "" {
		return nil
	}
	n, err := strconv.ParseUint(v, 10, 32)
	if err != nil {
		return nil
	}
	u := uint(n)
	return &u
}

// parseFloat64Ptr 解析 *float64 类型 query 参数
func parseFloat64Ptr(ctx plugin.Context, key string) *float64 {
	v := ctx.Query(key)
	if v == "" {
		return nil
	}
	n, err := strconv.ParseFloat(v, 64)
	if err != nil {
		return nil
	}
	return &n
}

// parseBoolPtr 解析 *bool 类型 query 参数
func parseBoolPtr(ctx plugin.Context, key string) *bool {
	v := ctx.Query(key)
	if v == "" {
		return nil
	}
	b := strings.ToLower(v) == "true" || v == "1"
	return &b
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

// failParam 参数错误响应
func failParam(msg string) *response.Response {
	return response.Fail(CodeDistributionParamError, msg)
}

// failUnauthorized 未登录响应
func failUnauthorized() *response.Response {
	return response.Fail(http.StatusUnauthorized, "请先登录")
}
