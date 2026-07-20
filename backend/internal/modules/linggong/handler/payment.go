// Package handler 同城零工兼职 HTTP 处理层 - 薪资支付
package handler

import (
	"net/http"

	"wuchang-tongcheng/internal/core/plugin"
	"wuchang-tongcheng/internal/core/response"
	"wuchang-tongcheng/internal/modules/linggong/dto"
	"wuchang-tongcheng/internal/modules/linggong/service"
	"wuchang-tongcheng/internal/pkg/utils"
)

// PaymentHandler 薪资支付 HTTP 处理器
type PaymentHandler struct {
	service service.PaymentService
}

// NewPaymentHandler 创建 PaymentHandler 实例
func NewPaymentHandler(svc service.PaymentService) *PaymentHandler {
	return &PaymentHandler{service: svc}
}

// ===== C 端 =====

// Create 创建支付记录
// POST /api/v1/linggong/payments  （需登录）
func (h *PaymentHandler) Create(ctx plugin.Context) {
	userID, _, _, _ := getUserProfile(ctx)
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
	info, err := h.service.Create(regionID, userID, &req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeLinggongPaymentError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("支付记录已创建", info))
}

// Update 更新支付记录
// PUT /api/v1/linggong/payments/:id  （需登录）
func (h *PaymentHandler) Update(ctx plugin.Context) {
	userID, _, _, _ := getUserProfile(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusOK, response.Unauthorized("请先登录"))
		return
	}

	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}

	var req dto.UpdatePaymentRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}

	if err := h.service.Update(id, userID, &req); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeLinggongPaymentNotFound, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("更新成功", nil))
}

// GetByID 支付详情
// GET /api/v1/linggong/payments/:id  （公开）
func (h *PaymentHandler) GetByID(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	info, err := h.service.GetByID(id)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeLinggongPaymentNotFound, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(info))
}

// GetByPaymentNo 按支付编号查询
// GET /api/v1/linggong/payments/no/:payment_no  （公开）
func (h *PaymentHandler) GetByPaymentNo(ctx plugin.Context) {
	paymentNo := ctx.Param("payment_no")
	if paymentNo == "" {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的支付编号"))
		return
	}
	info, err := h.service.GetByPaymentNo(paymentNo)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeLinggongPaymentNotFound, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(info))
}

// List 支付列表
// GET /api/v1/linggong/payments  （公开）
func (h *PaymentHandler) List(ctx plugin.Context) {
	var req dto.PaymentListRequest
	_ = ctx.Bind(&req)

	regionID := getRegionID(ctx)
	pagination, list, err := h.service.List(regionID, &req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeLinggongError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// ListByLinggong 按岗位查询支付
// GET /api/v1/linggong/:id/payments  （公开）
func (h *PaymentHandler) ListByLinggong(ctx plugin.Context) {
	linggongID, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的岗位ID"))
		return
	}
	page, pageSize := parsePagination(ctx)
	pagination, list, err := h.service.ListByLinggong(linggongID, page, pageSize)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeLinggongError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// ListByEmployer 按雇主查询支付
// GET /api/v1/linggong/payments/employer/:id  （需登录）
func (h *PaymentHandler) ListByEmployer(ctx plugin.Context) {
	employerID, err := parseSubID(ctx, "id")
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的雇主ID"))
		return
	}
	page, pageSize := parsePagination(ctx)
	pagination, list, err := h.service.ListByEmployer(employerID, page, pageSize)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeLinggongError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// ListByWorker 按求职者查询支付
// GET /api/v1/linggong/payments/worker/:id  （需登录）
func (h *PaymentHandler) ListByWorker(ctx plugin.Context) {
	workerID, err := parseSubID(ctx, "id")
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的求职者ID"))
		return
	}
	page, pageSize := parsePagination(ctx)
	pagination, list, err := h.service.ListByWorker(workerID, page, pageSize)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeLinggongError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// UpdateStatus 更新支付状态
// PUT /api/v1/linggong/payments/:id/status  （需登录）
func (h *PaymentHandler) UpdateStatus(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	var req dto.PaymentStatusUpdateRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	if err := h.service.UpdateStatus(id, &req); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeLinggongPaymentError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("状态更新成功", nil))
}

// Settle 薪资结算
// POST /api/v1/linggong/payments/:id/settle  （需登录）
func (h *PaymentHandler) Settle(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	var req dto.PaymentSettleRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	if err := h.service.Settle(id, &req); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeLinggongPaymentError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("结算成功", nil))
}

// ===== M 端管理 =====

// AdminList 管理后台支付列表
// GET /api/v1/linggong/admin/payments  （需 linggong:audit 权限）
func (h *PaymentHandler) AdminList(ctx plugin.Context) {
	var req dto.PaymentAdminListRequest
	_ = ctx.Bind(&req)

	pagination, list, err := h.service.AdminList(&req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeLinggongError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}
