// Package handler SKU 规格 HTTP 处理层
// 依据 v3.2.1 架构方案：对标闲鱼/转转 SKU 规格管理
package handler

import (
	"net/http"
	"strconv"

	"wuchang-tongcheng/internal/core/plugin"
	"wuchang-tongcheng/internal/core/response"
	"wuchang-tongcheng/internal/modules/ershou/dto"
	"wuchang-tongcheng/internal/modules/ershou/service"
	"wuchang-tongcheng/internal/pkg/utils"
)

// SKUHandler SKU 规格 HTTP 处理器
type SKUHandler struct {
	service service.SKUService
}

// NewSKUHandler 创建 SKU Handler 实例
func NewSKUHandler(svc service.SKUService) *SKUHandler {
	return &SKUHandler{service: svc}
}

// List 商品 SKU 列表
// GET /api/v1/ershou/:id/skus  （公开）
func (h *SKUHandler) List(ctx plugin.Context) {
	ershouID, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	resp, err := h.service.List(ershouID)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeErshouError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(resp))
}

// Create 新增 SKU
// POST /api/v1/ershou/:id/skus  （需登录 + 仅发布者本人）
func (h *SKUHandler) Create(ctx plugin.Context) {
	userID, _, _, _ := getUserProfile(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusOK, response.Unauthorized("请先登录"))
		return
	}
	ershouID, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	var req dto.SKURequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	resp, err := h.service.Create(ershouID, userID, &req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeErshouError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("SKU 创建成功", resp))
}

// Update 更新 SKU
// PUT /api/v1/ershou/:id/skus/:sku_id  （需登录 + 仅发布者本人）
func (h *SKUHandler) Update(ctx plugin.Context) {
	userID, _, _, _ := getUserProfile(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusOK, response.Unauthorized("请先登录"))
		return
	}
	ershouID, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	skuID, err := parseSubID(ctx, "sku_id")
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的SKU ID"))
		return
	}
	var req dto.SKURequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	resp, err := h.service.Update(ershouID, skuID, userID, &req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeErshouError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("SKU 更新成功", resp))
}

// Delete 删除 SKU
// DELETE /api/v1/ershou/:id/skus/:sku_id  （需登录 + 仅发布者本人）
func (h *SKUHandler) Delete(ctx plugin.Context) {
	userID, _, _, _ := getUserProfile(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusOK, response.Unauthorized("请先登录"))
		return
	}
	ershouID, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	skuID, err := parseSubID(ctx, "sku_id")
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的SKU ID"))
		return
	}
	if err := h.service.Delete(ershouID, skuID, userID); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeErshouError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("SKU 删除成功", nil))
}

// parseSubID 解析 URL 中的自定义子参数（如 :sku_id / :order_id / :brand_id 等）
func parseSubID(ctx plugin.Context, key string) (uint, error) {
	idStr := ctx.Param(key)
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		return 0, err
	}
	return uint(id), nil
}
