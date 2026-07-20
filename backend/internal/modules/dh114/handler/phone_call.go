// Package handler 同城114 HTTP 处理层 - 电话拨打记录
// 依据 v3.2.1 架构方案：一键拨号核心
// 记录用户点击拨号/直接拨打的次数与设备信息
package handler

import (
	"net/http"

	"wuchang-tongcheng/internal/core/plugin"
	"wuchang-tongcheng/internal/core/response"
	"wuchang-tongcheng/internal/modules/dh114/dto"
	"wuchang-tongcheng/internal/modules/dh114/service"
	"wuchang-tongcheng/internal/pkg/utils"
)

// PhoneCallHandler 电话拨打记录 HTTP 处理器
type PhoneCallHandler struct {
	svc service.PhoneCallService
}

// NewPhoneCallHandler 创建电话拨打 Handler 实例
func NewPhoneCallHandler(svc service.PhoneCallService) *PhoneCallHandler {
	return &PhoneCallHandler{svc: svc}
}

// GetByID 电话拨打记录详情
// GET /api/v1/dh114/phone-calls/:id  （公开）
func (h *PhoneCallHandler) GetByID(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	info, err := h.svc.GetByID(id)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeDh114PhoneCallError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(info))
}

// List 电话拨打记录列表
// GET /api/v1/dh114/phone-calls  （公开）
func (h *PhoneCallHandler) List(ctx plugin.Context) {
	var req dto.PhoneCallListRequest
	_ = ctx.Bind(&req)
	if req.Page == 0 {
		req.Page, req.PageSize = parsePagination(ctx)
	}
	pagination, list, err := h.svc.List(&req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeDh114PhoneCallError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// ListByDh114 按商户列出电话拨打记录
// GET /api/v1/dh114/dh114/:id/phone-calls  （公开）
func (h *PhoneCallHandler) ListByDh114(ctx plugin.Context) {
	dh114ID, err := parseSubID(ctx, "id")
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的商户ID"))
		return
	}
	page, pageSize := parsePagination(ctx)
	pagination, list, err := h.svc.ListByDh114(dh114ID, page, pageSize)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeDh114PhoneCallError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// ListByCaller 按拨打者列出电话拨打记录
// GET /api/v1/dh114/phone-calls/mine  （需登录）
func (h *PhoneCallHandler) ListByCaller(ctx plugin.Context) {
	userID, _, _, _ := getUserProfile(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusOK, response.Unauthorized("请先登录"))
		return
	}
	page, pageSize := parsePagination(ctx)
	pagination, list, err := h.svc.ListByCaller(userID, page, pageSize)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeDh114PhoneCallError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// CountByDh114 商户电话拨打总数
// GET /api/v1/dh114/dh114/:id/phone-calls/count  （公开）
func (h *PhoneCallHandler) CountByDh114(ctx plugin.Context) {
	dh114ID, err := parseSubID(ctx, "id")
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的商户ID"))
		return
	}
	count, err := h.svc.CountByDh114(dh114ID)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeDh114PhoneCallError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(map[string]int64{"count": count}))
}

// CountTodayByDh114 商户今日电话拨打数
// GET /api/v1/dh114/dh114/:id/phone-calls/today-count  （公开）
func (h *PhoneCallHandler) CountTodayByDh114(ctx plugin.Context) {
	dh114ID, err := parseSubID(ctx, "id")
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的商户ID"))
		return
	}
	count, err := h.svc.CountTodayByDh114(dh114ID)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeDh114PhoneCallError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(map[string]int64{"count": count}))
}

// AdminList 电话拨打记录列表（M 端）
// GET /api/v1/dh114/admin/phone-calls  （需 dh114:audit 权限）
func (h *PhoneCallHandler) AdminList(ctx plugin.Context) {
	var req dto.PhoneCallAdminListRequest
	_ = ctx.Bind(&req)
	if req.Page == 0 {
		req.Page, req.PageSize = parsePagination(ctx)
	}
	pagination, list, err := h.svc.AdminList(&req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeDh114PhoneCallError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// Create 记录电话拨打（C 端单独入口，通常由 dh114.handler.RecordCall 调用）
// POST /api/v1/dh114/phone-calls  （需登录）
func (h *PhoneCallHandler) Create(ctx plugin.Context) {
	userID, _, _, _ := getUserProfile(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusOK, response.Unauthorized("请先登录"))
		return
	}
	var req dto.PhoneCallRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	regionID := getRegionID(ctx)
	ip := getClientIP(ctx)
	userAgent := ctx.GetHeader("User-Agent")
	info, err := h.svc.Create(regionID, userID, &req, ip, userAgent)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeDh114PhoneCallError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("已记录拨打", info))
}
