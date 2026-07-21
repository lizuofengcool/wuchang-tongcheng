// Package handler 商户中台 HTTP 处理层 - 结算
package handler

import (
	"net/http"
	"strconv"

	"wuchang-tongcheng/internal/core/plugin"
	"wuchang-tongcheng/internal/core/response"
	"wuchang-tongcheng/internal/modules/merchant/dto"
	"wuchang-tongcheng/internal/modules/merchant/service"
	"wuchang-tongcheng/internal/pkg/utils"
)

// SettleHandler 结算 HTTP 处理器
type SettleHandler struct {
	svc service.SettleService
}

// NewSettleHandler 创建结算 Handler 实例
func NewSettleHandler(svc service.SettleService) *SettleHandler {
	return &SettleHandler{svc: svc}
}

// List 结算列表
// GET /api/v1/merchant/settles
func (h *SettleHandler) List(ctx plugin.Context) {
	var req dto.SettleListRequest
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

// GetByID 结算详情
// GET /api/v1/merchant/settles/:id
func (h *SettleHandler) GetByID(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	info, err := h.svc.GetByID(id)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(CodeMerchantSettleNotFound, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(info))
}

// ListByShop 按店铺查询结算
// GET /api/v1/merchant/shops/:id/settles
func (h *SettleHandler) ListByShop(ctx plugin.Context) {
	shopID, err := parseSubID(ctx, "id")
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的店铺ID"))
		return
	}
	page, pageSize := parsePagination(ctx)
	pagination, list, err := h.svc.ListByShop(shopID, page, pageSize)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(CodeMerchantError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// Generate 生成结算单（M 端）
// POST /api/v1/merchant/admin/settles
func (h *SettleHandler) Generate(ctx plugin.Context) {
	var req dto.SettleGenerateRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, failParam("参数错误: "+err.Error()))
		return
	}
	info, err := h.svc.Generate(&req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(CodeMerchantSettleExists, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("结算单生成成功", info))
}

// Withdraw 提现申请（M 端）
// PUT /api/v1/merchant/admin/settles/:id/withdraw
func (h *SettleHandler) Withdraw(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	if err := h.svc.Withdraw(id); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(CodeMerchantSettleStatusInvalid, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("提现申请已提交", nil))
}

// AuditWithdraw 提现审核（M 端）
// PUT /api/v1/merchant/admin/settles/:id/audit
func (h *SettleHandler) AuditWithdraw(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	var req dto.SettleAuditRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, failParam("参数错误: "+err.Error()))
		return
	}
	if err := h.svc.AuditWithdraw(id, &req); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(CodeMerchantSettleStatusInvalid, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("审核完成", nil))
}

// SummaryByShop 按店铺汇总
// GET /api/v1/merchant/admin/settles/summary-by-shop
func (h *SettleHandler) SummaryByShop(ctx plugin.Context) {
	shopIDStr := ctx.Query("shop_id")
	if shopIDStr == "" {
		ctx.JSON(http.StatusOK, failParam("shop_id 必填"))
		return
	}
	n, err := strconv.ParseUint(shopIDStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusOK, failParam("无效的 shop_id"))
		return
	}
	shopID := uint(n)
	summary, err := h.svc.SummaryByShop(shopID)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(CodeMerchantError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(summary))
}

// SummaryByPeriod 按周期汇总
// GET /api/v1/merchant/admin/settles/summary-by-period
func (h *SettleHandler) SummaryByPeriod(ctx plugin.Context) {
	period := ctx.Query("period")
	if period == "" {
		ctx.JSON(http.StatusOK, failParam("period 必填"))
		return
	}
	summary, err := h.svc.SummaryByPeriod(period)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(CodeMerchantError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(summary))
}
