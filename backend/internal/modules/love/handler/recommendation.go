// Package handler love 相亲交友 HTTP 处理层 - 推荐池
// 依据需求文档 1.10：4 维数据隔离（region_id + user_id）
// 依据 v3.2.1 架构方案：5 维度对标灵魂匹配（兴趣/性格/价值观/位置/年龄）
package handler

import (
	"net/http"

	"wuchang-tongcheng/internal/core/plugin"
	"wuchang-tongcheng/internal/core/response"
	"wuchang-tongcheng/internal/modules/love/dto"
	"wuchang-tongcheng/internal/modules/love/service"
	"wuchang-tongcheng/internal/pkg/utils"
)

// RecommendationHandler 推荐 HTTP 处理器
type RecommendationHandler struct {
	service service.LoveRecommendationService
}

// NewRecommendationHandler 创建 RecommendationHandler 实例
func NewRecommendationHandler(svc service.LoveRecommendationService) *RecommendationHandler {
	return &RecommendationHandler{service: svc}
}

// Generate 生成推荐
// POST /api/v1/love/recommendations/generate  （需登录）
func (h *RecommendationHandler) Generate(ctx plugin.Context) {
	userID, ok := requireLogin(ctx)
	if !ok {
		return
	}
	var req dto.GenerateLoveRecommendationsRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	loveIDStr := ctx.Query("love_id")
	loveID, _ := parseUint(loveIDStr)
	list, err := h.service.Generate(userID, loveID, &req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeLoveError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(list))
}

// GetByID 推荐详情
// GET /api/v1/love/recommendations/:id  （需登录）
func (h *RecommendationHandler) GetByID(ctx plugin.Context) {
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

// List 我的推荐列表
// GET /api/v1/love/recommendations  （需登录）
func (h *RecommendationHandler) List(ctx plugin.Context) {
	userID, ok := requireLogin(ctx)
	if !ok {
		return
	}
	var req dto.LoveRecommendationListRequest
	_ = ctx.Bind(&req)
	pagination, list, err := h.service.List(&req, userID)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeLoveError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// ListByType 按推荐类型查询
// GET /api/v1/love/recommendations/by-type/:type  （需登录）
func (h *RecommendationHandler) ListByType(ctx plugin.Context) {
	userID, ok := requireLogin(ctx)
	if !ok {
		return
	}
	recType := ctx.Param("type")
	var req dto.LoveRecommendationListRequest
	_ = ctx.Bind(&req)
	pagination, list, err := h.service.ListByUserAndType(userID, recType, &req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeLoveError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// Action 推荐操作（view/like/dislike/super_like/skip/dismiss）
// POST /api/v1/love/recommendations/:id/action  （需登录）
func (h *RecommendationHandler) Action(ctx plugin.Context) {
	userID, ok := requireLogin(ctx)
	if !ok {
		return
	}
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	var req dto.LoveRecommendationActionRequest
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

// Stats 推荐统计
// GET /api/v1/love/recommendations/stats  （需登录）
func (h *RecommendationHandler) Stats(ctx plugin.Context) {
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
