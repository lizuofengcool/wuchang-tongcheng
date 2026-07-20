// Package handler 同城114 HTTP 处理层 - 商户认证
// 依据 v3.2.1 架构方案：营业执照 + 实地认证 + 品牌授权
package handler

import (
	"net/http"

	"wuchang-tongcheng/internal/core/plugin"
	"wuchang-tongcheng/internal/core/response"
	"wuchang-tongcheng/internal/modules/dh114/dto"
	"wuchang-tongcheng/internal/modules/dh114/service"
	"wuchang-tongcheng/internal/pkg/utils"
)

// VerificationHandler 商户认证 HTTP 处理器
type VerificationHandler struct {
	svc service.VerificationService
}

// NewVerificationHandler 创建商户认证 Handler 实例
func NewVerificationHandler(svc service.VerificationService) *VerificationHandler {
	return &VerificationHandler{svc: svc}
}

// Create 提交认证申请（C 端）
// POST /api/v1/dh114/verifications  （需登录）
func (h *VerificationHandler) Create(ctx plugin.Context) {
	userID, _, _, _ := getUserProfile(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusOK, response.Unauthorized("请先登录"))
		return
	}
	var req dto.CreateVerificationRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	regionID := getRegionID(ctx)
	info, err := h.svc.Create(regionID, userID, &req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeDh114VerificationError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("认证申请已提交", info))
}

// Update 更新认证信息
// PUT /api/v1/dh114/verifications/:id  （需登录）
func (h *VerificationHandler) Update(ctx plugin.Context) {
	userID, _, _, _ := getUserProfile(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusOK, response.Unauthorized("请先登录"))
		return
	}
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	var req dto.UpdateVerificationRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	if err := h.svc.Update(id, userID, &req); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeDh114VerificationError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("更新成功", nil))
}

// Delete 删除认证记录
// DELETE /api/v1/dh114/verifications/:id  （需登录）
func (h *VerificationHandler) Delete(ctx plugin.Context) {
	userID, _, _, _ := getUserProfile(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusOK, response.Unauthorized("请先登录"))
		return
	}
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	if err := h.svc.Delete(id, userID); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeDh114VerificationError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("删除成功", nil))
}

// GetByID 认证详情
// GET /api/v1/dh114/verifications/:id  （公开）
func (h *VerificationHandler) GetByID(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	info, err := h.svc.GetByID(id)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeDh114VerificationNotFound, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(info))
}

// List 认证列表
// GET /api/v1/dh114/verifications  （公开）
func (h *VerificationHandler) List(ctx plugin.Context) {
	var req dto.VerificationListRequest
	_ = ctx.Bind(&req)
	if req.Page == 0 {
		req.Page, req.PageSize = parsePagination(ctx)
	}
	pagination, list, err := h.svc.List(&req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeDh114VerificationError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// ListByDh114 按商户列出认证记录
// GET /api/v1/dh114/dh114/:id/verifications  （公开）
func (h *VerificationHandler) ListByDh114(ctx plugin.Context) {
	dh114ID, err := parseSubID(ctx, "id")
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的商户ID"))
		return
	}
	page, pageSize := parsePagination(ctx)
	pagination, list, err := h.svc.ListByDh114(dh114ID, page, pageSize)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeDh114VerificationError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// ListMine 我的认证记录
// GET /api/v1/dh114/verifications/mine  （需登录）
func (h *VerificationHandler) ListMine(ctx plugin.Context) {
	userID, _, _, _ := getUserProfile(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusOK, response.Unauthorized("请先登录"))
		return
	}
	page, pageSize := parsePagination(ctx)
	pagination, list, err := h.svc.ListByUser(userID, page, pageSize)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeDh114VerificationError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// FindLatest 查询商户最新认证
// GET /api/v1/dh114/dh114/:id/verifications/latest  （公开）
func (h *VerificationHandler) FindLatest(ctx plugin.Context) {
	dh114ID, err := parseSubID(ctx, "id")
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的商户ID"))
		return
	}
	vType := ctx.DefaultQuery("verification_type", "")
	info, err := h.svc.FindLatestByDh114(dh114ID, vType)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeDh114VerificationNotFound, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(info))
}

// AdminList 认证列表（M 端）
// GET /api/v1/dh114/admin/verifications  （需 dh114:audit 权限）
func (h *VerificationHandler) AdminList(ctx plugin.Context) {
	var req dto.VerificationListRequest
	_ = ctx.Bind(&req)
	if req.Page == 0 {
		req.Page, req.PageSize = parsePagination(ctx)
	}
	pagination, list, err := h.svc.AdminList(&req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeDh114VerificationError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// Audit 审核认证（M 端）
// PUT /api/v1/dh114/admin/verifications/:id/audit  （需 dh114:audit 权限）
func (h *VerificationHandler) Audit(ctx plugin.Context) {
	userID, _, _, _ := getUserProfile(ctx)
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	var req dto.VerificationAuditRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	if err := h.svc.Audit(id, userID, &req); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeDh114VerificationError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("审核完成", nil))
}
