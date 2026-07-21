// Package handler 多租户分站 HTTP 处理层 - 域名
package handler

import (
	"net/http"

	"wuchang-tongcheng/internal/core/plugin"
	"wuchang-tongcheng/internal/core/response"
	"wuchang-tongcheng/internal/modules/tenant/dto"
	"wuchang-tongcheng/internal/modules/tenant/service"
	"wuchang-tongcheng/internal/pkg/utils"
)

// DomainHandler 域名 HTTP 处理器
type DomainHandler struct {
	svc service.DomainService
}

// NewDomainHandler 创建域名 Handler 实例
func NewDomainHandler(svc service.DomainService) *DomainHandler {
	return &DomainHandler{svc: svc}
}

// List 域名列表（admin）
// GET /api/v1/tenant/admin/domains
func (h *DomainHandler) List(ctx plugin.Context) {
	var req dto.DomainListRequest
	_ = ctx.Bind(&req)
	if req.Page == 0 {
		req.Page, req.PageSize = parsePagination(ctx)
	}
	pagination, list, err := h.svc.List(&req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(CodeTenantDomainError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// GetByID 域名详情（admin）
// GET /api/v1/tenant/admin/domains/:id
func (h *DomainHandler) GetByID(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		invalidID(ctx)
		return
	}
	info, err := h.svc.GetByID(id)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(CodeTenantDomainNotFound, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(info))
}

// Create 绑定域名（admin）
// POST /api/v1/tenant/admin/domains
func (h *DomainHandler) Create(ctx plugin.Context) {
	var req dto.CreateDomainRequest
	if err := ctx.Bind(&req); err != nil {
		badRequest(ctx, "参数错误: "+err.Error())
		return
	}
	info, err := h.svc.Create(&req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(CodeTenantDomainError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("域名绑定成功", info))
}

// Update 更新域名 SSL 状态（admin）
// PUT /api/v1/tenant/admin/domains/:id
func (h *DomainHandler) Update(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		invalidID(ctx)
		return
	}
	var req dto.UpdateDomainRequest
	if err := ctx.Bind(&req); err != nil {
		badRequest(ctx, "参数错误: "+err.Error())
		return
	}
	if err := h.svc.Update(id, &req); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(CodeTenantDomainError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("更新成功", nil))
}

// Delete 删除域名绑定（admin）
// DELETE /api/v1/tenant/admin/domains/:id
func (h *DomainHandler) Delete(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		invalidID(ctx)
		return
	}
	if err := h.svc.Delete(id); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(CodeTenantDomainError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("删除成功", nil))
}

// SetPrimary 设置主域名（admin）
// PUT /api/v1/tenant/admin/domains/:id/primary
func (h *DomainHandler) SetPrimary(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		invalidID(ctx)
		return
	}
	if err := h.svc.SetPrimary(id); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(CodeTenantDomainError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("已设为主域名", nil))
}

// UpdateSSL 更新 SSL 状态（admin）
// PUT /api/v1/tenant/admin/domains/:id/ssl
func (h *DomainHandler) UpdateSSL(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		invalidID(ctx)
		return
	}
	var req dto.UpdateSSLStatusRequest
	if err := ctx.Bind(&req); err != nil {
		badRequest(ctx, "参数错误: "+err.Error())
		return
	}
	if err := h.svc.UpdateSSLStatus(id, req.SSLStatus); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(CodeTenantSSLInvalid, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("SSL 状态已更新", nil))
}

// ListByStation 按分站查询域名（admin）
// GET /api/v1/tenant/admin/domains/by-station/:station_id
func (h *DomainHandler) ListByStation(ctx plugin.Context) {
	stationID, err := parseSubID(ctx, "station_id")
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的分站ID"))
		return
	}
	list, err := h.svc.ListByStation(stationID)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(CodeTenantDomainError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(list))
}
