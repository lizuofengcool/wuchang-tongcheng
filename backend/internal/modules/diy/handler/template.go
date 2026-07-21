// Package handler DIY 前端页面中台 HTTP 处理层 - 模板（template 子域）
package handler

import (
	"net/http"

	"wuchang-tongcheng/internal/core/plugin"
	"wuchang-tongcheng/internal/core/response"
	"wuchang-tongcheng/internal/modules/diy/dto"
	"wuchang-tongcheng/internal/modules/diy/service"
	"wuchang-tongcheng/internal/pkg/utils"
)

// TemplateHandler 模板 HTTP 处理器
type TemplateHandler struct {
	svc      service.TemplateService
	pageSvc  service.PageService
}

// NewTemplateHandler 创建模板 Handler 实例
func NewTemplateHandler(svc service.TemplateService, pageSvc service.PageService) *TemplateHandler {
	return &TemplateHandler{svc: svc, pageSvc: pageSvc}
}

// List 模板列表（admin 权限）
// GET /api/v1/diy/admin/templates
func (h *TemplateHandler) List(ctx plugin.Context) {
	var req dto.TemplateListRequest
	_ = ctx.Bind(&req)
	if req.Page == 0 {
		req.Page, req.PageSize = parsePagination(ctx)
	}
	pagination, list, err := h.svc.List(&req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(CodeDiyTemplateError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// GetByID 模板详情（admin 权限）
// GET /api/v1/diy/admin/templates/:id
func (h *TemplateHandler) GetByID(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		badRequest(ctx, "无效的ID")
		return
	}
	info, err := h.svc.GetByID(id)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(CodeDiyTemplateNotFound, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(info))
}

// Create 创建模板（admin 权限）
// POST /api/v1/diy/admin/templates
func (h *TemplateHandler) Create(ctx plugin.Context) {
	var req dto.CreateTemplateRequest
	if err := ctx.Bind(&req); err != nil {
		badRequest(ctx, "参数错误: "+err.Error())
		return
	}
	info, err := h.svc.Create(&req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(CodeDiyTemplateError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("创建成功", info))
}

// Update 更新模板（admin 权限）
// PUT /api/v1/diy/admin/templates/:id
func (h *TemplateHandler) Update(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		badRequest(ctx, "无效的ID")
		return
	}
	var req dto.UpdateTemplateRequest
	if err := ctx.Bind(&req); err != nil {
		badRequest(ctx, "参数错误: "+err.Error())
		return
	}
	if err := h.svc.Update(id, &req); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(CodeDiyTemplateError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("更新成功", nil))
}

// Delete 删除模板（admin 权限）
// DELETE /api/v1/diy/admin/templates/:id
func (h *TemplateHandler) Delete(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		badRequest(ctx, "无效的ID")
		return
	}
	if err := h.svc.Delete(id); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(CodeDiyTemplateError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("删除成功", nil))
}

// Apply 应用模板创建新页面（需登录）
// POST /api/v1/diy/templates/:id/apply
func (h *TemplateHandler) Apply(ctx plugin.Context) {
	userID, _, _, _ := getUserProfile(ctx)
	if userID == 0 {
		unauthorized(ctx)
		return
	}
	id, err := parseID(ctx)
	if err != nil {
		badRequest(ctx, "无效的ID")
		return
	}
	var req dto.ApplyTemplateRequest
	if err := ctx.Bind(&req); err != nil {
		badRequest(ctx, "参数错误: "+err.Error())
		return
	}
	regionID := getRegionID(ctx)
	info, err := h.svc.ApplyTemplate(id, regionID, userID, &req, h.pageSvc)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(CodeDiyTemplateError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("应用成功", info))
}
