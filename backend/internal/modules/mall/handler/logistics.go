// Package handler 同城商城 HTTP 处理层 - 物流
package handler

import (
	"net/http"

	"wuchang-tongcheng/internal/core/plugin"
	"wuchang-tongcheng/internal/core/response"
	"wuchang-tongcheng/internal/modules/mall/dto"
	"wuchang-tongcheng/internal/modules/mall/service"
	"wuchang-tongcheng/internal/pkg/utils"
)

// LogisticsHandler 物流 HTTP 处理器
type LogisticsHandler struct {
	svc service.LogisticsService
}

// NewLogisticsHandler 创建物流 Handler 实例
func NewLogisticsHandler(svc service.LogisticsService) *LogisticsHandler {
	return &LogisticsHandler{svc: svc}
}

// Create 创建物流记录
// POST /api/v1/mall/admin/logistics  （需 mall:audit 权限，卖家发货时调用）
func (h *LogisticsHandler) Create(ctx plugin.Context) {
	userID, _, _, _ := getUserProfile(ctx)
	var req dto.CreateLogisticsRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	regionID := getRegionID(ctx)
	info, err := h.svc.Create(regionID, userID, &req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeMallLogisticsError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("创建成功", info))
}

// Update 更新物流
// PUT /api/v1/mall/admin/logistics/:id  （需 mall:audit 权限）
func (h *LogisticsHandler) Update(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	var req dto.UpdateLogisticsRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	if err := h.svc.Update(id, &req); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeMallLogisticsError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("更新成功", nil))
}

// Delete 删除物流记录
// DELETE /api/v1/mall/admin/logistics/:id  （需 mall:audit 权限）
func (h *LogisticsHandler) Delete(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	if err := h.svc.Delete(id); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeMallLogisticsNotFound, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("删除成功", nil))
}

// GetByID 物流详情
// GET /api/v1/mall/logistics/:id  （需登录）
func (h *LogisticsHandler) GetByID(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	info, err := h.svc.GetByID(id)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeMallLogisticsNotFound, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(info))
}

// GetByOrderID 按订单 ID 查询物流
// GET /api/v1/mall/logistics/by-order/:order_id  （需登录）
func (h *LogisticsHandler) GetByOrderID(ctx plugin.Context) {
	orderID, err := parseSubID(ctx, "order_id")
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的订单ID"))
		return
	}
	info, err := h.svc.GetByOrderID(orderID)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeMallLogisticsNotFound, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(info))
}

// GetByTrackingNo 按物流单号查询
// GET /api/v1/mall/logistics/by-tracking-no  （公开）
func (h *LogisticsHandler) GetByTrackingNo(ctx plugin.Context) {
	trackingNo := ctx.Query("tracking_no")
	if trackingNo == "" {
		ctx.JSON(http.StatusOK, response.BadRequest("物流单号不能为空"))
		return
	}
	info, err := h.svc.GetByTrackingNo(trackingNo)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeMallLogisticsNotFound, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(info))
}

// List 物流列表（管理后台）
// GET /api/v1/mall/admin/logistics  （需 mall:audit 权限）
func (h *LogisticsHandler) List(ctx plugin.Context) {
	var req dto.LogisticsListRequest
	_ = ctx.Bind(&req)
	if req.Page == 0 {
		req.Page, req.PageSize = parsePagination(ctx)
	}
	pagination, list, err := h.svc.List(&req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeMallLogisticsError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// ListByUser 按用户列出
// GET /api/v1/mall/logistics/mine  （需登录）
func (h *LogisticsHandler) ListByUser(ctx plugin.Context) {
	userID, ok := requireLogin(ctx)
	if ok {
		return
	}
	page, pageSize := parsePagination(ctx)
	pagination, list, err := h.svc.ListByUser(userID, page, pageSize)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeMallLogisticsError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// ListByShop 按店铺列出
// GET /api/v1/mall/logistics/by-shop/:shop_id  （需登录）
func (h *LogisticsHandler) ListByShop(ctx plugin.Context) {
	shopID, err := parseSubID(ctx, "shop_id")
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的店铺ID"))
		return
	}
	page, pageSize := parsePagination(ctx)
	pagination, list, err := h.svc.ListByShop(shopID, page, pageSize)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeMallLogisticsError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// UpdateStatus 更新物流状态
// PUT /api/v1/mall/admin/logistics/:id/status  （需 mall:audit 权限）
func (h *LogisticsHandler) UpdateStatus(ctx plugin.Context) {
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
	if err := h.svc.UpdateStatus(id, req.Status); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeMallLogisticsError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("状态已更新", nil))
}

// Callback 物流状态回调（第三方物流回调）
// POST /api/v1/mall/logistics/callback  （公开，由第三方物流调用）
func (h *LogisticsHandler) Callback(ctx plugin.Context) {
	var req dto.UpdateLogisticsStatusRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	if err := h.svc.UpdateStatusByCallback(&req); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeMallLogisticsError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("回调处理成功", nil))
}
