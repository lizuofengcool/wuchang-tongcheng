// Package handler 同城114 HTTP 处理层 - 商户详情 + 营业时间 + 菜单
package handler

import (
	"net/http"
	"strconv"

	"wuchang-tongcheng/internal/core/plugin"
	"wuchang-tongcheng/internal/core/response"
	"wuchang-tongcheng/internal/modules/dh114/dto"
	"wuchang-tongcheng/internal/modules/dh114/service"
	"wuchang-tongcheng/internal/pkg/utils"
)

// BusinessHandler 商户详情 + 营业时间 + 菜单 HTTP 处理器
type BusinessHandler struct {
	svc service.BusinessService
}

// NewBusinessHandler 创建 Business Handler 实例
func NewBusinessHandler(svc service.BusinessService) *BusinessHandler {
	return &BusinessHandler{svc: svc}
}

// ===== 商户详情 =====

// Create 创建商户详情
// POST /api/v1/dh114/businesses  （需登录）
func (h *BusinessHandler) Create(ctx plugin.Context) {
	userID, _, _, _ := getUserProfile(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusOK, response.Unauthorized("请先登录"))
		return
	}
	var req dto.CreateBusinessRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	regionID := getRegionID(ctx)
	info, err := h.svc.Create(regionID, userID, &req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeDh114BusinessError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("商户详情创建成功", info))
}

// Update 更新商户详情
// PUT /api/v1/dh114/businesses/:id  （需登录）
func (h *BusinessHandler) Update(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	var req dto.UpdateBusinessRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	if err := h.svc.Update(id, &req); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeDh114BusinessError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("更新成功", nil))
}

// Delete 删除商户详情
// DELETE /api/v1/dh114/businesses/:id  （需登录）
func (h *BusinessHandler) Delete(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	if err := h.svc.Delete(id); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeDh114BusinessError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("删除成功", nil))
}

// GetByID 商户详情
// GET /api/v1/dh114/businesses/:id  （公开）
func (h *BusinessHandler) GetByID(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	info, err := h.svc.GetByID(id)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeDh114BusinessNotFound, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(info))
}

// GetByDh114ID 按商户 ID 获取详情
// GET /api/v1/dh114/:id/business  （公开）
func (h *BusinessHandler) GetByDh114ID(ctx plugin.Context) {
	dh114ID, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	info, err := h.svc.GetByDh114ID(dh114ID)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeDh114BusinessNotFound, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(info))
}

// List 商户详情列表
// GET /api/v1/dh114/businesses  （公开）
func (h *BusinessHandler) List(ctx plugin.Context) {
	var req dto.BusinessListRequest
	_ = ctx.Bind(&req)
	if req.Page == 0 {
		req.Page, req.PageSize = parsePagination(ctx)
	}
	pagination, list, err := h.svc.List(&req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeDh114BusinessError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// UpdateVerificationStatus 更新商户认证状态
// PUT /api/v1/dh114/businesses/:id/verification  （需 content:audit 权限）
func (h *BusinessHandler) UpdateVerificationStatus(ctx plugin.Context) {
	dh114ID, err := parseSubID(ctx, "id")
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的商户ID"))
		return
	}
	var req struct {
		Status int `json:"status" binding:"oneof=0 1 2 3"`
	}
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	if err := h.svc.UpdateVerificationStatus(dh114ID, req.Status); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeDh114BusinessError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("认证状态已更新", nil))
}

// ===== 营业时间 =====

// ListBusinessHours 营业时间列表
// GET /api/v1/dh114/:id/business-hours  （公开）
func (h *BusinessHandler) ListBusinessHours(ctx plugin.Context) {
	dh114ID, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	list, err := h.svc.ListBusinessHours(dh114ID)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeDh114BusinessError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(list))
}

// ReplaceBusinessHours 批量替换营业时间
// PUT /api/v1/dh114/:id/business-hours  （需登录）
func (h *BusinessHandler) ReplaceBusinessHours(ctx plugin.Context) {
	dh114ID, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	var req dto.BatchReplaceHoursRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	req.Dh114ID = dh114ID
	if err := h.svc.ReplaceBusinessHours(dh114ID, &req); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeDh114BusinessError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("营业时间已更新", nil))
}

// ===== 菜单 =====

// CreateMenu 创建菜单/服务项目
// POST /api/v1/dh114/menus  （需登录）
func (h *BusinessHandler) CreateMenu(ctx plugin.Context) {
	userID, _, _, _ := getUserProfile(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusOK, response.Unauthorized("请先登录"))
		return
	}
	var req dto.CreateMenuRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	regionID := getRegionID(ctx)
	info, err := h.svc.CreateMenu(regionID, userID, &req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeDh114MenuError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("菜单创建成功", info))
}

// UpdateMenu 更新菜单
// PUT /api/v1/dh114/menus/:id  （需登录）
func (h *BusinessHandler) UpdateMenu(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	var req dto.UpdateMenuRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	if err := h.svc.UpdateMenu(id, &req); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeDh114MenuError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("更新成功", nil))
}

// DeleteMenu 删除菜单
// DELETE /api/v1/dh114/menus/:id  （需登录）
func (h *BusinessHandler) DeleteMenu(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	if err := h.svc.DeleteMenu(id); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeDh114MenuError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("删除成功", nil))
}

// GetMenuByID 菜单详情
// GET /api/v1/dh114/menus/:id  （公开）
func (h *BusinessHandler) GetMenuByID(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	info, err := h.svc.GetMenuByID(id)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeDh114MenuNotFound, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(info))
}

// ListMenus 菜单列表
// GET /api/v1/dh114/menus  （公开）
func (h *BusinessHandler) ListMenus(ctx plugin.Context) {
	var req dto.MenuListRequest
	_ = ctx.Bind(&req)
	if req.Page == 0 {
		req.Page, req.PageSize = parsePagination(ctx)
	}
	pagination, list, err := h.svc.ListMenus(&req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeDh114MenuError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// ListMenusByDh114 按商户列出菜单
// GET /api/v1/dh114/:id/menus  （公开）
func (h *BusinessHandler) ListMenusByDh114(ctx plugin.Context) {
	dh114ID, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	onlyActive := ctx.Query("only_active") == "true" || ctx.Query("only_active") == "1"
	list, err := h.svc.ListMenusByDh114(dh114ID, onlyActive)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeDh114MenuError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(list))
}

// ListSignatureMenus 招牌菜单
// GET /api/v1/dh114/:id/menus/signature  （公开）
func (h *BusinessHandler) ListSignatureMenus(ctx plugin.Context) {
	dh114ID, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	list, err := h.svc.ListSignatureMenus(dh114ID)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeDh114MenuError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(list))
}

// ReplaceMenus 批量替换菜单
// PUT /api/v1/dh114/:id/menus  （需登录）
func (h *BusinessHandler) ReplaceMenus(ctx plugin.Context) {
	userID, _, _, _ := getUserProfile(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusOK, response.Unauthorized("请先登录"))
		return
	}
	dh114ID, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	var req dto.BatchReplaceMenusRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	req.Dh114ID = dh114ID
	regionID := getRegionID(ctx)
	if err := h.svc.ReplaceMenus(regionID, userID, &req); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeDh114MenuError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("菜单已批量更新", nil))
}

// IncrMenuOrderCount 增加菜单点单数（下单时调用）
// POST /api/v1/dh114/menus/:id/order-count  （公开）
func (h *BusinessHandler) IncrMenuOrderCount(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	countStr := ctx.DefaultQuery("count", "1")
	count, _ := strconv.Atoi(countStr)
	if err := h.svc.IncrMenuOrderCount(id, count); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeDh114MenuError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("点单数已更新", nil))
}
