// Package handler 同城商城 HTTP 处理层 - 举报
package handler

import (
	"net/http"

	"wuchang-tongcheng/internal/core/plugin"
	"wuchang-tongcheng/internal/core/response"
	"wuchang-tongcheng/internal/modules/mall/dto"
	"wuchang-tongcheng/internal/modules/mall/service"
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

// Create 创建举报
// POST /api/v1/mall/reports  （需登录）
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
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeMallReportError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("举报已提交", info))
}

// GetByID 举报详情
// GET /api/v1/mall/admin/reports/:id  （需 mall:audit 权限）
func (h *ReportHandler) GetByID(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	info, err := h.svc.GetByID(id)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeMallReportNotFound, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(info))
}

// Process 处理举报
// PUT /api/v1/mall/admin/reports/:id/process  （需 mall:audit 权限）
func (h *ReportHandler) Process(ctx plugin.Context) {
	handlerID, _, _, _ := getUserProfile(ctx)
	handlerName, _ := ctx.Get("username")
	name, _ := handlerName.(string)
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
	if err := h.svc.Process(id, handlerID, name, &req); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeMallReportError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("处理成功", nil))
}

// Delete 删除举报
// DELETE /api/v1/mall/admin/reports/:id  （需 mall:audit 权限）
func (h *ReportHandler) Delete(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	if err := h.svc.Delete(id); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeMallReportNotFound, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("删除成功", nil))
}

// List 举报列表
// GET /api/v1/mall/admin/reports  （需 mall:audit 权限）
func (h *ReportHandler) List(ctx plugin.Context) {
	var req dto.ReportListRequest
	_ = ctx.Bind(&req)
	if req.Page == 0 {
		req.Page, req.PageSize = parsePagination(ctx)
	}
	regionID := getRegionID(ctx)
	pagination, list, stats, err := h.svc.List(regionID, &req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeMallReportError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(map[string]interface{}{
		"list":       list,
		"pagination": utils.PageResult(list, pagination),
		"stats":      stats,
	}))
}

// ListByUser 按举报人列出
// GET /api/v1/mall/reports/mine  （需登录）
func (h *ReportHandler) ListByUser(ctx plugin.Context) {
	userID, ok := requireLogin(ctx)
	if ok {
		return
	}
	page, pageSize := parsePagination(ctx)
	pagination, list, err := h.svc.ListByUser(userID, page, pageSize)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeMallReportError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// ListByTarget 按被举报对象列出
// GET /api/v1/mall/reports/by-target  （需登录）
func (h *ReportHandler) ListByTarget(ctx plugin.Context) {
	targetType := ctx.Query("target_type")
	if targetType == "" {
		ctx.JSON(http.StatusOK, response.BadRequest("目标类型不能为空"))
		return
	}
	targetID := uint(parseQueryInt(ctx, "target_id", 0))
	if targetID == 0 {
		ctx.JSON(http.StatusOK, response.BadRequest("目标 ID 不能为空"))
		return
	}
	list, err := h.svc.ListByTarget(targetType, targetID)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeMallReportError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(list))
}

// Stats 举报统计
// GET /api/v1/mall/admin/reports/stats  （需 mall:audit 权限）
func (h *ReportHandler) Stats(ctx plugin.Context) {
	regionID := getRegionID(ctx)
	stats, err := h.svc.Stats(regionID)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeMallReportError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(stats))
}
