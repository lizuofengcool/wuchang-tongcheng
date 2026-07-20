// Package handler 同城114 HTTP 处理层 - 审核规则 + 推荐商家
// 依据 v3.2.1 架构方案：敏感词/违禁内容/联系方式/价格校验/频率
// 推荐商家：首页推荐/分类推荐/附近推荐/个性化推荐
package handler

import (
	"net/http"

	"wuchang-tongcheng/internal/core/plugin"
	"wuchang-tongcheng/internal/core/response"
	"wuchang-tongcheng/internal/modules/dh114/dto"
	"wuchang-tongcheng/internal/modules/dh114/service"
	"wuchang-tongcheng/internal/pkg/utils"
)

// =====================================================================
// AuditRuleHandler 审核规则
// =====================================================================

// AuditRuleHandler 审核规则 HTTP 处理器
type AuditRuleHandler struct {
	svc service.AuditRuleService
}

// NewAuditRuleHandler 创建审核规则 Handler 实例
func NewAuditRuleHandler(svc service.AuditRuleService) *AuditRuleHandler {
	return &AuditRuleHandler{svc: svc}
}

// Create 创建审核规则（M 端）
// POST /api/v1/dh114/admin/audit-rules  （需 dh114:audit 权限）
func (h *AuditRuleHandler) Create(ctx plugin.Context) {
	var req dto.CreateAuditRuleRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	info, err := h.svc.Create(&req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeDh114AuditRuleError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("审核规则创建成功", info))
}

// Update 更新审核规则
// PUT /api/v1/dh114/admin/audit-rules/:id  （需 dh114:audit 权限）
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
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeDh114AuditRuleError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("更新成功", nil))
}

// Delete 删除审核规则
// DELETE /api/v1/dh114/admin/audit-rules/:id  （需 dh114:audit 权限）
func (h *AuditRuleHandler) Delete(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	if err := h.svc.Delete(id); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeDh114AuditRuleError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("删除成功", nil))
}

// GetByID 审核规则详情
// GET /api/v1/dh114/audit-rules/:id  （公开）
func (h *AuditRuleHandler) GetByID(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	info, err := h.svc.GetByID(id)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeDh114AuditRuleNotFound, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(info))
}

// List 审核规则列表
// GET /api/v1/dh114/audit-rules  （公开）
func (h *AuditRuleHandler) List(ctx plugin.Context) {
	var req dto.AuditRuleListRequest
	_ = ctx.Bind(&req)
	if req.Page == 0 {
		req.Page, req.PageSize = parsePagination(ctx)
	}
	pagination, list, err := h.svc.List(&req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeDh114AuditRuleError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// ListEnabled 列出已启用的审核规则
// GET /api/v1/dh114/audit-rules/enabled  （公开）
func (h *AuditRuleHandler) ListEnabled(ctx plugin.Context) {
	list, err := h.svc.ListEnabled()
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeDh114AuditRuleError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(list))
}

// ListByType 按类型列出审核规则
// GET /api/v1/dh114/audit-rules/type/:rule_type  （公开）
func (h *AuditRuleHandler) ListByType(ctx plugin.Context) {
	ruleType := ctx.Param("rule_type")
	if ruleType == "" {
		ctx.JSON(http.StatusOK, response.BadRequest("rule_type 参数不能为空"))
		return
	}
	list, err := h.svc.ListByType(ruleType)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeDh114AuditRuleError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(list))
}

// UpdateStatus 更新审核规则状态
// PUT /api/v1/dh114/admin/audit-rules/:id/status  （需 dh114:audit 权限）
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
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeDh114AuditRuleError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("状态已更新", nil))
}

// =====================================================================
// RecommendationHandler 推荐商家
// =====================================================================

// RecommendationHandler 推荐商家 HTTP 处理器
type RecommendationHandler struct {
	svc service.RecommendationService
}

// NewRecommendationHandler 创建推荐 Handler 实例
func NewRecommendationHandler(svc service.RecommendationService) *RecommendationHandler {
	return &RecommendationHandler{svc: svc}
}

// ListByType 按推荐类型列出推荐（C 端）
// GET /api/v1/dh114/recommendations  （公开）
func (h *RecommendationHandler) ListByType(ctx plugin.Context) {
	recommendType := ctx.DefaultQuery("recommend_type", "home")
	regionID := getRegionID(ctx)
	page, pageSize := parsePagination(ctx)
	pagination, list, err := h.svc.ListByType(recommendType, regionID, page, pageSize)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeDh114RecommendationError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// ListByUser 按用户列出推荐
// GET /api/v1/dh114/recommendations/mine  （需登录）
func (h *RecommendationHandler) ListByUser(ctx plugin.Context) {
	userID, _, _, _ := getUserProfile(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusOK, response.Unauthorized("请先登录"))
		return
	}
	page, pageSize := parsePagination(ctx)
	pagination, list, err := h.svc.ListByUser(userID, page, pageSize)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeDh114RecommendationError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// ListByDh114 按商户列出推荐
// GET /api/v1/dh114/recommendations/by-dh114/:id  （公开）
func (h *RecommendationHandler) ListByDh114(ctx plugin.Context) {
	dh114ID, err := parseSubID(ctx, "id")
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的商户ID"))
		return
	}
	page, pageSize := parsePagination(ctx)
	pagination, list, err := h.svc.ListByDh114(dh114ID, page, pageSize)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeDh114RecommendationError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// GetByID 推荐详情
// GET /api/v1/dh114/recommendations/:id  （公开）
func (h *RecommendationHandler) GetByID(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	info, err := h.svc.GetByID(id)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeDh114RecommendationNotFound, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(info))
}

// MarkClicked 标记已点击
// POST /api/v1/dh114/recommendations/:id/click  （需登录）
func (h *RecommendationHandler) MarkClicked(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	if err := h.svc.MarkClicked(id); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeDh114RecommendationError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("已记录点击", nil))
}

// MarkContacted 标记已联系
// POST /api/v1/dh114/recommendations/:id/contact  （需登录）
func (h *RecommendationHandler) MarkContacted(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	if err := h.svc.MarkContacted(id); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeDh114RecommendationError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("已记录联系", nil))
}

// MarkDismissed 标记已忽略
// POST /api/v1/dh114/recommendations/:id/dismiss  （需登录）
func (h *RecommendationHandler) MarkDismissed(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	if err := h.svc.MarkDismissed(id); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeDh114RecommendationError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("已记录忽略", nil))
}

// AdminList 推荐列表（M 端）
// GET /api/v1/dh114/admin/recommendations  （需 dh114:audit 权限）
func (h *RecommendationHandler) AdminList(ctx plugin.Context) {
	var req dto.RecommendationListRequest
	_ = ctx.Bind(&req)
	if req.Page == 0 {
		req.Page, req.PageSize = parsePagination(ctx)
	}
	pagination, list, err := h.svc.List(&req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeDh114RecommendationError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// AdminCreate 创建推荐（M 端）
// POST /api/v1/dh114/admin/recommendations  （需 dh114:audit 权限）
func (h *RecommendationHandler) AdminCreate(ctx plugin.Context) {
	var req dto.RecommendationListRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	regionID := getRegionID(ctx)
	info, err := h.svc.Create(regionID, &req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeDh114RecommendationError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("推荐创建成功", info))
}

// AdminDelete 删除推荐（M 端）
// DELETE /api/v1/dh114/admin/recommendations/:id  （需 dh114:audit 权限）
func (h *RecommendationHandler) AdminDelete(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	if err := h.svc.Delete(id); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeDh114RecommendationError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("删除成功", nil))
}
