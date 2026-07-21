// Package handler 同城商城 HTTP 处理层 - 骑手
package handler

import (
	"net/http"

	"wuchang-tongcheng/internal/core/plugin"
	"wuchang-tongcheng/internal/core/response"
	"wuchang-tongcheng/internal/modules/mall/dto"
	"wuchang-tongcheng/internal/modules/mall/service"
	"wuchang-tongcheng/internal/pkg/utils"
)

// RiderHandler 骑手 HTTP 处理器
type RiderHandler struct {
	svc service.RiderService
}

// NewRiderHandler 创建骑手 Handler 实例
func NewRiderHandler(svc service.RiderService) *RiderHandler {
	return &RiderHandler{svc: svc}
}

// Apply 骑手申请
// POST /api/v1/mall/riders/apply  （需登录）
func (h *RiderHandler) Apply(ctx plugin.Context) {
	userID, ok := requireLogin(ctx)
	if ok {
		return
	}
	var req dto.RiderApplyRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	regionID := getRegionID(ctx)
	info, err := h.svc.Apply(regionID, userID, &req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeMallRiderError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("骑手申请已提交", info))
}

// GetByUserID 我的骑手资料
// GET /api/v1/mall/riders/mine  （需登录）
func (h *RiderHandler) GetByUserID(ctx plugin.Context) {
	userID, ok := requireLogin(ctx)
	if ok {
		return
	}
	info, err := h.svc.GetByUserID(userID)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeMallRiderNotFound, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(info))
}

// Update 更新骑手资料
// PUT /api/v1/mall/riders/:id  （需登录）
func (h *RiderHandler) Update(ctx plugin.Context) {
	userID, ok := requireLogin(ctx)
	if ok {
		return
	}
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	var req dto.RiderUpdateRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	if err := h.svc.Update(id, userID, &req); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeMallRiderError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("更新成功", nil))
}

// Online 骑手上线
// PUT /api/v1/mall/riders/:id/online  （需登录）
func (h *RiderHandler) Online(ctx plugin.Context) {
	userID, ok := requireLogin(ctx)
	if ok {
		return
	}
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	if err := h.svc.Online(id, userID); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeMallRiderError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("已上线", nil))
}

// Offline 骑手下线
// PUT /api/v1/mall/riders/:id/offline  （需登录）
func (h *RiderHandler) Offline(ctx plugin.Context) {
	userID, ok := requireLogin(ctx)
	if ok {
		return
	}
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	if err := h.svc.Offline(id, userID); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeMallRiderError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("已下线", nil))
}

// Earnings 我的收益
// GET /api/v1/mall/riders/earnings  （需登录）
func (h *RiderHandler) Earnings(ctx plugin.Context) {
	userID, ok := requireLogin(ctx)
	if ok {
		return
	}
	resp, err := h.svc.Earnings(userID)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeMallRiderError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(resp))
}

// AdminList 管理后台骑手列表
// GET /api/v1/mall/admin/riders  （需 mall:audit 权限）
func (h *RiderHandler) AdminList(ctx plugin.Context) {
	var req dto.RiderListRequest
	_ = ctx.Bind(&req)
	if req.Page == 0 {
		req.Page, req.PageSize = parsePagination(ctx)
	}
	if req.RegionID == 0 {
		req.RegionID = getRegionID(ctx)
	}
	pagination, list, err := h.svc.AdminList(&req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeMallRiderError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// AdminGetByID 管理后台骑手详情
// GET /api/v1/mall/admin/riders/:id  （需 mall:audit 权限）
func (h *RiderHandler) AdminGetByID(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	info, err := h.svc.GetByID(id)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeMallRiderNotFound, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(info))
}

// AdminAudit 管理后台骑手审核
// PUT /api/v1/mall/admin/riders/:id/audit  （需 mall:audit 权限）
func (h *RiderHandler) AdminAudit(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	var req dto.RiderAuditRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	if err := h.svc.Audit(id, &req); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeMallRiderError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("审核完成", nil))
}

// AdminUpdateStatus 管理后台骑手状态更新（冻结/解冻）
// PUT /api/v1/mall/admin/riders/:id/status  （需 mall:audit 权限）
func (h *RiderHandler) AdminUpdateStatus(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	var req dto.RiderStatusUpdateRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	if err := h.svc.UpdateStatus(id, &req); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeMallRiderError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("状态已更新", nil))
}
