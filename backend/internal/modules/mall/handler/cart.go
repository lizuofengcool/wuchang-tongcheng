// Package handler 同城商城 HTTP 处理层 - 购物车
package handler

import (
	"net/http"

	"wuchang-tongcheng/internal/core/plugin"
	"wuchang-tongcheng/internal/core/response"
	"wuchang-tongcheng/internal/modules/mall/dto"
	"wuchang-tongcheng/internal/modules/mall/service"
	"wuchang-tongcheng/internal/pkg/utils"
)

// CartHandler 购物车 HTTP 处理器
type CartHandler struct {
	svc service.CartService
}

// NewCartHandler 创建购物车 Handler 实例
func NewCartHandler(svc service.CartService) *CartHandler {
	return &CartHandler{svc: svc}
}

// Add 加入购物车
// POST /api/v1/mall/cart  （需登录）
func (h *CartHandler) Add(ctx plugin.Context) {
	userID, ok := requireLogin(ctx)
	if ok {
		return
	}
	var req dto.AddCartRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	regionID := getRegionID(ctx)
	info, err := h.svc.Add(regionID, userID, &req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeMallCartError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("已加入购物车", info))
}

// Update 更新购物车项
// PUT /api/v1/mall/cart/:id  （需登录）
func (h *CartHandler) Update(ctx plugin.Context) {
	userID, ok := requireLogin(ctx)
	if ok {
		return
	}
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	var req dto.UpdateCartRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	if err := h.svc.Update(id, userID, &req); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeMallCartError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("更新成功", nil))
}

// BatchUpdate 批量更新购物车
// PUT /api/v1/mall/cart/batch  （需登录）
func (h *CartHandler) BatchUpdate(ctx plugin.Context) {
	userID, ok := requireLogin(ctx)
	if ok {
		return
	}
	var req dto.BatchUpdateCartRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	if err := h.svc.BatchUpdate(userID, &req); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeMallCartError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("批量更新成功", nil))
}

// Delete 删除购物车项
// DELETE /api/v1/mall/cart/:id  （需登录）
func (h *CartHandler) Delete(ctx plugin.Context) {
	userID, ok := requireLogin(ctx)
	if ok {
		return
	}
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	if err := h.svc.Delete(id, userID); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeMallCartError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("删除成功", nil))
}

// BatchDelete 批量删除购物车
// DELETE /api/v1/mall/cart/batch  （需登录）
func (h *CartHandler) BatchDelete(ctx plugin.Context) {
	userID, ok := requireLogin(ctx)
	if ok {
		return
	}
	var req dto.BatchDeleteRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	if err := h.svc.BatchDelete(userID, req.IDs); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeMallCartError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("批量删除成功", nil))
}

// ClearByUser 清空用户购物车
// DELETE /api/v1/mall/cart  （需登录）
func (h *CartHandler) ClearByUser(ctx plugin.Context) {
	userID, ok := requireLogin(ctx)
	if ok {
		return
	}
	if err := h.svc.ClearByUser(userID); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeMallCartError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("已清空购物车", nil))
}

// ClearByUserAndShop 清空用户某店铺购物车
// DELETE /api/v1/mall/cart/by-shop/:shop_id  （需登录）
func (h *CartHandler) ClearByUserAndShop(ctx plugin.Context) {
	userID, ok := requireLogin(ctx)
	if ok {
		return
	}
	shopID, err := parseSubID(ctx, "shop_id")
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的店铺ID"))
		return
	}
	if err := h.svc.ClearByUserAndShop(userID, shopID); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeMallCartError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("已清空该店铺购物车", nil))
}

// SelectAll 全选/取消全选
// PUT /api/v1/mall/cart/select-all  （需登录）
func (h *CartHandler) SelectAll(ctx plugin.Context) {
	userID, ok := requireLogin(ctx)
	if ok {
		return
	}
	var req dto.SelectAllCartRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	if err := h.svc.SelectAll(userID, &req); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeMallCartError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("操作成功", nil))
}

// GetByID 购物车项详情
// GET /api/v1/mall/cart/:id  （需登录）
func (h *CartHandler) GetByID(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	info, err := h.svc.GetByID(id)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeMallCartNotFound, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(info))
}

// ListByUser 用户的购物车列表
// GET /api/v1/mall/cart  （需登录）
func (h *CartHandler) ListByUser(ctx plugin.Context) {
	userID, ok := requireLogin(ctx)
	if ok {
		return
	}
	list, err := h.svc.ListByUser(userID)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeMallCartError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(list))
}

// ListByUserAndShop 按用户和店铺列出购物车
// GET /api/v1/mall/cart/by-shop/:shop_id  （需登录）
func (h *CartHandler) ListByUserAndShop(ctx plugin.Context) {
	userID, ok := requireLogin(ctx)
	if ok {
		return
	}
	shopID, err := parseSubID(ctx, "shop_id")
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的店铺ID"))
		return
	}
	list, err := h.svc.ListByUserAndShop(userID, shopID)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeMallCartError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(list))
}

// ListSelected 已选中的购物车项
// GET /api/v1/mall/cart/selected  （需登录）
func (h *CartHandler) ListSelected(ctx plugin.Context) {
	userID, ok := requireLogin(ctx)
	if ok {
		return
	}
	list, err := h.svc.ListSelected(userID)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeMallCartError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(list))
}

// ListGroupByShop 按店铺分组列出购物车
// GET /api/v1/mall/cart/group-by-shop  （需登录）
func (h *CartHandler) ListGroupByShop(ctx plugin.Context) {
	userID, ok := requireLogin(ctx)
	if ok {
		return
	}
	list, err := h.svc.ListGroupByShop(userID)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeMallCartError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(list))
}

// Summary 购物车汇总
// GET /api/v1/mall/cart/summary  （需登录）
func (h *CartHandler) Summary(ctx plugin.Context) {
	userID, ok := requireLogin(ctx)
	if ok {
		return
	}
	var req dto.CartSummaryRequest
	_ = ctx.Bind(&req)
	summary, err := h.svc.Summary(userID, &req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeMallCartError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(summary))
}

// CountByUser 用户购物车数量
// GET /api/v1/mall/cart/count  （需登录）
func (h *CartHandler) CountByUser(ctx plugin.Context) {
	userID, ok := requireLogin(ctx)
	if ok {
		return
	}
	count, err := h.svc.CountByUser(userID)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeMallCartError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(map[string]int64{"count": count}))
}

// CountSelectedByUser 用户已选中购物车数量
// GET /api/v1/mall/cart/selected/count  （需登录）
func (h *CartHandler) CountSelectedByUser(ctx plugin.Context) {
	userID, ok := requireLogin(ctx)
	if ok {
		return
	}
	count, err := h.svc.CountSelectedByUser(userID)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeMallCartError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(map[string]int64{"count": count}))
}
