// Package handler 投递记录 HTTP 处理层
package handler

import (
	"net/http"

	"wuchang-tongcheng/internal/core/plugin"
	"wuchang-tongcheng/internal/core/response"
	"wuchang-tongcheng/internal/modules/job/dto"
	"wuchang-tongcheng/internal/modules/job/service"
	"wuchang-tongcheng/internal/pkg/utils"
)

// ApplicationHandler 投递记录 HTTP 处理器
type ApplicationHandler struct {
	svc service.ApplicationService
}

// NewApplicationHandler 创建投递记录 Handler 实例
func NewApplicationHandler(svc service.ApplicationService) *ApplicationHandler {
	return &ApplicationHandler{svc: svc}
}

// Create 投递简历
// POST /api/v1/job/applications
func (h *ApplicationHandler) Create(ctx plugin.Context) {
	userID, username, _, avatar := getUserProfile(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusOK, response.Unauthorized("请先登录"))
		return
	}
	var req dto.ApplicationCreateRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	regionID := getRegionID(ctx)
	info, err := h.svc.Create(regionID, userID, username, avatar, &req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeApplicationError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("投递成功", info))
}

// GetByID 投递详情
// GET /api/v1/job/applications/:id
func (h *ApplicationHandler) GetByID(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	userID, _, _, _ := getUserProfile(ctx)
	info, err := h.svc.GetByID(id, userID)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeApplicationNotFound, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(info))
}

// List 投递列表
// GET /api/v1/job/applications
func (h *ApplicationHandler) List(ctx plugin.Context) {
	userID, _, _, _ := getUserProfile(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusOK, response.Unauthorized("请先登录"))
		return
	}
	var req dto.ApplicationListQuery
	_ = ctx.Bind(&req)
	pagination, list, err := h.svc.List(userID, &req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeApplicationError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// ListByJobID 职位的投递列表
// GET /api/v1/job/jobs/:id/applications
func (h *ApplicationHandler) ListByJobID(ctx plugin.Context) {
	userID, _, _, _ := getUserProfile(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusOK, response.Unauthorized("请先登录"))
		return
	}
	jobID, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	page, pageSize := parsePagination(ctx)
	pagination, list, err := h.svc.ListByJobID(jobID, userID, page, pageSize)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeApplicationError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// StatusUpdate 投递状态变更
// PUT /api/v1/job/applications/:id/status
func (h *ApplicationHandler) StatusUpdate(ctx plugin.Context) {
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
	var req dto.ApplicationStatusUpdateRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	info, err := h.svc.StatusUpdate(id, userID, &req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeApplicationError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(info))
}

// BatchAction 批量操作
// POST /api/v1/job/applications/batch
func (h *ApplicationHandler) BatchAction(ctx plugin.Context) {
	userID, _, _, _ := getUserProfile(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusOK, response.Unauthorized("请先登录"))
		return
	}
	var req dto.ApplicationBatchActionRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	resp, err := h.svc.BatchAction(userID, &req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeApplicationError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(resp))
}

// Stats 投递统计
// GET /api/v1/job/applications/stats
func (h *ApplicationHandler) Stats(ctx plugin.Context) {
	userID, _, _, _ := getUserProfile(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusOK, response.Unauthorized("请先登录"))
		return
	}
	role := ctx.DefaultQuery("role", "applicant")
	stats, err := h.svc.Stats(userID, role)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeApplicationError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(stats))
}
