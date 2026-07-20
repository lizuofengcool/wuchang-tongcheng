// Package handler 同城114 HTTP 处理层 - 举报
// 提供用户对商户/评价/团购等内容的举报功能
package handler

import (
	"net/http"

	"wuchang-tongcheng/internal/core/plugin"
	"wuchang-tongcheng/internal/core/response"
	"wuchang-tongcheng/internal/modules/dh114/dto"
	"wuchang-tongcheng/internal/modules/dh114/service"
	"wuchang-tongcheng/internal/pkg/utils"
)

// ReportHandler 举报 HTTP 处理器
type ReportHandler struct {
	svc service.ReportService
}

// NewReportHandler 创建举报 Handler 实例
func NewReportHandler(svc service.ReportService) *ReportHandler {
	return &ReportHandler{svc: svc}
}

// List 举报列表
// GET /api/v1/dh114/reports  （公开，无需登录）
// 返回结构：{list, total, stats:{total, pending, processed}}
func (h *ReportHandler) List(ctx plugin.Context) {
	var req dto.ReportListRequest
	_ = ctx.Bind(&req)
	if req.Page == 0 {
		req.Page, req.PageSize = parsePagination(ctx)
	}
	regionID := getRegionID(ctx)
	pagination, list, stats, err := h.svc.List(regionID, &req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeDh114ReportError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(&dto.ReportListResponse{
		List:  list,
		Total: pagination.Total,
		Stats: *stats,
	}))
}

// Create 创建举报
// POST /api/v1/dh114/reports  （需登录）
func (h *ReportHandler) Create(ctx plugin.Context) {
	userID, username, _, _ := getUserProfile(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusOK, response.Unauthorized("请先登录"))
		return
	}
	var req dto.CreateReportRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	regionID := getRegionID(ctx)
	info, err := h.svc.Create(regionID, userID, username, &req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeDh114ReportError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("举报提交成功", info))
}

// Process 处理举报（M 端）
// PUT /api/v1/dh114/admin/reports/:id/process  （需 dh114:audit 权限）
func (h *ReportHandler) Process(ctx plugin.Context) {
	handlerID, handlerName, _, _ := getUserProfile(ctx)
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	var req dto.ProcessReportRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	if err := h.svc.Process(id, handlerID, handlerName, &req); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeDh114ReportError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("处理完成", nil))
}

// Delete 删除举报（M 端）
// DELETE /api/v1/dh114/admin/reports/:id  （需 dh114:audit 权限）
func (h *ReportHandler) Delete(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	if err := h.svc.Delete(id); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeDh114ReportError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("删除成功", nil))
}
