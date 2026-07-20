// Package handler 同城拼车出行 HTTP 处理层 - 审核规则
// 依据 v3.2.1 架构方案：对标哈啰出行/嘀嗒出行/滴滴顺风车
package handler

import (
	"net/http"

	"wuchang-tongcheng/internal/core/plugin"
	"wuchang-tongcheng/internal/core/response"
	"wuchang-tongcheng/internal/modules/pinche/dto"
	"wuchang-tongcheng/internal/modules/pinche/service"
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
// GET /api/v1/pinche/audit-rules/:id  （公开）
func (h *AuditRuleHandler) GetByID(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	info, err := h.service.GetByID(id)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodePincheAuditRuleError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(info))
}

// List 审核规则列表
// GET /api/v1/pinche/audit-rules  （公开）
func (h *AuditRuleHandler) List(ctx plugin.Context) {
	var req dto.AuditRuleListRequest
	_ = ctx.Bind(&req)

	regionID := getRegionID(ctx)
	pagination, list, err := h.service.List(regionID, &req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodePincheAuditRuleError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// ListEnabled 启用中的规则列表
// GET /api/v1/pinche/audit-rules/enabled  （公开）
func (h *AuditRuleHandler) ListEnabled(ctx plugin.Context) {
	regionID := getRegionID(ctx)
	list, err := h.service.ListEnabled(regionID)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodePincheAuditRuleError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(list))
}

// ListByType 按类型查询规则
// GET /api/v1/pinche/audit-rules/type/:rule_type  （公开）
func (h *AuditRuleHandler) ListByType(ctx plugin.Context) {
	ruleType := ctx.Param("rule_type")
	if ruleType == "" {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的规则类型"))
		return
	}
	regionID := getRegionID(ctx)
	list, err := h.service.ListByType(regionID, ruleType)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodePincheAuditRuleError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(list))
}

// ===== M 端管理 =====

// AdminList 管理后台审核规则列表（跨地区）
// GET /api/v1/pinche/admin/audit-rules  （需 pinche:audit 权限）
func (h *AuditRuleHandler) AdminList(ctx plugin.Context) {
	var req dto.AuditRuleListRequest
	_ = ctx.Bind(&req)

	// 跨地区：regionID=0
	pagination, list, err := h.service.List(0, &req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodePincheAuditRuleError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// Create 创建审核规则
// POST /api/v1/pinche/admin/audit-rules  （需 pinche:audit 权限）
func (h *AuditRuleHandler) Create(ctx plugin.Context) {
	var req dto.CreateAuditRuleRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	regionID := getRegionID(ctx)
	info, err := h.service.Create(regionID, &req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodePincheAuditRuleError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("规则创建成功", info))
}

// Update 更新审核规则
// PUT /api/v1/pinche/admin/audit-rules/:id  （需 pinche:audit 权限）
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
		ctx.JSON(http.StatusOK, response.Fail(utils.CodePincheAuditRuleError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("更新成功", nil))
}

// Delete 删除审核规则
// DELETE /api/v1/pinche/admin/audit-rules/:id  （需 pinche:audit 权限）
func (h *AuditRuleHandler) Delete(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	if err := h.service.Delete(id); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodePincheAuditRuleError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("删除成功", nil))
}

// UpdateStatus 更新审核规则状态（启用/禁用）
// PUT /api/v1/pinche/admin/audit-rules/:id/status  （需 pinche:audit 权限）
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
	if err := h.service.UpdateStatus(id, req.Status); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodePincheStatusInvalid, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("状态更新成功", nil))
}

// IncrHitCount 命中次数 +1（内部调用）
// POST /api/v1/pinche/admin/audit-rules/:id/hit  （需 pinche:audit 权限）
func (h *AuditRuleHandler) IncrHitCount(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	if err := h.service.IncrHitCount(id); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodePincheAuditRuleError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("已记录命中", nil))
}
