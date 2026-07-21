// Package handler 商户中台 HTTP 处理层 - 认证
package handler

import (
	"net/http"

	"wuchang-tongcheng/internal/core/plugin"
	"wuchang-tongcheng/internal/core/response"
	"wuchang-tongcheng/internal/modules/merchant/dto"
	"wuchang-tongcheng/internal/modules/merchant/service"
	"wuchang-tongcheng/internal/pkg/utils"
)

// VerificationHandler 认证 HTTP 处理器
type VerificationHandler struct {
	svc service.VerificationService
}

// NewVerificationHandler 创建认证 Handler 实例
func NewVerificationHandler(svc service.VerificationService) *VerificationHandler {
	return &VerificationHandler{svc: svc}
}

// Create 提交认证（需登录）
// POST /api/v1/merchant/verifications
func (h *VerificationHandler) Create(ctx plugin.Context) {
	userID, _, _, _ := getUserProfile(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusOK, failUnauthorized())
		return
	}
	var req dto.CreateVerificationRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, failParam("参数错误: "+err.Error()))
		return
	}
	regionID := getRegionID(ctx)
	info, err := h.svc.Create(regionID, userID, &req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(CodeMerchantError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("认证已提交", info))
}

// Update 更新认证（需登录）
// PUT /api/v1/merchant/verifications/:id
func (h *VerificationHandler) Update(ctx plugin.Context) {
	userID, _, _, _ := getUserProfile(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusOK, failUnauthorized())
		return
	}
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	var req dto.UpdateVerificationRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, failParam("参数错误: "+err.Error()))
		return
	}
	if err := h.svc.Update(id, userID, &req); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(CodeMerchantVerifyNotFound, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("更新成功", nil))
}

// Delete 删除认证（需登录）
// DELETE /api/v1/merchant/verifications/:id
func (h *VerificationHandler) Delete(ctx plugin.Context) {
	userID, _, _, _ := getUserProfile(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusOK, failUnauthorized())
		return
	}
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	if err := h.svc.Delete(id, userID); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(CodeMerchantVerifyNotFound, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("删除成功", nil))
}

// GetByID 认证详情（公开）
// GET /api/v1/merchant/verifications/:id
func (h *VerificationHandler) GetByID(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	info, err := h.svc.GetByID(id)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(CodeMerchantVerifyNotFound, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(info))
}

// List 认证列表（公开）
// GET /api/v1/merchant/verifications
func (h *VerificationHandler) List(ctx plugin.Context) {
	var req dto.VerificationListRequest
	_ = ctx.Bind(&req)
	if req.Page == 0 {
		req.Page, req.PageSize = parsePagination(ctx)
	}
	pagination, list, err := h.svc.List(&req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(CodeMerchantError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// ListByShop 按商户列出认证（公开）
// GET /api/v1/merchant/shops/:id/verifications
func (h *VerificationHandler) ListByShop(ctx plugin.Context) {
	shopID, err := parseSubID(ctx, "id")
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的商户ID"))
		return
	}
	page, pageSize := parsePagination(ctx)
	pagination, list, err := h.svc.ListByShop(shopID, page, pageSize)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(CodeMerchantError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// ===== M 端管理 =====

// AdminList 管理后台认证列表
// GET /api/v1/merchant/admin/verifications
func (h *VerificationHandler) AdminList(ctx plugin.Context) {
	var req dto.VerificationListRequest
	_ = ctx.Bind(&req)
	if req.Page == 0 {
		req.Page, req.PageSize = parsePagination(ctx)
	}
	pagination, list, err := h.svc.AdminList(&req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(CodeMerchantError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// Audit 审核认证（M 端）
// PUT /api/v1/merchant/admin/verifications/:id/audit
func (h *VerificationHandler) Audit(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	var req dto.VerificationAuditRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, failParam("参数错误: "+err.Error()))
		return
	}
	if err := h.svc.Audit(id, &req); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(CodeMerchantVerifyAudited, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("审核完成", nil))
}
