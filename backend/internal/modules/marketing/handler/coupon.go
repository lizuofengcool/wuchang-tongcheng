// Package handler 营销活动中台 HTTP 处理层 - 优惠券（coupon 子域）
package handler

import (
	"net/http"

	"wuchang-tongcheng/internal/core/plugin"
	"wuchang-tongcheng/internal/core/response"
	"wuchang-tongcheng/internal/modules/marketing/dto"
	"wuchang-tongcheng/internal/modules/marketing/service"
	"wuchang-tongcheng/internal/pkg/utils"
)

// CouponHandler 优惠券 HTTP 处理器
type CouponHandler struct {
	svc service.CouponService
}

// NewCouponHandler 创建优惠券 Handler 实例
func NewCouponHandler(svc service.CouponService) *CouponHandler {
	return &CouponHandler{svc: svc}
}

// Create 创建优惠券（M 端）
// POST /api/v1/marketing/coupons  （需 marketing:manage 权限）
func (h *CouponHandler) Create(ctx plugin.Context) {
	var req dto.CreateCouponRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	regionID := getRegionID(ctx)
	info, err := h.svc.Create(regionID, &req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(CodeMarketingCouponError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("创建成功", info))
}

// Update 更新优惠券
// PUT /api/v1/marketing/coupons/:id  （需 marketing:manage 权限）
func (h *CouponHandler) Update(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	var req dto.UpdateCouponRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	if err := h.svc.Update(id, &req); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(CodeMarketingCouponError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("更新成功", nil))
}

// Delete 删除优惠券
// DELETE /api/v1/marketing/coupons/:id  （需 marketing:manage 权限）
func (h *CouponHandler) Delete(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	if err := h.svc.Delete(id); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(CodeMarketingCouponError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("删除成功", nil))
}

// GetByID 优惠券详情
// GET /api/v1/marketing/coupons/:id  （公开）
func (h *CouponHandler) GetByID(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	info, err := h.svc.GetByID(id)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(CodeMarketingCouponNotFound, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(info))
}

// List 优惠券列表（M 端）
// GET /api/v1/marketing/coupons  （需 marketing:manage 权限）
func (h *CouponHandler) List(ctx plugin.Context) {
	var req dto.CouponListRequest
	_ = ctx.Bind(&req)
	if req.Page == 0 {
		req.Page, req.PageSize = parsePagination(ctx)
	}
	regionID := getRegionID(ctx)
	pagination, list, err := h.svc.List(regionID, &req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(CodeMarketingCouponError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// ListAvailable 可领取优惠券列表（C 端）
// GET /api/v1/marketing/coupons/available  （公开）
func (h *CouponHandler) ListAvailable(ctx plugin.Context) {
	page, pageSize := parsePagination(ctx)
	regionID := getRegionID(ctx)
	pagination, list, err := h.svc.ListAvailable(regionID, page, pageSize)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(CodeMarketingCouponError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// Receive 领取优惠券（C 端）
// POST /api/v1/marketing/coupons/:id/receive  （需登录）
func (h *CouponHandler) Receive(ctx plugin.Context) {
	userID, _, _, _ := getUserProfile(ctx)
	if userID == 0 {
		unauthorized(ctx)
		return
	}
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	var req dto.CouponReceiveRequest
	_ = ctx.Bind(&req)
	if err := h.svc.Receive(userID, id, req.Source); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(CodeMarketingCouponError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("领取成功", nil))
}

// Use 使用优惠券（C 端，订单下单时调用）
// POST /api/v1/marketing/user-coupons/:id/use  （需登录）
func (h *CouponHandler) Use(ctx plugin.Context) {
	userID, _, _, _ := getUserProfile(ctx)
	if userID == 0 {
		unauthorized(ctx)
		return
	}
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	var req dto.CouponUseRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	if err := h.svc.Use(id, req.OrderID); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(CodeMarketingCouponError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("使用成功", nil))
}

// Refund 退还优惠券（订单退款时调用）
// POST /api/v1/marketing/user-coupons/:id/refund  （需登录）
func (h *CouponHandler) Refund(ctx plugin.Context) {
	userID, _, _, _ := getUserProfile(ctx)
	if userID == 0 {
		unauthorized(ctx)
		return
	}
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	if err := h.svc.Refund(id); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(CodeMarketingCouponError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("退还成功", nil))
}

// ListMine 我的优惠券列表（C 端）
// GET /api/v1/marketing/my-coupons  （需登录）
func (h *CouponHandler) ListMine(ctx plugin.Context) {
	userID, _, _, _ := getUserProfile(ctx)
	if userID == 0 {
		unauthorized(ctx)
		return
	}
	var req dto.UserCouponListRequest
	_ = ctx.Bind(&req)
	if req.Page == 0 {
		req.Page, req.PageSize = parsePagination(ctx)
	}
	pagination, list, err := h.svc.ListMine(userID, &req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(CodeMarketingCouponError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// Statistics 优惠券统计（M 端）
// GET /api/v1/marketing/coupons/statistics  （需 marketing:manage 权限）
func (h *CouponHandler) Statistics(ctx plugin.Context) {
	regionID := getRegionID(ctx)
	stats, err := h.svc.Statistics(regionID)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(CodeMarketingCouponError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(stats))
}
