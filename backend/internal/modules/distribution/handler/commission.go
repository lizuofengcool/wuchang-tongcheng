// Package handler 分销合伙人中台 HTTP 处理层 - 佣金记录
package handler

import (
	"net/http"

	"wuchang-tongcheng/internal/core/plugin"
	"wuchang-tongcheng/internal/core/response"
	"wuchang-tongcheng/internal/modules/distribution/dto"
	"wuchang-tongcheng/internal/modules/distribution/service"
	"wuchang-tongcheng/internal/pkg/utils"
)

// CommissionHandler 佣金处理器
type CommissionHandler struct {
	svc service.CommissionService
}

// NewCommissionHandler 创建 CommissionHandler 实例
func NewCommissionHandler(svc service.CommissionService) *CommissionHandler {
	return &CommissionHandler{svc: svc}
}

// List 佣金列表
// GET /api/v1/distribution/commissions
func (h *CommissionHandler) List(ctx plugin.Context) {
	var req dto.CommissionListRequest
	_ = ctx.Bind(&req)
	if req.Page == 0 {
		req.Page, req.PageSize = parsePagination(ctx)
	}
	pagination, list, err := h.svc.List(&req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(CodeDistributionError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// GetByID 详情
// GET /api/v1/distribution/commissions/:id
func (h *CommissionHandler) GetByID(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, failParam("无效的ID"))
		return
	}
	info, err := h.svc.GetByID(id)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(CodeDistributionCommissionNotFound, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(info))
}

// ListMine 我的佣金（需登录）
// GET /api/v1/distribution/commissions/mine
func (h *CommissionHandler) ListMine(ctx plugin.Context) {
	userID, _, _, _ := getUserProfile(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusOK, failUnauthorized())
		return
	}
	partnerID := parseUintPtr(ctx, "partner_id")
	if partnerID == nil || *partnerID == 0 {
		ctx.JSON(http.StatusOK, failParam("partner_id 不能为空"))
		return
	}
	page, pageSize := parsePagination(ctx)
	pagination, list, err := h.svc.ListByPartner(*partnerID, page, pageSize)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(CodeDistributionError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// Summary 佣金汇总
// GET /api/v1/distribution/commissions/summary
func (h *CommissionHandler) Summary(ctx plugin.Context) {
	var req dto.CommissionSummaryRequest
	_ = ctx.Bind(&req)
	if req.PartnerID == 0 {
		ctx.JSON(http.StatusOK, failParam("partner_id 不能为空"))
		return
	}
	info, err := h.svc.Summary(req.PartnerID)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(CodeDistributionError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(info))
}

// ===== 管理后台 =====

// AdminList 管理后台列表
// GET /api/v1/distribution/admin/commissions
func (h *CommissionHandler) AdminList(ctx plugin.Context) {
	var req dto.CommissionListRequest
	_ = ctx.Bind(&req)
	if req.Page == 0 {
		req.Page, req.PageSize = parsePagination(ctx)
	}
	pagination, list, err := h.svc.List(&req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(CodeDistributionError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// AdminCreate 管理后台创建佣金记录（订单结算时触发）
// POST /api/v1/distribution/admin/commissions
func (h *CommissionHandler) AdminCreate(ctx plugin.Context) {
	var req dto.CommissionCreateRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, failParam("参数错误: "+err.Error()))
		return
	}
	info, err := h.svc.Create(&req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(CodeDistributionError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("佣金记录已创建", info))
}

// AdminSettle 管理后台单条结算
// PUT /api/v1/distribution/admin/commissions/:id/settle
func (h *CommissionHandler) AdminSettle(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, failParam("无效的ID"))
		return
	}
	if err := h.svc.Settle(id); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(CodeDistributionCommissionStatusInvalid, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("已结算", nil))
}

// AdminBatchSettle 管理后台批量结算
// POST /api/v1/distribution/admin/commissions/batch-settle
func (h *CommissionHandler) AdminBatchSettle(ctx plugin.Context) {
	var req dto.CommissionSettleRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, failParam("参数错误: "+err.Error()))
		return
	}
	result, err := h.svc.BatchSettle(req.IDs)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(CodeDistributionError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(result))
}

// AdminCancel 管理后台取消佣金
// PUT /api/v1/distribution/admin/commissions/:id/cancel
func (h *CommissionHandler) AdminCancel(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, failParam("无效的ID"))
		return
	}
	if err := h.svc.Cancel(id); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(CodeDistributionCommissionStatusInvalid, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("已取消", nil))
}

// AdminSummary 管理后台汇总
// GET /api/v1/distribution/admin/commissions/summary
func (h *CommissionHandler) AdminSummary(ctx plugin.Context) {
	var req dto.CommissionSummaryRequest
	_ = ctx.Bind(&req)
	if req.PartnerID == 0 {
		ctx.JSON(http.StatusOK, failParam("partner_id 不能为空"))
		return
	}
	info, err := h.svc.Summary(req.PartnerID)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(CodeDistributionError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(info))
}
