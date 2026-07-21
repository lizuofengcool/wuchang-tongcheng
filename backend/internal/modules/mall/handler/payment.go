// Package handler 同城商城 HTTP 处理层 - 支付
package handler

import (
	"net/http"

	"wuchang-tongcheng/internal/core/plugin"
	"wuchang-tongcheng/internal/core/response"
	"wuchang-tongcheng/internal/modules/mall/dto"
	"wuchang-tongcheng/internal/modules/mall/service"
	"wuchang-tongcheng/internal/pkg/utils"
)

// PaymentHandler 支付 HTTP 处理器
type PaymentHandler struct {
	svc service.PaymentService
}

// NewPaymentHandler 创建支付 Handler 实例
func NewPaymentHandler(svc service.PaymentService) *PaymentHandler {
	return &PaymentHandler{svc: svc}
}

// Create 创建支付单
// POST /api/v1/mall/payments  （需登录）
func (h *PaymentHandler) Create(ctx plugin.Context) {
	userID, ok := requireLogin(ctx)
	if ok {
		return
	}
	var req dto.CreatePaymentRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	regionID := getRegionID(ctx)
	info, err := h.svc.Create(regionID, userID, &req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeMallPaymentError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("创建支付单成功", info))
}

// Callback 第三方支付回调
// POST /api/v1/mall/payments/callback  （公开，由第三方支付调用）
func (h *PaymentHandler) Callback(ctx plugin.Context) {
	var req dto.PaymentCallbackRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	if err := h.svc.HandleCallback(&req); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeMallPaymentError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("回调处理成功", nil))
}

// Close 关闭支付单
// PUT /api/v1/mall/payments/:id/close  （需登录）
func (h *PaymentHandler) Close(ctx plugin.Context) {
	userID, ok := requireLogin(ctx)
	if ok {
		return
	}
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	var req dto.ClosePaymentRequest
	_ = ctx.Bind(&req)
	_ = userID
	if err := h.svc.Close(id, &req); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeMallPaymentError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("支付单已关闭", nil))
}

// GetByID 支付详情
// GET /api/v1/mall/payments/:id  （需登录）
func (h *PaymentHandler) GetByID(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	info, err := h.svc.GetByID(id)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeMallPaymentNotFound, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(info))
}

// GetByPaymentNo 按支付单号查询
// GET /api/v1/mall/payments/by-no/:payment_no  （需登录）
func (h *PaymentHandler) GetByPaymentNo(ctx plugin.Context) {
	paymentNo := ctx.Param("payment_no")
	if paymentNo == "" {
		ctx.JSON(http.StatusOK, response.BadRequest("支付单号不能为空"))
		return
	}
	info, err := h.svc.GetByPaymentNo(paymentNo)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeMallPaymentNotFound, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(info))
}

// GetByOrderID 按订单 ID 查询支付单
// GET /api/v1/mall/payments/by-order/:order_id  （需登录）
func (h *PaymentHandler) GetByOrderID(ctx plugin.Context) {
	orderID, err := parseSubID(ctx, "order_id")
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的订单ID"))
		return
	}
	info, err := h.svc.GetByOrderID(orderID)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeMallPaymentNotFound, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(info))
}

// List 支付列表（管理后台）
// GET /api/v1/mall/admin/payments  （需 mall:audit 权限）
func (h *PaymentHandler) List(ctx plugin.Context) {
	var req dto.PaymentListRequest
	_ = ctx.Bind(&req)
	if req.Page == 0 {
		req.Page, req.PageSize = parsePagination(ctx)
	}
	pagination, list, err := h.svc.List(&req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeMallPaymentError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// ListByUser 按用户列出
// GET /api/v1/mall/payments/mine  （需登录）
func (h *PaymentHandler) ListByUser(ctx plugin.Context) {
	userID, ok := requireLogin(ctx)
	if ok {
		return
	}
	page, pageSize := parsePagination(ctx)
	pagination, list, err := h.svc.ListByUser(userID, page, pageSize)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeMallPaymentError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// ListByShop 按店铺列出
// GET /api/v1/mall/payments/by-shop/:shop_id  （需登录）
func (h *PaymentHandler) ListByShop(ctx plugin.Context) {
	shopID, err := parseSubID(ctx, "shop_id")
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的店铺ID"))
		return
	}
	page, pageSize := parsePagination(ctx)
	pagination, list, err := h.svc.ListByShop(shopID, page, pageSize)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeMallPaymentError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// Stats 支付统计
// GET /api/v1/mall/admin/payments/stats  （需 mall:audit 权限）
func (h *PaymentHandler) Stats(ctx plugin.Context) {
	regionID := getRegionID(ctx)
	shopID := uint(parseQueryInt(ctx, "shop_id", 0))
	startDate := ctx.Query("start_date")
	endDate := ctx.Query("end_date")
	stats, err := h.svc.Stats(regionID, shopID, startDate, endDate)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeMallPaymentError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(stats))
}
