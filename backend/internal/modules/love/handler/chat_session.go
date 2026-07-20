// Package handler love 相亲交友 HTTP 处理层 - 聊天会话
// 依据需求文档 1.10：4 维数据隔离（region_id + user_id）
package handler

import (
	"net/http"

	"wuchang-tongcheng/internal/core/plugin"
	"wuchang-tongcheng/internal/core/response"
	"wuchang-tongcheng/internal/modules/love/dto"
	"wuchang-tongcheng/internal/modules/love/service"
	"wuchang-tongcheng/internal/pkg/utils"
)

// ChatSessionHandler 聊天会话 HTTP 处理器
type ChatSessionHandler struct {
	service service.LoveChatSessionService
}

// NewChatSessionHandler 创建 ChatSessionHandler 实例
func NewChatSessionHandler(svc service.LoveChatSessionService) *ChatSessionHandler {
	return &ChatSessionHandler{service: svc}
}

// GetByID 会话详情
// GET /api/v1/love/chat-sessions/:id  （需登录）
func (h *ChatSessionHandler) GetByID(ctx plugin.Context) {
	userID, ok := requireLogin(ctx)
	if !ok {
		return
	}
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	info, err := h.service.GetByID(id, userID)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeLoveError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(info))
}

// GetByMatchID 按匹配 ID 查询会话
// GET /api/v1/love/chat-sessions/by-match/:id  （需登录）
func (h *ChatSessionHandler) GetByMatchID(ctx plugin.Context) {
	_, ok := requireLogin(ctx)
	if !ok {
		return
	}
	matchID, err := parseSubID(ctx, "id")
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的 match_id"))
		return
	}
	info, err := h.service.GetByMatchID(matchID)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeLoveError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(info))
}

// List 我的会话列表
// GET /api/v1/love/chat-sessions  （需登录）
func (h *ChatSessionHandler) List(ctx plugin.Context) {
	userID, ok := requireLogin(ctx)
	if !ok {
		return
	}
	var req dto.LoveChatSessionListRequest
	_ = ctx.Bind(&req)
	pagination, list, err := h.service.List(&req, userID)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeLoveError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// Action 会话操作（mute/unmute/pin/unpin/delete/dissolve）
// POST /api/v1/love/chat-sessions/:id/action  （需登录）
func (h *ChatSessionHandler) Action(ctx plugin.Context) {
	userID, ok := requireLogin(ctx)
	if !ok {
		return
	}
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	var req dto.LoveChatSessionActionRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	req.ID = id
	if err := h.service.Action(&req, userID); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeLoveError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("操作成功", nil))
}

// MarkRead 标记已读
// POST /api/v1/love/chat-sessions/:id/read  （需登录）
func (h *ChatSessionHandler) MarkRead(ctx plugin.Context) {
	userID, ok := requireLogin(ctx)
	if !ok {
		return
	}
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	if err := h.service.MarkRead(id, userID); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeLoveError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("已读", nil))
}

// CountActive 我的活跃会话数
// GET /api/v1/love/chat-sessions/active-count  （需登录）
func (h *ChatSessionHandler) CountActive(ctx plugin.Context) {
	userID, ok := requireLogin(ctx)
	if !ok {
		return
	}
	count, err := h.service.CountActiveByUser(userID)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeLoveError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(map[string]int64{"count": count}))
}
