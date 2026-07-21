// Package handler 商户中台 HTTP 处理层 - 通用辅助函数与错误码
// 依据 dh114 handler/common.go（复制并裁剪）
//
// 错误码范围：5701-5730（与 mall 模块 5401-5430、dh114 模块 5301-5340 等并列）
package handler

import (
	"net/http"
	"strconv"
	"strings"

	"wuchang-tongcheng/internal/core/middleware"
	"wuchang-tongcheng/internal/core/plugin"
	"wuchang-tongcheng/internal/core/response"
)

// ===== 商户中台错误码（5701-5730） =====
const (
	CodeMerchantError              = 5701 // 商户中台通用错误
	CodeMerchantShopNotFound       = 5702 // 店铺不存在
	CodeMerchantShopExists         = 5703 // 店铺已存在/已被认领
	CodeMerchantShopNoPermission   = 5704 // 无权操作店铺
	CodeMerchantShopStatusInvalid  = 5705 // 店铺状态不允许此操作
	CodeMerchantStaffNotFound      = 5706 // 员工不存在
	CodeMerchantStaffExists        = 5707 // 员工已存在
	CodeMerchantStaffRoleInvalid   = 5708 // 员工角色无效
	CodeMerchantSettleNotFound     = 5709 // 结算单不存在
	CodeMerchantSettleExists       = 5710 // 该周期结算单已存在
	CodeMerchantSettleStatusInvalid = 5711 // 结算单状态不允许此操作
	CodeMerchantSettleAmountInvalid = 5712 // 结算金额无效
	CodeMerchantCategoryNotFound   = 5713 // 类目不存在
	CodeMerchantCategoryHasChildren = 5714 // 类目下有子类目
	CodeMerchantCategoryStatusInvalid = 5715 // 类目状态无效
	CodeMerchantVerifyNotFound     = 5716 // 认证记录不存在
	CodeMerchantVerifyAudited      = 5717 // 认证已审核
	CodeMerchantVerifyStatusInvalid = 5718 // 认证状态不允许此操作
	CodeMerchantApplyError         = 5719 // 入驻失败
	CodeMerchantClaimError         = 5720 // 认领失败
	CodeMerchantPermissionDenied   = 5721 // 权限不足
	CodeMerchantParamError         = 5722 // 参数错误
	CodeMerchantStaffNoPermission   = 5723 // 无权操作员工
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

// parseSubID 解析 URL 中的指定参数（如 :shop_id/:staff_id 等）
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
	return response.Fail(CodeMerchantParamError, msg)
}

// failUnauthorized 未登录响应
func failUnauthorized() *response.Response {
	return response.Fail(http.StatusUnauthorized, "请先登录")
}
