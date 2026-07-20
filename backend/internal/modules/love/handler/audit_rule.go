// Package handler love 相亲交友 HTTP 处理层 - 审核规则
// 依据需求文档 1.10：4 维数据隔离（region_id + user_id）
package handler

import (
	"net/http"

	"wuchang-tongcheng/internal/core/plugin"
	"wuchang-tongcheng/internal/core/response"
	"wuchang-tongcheng/internal/modules/love/dto"
	"wuchang-tongcheng/internal/modules/love/service"
	"wuchang-tongcheng/internal/pkg/utils"
)

// AuditRuleHandler 审核规则 HTTP 处理器（M 端配置 + C 端检查）
type AuditRuleHandler struct {
	service service.LoveAuditRuleService
}

// NewAuditRuleHandler 创建 AuditRuleHandler 实例
func NewAuditRuleHandler(svc service.LoveAuditRuleService) *AuditRuleHandler {
	return &AuditRuleHandler{service: svc}
}

// Create 创建审核规则
// POST /api/v1/love/admin/audit-rules  （需 love:audit 权限）
func (h *AuditRuleHandler) Create(ctx plugin.Context) {
	var req dto.CreateLoveAuditRuleRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	info, err := h.service.Create(&req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeLoveError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("创建成功", info))
}

// Update 更新审核规则
// PUT /api/v1/love/admin/audit-rules/:id  （需 love:audit 权限）
func (h *AuditRuleHandler) Update(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	var req dto.UpdateLoveAuditRuleRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	if err := h.service.Update(id, &req); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeLoveError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("更新成功", nil))
}

// Delete 删除审核规则
// DELETE /api/v1/love/admin/audit-rules/:id  （需 love:audit 权限）
func (h *AuditRuleHandler) Delete(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	if err := h.service.Delete(id); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeLoveError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("删除成功", nil))
}

// GetByID 审核规则详情
// GET /api/v1/love/audit-rules/:id  （公开）
func (h *AuditRuleHandler) GetByID(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	info, err := h.service.GetByID(id)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeLoveError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(info))
}

// GetByRuleKey 按 rule_key 查询
// GET /api/v1/love/audit-rules/key/:key  （公开）
func (h *AuditRuleHandler) GetByRuleKey(ctx plugin.Context) {
	key := ctx.Param("key")
	if key == "" {
		ctx.JSON(http.StatusOK, response.BadRequest("key 不能为空"))
		return
	}
	info, err := h.service.GetByRuleKey(key)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeLoveError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(info))
}

// List 审核规则列表
// GET /api/v1/love/audit-rules  （公开）
func (h *AuditRuleHandler) List(ctx plugin.Context) {
	var req dto.LoveAuditRuleListRequest
	_ = ctx.Bind(&req)
	pagination, list, err := h.service.List(&req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeLoveError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// ListAll 全部启用审核规则
// GET /api/v1/love/audit-rules/all  （公开）
func (h *AuditRuleHandler) ListAll(ctx plugin.Context) {
	list, err := h.service.ListAll()
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeLoveError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(list))
}

// UpdateStatus 更新审核规则状态
// PUT /api/v1/love/admin/audit-rules/:id/status  （需 love:audit 权限）
func (h *AuditRuleHandler) UpdateStatus(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	var req struct {
		Status int `json:"status" binding:"required,oneof=0 1"`
	}
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	if err := h.service.UpdateStatus(id, req.Status); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeLoveError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("更新成功", nil))
}

// BatchUpdateStatus 批量更新状态
// PUT /api/v1/love/admin/audit-rules/batch-status  （需 love:audit 权限）
func (h *AuditRuleHandler) BatchUpdateStatus(ctx plugin.Context) {
	var req struct {
		IDs    []uint `json:"ids" binding:"required,min=1"`
		Status int    `json:"status" binding:"required,oneof=0 1"`
	}
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	if err := h.service.BatchUpdateStatus(req.IDs, req.Status); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeLoveError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("批量更新成功", nil))
}

// Check 内容审核检查
// POST /api/v1/love/audit-rules/check  （公开）
func (h *AuditRuleHandler) Check(ctx plugin.Context) {
	var req dto.LoveAuditCheckRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	resp, err := h.service.Check(&req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeLoveError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(resp))
}
