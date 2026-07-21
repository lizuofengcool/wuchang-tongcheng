// Package handler 同城商城 HTTP 处理层 - 审核规则
package handler

import (
	"net/http"

	"wuchang-tongcheng/internal/core/plugin"
	"wuchang-tongcheng/internal/core/response"
	"wuchang-tongcheng/internal/modules/mall/dto"
	"wuchang-tongcheng/internal/modules/mall/service"
	"wuchang-tongcheng/internal/pkg/utils"
)

// AuditRuleHandler 审核规则 HTTP 处理器
type AuditRuleHandler struct {
	svc service.AuditRuleService
}

// NewAuditRuleHandler 创建审核规则 Handler 实例
func NewAuditRuleHandler(svc service.AuditRuleService) *AuditRuleHandler {
	return &AuditRuleHandler{svc: svc}
}

// Create 创建审核规则
// POST /api/v1/mall/admin/audit-rules  （需 mall:audit 权限）
func (h *AuditRuleHandler) Create(ctx plugin.Context) {
	var req dto.CreateAuditRuleRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	info, err := h.svc.Create(&req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeMallAuditRuleError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("创建成功", info))
}

// Update 更新审核规则
// PUT /api/v1/mall/admin/audit-rules/:id  （需 mall:audit 权限）
func (h *AuditRuleHandler) Update(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	var req dto.UpdateAuditRuleRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	if err := h.svc.Update(id, &req); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeMallAuditRuleError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("更新成功", nil))
}

// Delete 删除审核规则
// DELETE /api/v1/mall/admin/audit-rules/:id  （需 mall:audit 权限）
func (h *AuditRuleHandler) Delete(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	if err := h.svc.Delete(id); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeMallAuditRuleNotFound, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("删除成功", nil))
}

// GetByID 审核规则详情
// GET /api/v1/mall/audit-rules/:id  （公开只读）
func (h *AuditRuleHandler) GetByID(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	info, err := h.svc.GetByID(id)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeMallAuditRuleNotFound, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(info))
}

// List 审核规则列表
// GET /api/v1/mall/audit-rules  （公开只读）
func (h *AuditRuleHandler) List(ctx plugin.Context) {
	var req dto.AuditRuleListRequest
	_ = ctx.Bind(&req)
	if req.Page == 0 {
		req.Page, req.PageSize = parsePagination(ctx)
	}
	pagination, list, err := h.svc.List(&req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeMallAuditRuleError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// ListEnabled 列出已启用的审核规则
// GET /api/v1/mall/audit-rules/enabled  （公开只读）
func (h *AuditRuleHandler) ListEnabled(ctx plugin.Context) {
	list, err := h.svc.ListEnabled()
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeMallAuditRuleError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(list))
}

// ListByType 按类型列出审核规则
// GET /api/v1/mall/audit-rules/type/:rule_type  （公开只读）
func (h *AuditRuleHandler) ListByType(ctx plugin.Context) {
	ruleType := ctx.Param("rule_type")
	if ruleType == "" {
		ctx.JSON(http.StatusOK, response.BadRequest("规则类型不能为空"))
		return
	}
	list, err := h.svc.ListByType(ruleType)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeMallAuditRuleError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(list))
}

// UpdateStatus 更新审核规则状态
// PUT /api/v1/mall/admin/audit-rules/:id/status  （需 mall:audit 权限）
func (h *AuditRuleHandler) UpdateStatus(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	var req dto.UpdateStatusRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	if err := h.svc.UpdateStatus(id, req.Status); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeMallAuditRuleError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("状态已更新", nil))
}

// Check 内容审核检测
// POST /api/v1/mall/admin/audit-rules/check  （需 mall:audit 权限）
func (h *AuditRuleHandler) Check(ctx plugin.Context) {
	var req dto.AuditCheckRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	resp, err := h.svc.Check(&req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeMallAuditRuleError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(resp))
}
