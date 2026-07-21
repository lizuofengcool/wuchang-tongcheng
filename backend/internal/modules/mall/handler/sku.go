// Package handler 同城商城 HTTP 处理层 - SKU 规格表
package handler

import (
	"net/http"

	"wuchang-tongcheng/internal/core/plugin"
	"wuchang-tongcheng/internal/core/response"
	"wuchang-tongcheng/internal/modules/mall/dto"
	"wuchang-tongcheng/internal/modules/mall/service"
	"wuchang-tongcheng/internal/pkg/utils"
)

// SkuHandler SKU HTTP 处理器
type SkuHandler struct {
	svc service.SkuService
}

// NewSkuHandler 创建 SKU Handler 实例
func NewSkuHandler(svc service.SkuService) *SkuHandler {
	return &SkuHandler{svc: svc}
}

// Create 创建 SKU
// POST /api/v1/mall/skus  （需登录，卖家操作）
func (h *SkuHandler) Create(ctx plugin.Context) {
	userID, ok := requireLogin(ctx)
	if ok {
		return
	}
	var req dto.CreateSkuRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	regionID := getRegionID(ctx)
	productID := uint(parseQueryInt(ctx, "product_id", 0))
	if productID == 0 {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: product_id 必填"))
		return
	}
	info, err := h.svc.Create(regionID, userID, productID, &req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeMallSkuError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("创建成功", info))
}

// Update 更新 SKU
// PUT /api/v1/mall/skus/:id  （需登录）
func (h *SkuHandler) Update(ctx plugin.Context) {
	userID, ok := requireLogin(ctx)
	if ok {
		return
	}
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	var req dto.UpdateSkuRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	if err := h.svc.Update(id, userID, &req); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeMallSkuError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("更新成功", nil))
}

// Delete 删除 SKU
// DELETE /api/v1/mall/skus/:id  （需登录）
func (h *SkuHandler) Delete(ctx plugin.Context) {
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
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeMallSkuError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("删除成功", nil))
}

// GetByID SKU 详情
// GET /api/v1/mall/skus/:id  （公开）
func (h *SkuHandler) GetByID(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	info, err := h.svc.GetByID(id)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeMallSkuNotFound, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(info))
}

// ListByProduct 按商品列出 SKU
// GET /api/v1/mall/skus/by-product/:product_id  （公开）
func (h *SkuHandler) ListByProduct(ctx plugin.Context) {
	productID, err := parseSubID(ctx, "product_id")
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的商品ID"))
		return
	}
	list, err := h.svc.ListByProduct(productID)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeMallSkuError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(list))
}

// ListByShop 按店铺列出 SKU
// GET /api/v1/mall/skus/by-shop/:shop_id  （公开）
func (h *SkuHandler) ListByShop(ctx plugin.Context) {
	shopID, err := parseSubID(ctx, "shop_id")
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的店铺ID"))
		return
	}
	var req dto.SkuListRequest
	_ = ctx.Bind(&req)
	if req.Page == 0 {
		req.Page, req.PageSize = parsePagination(ctx)
	}
	pagination, list, err := h.svc.ListByShop(shopID, &req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeMallSkuError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// UpdateStock 更新 SKU 库存
// PUT /api/v1/mall/skus/:id/stock  （需登录）
func (h *SkuHandler) UpdateStock(ctx plugin.Context) {
	userID, ok := requireLogin(ctx)
	if ok {
		return
	}
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	var req dto.SkuStockUpdate
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	stock := req.Stock
	if err := h.svc.UpdateStock(id, userID, stock); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeMallSkuError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("库存已更新", nil))
}

// BatchUpdateStock 批量更新库存
// PUT /api/v1/mall/skus/batch-stock  （需登录）
func (h *SkuHandler) BatchUpdateStock(ctx plugin.Context) {
	userID, ok := requireLogin(ctx)
	if ok {
		return
	}
	var req dto.BatchUpdateStockRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	if err := h.svc.BatchUpdateStock(userID, &req); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeMallSkuError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("批量更新成功", nil))
}
