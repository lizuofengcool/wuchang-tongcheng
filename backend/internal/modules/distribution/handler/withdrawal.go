// Package handler 分销合伙人中台 HTTP 处理层 - 提现
package handler

import (
	"net/http"

	"wuchang-tongcheng/internal/core/plugin"
	"wuchang-tongcheng/internal/core/response"
	"wuchang-tongcheng/internal/modules/distribution/dto"
	"wuchang-tongcheng/internal/modules/distribution/service"
	"wuchang-tongcheng/internal/pkg/utils"
)

// WithdrawalHandler 提现处理器
type WithdrawalHandler struct {
	svc service.WithdrawalService
}

// NewWithdrawalHandler 创建 WithdrawalHandler 实例
func NewWithdrawalHandler(svc service.WithdrawalService) *WithdrawalHandler {
	return &WithdrawalHandler{svc: svc}
}

// Apply 申请提现（C 端，需登录）
// POST /api/v1/distribution/withdrawals/apply
func (h *WithdrawalHandler) Apply(ctx plugin.Context) {
	userID, _, _, _ := getUserProfile(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusOK, failUnauthorized())
		return
	}
	var req dto.WithdrawalApplyRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, failParam("参数错误: "+err.Error()))
		return
	}
	// 从 body 扩展 partner_id
	type applyReq struct {
		dto.WithdrawalApplyRequest
		PartnerID uint `json:"partner_id" binding:"required"`
	}
	var ar applyReq
	_ = ctx.Bind(&ar)
	if ar.PartnerID == 0 {
		ctx.JSON(http.StatusOK, failParam("partner_id 不能为空"))
		return
	}
	info, err := h.svc.Apply(ar.PartnerID, &req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(CodeDistributionError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("提现申请已提交", info))
}

// ListMine 我的提现（需登录）
// GET /api/v1/distribution/withdrawals/mine
func (h *WithdrawalHandler) ListMine(ctx plugin.Context) {
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

// List 提现列表
// GET /api/v1/distribution/withdrawals
func (h *WithdrawalHandler) List(ctx plugin.Context) {
	var req dto.WithdrawalListRequest
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
// GET /api/v1/distribution/withdrawals/:id
func (h *WithdrawalHandler) GetByID(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, failParam("无效的ID"))
		return
	}
	info, err := h.svc.GetByID(id)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(CodeDistributionWithdrawalNotFound, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(info))
}

// ===== 管理后台 =====

// AdminList 管理后台列表
// GET /api/v1/distribution/admin/withdrawals
func (h *WithdrawalHandler) AdminList(ctx plugin.Context) {
	var req dto.WithdrawalListRequest
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

// AdminPending 待审核列表
// GET /api/v1/distribution/admin/withdrawals/pending
func (h *WithdrawalHandler) AdminPending(ctx plugin.Context) {
	page, pageSize := parsePagination(ctx)
	pagination, list, err := h.svc.ListPending(page, pageSize)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(CodeDistributionError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// AdminAudit 审核提现（通过/拒绝）
// PUT /api/v1/distribution/admin/withdrawals/:id/audit
func (h *WithdrawalHandler) AdminAudit(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, failParam("无效的ID"))
		return
	}
	var req dto.WithdrawalAuditRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, failParam("参数错误: "+err.Error()))
		return
	}
	if err := h.svc.Audit(id, &req); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(CodeDistributionWithdrawalStatusInvalid, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("审核完成", nil))
}

// AdminPay 打款确认
// PUT /api/v1/distribution/admin/withdrawals/:id/pay
func (h *WithdrawalHandler) AdminPay(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, failParam("无效的ID"))
		return
	}
	var req dto.WithdrawalPayRequest
	_ = ctx.Bind(&req)
	if err := h.svc.Pay(id, &req); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(CodeDistributionWithdrawalStatusInvalid, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("已确认打款", nil))
}

// AdminReject 拒绝提现
// PUT /api/v1/distribution/admin/withdrawals/:id/reject
func (h *WithdrawalHandler) AdminReject(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, failParam("无效的ID"))
		return
	}
	var body struct {
		Reason string `json:"reason" binding:"max=500"`
	}
	if err := ctx.Bind(&body); err != nil {
		ctx.JSON(http.StatusOK, failParam("参数错误: "+err.Error()))
		return
	}
	if err := h.svc.Reject(id, body.Reason); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(CodeDistributionWithdrawalStatusInvalid, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("已拒绝", nil))
}
