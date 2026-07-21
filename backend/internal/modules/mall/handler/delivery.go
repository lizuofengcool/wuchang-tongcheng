// Package handler 同城商城 HTTP 处理层 - 配送单
package handler

import (
	"net/http"

	"wuchang-tongcheng/internal/core/plugin"
	"wuchang-tongcheng/internal/core/response"
	"wuchang-tongcheng/internal/modules/mall/dto"
	"wuchang-tongcheng/internal/modules/mall/service"
	"wuchang-tongcheng/internal/pkg/utils"
)

// DeliveryHandler 配送单 HTTP 处理器
type DeliveryHandler struct {
	svc service.DeliveryService
}

// NewDeliveryHandler 创建配送单 Handler 实例
func NewDeliveryHandler(svc service.DeliveryService) *DeliveryHandler {
	return &DeliveryHandler{svc: svc}
}

// List 配送单列表（抢单大厅）
// GET /api/v1/mall/deliveries  （公开）
func (h *DeliveryHandler) List(ctx plugin.Context) {
	var req dto.DeliveryListRequest
	_ = ctx.Bind(&req)
	if req.Page == 0 {
		req.Page, req.PageSize = parsePagination(ctx)
	}
	regionID := getRegionID(ctx)
	pagination, list, err := h.svc.List(regionID, &req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeMallRiderError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// GetByID 配送单详情
// GET /api/v1/mall/deliveries/:id  （公开）
func (h *DeliveryHandler) GetByID(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	info, err := h.svc.GetByID(id)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeMallDeliveryNotFound, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(info))
}

// ListByRider 我的配送单
// GET /api/v1/mall/riders/deliveries  （需登录）
func (h *DeliveryHandler) ListByRider(ctx plugin.Context) {
	userID, ok := requireLogin(ctx)
	if ok {
		return
	}
	var req dto.DeliveryListRequest
	_ = ctx.Bind(&req)
	if req.Page == 0 {
		req.Page, req.PageSize = parsePagination(ctx)
	}
	pagination, list, err := h.svc.ListByRider(userID, &req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeMallDeliveryNotFound, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// Grab 骑手抢单
// POST /api/v1/mall/deliveries/:id/grab  （需登录）
func (h *DeliveryHandler) Grab(ctx plugin.Context) {
	userID, ok := requireLogin(ctx)
	if ok {
		return
	}
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	if err := h.svc.Grab(id, userID); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeMallDeliveryGrabbed, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("抢单成功", nil))
}

// ArriveShop 骑手到店
// PUT /api/v1/mall/deliveries/:id/arrive-shop  （需登录）
func (h *DeliveryHandler) ArriveShop(ctx plugin.Context) {
	userID, ok := requireLogin(ctx)
	if ok {
		return
	}
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	if err := h.svc.ArriveShop(id, userID); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeMallDeliveryStatusInvalid, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("已到店", nil))
}

// Pickup 骑手取货
// PUT /api/v1/mall/deliveries/:id/pickup  （需登录）
func (h *DeliveryHandler) Pickup(ctx plugin.Context) {
	userID, ok := requireLogin(ctx)
	if ok {
		return
	}
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	if err := h.svc.Pickup(id, userID); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeMallDeliveryStatusInvalid, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("已取货", nil))
}

// Deliver 开始配送
// PUT /api/v1/mall/deliveries/:id/deliver  （需登录）
func (h *DeliveryHandler) Deliver(ctx plugin.Context) {
	userID, ok := requireLogin(ctx)
	if ok {
		return
	}
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	if err := h.svc.Deliver(id, userID); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeMallDeliveryStatusInvalid, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("开始配送", nil))
}

// Complete 送达
// PUT /api/v1/mall/deliveries/:id/complete  （需登录）
func (h *DeliveryHandler) Complete(ctx plugin.Context) {
	userID, ok := requireLogin(ctx)
	if ok {
		return
	}
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	if err := h.svc.Complete(id, userID); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeMallDeliveryStatusInvalid, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("已送达", nil))
}

// Cancel 取消配送单
// PUT /api/v1/mall/deliveries/:id/cancel  （需登录）
func (h *DeliveryHandler) Cancel(ctx plugin.Context) {
	userID, ok := requireLogin(ctx)
	if ok {
		return
	}
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	var req dto.DeliveryCancelRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	if err := h.svc.Cancel(id, userID, &req); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeMallDeliveryStatusInvalid, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("已取消", nil))
}

// AdminList 管理后台配送单列表
// GET /api/v1/mall/admin/deliveries  （需 mall:audit 权限）
func (h *DeliveryHandler) AdminList(ctx plugin.Context) {
	var req dto.DeliveryListRequest
	_ = ctx.Bind(&req)
	if req.Page == 0 {
		req.Page, req.PageSize = parsePagination(ctx)
	}
	if req.RegionID == 0 {
		req.RegionID = getRegionID(ctx)
	}
	pagination, list, err := h.svc.AdminList(&req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeMallRiderError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}
