// Package handler 多租户分站 HTTP 处理层 - 分站
package handler

import (
	"net/http"

	"wuchang-tongcheng/internal/core/plugin"
	"wuchang-tongcheng/internal/core/response"
	"wuchang-tongcheng/internal/modules/tenant/dto"
	"wuchang-tongcheng/internal/modules/tenant/service"
	"wuchang-tongcheng/internal/pkg/utils"
)

// StationHandler 分站 HTTP 处理器
type StationHandler struct {
	svc service.StationService
}

// NewStationHandler 创建分站 Handler 实例
func NewStationHandler(svc service.StationService) *StationHandler {
	return &StationHandler{svc: svc}
}

// GetCurrent 获取当前分站（根据域名识别，公开）
// GET /api/v1/tenant/stations/current?domain=xxx
func (h *StationHandler) GetCurrent(ctx plugin.Context) {
	domain := ctx.Query("domain")
	if domain == "" {
		ctx.JSON(http.StatusOK, response.BadRequest("domain 参数不能为空"))
		return
	}
	info, err := h.svc.GetByDomain(domain)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(CodeTenantNotFound, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(info))
}

// List 分站列表（需登录）
// GET /api/v1/tenant/stations
func (h *StationHandler) List(ctx plugin.Context) {
	var req dto.StationListRequest
	_ = ctx.Bind(&req)
	if req.Page == 0 {
		req.Page, req.PageSize = parsePagination(ctx)
	}
	pagination, list, err := h.svc.List(&req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(CodeTenantError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// GetByID 分站详情（admin）
// GET /api/v1/tenant/admin/stations/:id
func (h *StationHandler) GetByID(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		invalidID(ctx)
		return
	}
	info, err := h.svc.GetByID(id)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(CodeTenantNotFound, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(info))
}

// Create 创建分站（admin）
// POST /api/v1/tenant/admin/stations
func (h *StationHandler) Create(ctx plugin.Context) {
	var req dto.CreateStationRequest
	if err := ctx.Bind(&req); err != nil {
		badRequest(ctx, "参数错误: "+err.Error())
		return
	}
	info, err := h.svc.Create(&req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(CodeTenantError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("分站创建成功", info))
}

// Update 更新分站（admin）
// PUT /api/v1/tenant/admin/stations/:id
func (h *StationHandler) Update(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		invalidID(ctx)
		return
	}
	var req dto.UpdateStationRequest
	if err := ctx.Bind(&req); err != nil {
		badRequest(ctx, "参数错误: "+err.Error())
		return
	}
	if err := h.svc.Update(id, &req); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(CodeTenantError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("更新成功", nil))
}

// Delete 删除分站（admin）
// DELETE /api/v1/tenant/admin/stations/:id
func (h *StationHandler) Delete(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		invalidID(ctx)
		return
	}
	if err := h.svc.Delete(id); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(CodeTenantError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("删除成功", nil))
}

// UpdateStatus 启停分站（admin）
// PUT /api/v1/tenant/admin/stations/:id/status
func (h *StationHandler) UpdateStatus(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		invalidID(ctx)
		return
	}
	var req dto.UpdateStationStatusRequest
	if err := ctx.Bind(&req); err != nil {
		badRequest(ctx, "参数错误: "+err.Error())
		return
	}
	if err := h.svc.UpdateStatus(id, req.Status); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(CodeTenantStatusInvalid, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("状态已更新", nil))
}

// CopyConfig 配置复制（admin）
// POST /api/v1/tenant/admin/stations/copy-config
func (h *StationHandler) CopyConfig(ctx plugin.Context) {
	var req dto.CopyConfigRequest
	if err := ctx.Bind(&req); err != nil {
		badRequest(ctx, "参数错误: "+err.Error())
		return
	}
	result, err := h.svc.CopyConfig(&req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(CodeTenantCopyError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("配置复制完成", result))
}
