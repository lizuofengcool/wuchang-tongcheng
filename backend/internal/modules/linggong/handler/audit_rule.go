// Package handler 同城零工兼职 HTTP 处理层 - 审核规则
// 注意：BaseModel 无 region_id，使用全局规则
package handler

import (
	"net/http"

	"wuchang-tongcheng/internal/core/plugin"
	"wuchang-tongcheng/internal/core/response"
	"wuchang-tongcheng/internal/modules/linggong/dto"
	"wuchang-tongcheng/internal/modules/linggong/service"
	"wuchang-tongcheng/internal/pkg/utils"
)

// AuditRuleHandler 审核规则 HTTP 处理器
type AuditRuleHandler struct {
	service service.AuditRuleService
}

// NewAuditRuleHandler 创建 AuditRuleHandler 实例
func NewAuditRuleHandler(svc service.AuditRuleService) *AuditRuleHandler {
	return &AuditRuleHandler{service: svc}
}

// ===== C 端只读 =====

// GetByID 审核规则详情
// GET /api/v1/linggong/audit-rules/:id  （公开）
func (h *AuditRuleHandler) GetByID(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	info, err := h.service.GetByID(id)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeLinggongAuditRuleNotFound, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(info))
}

// List 审核规则列表
// GET /api/v1/linggong/audit-rules  （公开）
func (h *AuditRuleHandler) List(ctx plugin.Context) {
	var req dto.AuditRuleListRequest
	_ = ctx.Bind(&req)

	pagination, list, err := h.service.List(&req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeLinggongAuditRuleNotFound, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// ListEnabled 启用中的规则列表
// GET /api/v1/linggong/audit-rules/enabled  （公开）
func (h *AuditRuleHandler) ListEnabled(ctx plugin.Context) {
	list, err := h.service.ListEnabled()
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeLinggongAuditRuleNotFound, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(list))
}

// ListByType 按类型查询规则
// GET /api/v1/linggong/audit-rules/type/:rule_type  （公开）
func (h *AuditRuleHandler) ListByType(ctx plugin.Context) {
	ruleType := ctx.Param("rule_type")
	if ruleType == "" {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的规则类型"))
		return
	}
	list, err := h.service.ListByRuleType(ruleType)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeLinggongAuditRuleNotFound, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(list))
}

// ===== M 端管理 =====

// AdminList 管理后台审核规则列表
// GET /api/v1/linggong/admin/audit-rules  （需 linggong:audit 权限）
func (h *AuditRuleHandler) AdminList(ctx plugin.Context) {
	var req dto.AuditRuleListRequest
	_ = ctx.Bind(&req)

	pagination, list, err := h.service.List(&req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeLinggongAuditRuleNotFound, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// Create 创建审核规则
// POST /api/v1/linggong/admin/audit-rules  （需 linggong:audit 权限）
func (h *AuditRuleHandler) Create(ctx plugin.Context) {
	var req dto.CreateAuditRuleRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	info, err := h.service.Create(&req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeLinggongAuditRuleNotFound, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("规则创建成功", info))
}

// Update 更新审核规则
// PUT /api/v1/linggong/admin/audit-rules/:id  （需 linggong:audit 权限）
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
	if err := h.service.Update(id, &req); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeLinggongAuditRuleNotFound, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("更新成功", nil))
}

// Delete 删除审核规则
// DELETE /api/v1/linggong/admin/audit-rules/:id  （需 linggong:audit 权限）
func (h *AuditRuleHandler) Delete(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	if err := h.service.Delete(id); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeLinggongAuditRuleNotFound, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("删除成功", nil))
}

// AdminUpdateStatus 更新审核规则状态（启用/禁用）
// PUT /api/v1/linggong/admin/audit-rules/:id/status  （需 linggong:audit 权限）
func (h *AuditRuleHandler) AdminUpdateStatus(ctx plugin.Context) {
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
	if err := h.service.UpdateStatus(id, req.Status); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeLinggongStatusInvalid, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("状态更新成功", nil))
}

// BatchDelete 批量删除审核规则
// POST /api/v1/linggong/admin/audit-rules/batch-delete  （需 linggong:audit 权限）
func (h *AuditRuleHandler) BatchDelete(ctx plugin.Context) {
	var req dto.BatchDeleteRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	if err := h.service.BatchDelete(req.IDs); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeLinggongAuditRuleNotFound, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("批量删除完成", nil))
}
