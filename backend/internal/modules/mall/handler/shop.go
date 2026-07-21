// Package handler 同城商城 HTTP 处理层 - 店铺
package handler

import (
	"net/http"

	"wuchang-tongcheng/internal/core/plugin"
	"wuchang-tongcheng/internal/core/response"
	"wuchang-tongcheng/internal/modules/mall/dto"
	"wuchang-tongcheng/internal/modules/mall/service"
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

// Create 创建店铺
// POST /api/v1/mall/shops  （需登录）
func (h *ShopHandler) Create(ctx plugin.Context) {
	userID, ok := requireLogin(ctx)
	if ok {
		return
	}
	var req dto.CreateShopRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	regionID := getRegionID(ctx)
	_, userName, _, _ := getUserProfile(ctx)
	info, err := h.svc.Create(regionID, userID, userName, &req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeMallShopError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("创建成功", info))
}

// Update 更新店铺
// PUT /api/v1/mall/shops/:id  （需登录）
func (h *ShopHandler) Update(ctx plugin.Context) {
	userID, ok := requireLogin(ctx)
	if ok {
		return
	}
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	var req dto.UpdateShopRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	if err := h.svc.Update(id, userID, &req); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeMallShopError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("更新成功", nil))
}

// Delete 删除店铺
// DELETE /api/v1/mall/shops/:id  （需登录）
func (h *ShopHandler) Delete(ctx plugin.Context) {
	userID, ok := requireLogin(ctx)
	if ok {
		return
	}
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	if err := h.svc.Delete(id, userID); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeMallShopError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("删除成功", nil))
}

// GetByID 店铺详情
// GET /api/v1/mall/shops/:id  （公开）
func (h *ShopHandler) GetByID(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	info, err := h.svc.GetByID(id)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeMallShopNotFound, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(info))
}

// GetByUserID 根据用户 ID 获取店铺
// GET /api/v1/mall/shops/mine  （需登录）
func (h *ShopHandler) GetByUserID(ctx plugin.Context) {
	userID, ok := requireLogin(ctx)
	if ok {
		return
	}
	info, err := h.svc.GetByUserID(userID)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeMallShopNotFound, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(info))
}

// List 店铺列表
// GET /api/v1/mall/shops  （公开）
func (h *ShopHandler) List(ctx plugin.Context) {
	var req dto.ShopListRequest
	_ = ctx.Bind(&req)
	if req.Page == 0 {
		req.Page, req.PageSize = parsePagination(ctx)
	}
	regionID := getRegionID(ctx)
	pagination, list, err := h.svc.List(regionID, &req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeMallShopError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// AdminList 管理后台店铺列表
// GET /api/v1/mall/admin/shops  （需 mall:audit 权限）
func (h *ShopHandler) AdminList(ctx plugin.Context) {
	var req dto.ShopAdminListRequest
	_ = ctx.Bind(&req)
	if req.Page == 0 {
		req.Page, req.PageSize = parsePagination(ctx)
	}
	if req.RegionID == 0 {
		req.RegionID = getRegionID(ctx)
	}
	pagination, list, err := h.svc.AdminList(&req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeMallShopError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// Search 搜索店铺
// GET /api/v1/mall/shops/search  （公开）
func (h *ShopHandler) Search(ctx plugin.Context) {
	keyword := ctx.Query("keyword")
	page, pageSize := parsePagination(ctx)
	regionID := getRegionID(ctx)
	pagination, list, err := h.svc.Search(regionID, keyword, page, pageSize)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeMallShopError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// ListByUser 按用户列出店铺
// GET /api/v1/mall/shops/by-user/:user_id  （公开）
func (h *ShopHandler) ListByUser(ctx plugin.Context) {
	userID, err := parseSubID(ctx, "user_id")
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的用户ID"))
		return
	}
	page, pageSize := parsePagination(ctx)
	pagination, list, err := h.svc.ListByUser(userID, page, pageSize)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeMallShopError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// ListByCategory 按分类列出店铺
// GET /api/v1/mall/shops/by-category/:category_id  （公开）
func (h *ShopHandler) ListByCategory(ctx plugin.Context) {
	categoryID, err := parseSubID(ctx, "category_id")
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的分类ID"))
		return
	}
	page, pageSize := parsePagination(ctx)
	regionID := getRegionID(ctx)
	pagination, list, err := h.svc.ListByCategory(regionID, categoryID, page, pageSize)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeMallShopError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// UpdateStatus 更新店铺状态
// PUT /api/v1/mall/shops/:id/status  （需登录）
func (h *ShopHandler) UpdateStatus(ctx plugin.Context) {
	userID, ok := requireLogin(ctx)
	if ok {
		return
	}
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
	_ = userID
	if err := h.svc.UpdateStatus(id, req.Status); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeMallShopError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("状态已更新", nil))
}

// Audit 审核店铺
// PUT /api/v1/mall/admin/shops/:id/audit  （需 mall:audit 权限）
func (h *ShopHandler) Audit(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	var req dto.AuditRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	if err := h.svc.Audit(id, req.AuditStatus, req.AuditReason); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeMallShopError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("审核完成", nil))
}

// UpdatePromotion 更新店铺推广配置
// PUT /api/v1/mall/admin/shops/:id/promotion  （需 mall:audit 权限）
func (h *ShopHandler) UpdatePromotion(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	var req dto.ShopPromotionRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	if err := h.svc.UpdatePromotion(id, &req); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeMallShopError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("推广配置已更新", nil))
}

// IncrViewCount 增加店铺浏览数
// POST /api/v1/mall/shops/:id/view  （公开）
func (h *ShopHandler) IncrViewCount(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	if err := h.svc.IncrViewCount(id); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeMallShopError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("已记录浏览", nil))
}
