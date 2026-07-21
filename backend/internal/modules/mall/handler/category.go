// Package handler 同城商城 HTTP 处理层 - 商品分类
package handler

import (
	"net/http"

	"wuchang-tongcheng/internal/core/plugin"
	"wuchang-tongcheng/internal/core/response"
	"wuchang-tongcheng/internal/modules/mall/dto"
	"wuchang-tongcheng/internal/modules/mall/service"
	"wuchang-tongcheng/internal/pkg/utils"
)

// CategoryHandler 商品分类 HTTP 处理器
type CategoryHandler struct {
	svc service.CategoryService
}

// NewCategoryHandler 创建分类 Handler 实例
func NewCategoryHandler(svc service.CategoryService) *CategoryHandler {
	return &CategoryHandler{svc: svc}
}

// Create 创建分类
// POST /api/v1/mall/admin/categories  （需 mall:audit 权限）
func (h *CategoryHandler) Create(ctx plugin.Context) {
	var req dto.CreateCategoryRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	regionID := getRegionID(ctx)
	info, err := h.svc.Create(regionID, &req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeMallCategoryError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("创建成功", info))
}

// Update 更新分类
// PUT /api/v1/mall/admin/categories/:id  （需 mall:audit 权限）
func (h *CategoryHandler) Update(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	var req dto.UpdateCategoryRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	if err := h.svc.Update(id, &req); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeMallCategoryError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("更新成功", nil))
}

// Delete 删除分类
// DELETE /api/v1/mall/admin/categories/:id  （需 mall:audit 权限）
func (h *CategoryHandler) Delete(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	if err := h.svc.Delete(id); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeMallCategoryError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("删除成功", nil))
}

// GetByID 分类详情
// GET /api/v1/mall/categories/:id  （公开）
func (h *CategoryHandler) GetByID(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	info, err := h.svc.GetByID(id)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeMallCategoryNotFound, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(info))
}

// List 分类列表
// GET /api/v1/mall/categories  （公开）
func (h *CategoryHandler) List(ctx plugin.Context) {
	var req dto.CategoryListRequest
	_ = ctx.Bind(&req)
	if req.Page == 0 {
		req.Page, req.PageSize = parsePagination(ctx)
	}
	regionID := getRegionID(ctx)
	pagination, list, err := h.svc.List(regionID, &req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeMallCategoryError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// ListTree 分类树形列表
// GET /api/v1/mall/categories/tree  （公开）
func (h *CategoryHandler) ListTree(ctx plugin.Context) {
	regionID := getRegionID(ctx)
	list, err := h.svc.ListTree(regionID)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeMallCategoryError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(list))
}

// ListByParent 按父级查询子分类
// GET /api/v1/mall/categories/by-parent  （公开）
func (h *CategoryHandler) ListByParent(ctx plugin.Context) {
	parentID := uint(parseQueryInt(ctx, "parent_id", 0))
	list, err := h.svc.ListByParent(parentID)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeMallCategoryError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(list))
}

// ListEnabled 启用中的分类列表
// GET /api/v1/mall/categories/enabled  （公开）
func (h *CategoryHandler) ListEnabled(ctx plugin.Context) {
	regionID := getRegionID(ctx)
	list, err := h.svc.ListEnabled(regionID)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeMallCategoryError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(list))
}

// UpdateStatus 更新分类状态
// PUT /api/v1/mall/admin/categories/:id/status  （需 mall:audit 权限）
func (h *CategoryHandler) UpdateStatus(ctx plugin.Context) {
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
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeMallCategoryError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("状态已更新", nil))
}
