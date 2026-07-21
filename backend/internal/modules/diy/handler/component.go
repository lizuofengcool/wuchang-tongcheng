// Package handler DIY 前端页面中台 HTTP 处理层 - 组件（component 子域）
package handler

import (
	"net/http"

	"wuchang-tongcheng/internal/core/plugin"
	"wuchang-tongcheng/internal/core/response"
	"wuchang-tongcheng/internal/modules/diy/dto"
	"wuchang-tongcheng/internal/modules/diy/service"
	"wuchang-tongcheng/internal/pkg/utils"
)

// ComponentHandler 组件 HTTP 处理器
type ComponentHandler struct {
	svc service.ComponentService
}

// NewComponentHandler 创建组件 Handler 实例
func NewComponentHandler(svc service.ComponentService) *ComponentHandler {
	return &ComponentHandler{svc: svc}
}

// List 组件列表（admin 权限）
// GET /api/v1/diy/admin/components
func (h *ComponentHandler) List(ctx plugin.Context) {
	var req dto.ComponentListRequest
	_ = ctx.Bind(&req)
	if req.Page == 0 {
		req.Page, req.PageSize = parsePagination(ctx)
	}
	pagination, list, err := h.svc.List(&req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(CodeDiyComponentError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// GetByID 组件详情（admin 权限）
// GET /api/v1/diy/admin/components/:id
func (h *ComponentHandler) GetByID(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		badRequest(ctx, "无效的ID")
		return
	}
	info, err := h.svc.GetByID(id)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(CodeDiyComponentNotFound, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(info))
}

// Create 创建组件（admin 权限）
// POST /api/v1/diy/admin/components
func (h *ComponentHandler) Create(ctx plugin.Context) {
	var req dto.CreateComponentRequest
	if err := ctx.Bind(&req); err != nil {
		badRequest(ctx, "参数错误: "+err.Error())
		return
	}
	info, err := h.svc.Create(&req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(CodeDiyComponentError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("创建成功", info))
}

// Update 更新组件（admin 权限）
// PUT /api/v1/diy/admin/components/:id
func (h *ComponentHandler) Update(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		badRequest(ctx, "无效的ID")
		return
	}
	var req dto.UpdateComponentRequest
	if err := ctx.Bind(&req); err != nil {
		badRequest(ctx, "参数错误: "+err.Error())
		return
	}
	if err := h.svc.Update(id, &req); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(CodeDiyComponentError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("更新成功", nil))
}

// Delete 删除组件（admin 权限）
// DELETE /api/v1/diy/admin/components/:id
func (h *ComponentHandler) Delete(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		badRequest(ctx, "无效的ID")
		return
	}
	if err := h.svc.Delete(id); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(CodeDiyComponentError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("删除成功", nil))
}

// ListByCategory 按分类获取组件（公开，C 端编辑器使用）
// GET /api/v1/diy/components/by-category/:category
func (h *ComponentHandler) ListByCategory(ctx plugin.Context) {
	category := ctx.Param("category")
	if category == "" {
		badRequest(ctx, "分类不能为空")
		return
	}
	page, pageSize := parsePagination(ctx)
	pagination, list, err := h.svc.ListByCategory(category, page, pageSize)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(CodeDiyComponentError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}
