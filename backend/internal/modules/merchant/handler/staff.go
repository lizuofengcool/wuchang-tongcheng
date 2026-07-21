// Package handler 商户中台 HTTP 处理层 - 员工
package handler

import (
	"net/http"

	"wuchang-tongcheng/internal/core/plugin"
	"wuchang-tongcheng/internal/core/response"
	"wuchang-tongcheng/internal/modules/merchant/dto"
	"wuchang-tongcheng/internal/modules/merchant/service"
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

// Create 添加员工（需登录）
// POST /api/v1/merchant/staff
func (h *StaffHandler) Create(ctx plugin.Context) {
	userID, _, _, _ := getUserProfile(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusOK, failUnauthorized())
		return
	}
	var req dto.CreateStaffRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, failParam("参数错误: "+err.Error()))
		return
	}
	info, err := h.svc.Create(&req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(CodeMerchantStaffExists, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("员工添加成功", info))
}

// Update 更新员工（需登录）
// PUT /api/v1/merchant/staff/:id
func (h *StaffHandler) Update(ctx plugin.Context) {
	userID, _, _, _ := getUserProfile(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusOK, failUnauthorized())
		return
	}
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	var req dto.UpdateStaffRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, failParam("参数错误: "+err.Error()))
		return
	}
	if err := h.svc.Update(id, userID, &req); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(CodeMerchantStaffNotFound, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("更新成功", nil))
}

// Delete 删除员工（需登录）
// DELETE /api/v1/merchant/staff/:id
func (h *StaffHandler) Delete(ctx plugin.Context) {
	userID, _, _, _ := getUserProfile(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusOK, failUnauthorized())
		return
	}
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	if err := h.svc.Delete(id, userID); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(CodeMerchantStaffNotFound, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("删除成功", nil))
}

// GetByID 员工详情（公开）
// GET /api/v1/merchant/staff/:id
func (h *StaffHandler) GetByID(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	info, err := h.svc.GetByID(id)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(CodeMerchantStaffNotFound, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(info))
}

// List 员工列表
// GET /api/v1/merchant/staff
func (h *StaffHandler) List(ctx plugin.Context) {
	var req dto.StaffListRequest
	_ = ctx.Bind(&req)
	if req.Page == 0 {
		req.Page, req.PageSize = parsePagination(ctx)
	}
	pagination, list, err := h.svc.List(&req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(CodeMerchantError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// AssignPermissions 权限分配（需登录）
// PUT /api/v1/merchant/staff/:id/permissions
func (h *StaffHandler) AssignPermissions(ctx plugin.Context) {
	userID, _, _, _ := getUserProfile(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusOK, failUnauthorized())
		return
	}
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	var req dto.StaffPermissionUpdateRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, failParam("参数错误: "+err.Error()))
		return
	}
	if err := h.svc.AssignPermissions(id, userID, req.Permissions); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(CodeMerchantStaffNotFound, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("权限分配成功", nil))
}

// SwitchRole 角色切换（需登录）
// PUT /api/v1/merchant/staff/:id/role
func (h *StaffHandler) SwitchRole(ctx plugin.Context) {
	userID, _, _, _ := getUserProfile(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusOK, failUnauthorized())
		return
	}
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	var req dto.StaffRoleSwitchRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, failParam("参数错误: "+err.Error()))
		return
	}
	if err := h.svc.SwitchRole(id, userID, req.Role); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(CodeMerchantStaffRoleInvalid, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("角色切换成功", nil))
}
