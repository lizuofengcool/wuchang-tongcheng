// Package handler love 相亲交友 HTTP 处理层 - 通知
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

// NotificationHandler 通知 HTTP 处理器
type NotificationHandler struct {
	service service.LoveNotificationService
}

// NewNotificationHandler 创建 NotificationHandler 实例
func NewNotificationHandler(svc service.LoveNotificationService) *NotificationHandler {
	return &NotificationHandler{service: svc}
}

// List 我的通知列表
// GET /api/v1/love/notifications  （需登录）
func (h *NotificationHandler) List(ctx plugin.Context) {
	userID, ok := requireLogin(ctx)
	if !ok {
		return
	}
	var req dto.LoveNotificationListRequest
	_ = ctx.Bind(&req)
	pagination, list, err := h.service.List(&req, userID)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeLoveError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// ListUnread 未读通知列表
// GET /api/v1/love/notifications/unread  （需登录）
func (h *NotificationHandler) ListUnread(ctx plugin.Context) {
	userID, ok := requireLogin(ctx)
	if !ok {
		return
	}
	var req dto.LoveNotificationListRequest
	_ = ctx.Bind(&req)
	pagination, list, err := h.service.ListUnread(userID, &req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeLoveError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// GetByID 通知详情
// GET /api/v1/love/notifications/:id  （需登录）
func (h *NotificationHandler) GetByID(ctx plugin.Context) {
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

// MarkRead 标记已读
// POST /api/v1/love/notifications/:id/read  （需登录）
func (h *NotificationHandler) MarkRead(ctx plugin.Context) {
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

// MarkAllRead 全部已读
// POST /api/v1/love/notifications/read-all  （需登录）
func (h *NotificationHandler) MarkAllRead(ctx plugin.Context) {
	userID, ok := requireLogin(ctx)
	if !ok {
		return
	}
	if err := h.service.MarkAllRead(userID); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeLoveError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("全部已读", nil))
}

// BatchMarkRead 批量已读
// POST /api/v1/love/notifications/batch-read  （需登录）
func (h *NotificationHandler) BatchMarkRead(ctx plugin.Context) {
	userID, ok := requireLogin(ctx)
	if !ok {
		return
	}
	var req dto.LoveNotificationBatchReadRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	if err := h.service.BatchMarkRead(req.IDs, userID); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeLoveError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("批量已读", nil))
}

// Delete 删除通知
// DELETE /api/v1/love/notifications/:id  （需登录）
func (h *NotificationHandler) Delete(ctx plugin.Context) {
	userID, ok := requireLogin(ctx)
	if !ok {
		return
	}
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	if err := h.service.Delete(id, userID); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeLoveError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("删除成功", nil))
}

// CountUnread 未读数
// GET /api/v1/love/notifications/unread-count  （需登录）
func (h *NotificationHandler) CountUnread(ctx plugin.Context) {
	userID, ok := requireLogin(ctx)
	if !ok {
		return
	}
	count, err := h.service.CountUnread(userID)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeLoveError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(map[string]int64{"count": count}))
}

// Stats 通知统计
// GET /api/v1/love/notifications/stats  （需登录）
func (h *NotificationHandler) Stats(ctx plugin.Context) {
	userID, ok := requireLogin(ctx)
	if !ok {
		return
	}
	resp, err := h.service.Stats(userID)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeLoveError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(resp))
}
