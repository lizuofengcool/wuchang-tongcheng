// Package handler 同城拼车出行 HTTP 处理层 - 路线 + 常用路线收藏
// 依据 v3.2.1 架构方案：对标哈啰出行/嘀嗒出行/滴滴顺风车
package handler

import (
	"net/http"

	"wuchang-tongcheng/internal/core/plugin"
	"wuchang-tongcheng/internal/core/response"
	"wuchang-tongcheng/internal/modules/pinche/dto"
	"wuchang-tongcheng/internal/modules/pinche/service"
	"wuchang-tongcheng/internal/pkg/utils"
)

// RouteHandler 路线 HTTP 处理器
type RouteHandler struct {
	service service.RouteService
}

// NewRouteHandler 创建 RouteHandler 实例
func NewRouteHandler(svc service.RouteService) *RouteHandler {
	return &RouteHandler{service: svc}
}

// ===== C 端 =====

// Create 创建路线
// POST /api/v1/pinche/routes  （需登录）
func (h *RouteHandler) Create(ctx plugin.Context) {
	userID, _, _, _ := getUserProfile(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusOK, response.Unauthorized("请先登录"))
		return
	}

	var req dto.CreateRouteRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}

	regionID := getRegionID(ctx)
	info, err := h.service.Create(regionID, userID, &req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodePincheError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("路线创建成功", info))
}

// Update 更新路线（仅本人）
// PUT /api/v1/pinche/routes/:id  （需登录）
func (h *RouteHandler) Update(ctx plugin.Context) {
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

	var req dto.UpdateRouteRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}

	if err := h.service.Update(id, userID, &req); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodePincheRouteNotFound, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("更新成功", nil))
}

// Delete 删除路线（仅本人）
// DELETE /api/v1/pinche/routes/:id  （需登录）
func (h *RouteHandler) Delete(ctx plugin.Context) {
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

	if err := h.service.Delete(id, userID); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodePincheRouteNotFound, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("删除成功", nil))
}

// GetByID 路线详情
// GET /api/v1/pinche/routes/:id  （公开）
func (h *RouteHandler) GetByID(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}

	info, err := h.service.GetByID(id)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodePincheRouteNotFound, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(info))
}

// List 路线列表
// GET /api/v1/pinche/routes  （公开）
func (h *RouteHandler) List(ctx plugin.Context) {
	var req dto.RouteListRequest
	_ = ctx.Bind(&req)

	regionID := getRegionID(ctx)
	pagination, list, err := h.service.List(regionID, &req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodePincheError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// ListMine 我的路线
// GET /api/v1/pinche/routes/mine  （需登录）
func (h *RouteHandler) ListMine(ctx plugin.Context) {
	userID, _, _, _ := getUserProfile(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusOK, response.Unauthorized("请先登录"))
		return
	}
	page, pageSize := parsePagination(ctx)
	pagination, list, err := h.service.ListByUser(userID, page, pageSize)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodePincheError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// ListCommon 常用路线
// GET /api/v1/pinche/routes/common  （公开）
func (h *RouteHandler) ListCommon(ctx plugin.Context) {
	regionID := getRegionID(ctx)
	page, pageSize := parsePagination(ctx)
	pagination, list, err := h.service.ListCommon(regionID, page, pageSize)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodePincheError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// Fav 收藏/取消收藏路线
// POST /api/v1/pinche/routes/:id/fav  （需登录）
func (h *RouteHandler) Fav(ctx plugin.Context) {
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
	res, err := h.service.FavRoute(userID, id)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodePincheRouteFavoriteError, err.Error()))
		return
	}
	if res.HasFaved {
		ctx.JSON(http.StatusOK, response.SuccessWithMessage("收藏成功", res))
	} else {
		ctx.JSON(http.StatusOK, response.SuccessWithMessage("已取消收藏", res))
	}
}

// ListFavs 我的收藏路线
// GET /api/v1/pinche/routes/favorites  （需登录）
func (h *RouteHandler) ListFavs(ctx plugin.Context) {
	userID, _, _, _ := getUserProfile(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusOK, response.Unauthorized("请先登录"))
		return
	}
	page, pageSize := parsePagination(ctx)
	pagination, list, err := h.service.ListFavRoutes(userID, page, pageSize)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodePincheError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// IncrUseCount 增加路线使用次数
// POST /api/v1/pinche/routes/:id/use  （需登录）
func (h *RouteHandler) IncrUseCount(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	if err := h.service.IncrUseCount(id); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodePincheError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("已记录使用", nil))
}
