// Package handler IM 消息中台精简版HTTP处理层
package handler

import (
	"net/http"
	"strconv"

	"wuchang-tongcheng/internal/core/middleware"
	"wuchang-tongcheng/internal/core/plugin"
	"wuchang-tongcheng/internal/core/response"
	"wuchang-tongcheng/internal/modules/im/dto"
	"wuchang-tongcheng/internal/modules/im/service"
)

// Handler IM 中台 HTTP 处理器
type Handler struct {
	svc service.IMService
}

// NewHandler 创建 Handler 实例
func NewHandler(svc service.IMService) *Handler {
	return &Handler{svc: svc}
}

// getUserID 从上下文获取登录用户ID
func getUserID(ctx plugin.Context) uint {
	if v, ok := ctx.Get(middleware.ContextUserID); ok {
		if id, ok := v.(uint); ok {
			return id
		}
	}
	return 0
}

// getRegionID 从上下文获取地区ID
func getRegionID(ctx plugin.Context) uint {
	if v, ok := ctx.Get(middleware.RegionIDKey); ok {
		if id, ok := v.(uint); ok {
			return id
		}
	}
	return middleware.DefaultRegionID
}

// parsePagination 解析分页
func parsePagination(ctx plugin.Context) (page, pageSize int) {
	page, _ = strconv.Atoi(ctx.DefaultQuery("page", "1"))
	pageSize, _ = strconv.Atoi(ctx.DefaultQuery("page_size", "20"))
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	return
}

// CreateSession 创建会话
// POST /api/v1/im/sessions
func (h *Handler) CreateSession(ctx plugin.Context) {
	userID := getUserID(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusOK, response.Unauthorized("请先登录"))
		return
	}
	var req dto.CreateSessionRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	info, err := h.svc.CreateSession(getRegionID(ctx), userID, &req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(2901, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("创建会话成功", info))
}

// GetSession 查询会话
// GET /api/v1/im/sessions/:session_id
func (h *Handler) GetSession(ctx plugin.Context) {
	sessionID := ctx.Param("session_id")
	info, err := h.svc.GetSession(sessionID)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(2902, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(info))
}

// ListSessions 我的会话列表
// GET /api/v1/im/sessions
func (h *Handler) ListSessions(ctx plugin.Context) {
	userID := getUserID(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusOK, response.Unauthorized("请先登录"))
		return
	}
	page, pageSize := parsePagination(ctx)
	list, total, err := h.svc.ListSessions(userID, page, pageSize)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(2901, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(response.NewPageResult(list, total, page, pageSize)))
}

// SendMessage 发送消息
// POST /api/v1/im/messages
func (h *Handler) SendMessage(ctx plugin.Context) {
	userID := getUserID(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusOK, response.Unauthorized("请先登录"))
		return
	}
	var req dto.SendMessageRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	info, err := h.svc.SendMessage(getRegionID(ctx), userID, &req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(2901, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("发送成功", info))
}

// GetHistory 历史消息
// GET /api/v1/im/sessions/:session_id/messages
func (h *Handler) GetHistory(ctx plugin.Context) {
	userID := getUserID(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusOK, response.Unauthorized("请先登录"))
		return
	}
	sessionID := ctx.Param("session_id")
	page, pageSize := parsePagination(ctx)
	list, total, err := h.svc.GetHistory(sessionID, userID, page, pageSize)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(2901, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(response.NewPageResult(list, total, page, pageSize)))
}

// MarkRead 标记会话已读
// POST /api/v1/im/sessions/:session_id/read
func (h *Handler) MarkRead(ctx plugin.Context) {
	userID := getUserID(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusOK, response.Unauthorized("请先登录"))
		return
	}
	sessionID := ctx.Param("session_id")
	if err := h.svc.MarkRead(sessionID, userID); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(2901, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("已标记已读", nil))
}

// ListNotifications 系统通知列表
// GET /api/v1/im/notifications
func (h *Handler) ListNotifications(ctx plugin.Context) {
	userID := getUserID(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusOK, response.Unauthorized("请先登录"))
		return
	}
	page, pageSize := parsePagination(ctx)
	list, total, err := h.svc.ListNotifications(userID, page, pageSize)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(2901, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(response.NewPageResult(list, total, page, pageSize)))
}

// ListUnreadNotifications 未读通知
// GET /api/v1/im/notifications/unread
func (h *Handler) ListUnreadNotifications(ctx plugin.Context) {
	userID := getUserID(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusOK, response.Unauthorized("请先登录"))
		return
	}
	list, total, err := h.svc.ListUnreadNotifications(userID)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(2901, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(response.NewPageResult(list, total, 1, 20)))
}

// MarkAllNotificationsRead 标记全部通知已读
// POST /api/v1/im/notifications/read-all
func (h *Handler) MarkAllNotificationsRead(ctx plugin.Context) {
	userID := getUserID(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusOK, response.Unauthorized("请先登录"))
		return
	}
	if err := h.svc.MarkAllNotificationsRead(userID); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(2901, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("全部已读", nil))
}

// BindPrivacyNumber 绑定隐私号码（业务模块调用，如 ershou 拨打电话前绑定）
// POST /api/v1/im/privacy-numbers
func (h *Handler) BindPrivacyNumber(ctx plugin.Context) {
	var req dto.BindPrivacyNumberRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	info, err := h.svc.BindPrivacyNumber(getRegionID(ctx), &req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(2901, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("绑定成功", info))
}

// UnbindPrivacyNumber 解绑隐私号码
// POST /api/v1/im/privacy-numbers/unbind
func (h *Handler) UnbindPrivacyNumber(ctx plugin.Context) {
	var req dto.UnbindPrivacyNumberRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	if err := h.svc.UnbindPrivacyNumber(&req); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(2901, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("解绑成功", nil))
}
