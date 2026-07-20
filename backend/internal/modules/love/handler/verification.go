// Package handler love 相亲交友 HTTP 处理层 - 认证
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

// VerificationHandler 认证 HTTP 处理器
type VerificationHandler struct {
	service service.LoveVerificationService
}

// NewVerificationHandler 创建 VerificationHandler 实例
func NewVerificationHandler(svc service.LoveVerificationService) *VerificationHandler {
	return &VerificationHandler{service: svc}
}

// Submit 提交认证申请
// POST /api/v1/love/verifications  （需登录）
func (h *VerificationHandler) Submit(ctx plugin.Context) {
	userID, ok := requireLogin(ctx)
	if !ok {
		return
	}
	var req dto.CreateLoveVerificationRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	loveIDStr := ctx.Query("love_id")
	loveID, _ := parseUint(loveIDStr)
	info, err := h.service.Submit(loveID, userID, &req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeLoveError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("提交成功", info))
}

// GetByID 认证详情
// GET /api/v1/love/verifications/:id  （公开）
func (h *VerificationHandler) GetByID(ctx plugin.Context) {
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

// GetByLoveID 指定用户的认证列表
// GET /api/v1/love/loves/:id/verifications  （公开）
func (h *VerificationHandler) GetByLoveID(ctx plugin.Context) {
	loveID, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	list, err := h.service.GetByLoveID(loveID)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeLoveError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(list))
}

// GetByUserID 当前用户的认证列表
// GET /api/v1/love/verifications/me  （需登录）
func (h *VerificationHandler) GetByUserID(ctx plugin.Context) {
	userID, ok := requireLogin(ctx)
	if !ok {
		return
	}
	list, err := h.service.GetByUserID(userID)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeLoveError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(list))
}

// List 认证列表
// GET /api/v1/love/verifications  （公开）
func (h *VerificationHandler) List(ctx plugin.Context) {
	var req dto.LoveVerificationListRequest
	_ = ctx.Bind(&req)
	pagination, list, err := h.service.List(&req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeLoveError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// ===== M 端 =====

// Audit 管理后台：审核认证
// PUT /api/v1/love/admin/verifications/:id/audit  （需 love:audit 权限）
func (h *VerificationHandler) Audit(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	var req dto.LoveVerificationAuditRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	req.ID = id
	// handledBy / handledName 取自 JWT 上下文
	handledBy, _ := ctx.Get(middleware.ContextUserID)
	handledByName, _ := ctx.Get(middleware.ContextUsername)
	var uid uint
	var name string
	if v, ok := handledBy.(uint); ok {
		uid = v
	}
	if v, ok := handledByName.(string); ok {
		name = v
	}
	if err := h.service.Audit(id, uid, name, &req); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeLoveError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("审核成功", nil))
}

// CountPending 管理后台：待审核数量
// GET /api/v1/love/admin/verifications/pending-count  （需 love:audit 权限）
func (h *VerificationHandler) CountPending(ctx plugin.Context) {
	count, err := h.service.CountPending()
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeLoveError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(map[string]int64{"count": count}))
}
