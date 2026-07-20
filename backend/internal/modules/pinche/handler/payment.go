// Package handler 同城拼车出行 HTTP 处理层 - 支付（含 ETC）
// 依据 v3.2.1 架构方案：对标哈啰出行/嘀嗒出行/滴滴顺风车
package handler

import (
	"net/http"

	"wuchang-tongcheng/internal/core/plugin"
	"wuchang-tongcheng/internal/core/response"
	"wuchang-tongcheng/internal/modules/pinche/dto"
	"wuchang-tongcheng/internal/modules/pinche/service"
	"wuchang-tongcheng/internal/pkg/utils"
)

// PaymentHandler 支付 HTTP 处理器
type PaymentHandler struct {
	service service.PaymentService
}

// NewPaymentHandler 创建 PaymentHandler 实例
func NewPaymentHandler(svc service.PaymentService) *PaymentHandler {
	return &PaymentHandler{service: svc}
}

// ===== C 端 =====

// Create 创建支付
// POST /api/v1/pinche/payments  （需登录）
func (h *PaymentHandler) Create(ctx plugin.Context) {
	userID, username, _, _ := getUserProfile(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusOK, response.Unauthorized("请先登录"))
		return
	}

	var req dto.CreatePaymentRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}

	regionID := getRegionID(ctx)
	info, err := h.service.Create(regionID, userID, username, &req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodePinchePaymentFailed, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("支付单已创建", info))
}

// Callback 支付回调（第三方支付平台回调）
// POST /api/v1/pinche/payments/callback  （公开，由支付平台调用）
func (h *PaymentHandler) Callback(ctx plugin.Context) {
	var req dto.PaymentCallbackRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	info, err := h.service.Callback(&req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodePinchePaymentFailed, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(info))
}

// GetByID 支付详情
// GET /api/v1/pinche/payments/:id  （需登录）
func (h *PaymentHandler) GetByID(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	info, err := h.service.GetByID(id)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodePinchePaymentNotFound, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(info))
}

// GetByPaymentNo 按支付单号查询
// GET /api/v1/pinche/payments/no/:payment_no  （需登录）
func (h *PaymentHandler) GetByPaymentNo(ctx plugin.Context) {
	paymentNo := ctx.Param("payment_no")
	if paymentNo == "" {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的支付单号"))
		return
	}
	info, err := h.service.GetByPaymentNo(paymentNo)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodePinchePaymentNotFound, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(info))
}

// ListByPayer 我的支付（付款方）
// GET /api/v1/pinche/payments/payer  （需登录）
func (h *PaymentHandler) ListByPayer(ctx plugin.Context) {
	userID, _, _, _ := getUserProfile(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusOK, response.Unauthorized("请先登录"))
		return
	}
	page, pageSize := parsePagination(ctx)
	pagination, list, err := h.service.ListByPayer(userID, page, pageSize)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodePincheError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// ListByPayee 我的收入（收款方）
// GET /api/v1/pinche/payments/payee  （需登录）
func (h *PaymentHandler) ListByPayee(ctx plugin.Context) {
	userID, _, _, _ := getUserProfile(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusOK, response.Unauthorized("请先登录"))
		return
	}
	page, pageSize := parsePagination(ctx)
	pagination, list, err := h.service.ListByPayee(userID, page, pageSize)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodePincheError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// ListByBooking 按预订查询支付
// GET /api/v1/pinche/bookings/:id/payments  （需登录）
func (h *PaymentHandler) ListByBooking(ctx plugin.Context) {
	bookingID, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的预订ID"))
		return
	}
	page, pageSize := parsePagination(ctx)
	pagination, list, err := h.service.ListByBooking(bookingID, page, pageSize)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodePincheError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// ETCSettlement ETC 结算
// POST /api/v1/pinche/payments/etc  （需登录）
func (h *PaymentHandler) ETCSettlement(ctx plugin.Context) {
	var req dto.ETCSettlementRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	info, err := h.service.ETCSettlement(&req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodePinchePaymentFailed, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("ETC 结算成功", info))
}

// ===== M 端管理 =====

// AdminList 管理后台支付列表
// GET /api/v1/pinche/admin/payments  （需 pinche:audit 权限）
func (h *PaymentHandler) AdminList(ctx plugin.Context) {
	var req dto.PaymentListRequest
	_ = ctx.Bind(&req)

	pagination, list, err := h.service.AdminList(&req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodePincheError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// UpdateStatus 管理后台更新支付状态
// PUT /api/v1/pinche/admin/payments/:id/status  （需 pinche:audit 权限）
func (h *PaymentHandler) UpdateStatus(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	var req dto.UpdateStatusRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	if err := h.service.UpdateStatus(id, req.Status); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodePincheStatusInvalid, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("状态更新成功", nil))
}
