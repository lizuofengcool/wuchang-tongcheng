// Package handler love 相亲交友 HTTP 处理层 - 印象标签
package handler

import (
	"net/http"

	"wuchang-tongcheng/internal/core/plugin"
	"wuchang-tongcheng/internal/core/response"
	"wuchang-tongcheng/internal/modules/love/dto"
	"wuchang-tongcheng/internal/modules/love/service"
	"wuchang-tongcheng/internal/pkg/utils"
)

// ImpressionHandler 印象 HTTP 处理器
type ImpressionHandler struct {
	service service.LoveImpressionService
}

// NewImpressionHandler 创建 ImpressionHandler 实例
func NewImpressionHandler(svc service.LoveImpressionService) *ImpressionHandler {
	return &ImpressionHandler{service: svc}
}

// Create 创建印象
// POST /api/v1/love/impressions  （需登录）
func (h *ImpressionHandler) Create(ctx plugin.Context) {
	userID, ok := requireLogin(ctx)
	if !ok {
		return
	}
	var req dto.CreateLoveImpressionRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	loveIDStr := ctx.Query("love_id")
	loveID, _ := parseUint(loveIDStr)
	info, err := h.service.Create(loveID, userID, &req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeLoveError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("创建成功", info))
}

// Delete 删除印象
// DELETE /api/v1/love/impressions/:id  （需登录）
func (h *ImpressionHandler) Delete(ctx plugin.Context) {
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

// GetByID 印象详情
// GET /api/v1/love/impressions/:id  （公开）
func (h *ImpressionHandler) GetByID(ctx plugin.Context) {
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

// List 印象列表
// GET /api/v1/love/impressions  （公开）
func (h *ImpressionHandler) List(ctx plugin.Context) {
	var req dto.LoveImpressionListRequest
	_ = ctx.Bind(&req)
	pagination, list, err := h.service.List(&req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeLoveError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// ListByLoveID 指定用户的印象列表
// GET /api/v1/love/loves/:id/impressions  （公开）
func (h *ImpressionHandler) ListByLoveID(ctx plugin.Context) {
	loveID, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	var req dto.LoveImpressionListRequest
	_ = ctx.Bind(&req)
	pagination, list, err := h.service.ListByLoveID(loveID, &req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeLoveError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// Stats 印象统计
// GET /api/v1/love/loves/:id/impressions/stats  （公开）
func (h *ImpressionHandler) Stats(ctx plugin.Context) {
	loveID, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	resp, err := h.service.Stats(loveID)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeLoveError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(resp))
}
