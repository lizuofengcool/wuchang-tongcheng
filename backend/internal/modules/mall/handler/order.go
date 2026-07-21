// Package handler 同城商城 HTTP 处理层 - 订单
package handler

import (
	"net/http"

	"wuchang-tongcheng/internal/core/plugin"
	"wuchang-tongcheng/internal/core/response"
	"wuchang-tongcheng/internal/modules/mall/dto"
	"wuchang-tongcheng/internal/modules/mall/service"
	"wuchang-tongcheng/internal/pkg/utils"
)

// OrderHandler 订单 HTTP 处理器
type OrderHandler struct {
	svc service.OrderService
}

// NewOrderHandler 创建订单 Handler 实例
func NewOrderHandler(svc service.OrderService) *OrderHandler {
	return &OrderHandler{svc: svc}
}

// Create 创建订单
// POST /api/v1/mall/orders  （需登录）
func (h *OrderHandler) Create(ctx plugin.Context) {
	userID, username, phone, avatar := getUserProfile(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusOK, response.Unauthorized("请先登录"))
		return
	}
	var req dto.CreateOrderRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	regionID := getRegionID(ctx)
	ip := getClientIP(ctx)
	userAgent := ctx.GetHeader("User-Agent")
	info, err := h.svc.Create(regionID, userID, username, phone, avatar, ip, userAgent, &req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeMallOrderError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("下单成功", info))
}

// Cancel 买家取消订单
// PUT /api/v1/mall/orders/:id/cancel  （需登录）
func (h *OrderHandler) Cancel(ctx plugin.Context) {
	userID, ok := requireLogin(ctx)
	if ok {
		return
	}
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	var req dto.CancelOrderRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	if err := h.svc.Cancel(id, userID, &req); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeMallOrderError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("取消成功", nil))
}

// AdminClose 管理后台关闭订单
// PUT /api/v1/mall/admin/orders/:id/close  （需 mall:audit 权限）
func (h *OrderHandler) AdminClose(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	var req dto.AdminCloseOrderRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	if err := h.svc.AdminClose(id, &req); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeMallOrderError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("订单已关闭", nil))
}

// Ship 卖家发货
// PUT /api/v1/mall/orders/:id/ship  （需登录，卖家操作）
func (h *OrderHandler) Ship(ctx plugin.Context) {
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
	var req dto.ShipOrderRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	if err := h.svc.Ship(id, userID, &req); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeMallOrderError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("发货成功", nil))
}

// Confirm 买家确认收货
// PUT /api/v1/mall/orders/:id/confirm  （需登录）
func (h *OrderHandler) Confirm(ctx plugin.Context) {
	userID, ok := requireLogin(ctx)
	if ok {
		return
	}
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	if err := h.svc.Confirm(id, userID); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeMallOrderError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("确认收货成功", nil))
}

// Complete 完成订单
// PUT /api/v1/mall/admin/orders/:id/complete  （需 mall:audit 权限）
func (h *OrderHandler) Complete(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	if err := h.svc.Complete(id); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeMallOrderError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("订单已完成", nil))
}

// Delete 删除订单
// DELETE /api/v1/mall/admin/orders/:id  （需 mall:audit 权限）
func (h *OrderHandler) Delete(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	if err := h.svc.Delete(id); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeMallOrderNotFound, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("删除成功", nil))
}

// GetByID 订单详情
// GET /api/v1/mall/orders/:id  （需登录）
func (h *OrderHandler) GetByID(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	info, err := h.svc.GetByID(id)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeMallOrderNotFound, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(info))
}

// GetByOrderNo 按订单号查询
// GET /api/v1/mall/orders/by-no/:order_no  （需登录）
func (h *OrderHandler) GetByOrderNo(ctx plugin.Context) {
	orderNo := ctx.Param("order_no")
	if orderNo == "" {
		ctx.JSON(http.StatusOK, response.BadRequest("订单号不能为空"))
		return
	}
	info, err := h.svc.GetByOrderNo(orderNo)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeMallOrderNotFound, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(info))
}

// ListByUser 按买家列出订单
// GET /api/v1/mall/orders/mine  （需登录）
func (h *OrderHandler) ListByUser(ctx plugin.Context) {
	userID, ok := requireLogin(ctx)
	if ok {
		return
	}
	var req dto.OrderListRequest
	_ = ctx.Bind(&req)
	if req.Page == 0 {
		req.Page, req.PageSize = parsePagination(ctx)
	}
	pagination, list, err := h.svc.ListByUser(userID, &req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeMallOrderError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// ListByShop 按店铺列出订单
// GET /api/v1/mall/orders/by-shop/:shop_id  （需登录，卖家操作）
func (h *OrderHandler) ListByShop(ctx plugin.Context) {
	shopID, err := parseSubID(ctx, "shop_id")
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的店铺ID"))
		return
	}
	var req dto.OrderListRequest
	_ = ctx.Bind(&req)
	if req.Page == 0 {
		req.Page, req.PageSize = parsePagination(ctx)
	}
	pagination, list, err := h.svc.ListByShop(shopID, &req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeMallOrderError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// AdminList 管理后台订单列表
// GET /api/v1/mall/admin/orders  （需 mall:audit 权限）
func (h *OrderHandler) AdminList(ctx plugin.Context) {
	var req dto.AdminOrderListRequest
	_ = ctx.Bind(&req)
	if req.Page == 0 {
		req.Page, req.PageSize = parsePagination(ctx)
	}
	if req.RegionID == 0 {
		req.RegionID = getRegionID(ctx)
	}
	pagination, list, err := h.svc.AdminList(&req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeMallOrderError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// BatchUpdateStatus 批量更新订单状态
// POST /api/v1/mall/admin/orders/batch-status  （需 mall:audit 权限）
func (h *OrderHandler) BatchUpdateStatus(ctx plugin.Context) {
	var req dto.BatchStatusUpdateRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	if err := h.svc.BatchUpdateStatus(req.IDs, req.Status); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeMallOrderError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("批量更新成功", nil))
}

// CountByStatus 按状态统计订单数
// GET /api/v1/mall/orders/count-by-status  （需登录）
func (h *OrderHandler) CountByStatus(ctx plugin.Context) {
	userID, ok := requireLogin(ctx)
	if ok {
		return
	}
	status := parseQueryInt(ctx, "status", -1)
	count, err := h.svc.CountByStatus(userID, status)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeMallOrderError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(map[string]int64{"count": count}))
}

// Summary 订单汇总
// GET /api/v1/mall/orders/summary  （需登录）
func (h *OrderHandler) Summary(ctx plugin.Context) {
	userID, _, _, _ := getUserProfile(ctx)
	regionID := getRegionID(ctx)
	shopID := uint(parseQueryInt(ctx, "shop_id", 0))
	startDate := ctx.Query("start_date")
	endDate := ctx.Query("end_date")
	summary, err := h.svc.Summary(regionID, userID, shopID, startDate, endDate)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeMallOrderError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(summary))
}

// AutoClose 自动关闭超时未付款订单（定时任务）
// POST /api/v1/mall/admin/orders/auto-close  （需 mall:audit 权限）
func (h *OrderHandler) AutoClose(ctx plugin.Context) {
	count, err := h.svc.AutoClose()
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeMallOrderError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(map[string]int{"closed_count": count}))
}

// AutoConfirm 自动确认收货（定时任务）
// POST /api/v1/mall/admin/orders/auto-confirm  （需 mall:audit 权限）
func (h *OrderHandler) AutoConfirm(ctx plugin.Context) {
	count, err := h.svc.AutoConfirm()
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeMallOrderError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(map[string]int{"confirmed_count": count}))
}

// AutoReview 自动评价（定时任务）
// POST /api/v1/mall/admin/orders/auto-review  （需 mall:audit 权限）
func (h *OrderHandler) AutoReview(ctx plugin.Context) {
	count, err := h.svc.AutoReview()
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeMallOrderError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(map[string]int{"reviewed_count": count}))
}
