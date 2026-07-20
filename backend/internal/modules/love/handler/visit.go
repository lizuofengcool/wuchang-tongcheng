// Package handler love 相亲交友 HTTP 处理层 - 访客
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

// VisitHandler 访客 HTTP 处理器
type VisitHandler struct {
	service service.LoveVisitService
}

// NewVisitHandler 创建 VisitHandler 实例
func NewVisitHandler(svc service.LoveVisitService) *VisitHandler {
	return &VisitHandler{service: svc}
}

// Visit 访问他人主页（记录访客）
// POST /api/v1/love/visits  （需登录）
func (h *VisitHandler) Visit(ctx plugin.Context) {
	userID, username, avatar := getUserProfile(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusOK, response.Unauthorized("请先登录"))
		return
	}
	var req dto.CreateLoveVisitRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	loveIDStr := ctx.Query("love_id")
	loveID, _ := parseUint(loveIDStr)
	genderStr := ctx.Query("gender")
	gender, _ := parseUint(genderStr)
	info, err := h.service.Visit(userID, loveID, username, avatar, int(gender), &req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeLoveError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(info))
}

// GetByID 访客记录详情
// GET /api/v1/love/visits/:id  （需登录）
func (h *VisitHandler) GetByID(ctx plugin.Context) {
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

// List 我的访客列表（别人来看我）
// GET /api/v1/love/visits  （需登录）
func (h *VisitHandler) List(ctx plugin.Context) {
	userID, ok := requireLogin(ctx)
	if !ok {
		return
	}
	var req dto.LoveVisitListRequest
	_ = ctx.Bind(&req)
	pagination, list, err := h.service.List(&req, userID)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeLoveError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// ListByVisitor 我看过谁
// GET /api/v1/love/visits/by-visitor  （需登录）
func (h *VisitHandler) ListByVisitor(ctx plugin.Context) {
	userID, ok := requireLogin(ctx)
	if !ok {
		return
	}
	var req dto.LoveVisitListRequest
	_ = ctx.Bind(&req)
	pagination, list, err := h.service.ListByVisitor(userID, &req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeLoveError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// ListUnread 未读访客
// GET /api/v1/love/visits/unread  （需登录）
func (h *VisitHandler) ListUnread(ctx plugin.Context) {
	userID, ok := requireLogin(ctx)
	if !ok {
		return
	}
	var req dto.LoveVisitListRequest
	_ = ctx.Bind(&req)
	pagination, list, err := h.service.ListUnread(userID, &req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeLoveError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// MarkRead 标记已读
// POST /api/v1/love/visits/:id/read  （需登录）
func (h *VisitHandler) MarkRead(ctx plugin.Context) {
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
// POST /api/v1/love/visits/read-all  （需登录）
func (h *VisitHandler) MarkAllRead(ctx plugin.Context) {
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

// Stats 访客统计
// GET /api/v1/love/visits/stats  （需登录）
func (h *VisitHandler) Stats(ctx plugin.Context) {
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
