// Package handler 经纪人 HTTP 处理层
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

// AgentHandler 经纪人 HTTP 处理器
type AgentHandler struct {
	service service.AgentService
}

// NewAgentHandler 创建 AgentHandler 实例
func NewAgentHandler(svc service.AgentService) *AgentHandler {
	return &AgentHandler{service: svc}
}

// ===== C 端 =====

// Create 申请成为经纪人
// POST /api/v1/house/agents  （需登录）
func (h *AgentHandler) Create(ctx plugin.Context) {
	userID, _, _, _ := getUserProfile(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusOK, response.Unauthorized("请先登录"))
		return
	}
	var req dto.AgentCreateRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	regionID := getRegionID(ctx)
	info, err := h.service.Create(regionID, userID, &req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeHouseAgentError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("申请已提交", info))
}

// Update 更新经纪人信息（仅本人）
// PUT /api/v1/house/agents/:id  （需登录）
func (h *AgentHandler) Update(ctx plugin.Context) {
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
	var req dto.AgentCreateRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	if err := h.service.Update(id, userID, &req); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeHouseAgentError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("更新成功", nil))
}

// GetByID 获取经纪人详情
// GET /api/v1/house/agents/:id  （公开）
func (h *AgentHandler) GetByID(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	userID, _, _, _ := getUserProfile(ctx)
	info, err := h.service.GetByID(id, userID)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeHouseAgentNotFound, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(info))
}

// List 经纪人列表
// GET /api/v1/house/agents  （公开）
func (h *AgentHandler) List(ctx plugin.Context) {
	var req dto.AgentListQuery
	_ = ctx.Bind(&req)
	regionID := getRegionID(ctx)
	pagination, list, err := h.service.List(regionID, &req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeHouseAgentError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// GetMine 查询当前用户的经纪人信息
// GET /api/v1/house/agents/mine  （需登录）
func (h *AgentHandler) GetMine(ctx plugin.Context) {
	userID, _, _, _ := getUserProfile(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusOK, response.Unauthorized("请先登录"))
		return
	}
	info, err := h.service.GetMine(userID)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeHouseAgentNotFound, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(info))
}

// ===== 关注 =====

// Follow 关注/取消关注经纪人（toggle 语义）
// POST /api/v1/house/agents/:id/follow  （需登录）
func (h *AgentHandler) Follow(ctx plugin.Context) {
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
	var req dto.AgentFollowRequest
	_ = ctx.Bind(&req)
	res, err := h.service.Follow(userID, id, req.Notify)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeHouseAgentError, err.Error()))
		return
	}
	if res.HasFaved {
		ctx.JSON(http.StatusOK, response.SuccessWithMessage("关注成功", res))
	} else {
		ctx.JSON(http.StatusOK, response.SuccessWithMessage("已取消关注", res))
	}
}

// FollowStatus 查询关注状态
// GET /api/v1/house/agents/:id/follow  （公开，未登录返回 has_faved=false）
func (h *AgentHandler) FollowStatus(ctx plugin.Context) {
	userID, _, _, _ := getUserProfile(ctx)
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	res, err := h.service.FollowStatus(userID, id)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeHouseAgentNotFound, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(res))
}

// ===== M 端管理 =====

// AdminList 管理后台经纪人列表
// GET /api/v1/admin/house/agents  （需 house:audit 权限）
func (h *AgentHandler) AdminList(ctx plugin.Context) {
	var req dto.AgentAdminListQuery
	_ = ctx.Bind(&req)
	pagination, list, err := h.service.AdminList(&req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeHouseAgentError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// Audit 审核经纪人（通过/拒绝/冻结/撤销）
// PUT /api/v1/admin/house/agents/:id/audit  （需 house:audit 权限）
func (h *AgentHandler) Audit(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	var req dto.AgentAuditRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	if err := h.service.Audit(id, req.Status, req.Reason); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeHouseAuditError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("审核完成", nil))
}

// UpdateOnlineStatus 经纪人本人更新在线状态
// PUT /api/v1/house/agents/online-status  （需登录 + 经纪人本人）
func (h *AgentHandler) UpdateOnlineStatus(ctx plugin.Context) {
	userID, _, _, _ := getUserProfile(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusOK, response.Unauthorized("请先登录"))
		return
	}
	var req struct {
		OnlineStatus int `json:"online_status" binding:"oneof=0 1 2 3"`
	}
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	if err := h.service.UpdateOnlineStatus(userID, req.OnlineStatus); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeHouseAgentError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("在线状态更新成功", nil))
}
