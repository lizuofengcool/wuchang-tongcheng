// Package handler 风控审核中台精简版HTTP处理层
package handler

import (
	"net/http"
	"strconv"

	"wuchang-tongcheng/internal/core/middleware"
	"wuchang-tongcheng/internal/core/plugin"
	"wuchang-tongcheng/internal/core/response"
	"wuchang-tongcheng/internal/modules/risk/dto"
	"wuchang-tongcheng/internal/modules/risk/service"
)

// Handler 风控中台 HTTP 处理器
type Handler struct {
	svc service.RiskService
}

// NewHandler 创建 Handler 实例
func NewHandler(svc service.RiskService) *Handler {
	return &Handler{svc: svc}
}

// getUserID 从上下文获取登录用户ID
func getUserID(ctx plugin.Context) uint {
	if v, ok := ctx.Get(middleware.ContextUserID); ok {
		if id, ok := v.(uint); ok {
			return id
		}
	}
	return 0
}

// getRegionID 从上下文获取地区ID
func getRegionID(ctx plugin.Context) uint {
	if v, ok := ctx.Get(middleware.RegionIDKey); ok {
		if id, ok := v.(uint); ok {
			return id
		}
	}
	return middleware.DefaultRegionID
}

// parsePagination 解析分页
func parsePagination(ctx plugin.Context) (page, pageSize int) {
	page, _ = strconv.Atoi(ctx.DefaultQuery("page", "1"))
	pageSize, _ = strconv.Atoi(ctx.DefaultQuery("page_size", "20"))
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	return
}

// CreateReport 创建举报
// POST /api/v1/risk/reports （需登录）
func (h *Handler) CreateReport(ctx plugin.Context) {
	userID := getUserID(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusOK, response.Unauthorized("请先登录"))
		return
	}
	var req dto.ReportRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	info, err := h.svc.CreateReport(getRegionID(ctx), userID, &req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(3001, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("举报成功", info))
}

// GetReport 查询举报
// GET /api/v1/risk/reports/:id
func (h *Handler) GetReport(ctx plugin.Context) {
	idStr := ctx.Param("id")
	id, _ := strconv.ParseUint(idStr, 10, 32)
	if id == 0 {
		ctx.JSON(http.StatusOK, response.BadRequest("id 不能为空"))
		return
	}
	info, err := h.svc.GetReport(uint(id))
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(3002, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(info))
}

// ListReports 举报列表（M 端）
// GET /api/v1/risk/reports
func (h *Handler) ListReports(ctx plugin.Context) {
	page, pageSize := parsePagination(ctx)
	status, _ := strconv.Atoi(ctx.DefaultQuery("status", "-1"))
	req := &dto.ReportListRequest{
		Status:     status,
		ReportType: ctx.Query("report_type"),
		BizModule:  ctx.Query("biz_module"),
		Page:       page,
		PageSize:   pageSize,
	}
	list, total, err := h.svc.ListReports(req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(3001, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(response.NewPageResult(list, total, page, pageSize)))
}

// HandleReport 处理举报（M 端）
// POST /api/v1/risk/reports/handle
func (h *Handler) HandleReport(ctx plugin.Context) {
	userID := getUserID(ctx)
	var req dto.HandleReportRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	if err := h.svc.HandleReport(userID, &req); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(3001, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("处理成功", nil))
}

// AddSensitiveWord 添加敏感词（M 端）
// POST /api/v1/risk/sensitive-words
func (h *Handler) AddSensitiveWord(ctx plugin.Context) {
	var req dto.SensitiveWordRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	if err := h.svc.AddSensitiveWord(&req); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(3001, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("添加成功", nil))
}

// DeleteSensitiveWord 删除敏感词（M 端）
// DELETE /api/v1/risk/sensitive-words/:id
func (h *Handler) DeleteSensitiveWord(ctx plugin.Context) {
	idStr := ctx.Param("id")
	id, _ := strconv.ParseUint(idStr, 10, 32)
	if id == 0 {
		ctx.JSON(http.StatusOK, response.BadRequest("id 不能为空"))
		return
	}
	if err := h.svc.DeleteSensitiveWord(uint(id)); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(3001, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("删除成功", nil))
}

// ListSensitiveWords 敏感词列表（M 端）
// GET /api/v1/risk/sensitive-words
func (h *Handler) ListSensitiveWords(ctx plugin.Context) {
	page, pageSize := parsePagination(ctx)
	wordType := ctx.Query("word_type")
	list, total, err := h.svc.ListSensitiveWords(wordType, page, pageSize)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(3001, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(response.NewPageResult(list, total, page, pageSize)))
}

// CheckText 文本审核（敏感词检测）
// POST /api/v1/risk/check-text
func (h *Handler) CheckText(ctx plugin.Context) {
	var req dto.CheckTextRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	resp, err := h.svc.CheckText(&req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(3001, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(resp))
}

// AuditContent 综合内容审核（供其他模块调用）
// POST /api/v1/risk/audit
func (h *Handler) AuditContent(ctx plugin.Context) {
	var req dto.AuditResultRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	resp, err := h.svc.AuditContent(getRegionID(ctx), &req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(3001, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(resp))
}

// AddToBlacklist 加入黑名单（M 端）
// POST /api/v1/risk/blacklist
func (h *Handler) AddToBlacklist(ctx plugin.Context) {
	userID := getUserID(ctx)
	var req dto.BlacklistRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	if err := h.svc.AddToBlacklist(userID, &req); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(3001, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("加入黑名单成功", nil))
}

// CheckBlacklist 检查黑名单
// POST /api/v1/risk/blacklist/check
func (h *Handler) CheckBlacklist(ctx plugin.Context) {
	var req dto.BlacklistCheckRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	in, err := h.svc.CheckBlacklist(req.TargetType, req.TargetValue)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(3001, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(map[string]bool{"in_blacklist": in}))
}

// ListBlacklist 黑名单列表（M 端）
// GET /api/v1/risk/blacklist
func (h *Handler) ListBlacklist(ctx plugin.Context) {
	page, pageSize := parsePagination(ctx)
	targetType := ctx.Query("target_type")
	list, total, err := h.svc.ListBlacklist(targetType, page, pageSize)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(3001, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(response.NewPageResult(list, total, page, pageSize)))
}

// GetUserScore 查询用户风险分
// GET /api/v1/risk/scores/:user_id
func (h *Handler) GetUserScore(ctx plugin.Context) {
	userIDStr := ctx.Param("user_id")
	userID, _ := strconv.ParseUint(userIDStr, 10, 32)
	if userID == 0 {
		ctx.JSON(http.StatusOK, response.BadRequest("user_id 不能为空"))
		return
	}
	info, err := h.svc.GetUserScore(uint(userID))
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(3002, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(info))
}

// ListUserViolations 用户违规列表
// GET /api/v1/risk/violations/:user_id
func (h *Handler) ListUserViolations(ctx plugin.Context) {
	userIDStr := ctx.Param("user_id")
	userID, _ := strconv.ParseUint(userIDStr, 10, 32)
	if userID == 0 {
		ctx.JSON(http.StatusOK, response.BadRequest("user_id 不能为空"))
		return
	}
	page, pageSize := parsePagination(ctx)
	list, total, err := h.svc.ListUserViolations(uint(userID), page, pageSize)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(3001, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(response.NewPageResult(list, total, page, pageSize)))
}

// AppealViolation 申诉
// POST /api/v1/risk/violations/appeal
func (h *Handler) AppealViolation(ctx plugin.Context) {
	userID := getUserID(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusOK, response.Unauthorized("请先登录"))
		return
	}
	var req dto.AppealRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	if err := h.svc.AppealViolation(userID, &req); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(3001, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("申诉已提交", nil))
}
