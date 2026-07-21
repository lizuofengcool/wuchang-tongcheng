// Package handler 商户中台 HTTP 处理层 - 店铺
// 依据架构设计 4.4：商家入驻/认领/店铺管理/信用分/等级
package handler

import (
	"net/http"

	"wuchang-tongcheng/internal/core/plugin"
	"wuchang-tongcheng/internal/core/response"
	"wuchang-tongcheng/internal/modules/merchant/dto"
	"wuchang-tongcheng/internal/modules/merchant/service"
	"wuchang-tongcheng/internal/pkg/utils"
)

// ShopHandler 店铺 HTTP 处理器
type ShopHandler struct {
	svc service.ShopService
}

// NewShopHandler 创建店铺 Handler 实例
func NewShopHandler(svc service.ShopService) *ShopHandler {
	return &ShopHandler{svc: svc}
}

// List 店铺列表（公开）
// GET /api/v1/merchant/shops
func (h *ShopHandler) List(ctx plugin.Context) {
	var req dto.ShopListRequest
	_ = ctx.Bind(&req)
	if req.Page == 0 {
		req.Page, req.PageSize = parsePagination(ctx)
	}
	regionID := getRegionID(ctx)
	pagination, list, err := h.svc.List(regionID, &req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(CodeMerchantError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// GetByID 店铺详情（公开）
// GET /api/v1/merchant/shops/:id
func (h *ShopHandler) GetByID(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	info, err := h.svc.GetByID(id)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(CodeMerchantShopNotFound, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(info))
}

// Apply 商户入驻（需登录）
// POST /api/v1/merchant/shops/apply
func (h *ShopHandler) Apply(ctx plugin.Context) {
	userID, _, _, _ := getUserProfile(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusOK, failUnauthorized())
		return
	}
	var req dto.CreateShopRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, failParam("参数错误: "+err.Error()))
		return
	}
	regionID := getRegionID(ctx)
	info, err := h.svc.Apply(regionID, userID, &req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(CodeMerchantApplyError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("入驻申请已提交", info))
}

// Update 更新店铺（需登录）
// PUT /api/v1/merchant/shops/:id
func (h *ShopHandler) Update(ctx plugin.Context) {
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
	var req dto.UpdateShopRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, failParam("参数错误: "+err.Error()))
		return
	}
	if err := h.svc.Update(id, userID, &req); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(CodeMerchantShopNoPermission, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("更新成功", nil))
}

// ListMine 我的店铺（需登录）
// GET /api/v1/merchant/shops/mine
func (h *ShopHandler) ListMine(ctx plugin.Context) {
	userID, _, _, _ := getUserProfile(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusOK, failUnauthorized())
		return
	}
	page, pageSize := parsePagination(ctx)
	pagination, list, err := h.svc.ListMine(userID, page, pageSize)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(CodeMerchantError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// Claim 认领店铺（需登录）
// POST /api/v1/merchant/shops/claim
func (h *ShopHandler) Claim(ctx plugin.Context) {
	userID, _, _, _ := getUserProfile(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusOK, failUnauthorized())
		return
	}
	var req dto.ShopClaimRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, failParam("参数错误: "+err.Error()))
		return
	}
	info, err := h.svc.Claim(req.ShopID, userID)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(CodeMerchantClaimError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("认领成功", info))
}

// Search 搜索店铺（公开）
// GET /api/v1/merchant/shops/search
func (h *ShopHandler) Search(ctx plugin.Context) {
	var req dto.ShopListRequest
	_ = ctx.Bind(&req)
	if req.Page == 0 {
		req.Page, req.PageSize = parsePagination(ctx)
	}
	regionID := getRegionID(ctx)
	pagination, list, err := h.svc.Search(regionID, &req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(CodeMerchantError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// ===== M 端管理（需 merchant:audit 权限） =====

// AdminList 管理后台店铺列表
// GET /api/v1/merchant/admin/shops
func (h *ShopHandler) AdminList(ctx plugin.Context) {
	var req dto.ShopAdminListRequest
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

// AdminGetByID 管理后台店铺详情
// GET /api/v1/merchant/admin/shops/:id
func (h *ShopHandler) AdminGetByID(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	info, err := h.svc.AdminGetByID(id)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(CodeMerchantShopNotFound, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(info))
}

// UpdateStatus 更新店铺状态（M 端）
// PUT /api/v1/merchant/admin/shops/:id/status
func (h *ShopHandler) UpdateStatus(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	var req dto.ShopStatusUpdateRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, failParam("参数错误: "+err.Error()))
		return
	}
	if err := h.svc.UpdateStatus(id, req.Status); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(CodeMerchantShopStatusInvalid, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("状态更新成功", nil))
}

// UpdateCreditScore 信用分调整（M 端）
// PUT /api/v1/merchant/admin/shops/:id/credit
func (h *ShopHandler) UpdateCreditScore(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	var req dto.ShopCreditAdjustRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, failParam("参数错误: "+err.Error()))
		return
	}
	if err := h.svc.UpdateCreditScore(id, req.Delta, req.Reason); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(CodeMerchantShopNotFound, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("信用分调整成功", nil))
}

// UpdateLevel 等级调整（M 端）
// PUT /api/v1/merchant/admin/shops/:id/level
func (h *ShopHandler) UpdateLevel(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	var req dto.ShopLevelUpdateRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, failParam("参数错误: "+err.Error()))
		return
	}
	if err := h.svc.UpdateLevel(id, req.Level); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(CodeMerchantShopNotFound, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("等级调整成功", nil))
}
