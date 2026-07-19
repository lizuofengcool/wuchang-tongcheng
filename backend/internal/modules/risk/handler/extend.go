// Package handler 风控中台扩展 HTTP 处理层
// 依据 015_risk_full.sql：证据/申诉/规则/评分记录/审核日志
package handler

import (
	"net/http"
	"strconv"

	"wuchang-tongcheng/internal/core/plugin"
	"wuchang-tongcheng/internal/core/response"
	"wuchang-tongcheng/internal/modules/risk/dto"
	"wuchang-tongcheng/internal/modules/risk/service"
	"wuchang-tongcheng/internal/pkg/utils"
)

// ExtendHandler 风控中台扩展处理器
type ExtendHandler struct {
	extSvc service.RiskExtendService
}

// NewExtendHandler 创建扩展 Handler 实例
func NewExtendHandler(extSvc service.RiskExtendService) *ExtendHandler {
	return &ExtendHandler{extSvc: extSvc}
}

// ===== 举报证据 =====

// AddEvidence 添加举报证据
// POST /api/v1/risk/reports/evidence
func (h *ExtendHandler) AddEvidence(ctx plugin.Context) {
	userID := getUserID(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusOK, response.Unauthorized("请先登录"))
		return
	}
	var req dto.AddEvidenceRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	info, err := h.extSvc.AddEvidence(getRegionID(ctx), userID, &req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeRiskError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("证据添加成功", info))
}

// ListEvidenceByReport 查询举报证据
// GET /api/v1/risk/reports/:id/evidence
func (h *ExtendHandler) ListEvidenceByReport(ctx plugin.Context) {
	reportID, _ := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if reportID == 0 {
		ctx.JSON(http.StatusOK, response.BadRequest("举报ID无效"))
		return
	}
	list, err := h.extSvc.ListEvidenceByReport(uint(reportID))
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeRiskEvidenceNotFound, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(list))
}

// DeleteEvidence 删除证据（M 端）
// DELETE /api/v1/risk/reports/evidence/:id
func (h *ExtendHandler) DeleteEvidence(ctx plugin.Context) {
	id, _ := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if id == 0 {
		ctx.JSON(http.StatusOK, response.BadRequest("证据ID无效"))
		return
	}
	if err := h.extSvc.DeleteEvidence(uint(id)); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeRiskEvidenceNotFound, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("删除成功", nil))
}

// ===== 申诉 =====

// CreateAppeal 创建申诉
// POST /api/v1/risk/appeals
func (h *ExtendHandler) CreateAppeal(ctx plugin.Context) {
	userID := getUserID(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusOK, response.Unauthorized("请先登录"))
		return
	}
	var req dto.CreateAppealRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	info, err := h.extSvc.CreateAppeal(getRegionID(ctx), userID, &req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeRiskAppealNotFound, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("申诉已提交", info))
}

// ListMyAppeals 我的申诉列表
// GET /api/v1/risk/appeals
func (h *ExtendHandler) ListMyAppeals(ctx plugin.Context) {
	userID := getUserID(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusOK, response.Unauthorized("请先登录"))
		return
	}
	page, pageSize := parsePagination(ctx)
	list, total, err := h.extSvc.ListMyAppeals(userID, page, pageSize)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeRiskError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(response.NewPageResult(list, total, page, pageSize)))
}

// ListAppeals 申诉列表（M 端）
// GET /api/v1/risk/admin/appeals
func (h *ExtendHandler) ListAppeals(ctx plugin.Context) {
	status, _ := strconv.Atoi(ctx.DefaultQuery("status", "-1"))
	page, pageSize := parsePagination(ctx)
	list, total, err := h.extSvc.ListAppeals(status, page, pageSize)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeRiskError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(response.NewPageResult(list, total, page, pageSize)))
}

// HandleAppeal 处理申诉（M 端）
// POST /api/v1/risk/admin/appeals/handle
func (h *ExtendHandler) HandleAppeal(ctx plugin.Context) {
	handlerID := getUserID(ctx)
	var req dto.HandleAppealRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	if err := h.extSvc.HandleAppeal(handlerID, &req); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeRiskAppealNotFound, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("申诉已处理", nil))
}

// ===== 风控规则 =====

// CreateRule 创建风控规则（M 端）
// POST /api/v1/risk/admin/rules
func (h *ExtendHandler) CreateRule(ctx plugin.Context) {
	var req dto.CreateRuleRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	info, err := h.extSvc.CreateRule(getRegionID(ctx), &req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeRiskRuleNotFound, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("规则创建成功", info))
}

// UpdateRule 更新风控规则（M 端）
// POST /api/v1/risk/admin/rules/:id
func (h *ExtendHandler) UpdateRule(ctx plugin.Context) {
	id, _ := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if id == 0 {
		ctx.JSON(http.StatusOK, response.BadRequest("规则ID无效"))
		return
	}
	var req dto.UpdateRuleRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	if err := h.extSvc.UpdateRule(uint(id), &req); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeRiskRuleNotFound, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("更新成功", nil))
}

// DeleteRule 删除风控规则（M 端）
// DELETE /api/v1/risk/admin/rules/:id
func (h *ExtendHandler) DeleteRule(ctx plugin.Context) {
	id, _ := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if id == 0 {
		ctx.JSON(http.StatusOK, response.BadRequest("规则ID无效"))
		return
	}
	if err := h.extSvc.DeleteRule(uint(id)); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeRiskRuleNotFound, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("删除成功", nil))
}

// ListRules 风控规则列表
// GET /api/v1/risk/rules
func (h *ExtendHandler) ListRules(ctx plugin.Context) {
	ruleType := ctx.Query("rule_type")
	page, pageSize := parsePagination(ctx)
	list, total, err := h.extSvc.ListRules(ruleType, page, pageSize)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeRiskError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(response.NewPageResult(list, total, page, pageSize)))
}

// GetRule 查询规则详情
// GET /api/v1/risk/rules/:id
func (h *ExtendHandler) GetRule(ctx plugin.Context) {
	id, _ := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if id == 0 {
		ctx.JSON(http.StatusOK, response.BadRequest("规则ID无效"))
		return
	}
	info, err := h.extSvc.GetRule(uint(id))
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeRiskRuleNotFound, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(info))
}

// ===== 风险评分记录 =====

// ListMyScoreRecords 我的风险评分记录
// GET /api/v1/risk/score-records
func (h *ExtendHandler) ListMyScoreRecords(ctx plugin.Context) {
	userID := getUserID(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusOK, response.Unauthorized("请先登录"))
		return
	}
	page, pageSize := parsePagination(ctx)
	list, total, err := h.extSvc.ListMyScoreRecords(userID, page, pageSize)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeRiskError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(response.NewPageResult(list, total, page, pageSize)))
}

// ListScoreRecordsByLevel 按等级查询评分记录（M 端）
// GET /api/v1/risk/admin/score-records
func (h *ExtendHandler) ListScoreRecordsByLevel(ctx plugin.Context) {
	level := ctx.DefaultQuery("level", "")
	page, pageSize := parsePagination(ctx)
	list, total, err := h.extSvc.ListScoreRecordsByLevel(level, page, pageSize)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeRiskError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(response.NewPageResult(list, total, page, pageSize)))
}

// ===== 审核日志 =====

// ListAuditLogs 审核日志列表（M 端）
// GET /api/v1/risk/admin/audit-logs
func (h *ExtendHandler) ListAuditLogs(ctx plugin.Context) {
	auditorID, _ := strconv.ParseUint(ctx.Query("auditor_id"), 10, 32)
	action := ctx.Query("action")
	targetType := ctx.Query("target_type")
	page, pageSize := parsePagination(ctx)
	list, total, err := h.extSvc.ListAuditLogs(uint(auditorID), action, targetType, page, pageSize)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeRiskAuditLogNotFound, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(response.NewPageResult(list, total, page, pageSize)))
}

// ===== 统计 =====

// Statistics 风控统计（M 端）
// GET /api/v1/risk/admin/statistics
func (h *ExtendHandler) Statistics(ctx plugin.Context) {
	resp, err := h.extSvc.Statistics()
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeRiskError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(resp))
}
