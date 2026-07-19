// Package handler 店铺 HTTP 处理层
// 依据 v3.2.1 架构方案：对标转转商家版
package handler

import (
	"net/http"
	"strconv"

	"wuchang-tongcheng/internal/core/plugin"
	"wuchang-tongcheng/internal/core/response"
	"wuchang-tongcheng/internal/modules/ershou/dto"
	"wuchang-tongcheng/internal/modules/ershou/service"
	"wuchang-tongcheng/internal/pkg/utils"
)

// ShopHandler 店铺 HTTP 处理器
type ShopHandler struct {
	service service.ShopService
}

// NewShopHandler 创建店铺 Handler 实例
func NewShopHandler(svc service.ShopService) *ShopHandler {
	return &ShopHandler{service: svc}
}

// Create 用户开通店铺
// POST /api/v1/ershou/shops  （需登录）
func (h *ShopHandler) Create(ctx plugin.Context) {
	userID, _, _, _ := getUserProfile(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusOK, response.Unauthorized("请先登录"))
		return
	}
	var req dto.ShopCreateRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	resp, err := h.service.Create(userID, &req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeShopError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("店铺创建成功，等待审核", resp))
}

// GetByID 店铺详情
// GET /api/v1/ershou/shops/:id  （公开）
func (h *ShopHandler) GetByID(ctx plugin.Context) {
	shopID, err := parseSubID(ctx, "id")
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的店铺ID"))
		return
	}
	resp, err := h.service.GetByID(shopID)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeShopNotFound, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(resp))
}

// GetMyShop 我的店铺
// GET /api/v1/ershou/shops/mine  （需登录）
func (h *ShopHandler) GetMyShop(ctx plugin.Context) {
	userID, _, _, _ := getUserProfile(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusOK, response.Unauthorized("请先登录"))
		return
	}
	resp, err := h.service.GetByUserID(userID)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeShopNotFound, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(resp))
}

// Update 更新店铺
// PUT /api/v1/ershou/shops/:id  （需登录 + 仅店主）
func (h *ShopHandler) Update(ctx plugin.Context) {
	userID, _, _, _ := getUserProfile(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusOK, response.Unauthorized("请先登录"))
		return
	}
	shopID, err := parseSubID(ctx, "id")
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的店铺ID"))
		return
	}
	var req dto.ShopUpdateRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	resp, err := h.service.Update(shopID, userID, &req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeShopError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("店铺更新成功", resp))
}

// List 店铺列表
// GET /api/v1/ershou/shops  （公开）
func (h *ShopHandler) List(ctx plugin.Context) {
	var query dto.ShopListQuery
	_ = ctx.Bind(&query)
	if query.Page == 0 {
		query.Page = 1
	}
	if query.PageSize == 0 {
		query.PageSize = 10
	}
	pagination, list, err := h.service.List(query)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeShopError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// Audit 店铺审核（M 端）
// PUT /api/v1/ershou/shops/:id/audit  （需登录 + content:audit）
func (h *ShopHandler) Audit(ctx plugin.Context) {
	userID, _, _, _ := getUserProfile(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusOK, response.Unauthorized("请先登录"))
		return
	}
	shopID, err := parseSubID(ctx, "id")
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的店铺ID"))
		return
	}
	approvedStr := ctx.PostForm("approved")
	rejectReason := ctx.PostForm("reject_reason")
	approved := approvedStr == "true" || approvedStr == "1"
	resp, err := h.service.Audit(shopID, userID, approved, rejectReason)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeShopAuditError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("店铺审核完成", resp))
}

// UpdateStatus 更新店铺状态（M 端冻结/关闭/恢复）
// PUT /api/v1/ershou/shops/:id/status  （需登录 + content:audit）
func (h *ShopHandler) UpdateStatus(ctx plugin.Context) {
	userID, _, _, _ := getUserProfile(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusOK, response.Unauthorized("请先登录"))
		return
	}
	shopID, err := parseSubID(ctx, "id")
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的店铺ID"))
		return
	}
	statusStr := ctx.PostForm("status")
	status, err := strconv.Atoi(statusStr)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的状态"))
		return
	}
	resp, err := h.service.UpdateStatus(shopID, userID, status)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeShopError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("店铺状态更新成功", resp))
}

// Follow 关注店铺
// POST /api/v1/ershou/shops/:id/follow  （需登录）
func (h *ShopHandler) Follow(ctx plugin.Context) {
	userID, _, _, _ := getUserProfile(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusOK, response.Unauthorized("请先登录"))
		return
	}
	shopID, err := parseSubID(ctx, "id")
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的店铺ID"))
		return
	}
	var req dto.ShopFollowRequest
	_ = ctx.Bind(&req)
	if err := h.service.Follow(shopID, userID, &req); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeShopError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("关注成功", nil))
}

// Unfollow 取关店铺
// DELETE /api/v1/ershou/shops/:id/follow  （需登录）
func (h *ShopHandler) Unfollow(ctx plugin.Context) {
	userID, _, _, _ := getUserProfile(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusOK, response.Unauthorized("请先登录"))
		return
	}
	shopID, err := parseSubID(ctx, "id")
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的店铺ID"))
		return
	}
	if err := h.service.Unfollow(shopID, userID); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeShopError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("已取关", nil))
}

// ListFollowers 店铺粉丝列表
// GET /api/v1/ershou/shops/:id/followers  （需登录 + 店主或公开）
func (h *ShopHandler) ListFollowers(ctx plugin.Context) {
	shopID, err := parseSubID(ctx, "id")
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的店铺ID"))
		return
	}
	page, pageSize := parsePagination(ctx)
	pagination := utils.NewPagination(page, pageSize)
	pagination, list, err := h.service.ListFollowers(shopID, pagination)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeShopError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// ListUserFollowing 我的关注店铺列表
// GET /api/v1/ershou/shops/following  （需登录）
func (h *ShopHandler) ListUserFollowing(ctx plugin.Context) {
	userID, _, _, _ := getUserProfile(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusOK, response.Unauthorized("请先登录"))
		return
	}
	page, pageSize := parsePagination(ctx)
	pagination := utils.NewPagination(page, pageSize)
	pagination, list, err := h.service.ListUserFollowing(userID, pagination)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeShopError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}
