// Package handler 同城商城 HTTP 处理层 - 退款
package handler

import (
	"net/http"

	"wuchang-tongcheng/internal/core/plugin"
	"wuchang-tongcheng/internal/core/response"
	"wuchang-tongcheng/internal/modules/mall/dto"
	"wuchang-tongcheng/internal/modules/mall/service"
	"wuchang-tongcheng/internal/pkg/utils"
)

// RefundHandler 退款 HTTP 处理器
type RefundHandler struct {
	svc service.RefundService
}

// NewRefundHandler 创建退款 Handler 实例
func NewRefundHandler(svc service.RefundService) *RefundHandler {
	return &RefundHandler{svc: svc}
}

// Create 创建退款单
// POST /api/v1/mall/refunds  （需登录）
func (h *RefundHandler) Create(ctx plugin.Context) {
	userID, ok := requireLogin(ctx)
	if ok {
		return
	}
	var req dto.CreateRefundRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	regionID := getRegionID(ctx)
	info, err := h.svc.Create(regionID, userID, &req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeMallRefundError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("退款申请已提交", info))
}

// SellerProcess 卖家处理退款
// PUT /api/v1/mall/refunds/:id/seller-process  （需登录，卖家操作）
func (h *RefundHandler) SellerProcess(ctx plugin.Context) {
	userID, ok := requireLogin(ctx)
	if ok {
		return
	}
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	var req dto.SellerProcessRefundRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	if err := h.svc.SellerProcess(id, userID, &req); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeMallRefundError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("处理成功", nil))
}

// AdminProcess 管理后台处理退款
// PUT /api/v1/mall/admin/refunds/:id/process  （需 mall:audit 权限）
func (h *RefundHandler) AdminProcess(ctx plugin.Context) {
	adminID, _, _, _ := getUserProfile(ctx)
	adminName, _ := ctx.Get("username")
	name, _ := adminName.(string)
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	var req dto.AdminProcessRefundRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	if err := h.svc.AdminProcess(id, adminID, name, &req); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeMallRefundError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("处理成功", nil))
}

// Ship 买家退货物流
// PUT /api/v1/mall/refunds/:id/ship  （需登录）
func (h *RefundHandler) Ship(ctx plugin.Context) {
	userID, ok := requireLogin(ctx)
	if ok {
		return
	}
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	var req dto.RefundShipRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	if err := h.svc.Ship(id, userID, &req); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeMallRefundError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("退货物流已填写", nil))
}

// GetByID 退款详情
// GET /api/v1/mall/refunds/:id  （需登录）
func (h *RefundHandler) GetByID(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	info, err := h.svc.GetByID(id)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeMallRefundNotFound, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(info))
}

// GetByRefundNo 按退款单号查询
// GET /api/v1/mall/refunds/by-no/:refund_no  （需登录）
func (h *RefundHandler) GetByRefundNo(ctx plugin.Context) {
	refundNo := ctx.Param("refund_no")
	if refundNo == "" {
		ctx.JSON(http.StatusOK, response.BadRequest("退款单号不能为空"))
		return
	}
	info, err := h.svc.GetByRefundNo(refundNo)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeMallRefundNotFound, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(info))
}

// List 退款列表（管理后台）
// GET /api/v1/mall/admin/refunds  （需 mall:audit 权限）
func (h *RefundHandler) List(ctx plugin.Context) {
	var req dto.RefundListRequest
	_ = ctx.Bind(&req)
	if req.Page == 0 {
		req.Page, req.PageSize = parsePagination(ctx)
	}
	pagination, list, err := h.svc.List(&req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeMallRefundError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// ListByUser 按用户列出
// GET /api/v1/mall/refunds/mine  （需登录）
func (h *RefundHandler) ListByUser(ctx plugin.Context) {
	userID, ok := requireLogin(ctx)
	if ok {
		return
	}
	page, pageSize := parsePagination(ctx)
	pagination, list, err := h.svc.ListByUser(userID, page, pageSize)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeMallRefundError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// ListByShop 按店铺列出
// GET /api/v1/mall/refunds/by-shop/:shop_id  （需登录）
func (h *RefundHandler) ListByShop(ctx plugin.Context) {
	shopID, err := parseSubID(ctx, "shop_id")
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的店铺ID"))
		return
	}
	page, pageSize := parsePagination(ctx)
	pagination, list, err := h.svc.ListByShop(shopID, page, pageSize)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeMallRefundError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// ListByOrder 按订单列出
// GET /api/v1/mall/refunds/by-order/:order_id  （需登录）
func (h *RefundHandler) ListByOrder(ctx plugin.Context) {
	orderID, err := parseSubID(ctx, "order_id")
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的订单ID"))
		return
	}
	list, err := h.svc.ListByOrder(orderID)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeMallRefundError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(list))
}

// Stats 退款统计
// GET /api/v1/mall/admin/refunds/stats  （需 mall:audit 权限）
func (h *RefundHandler) Stats(ctx plugin.Context) {
	regionID := getRegionID(ctx)
	shopID := uint(parseQueryInt(ctx, "shop_id", 0))
	startDate := ctx.Query("start_date")
	endDate := ctx.Query("end_date")
	stats, err := h.svc.Stats(regionID, shopID, startDate, endDate)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeMallRefundError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(stats))
}
