// Package handler 小区信息 HTTP 处理层
// 依据 v3.2.1 架构方案第五章：对标贝壳/链家
package handler

import (
	"net/http"

	"wuchang-tongcheng/internal/core/plugin"
	"wuchang-tongcheng/internal/core/response"
	"wuchang-tongcheng/internal/modules/house/dto"
	"wuchang-tongcheng/internal/modules/house/service"
	"wuchang-tongcheng/internal/pkg/utils"
)

// CommunityHandler 小区信息 HTTP 处理器
type CommunityHandler struct {
	service service.CommunityService
}

// NewCommunityHandler 创建 CommunityHandler 实例
func NewCommunityHandler(svc service.CommunityService) *CommunityHandler {
	return &CommunityHandler{service: svc}
}

// ===== C 端 =====

// Create 创建小区（C 端用户可补充）
// POST /api/v1/house/communities  （需登录）
func (h *CommunityHandler) Create(ctx plugin.Context) {
	userID, _, _, _ := getUserProfile(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusOK, response.Unauthorized("请先登录"))
		return
	}
	var req dto.CommunityCreateRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	regionID := getRegionID(ctx)
	info, err := h.service.Create(regionID, userID, &req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeHouseCommunityError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("创建成功", info))
}

// Update 更新小区（M 端）
// PUT /api/v1/admin/house/communities/:id  （需 house:audit 权限）
func (h *CommunityHandler) Update(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	var req dto.CommunityCreateRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	userID, _, _, _ := getUserProfile(ctx)
	if err := h.service.Update(id, userID, &req); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeHouseCommunityError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("更新成功", nil))
}

// GetByID 获取小区详情
// GET /api/v1/house/communities/:id  （公开）
func (h *CommunityHandler) GetByID(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	userID, _, _, _ := getUserProfile(ctx)
	info, err := h.service.GetByID(id, userID)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeHouseCommunityNotFound, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(info))
}

// List 小区列表
// GET /api/v1/house/communities  （公开）
func (h *CommunityHandler) List(ctx plugin.Context) {
	var req dto.CommunityListQuery
	_ = ctx.Bind(&req)
	regionID := getRegionID(ctx)
	pagination, list, err := h.service.List(regionID, &req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeHouseCommunityError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// ListNearby 附近小区
// GET /api/v1/house/communities/nearby  （公开）
func (h *CommunityHandler) ListNearby(ctx plugin.Context) {
	var req dto.CommunityListQuery
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	if req.Latitude < -90 || req.Latitude > 90 || req.Longitude < -180 || req.Longitude > 180 {
		ctx.JSON(http.StatusOK, response.BadRequest("经纬度参数无效"))
		return
	}
	regionID := getRegionID(ctx)
	pagination, list, err := h.service.ListNearby(regionID, &req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeHouseCommunityError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// ===== 关注 =====

// Follow 关注/取消关注小区（toggle 语义）
// POST /api/v1/house/communities/:id/follow  （需登录）
func (h *CommunityHandler) Follow(ctx plugin.Context) {
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
	var req dto.CommunityFollowRequest
	_ = ctx.Bind(&req)
	res, err := h.service.Follow(userID, id, req.Notify)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeHouseCommunityError, err.Error()))
		return
	}
	if res.HasFaved {
		ctx.JSON(http.StatusOK, response.SuccessWithMessage("关注成功", res))
	} else {
		ctx.JSON(http.StatusOK, response.SuccessWithMessage("已取消关注", res))
	}
}

// FollowStatus 查询关注状态
// GET /api/v1/house/communities/:id/follow  （公开，未登录返回 has_faved=false）
func (h *CommunityHandler) FollowStatus(ctx plugin.Context) {
	userID, _, _, _ := getUserProfile(ctx)
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	res, err := h.service.FollowStatus(userID, id)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeHouseCommunityNotFound, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(res))
}

// ===== M 端 =====

// AdminList 管理后台小区列表
// GET /api/v1/admin/house/communities  （需 house:audit 权限）
func (h *CommunityHandler) AdminList(ctx plugin.Context) {
	var req dto.CommunityAdminListQuery
	_ = ctx.Bind(&req)
	pagination, list, err := h.service.AdminList(&req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeHouseCommunityError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// UpdateStatus 强制下架/恢复小区
// PUT /api/v1/admin/house/communities/:id/status  （需 house:audit 权限）
func (h *CommunityHandler) UpdateStatus(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	var req dto.UpdateHouseStatusRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	if err := h.service.UpdateStatus(id, req.Status); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeHouseCommunityError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("状态更新成功", nil))
}
