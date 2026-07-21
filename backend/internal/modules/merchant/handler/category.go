// Package handler 商户中台 HTTP 处理层 - 类目
package handler

import (
	"net/http"

	"wuchang-tongcheng/internal/core/plugin"
	"wuchang-tongcheng/internal/core/response"
	"wuchang-tongcheng/internal/modules/merchant/dto"
	"wuchang-tongcheng/internal/modules/merchant/service"
	"wuchang-tongcheng/internal/pkg/utils"
)

// CategoryHandler 类目 HTTP 处理器
type CategoryHandler struct {
	svc service.CategoryService
}

// NewCategoryHandler 创建类目 Handler 实例
func NewCategoryHandler(svc service.CategoryService) *CategoryHandler {
	return &CategoryHandler{svc: svc}
}

// Tree 类目树（公开）
// GET /api/v1/merchant/categories/tree
func (h *CategoryHandler) Tree(ctx plugin.Context) {
	tree, err := h.svc.Tree()
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(CodeMerchantError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(tree))
}

// List 类目列表（公开）
// GET /api/v1/merchant/categories
func (h *CategoryHandler) List(ctx plugin.Context) {
	var req dto.CategoryListRequest
	_ = ctx.Bind(&req)
	if req.Page == 0 {
		req.Page, req.PageSize = parsePagination(ctx)
	}
	pagination, list, err := h.svc.List(&req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(CodeMerchantError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// GetByID 类目详情（公开）
// GET /api/v1/merchant/categories/:id
func (h *CategoryHandler) GetByID(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	info, err := h.svc.GetByID(id)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(CodeMerchantCategoryNotFound, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(info))
}

// Create 创建类目（M 端）
// POST /api/v1/merchant/admin/categories
func (h *CategoryHandler) Create(ctx plugin.Context) {
	var req dto.CreateCategoryRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, failParam("参数错误: "+err.Error()))
		return
	}
	info, err := h.svc.Create(&req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(CodeMerchantError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("类目创建成功", info))
}

// Update 更新类目（M 端）
// PUT /api/v1/merchant/admin/categories/:id
func (h *CategoryHandler) Update(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	var req dto.UpdateCategoryRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, failParam("参数错误: "+err.Error()))
		return
	}
	if err := h.svc.Update(id, &req); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(CodeMerchantCategoryNotFound, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("更新成功", nil))
}

// Delete 删除类目（M 端）
// DELETE /api/v1/merchant/admin/categories/:id
func (h *CategoryHandler) Delete(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	if err := h.svc.Delete(id); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(CodeMerchantCategoryHasChildren, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("删除成功", nil))
}

// UpdateStatus 更新类目状态（M 端）
// PUT /api/v1/merchant/admin/categories/:id/status
func (h *CategoryHandler) UpdateStatus(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	var req dto.CategoryStatusUpdateRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, failParam("参数错误: "+err.Error()))
		return
	}
	if err := h.svc.UpdateStatus(id, req.Status); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(CodeMerchantCategoryStatusInvalid, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("状态更新成功", nil))
}
