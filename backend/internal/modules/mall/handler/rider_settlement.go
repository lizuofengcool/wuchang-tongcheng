// Package handler 同城商城 HTTP 处理层 - 骑手结算
package handler

import (
	"net/http"

	"wuchang-tongcheng/internal/core/plugin"
	"wuchang-tongcheng/internal/core/response"
	"wuchang-tongcheng/internal/modules/mall/dto"
	"wuchang-tongcheng/internal/modules/mall/service"
	"wuchang-tongcheng/internal/pkg/utils"
)

// RiderSettlementHandler 骑手结算 HTTP 处理器
type RiderSettlementHandler struct {
	svc service.RiderSettlementService
}

// NewRiderSettlementHandler 创建骑手结算 Handler 实例
func NewRiderSettlementHandler(svc service.RiderSettlementService) *RiderSettlementHandler {
	return &RiderSettlementHandler{svc: svc}
}

// ListByUser 我的结算
// GET /api/v1/mall/riders/settlements  （需登录）
func (h *RiderSettlementHandler) ListByUser(ctx plugin.Context) {
	userID, ok := requireLogin(ctx)
	if ok {
		return
	}
	var req dto.RiderSettlementListRequest
	_ = ctx.Bind(&req)
	if req.Page == 0 {
		req.Page, req.PageSize = parsePagination(ctx)
	}
	pagination, list, err := h.svc.ListByUser(userID, &req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeMallRiderSettlementNotFound, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// Withdraw 提现申请
// POST /api/v1/mall/riders/settlements/withdraw  （需登录）
func (h *RiderSettlementHandler) Withdraw(ctx plugin.Context) {
	userID, ok := requireLogin(ctx)
	if ok {
		return
	}
	var req dto.RiderSettlementWithdrawRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	if err := h.svc.Withdraw(userID, &req); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeMallRiderSettlementError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("提现申请已提交", nil))
}

// AdminList 管理后台结算列表
// GET /api/v1/mall/admin/rider-settlements  （需 mall:audit 权限）
func (h *RiderSettlementHandler) AdminList(ctx plugin.Context) {
	var req dto.RiderSettlementListRequest
	_ = ctx.Bind(&req)
	if req.Page == 0 {
		req.Page, req.PageSize = parsePagination(ctx)
	}
	if req.RegionID == 0 {
		req.RegionID = getRegionID(ctx)
	}
	pagination, list, err := h.svc.List(&req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeMallRiderSettlementError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// AdminAudit 管理后台结算审核
// PUT /api/v1/mall/admin/rider-settlements/:id/audit  （需 mall:audit 权限）
func (h *RiderSettlementHandler) AdminAudit(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	var req dto.RiderSettlementAuditRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	if err := h.svc.Audit(id, &req); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeMallRiderSettlementError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("结算审核完成", nil))
}
