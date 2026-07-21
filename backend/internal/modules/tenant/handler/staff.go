// Package handler 多租户分站 HTTP 处理层 - 员工
package handler

import (
	"net/http"

	"wuchang-tongcheng/internal/core/plugin"
	"wuchang-tongcheng/internal/core/response"
	"wuchang-tongcheng/internal/modules/tenant/dto"
	"wuchang-tongcheng/internal/modules/tenant/service"
	"wuchang-tongcheng/internal/pkg/utils"
)

// StaffHandler 员工 HTTP 处理器
type StaffHandler struct {
	svc service.StaffService
}

// NewStaffHandler 创建员工 Handler 实例
func NewStaffHandler(svc service.StaffService) *StaffHandler {
	return &StaffHandler{svc: svc}
}

// List 员工列表（admin）
// GET /api/v1/tenant/admin/staff
func (h *StaffHandler) List(ctx plugin.Context) {
	var req dto.StaffListRequest
	_ = ctx.Bind(&req)
	if req.Page == 0 {
		req.Page, req.PageSize = parsePagination(ctx)
	}
	pagination, list, err := h.svc.List(&req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(CodeTenantStaffError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// GetByID 员工详情（admin）
// GET /api/v1/tenant/admin/staff/:id
func (h *StaffHandler) GetByID(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		invalidID(ctx)
		return
	}
	info, err := h.svc.GetByID(id)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(CodeTenantStaffNotFound, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(info))
}

// Create 创建员工（admin）
// POST /api/v1/tenant/admin/staff
func (h *StaffHandler) Create(ctx plugin.Context) {
	var req dto.CreateStaffRequest
	if err := ctx.Bind(&req); err != nil {
		badRequest(ctx, "参数错误: "+err.Error())
		return
	}
	info, err := h.svc.Create(&req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(CodeTenantStaffError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("员工添加成功", info))
}

// Update 更新员工（角色/权限/状态）（admin）
// PUT /api/v1/tenant/admin/staff/:id
func (h *StaffHandler) Update(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		invalidID(ctx)
		return
	}
	var req dto.UpdateStaffRequest
	if err := ctx.Bind(&req); err != nil {
		badRequest(ctx, "参数错误: "+err.Error())
		return
	}
	if err := h.svc.Update(id, &req); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(CodeTenantStaffError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("更新成功", nil))
}

// Delete 删除员工（admin）
// DELETE /api/v1/tenant/admin/staff/:id
func (h *StaffHandler) Delete(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		invalidID(ctx)
		return
	}
	if err := h.svc.Delete(id); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(CodeTenantStaffError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("删除成功", nil))
}

// ListByStation 按分站查询员工（admin）
// GET /api/v1/tenant/admin/staff/by-station/:station_id
func (h *StaffHandler) ListByStation(ctx plugin.Context) {
	stationID, err := parseSubID(ctx, "station_id")
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的分站ID"))
		return
	}
	list, err := h.svc.ListByStation(stationID)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(CodeTenantStaffError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(list))
}
