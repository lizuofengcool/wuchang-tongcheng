// Package handler DIY 前端页面中台 HTTP 处理层 - 页面（page 子域）
package handler

import (
	"net/http"

	"wuchang-tongcheng/internal/core/plugin"
	"wuchang-tongcheng/internal/core/response"
	"wuchang-tongcheng/internal/modules/diy/dto"
	"wuchang-tongcheng/internal/modules/diy/service"
	"wuchang-tongcheng/internal/pkg/utils"
)

// PageHandler 页面 HTTP 处理器
type PageHandler struct {
	svc service.PageService
}

// NewPageHandler 创建页面 Handler 实例
func NewPageHandler(svc service.PageService) *PageHandler {
	return &PageHandler{svc: svc}
}

// GetBySlug 按 slug 获取已发布页面（公开）
// GET /api/v1/diy/pages/by-slug/:slug
func (h *PageHandler) GetBySlug(ctx plugin.Context) {
	slug := ctx.Param("slug")
	if slug == "" {
		badRequest(ctx, "slug 不能为空")
		return
	}
	regionID := getRegionID(ctx)
	info, err := h.svc.GetBySlug(regionID, slug)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(CodeDiyPageNotFound, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(info))
}

// GetByID 获取页面详情
// GET /api/v1/diy/pages/:id  （公开）
func (h *PageHandler) GetByID(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		badRequest(ctx, "无效的ID")
		return
	}
	info, err := h.svc.GetByID(id)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(CodeDiyPageNotFound, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(info))
}

// List 已发布页面列表（公开）
// GET /api/v1/diy/pages
func (h *PageHandler) List(ctx plugin.Context) {
	var req dto.PageListRequest
	_ = ctx.Bind(&req)
	if req.Page == 0 {
		req.Page, req.PageSize = parsePagination(ctx)
	}
	regionID := getRegionID(ctx)
	pagination, list, err := h.svc.List(regionID, &req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(CodeDiyPageError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// ListMine 我的页面列表（需登录）
// GET /api/v1/diy/pages/mine
func (h *PageHandler) ListMine(ctx plugin.Context) {
	userID, _, _, _ := getUserProfile(ctx)
	if userID == 0 {
		unauthorized(ctx)
		return
	}
	var req dto.PageListRequest
	_ = ctx.Bind(&req)
	if req.Page == 0 {
		req.Page, req.PageSize = parsePagination(ctx)
	}
	pagination, list, err := h.svc.ListMine(userID, &req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(CodeDiyPageError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// Create 创建页面（需登录）
// POST /api/v1/diy/pages
func (h *PageHandler) Create(ctx plugin.Context) {
	userID, _, _, _ := getUserProfile(ctx)
	if userID == 0 {
		unauthorized(ctx)
		return
	}
	var req dto.CreatePageRequest
	if err := ctx.Bind(&req); err != nil {
		badRequest(ctx, "参数错误: "+err.Error())
		return
	}
	regionID := getRegionID(ctx)
	info, err := h.svc.Create(regionID, userID, &req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(CodeDiyPageError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("创建成功", info))
}

// Update 更新页面（需登录）
// PUT /api/v1/diy/pages/:id
func (h *PageHandler) Update(ctx plugin.Context) {
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
	var req dto.UpdatePageRequest
	if err := ctx.Bind(&req); err != nil {
		badRequest(ctx, "参数错误: "+err.Error())
		return
	}
	if err := h.svc.Update(id, userID, &req); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(CodeDiyPageError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("更新成功", nil))
}

// Delete 删除页面（需登录）
// DELETE /api/v1/diy/pages/:id
func (h *PageHandler) Delete(ctx plugin.Context) {
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
	if err := h.svc.Delete(id, userID); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(CodeDiyPageError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("删除成功", nil))
}

// Publish 发布页面（需登录）
// POST /api/v1/diy/pages/:id/publish
func (h *PageHandler) Publish(ctx plugin.Context) {
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
	if err := h.svc.Publish(id, userID); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(CodeDiyPageStatusInvalid, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("发布成功", nil))
}

// Offline 下线页面（需登录）
// POST /api/v1/diy/pages/:id/offline
func (h *PageHandler) Offline(ctx plugin.Context) {
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
	if err := h.svc.Offline(id, userID); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(CodeDiyPageStatusInvalid, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("已下线", nil))
}

// Copy 复制页面（需登录）
// POST /api/v1/diy/pages/:id/copy
func (h *PageHandler) Copy(ctx plugin.Context) {
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
	var req struct {
		Title string `json:"title"`
	}
	_ = ctx.Bind(&req)
	info, err := h.svc.Copy(id, userID, req.Title)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(CodeDiyPageError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("复制成功", info))
}

// ===== M 端管理 =====

// AdminList 管理后台页面列表
// GET /api/v1/diy/admin/pages
func (h *PageHandler) AdminList(ctx plugin.Context) {
	var req dto.PageListAdminRequest
	_ = ctx.Bind(&req)
	if req.Page == 0 {
		req.Page, req.PageSize = parsePagination(ctx)
	}
	pagination, list, err := h.svc.AdminList(&req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(CodeDiyPageError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// AdminGetByID 管理后台获取详情
// GET /api/v1/diy/admin/pages/:id
func (h *PageHandler) AdminGetByID(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		badRequest(ctx, "无效的ID")
		return
	}
	info, err := h.svc.AdminGetByID(id)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(CodeDiyPageNotFound, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(info))
}

// AdminUpdateStatus 管理后台更新状态
// PUT /api/v1/diy/admin/pages/:id/status
func (h *PageHandler) AdminUpdateStatus(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		badRequest(ctx, "无效的ID")
		return
	}
	var req struct {
		Status int `json:"status" binding:"oneof=0 1 2"`
	}
	if err := ctx.Bind(&req); err != nil {
		badRequest(ctx, "参数错误: "+err.Error())
		return
	}
	if err := h.svc.AdminUpdateStatus(id, req.Status); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(CodeDiyPageStatusInvalid, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("状态已更新", nil))
}
