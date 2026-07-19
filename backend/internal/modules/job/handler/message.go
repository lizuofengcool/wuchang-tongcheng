// Package handler 沟通消息 HTTP 处理层
package handler

import (
	"net/http"

	"wuchang-tongcheng/internal/core/plugin"
	"wuchang-tongcheng/internal/core/response"
	"wuchang-tongcheng/internal/modules/job/dto"
	"wuchang-tongcheng/internal/modules/job/service"
	"wuchang-tongcheng/internal/pkg/utils"
)

// MessageHandler 消息 HTTP 处理器
type MessageHandler struct {
	svc service.MessageService
}

// NewMessageHandler 创建消息 Handler 实例
func NewMessageHandler(svc service.MessageService) *MessageHandler {
	return &MessageHandler{svc: svc}
}

// Create 发送消息
// POST /api/v1/job/messages
func (h *MessageHandler) Create(ctx plugin.Context) {
	userID, username, _, avatar := getUserProfile(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusOK, response.Unauthorized("请先登录"))
		return
	}
	var req dto.MessageCreateRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	regionID := getRegionID(ctx)
	info, err := h.svc.Create(regionID, userID, username, avatar, &req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeMessageError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("发送成功", info))
}

// GetByID 消息详情
// GET /api/v1/job/messages/:id
func (h *MessageHandler) GetByID(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	userID, _, _, _ := getUserProfile(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusOK, response.Unauthorized("请先登录"))
		return
	}
	info, err := h.svc.GetByID(id, userID)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeMessageError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(info))
}

// List 消息列表
// GET /api/v1/job/messages
func (h *MessageHandler) List(ctx plugin.Context) {
	userID, _, _, _ := getUserProfile(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusOK, response.Unauthorized("请先登录"))
		return
	}
	var req dto.MessageListQuery
	_ = ctx.Bind(&req)
	pagination, list, err := h.svc.List(userID, &req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeMessageError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// ListByConversation 会话消息列表
// GET /api/v1/job/conversations/:conversation_id/messages
func (h *MessageHandler) ListByConversation(ctx plugin.Context) {
	userID, _, _, _ := getUserProfile(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusOK, response.Unauthorized("请先登录"))
		return
	}
	convID := ctx.Param("conversation_id")
	if convID == "" {
		ctx.JSON(http.StatusOK, response.BadRequest("会话 ID 不能为空"))
		return
	}
	page, pageSize := parsePagination(ctx)
	pagination, list, err := h.svc.ListByConversation(convID, userID, page, pageSize)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeMessageError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// ListConversations 会话列表
// GET /api/v1/job/conversations
func (h *MessageHandler) ListConversations(ctx plugin.Context) {
	userID, _, _, _ := getUserProfile(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusOK, response.Unauthorized("请先登录"))
		return
	}
	list, err := h.svc.ListConversations(userID)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeMessageError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(list))
}

// MarkRead 标记会话已读
// PUT /api/v1/job/conversations/:conversation_id/read
func (h *MessageHandler) MarkRead(ctx plugin.Context) {
	userID, _, _, _ := getUserProfile(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusOK, response.Unauthorized("请先登录"))
		return
	}
	convID := ctx.Param("conversation_id")
	if convID == "" {
		ctx.JSON(http.StatusOK, response.BadRequest("会话 ID 不能为空"))
		return
	}
	if err := h.svc.MarkRead(convID, userID); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeMessageError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("标记已读成功", nil))
}

// CountUnread 未读消息数
// GET /api/v1/job/messages/unread/count
func (h *MessageHandler) CountUnread(ctx plugin.Context) {
	userID, _, _, _ := getUserProfile(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusOK, response.Unauthorized("请先登录"))
		return
	}
	count, err := h.svc.CountUnread(userID)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeMessageError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(map[string]int64{"unread_count": count}))
}

// Delete 删除消息
// DELETE /api/v1/job/messages/:id
func (h *MessageHandler) Delete(ctx plugin.Context) {
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
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeMessageError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("删除成功", nil))
}

// BatchDelete 批量删除
// POST /api/v1/job/messages/batch-delete
func (h *MessageHandler) BatchDelete(ctx plugin.Context) {
	userID, _, _, _ := getUserProfile(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusOK, response.Unauthorized("请先登录"))
		return
	}
	var req dto.MessageBatchDeleteRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	if err := h.svc.BatchDelete(userID, req.IDs); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeMessageError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("批量删除成功", nil))
}

// Recall 撤回消息
// PUT /api/v1/job/messages/:id/recall
func (h *MessageHandler) Recall(ctx plugin.Context) {
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
	if err := h.svc.Recall(id, userID); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeMessageError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("撤回成功", nil))
}
