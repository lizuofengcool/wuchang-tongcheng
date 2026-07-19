// Package handler 支付财务中台精简版HTTP处理层
package handler

import (
	"net/http"
	"strconv"

	"wuchang-tongcheng/internal/core/middleware"
	"wuchang-tongcheng/internal/core/plugin"
	"wuchang-tongcheng/internal/core/response"
	"wuchang-tongcheng/internal/modules/pay/dto"
	"wuchang-tongcheng/internal/modules/pay/service"
)

// Handler 支付中台 HTTP 处理器
type Handler struct {
	svc service.PayService
}

// NewHandler 创建 Handler 实例
func NewHandler(svc service.PayService) *Handler {
	return &Handler{svc: svc}
}

// getUserID 从上下文获取登录用户ID
func getUserID(ctx plugin.Context) uint {
	if v, ok := ctx.Get(middleware.ContextUserID); ok {
		if id, ok := v.(uint); ok {
			return id
		}
	}
	return 0
}

// getRegionID 从上下文获取地区ID
func getRegionID(ctx plugin.Context) uint {
	if v, ok := ctx.Get(middleware.RegionIDKey); ok {
		if id, ok := v.(uint); ok {
			return id
		}
	}
	return middleware.DefaultRegionID
}

// parsePagination 解析分页
func parsePagination(ctx plugin.Context) (page, pageSize int) {
	page, _ = strconv.Atoi(ctx.DefaultQuery("page", "1"))
	pageSize, _ = strconv.Atoi(ctx.DefaultQuery("page_size", "10"))
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}
	return
}

// CreatePayment 创建支付订单
// POST /api/v1/pay/orders （需登录）
func (h *Handler) CreatePayment(ctx plugin.Context) {
	userID := getUserID(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusOK, response.Unauthorized("请先登录"))
		return
	}
	var req dto.CreatePaymentRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	info, err := h.svc.CreatePayment(getRegionID(ctx), userID, &req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(2801, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("创建订单成功", info))
}

// ConfirmPayment 确认支付（第三方回调）
// POST /api/v1/pay/orders/confirm
func (h *Handler) ConfirmPayment(ctx plugin.Context) {
	var req dto.ConfirmPaymentRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	if err := h.svc.ConfirmPayment(&req); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(2801, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("支付确认成功", nil))
}

// GetPayment 查询订单
// GET /api/v1/pay/orders/:order_no
func (h *Handler) GetPayment(ctx plugin.Context) {
	orderNo := ctx.Param("order_no")
	if orderNo == "" {
		ctx.JSON(http.StatusOK, response.BadRequest("订单号不能为空"))
		return
	}
	info, err := h.svc.GetPayment(orderNo)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(2802, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(info))
}

// ListPayments 我的订单列表
// GET /api/v1/pay/orders
func (h *Handler) ListPayments(ctx plugin.Context) {
	userID := getUserID(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusOK, response.Unauthorized("请先登录"))
		return
	}
	page, pageSize := parsePagination(ctx)
	list, total, err := h.svc.ListPayments(userID, page, pageSize)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(2801, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(response.NewPageResult(list, total, page, pageSize)))
}

// ConfirmEscrow 确认收货（放款担保账户）
// POST /api/v1/pay/escrow/confirm
func (h *Handler) ConfirmEscrow(ctx plugin.Context) {
	userID := getUserID(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusOK, response.Unauthorized("请先登录"))
		return
	}
	var req dto.ConfirmEscrowRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	if err := h.svc.ConfirmEscrow(req.OrderNo); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(2801, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("确认收货成功", nil))
}

// GetEscrow 查询担保账户
// GET /api/v1/pay/escrow/:order_no
func (h *Handler) GetEscrow(ctx plugin.Context) {
	orderNo := ctx.Param("order_no")
	info, err := h.svc.GetEscrow(orderNo)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(2802, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(info))
}

// Refund 申请退款
// POST /api/v1/pay/refunds
func (h *Handler) Refund(ctx plugin.Context) {
	var req dto.RefundRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	info, err := h.svc.Refund(getRegionID(ctx), &req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(2801, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("退款成功", info))
}

// GetRefund 查询退款单
// GET /api/v1/pay/refunds/:refund_no
func (h *Handler) GetRefund(ctx plugin.Context) {
	refundNo := ctx.Param("refund_no")
	info, err := h.svc.GetRefund(refundNo)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(2802, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(info))
}

// Withdraw 申请提现
// POST /api/v1/pay/withdrawals
func (h *Handler) Withdraw(ctx plugin.Context) {
	userID := getUserID(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusOK, response.Unauthorized("请先登录"))
		return
	}
	var req dto.WithdrawRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	info, err := h.svc.Withdraw(getRegionID(ctx), userID, &req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(2801, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("提现申请已提交", info))
}

// ListMyWithdrawals 我的提现列表
// GET /api/v1/pay/withdrawals
func (h *Handler) ListMyWithdrawals(ctx plugin.Context) {
	userID := getUserID(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusOK, response.Unauthorized("请先登录"))
		return
	}
	page, pageSize := parsePagination(ctx)
	list, total, err := h.svc.ListMyWithdrawals(userID, page, pageSize)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(2801, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(response.NewPageResult(list, total, page, pageSize)))
}

// HandleWithdrawal 处理提现申请（M 端）
// POST /api/v1/pay/admin/withdrawals/handle
func (h *Handler) HandleWithdrawal(ctx plugin.Context) {
	userID := getUserID(ctx)
	var req dto.WithdrawActionRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	if err := h.svc.HandleWithdrawal(&req, userID); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(2801, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("处理成功", nil))
}

// ListPendingWithdrawals 待审核提现列表（M 端）
// GET /api/v1/pay/admin/withdrawals/pending
func (h *Handler) ListPendingWithdrawals(ctx plugin.Context) {
	page, pageSize := parsePagination(ctx)
	list, total, err := h.svc.ListPendingWithdrawals(page, pageSize)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(2801, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(response.NewPageResult(list, total, page, pageSize)))
}

// Settle 触发结算（M 端）
// POST /api/v1/pay/admin/settlements
func (h *Handler) Settle(ctx plugin.Context) {
	var req dto.SettleRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	info, err := h.svc.Settle(getRegionID(ctx), &req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(2801, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("结算成功", info))
}

// ListSettlements 商家结算列表
// GET /api/v1/pay/settlements
func (h *Handler) ListSettlements(ctx plugin.Context) {
	merchantIDStr := ctx.Query("merchant_id")
	merchantID, _ := strconv.ParseUint(merchantIDStr, 10, 32)
	page, pageSize := parsePagination(ctx)
	list, total, err := h.svc.ListSettlements(uint(merchantID), page, pageSize)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(2801, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(response.NewPageResult(list, total, page, pageSize)))
}

// GetAccount 查询资金账户
// GET /api/v1/pay/account
func (h *Handler) GetAccount(ctx plugin.Context) {
	userID := getUserID(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusOK, response.Unauthorized("请先登录"))
		return
	}
	info, err := h.svc.GetAccount(userID)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(2802, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(info))
}
