// Package handler 营销活动中台 HTTP 处理层 - 营销活动（activity 子域）
package handler

import (
	"net/http"

	"wuchang-tongcheng/internal/core/plugin"
	"wuchang-tongcheng/internal/core/response"
	"wuchang-tongcheng/internal/modules/marketing/dto"
	"wuchang-tongcheng/internal/modules/marketing/service"
	"wuchang-tongcheng/internal/pkg/utils"
)

// ActivityHandler 营销活动 HTTP 处理器
type ActivityHandler struct {
	svc service.ActivityService
}

// NewActivityHandler 创建营销活动 Handler 实例
func NewActivityHandler(svc service.ActivityService) *ActivityHandler {
	return &ActivityHandler{svc: svc}
}

// Create 创建营销活动（M 端）
// POST /api/v1/marketing/activities  （需 marketing:manage 权限）
func (h *ActivityHandler) Create(ctx plugin.Context) {
	var req dto.CreateActivityRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	regionID := getRegionID(ctx)
	info, err := h.svc.Create(regionID, &req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(CodeMarketingActivityError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("创建成功", info))
}

// Update 更新营销活动
// PUT /api/v1/marketing/activities/:id  （需 marketing:manage 权限）
func (h *ActivityHandler) Update(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	var req dto.UpdateActivityRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	if err := h.svc.Update(id, &req); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(CodeMarketingActivityError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("更新成功", nil))
}

// Delete 删除营销活动
// DELETE /api/v1/marketing/activities/:id  （需 marketing:manage 权限）
func (h *ActivityHandler) Delete(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	if err := h.svc.Delete(id); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(CodeMarketingActivityError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("删除成功", nil))
}

// GetByID 活动详情
// GET /api/v1/marketing/activities/:id  （公开）
func (h *ActivityHandler) GetByID(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	info, err := h.svc.GetByID(id)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(CodeMarketingActivityNotFound, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(info))
}

// List 活动列表（M 端）
// GET /api/v1/marketing/activities  （需 marketing:manage 权限）
func (h *ActivityHandler) List(ctx plugin.Context) {
	var req dto.ActivityListRequest
	_ = ctx.Bind(&req)
	if req.Page == 0 {
		req.Page, req.PageSize = parsePagination(ctx)
	}
	regionID := getRegionID(ctx)
	pagination, list, err := h.svc.List(regionID, &req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(CodeMarketingActivityError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// ListOngoing 进行中的活动（C 端）
// GET /api/v1/marketing/activities/ongoing  （公开）
func (h *ActivityHandler) ListOngoing(ctx plugin.Context) {
	page, pageSize := parsePagination(ctx)
	regionID := getRegionID(ctx)
	pagination, list, err := h.svc.ListOngoing(regionID, page, pageSize)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(CodeMarketingActivityError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// ListUpcoming 即将开始的活动（C 端）
// GET /api/v1/marketing/activities/upcoming  （公开）
func (h *ActivityHandler) ListUpcoming(ctx plugin.Context) {
	page, pageSize := parsePagination(ctx)
	regionID := getRegionID(ctx)
	pagination, list, err := h.svc.ListUpcoming(regionID, page, pageSize)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(CodeMarketingActivityError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// ListEnded 已结束的活动（C 端）
// GET /api/v1/marketing/activities/ended  （公开）
func (h *ActivityHandler) ListEnded(ctx plugin.Context) {
	page, pageSize := parsePagination(ctx)
	regionID := getRegionID(ctx)
	pagination, list, err := h.svc.ListEnded(regionID, page, pageSize)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(CodeMarketingActivityError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// UpdateStatus 手动更新活动状态（M 端）
// PUT /api/v1/marketing/activities/:id/status  （需 marketing:manage 权限）
func (h *ActivityHandler) UpdateStatus(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	var req dto.AdminUpdateStatusRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	if err := h.svc.UpdateStatus(id, req.Status); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(CodeMarketingActivityError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("状态已更新", nil))
}

// Statistics 活动统计（M 端）
// GET /api/v1/marketing/activities/statistics  （需 marketing:manage 权限）
func (h *ActivityHandler) Statistics(ctx plugin.Context) {
	regionID := getRegionID(ctx)
	stats, err := h.svc.Statistics(regionID)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(CodeMarketingActivityError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(stats))
}
