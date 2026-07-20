// Package handler love 相亲交友 HTTP 处理层 - 喜欢/不喜欢/心动信号
package handler

import (
	"net/http"

	"wuchang-tongcheng/internal/core/plugin"
	"wuchang-tongcheng/internal/core/response"
	"wuchang-tongcheng/internal/modules/love/dto"
	"wuchang-tongcheng/internal/modules/love/service"
	"wuchang-tongcheng/internal/pkg/utils"
)

// LikeHandler 喜欢 HTTP 处理器
type LikeHandler struct {
	service service.LoveLikeService
}

// NewLikeHandler 创建 LikeHandler 实例
func NewLikeHandler(svc service.LoveLikeService) *LikeHandler {
	return &LikeHandler{service: svc}
}

// Act 执行喜欢/不喜欢/跳过/心动信号动作
// POST /api/v1/love/likes  （需登录）
func (h *LikeHandler) Act(ctx plugin.Context) {
	userID, ok := requireLogin(ctx)
	if !ok {
		return
	}
	var req dto.CreateLoveLikeRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	loveIDStr := ctx.Query("love_id")
	loveID, _ := parseUint(loveIDStr)
	info, err := h.service.Act(loveID, userID, &req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeLoveError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("操作成功", info))
}

// Undo 撤销操作
// POST /api/v1/love/likes/:id/undo  （需登录）
func (h *LikeHandler) Undo(ctx plugin.Context) {
	userID, ok := requireLogin(ctx)
	if !ok {
		return
	}
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	var req dto.UndoLoveLikeRequest
	_ = ctx.Bind(&req)
	if err := h.service.Undo(id, userID, &req); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeLoveError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("撤销成功", nil))
}

// GetByID 查询喜欢记录详情
// GET /api/v1/love/likes/:id  （需登录）
func (h *LikeHandler) GetByID(ctx plugin.Context) {
	userID, ok := requireLogin(ctx)
	if !ok {
		return
	}
	_ = userID
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	info, err := h.service.GetByID(id)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeLoveError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(info))
}

// List 我喜欢/被喜欢列表
// GET /api/v1/love/likes  （需登录）
func (h *LikeHandler) List(ctx plugin.Context) {
	userID, ok := requireLogin(ctx)
	if !ok {
		return
	}
	var req dto.LoveLikeListRequest
	_ = ctx.Bind(&req)
	pagination, list, err := h.service.List(&req, userID)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeLoveError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// ListMatched 匹配列表
// GET /api/v1/love/likes/matched  （需登录）
func (h *LikeHandler) ListMatched(ctx plugin.Context) {
	userID, ok := requireLogin(ctx)
	if !ok {
		return
	}
	var req dto.LoveLikeListRequest
	_ = ctx.Bind(&req)
	pagination, list, err := h.service.ListMatched(userID, &req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeLoveError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// HasLiked 查询是否已喜欢
// GET /api/v1/love/likes/check  （需登录）
func (h *LikeHandler) HasLiked(ctx plugin.Context) {
	userID, ok := requireLogin(ctx)
	if !ok {
		return
	}
	targetUserIDStr := ctx.Query("target_user_id")
	targetUserID, err := parseUint(targetUserIDStr)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("target_user_id 无效"))
		return
	}
	liked, err := h.service.HasLiked(userID, targetUserID)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeLoveError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(map[string]bool{"liked": liked}))
}

// TodayStats 今日喜欢统计（依据会员等级限制）
// GET /api/v1/love/likes/today-stats  （需登录）
func (h *LikeHandler) TodayStats(ctx plugin.Context) {
	userID, ok := requireLogin(ctx)
	if !ok {
		return
	}
	memberLevelStr := ctx.Query("member_level")
	memberLevel, _ := parseUint(memberLevelStr)
	resp, err := h.service.TodayStats(userID, int(memberLevel))
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeLoveError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(resp))
}
