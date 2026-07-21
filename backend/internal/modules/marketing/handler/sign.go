// Package handler 营销活动中台 HTTP 处理层 - 签到（sign 子域）
package handler

import (
	"net/http"

	"wuchang-tongcheng/internal/core/plugin"
	"wuchang-tongcheng/internal/core/response"
	"wuchang-tongcheng/internal/modules/marketing/dto"
	"wuchang-tongcheng/internal/modules/marketing/service"
	"wuchang-tongcheng/internal/pkg/utils"
)

// SignHandler 签到 HTTP 处理器
type SignHandler struct {
	svc service.SignService
}

// NewSignHandler 创建签到 Handler 实例
func NewSignHandler(svc service.SignService) *SignHandler {
	return &SignHandler{svc: svc}
}

// CheckIn 每日签到（C 端）
// POST /api/v1/marketing/sign/check-in  （需登录）
func (h *SignHandler) CheckIn(ctx plugin.Context) {
	userID, _, _, _ := getUserProfile(ctx)
	if userID == 0 {
		unauthorized(ctx)
		return
	}
	resp, err := h.svc.CheckIn(userID)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(CodeMarketingSignError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("签到成功", resp))
}

// GetCalendar 签到日历（C 端）
// GET /api/v1/marketing/sign/calendar  （需登录）
func (h *SignHandler) GetCalendar(ctx plugin.Context) {
	userID, _, _, _ := getUserProfile(ctx)
	if userID == 0 {
		unauthorized(ctx)
		return
	}
	month := ctx.Query("month")
	resp, err := h.svc.GetCalendar(userID, month)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(CodeMarketingSignError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(resp))
}

// ===== 签到规则管理（M 端） =====

// CreateRule 创建签到规则
// POST /api/v1/marketing/sign-rules  （需 marketing:manage 权限）
func (h *SignHandler) CreateRule(ctx plugin.Context) {
	var req dto.CreateSignRuleRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	info, err := h.svc.CreateRule(&req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(CodeMarketingSignRuleError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("创建成功", info))
}

// UpdateRule 更新签到规则
// PUT /api/v1/marketing/sign-rules/:id  （需 marketing:manage 权限）
func (h *SignHandler) UpdateRule(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	var req dto.UpdateSignRuleRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	if err := h.svc.UpdateRule(id, &req); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(CodeMarketingSignRuleError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("更新成功", nil))
}

// DeleteRule 删除签到规则
// DELETE /api/v1/marketing/sign-rules/:id  （需 marketing:manage 权限）
func (h *SignHandler) DeleteRule(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	if err := h.svc.DeleteRule(id); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(CodeMarketingSignRuleError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("删除成功", nil))
}

// GetRuleByID 签到规则详情
// GET /api/v1/marketing/sign-rules/:id  （需 marketing:manage 权限）
func (h *SignHandler) GetRuleByID(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	info, err := h.svc.GetRuleByID(id)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(CodeMarketingSignRuleNotFound, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(info))
}

// ListRules 签到规则列表
// GET /api/v1/marketing/sign-rules  （需 marketing:manage 权限）
func (h *SignHandler) ListRules(ctx plugin.Context) {
	var req dto.SignRuleListRequest
	_ = ctx.Bind(&req)
	if req.Page == 0 {
		req.Page, req.PageSize = parsePagination(ctx)
	}
	pagination, list, err := h.svc.ListRules(&req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(CodeMarketingSignRuleError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// ListEnabledRules 启用的签到规则列表（C 端）
// GET /api/v1/marketing/sign/rules/enabled  （公开）
func (h *SignHandler) ListEnabledRules(ctx plugin.Context) {
	list, err := h.svc.ListEnabledRules()
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(CodeMarketingSignRuleError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(list))
}
