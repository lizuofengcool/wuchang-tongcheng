// Package handler love 相亲交友 HTTP 处理层 - 匹配
package handler

import (
	"net/http"

	"wuchang-tongcheng/internal/core/plugin"
	"wuchang-tongcheng/internal/core/response"
	"wuchang-tongcheng/internal/modules/love/dto"
	"wuchang-tongcheng/internal/modules/love/service"
	"wuchang-tongcheng/internal/pkg/utils"
)

// MatchHandler 匹配 HTTP 处理器
type MatchHandler struct {
	service service.LoveMatchService
}

// NewMatchHandler 创建 MatchHandler 实例
func NewMatchHandler(svc service.LoveMatchService) *MatchHandler {
	return &MatchHandler{service: svc}
}

// GetByID 匹配详情
// GET /api/v1/love/matches/:id  （需登录）
func (h *MatchHandler) GetByID(ctx plugin.Context) {
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

// List 我的匹配列表
// GET /api/v1/love/matches  （需登录）
func (h *MatchHandler) List(ctx plugin.Context) {
	userID, ok := requireLogin(ctx)
	if !ok {
		return
	}
	var req dto.LoveMatchListRequest
	_ = ctx.Bind(&req)
	pagination, list, err := h.service.List(&req, userID)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeLoveError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// Dissolve 解除匹配
// POST /api/v1/love/matches/:id/dissolve  （需登录）
func (h *MatchHandler) Dissolve(ctx plugin.Context) {
	userID, ok := requireLogin(ctx)
	if !ok {
		return
	}
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	var req dto.DissolveMatchRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	if err := h.service.Dissolve(id, userID, &req); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeLoveError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("解除匹配成功", nil))
}

// CountByUser 我的匹配数
// GET /api/v1/love/matches/count  （需登录）
func (h *MatchHandler) CountByUser(ctx plugin.Context) {
	userID, ok := requireLogin(ctx)
	if !ok {
		return
	}
	count, err := h.service.CountByUser(userID)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeLoveError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(map[string]int64{"count": count}))
}

// CountToday 今日匹配数
// GET /api/v1/love/matches/today-count  （需登录）
func (h *MatchHandler) CountToday(ctx plugin.Context) {
	userID, ok := requireLogin(ctx)
	if !ok {
		return
	}
	count, err := h.service.CountTodayByUser(userID)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeLoveError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(map[string]int64{"count": count}))
}
