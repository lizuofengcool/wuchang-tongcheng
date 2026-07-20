// Package handler love 相亲交友 HTTP 处理层 - 举报
// 依据需求文档 1.10：4 维数据隔离（region_id + user_id）
package handler

import (
	"net/http"

	"wuchang-tongcheng/internal/core/middleware"
	"wuchang-tongcheng/internal/core/plugin"
	"wuchang-tongcheng/internal/core/response"
	"wuchang-tongcheng/internal/modules/love/dto"
	"wuchang-tongcheng/internal/modules/love/service"
	"wuchang-tongcheng/internal/pkg/utils"
)

// ReportHandler 举报 HTTP 处理器
type ReportHandler struct {
	service service.LoveReportService
}

// NewReportHandler 创建 ReportHandler 实例
func NewReportHandler(svc service.LoveReportService) *ReportHandler {
	return &ReportHandler{service: svc}
}

// ===== C 端 =====

// Create 提交举报
// POST /api/v1/love/reports  （需登录）
func (h *ReportHandler) Create(ctx plugin.Context) {
	userID, username, avatar := getUserProfile(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusOK, response.Unauthorized("请先登录"))
		return
	}
	var req dto.CreateLoveReportRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	loveIDStr := ctx.Query("love_id")
	loveID, _ := parseUint(loveIDStr)
	info, err := h.service.Create(userID, loveID, username, avatar, &req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeLoveReportError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("举报成功", info))
}

// GetByID 举报详情
// GET /api/v1/love/reports/:id  （需登录）
func (h *ReportHandler) GetByID(ctx plugin.Context) {
	userID, ok := requireLogin(ctx)
	if !ok {
		return
	}
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	info, err := h.service.GetByID(id, userID)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeLoveReportError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(info))
}

// ListByReporter 我的举报列表
// GET /api/v1/love/reports/mine  （需登录）
func (h *ReportHandler) ListByReporter(ctx plugin.Context) {
	userID, ok := requireLogin(ctx)
	if !ok {
		return
	}
	var req dto.LoveReportListRequest
	_ = ctx.Bind(&req)
	pagination, list, err := h.service.ListByReporter(userID, &req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeLoveReportError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// Appeal 提交申诉
// POST /api/v1/love/reports/:id/appeal  （需登录）
func (h *ReportHandler) Appeal(ctx plugin.Context) {
	userID, ok := requireLogin(ctx)
	if !ok {
		return
	}
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	var req dto.LoveReportAppealRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	req.ID = id
	if err := h.service.Appeal(&req, userID); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeLoveReportError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("申诉成功", nil))
}

// ===== M 端 =====

// List 管理后台：举报列表
// GET /api/v1/love/admin/reports  （需 love:audit 权限）
func (h *ReportHandler) List(ctx plugin.Context) {
	var req dto.LoveReportListRequest
	_ = ctx.Bind(&req)
	pagination, list, err := h.service.List(&req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeLoveReportError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// ListPending 管理后台：待处理举报
// GET /api/v1/love/admin/reports/pending  （需 love:audit 权限）
func (h *ReportHandler) ListPending(ctx plugin.Context) {
	var req dto.LoveReportListRequest
	_ = ctx.Bind(&req)
	pagination, list, err := h.service.ListPending(&req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeLoveReportError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// ListByTarget 管理后台：按被举报人查询
// GET /api/v1/love/admin/reports/by-target  （需 love:audit 权限）
func (h *ReportHandler) ListByTarget(ctx plugin.Context) {
	userIDStr := ctx.Query("user_id")
	targetUserID, err := parseUint(userIDStr)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("user_id 无效"))
		return
	}
	var req dto.LoveReportListRequest
	_ = ctx.Bind(&req)
	pagination, list, err := h.service.ListByTarget(targetUserID, &req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeLoveReportError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// Handle 处理举报
// PUT /api/v1/love/admin/reports/:id/handle  （需 love:audit 权限）
func (h *ReportHandler) Handle(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	var req dto.LoveReportHandleRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	req.ID = id
	// handledBy 取自 JWT 上下文
	handledBy, _ := ctx.Get(middleware.ContextUserID)
	var uid uint
	if v, ok := handledBy.(uint); ok {
		uid = v
	}
	if err := h.service.Handle(&req, uid); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeLoveReportError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("处理成功", nil))
}

// HandleAppeal 处理申诉
// PUT /api/v1/love/admin/reports/:id/appeal  （需 love:audit 权限）
func (h *ReportHandler) HandleAppeal(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	var req dto.LoveReportAppealHandleRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	req.ID = id
	handledBy, _ := ctx.Get(middleware.ContextUserID)
	var uid uint
	if v, ok := handledBy.(uint); ok {
		uid = v
	}
	if err := h.service.HandleAppeal(&req, uid); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeLoveReportError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("申诉处理成功", nil))
}

// UpdateRiskScore 更新风险分
// PUT /api/v1/love/admin/reports/:id/risk-score  （需 love:audit 权限）
func (h *ReportHandler) UpdateRiskScore(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	var req struct {
		Score int `json:"score" binding:"min=0,max=100"`
	}
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	if err := h.service.UpdateRiskScore(id, req.Score); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeLoveReportError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("更新成功", nil))
}

// Delete 删除举报
// DELETE /api/v1/love/admin/reports/:id  （需 love:audit 权限）
func (h *ReportHandler) Delete(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	if err := h.service.Delete(id); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeLoveReportError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("删除成功", nil))
}

// Stats 举报统计
// GET /api/v1/love/admin/reports/stats  （需 love:audit 权限）
func (h *ReportHandler) Stats(ctx plugin.Context) {
	resp, err := h.service.Stats()
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeLoveReportError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(resp))
}
