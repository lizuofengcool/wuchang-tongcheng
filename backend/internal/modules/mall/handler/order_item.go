// Package handler 同城商城 HTTP 处理层 - 订单明细
package handler

import (
	"net/http"

	"wuchang-tongcheng/internal/core/plugin"
	"wuchang-tongcheng/internal/core/response"
	"wuchang-tongcheng/internal/modules/mall/dto"
	"wuchang-tongcheng/internal/modules/mall/service"
	"wuchang-tongcheng/internal/pkg/utils"
)

// OrderItemHandler 订单明细 HTTP 处理器
type OrderItemHandler struct {
	svc service.OrderItemService
}

// NewOrderItemHandler 创建订单明细 Handler 实例
func NewOrderItemHandler(svc service.OrderItemService) *OrderItemHandler {
	return &OrderItemHandler{svc: svc}
}

// GetByID 订单明细详情
// GET /api/v1/mall/order-items/:id  （需登录）
func (h *OrderItemHandler) GetByID(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	info, err := h.svc.GetByID(id)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeMallOrderItemNotFound, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(info))
}

// ListByOrder 按订单列出明细
// GET /api/v1/mall/order-items/by-order/:order_id  （需登录）
func (h *OrderItemHandler) ListByOrder(ctx plugin.Context) {
	orderID, err := parseSubID(ctx, "order_id")
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的订单ID"))
		return
	}
	list, err := h.svc.ListByOrder(orderID)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeMallOrderItemError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(list))
}

// List 订单明细列表（管理后台）
// GET /api/v1/mall/admin/order-items  （需 mall:audit 权限）
func (h *OrderItemHandler) List(ctx plugin.Context) {
	var req dto.OrderItemListRequest
	_ = ctx.Bind(&req)
	if req.Page == 0 {
		req.Page, req.PageSize = parsePagination(ctx)
	}
	pagination, list, err := h.svc.List(&req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeMallOrderItemError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// ListByUser 按用户列出订单明细
// GET /api/v1/mall/order-items/mine  （需登录）
func (h *OrderItemHandler) ListByUser(ctx plugin.Context) {
	userID, ok := requireLogin(ctx)
	if ok {
		return
	}
	page, pageSize := parsePagination(ctx)
	pagination, list, err := h.svc.ListByUser(userID, page, pageSize)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeMallOrderItemError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// UpdateReviewStatus 更新评价状态（内部调用，一般不直接暴露）
// PUT /api/v1/mall/admin/order-items/:id/review-status  （需 mall:audit 权限）
func (h *OrderItemHandler) UpdateReviewStatus(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	var req struct {
		HasReview bool `json:"has_review"`
		ReviewID  uint `json:"review_id"`
	}
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	if err := h.svc.UpdateReviewStatus(id, req.HasReview, req.ReviewID); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeMallOrderItemError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("更新成功", nil))
}

// UpdateRefundStatus 更新退款状态（内部调用，一般不直接暴露）
// PUT /api/v1/mall/admin/order-items/:id/refund-status  （需 mall:audit 权限）
func (h *OrderItemHandler) UpdateRefundStatus(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	var req struct {
		RefundStatus int `json:"refund_status"`
		RefundID     uint `json:"refund_id"`
	}
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	if err := h.svc.UpdateRefundStatus(id, req.RefundStatus, req.RefundID); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeMallOrderItemError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("更新成功", nil))
}
