// Package handler 订单 HTTP 处理层
// 依据 v3.2.1 架构方案：11 状态机订单（待支付→已支付→已发货→已完成/已取消）
package handler

import (
	"net/http"

	"wuchang-tongcheng/internal/core/plugin"
	"wuchang-tongcheng/internal/core/response"
	"wuchang-tongcheng/internal/modules/ershou/dto"
	"wuchang-tongcheng/internal/modules/ershou/service"
	"wuchang-tongcheng/internal/pkg/utils"
)

// OrderHandler 订单 HTTP 处理器
type OrderHandler struct {
	service service.OrderService
}

// NewOrderHandler 创建订单 Handler 实例
func NewOrderHandler(svc service.OrderService) *OrderHandler {
	return &OrderHandler{service: svc}
}

// Create 创建订单（C端买家下单）
// POST /api/v1/ershou/orders  （需登录）
func (h *OrderHandler) Create(ctx plugin.Context) {
	userID, _, _, _ := getUserProfile(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusOK, response.Unauthorized("请先登录"))
		return
	}
	var req dto.OrderCreateRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	regionID := getRegionID(ctx)
	resp, err := h.service.Create(regionID, userID, &req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeOrderError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("下单成功", resp))
}

// GetByID 订单详情（买卖双方可查）
// GET /api/v1/ershou/orders/:id  （需登录）
func (h *OrderHandler) GetByID(ctx plugin.Context) {
	userID, _, _, _ := getUserProfile(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusOK, response.Unauthorized("请先登录"))
		return
	}
	orderID, err := parseSubID(ctx, "id")
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的订单ID"))
		return
	}
	resp, err := h.service.GetByID(orderID, userID)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeOrderNotFound, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(resp))
}

// List 订单列表（按 buyer/seller/all 角色查询）
// GET /api/v1/ershou/orders  （需登录）
func (h *OrderHandler) List(ctx plugin.Context) {
	userID, _, _, _ := getUserProfile(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusOK, response.Unauthorized("请先登录"))
		return
	}
	var req dto.OrderQuery
	_ = ctx.Bind(&req)
	pagination, list, err := h.service.List(userID, req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeOrderError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// UpdateStatus 订单状态机变更（pay/ship/receive/cancel/complete）
// PUT /api/v1/ershou/orders/:id/status  （需登录）
func (h *OrderHandler) UpdateStatus(ctx plugin.Context) {
	userID, _, _, _ := getUserProfile(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusOK, response.Unauthorized("请先登录"))
		return
	}
	orderID, err := parseSubID(ctx, "id")
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的订单ID"))
		return
	}
	var req dto.OrderStatusUpdateRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	resp, err := h.service.UpdateStatus(orderID, userID, &req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeOrderError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("订单状态更新成功", resp))
}

// Pay 快捷支付接口（语义同 UpdateStatus action=pay）
// POST /api/v1/ershou/orders/:id/pay  （需登录 + 买家）
func (h *OrderHandler) Pay(ctx plugin.Context) {
	userID, _, _, _ := getUserProfile(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusOK, response.Unauthorized("请先登录"))
		return
	}
	orderID, err := parseSubID(ctx, "id")
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的订单ID"))
		return
	}
	resp, err := h.service.Pay(orderID, userID)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeOrderError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("支付成功", resp))
}

// Ship 快捷发货接口（语义同 UpdateStatus action=ship）
// POST /api/v1/ershou/orders/:id/ship  （需登录 + 卖家）
func (h *OrderHandler) Ship(ctx plugin.Context) {
	userID, _, _, _ := getUserProfile(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusOK, response.Unauthorized("请先登录"))
		return
	}
	orderID, err := parseSubID(ctx, "id")
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的订单ID"))
		return
	}
	resp, err := h.service.Ship(orderID, userID)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeOrderError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("发货成功", resp))
}

// Receive 快捷确认收货接口（语义同 UpdateStatus action=receive）
// POST /api/v1/ershou/orders/:id/receive  （需登录 + 买家）
func (h *OrderHandler) Receive(ctx plugin.Context) {
	userID, _, _, _ := getUserProfile(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusOK, response.Unauthorized("请先登录"))
		return
	}
	orderID, err := parseSubID(ctx, "id")
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的订单ID"))
		return
	}
	resp, err := h.service.Receive(orderID, userID)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeOrderError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("确认收货成功", resp))
}

// Cancel 快捷取消订单接口（语义同 UpdateStatus action=cancel）
// POST /api/v1/ershou/orders/:id/cancel  （需登录）
func (h *OrderHandler) Cancel(ctx plugin.Context) {
	userID, _, _, _ := getUserProfile(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusOK, response.Unauthorized("请先登录"))
		return
	}
	orderID, err := parseSubID(ctx, "id")
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的订单ID"))
		return
	}
	var req struct {
		Remark string `json:"remark"`
	}
	_ = ctx.Bind(&req)
	resp, err := h.service.Cancel(orderID, userID, req.Remark)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeOrderError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("订单已取消", resp))
}
