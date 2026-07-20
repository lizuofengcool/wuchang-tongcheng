// Package handler 同城114 HTTP 处理层 - 收藏
// 依据 v3.2.1 架构方案：商户/团购/优惠券收藏
package handler

import (
	"net/http"
	"strings"

	"wuchang-tongcheng/internal/core/plugin"
	"wuchang-tongcheng/internal/core/response"
	"wuchang-tongcheng/internal/modules/dh114/dto"
	"wuchang-tongcheng/internal/modules/dh114/service"
	"wuchang-tongcheng/internal/pkg/utils"
)

// FavoriteHandler 收藏 HTTP 处理器
type FavoriteHandler struct {
	svc service.FavoriteService
}

// NewFavoriteHandler 创建收藏 Handler 实例
func NewFavoriteHandler(svc service.FavoriteService) *FavoriteHandler {
	return &FavoriteHandler{svc: svc}
}

// Create 创建收藏（C 端）
// POST /api/v1/dh114/favorites  （需登录）
func (h *FavoriteHandler) Create(ctx plugin.Context) {
	userID, _, _, _ := getUserProfile(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusOK, response.Unauthorized("请先登录"))
		return
	}
	var req dto.CreateFavoriteRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	info, err := h.svc.Create(userID, &req)
	if err != nil {
		// 已收藏视为成功（幂等）
		if strings.Contains(err.Error(), "已收藏") {
			ctx.JSON(http.StatusOK, response.SuccessWithMessage("已收藏", nil))
			return
		}
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeDh114FavoriteError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("收藏成功", info))
}

// Delete 删除收藏
// DELETE /api/v1/dh114/favorites/:id  （需登录）
func (h *FavoriteHandler) Delete(ctx plugin.Context) {
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
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeDh114FavoriteError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("取消收藏成功", nil))
}

// DeleteByTarget 按目标取消收藏
// DELETE /api/v1/dh114/favorites/by-target  （需登录）
func (h *FavoriteHandler) DeleteByTarget(ctx plugin.Context) {
	userID, _, _, _ := getUserProfile(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusOK, response.Unauthorized("请先登录"))
		return
	}
	var req dto.CreateFavoriteRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	favType := req.FavoriteType
	if favType == "" {
		favType = "business"
	}
	if err := h.svc.DeleteByTarget(userID, req.Dh114ID, favType); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeDh114FavoriteError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("取消收藏成功", nil))
}

// Update 更新收藏（备注/分组）
// PUT /api/v1/dh114/favorites/:id  （需登录）
func (h *FavoriteHandler) Update(ctx plugin.Context) {
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
	var req dto.UpdateFavoriteRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	if err := h.svc.Update(id, userID, &req); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeDh114FavoriteError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("更新成功", nil))
}

// GetByID 收藏详情
// GET /api/v1/dh114/favorites/:id  （需登录）
func (h *FavoriteHandler) GetByID(ctx plugin.Context) {
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
	info, err := h.svc.GetByID(id)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeDh114FavoriteError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(info))
}

// List 收藏列表
// GET /api/v1/dh114/favorites  （需登录）
func (h *FavoriteHandler) List(ctx plugin.Context) {
	userID, _, _, _ := getUserProfile(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusOK, response.Unauthorized("请先登录"))
		return
	}
	var req dto.FavoriteListRequest
	_ = ctx.Bind(&req)
	if req.Page == 0 {
		req.Page, req.PageSize = parsePagination(ctx)
	}
	pagination, list, err := h.svc.List(userID, &req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeDh114FavoriteError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// ListByType 按类型列出收藏
// GET /api/v1/dh114/favorites/by-type  （需登录）
func (h *FavoriteHandler) ListByType(ctx plugin.Context) {
	userID, _, _, _ := getUserProfile(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusOK, response.Unauthorized("请先登录"))
		return
	}
	favType := ctx.DefaultQuery("favorite_type", "business")
	page, pageSize := parsePagination(ctx)
	pagination, list, err := h.svc.ListByType(userID, favType, page, pageSize)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeDh114FavoriteError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// ListByGroup 按分组列出收藏
// GET /api/v1/dh114/favorites/by-group  （需登录）
func (h *FavoriteHandler) ListByGroup(ctx plugin.Context) {
	userID, _, _, _ := getUserProfile(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusOK, response.Unauthorized("请先登录"))
		return
	}
	groupIDStr := ctx.Query("group_id")
	groupID, err := parseUintStr(groupIDStr)
	if err != nil || groupID == 0 {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的分组ID"))
		return
	}
	page, pageSize := parsePagination(ctx)
	pagination, list, err := h.svc.ListByGroup(userID, groupID, page, pageSize)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeDh114FavoriteError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// HasFaved 查询是否已收藏
// GET /api/v1/dh114/favorites/has  （需登录）
func (h *FavoriteHandler) HasFaved(ctx plugin.Context) {
	userID, _, _, _ := getUserProfile(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusOK, response.Unauthorized("请先登录"))
		return
	}
	dh114ID, err := parseUintStr(ctx.Query("id"))
	if err != nil || dh114ID == 0 {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的商户ID"))
		return
	}
	favType := ctx.DefaultQuery("favorite_type", "business")
	has, err := h.svc.HasFaved(userID, dh114ID, favType)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeDh114FavoriteError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(map[string]bool{"has_faved": has}))
}

// parseUintStr 解析字符串为 uint
func parseUintStr(s string) (uint, error) {
	var n uint
	for _, c := range s {
		if c < '0' || c > '9' {
			return 0, errInvalidID
		}
		n = n*10 + uint(c-'0')
	}
	return n, nil
}

var errInvalidID = &invalidIDError{}

type invalidIDError struct{}

func (e *invalidIDError) Error() string { return "invalid id" }
