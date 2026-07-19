// Package handler IM 中台扩展 HTTP 处理层
// 依据 013_im_full.sql：群组/群成员/用户设置/消息撤回/统计
package handler

import (
	"net/http"

	"wuchang-tongcheng/internal/core/plugin"
	"wuchang-tongcheng/internal/core/response"
	"wuchang-tongcheng/internal/modules/im/dto"
	"wuchang-tongcheng/internal/modules/im/service"
	"wuchang-tongcheng/internal/pkg/utils"
)

// ExtendHandler IM 中台扩展处理器
type ExtendHandler struct {
	extSvc service.IMExtendService
}

// NewExtendHandler 创建扩展 Handler 实例
func NewExtendHandler(extSvc service.IMExtendService) *ExtendHandler {
	return &ExtendHandler{extSvc: extSvc}
}

// ===== 群组 =====

// CreateGroup 创建群组
// POST /api/v1/im/groups
func (h *ExtendHandler) CreateGroup(ctx plugin.Context) {
	userID := getUserID(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusOK, response.Unauthorized("请先登录"))
		return
	}
	var req dto.CreateGroupRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	info, err := h.extSvc.CreateGroup(getRegionID(ctx), userID, &req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeIMError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("创建群组成功", info))
}

// UpdateGroup 更新群组
// POST /api/v1/im/groups/:group_id
func (h *ExtendHandler) UpdateGroup(ctx plugin.Context) {
	userID := getUserID(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusOK, response.Unauthorized("请先登录"))
		return
	}
	groupID := ctx.Param("group_id")
	if groupID == "" {
		ctx.JSON(http.StatusOK, response.BadRequest("群组ID不能为空"))
		return
	}
	var req dto.UpdateGroupRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	if err := h.extSvc.UpdateGroup(userID, groupID, &req); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeIMGroupNotFound, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("更新成功", nil))
}

// DissolveGroup 解散群组
// DELETE /api/v1/im/groups/:group_id
func (h *ExtendHandler) DissolveGroup(ctx plugin.Context) {
	userID := getUserID(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusOK, response.Unauthorized("请先登录"))
		return
	}
	groupID := ctx.Param("group_id")
	if err := h.extSvc.DissolveGroup(userID, groupID); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeIMGroupNotFound, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("群组已解散", nil))
}

// GetGroup 查询群组
// GET /api/v1/im/groups/:group_id
func (h *ExtendHandler) GetGroup(ctx plugin.Context) {
	groupID := ctx.Param("group_id")
	info, err := h.extSvc.GetGroup(groupID)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeIMGroupNotFound, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(info))
}

// ListMyGroups 我的群组列表
// GET /api/v1/im/groups
func (h *ExtendHandler) ListMyGroups(ctx plugin.Context) {
	userID := getUserID(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusOK, response.Unauthorized("请先登录"))
		return
	}
	page, pageSize := parsePagination(ctx)
	list, total, err := h.extSvc.ListMyGroups(userID, page, pageSize)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeIMError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(response.NewPageResult(list, total, page, pageSize)))
}

// ===== 群成员 =====

// AddGroupMembers 添加群成员
// POST /api/v1/im/groups/members
func (h *ExtendHandler) AddGroupMembers(ctx plugin.Context) {
	userID := getUserID(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusOK, response.Unauthorized("请先登录"))
		return
	}
	var req dto.AddGroupMembersRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	if err := h.extSvc.AddGroupMembers(userID, &req); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeIMError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("成员添加成功", nil))
}

// RemoveMember 移除群成员
// POST /api/v1/im/groups/members/remove
func (h *ExtendHandler) RemoveMember(ctx plugin.Context) {
	userID := getUserID(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusOK, response.Unauthorized("请先登录"))
		return
	}
	var req dto.RemoveGroupMemberRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	if err := h.extSvc.RemoveMember(userID, &req); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeIMGroupMemberNotFound, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("成员已移除", nil))
}

// ListGroupMembers 群成员列表
// GET /api/v1/im/groups/:group_id/members
func (h *ExtendHandler) ListGroupMembers(ctx plugin.Context) {
	groupID := ctx.Param("group_id")
	list, err := h.extSvc.ListGroupMembers(groupID)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeIMError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(list))
}

// ===== 用户设置 =====

// GetMySetting 获取我的 IM 设置
// GET /api/v1/im/settings
func (h *ExtendHandler) GetMySetting(ctx plugin.Context) {
	userID := getUserID(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusOK, response.Unauthorized("请先登录"))
		return
	}
	info, err := h.extSvc.GetMySetting(userID, getRegionID(ctx))
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeIMUserSettingsNotFound, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(info))
}

// UpdateMySetting 更新我的 IM 设置
// POST /api/v1/im/settings
func (h *ExtendHandler) UpdateMySetting(ctx plugin.Context) {
	userID := getUserID(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusOK, response.Unauthorized("请先登录"))
		return
	}
	var req dto.UpdateUserSettingRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	if err := h.extSvc.UpdateMySetting(userID, getRegionID(ctx), &req); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeIMError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("设置已更新", nil))
}

// ===== 消息撤回 =====

// RecallMessage 撤回消息
// POST /api/v1/im/messages/recall
func (h *ExtendHandler) RecallMessage(ctx plugin.Context) {
	userID := getUserID(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusOK, response.Unauthorized("请先登录"))
		return
	}
	var req dto.RecallMessageRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	if err := h.extSvc.RecallMessage(userID, &req); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeIMMessageNotFound, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("消息已撤回", nil))
}

// ===== 统计 =====

// Statistics IM 总览统计（M 端）
// GET /api/v1/im/admin/statistics
func (h *ExtendHandler) Statistics(ctx plugin.Context) {
	resp, err := h.extSvc.Statistics()
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeIMError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(resp))
}
