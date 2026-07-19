// Package handler 房源发布 HTTP 处理层
// 依据 v3.2.1 架构方案第五章：与 houses 主表 1:1 冗余发布信息
package handler

import (
	"net/http"

	"wuchang-tongcheng/internal/core/plugin"
	"wuchang-tongcheng/internal/core/response"
	"wuchang-tongcheng/internal/modules/house/dto"
	"wuchang-tongcheng/internal/modules/house/service"
	"wuchang-tongcheng/internal/pkg/utils"
)

// ListingHandler 房源发布 HTTP 处理器
type ListingHandler struct {
	service service.ListingService
}

// NewListingHandler 创建 ListingHandler 实例
func NewListingHandler(svc service.ListingService) *ListingHandler {
	return &ListingHandler{service: svc}
}

// ===== C 端 =====

// Create 创建发布
// POST /api/v1/house/listings  （需登录）
func (h *ListingHandler) Create(ctx plugin.Context) {
	userID, username, phone, avatar := getUserProfile(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusOK, response.Unauthorized("请先登录"))
		return
	}
	var req dto.CreateListingRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	regionID := getRegionID(ctx)
	info, err := h.service.Create(regionID, userID, username, phone, avatar, &req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeHousePublishError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("发布成功", info))
}

// Update 更新发布（仅发布者本人）
// PUT /api/v1/house/listings/:id  （需登录）
func (h *ListingHandler) Update(ctx plugin.Context) {
	userID, _, _, _ := getUserProfile(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusOK, response.Unauthorized("请先登录"))
		return
	}
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	var req dto.UpdateListingRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	if err := h.service.Update(id, userID, &req); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeHouseError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("更新成功", nil))
}

// Delete 删除发布（仅发布者本人）
// DELETE /api/v1/house/listings/:id  （需登录）
func (h *ListingHandler) Delete(ctx plugin.Context) {
	userID, _, _, _ := getUserProfile(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusOK, response.Unauthorized("请先登录"))
		return
	}
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	if err := h.service.Delete(id, userID); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeHouseError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("删除成功", nil))
}

// GetByID 获取发布详情（同时增加浏览量）
// GET /api/v1/house/listings/:id  （公开）
func (h *ListingHandler) GetByID(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	userID, _, _, _ := getUserProfile(ctx)
	info, err := h.service.GetByID(id, userID)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeHouseNotFound, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(info))
}

// List 发布列表（C 端，仅返回已发布）
// GET /api/v1/house/listings  （公开）
func (h *ListingHandler) List(ctx plugin.Context) {
	var req dto.ListingListQuery
	_ = ctx.Bind(&req)
	regionID := getRegionID(ctx)
	pagination, list, err := h.service.List(regionID, &req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeHouseError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// ListMine 我的发布列表
// GET /api/v1/house/listings/mine  （需登录）
func (h *ListingHandler) ListMine(ctx plugin.Context) {
	userID, _, _, _ := getUserProfile(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusOK, response.Unauthorized("请先登录"))
		return
	}
	page, pageSize := parsePagination(ctx)
	pagination, list, err := h.service.ListMine(userID, page, pageSize)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeHouseError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// Refresh 刷新发布（提升排序权重）
// POST /api/v1/house/listings/:id/refresh  （需登录 + 发布者本人）
func (h *ListingHandler) Refresh(ctx plugin.Context) {
	userID, _, _, _ := getUserProfile(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusOK, response.Unauthorized("请先登录"))
		return
	}
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	if err := h.service.Refresh(id, userID); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeHouseError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("刷新成功", nil))
}

// ===== M 端管理 =====

// AdminList 管理后台发布列表
// GET /api/v1/admin/house/listings  （需 house:audit 权限）
func (h *ListingHandler) AdminList(ctx plugin.Context) {
	var req dto.ListingAdminListQuery
	_ = ctx.Bind(&req)
	pagination, list, err := h.service.AdminList(&req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeHouseError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// Audit 审核发布（通过/拒绝）
// PUT /api/v1/admin/house/listings/:id/audit  （需 house:audit 权限）
func (h *ListingHandler) Audit(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	var req dto.ListingAuditRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	if err := h.service.Audit(id, req.AuditStatus, req.AuditReason); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeHouseAuditError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("审核完成", nil))
}

// UpdateStatus 强制下架/恢复
// PUT /api/v1/admin/house/listings/:id/status  （需 house:audit 权限）
func (h *ListingHandler) UpdateStatus(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	var req dto.UpdateHouseStatusRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	if err := h.service.UpdateStatus(id, req.Status); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeHouseAuditError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("状态更新成功", nil))
}
