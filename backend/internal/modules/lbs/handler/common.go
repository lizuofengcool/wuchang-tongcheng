// Package handler LBS地图中台 HTTP 处理层 - 通用辅助函数
// 依据 v3.2.1 架构方案第 4.8 节：高德定位/附近检索/距离排序/POI/路线规划/地理围栏/分站区域隔离
package handler

import (
	"net/http"
	"strconv"
	"strings"

	"wuchang-tongcheng/internal/core/middleware"
	"wuchang-tongcheng/internal/core/plugin"
	"wuchang-tongcheng/internal/core/response"
)

// ===== 错误码常量（5501-5530） =====
// 注：不修改 backend/internal/pkg/utils/error_code.go，
//     直接在 handler 中以数字常量使用 response.Fail(code, msg)
const (
	CodeLBSError             = 5501 // LBS 通用错误
	CodeLBSPOINotFound        = 5502 // POI 不存在
	CodeLBSPOICreateError     = 5503 // POI 创建错误
	CodeLBSPOIUpdateError     = 5504 // POI 更新错误
	CodeLBSPOIDeleteError     = 5505 // POI 删除错误
	CodeLBSRegionNotFound     = 5506 // 区域不存在
	CodeLBSRegionCreateError  = 5507 // 区域创建错误
	CodeLBSGeofenceNotFound  = 5508 // 围栏不存在
	CodeLBSGeofenceCreateError = 5509 // 围栏创建错误
	CodeLBSRouteError         = 5510 // 路线规划错误
	CodeLBSDistanceError      = 5511 // 距离计算错误
	CodeLBSNoPermission       = 5512 // 无权操作
	CodeLBSStatusInvalid      = 5513 // 状态不允许此操作
	CodeLBSGeofenceTypeError  = 5514 // 围栏类型错误
	CodeLBSGeofencePointsError = 5515 // 围栏顶点格式错误
	CodeLBSRegionByLocation   = 5516 // 根据经纬度判断分站失败
	CodeLBSPOINearbyError     = 5517 // 附近检索错误
	CodeLBSPOIListError       = 5518 // POI 列表查询错误
	CodeLBSRegionListError   = 5519 // 区域列表查询错误
	CodeLBSGeofenceListError = 5520 // 围栏列表查询错误
	CodeLBSCheckPointError   = 5521 // 围栏判断错误
	CodeLBSParamInvalid      = 5522 // 参数错误
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

// parseSubID 解析 URL 中的指定参数
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

// parseBoolPtr 解析 *bool 类型 query 参数
func parseBoolPtr(ctx plugin.Context, key string) *bool {
	v := ctx.Query(key)
	if v == "" {
		return nil
	}
	b := strings.ToLower(v) == "true" || v == "1"
	return &b
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

// parseFloat64 从 query 解析 float64 参数（必填）
func parseFloat64(ctx plugin.Context, key string) (float64, error) {
	v := ctx.Query(key)
	if v == "" {
		return 0, errLBSParamRequired
	}
	return strconv.ParseFloat(v, 64)
}

// errLBSParamRequired 参数必填错误
var errLBSParamRequired = &lbsError{msg: "参数不能为空"}

// lbsError 内部错误类型
type lbsError struct {
	msg string
}

func (e *lbsError) Error() string { return e.msg }

// failByError 根据错误返回标准响应
func failByError(ctx plugin.Context, code int, err error) {
	msg := err.Error()
	if msg == "" {
		msg = "LBS 错误"
	}
	ctx.JSON(http.StatusOK, response.Fail(code, msg))
}

// failMsg 返回标准失败响应
func failMsg(ctx plugin.Context, code int, msg string) {
	if msg == "" {
		msg = "LBS 错误"
	}
	ctx.JSON(http.StatusOK, response.Fail(code, msg))
}
